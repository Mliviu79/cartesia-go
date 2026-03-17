package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestPronunciationDictsCreate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/pronunciation-dicts/" {
			t.Errorf("expected path /pronunciation-dicts/, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.PronunciationDictCreateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name != "Tech Terms" {
			t.Errorf("expected name 'Tech Terms', got %s", body.Name)
		}
		if len(body.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(body.Items))
		}
		if body.Items[0].Text != "API" || body.Items[0].Alias != "A P I" {
			t.Errorf("unexpected first item: %+v", body.Items[0])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.PronunciationDict{
			ID:        "pd-1",
			Name:      body.Name,
			Items:     body.Items,
			OwnerID:   "user-1",
			Pinned:    false,
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.PronunciationDicts.Create(context.Background(), cartesia.PronunciationDictCreateParams{
		Name: "Tech Terms",
		Items: []cartesia.PronunciationDictItem{
			{Text: "API", Alias: "A P I"},
			{Text: "SDK", Alias: "S D K"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "pd-1" {
		t.Errorf("expected ID pd-1, got %s", got.ID)
	}
	if len(got.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(got.Items))
	}
}

func TestPronunciationDictsRetrieve(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/pronunciation-dicts/pd-1" {
			t.Errorf("expected path /pronunciation-dicts/pd-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.PronunciationDict{
			ID:   "pd-1",
			Name: "Tech Terms",
			Items: []cartesia.PronunciationDictItem{
				{Text: "API", Alias: "A P I"},
			},
			OwnerID:   "user-1",
			Pinned:    true,
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.PronunciationDicts.Retrieve(context.Background(), "pd-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "pd-1" {
		t.Errorf("expected ID pd-1, got %s", got.ID)
	}
	if got.OwnerID != "user-1" {
		t.Errorf("expected OwnerID user-1, got %s", got.OwnerID)
	}
	if !got.Pinned {
		t.Error("expected Pinned to be true")
	}
	if len(got.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(got.Items))
	}
}

func TestPronunciationDictsUpdate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/pronunciation-dicts/pd-1" {
			t.Errorf("expected path /pronunciation-dicts/pd-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.PronunciationDictUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name == nil || *body.Name != "Updated Dict" {
			t.Errorf("expected name 'Updated Dict', got %v", body.Name)
		}
		if len(body.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(body.Items))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.PronunciationDict{
			ID:        "pd-1",
			Name:      *body.Name,
			Items:     body.Items,
			OwnerID:   "user-1",
			Pinned:    false,
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.PronunciationDicts.Update(context.Background(), "pd-1", cartesia.PronunciationDictUpdateParams{
		Name: cartesia.String("Updated Dict"),
		Items: []cartesia.PronunciationDictItem{
			{Text: "CLI", Alias: "C L I"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Dict" {
		t.Errorf("expected name 'Updated Dict', got %s", got.Name)
	}
}

func TestPronunciationDictsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/pronunciation-dicts/") {
			t.Errorf("expected path to start with /pronunciation-dicts/, got %s", uri)
		}
		if !strings.Contains(uri, "limit=10") {
			t.Errorf("expected limit=10 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.PronunciationDict{
				{ID: "pd-1", Name: "Dict One", OwnerID: "user-1", CreatedAt: "2025-01-01T00:00:00Z"},
				{ID: "pd-2", Name: "Dict Two", OwnerID: "user-1", CreatedAt: "2025-01-02T00:00:00Z"},
			},
			"has_more": false,
		})
	}))

	got, err := client.PronunciationDicts.List(context.Background(), &cartesia.ListParams{
		Limit: cartesia.Int(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 dicts, got %d", len(got.Data))
	}
	if got.HasMore {
		t.Error("expected HasMore to be false")
	}
}

func TestPronunciationDictsDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/pronunciation-dicts/pd-1" {
			t.Errorf("expected path /pronunciation-dicts/pd-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.PronunciationDicts.Delete(context.Background(), "pd-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
