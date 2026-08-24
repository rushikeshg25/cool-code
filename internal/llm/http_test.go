package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostJSONRetriesTransientErrors(t *testing.T) {
	old := retryBaseDelay
	retryBaseDelay = time.Millisecond
	defer func() { retryBaseDelay = old }()

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.WriteHeader(http.StatusInternalServerError)
		case 2:
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer srv.Close()

	var out struct {
		OK bool `json:"ok"`
	}
	if err := postJSON(context.Background(), srv.URL, nil, map[string]any{}, &out); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
	if !out.OK {
		t.Fatal("response not decoded")
	}
}

func TestPostJSONNoRetryOnClientError(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	err := postJSON(context.Background(), srv.URL, nil, map[string]any{}, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("expected HTTP 400 error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("400 must not be retried, got %d attempts", calls)
	}
}

func TestPostJSONCancelDuringBackoff(t *testing.T) {
	old := retryBaseDelay
	retryBaseDelay = time.Minute // long enough that only cancellation can end the sleep
	defer func() { retryBaseDelay = old }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- postJSON(ctx, srv.URL, nil, map[string]any{}, &struct{}{}) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("postJSON did not return after cancellation")
	}
}

func TestHTTPErrorRedactsSecretsAndEndpoint(t *testing.T) {
	t.Setenv("COOLCODE_TEST_API_KEY", "super-secret-api-value")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"apiKey=super-secret-api-value"}}`))
	}))
	defer srv.Close()
	err := postJSON(context.Background(), srv.URL+"/internal/provider", nil, map[string]any{}, &struct{}{})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-api-value") || strings.Contains(err.Error(), srv.URL) {
		t.Fatalf("sensitive error details leaked: %v", err)
	}
}
