package cartesia

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error returned by the Cartesia API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
	Header     http.Header
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cartesia: %d %s", e.StatusCode, e.Message)
}

// ConnectionError represents a network-level connection failure.
type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("cartesia: connection error: %s", e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a request timeout.
type TimeoutError struct {
	Err error
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("cartesia: timeout: %s", e.Err)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// newAPIError reads the response body and constructs an APIError.
func newAPIError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
		Body:       body,
		Header:     resp.Header,
	}
}

// Error classification helpers.

// IsNotFound reports whether err is a 404 Not Found API error.
func IsNotFound(err error) bool { return hasStatusCode(err, 404) }

// IsBadRequest reports whether err is a 400 Bad Request API error.
func IsBadRequest(err error) bool { return hasStatusCode(err, 400) }

// IsUnauthorized reports whether err is a 401 Unauthorized API error.
func IsUnauthorized(err error) bool { return hasStatusCode(err, 401) }

// IsForbidden reports whether err is a 403 Forbidden API error.
func IsForbidden(err error) bool { return hasStatusCode(err, 403) }

// IsConflict reports whether err is a 409 Conflict API error.
func IsConflict(err error) bool { return hasStatusCode(err, 409) }

// IsUnprocessableEntity reports whether err is a 422 Unprocessable Entity API error.
func IsUnprocessableEntity(err error) bool { return hasStatusCode(err, 422) }

// IsRateLimited reports whether err is a 429 Rate Limited API error.
func IsRateLimited(err error) bool { return hasStatusCode(err, 429) }

// IsServerError reports whether err is a 5xx server error.
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// IsConnectionError reports whether err is a connection-level error.
func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr)
}

// IsTimeoutError reports whether err is a timeout error.
func IsTimeoutError(err error) bool {
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

func hasStatusCode(err error, code int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == code
	}
	return false
}
