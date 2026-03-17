package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestFineTunesCreate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/fine-tunes/" {
			t.Errorf("expected path /fine-tunes/, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.FineTuneCreateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Dataset != "ds-1" {
			t.Errorf("expected dataset ds-1, got %s", body.Dataset)
		}
		if body.Description != "Fine tune description" {
			t.Errorf("expected description, got %s", body.Description)
		}
		if body.Language != "en" {
			t.Errorf("expected language en, got %s", body.Language)
		}
		if body.ModelID != "sonic-2" {
			t.Errorf("expected model_id sonic-2, got %s", body.ModelID)
		}
		if body.Name != "My Fine Tune" {
			t.Errorf("expected name 'My Fine Tune', got %s", body.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.FineTune{
			ID:          "ft-1",
			Dataset:     body.Dataset,
			Description: body.Description,
			Language:    body.Language,
			ModelID:     body.ModelID,
			Name:        body.Name,
			Status:      "pending",
		})
	}))

	got, err := client.FineTunes.Create(context.Background(), cartesia.FineTuneCreateParams{
		Dataset:     "ds-1",
		Description: "Fine tune description",
		Language:    "en",
		ModelID:     "sonic-2",
		Name:        "My Fine Tune",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "ft-1" {
		t.Errorf("expected ID ft-1, got %s", got.ID)
	}
	if got.Status != "pending" {
		t.Errorf("expected status pending, got %s", got.Status)
	}
}

func TestFineTunesRetrieve(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fine-tunes/ft-1" {
			t.Errorf("expected path /fine-tunes/ft-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.FineTune{
			ID:       "ft-1",
			Name:     "My Fine Tune",
			Status:   "completed",
			Dataset:  "ds-1",
			Language: "en",
			ModelID:  "sonic-2",
		})
	}))

	got, err := client.FineTunes.Retrieve(context.Background(), "ft-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "ft-1" {
		t.Errorf("expected ID ft-1, got %s", got.ID)
	}
	if got.Status != "completed" {
		t.Errorf("expected status completed, got %s", got.Status)
	}
}

func TestFineTunesList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/fine-tunes/") {
			t.Errorf("expected path to start with /fine-tunes/, got %s", uri)
		}
		if !strings.Contains(uri, "limit=5") {
			t.Errorf("expected limit=5 in URI, got %s", uri)
		}
		if !strings.Contains(uri, "starting_after=ft-0") {
			t.Errorf("expected starting_after=ft-0 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.FineTune{
				{ID: "ft-1", Name: "FT One", Status: "completed"},
			},
			"has_more": false,
		})
	}))

	got, err := client.FineTunes.List(context.Background(), &cartesia.ListParams{
		Limit:         cartesia.Int(5),
		StartingAfter: cartesia.String("ft-0"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 fine tune, got %d", len(got.Data))
	}
	if got.Data[0].ID != "ft-1" {
		t.Errorf("expected ID ft-1, got %s", got.Data[0].ID)
	}
}

func TestFineTunesDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/fine-tunes/ft-1" {
			t.Errorf("expected path /fine-tunes/ft-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.FineTunes.Delete(context.Background(), "ft-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFineTunesListVoices(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/fine-tunes/ft-1/voices" {
			t.Errorf("expected path /fine-tunes/ft-1/voices, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.Voice{
				{ID: "voice-1", Name: "Fine Tuned Voice", Language: "en", IsOwner: true},
			},
			"has_more": false,
		})
	}))

	got, err := client.FineTunes.ListVoices(context.Background(), "ft-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 voice, got %d", len(got.Data))
	}
	if got.Data[0].Name != "Fine Tuned Voice" {
		t.Errorf("expected name 'Fine Tuned Voice', got %s", got.Data[0].Name)
	}
}
