package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	idempotencyHeader   = "Idempotency-Key"
	processingTTL       = 30 * time.Second
	completedRecordTTL  = 24 * time.Hour
	defaultRequestDelay = 2 * time.Second
	storeWriteTimeout   = 2 * time.Second
)

type storeState string

const (
	stateProcessing storeState = "processing"
	stateCompleted  storeState = "completed"
)

type paymentReceipt struct {
	Status        string `json:"status"`
	Amount        int    `json:"amount"`
	TransactionID string `json:"transaction_id"`
}

type storedResponse struct {
	StatusCode int         `json:"status_code"`
	Header     http.Header `json:"header"`
	Body       []byte      `json:"body"`
}

type idempotencyRecord struct {
	State    storeState      `json:"state"`
	Response *storedResponse `json:"response,omitempty"`
}

type IdempotencyStore interface {
	Get(ctx context.Context, key string) (idempotencyRecord, bool, error)
	TryStart(ctx context.Context, key string) (bool, error)
	Complete(ctx context.Context, key string, response storedResponse) error
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(ctx context.Context, cfg RedisConfig) (*RedisIdempotencyStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &RedisIdempotencyStore{client: client}, nil
}

func (s *RedisIdempotencyStore) Close() error {
	return s.client.Close()
}

func (s *RedisIdempotencyStore) Get(ctx context.Context, key string) (idempotencyRecord, bool, error) {
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return idempotencyRecord{}, false, nil
	}
	if err != nil {
		return idempotencyRecord{}, false, err
	}

	var record idempotencyRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		return idempotencyRecord{}, false, fmt.Errorf("decode idempotency record: %w", err)
	}

	return record, true, nil
}

func (s *RedisIdempotencyStore) TryStart(ctx context.Context, key string) (bool, error) {
	payload, err := json.Marshal(idempotencyRecord{State: stateProcessing})
	if err != nil {
		return false, fmt.Errorf("encode processing record: %w", err)
	}

	started, err := s.client.SetNX(ctx, key, payload, processingTTL).Result()
	if err != nil {
		return false, err
	}

	return started, nil
}

func (s *RedisIdempotencyStore) Complete(ctx context.Context, key string, response storedResponse) error {
	payload, err := json.Marshal(idempotencyRecord{
		State:    stateCompleted,
		Response: &response,
	})
	if err != nil {
		return fmt.Errorf("encode completed record: %w", err)
	}

	return s.client.Set(ctx, key, payload, completedRecordTTL).Err()
}

type MemoryIdempotencyStore struct {
	mu      sync.Mutex
	records map[string]idempotencyRecord
}

func NewMemoryIdempotencyStore() *MemoryIdempotencyStore {
	return &MemoryIdempotencyStore{
		records: make(map[string]idempotencyRecord),
	}
}

func (s *MemoryIdempotencyStore) Get(_ context.Context, key string) (idempotencyRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.records[key]
	return record, exists, nil
}

func (s *MemoryIdempotencyStore) TryStart(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.records[key]; exists {
		return false, nil
	}

	s.records[key] = idempotencyRecord{State: stateProcessing}
	return true, nil
}

func (s *MemoryIdempotencyStore) Complete(_ context.Context, key string, response storedResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[key] = idempotencyRecord{
		State:    stateCompleted,
		Response: &response,
	}
	return nil
}

func IdempotencyMiddleware(store IdempotencyStore, logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get(idempotencyHeader))
		if key == "" {
			http.Error(w, "missing Idempotency-Key header", http.StatusBadRequest)
			return
		}

		record, exists, err := store.Get(r.Context(), key)
		if err != nil {
			logger.Printf("[%s] failed to read idempotency store: %v", key, err)
			http.Error(w, "idempotency store unavailable", http.StatusInternalServerError)
			return
		}
		if exists {
			respondFromRecord(w, logger, key, record)
			return
		}

		started, err := store.TryStart(r.Context(), key)
		if err != nil {
			logger.Printf("[%s] failed to mark request as processing: %v", key, err)
			http.Error(w, "idempotency store unavailable", http.StatusInternalServerError)
			return
		}
		if !started {
			record, exists, err := store.Get(r.Context(), key)
			if err != nil {
				logger.Printf("[%s] failed to re-read idempotency store: %v", key, err)
				http.Error(w, "idempotency store unavailable", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "request state changed unexpectedly", http.StatusConflict)
				return
			}

			respondFromRecord(w, logger, key, record)
			return
		}

		logger.Printf("[%s] processing started", key)

		recorder := newResponseCapture(w)
		next.ServeHTTP(recorder, r)

		storeCtx, cancel := context.WithTimeout(context.Background(), storeWriteTimeout)
		defer cancel()

		if err := store.Complete(storeCtx, key, recorder.StoredResponse()); err != nil {
			logger.Printf("[%s] failed to cache completed response: %v", key, err)
			return
		}

		logger.Printf("[%s] processing completed and cached", key)
	})
}

func respondFromRecord(w http.ResponseWriter, logger *log.Logger, key string, record idempotencyRecord) {
	switch record.State {
	case stateProcessing:
		logger.Printf("[%s] duplicate request blocked with 409", key)
		http.Error(w, "request is already being processed", http.StatusConflict)
	case stateCompleted:
		if record.Response == nil {
			http.Error(w, "cached response is corrupted", http.StatusInternalServerError)
			return
		}

		logger.Printf("[%s] returning cached response", key)
		copyHeaders(w.Header(), record.Response.Header)
		w.WriteHeader(record.Response.StatusCode)
		_, _ = w.Write(record.Response.Body)
	default:
		http.Error(w, "unknown idempotency state", http.StatusInternalServerError)
	}
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newResponseCapture(w http.ResponseWriter) *responseCapture {
	return &responseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseCapture) Write(body []byte) (int, error) {
	r.body.Write(body)
	return r.ResponseWriter.Write(body)
}

func (r *responseCapture) StoredResponse() storedResponse {
	return storedResponse{
		StatusCode: r.statusCode,
		Header:     cloneHeader(r.Header()),
		Body:       append([]byte(nil), r.body.Bytes()...),
	}
}

func runIdempotencyScenario(ctx context.Context, logger *log.Logger, store IdempotencyStore) error {
	paymentHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("business logic: heavy payment processing started")
		time.Sleep(defaultRequestDelay)

		receipt := paymentReceipt{
			Status:        "paid",
			Amount:        1000,
			TransactionID: newTransactionID(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(receipt); err != nil {
			logger.Printf("failed to write payment response: %v", err)
			return
		}
		logger.Printf("business logic: payment completed with transaction %s", receipt.TransactionID)
	})

	server := httptest.NewServer(IdempotencyMiddleware(store, logger, paymentHandler))
	defer server.Close()

	key := fmt.Sprintf("loan-payment-%d", time.Now().UnixNano())
	totalRequests := 6

	type result struct {
		name string
		code int
		body string
		err  error
	}

	results := make(chan result, totalRequests+1)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	for i := 1; i <= totalRequests; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start

			code, body, err := sendIdempotentRequest(ctx, server.Client(), server.URL, key)
			results <- result{
				name: fmt.Sprintf("parallel request %d", idx),
				code: code,
				body: body,
				err:  err,
			}
		}(i)
	}

	close(start)
	wg.Wait()

	code, body, err := sendIdempotentRequest(ctx, server.Client(), server.URL, key)
	results <- result{
		name: "follow-up request after completion",
		code: code,
		body: body,
		err:  err,
	}
	close(results)

	for item := range results {
		if item.err != nil {
			return fmt.Errorf("%s failed: %w", item.name, item.err)
		}
		logger.Printf("%s -> HTTP %d, body: %s", item.name, item.code, item.body)
	}

	return nil
}

func sendIdempotentRequest(parent context.Context, client *http.Client, url, key string) (int, string, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return 0, "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set(idempotencyHeader, key)

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("read body: %w", err)
	}

	return resp.StatusCode, strings.TrimSpace(string(body)), nil
}

func newTransactionID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return "uuid-" + hex.EncodeToString(raw[:])
}

func cloneHeader(header http.Header) http.Header {
	cloned := make(http.Header, len(header))
	for key, values := range header {
		copied := append([]string(nil), values...)
		cloned[key] = copied
	}
	return cloned
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
