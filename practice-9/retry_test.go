package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsRetryable(t *testing.T) {
	t.Parallel()

	timeoutErr := &net.DNSError{IsTimeout: true}
	if !IsRetryable(nil, timeoutErr) {
		t.Fatalf("expected timeout error to be retryable")
	}

	retryableStatuses := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, status := range retryableStatuses {
		resp := &http.Response{StatusCode: status}
		if !IsRetryable(resp, nil) {
			t.Fatalf("expected status %d to be retryable", status)
		}
	}

	nonRetryableStatuses := []int{
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusBadRequest,
	}
	for _, status := range nonRetryableStatuses {
		resp := &http.Response{StatusCode: status}
		if IsRetryable(resp, nil) {
			t.Fatalf("expected status %d to be non-retryable", status)
		}
	}
}

func TestExecutePaymentRetriesUntilSuccess(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)
		if call <= 3 {
			http.Error(w, `{"error":"temporary failure"}`, http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(paymentResponse{Status: "success"})
	}))
	defer server.Close()

	client := &PaymentClient{
		HTTPClient: server.Client(),
		URL:        server.URL,
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
		Logger:     testLogger(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.ExecutePayment(ctx, paymentRequest{
		OrderID: "test",
		Amount:  1000,
	})
	if err != nil {
		t.Fatalf("ExecutePayment returned error: %v", err)
	}

	if resp.Status != "success" {
		t.Fatalf("unexpected response status: %s", resp.Status)
	}

	if calls.Load() != 4 {
		t.Fatalf("expected 4 calls, got %d", calls.Load())
	}
}
