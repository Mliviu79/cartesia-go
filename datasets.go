package cartesia

import (
	"context"
	"net/http"
)

// DatasetsService manages datasets.
type DatasetsService struct {
	client *Client
	Files  *DatasetsFilesService
}

// Dataset represents a Cartesia dataset.
type Dataset struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// DatasetCreateParams are the parameters for creating a dataset.
type DatasetCreateParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DatasetUpdateParams are the parameters for updating a dataset.
type DatasetUpdateParams struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// DatasetsPage is a paginated list of datasets.
type DatasetsPage = CursorPage[Dataset]

// Create creates a new dataset.
func (s *DatasetsService) Create(ctx context.Context, params DatasetCreateParams) (*Dataset, error) {
	var res Dataset
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/datasets/", params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Retrieve gets a dataset by ID.
func (s *DatasetsService) Retrieve(ctx context.Context, datasetID string) (*Dataset, error) {
	var res Dataset
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/datasets/"+datasetID, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Update modifies a dataset.
func (s *DatasetsService) Update(ctx context.Context, datasetID string, params DatasetUpdateParams) error {
	return s.client.requestDrain(ctx, http.MethodPatch, "/datasets/"+datasetID, params)
}

// List returns datasets with pagination.
func (s *DatasetsService) List(ctx context.Context, params *ListParams) (*DatasetsPage, error) {
	var res DatasetsPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/datasets/", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes a dataset.
func (s *DatasetsService) Delete(ctx context.Context, datasetID string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/datasets/"+datasetID, nil)
}
