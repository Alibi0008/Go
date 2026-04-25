package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMaxRetries = 5
	defaultBaseDelay  = 500 * time.Millisecond
	defaultMaxDelay   = 5 * time.Second
)

var jitterRand = struct {
	mu  sync.Mutex
	src *rand.Rand
}{
	src: rand.New(rand.NewSource(time.Now().UnixNano())),
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int    `json:"amount"`
}

type paymentResponse struct {
	Status string `json:"status"`
}

type PaymentClient struct {
	HTTPClient *http.Client
	URL        string
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Logger     *log.Logger
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}

		return false
	}

	if resp == nil {
		return false
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	case http.StatusUnauthorized, http.StatusNotFound:
		return false
	default:
		return false
	}
}

func CalculateBackoff(attempt int) time.Duration {
	return calculateBackoff(attempt, defaultBaseDelay, defaultMaxDelay)
}

func calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	backoff := float64(baseDelay) * math.Pow(2, float64(attempt))
	current := time.Duration(backoff)
	if current > maxDelay {
		current = maxDelay
	}
	if current <= 0 {
		return 0
	}

	jitterRand.mu.Lock()
	defer jitterRand.mu.Unlock()

	// Full jitter spreads retries in [0, current].
	return time.Duration(jitterRand.src.Int63n(int64(current) + 1))
}

func (c *PaymentClient) ExecutePayment(ctx context.Context, payload paymentRequest) (paymentResponse, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 2 * time.Second}
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = defaultMaxRetries
	}
	if c.BaseDelay <= 0 {
		c.BaseDelay = defaultBaseDelay
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = defaultMaxDelay
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return paymentResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
		if err != nil {
			return paymentResponse{}, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()

			var result paymentResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
				return paymentResponse{}, fmt.Errorf("decode success response: %w", decodeErr)
			}

			c.Logger.Printf("Attempt %d: success!", attempt)
			return result, nil
		}

		lastErr = buildHTTPError(resp, err)
		retryable := IsRetryable(resp, err)

		if resp != nil && resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		if !retryable {
			return paymentResponse{}, fmt.Errorf("attempt %d failed with non-retryable error: %w", attempt, lastErr)
		}
		if attempt == c.MaxRetries {
			break
		}

		waitFor := calculateBackoff(attempt-1, c.BaseDelay, c.MaxDelay)
		c.Logger.Printf("Attempt %d failed: %v. Waiting %v before retry...", attempt, lastErr, waitFor)

		timer := time.NewTimer(waitFor)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return paymentResponse{}, fmt.Errorf("payment execution aborted: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return paymentResponse{}, fmt.Errorf("payment failed after %d attempts: %w", c.MaxRetries, lastErr)
}

func buildHTTPError(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("request failed without response")
	}
	return fmt.Errorf("unexpected status %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}

func runRetryScenario(logger *log.Logger) error {
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)
		logger.Printf("payment gateway received request #%d", call)

		if call <= 3 {
			http.Error(w, `{"error":"gateway temporarily unavailable"}`, http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paymentResponse{Status: "success"})
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &PaymentClient{
		HTTPClient: server.Client(),
		URL:        server.URL,
		MaxRetries: defaultMaxRetries,
		BaseDelay:  defaultBaseDelay,
		MaxDelay:   defaultMaxDelay,
		Logger:     logger,
	}

	result, err := client.ExecutePayment(ctx, paymentRequest{
		OrderID: "order-black-friday-1",
		Amount:  1000,
	})
	if err != nil {
		return err
	}

	logger.Printf("final payment result: %+v", result)
	return nil
}
