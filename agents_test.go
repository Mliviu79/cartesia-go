package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestAgentsRetrieve(t *testing.T) {
	want := cartesia.AgentSummary{
		ID:              "agent-1",
		Name:            "Test Agent",
		CreatedAt:       "2025-01-01T00:00:00Z",
		UpdatedAt:       "2025-01-02T00:00:00Z",
		DeploymentCount: 3,
		TTSLanguage:     "en",
		TTSVoice:        "voice-1",
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1" {
			t.Errorf("expected path /agents/agent-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))

	got, err := client.Agents.Retrieve(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("expected ID %s, got %s", want.ID, got.ID)
	}
	if got.Name != want.Name {
		t.Errorf("expected Name %s, got %s", want.Name, got.Name)
	}
	if got.DeploymentCount != want.DeploymentCount {
		t.Errorf("expected DeploymentCount %d, got %d", want.DeploymentCount, got.DeploymentCount)
	}
	if got.TTSLanguage != want.TTSLanguage {
		t.Errorf("expected TTSLanguage %s, got %s", want.TTSLanguage, got.TTSLanguage)
	}
	if got.TTSVoice != want.TTSVoice {
		t.Errorf("expected TTSVoice %s, got %s", want.TTSVoice, got.TTSVoice)
	}
}

func TestAgentsUpdate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1" {
			t.Errorf("expected path /agents/agent-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.AgentUpdateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name == nil || *body.Name != "Updated Agent" {
			t.Errorf("expected name Updated Agent, got %v", body.Name)
		}
		if body.TTSVoice == nil || *body.TTSVoice != "voice-2" {
			t.Errorf("expected tts_voice voice-2, got %v", body.TTSVoice)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.AgentSummary{
			ID:   "agent-1",
			Name: "Updated Agent",
		})
	}))

	got, err := client.Agents.Update(context.Background(), "agent-1", cartesia.AgentUpdateParams{
		Name:     cartesia.String("Updated Agent"),
		TTSVoice: cartesia.String("voice-2"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Agent" {
		t.Errorf("expected name Updated Agent, got %s", got.Name)
	}
}

func TestAgentsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents" {
			t.Errorf("expected path /agents, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.AgentListResponse{
			Summaries: []cartesia.AgentSummary{
				{ID: "agent-1", Name: "Agent One"},
				{ID: "agent-2", Name: "Agent Two"},
			},
		})
	}))

	got, err := client.Agents.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(got.Summaries))
	}
	if got.Summaries[0].ID != "agent-1" {
		t.Errorf("expected first agent ID agent-1, got %s", got.Summaries[0].ID)
	}
	if got.Summaries[1].ID != "agent-2" {
		t.Errorf("expected second agent ID agent-2, got %s", got.Summaries[1].ID)
	}
}

func TestAgentsDelete(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1" {
			t.Errorf("expected path /agents/agent-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Agents.Delete(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsListPhoneNumbers(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1/phone-numbers" {
			t.Errorf("expected path /agents/agent-1/phone-numbers, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]cartesia.AgentPhoneNumberDetail{
			{
				AgentID:           "agent-1",
				Number:            "+15551234567",
				CreatedAt:         "2025-01-01T00:00:00Z",
				UpdatedAt:         "2025-01-01T00:00:00Z",
				IsCartesiaManaged: true,
			},
		})
	}))

	got, err := client.Agents.ListPhoneNumbers(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 phone number, got %d", len(got))
	}
	if got[0].Number != "+15551234567" {
		t.Errorf("expected number +15551234567, got %s", got[0].Number)
	}
	if !got[0].IsCartesiaManaged {
		t.Error("expected IsCartesiaManaged to be true")
	}
}

func TestAgentsListTemplates(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/templates" {
			t.Errorf("expected path /agents/templates, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.AgentListTemplatesResponse{
			Templates: []cartesia.AgentTemplate{
				{
					ID:        "tmpl-1",
					Name:      "Basic Template",
					OwnerID:   "owner-1",
					RepoURL:   "https://github.com/example/repo",
					RootDir:   "/",
					CreatedAt: "2025-01-01T00:00:00Z",
					UpdatedAt: "2025-01-01T00:00:00Z",
				},
			},
		})
	}))

	got, err := client.Agents.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(got.Templates))
	}
	if got.Templates[0].Name != "Basic Template" {
		t.Errorf("expected name Basic Template, got %s", got.Templates[0].Name)
	}
	if got.Templates[0].RepoURL != "https://github.com/example/repo" {
		t.Errorf("expected repo URL https://github.com/example/repo, got %s", got.Templates[0].RepoURL)
	}
}
