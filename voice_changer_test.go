package cartesia_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestVoiceChangerChangeVoiceBytes(t *testing.T) {
	audioBytes := []byte("changed-voice-audio")

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/voice-changer/bytes" {
			t.Errorf("expected path /voice-changer/bytes, got %s", r.URL.Path)
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

		clipFile, clipHeader, err := r.FormFile("clip")
		if err != nil {
			t.Fatalf("expected clip field: %v", err)
		}
		defer clipFile.Close()
		if clipHeader.Filename != "input.wav" {
			t.Errorf("expected clip filename input.wav, got %s", clipHeader.Filename)
		}

		voiceID := r.FormValue("voice[id]")
		if voiceID != "voice-1" {
			t.Errorf("expected voice[id] voice-1, got %s", voiceID)
		}

		container := r.FormValue("output_format[container]")
		if container != "wav" {
			t.Errorf("expected output_format[container] wav, got %s", container)
		}

		encoding := r.FormValue("output_format[encoding]")
		if encoding != "pcm_s16le" {
			t.Errorf("expected output_format[encoding] pcm_s16le, got %s", encoding)
		}

		sampleRate := r.FormValue("output_format[sample_rate]")
		if sampleRate != "44100" {
			t.Errorf("expected output_format[sample_rate] 44100, got %s", sampleRate)
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioBytes)
	}))

	got, err := client.VoiceChanger.ChangeVoiceBytes(context.Background(), cartesia.VoiceChangerParams{
		Clip: cartesia.FileParam{
			Reader:   strings.NewReader("input audio data"),
			FileName: "input.wav",
		},
		VoiceID: "voice-1",
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

func TestVoiceChangerChangeVoiceSSE(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/voice-changer/sse" {
			t.Errorf("expected path /voice-changer/sse, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", contentType)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected ResponseWriter to be a Flusher")
		}

		fmt.Fprintf(w, "event: audio\ndata: {\"audio\":\"chunk1\"}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "event: done\ndata: {}\n\n")
		flusher.Flush()
	}))

	reader, err := client.VoiceChanger.ChangeVoiceSSE(context.Background(), cartesia.VoiceChangerParams{
		Clip: cartesia.FileParam{
			Reader:   strings.NewReader("input audio data"),
			FileName: "input.wav",
		},
		VoiceID: "voice-1",
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
