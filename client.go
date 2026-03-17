package cartesia

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

// Client is the Cartesia API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	apiKey     string
	token      string
	version    string
	maxRetries int
	logger     *zap.Logger
	tracer     trace.Tracer

	// Services
	AccessToken        *AccessTokenService
	Agents             *AgentsService
	Datasets           *DatasetsService
	FineTunes          *FineTunesService
	PronunciationDicts *PronunciationDictsService
	STT                *STTService
	TTS                *TTSService
	VoiceChanger       *VoiceChangerService
	Voices             *VoicesService
}

// NewClient creates a new Cartesia API client.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	u, _ := url.Parse(DefaultBaseURL)

	c := &Client{
		baseURL: u,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:     apiKey,
		version:    APIVersion,
		maxRetries: 2,
		logger:     zap.NewNop(),
		tracer:     noop.NewTracerProvider().Tracer(""),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.AccessToken = &AccessTokenService{client: c}
	c.Agents = &AgentsService{
		client: c,
		Calls:  &AgentsCallsService{client: c},
		Metrics: &AgentsMetricsService{
			client:  c,
			Results: &AgentsMetricsResultsService{client: c},
		},
		Deployments: &AgentsDeploymentsService{client: c},
	}
	c.Datasets = &DatasetsService{
		client: c,
		Files:  &DatasetsFilesService{client: c},
	}
	c.FineTunes = &FineTunesService{client: c}
	c.PronunciationDicts = &PronunciationDictsService{client: c}
	c.STT = &STTService{client: c}
	c.TTS = &TTSService{client: c}
	c.VoiceChanger = &VoiceChangerService{client: c}
	c.Voices = &VoicesService{client: c}

	return c
}

// GetStatusResponse is the response from the GET / endpoint.
type GetStatusResponse struct {
	OK      bool   `json:"ok"`
	Version string `json:"version"`
}

// GetStatus returns the API status.
func (c *Client) GetStatus(ctx context.Context) (*GetStatusResponse, error) {
	var res GetStatusResponse
	_, err := c.requestJSON(ctx, http.MethodGet, "/", nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
