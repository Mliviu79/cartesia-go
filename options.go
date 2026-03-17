package cartesia

import (
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(rawURL string) ClientOption {
	return func(c *Client) {
		u, err := url.Parse(rawURL)
		if err == nil && u.Scheme != "" {
			c.baseURL = u
		}
	}
}

// WithVersion sets the Cartesia-Version header value.
func WithVersion(v string) ClientOption {
	return func(c *Client) {
		c.version = v
	}
}

// WithMaxRetries sets the maximum number of retry attempts for failed requests.
// Set to 0 to disable retries. Default is 2.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

// WithLogger sets a structured logger for the client.
// If not set, a no-op logger is used.
func WithLogger(l *zap.Logger) ClientOption {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithTracer sets an OpenTelemetry tracer for the client.
// If not set, a no-op tracer is used.
func WithTracer(t trace.Tracer) ClientOption {
	return func(c *Client) {
		if t != nil {
			c.tracer = t
		}
	}
}

// WithToken sets a short-lived access token for authentication.
// When set, this takes precedence over the API key.
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}
