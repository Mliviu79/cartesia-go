package cartesia

import (
	"context"
	"net/http"
)

// AccessTokenService manages short-lived access tokens.
type AccessTokenService struct {
	client *Client
}

// AccessTokenGrants specifies which API capabilities the token grants.
type AccessTokenGrants struct {
	Agent *bool `json:"agent,omitempty"`
	STT   *bool `json:"stt,omitempty"`
	TTS   *bool `json:"tts,omitempty"`
}

// AccessTokenCreateParams are the parameters for creating an access token.
type AccessTokenCreateParams struct {
	ExpiresIn *int               `json:"expires_in,omitempty"`
	Grants    *AccessTokenGrants `json:"grants,omitempty"`
}

// AccessTokenCreateResponse is the response from creating an access token.
type AccessTokenCreateResponse struct {
	Token string `json:"token"`
}

// Create generates a short-lived access token.
func (s *AccessTokenService) Create(ctx context.Context, params ...AccessTokenCreateParams) (*AccessTokenCreateResponse, error) {
	var body any
	if len(params) > 0 {
		body = params[0]
	} else {
		body = struct{}{}
	}

	var res AccessTokenCreateResponse
	_, err := s.client.requestJSON(ctx, http.MethodPost, "/access-token", body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
