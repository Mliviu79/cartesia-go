package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func newTestClient(t *testing.T, handler http.Handler) *cartesia.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return cartesia.NewClient("test-key", cartesia.WithBaseURL(srv.URL), cartesia.WithMaxRetries(0))
}

func TestAccessTokenCreate_NoParams(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/access-token" {
			t.Errorf("expected path /access-token, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.AccessTokenCreateResponse{
			Token: "tok_abc123",
		})
	}))

	resp, err := client.AccessToken.Create(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "tok_abc123" {
		t.Errorf("expected token tok_abc123, got %s", resp.Token)
	}
}

func TestAccessTokenCreate_WithParams(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/access-token" {
			t.Errorf("expected path /access-token, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		var body cartesia.AccessTokenCreateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.ExpiresIn == nil || *body.ExpiresIn != 3600 {
			t.Errorf("expected expires_in=3600, got %v", body.ExpiresIn)
		}
		if body.Grants == nil {
			t.Fatal("expected grants to be set")
		}
		if body.Grants.TTS == nil || !*body.Grants.TTS {
			t.Error("expected grants.tts=true")
		}
		if body.Grants.STT == nil || !*body.Grants.STT {
			t.Error("expected grants.stt=true")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.AccessTokenCreateResponse{
			Token: "tok_with_params",
		})
	}))

	resp, err := client.AccessToken.Create(context.Background(), cartesia.AccessTokenCreateParams{
		ExpiresIn: cartesia.Int(3600),
		Grants: &cartesia.AccessTokenGrants{
			TTS: cartesia.Bool(true),
			STT: cartesia.Bool(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "tok_with_params" {
		t.Errorf("expected token tok_with_params, got %s", resp.Token)
	}
}
