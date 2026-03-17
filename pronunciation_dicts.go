package cartesia

import (
	"context"
	"net/http"
)

// PronunciationDictsService manages pronunciation dictionaries.
type PronunciationDictsService struct {
	client *Client
}

// PronunciationDictItem is a text-to-phoneme mapping.
type PronunciationDictItem struct {
	Alias string `json:"alias"`
	Text  string `json:"text"`
}

// PronunciationDict represents a pronunciation dictionary.
type PronunciationDict struct {
	ID        string                  `json:"id"`
	CreatedAt string                  `json:"created_at"`
	Items     []PronunciationDictItem `json:"items"`
	Name      string                  `json:"name"`
	OwnerID   string                  `json:"owner_id"`
	Pinned    bool                    `json:"pinned"`
}

// PronunciationDictCreateParams are the parameters for creating a dictionary.
type PronunciationDictCreateParams struct {
	Name  string                  `json:"name"`
	Items []PronunciationDictItem `json:"items,omitempty"`
}

// PronunciationDictUpdateParams are the parameters for updating a dictionary.
type PronunciationDictUpdateParams struct {
	Name  *string                 `json:"name,omitempty"`
	Items []PronunciationDictItem `json:"items,omitempty"`
}

// PronunciationDictsPage is a paginated list of pronunciation dictionaries.
type PronunciationDictsPage = CursorPage[PronunciationDict]

// Create creates a new pronunciation dictionary.
func (s *PronunciationDictsService) Create(ctx context.Context, params PronunciationDictCreateParams) (*PronunciationDict, error) {
	var res PronunciationDict
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/pronunciation-dicts/", params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Retrieve gets a pronunciation dictionary by ID.
func (s *PronunciationDictsService) Retrieve(ctx context.Context, id string) (*PronunciationDict, error) {
	var res PronunciationDict
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/pronunciation-dicts/"+id, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Update modifies a pronunciation dictionary.
func (s *PronunciationDictsService) Update(ctx context.Context, id string, params PronunciationDictUpdateParams) (*PronunciationDict, error) {
	var res PronunciationDict
	_, err := s.client.requestJSON(ctx, http.MethodPatch, "/pronunciation-dicts/"+id, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns pronunciation dictionaries with pagination.
func (s *PronunciationDictsService) List(ctx context.Context, params *ListParams) (*PronunciationDictsPage, error) {
	var res PronunciationDictsPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/pronunciation-dicts/", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes a pronunciation dictionary.
func (s *PronunciationDictsService) Delete(ctx context.Context, id string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/pronunciation-dicts/"+id, nil)
}
