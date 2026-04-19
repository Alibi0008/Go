package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetRateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/convert" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("from"); got != "USD" {
			t.Fatalf("unexpected from query: %s", got)
		}

		if got := r.URL.Query().Get("to"); got != "EUR" {
			t.Fatalf("unexpected to query: %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":0.92}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")
	if err != nil {
		t.Fatalf("GetRate returned unexpected error: %v", err)
	}

	if rate != 0.92 {
		t.Errorf("GetRate returned rate %v; want %v", rate, 0.92)
	}
}

func TestGetRateAPIBusinessError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"invalid currency pair"}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("AAA", "BBB")
	if err == nil {
		t.Fatal("GetRate expected an error, got nil")
	}

	if !strings.Contains(err.Error(), "api error: invalid currency pair") {
		t.Errorf("GetRate error = %q; want to contain %q", err.Error(), "api error: invalid currency pair")
	}
}

func TestGetRateMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("GetRate expected a decode error, got nil")
	}

	if !strings.Contains(err.Error(), "decode error") {
		t.Errorf("GetRate error = %q; want to contain %q", err.Error(), "decode error")
	}
}

func TestGetRateTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"base":"USD","target":"EUR","rate":0.92}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)
	service.Client.Timeout = 50 * time.Millisecond

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("GetRate expected a timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("GetRate error = %q; want to contain %q", err.Error(), "network error")
	}
}

func TestGetRateInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"internal server error"}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("GetRate expected an error, got nil")
	}

	if !strings.Contains(err.Error(), "api error: internal server error") {
		t.Errorf("GetRate error = %q; want to contain %q", err.Error(), "api error: internal server error")
	}
}

func TestGetRateEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("GetRate expected a decode error, got nil")
	}

	if !strings.Contains(err.Error(), "decode error") {
		t.Errorf("GetRate error = %q; want to contain %q", err.Error(), "decode error")
	}
}
