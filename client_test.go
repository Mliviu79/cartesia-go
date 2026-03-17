package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func TestNewClient_Defaults(t *testing.T) {
	c := cartesia.NewClient("test-key")
	if c == nil {
		t.Fatal("expected non-nil client")
	}

	// Verify services are initialized.
	if c.TTS == nil {
		t.Error("expected TTS service to be initialized")
	}
	if c.Voices == nil {
		t.Error("expected Voices service to be initialized")
	}
	if c.Agents == nil {
		t.Error("expected Agents service to be initialized")
	}
	if c.Datasets == nil {
		t.Error("expected Datasets service to be initialized")
	}
	if c.STT == nil {
		t.Error("expected STT service to be initialized")
	}

	// We can verify defaults by making a request and inspecting headers.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify version header uses default.
		if got := r.Header.Get("Cartesia-Version"); got != cartesia.APIVersion {
			t.Errorf("expected Cartesia-Version %q, got %q", cartesia.APIVersion, got)
		}
		// Verify auth header uses apiKey.
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Authorization 'Bearer test-key', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true, Version: "1.0"})
	}))
	defer ts.Close()

	c2 := cartesia.NewClient("test-key", cartesia.WithBaseURL(ts.URL))
	_, err := c2.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := noop.NewTracerProvider().Tracer("test")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token takes precedence.
		if got := r.Header.Get("Authorization"); got != "Bearer my-token" {
			t.Errorf("expected Authorization 'Bearer my-token', got %q", got)
		}
		// Verify custom version.
		if got := r.Header.Get("Cartesia-Version"); got != "2025-01-01" {
			t.Errorf("expected Cartesia-Version '2025-01-01', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(cartesia.GetStatusResponse{OK: true, Version: "1.0"})
	}))
	defer ts.Close()

	c := cartesia.NewClient("api-key",
		cartesia.WithBaseURL(ts.URL),
		cartesia.WithMaxRetries(5),
		cartesia.WithToken("my-token"),
		cartesia.WithLogger(logger),
		cartesia.WithTracer(tracer),
		cartesia.WithVersion("2025-01-01"),
	)

	resp, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Error("expected OK to be true")
	}
	if resp.Version != "1.0" {
		t.Errorf("expected Version '1.0', got %q", resp.Version)
	}
}

func TestGetStatus_ParsesResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/" {
			t.Errorf("expected path '/', got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"version": "2.5.0",
		})
	}))
	defer ts.Close()

	c := cartesia.NewClient("key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(0))
	resp, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Error("expected OK=true")
	}
	if resp.Version != "2.5.0" {
		t.Errorf("expected version '2.5.0', got %q", resp.Version)
	}
}

func TestGetStatus_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer ts.Close()

	c := cartesia.NewClient("bad-key", cartesia.WithBaseURL(ts.URL), cartesia.WithMaxRetries(0))
	_, err := c.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *cartesia.APIError
	if !cartesia.IsForbidden(err) {
		t.Error("expected IsForbidden to be true")
	}

	// Use errors.As pattern.
	if ok := isAPIError(err, &apiErr); !ok {
		t.Fatal("expected *APIError")
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "access denied" {
		t.Errorf("expected message 'access denied', got %q", apiErr.Message)
	}
}

// isAPIError is a test helper that uses errors.As.
func isAPIError(err error, target **cartesia.APIError) bool {
	return err != nil && func() bool {
		var ae *cartesia.APIError
		ok := false
		// Use a type switch for simplicity.
		for e := err; e != nil; {
			if a, ok2 := e.(*cartesia.APIError); ok2 {
				*target = a
				return true
			}
			if u, ok2 := e.(interface{ Unwrap() error }); ok2 {
				e = u.Unwrap()
			} else {
				break
			}
		}
		_ = ae
		_ = ok
		return false
	}()
}
