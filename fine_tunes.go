package cartesia

import (
	"context"
	"net/http"
)

// FineTunesService manages fine-tuning jobs.
type FineTunesService struct {
	client *Client
}

// FineTune represents a fine-tuning job.
type FineTune struct {
	ID          string `json:"id"`
	Dataset     string `json:"dataset"`
	Description string `json:"description"`
	Language    string `json:"language"`
	ModelID     string `json:"model_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
}

// FineTuneCreateParams are the parameters for creating a fine-tuning job.
type FineTuneCreateParams struct {
	Dataset     string `json:"dataset"`
	Description string `json:"description"`
	Language    string `json:"language"`
	ModelID     string `json:"model_id"`
	Name        string `json:"name"`
}

// FineTunesPage is a paginated list of fine-tuning jobs.
type FineTunesPage = CursorPage[FineTune]

// Create starts a new fine-tuning job.
func (s *FineTunesService) Create(ctx context.Context, params FineTuneCreateParams) (*FineTune, error) {
	var res FineTune
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/fine-tunes/", params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Retrieve gets a fine-tuning job by ID.
func (s *FineTunesService) Retrieve(ctx context.Context, id string) (*FineTune, error) {
	var res FineTune
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/fine-tunes/"+id, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns fine-tuning jobs with pagination.
func (s *FineTunesService) List(ctx context.Context, params *ListParams) (*FineTunesPage, error) {
	var res FineTunesPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/fine-tunes/", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes a fine-tuning job.
func (s *FineTunesService) Delete(ctx context.Context, id string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/fine-tunes/"+id, nil)
}

// ListVoices returns voices produced by a fine-tuning job.
func (s *FineTunesService) ListVoices(ctx context.Context, id string, params *ListParams) (*VoicesPage, error) {
	var res VoicesPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/fine-tunes/"+id+"/voices", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
