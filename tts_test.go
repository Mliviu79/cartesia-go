package cartesia_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestTTSGenerate(t *testing.T) {
	audioBytes := []byte("fake-wav-audio-bytes")

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tts/bytes" {
			t.Errorf("expected path /tts/bytes, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body cartesia.TTSRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.ModelID != "sonic-2" {
			t.Errorf("expected model_id sonic-2, got %s", body.ModelID)
		}
		if body.Transcript != "Hello, world!" {
			t.Errorf("expected transcript 'Hello, world!', got %s", body.Transcript)
		}
		if body.Voice.Mode != "id" || body.Voice.ID != "voice-1" {
			t.Errorf("unexpected voice: %+v", body.Voice)
		}
		if body.OutputFormat.Container != "wav" {
			t.Errorf("expected container wav, got %s", body.OutputFormat.Container)
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioBytes)
	}))

	got, err := client.TTS.Generate(context.Background(), cartesia.TTSRequest{
		ModelID:    "sonic-2",
		Transcript: "Hello, world!",
		Voice:      cartesia.VoiceSpecifier{Mode: "id", ID: "voice-1"},
		OutputFormat: cartesia.OutputFormat{
			Container:  "wav",
			Encoding:   "pcm_s16le",
			SampleRate: 44100,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Errorf("expected %q, got %q", string(audioBytes), string(got))
	}
}

func TestTTSGenerateSSE(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/tts/sse" {
			t.Errorf("expected path /tts/sse, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.TTSRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.ModelID != "sonic-2" {
			t.Errorf("expected model_id sonic-2, got %s", body.ModelID)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected ResponseWriter to be a Flusher")
		}

		fmt.Fprintf(w, "event: audio\ndata: {\"audio\":\"AQID\"}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: done\ndata: {}\n\n")
		flusher.Flush()
	}))

	reader, err := client.TTS.GenerateSSE(context.Background(), cartesia.TTSRequest{
		ModelID:    "sonic-2",
		Transcript: "Hello!",
		Voice:      cartesia.VoiceSpecifier{Mode: "id", ID: "voice-1"},
		OutputFormat: cartesia.OutputFormat{
			Container:  "raw",
			Encoding:   "pcm_s16le",
			SampleRate: 44100,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	event, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error reading first event: %v", err)
	}
	if event.Event != "audio" {
		t.Errorf("expected event type audio, got %s", event.Event)
	}

	event, err = reader.Next()
	if err != nil {
		t.Fatalf("unexpected error reading second event: %v", err)
	}
	if event.Event != "done" {
		t.Errorf("expected event type done, got %s", event.Event)
	}

	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestTTSInfill(t *testing.T) {
	audioBytes := []byte("infilled-audio-data")

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/infill/bytes" {
			t.Errorf("expected path /infill/bytes, got %s", r.URL.Path)
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

		leftFile, leftHeader, err := r.FormFile("left_audio")
		if err != nil {
			t.Fatalf("expected left_audio field: %v", err)
		}
		defer leftFile.Close()
		if leftHeader.Filename != "left.wav" {
			t.Errorf("expected left filename left.wav, got %s", leftHeader.Filename)
		}

		rightFile, rightHeader, err := r.FormFile("right_audio")
		if err != nil {
			t.Fatalf("expected right_audio field: %v", err)
		}
		defer rightFile.Close()
		if rightHeader.Filename != "right.wav" {
			t.Errorf("expected right filename right.wav, got %s", rightHeader.Filename)
		}

		transcript := r.FormValue("transcript")
		if transcript != "and then" {
			t.Errorf("expected transcript 'and then', got %s", transcript)
		}

		voiceID := r.FormValue("voice_id")
		if voiceID != "voice-1" {
			t.Errorf("expected voice_id voice-1, got %s", voiceID)
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioBytes)
	}))

	got, err := client.TTS.Infill(context.Background(), cartesia.TTSInfillParams{
		LeftAudio: cartesia.FileParam{
			Reader:   strings.NewReader("left audio data"),
			FileName: "left.wav",
		},
		RightAudio: cartesia.FileParam{
			Reader:   strings.NewReader("right audio data"),
			FileName: "right.wav",
		},
		Transcript: "and then",
		VoiceID:    "voice-1",
		ModelID:    "sonic-2",
		Language:   "en",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Errorf("expected %q, got %q", string(audioBytes), string(got))
	}
}
