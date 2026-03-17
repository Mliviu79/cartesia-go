package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestAgentsDeploymentsRetrieve(t *testing.T) {
	want := cartesia.Deployment{
		ID:                    "deploy-1",
		AgentID:               "agent-1",
		BuildCompletedAt:      "2025-01-01T00:01:00Z",
		BuildLogs:             "build ok",
		BuildStartedAt:        "2025-01-01T00:00:00Z",
		CreatedAt:             "2025-01-01T00:00:00Z",
		DeploymentCompletedAt: "2025-01-01T00:02:00Z",
		DeploymentStartedAt:   "2025-01-01T00:01:30Z",
		EnvVarCollectionID:    "env-1",
		GitCommitHash:         "abc123",
		IsLive:                true,
		IsPinned:              false,
		SourceCodeFileID:      "file-1",
		Status:                "live",
		UpdatedAt:             "2025-01-01T00:02:00Z",
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/deployments/deploy-1" {
			t.Errorf("expected path /agents/deployments/deploy-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))

	got, err := client.Agents.Deployments.Retrieve(context.Background(), "deploy-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("expected ID %s, got %s", want.ID, got.ID)
	}
	if got.AgentID != want.AgentID {
		t.Errorf("expected AgentID %s, got %s", want.AgentID, got.AgentID)
	}
	if !got.IsLive {
		t.Error("expected IsLive to be true")
	}
	if got.Status != "live" {
		t.Errorf("expected Status live, got %s", got.Status)
	}
	if got.GitCommitHash != "abc123" {
		t.Errorf("expected GitCommitHash abc123, got %s", got.GitCommitHash)
	}
	if got.EnvVarCollectionID != "env-1" {
		t.Errorf("expected EnvVarCollectionID env-1, got %s", got.EnvVarCollectionID)
	}
}

func TestAgentsDeploymentsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1/deployments" {
			t.Errorf("expected path /agents/agent-1/deployments, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]cartesia.Deployment{
			{ID: "deploy-1", AgentID: "agent-1", Status: "live", IsLive: true},
			{ID: "deploy-2", AgentID: "agent-1", Status: "stopped", IsLive: false},
		})
	}))

	got, err := client.Agents.Deployments.List(context.Background(), "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 deployments, got %d", len(got))
	}
	if got[0].ID != "deploy-1" {
		t.Errorf("expected first deployment ID deploy-1, got %s", got[0].ID)
	}
	if got[1].Status != "stopped" {
		t.Errorf("expected second deployment status stopped, got %s", got[1].Status)
	}
}
