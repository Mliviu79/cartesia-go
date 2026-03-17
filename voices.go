package cartesia

import (
	"context"
	"net/http"
)

// VoicesService manages voices.
type VoicesService struct {
	client *Client
}

// Voice represents a Cartesia voice.
type Voice struct {
	ID             string  `json:"id"`
	CreatedAt      string  `json:"created_at"`
	Description    string  `json:"description"`
	IsOwner        bool    `json:"is_owner"`
	IsPublic       bool    `json:"is_public"`
	Language       string  `json:"language"`
	Name           string  `json:"name"`
	Gender         *string `json:"gender,omitempty"`
	PreviewFileURL *string `json:"preview_file_url,omitempty"`
}

// VoiceMetadata is returned from clone and localize operations.
type VoiceMetadata struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	Language    string `json:"language"`
	Name        string `json:"name"`
	UserID      string `json:"user_id"`
}

// VoicesPage is a paginated list of voices.
type VoicesPage = CursorPage[Voice]

// VoiceUpdateParams are the parameters for updating a voice.
type VoiceUpdateParams struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Gender      *string `json:"gender,omitempty"`
}

// VoicesListParams are the query parameters for listing voices.
type VoicesListParams struct {
	Expand        []string `url:"expand,omitempty"`
	Gender        *string  `url:"gender,omitempty"`
	IsOwner       *bool    `url:"is_owner,omitempty"`
	Limit         *int     `url:"limit,omitempty"`
	Q             *string  `url:"q,omitempty"`
	StartingAfter *string  `url:"starting_after,omitempty"`
}

// VoiceGetParams are the query parameters for getting a voice.
type VoiceGetParams struct {
	Expand []string `url:"expand,omitempty"`
}

// VoiceCloneParams are the parameters for cloning a voice.
type VoiceCloneParams struct {
	Clip        FileParam
	Name        string
	Description string
	Language    string
	BaseVoiceID string
}

// VoiceLocalizeParams are the parameters for localizing a voice.
type VoiceLocalizeParams struct {
	VoiceID               string  `json:"voice_id"`
	Language              string  `json:"language"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	OriginalSpeakerGender string  `json:"original_speaker_gender"`
	Dialect               *string `json:"dialect,omitempty"`
}

// Update modifies a voice.
func (s *VoicesService) Update(ctx context.Context, id string, params VoiceUpdateParams) (*Voice, error) {
	var res Voice
	_, err := s.client.requestJSON(ctx, http.MethodPatch, "/voices/"+id, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns voices with pagination and filtering.
func (s *VoicesService) List(ctx context.Context, params *VoicesListParams) (*VoicesPage, error) {
	var res VoicesPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/voices", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes a voice.
func (s *VoicesService) Delete(ctx context.Context, id string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/voices/"+id, nil)
}

// Clone creates a new voice from an audio clip using multipart upload.
func (s *VoicesService) Clone(ctx context.Context, params VoiceCloneParams) (*VoiceMetadata, error) {
	form := NewMultipartForm()

	if params.Clip.Reader != nil {
		if err := form.WriteFile("clip", params.Clip.FileName, params.Clip.Reader); err != nil {
			return nil, err
		}
	}
	if params.Name != "" {
		_ = form.WriteField("name", params.Name)
	}
	if params.Description != "" {
		_ = form.WriteField("description", params.Description)
	}
	if params.Language != "" {
		_ = form.WriteField("language", params.Language)
	}
	if params.BaseVoiceID != "" {
		_ = form.WriteField("base_voice_id", params.BaseVoiceID)
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	var res VoiceMetadata
	_, err := s.client.requestMultipart(ctx, http.MethodPost, "/voices/clone", form, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get retrieves a single voice by ID.
func (s *VoicesService) Get(ctx context.Context, id string, params *VoiceGetParams) (*Voice, error) {
	var res Voice
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/voices/"+id, nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Localize creates a localized version of a voice.
func (s *VoicesService) Localize(ctx context.Context, params VoiceLocalizeParams) (*VoiceMetadata, error) {
	var res VoiceMetadata
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/voices/localize", params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
