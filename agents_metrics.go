package cartesia

import (
	"context"
	"io"
	"net/http"
)

// AgentsMetricsService manages agent metrics.
type AgentsMetricsService struct {
	client  *Client
	Results *AgentsMetricsResultsService
}

// AgentsMetricsResultsService manages metric results.
type AgentsMetricsResultsService struct {
	client *Client
}

// Metric represents an agent evaluation metric.
type Metric struct {
	ID          string  `json:"id"`
	CreatedAt   string  `json:"created_at"`
	Name        string  `json:"name"`
	Prompt      string  `json:"prompt"`
	DisplayName *string `json:"display_name,omitempty"`
}

// MetricCreateParams are the parameters for creating a metric.
type MetricCreateParams struct {
	Name        string  `json:"name"`
	Prompt      string  `json:"prompt"`
	DisplayName *string `json:"display_name,omitempty"`
}

// MetricListParams are the query parameters for listing metrics.
type MetricListParams struct {
	Limit         *int    `url:"limit,omitempty"`
	StartingAfter *string `url:"starting_after,omitempty"`
}

// MetricListResponse is a paginated list of metrics.
type MetricListResponse struct {
	Data     []Metric `json:"data"`
	HasMore  bool     `json:"has_more"`
	NextPage *string  `json:"next_page,omitempty"`
}

// MetricResult represents an evaluation result.
type MetricResult struct {
	ID           string  `json:"id"`
	AgentID      string  `json:"agentId"`
	CallID       string  `json:"callId"`
	CreatedAt    string  `json:"createdAt"`
	DeploymentID string  `json:"deploymentId"`
	MetricID     string  `json:"metricId"`
	MetricName   string  `json:"metricName"`
	Result       string  `json:"result"`
	Status       string  `json:"status"`
	Summary      string  `json:"summary"`
	JSONResult   any     `json:"jsonResult,omitempty"`
	RunID        *string `json:"runId,omitempty"`
	Value        any     `json:"value,omitempty"`
}

// MetricResultsListParams are the query parameters for listing metric results.
type MetricResultsListParams struct {
	AgentID       *string `url:"agent_id,omitempty"`
	CallID        *string `url:"call_id,omitempty"`
	DeploymentID  *string `url:"deployment_id,omitempty"`
	EndDate       *string `url:"end_date,omitempty"`
	Limit         *int    `url:"limit,omitempty"`
	MetricID      *string `url:"metric_id,omitempty"`
	StartDate     *string `url:"start_date,omitempty"`
	StartingAfter *string `url:"starting_after,omitempty"`
}

// MetricResultsExportParams are the query parameters for exporting metric results.
type MetricResultsExportParams struct {
	AgentID      *string `url:"agent_id,omitempty"`
	CallID       *string `url:"call_id,omitempty"`
	DeploymentID *string `url:"deployment_id,omitempty"`
	EndDate      *string `url:"end_date,omitempty"`
	MetricID     *string `url:"metric_id,omitempty"`
	StartDate    *string `url:"start_date,omitempty"`
}

// MetricResultsPage is a paginated list of metric results.
type MetricResultsPage = CursorPage[MetricResult]

// Create creates a new metric.
func (s *AgentsMetricsService) Create(ctx context.Context, params MetricCreateParams) (*Metric, error) {
	var res Metric
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/agents/metrics", params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Retrieve gets a metric by ID.
func (s *AgentsMetricsService) Retrieve(ctx context.Context, metricID string) (*Metric, error) {
	var res Metric
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/metrics/"+metricID, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns metrics with pagination.
func (s *AgentsMetricsService) List(ctx context.Context, params *MetricListParams) (*MetricListResponse, error) {
	var res MetricListResponse
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/agents/metrics", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// AddToAgent associates a metric with an agent.
func (s *AgentsMetricsService) AddToAgent(ctx context.Context, agentID, metricID string) error {
	return s.client.requestDrain(ctx, http.MethodPost, "/agents/"+agentID+"/metrics/"+metricID, nil)
}

// RemoveFromAgent dissociates a metric from an agent.
func (s *AgentsMetricsService) RemoveFromAgent(ctx context.Context, agentID, metricID string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/agents/"+agentID+"/metrics/"+metricID, nil)
}

// List returns metric results with filtering.
func (s *AgentsMetricsResultsService) List(ctx context.Context, params *MetricResultsListParams) (*MetricResultsPage, error) {
	var res MetricResultsPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/agents/metrics/results", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Export returns metric results as CSV.
func (s *AgentsMetricsResultsService) Export(ctx context.Context, params *MetricResultsExportParams) (string, error) {
	resp, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/agents/metrics/results/export", nil, params, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
