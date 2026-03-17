package cartesia_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestAgentsMetricsCreate(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/agents/metrics" {
			t.Errorf("expected path /agents/metrics, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		var body cartesia.MetricCreateParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Name != "response_quality" {
			t.Errorf("expected name response_quality, got %s", body.Name)
		}
		if body.Prompt != "Rate the response quality" {
			t.Errorf("expected prompt 'Rate the response quality', got %s", body.Prompt)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Metric{
			ID:        "metric-1",
			Name:      body.Name,
			Prompt:    body.Prompt,
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Agents.Metrics.Create(context.Background(), cartesia.MetricCreateParams{
		Name:   "response_quality",
		Prompt: "Rate the response quality",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "metric-1" {
		t.Errorf("expected ID metric-1, got %s", got.ID)
	}
	if got.Name != "response_quality" {
		t.Errorf("expected Name response_quality, got %s", got.Name)
	}
}

func TestAgentsMetricsRetrieve(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/metrics/metric-1" {
			t.Errorf("expected path /agents/metrics/metric-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.Metric{
			ID:        "metric-1",
			Name:      "accuracy",
			Prompt:    "Evaluate accuracy",
			CreatedAt: "2025-01-01T00:00:00Z",
		})
	}))

	got, err := client.Agents.Metrics.Retrieve(context.Background(), "metric-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "metric-1" {
		t.Errorf("expected ID metric-1, got %s", got.ID)
	}
	if got.Prompt != "Evaluate accuracy" {
		t.Errorf("expected Prompt 'Evaluate accuracy', got %s", got.Prompt)
	}
}

func TestAgentsMetricsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/agents/metrics") {
			t.Errorf("expected path to start with /agents/metrics, got %s", uri)
		}
		if !strings.Contains(uri, "limit=5") {
			t.Errorf("expected limit=5 in URI, got %s", uri)
		}
		if !strings.Contains(uri, "starting_after=cursor-abc") {
			t.Errorf("expected starting_after=cursor-abc in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cartesia.MetricListResponse{
			Data: []cartesia.Metric{
				{ID: "metric-2", Name: "latency", Prompt: "Check latency", CreatedAt: "2025-01-01T00:00:00Z"},
			},
			HasMore: true,
		})
	}))

	got, err := client.Agents.Metrics.List(context.Background(), &cartesia.MetricListParams{
		Limit:         cartesia.Int(5),
		StartingAfter: cartesia.String("cursor-abc"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(got.Data))
	}
	if !got.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestAgentsMetricsAddToAgent(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1/metrics/metric-1" {
			t.Errorf("expected path /agents/agent-1/metrics/metric-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Agents.Metrics.AddToAgent(context.Background(), "agent-1", "metric-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsMetricsRemoveFromAgent(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent-1/metrics/metric-1" {
			t.Errorf("expected path /agents/agent-1/metrics/metric-1, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Agents.Metrics.RemoveFromAgent(context.Background(), "agent-1", "metric-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsMetricsResultsList(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/agents/metrics/results") {
			t.Errorf("expected path to start with /agents/metrics/results, got %s", uri)
		}
		if !strings.Contains(uri, "agent_id=agent-1") {
			t.Errorf("expected agent_id=agent-1 in URI, got %s", uri)
		}
		if !strings.Contains(uri, "metric_id=metric-1") {
			t.Errorf("expected metric_id=metric-1 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []cartesia.MetricResult{
				{
					ID:           "result-1",
					AgentID:      "agent-1",
					CallID:       "call-1",
					MetricID:     "metric-1",
					MetricName:   "accuracy",
					Result:       "pass",
					Status:       "completed",
					Summary:      "Good accuracy",
					CreatedAt:    "2025-01-01T00:00:00Z",
					DeploymentID: "deploy-1",
				},
			},
			"has_more": false,
		})
	}))

	got, err := client.Agents.Metrics.Results.List(context.Background(), &cartesia.MetricResultsListParams{
		AgentID:  cartesia.String("agent-1"),
		MetricID: cartesia.String("metric-1"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got.Data))
	}
	if got.Data[0].Result != "pass" {
		t.Errorf("expected result pass, got %s", got.Data[0].Result)
	}
}

func TestAgentsMetricsResultsExport(t *testing.T) {
	csvData := "id,agent_id,metric_name,result\nresult-1,agent-1,accuracy,pass\n"

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		uri := r.RequestURI
		if !strings.HasPrefix(uri, "/agents/metrics/results/export") {
			t.Errorf("expected path to start with /agents/metrics/results/export, got %s", uri)
		}
		if !strings.Contains(uri, "agent_id=agent-1") {
			t.Errorf("expected agent_id=agent-1 in URI, got %s", uri)
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte(csvData))
	}))

	got, err := client.Agents.Metrics.Results.Export(context.Background(), &cartesia.MetricResultsExportParams{
		AgentID: cartesia.String("agent-1"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != csvData {
		t.Errorf("expected CSV %q, got %q", csvData, got)
	}
}
