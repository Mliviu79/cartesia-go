package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestSTTTranscribe(t *testing.T) {
	duration := 5.2
	lang := "en"

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/stt" {
			t.Errorf("expected path /stt, got %s", r.URL.Path)
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

		language := r.FormValue("language")
		if language != "en" {
			t.Errorf("expected language en, got %s", language)
		}

		model := r.FormValue("model")
		if model != "ink" {
			t.Errorf("expected model ink, got %s", model)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.STTTranscribeResponse{
			Text:     "Hello world",
			Duration: &duration,
			Language: &lang,
			Words: []cartesia.STTWord{
				{Word: "Hello", Start: 0.0, End: 0.5},
				{Word: "world", Start: 0.6, End: 1.0},
			},
		})
	}))

	got, err := client.STT.Transcribe(context.Background(), cartesia.STTTranscribeParams{
		File: cartesia.FileParam{
			Reader:   strings.NewReader("fake audio data"),
			FileName: "test.wav",
		},
		Language: "en",
		Model:    "ink",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "Hello world" {
		t.Errorf("expected text 'Hello world', got %s", got.Text)
	}
	if got.Duration == nil || *got.Duration != 5.2 {
		t.Errorf("expected duration 5.2, got %v", got.Duration)
	}
	if got.Language == nil || *got.Language != "en" {
		t.Errorf("expected language en, got %v", got.Language)
	}
	if len(got.Words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(got.Words))
	}
	if got.Words[0].Word != "Hello" {
		t.Errorf("expected first word Hello, got %s", got.Words[0].Word)
	}
	if got.Words[1].End != 1.0 {
		t.Errorf("expected second word end 1.0, got %f", got.Words[1].End)
	}
}

func TestSTTTranscribe_WithQueryParams(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/stt") {
			t.Errorf("expected path to start with /stt, got %s", uri)
		}
		if !strings.Contains(uri, "encoding=pcm_s16le") {
			t.Errorf("expected encoding=pcm_s16le in URI, got %s", uri)
		}
		if !strings.Contains(uri, "sample_rate=44100") {
			t.Errorf("expected sample_rate=44100 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.STTTranscribeResponse{
			Text: "Test transcription",
		})
	}))

	got, err := client.STT.Transcribe(context.Background(), cartesia.STTTranscribeParams{
		File: cartesia.FileParam{
			Reader:   strings.NewReader("fake audio data"),
			FileName: "test.raw",
		},
		Language:   "en",
		Encoding:   "pcm_s16le",
		SampleRate: cartesia.Int(44100),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "Test transcription" {
		t.Errorf("expected text 'Test transcription', got %s", got.Text)
	}
}
