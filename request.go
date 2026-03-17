package cartesia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/google/go-querystring/query"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// newHTTPRequest creates an HTTP request with standard auth and version headers.
func (c *Client) newHTTPRequest(ctx context.Context, method, path string) (*http.Request, error) {
	u := c.baseURL.JoinPath(path)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cartesia: create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	req.Header.Set("Cartesia-Version", c.version)
	req.Header.Set("User-Agent", fmt.Sprintf("cartesia-go/%s (%s; %s)", SDKVersion, runtime.GOOS, runtime.GOARCH))

	return req, nil
}

// executeWithRetry sends an HTTP request with automatic retry for transient failures.
func (c *Client) executeWithRetry(ctx context.Context, req *http.Request, bodyBytes []byte) (*http.Response, error) {
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("cartesia %s %s", req.Method, req.URL.Path),
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		),
	)
	defer span.End()

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := retryDelay(attempt)
			c.logger.Debug("retrying request",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.String("path", req.URL.Path),
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr != nil {
			c.logger.Warn("request failed", zap.Error(lastErr), zap.Int("attempt", attempt))
			if attempt == c.maxRetries {
				span.RecordError(lastErr)
				span.SetStatus(codes.Error, "connection error")
				return nil, &ConnectionError{Err: fmt.Errorf("after %d attempts: %w", attempt+1, lastErr)}
			}
			continue
		}

		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			c.logger.Debug("retryable status", zap.Int("status", resp.StatusCode))
			resp.Body.Close()
			continue
		}

		break
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode >= 400 {
		apiErr := newAPIError(resp)
		span.RecordError(apiErr)
		span.SetStatus(codes.Error, "api error")
		c.logger.Warn("API error",
			zap.Int("status", resp.StatusCode),
			zap.String("path", req.URL.Path),
		)
		return resp, apiErr
	}

	return resp, nil
}

func retryDelay(attempt int) time.Duration {
	base := math.Pow(2, float64(attempt-1)) * 500
	jitter := rand.Float64() * 500 //nolint:gosec
	return time.Duration(base+jitter) * time.Millisecond
}

func isRetryableStatus(code int) bool {
	return code == http.StatusRequestTimeout ||
		code == http.StatusConflict ||
		code == http.StatusTooManyRequests ||
		code >= http.StatusInternalServerError
}

// requestJSON sends a JSON-encoded request and optionally decodes the JSON response.
// If v is non-nil, the response body is decoded into v and closed.
// If v is nil, the caller receives the response with body open and must close it.
func (c *Client) requestJSON(ctx context.Context, method, path string, body, v any) (*http.Response, error) {
	req, err := c.newHTTPRequest(ctx, method, path)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("cartesia: marshal body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(bodyBytes))
	}

	resp, err := c.executeWithRetry(ctx, req, bodyBytes)
	if err != nil {
		return resp, err
	}

	if v != nil {
		defer resp.Body.Close()
		if decErr := json.NewDecoder(resp.Body).Decode(v); decErr != nil && decErr != io.EOF {
			return resp, fmt.Errorf("cartesia: decode response: %w", decErr)
		}
	}

	return resp, nil
}

// requestJSONQuery is like requestJSON but appends URL query parameters from a struct.
func (c *Client) requestJSONQuery(ctx context.Context, method, path string, body, queryParams, v any) (*http.Response, error) {
	path = appendQuery(path, queryParams)
	return c.requestJSON(ctx, method, path, body, v)
}

// requestMultipart sends a multipart/form-data request and optionally decodes the response.
func (c *Client) requestMultipart(ctx context.Context, method, path string, form *MultipartForm, v any) (*http.Response, error) {
	req, err := c.newHTTPRequest(ctx, method, path)
	if err != nil {
		return nil, err
	}

	bodyBytes := form.Bytes()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", form.ContentType())
	req.ContentLength = int64(len(bodyBytes))

	resp, err := c.executeWithRetry(ctx, req, bodyBytes)
	if err != nil {
		return resp, err
	}

	if v != nil {
		defer resp.Body.Close()
		if decErr := json.NewDecoder(resp.Body).Decode(v); decErr != nil && decErr != io.EOF {
			return resp, fmt.Errorf("cartesia: decode response: %w", decErr)
		}
	}

	return resp, nil
}

// requestMultipartQuery is like requestMultipart but appends URL query parameters.
func (c *Client) requestMultipartQuery(ctx context.Context, method, path string, form *MultipartForm, queryParams, v any) (*http.Response, error) {
	path = appendQuery(path, queryParams)
	return c.requestMultipart(ctx, method, path, form, v)
}

// requestDrain sends a JSON request and discards the response body.
func (c *Client) requestDrain(ctx context.Context, method, path string, body any) error {
	resp, err := c.requestJSON(ctx, method, path, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// requestDrainQuery is like requestDrain but appends URL query parameters.
func (c *Client) requestDrainQuery(ctx context.Context, method, path string, body, queryParams any) error {
	path = appendQuery(path, queryParams)
	return c.requestDrain(ctx, method, path, body)
}

// appendQuery appends query parameters from a struct to a URL path.
func appendQuery(path string, queryParams any) string {
	if queryParams == nil {
		return path
	}
	q, err := query.Values(queryParams)
	if err == nil && len(q) > 0 {
		path += "?" + q.Encode()
	}
	return path
}
