package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestDatasetsFilesList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1/files" {
			t.Errorf("expected path /datasets/ds-1/files, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.DatasetFile{
				{ID: "file-1", Filename: "audio.wav", Size: 1024, CreatedAt: "2025-01-01T00:00:00Z"},
				{ID: "file-2", Filename: "transcript.txt", Size: 256, CreatedAt: "2025-01-02T00:00:00Z"},
			},
			"has_more": false,
		})
	}))

	got, err := client.Datasets.Files.List(context.Background(), "ds-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 files, got %d", len(got.Data))
	}
	if got.Data[0].Filename != "audio.wav" {
		t.Errorf("expected filename audio.wav, got %s", got.Data[0].Filename)
	}
	if got.Data[0].Size != 1024 {
		t.Errorf("expected size 1024, got %d", got.Data[0].Size)
	}
}

func TestDatasetsFilesDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1/files/file-1" {
			t.Errorf("expected path /datasets/ds-1/files/file-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Datasets.Files.Delete(context.Background(), "ds-1", "file-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatasetsFilesUpload(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/datasets/ds-1/files" {
			t.Errorf("expected path /datasets/ds-1/files, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", contentType)
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("expected file field: %v", err)
		}
		defer file.Close()

		if header.Filename != "test.wav" {
			t.Errorf("expected filename test.wav, got %s", header.Filename)
		}

		purpose := r.FormValue("purpose")
		if purpose != "fine-tune" {
			t.Errorf("expected purpose fine-tune, got %s", purpose)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))

	err := client.Datasets.Files.Upload(context.Background(), "ds-1", cartesia.FileUploadParams{
		File:     strings.NewReader("fake audio content"),
		FileName: "test.wav",
		Purpose:  "fine-tune",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
