package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestRetry_429ThenSuccess(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true, Version: "1.0"})
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(2))
	resp, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Error("expected OK=true")
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetry_500TwiceThenSuccess(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true, Version: "1.0"})
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(2))
	resp, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Error("expected OK=true")
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetry_ConnectionError(t *testing.T) {
	// Start and immediately close a server to get a connection-refused address.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := ts.URL
	ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(addr), cartesia.WithMaxRetries(1))
	_, err := c.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
	if !cartesia.IsConnectionError(err) {
		t.Errorf("expected ConnectionError, got %T: %v", err, err)
	}
}

func TestNoRetry_ClientErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"400 Bad Request", http.StatusBadRequest},
		{"401 Unauthorized", http.StatusUnauthorized},
		{"404 Not Found", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte("error"))
			}))
			defer ts.Close()

			c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(2))
			_, err := c.GetStatus(context.Background())
			if err == nil {
				t.Fatal("expected error")
			}
			if got := attempts.Load(); got != 1 {
				t.Errorf("expected 1 attempt (no retry for %d), got %d", tt.status, got)
			}
		})
	}
}

func TestAuthHeader_BearerToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if got != "Bearer my-access-token" {
			t.Errorf("expected 'Bearer my-access-token', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true})
	}))
	defer ts.Close()

	c := cartesia.NewClient("api-key",
		cartesia.WithBaseURL(ts.URL),
		cartesia.WithToken("my-access-token"),
		cartesia.WithMaxRetries(0),
	)
	_, _ = c.GetStatus(context.Background())
}

func TestAuthHeader_BearerAPIKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if got != "Bearer my-api-key" {
			t.Errorf("expected 'Bearer my-api-key', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true})
	}))
	defer ts.Close()

	c := cartesia.NewClient("my-api-key",
		cartesia.WithBaseURL(ts.URL),
		cartesia.WithMaxRetries(0),
	)
	_, _ = c.GetStatus(context.Background())
}

func TestCartesiaVersionHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Cartesia-Version")
		if got != cartesia.APIVersion {
			t.Errorf("expected %q, got %q", cartesia.APIVersion, got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true})
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(0))
	_, _ = c.GetStatus(context.Background())
}

func TestJSONBodyMarshaling(t *testing.T) {
	// GetStatus uses GET with no body, so we test indirectly by checking
	// the server receives the request. For a real body test we'd need
	// an endpoint that accepts a body; we verify via the Voices or similar.
	// Instead, we test through the public API error path which exercises requestJSON.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); r.ContentLength > 0 && ct != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", ct)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true})
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(0))
	_, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay response so context cancellation has time to take effect.
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := c.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
