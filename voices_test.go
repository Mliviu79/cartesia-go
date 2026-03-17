package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestVoicesUpdate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/voices/voice-1" {
			t.Errorf("expected path /voices/voice-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.VoiceUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name != "Updated Voice" {
			t.Errorf("expected name 'Updated Voice', got %s", body.Name)
		}
		if body.Gender == nil || *body.Gender != "female" {
			t.Errorf("expected gender female, got %v", body.Gender)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Voice{
			ID:       "voice-1",
			Name:     body.Name,
			Gender:   body.Gender,
			Language: "en",
		})
	}))

	got, err := client.Voices.Update(context.Background(), "voice-1", cartesia.VoiceUpdateParams{
		Name:        "Updated Voice",
		Description: "A female voice",
		Gender:      cartesia.String("female"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Voice" {
		t.Errorf("expected name 'Updated Voice', got %s", got.Name)
	}
	if got.Gender == nil || *got.Gender != "female" {
		t.Errorf("expected gender female, got %v", got.Gender)
	}
}

func TestVoicesList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/voices") {
			t.Errorf("expected path to start with /voices, got %s", uri)
		}
		if !strings.Contains(uri, "gender=female") {
			t.Errorf("expected gender=female in URI, got %s", uri)
		}
		if !strings.Contains(uri, "is_owner=true") {
			t.Errorf("expected is_owner=true in URI, got %s", uri)
		}
		if !strings.Contains(uri, "q=search") {
			t.Errorf("expected q=search in URI, got %s", uri)
		}
		if !strings.Contains(uri, "limit=20") {
			t.Errorf("expected limit=20 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.Voice{
				{ID: "voice-1", Name: "Voice One", Language: "en", IsOwner: true, Gender: cartesia.String("female")},
			},
			"has_more": false,
		})
	}))

	got, err := client.Voices.List(context.Background(), &cartesia.VoicesListParams{
		Gender:  cartesia.String("female"),
		IsOwner: cartesia.Bool(true),
		Q:       cartesia.String("search term"),
		Limit:   cartesia.Int(20),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 voice, got %d", len(got.Data))
	}
	if got.Data[0].ID != "voice-1" {
		t.Errorf("expected ID voice-1, got %s", got.Data[0].ID)
	}
}

func TestVoicesDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/voices/voice-1" {
			t.Errorf("expected path /voices/voice-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Voices.Delete(context.Background(), "voice-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVoicesClone(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/voices/clone" {
			t.Errorf("expected path /voices/clone, got %s", r.URL.Path)
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
		if clipHeader.Filename != "sample.wav" {
			t.Errorf("expected clip filename sample.wav, got %s", clipHeader.Filename)
		}

		name := r.FormValue("name")
		if name != "Cloned Voice" {
			t.Errorf("expected name 'Cloned Voice', got %s", name)
		}
		description := r.FormValue("description")
		if description != "A cloned voice" {
			t.Errorf("expected description 'A cloned voice', got %s", description)
		}
		language := r.FormValue("language")
		if language != "en" {
			t.Errorf("expected language en, got %s", language)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.VoiceMetadata{
			ID:          "voice-new",
			Name:        "Cloned Voice",
			Description: "A cloned voice",
			Language:    "en",
			UserID:      "user-1",
			CreatedAt:   "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Voices.Clone(context.Background(), cartesia.VoiceCloneParams{
		Clip: cartesia.FileParam{
			Reader:   strings.NewReader("fake audio clip"),
			FileName: "sample.wav",
		},
		Name:        "Cloned Voice",
		Description: "A cloned voice",
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "voice-new" {
		t.Errorf("expected ID voice-new, got %s", got.ID)
	}
	if got.Name != "Cloned Voice" {
		t.Errorf("expected name 'Cloned Voice', got %s", got.Name)
	}
	if got.UserID != "user-1" {
		t.Errorf("expected UserID user-1, got %s", got.UserID)
	}
}

func TestVoicesGet(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/voices/voice-1") {
			t.Errorf("expected path to start with /voices/voice-1, got %s", uri)
		}
		if !strings.Contains(uri, "expand=embedding") {
			t.Errorf("expected expand=embedding in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Voice{
			ID:       "voice-1",
			Name:     "My Voice",
			Language: "en",
			IsOwner:  true,
			IsPublic: false,
		})
	}))

	got, err := client.Voices.Get(context.Background(), "voice-1", &cartesia.VoiceGetParams{
		Expand: []string{"embedding"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "voice-1" {
		t.Errorf("expected ID voice-1, got %s", got.ID)
	}
	if got.Name != "My Voice" {
		t.Errorf("expected name 'My Voice', got %s", got.Name)
	}
	if !got.IsOwner {
		t.Error("expected IsOwner to be true")
	}
}

func TestVoicesLocalize(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/voices/localize" {
			t.Errorf("expected path /voices/localize, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.VoiceLocalizeParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.VoiceID != "voice-1" {
			t.Errorf("expected voice_id voice-1, got %s", body.VoiceID)
		}
		if body.Language != "es" {
			t.Errorf("expected language es, got %s", body.Language)
		}
		if body.Name != "Localized Voice" {
			t.Errorf("expected name 'Localized Voice', got %s", body.Name)
		}
		if body.Description != "Spanish version" {
			t.Errorf("expected description 'Spanish version', got %s", body.Description)
		}
		if body.OriginalSpeakerGender != "male" {
			t.Errorf("expected original_speaker_gender male, got %s", body.OriginalSpeakerGender)
		}
		if body.Dialect == nil || *body.Dialect != "castilian" {
			t.Errorf("expected dialect castilian, got %v", body.Dialect)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.VoiceMetadata{
			ID:          "voice-localized",
			Name:        body.Name,
			Description: body.Description,
			Language:    body.Language,
			UserID:      "user-1",
			CreatedAt:   "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Voices.Localize(context.Background(), cartesia.VoiceLocalizeParams{
		VoiceID:               "voice-1",
		Language:              "es",
		Name:                  "Localized Voice",
		Description:           "Spanish version",
		OriginalSpeakerGender: "male",
		Dialect:               cartesia.String("castilian"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "voice-localized" {
		t.Errorf("expected ID voice-localized, got %s", got.ID)
	}
	if got.Language != "es" {
		t.Errorf("expected language es, got %s", got.Language)
	}
	if got.Name != "Localized Voice" {
		t.Errorf("expected name 'Localized Voice', got %s", got.Name)
	}
}
