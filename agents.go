package cartesia

import (
	"context"
	"net/http"
)

// AgentsService manages agents.
type AgentsService struct {
	client      *Client
	Calls       *AgentsCallsService
	Metrics     *AgentsMetricsService
	Deployments *AgentsDeploymentsService
}

// GitRepository represents a linked git repository.
type GitRepository struct {
	Account  string `json:"account"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// AgentPhoneNumber represents a phone number on an agent summary.
type AgentPhoneNumber struct {
	ID     string `json:"id"`
	Number string `json:"number"`
}

// AgentSummary represents a Cartesia agent.
type AgentSummary struct {
	ID                string            `json:"id"`
	CreatedAt         string            `json:"created_at"`
	DeploymentCount   int               `json:"deployment_count"`
	HasTextToAgentRun bool              `json:"has_text_to_agent_run"`
	Name              string            `json:"name"`
	TTSLanguage       string            `json:"tts_language"`
	TTSVoice          string            `json:"tts_voice"`
	UpdatedAt         string            `json:"updated_at"`
	DeletedAt         *string           `json:"deleted_at,omitempty"`
	Description       *string           `json:"description,omitempty"`
	GitDeployBranch   *string           `json:"git_deploy_branch,omitempty"`
	GitRepository     *GitRepository    `json:"git_repository,omitempty"`
	PhoneNumbers      []AgentPhoneNumber `json:"phone_numbers,omitempty"`
	WebhookID         *string           `json:"webhook_id,omitempty"`
}

// AgentUpdateParams are the parameters for updating an agent.
type AgentUpdateParams struct {
	Description *string `json:"description,omitempty"`
	Name        *string `json:"name,omitempty"`
	TTSLanguage *string `json:"tts_language,omitempty"`
	TTSVoice    *string `json:"tts_voice,omitempty"`
}

// AgentListResponse is the response from listing agents.
type AgentListResponse struct {
	Summaries []AgentSummary `json:"summaries"`
}

// AgentPhoneNumberDetail is a detailed phone number record.
type AgentPhoneNumberDetail struct {
	AgentID           string `json:"agent_id"`
	CreatedAt         string `json:"created_at"`
	IsCartesiaManaged bool   `json:"is_cartesia_managed"`
	Number            string `json:"number"`
	UpdatedAt         string `json:"updated_at"`
}

// AgentTemplate represents an agent template.
type AgentTemplate struct {
	ID              string   `json:"id"`
	CreatedAt       string   `json:"created_at"`
	Name            string   `json:"name"`
	OwnerID         string   `json:"owner_id"`
	RepoURL         string   `json:"repo_url"`
	RootDir         string   `json:"root_dir"`
	UpdatedAt       string   `json:"updated_at"`
	Dependencies    []string `json:"dependencies,omitempty"`
	Description     *string  `json:"description,omitempty"`
	RequiredEnvVars []string `json:"required_env_vars,omitempty"`
}

// AgentListTemplatesResponse is the response from listing agent templates.
type AgentListTemplatesResponse struct {
	Templates []AgentTemplate `json:"templates"`
}

// Retrieve gets an agent by ID.
func (s *AgentsService) Retrieve(ctx context.Context, agentID string) (*AgentSummary, error) {
	var res AgentSummary
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/"+agentID, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Update modifies an agent.
func (s *AgentsService) Update(ctx context.Context, agentID string, params AgentUpdateParams) (*AgentSummary, error) {
	var res AgentSummary
	_, err := s.client.requestJSON(ctx, http.MethodPatch, "/agents/"+agentID, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns all agents.
func (s *AgentsService) List(ctx context.Context) (*AgentListResponse, error) {
	var res AgentListResponse
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents", nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes an agent.
func (s *AgentsService) Delete(ctx context.Context, agentID string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/agents/"+agentID, nil)
}

// ListPhoneNumbers returns phone numbers assigned to an agent.
func (s *AgentsService) ListPhoneNumbers(ctx context.Context, agentID string) ([]AgentPhoneNumberDetail, error) {
	var res []AgentPhoneNumberDetail
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/"+agentID+"/phone-numbers", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListTemplates returns available agent templates.
func (s *AgentsService) ListTemplates(ctx context.Context) (*AgentListTemplatesResponse, error) {
	var res AgentListTemplatesResponse
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/templates", nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
