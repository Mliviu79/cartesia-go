package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestAgentsCallsRetrieve(t *testing.T) {
	startTime := "2025-01-01T00:00:00Z"
	endTime := "2025-01-01T00:05:00Z"
	summary := "Test call summary"
	deploymentID := "deploy-1"

	want := cartesia.AgentCall{
		ID:           "call-1",
		AgentID:      "agent-1",
		Status:       "completed",
		DeploymentID: &deploymentID,
		StartTime:    &startTime,
		EndTime:      &endTime,
		Summary:      &summary,
		TelephonyParams: &cartesia.TelephonyParams{
			From: "+15551111111",
			To:   "+15552222222",
		},
		Transcript: []cartesia.AgentTranscript{
			{
				Role:           "agent",
				StartTimestamp: 0.0,
				EndTimestamp:   2.5,
				TextChunks: []cartesia.TextChunk{
					{Text: "Hello", StartTimestamp: 0.0},
				},
			},
		},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/calls/call-1" {
			t.Errorf("expected path /agents/calls/call-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))

	got, err := client.Agents.Calls.Retrieve(context.Background(), "call-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "call-1" {
		t.Errorf("expected ID call-1, got %s", got.ID)
	}
	if got.AgentID != "agent-1" {
		t.Errorf("expected AgentID agent-1, got %s", got.AgentID)
	}
	if got.Status != "completed" {
		t.Errorf("expected Status completed, got %s", got.Status)
	}
	if got.DeploymentID == nil || *got.DeploymentID != "deploy-1" {
		t.Errorf("expected DeploymentID deploy-1, got %v", got.DeploymentID)
	}
	if got.TelephonyParams == nil {
		t.Fatal("expected TelephonyParams to be set")
	}
	if got.TelephonyParams.From != "+15551111111" {
		t.Errorf("expected From +15551111111, got %s", got.TelephonyParams.From)
	}
	if len(got.Transcript) != 1 {
		t.Fatalf("expected 1 transcript entry, got %d", len(got.Transcript))
	}
	if got.Transcript[0].Role != "agent" {
		t.Errorf("expected role agent, got %s", got.Transcript[0].Role)
	}
}

func TestAgentsCallsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		// Query params are appended to the path by appendQuery before URL construction,
		// so we parse them from the full request URI.
		fullURL := r.RequestURI
		if !strings.HasPrefix(fullURL, "/agents/calls") {
			t.Errorf("expected path to start with /agents/calls, got %s", fullURL)
		}
		if !strings.Contains(fullURL, "agent_id=agent-1") {
			t.Errorf("expected agent_id=agent-1 in URL, got %s", fullURL)
		}
		if !strings.Contains(fullURL, "limit=10") {
			t.Errorf("expected limit=10 in URL, got %s", fullURL)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.AgentCall{
				{ID: "call-1", AgentID: "agent-1", Status: "completed"},
				{ID: "call-2", AgentID: "agent-1", Status: "in_progress"},
			},
			"has_more": false,
		})
	}))

	got, err := client.Agents.Calls.List(context.Background(), cartesia.AgentCallsListParams{
		AgentID: "agent-1",
		Limit:   cartesia.Int(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(got.Data))
	}
	if got.Data[0].ID != "call-1" {
		t.Errorf("expected first call ID call-1, got %s", got.Data[0].ID)
	}
	if got.HasMore {
		t.Error("expected HasMore to be false")
	}
}

func TestAgentsCallsDownloadAudio(t *testing.T) {
	audioBytes := []byte("fake-audio-data-wav")

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/calls/call-1/audio" {
			t.Errorf("expected path /agents/calls/call-1/audio, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioBytes)
	}))

	got, err := client.Agents.Calls.DownloadAudio(context.Background(), "call-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Errorf("expected %q, got %q", string(audioBytes), string(got))
	}
}
