package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIdempotencyMiddlewareReturns400WithoutKey(t *testing.T) {
	t.Parallel()

	store := NewMemoryIdempotencyStore()
	handler := IdempotencyMiddleware(store, testLogger(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIdempotencyMiddlewareBlocksDuplicatesAndCachesResponse(t *testing.T) {
	t.Parallel()

	store := NewMemoryIdempotencyStore()
	var businessCalls atomic.Int32

	handler := IdempotencyMiddleware(store, testLogger(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		businessCalls.Add(1)
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(paymentReceipt{
			Status:        "paid",
			Amount:        1000,
			TransactionID: "uuid-test",
		})
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	key := "same-key"

	type result struct {
		code int
		body string
		err  error
	}

	start := make(chan struct{})
	results := make(chan result, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start

			code, body, err := sendIdempotentRequest(context.Background(), client, server.URL, key)
			results <- result{code: code, body: body, err: err}
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var sawConflict bool
	var sawSuccess bool
	for item := range results {
		if item.err != nil {
			t.Fatalf("request failed: %v", item.err)
		}
		switch item.code {
		case http.StatusConflict:
			sawConflict = true
		case http.StatusOK:
			sawSuccess = true
		default:
			t.Fatalf("unexpected status code %d", item.code)
		}
	}

	if !sawConflict || !sawSuccess {
		t.Fatalf("expected one 409 and one 200, got conflict=%v success=%v", sawConflict, sawSuccess)
	}

	if businessCalls.Load() != 1 {
		t.Fatalf("expected business logic to run once, got %d", businessCalls.Load())
	}

	code, body, err := sendIdempotentRequest(context.Background(), client, server.URL, key)
	if err != nil {
		t.Fatalf("follow-up request failed: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected cached response status 200, got %d", code)
	}
	if body == "" {
		t.Fatalf("expected cached response body")
	}

	if businessCalls.Load() != 1 {
		t.Fatalf("expected business logic to remain at one call, got %d", businessCalls.Load())
	}
}
