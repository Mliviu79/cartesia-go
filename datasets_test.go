package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestDatasetsCreate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/" {
			t.Errorf("expected path /datasets/, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.DatasetCreateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name != "My Dataset" {
			t.Errorf("expected name 'My Dataset', got %s", body.Name)
		}
		if body.Description != "Test description" {
			t.Errorf("expected description 'Test description', got %s", body.Description)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Dataset{
			ID:          "ds-1",
			Name:        body.Name,
			Description: body.Description,
			CreatedAt:   "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Datasets.Create(context.Background(), cartesia.DatasetCreateParams{
		Name:        "My Dataset",
		Description: "Test description",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "ds-1" {
		t.Errorf("expected ID ds-1, got %s", got.ID)
	}
	if got.Name != "My Dataset" {
		t.Errorf("expected Name 'My Dataset', got %s", got.Name)
	}
}

func TestDatasetsRetrieve(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1" {
			t.Errorf("expected path /datasets/ds-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Dataset{
			ID:          "ds-1",
			Name:        "My Dataset",
			Description: "A dataset",
			CreatedAt:   "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Datasets.Retrieve(context.Background(), "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "ds-1" {
		t.Errorf("expected ID ds-1, got %s", got.ID)
	}
	if got.Description != "A dataset" {
		t.Errorf("expected Description 'A dataset', got %s", got.Description)
	}
}

func TestDatasetsUpdate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1" {
			t.Errorf("expected path /datasets/ds-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.DatasetUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name == nil || *body.Name != "Updated Name" {
			t.Errorf("expected name 'Updated Name', got %v", body.Name)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Datasets.Update(context.Background(), "ds-1", cartesia.DatasetUpdateParams{
		Name: cartesia.String("Updated Name"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatasetsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/datasets/") {
			t.Errorf("expected path to start with /datasets/, got %s", uri)
		}
		if !strings.Contains(uri, "limit=10") {
			t.Errorf("expected limit=10 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.Dataset{
				{ID: "ds-1", Name: "Dataset One", Description: "First", CreatedAt: "2025-01-01T00:00:00Z"},
				{ID: "ds-2", Name: "Dataset Two", Description: "Second", CreatedAt: "2025-01-02T00:00:00Z"},
			},
			"has_more": true,
		})
	}))

	got, err := client.Datasets.List(context.Background(), &cartesia.ListParams{
		Limit: cartesia.Int(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 datasets, got %d", len(got.Data))
	}
	if !got.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestDatasetsDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1" {
			t.Errorf("expected path /datasets/ds-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Datasets.Delete(context.Background(), "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
