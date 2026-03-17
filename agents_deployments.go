package cartesia

import (
	"context"
	"net/http"
)

// AgentsDeploymentsService manages agent deployments.
type AgentsDeploymentsService struct {
	client *Client
}

// Deployment represents an agent deployment.
type Deployment struct {
	ID                     string  `json:"id"`
	AgentID                string  `json:"agent_id"`
	BuildCompletedAt       string  `json:"build_completed_at"`
	BuildLogs              string  `json:"build_logs"`
	BuildStartedAt         string  `json:"build_started_at"`
	CreatedAt              string  `json:"created_at"`
	DeploymentCompletedAt  string  `json:"deployment_completed_at"`
	DeploymentStartedAt    string  `json:"deployment_started_at"`
	EnvVarCollectionID     string  `json:"env_var_collection_id"`
	GitCommitHash          string  `json:"git_commit_hash"`
	IsLive                 bool    `json:"is_live"`
	IsPinned               bool    `json:"is_pinned"`
	SourceCodeFileID       string  `json:"source_code_file_id"`
	Status                 string  `json:"status"`
	UpdatedAt              string  `json:"updated_at"`
	BuildError             *string `json:"build_error,omitempty"`
	DeploymentError        *string `json:"deployment_error,omitempty"`
}

// Retrieve gets a deployment by ID.
func (s *AgentsDeploymentsService) Retrieve(ctx context.Context, deploymentID string) (*Deployment, error) {
	var res Deployment
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/deployments/"+deploymentID, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns deployments for an agent.
func (s *AgentsDeploymentsService) List(ctx context.Context, agentID string) ([]Deployment, error) {
	var res []Deployment
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/"+agentID+"/deployments", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
