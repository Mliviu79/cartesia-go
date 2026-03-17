package cartesia_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestAPIError_ErrorFormat(t *testing.T) {
	err := &cartesia.APIError{
		StatusCode: 404,
		Message:    "voice not found",
	}
	expected := "cartesia: 404 voice not found"
	if got := err.Error(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestAPIError_ErrorFormat_Empty(t *testing.T) {
	err := &cartesia.APIError{
		StatusCode: 500,
		Message:    "",
	}
	expected := "cartesia: 500 "
	if got := err.Error(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestIsNotFound(t *testing.T) {
	err := &cartesia.APIError{StatusCode: http.StatusNotFound, Message: "not found"}
	if !cartesia.IsNotFound(err) {
		t.Error("expected IsNotFound=true")
	}
	if cartesia.IsNotFound(&cartesia.APIError{StatusCode: 400}) {
		t.Error("expected IsNotFound=false for 400")
	}
	if cartesia.IsNotFound(fmt.Errorf("some error")) {
		t.Error("expected IsNotFound=false for non-APIError")
	}
}

func TestIsBadRequest(t *testing.T) {
	err := &cartesia.APIError{StatusCode: http.StatusBadRequest, Message: "bad request"}
	if !cartesia.IsBadRequest(err) {
		t.Error("expected IsBadRequest=true")
	}
	if cartesia.IsBadRequest(&cartesia.APIError{StatusCode: 404}) {
		t.Error("expected IsBadRequest=false for 404")
	}
}

func TestIsUnauthorized(t *testing.T) {
	err := &cartesia.APIError{StatusCode: http.StatusUnauthorized, Message: "unauthorized"}
	if !cartesia.IsUnauthorized(err) {
		t.Error("expected IsUnauthorized=true")
	}
	if cartesia.IsUnauthorized(&cartesia.APIError{StatusCode: 200}) {
		t.Error("expected IsUnauthorized=false for 200")
	}
}

func TestIsRateLimited(t *testing.T) {
	err := &cartesia.APIError{StatusCode: http.StatusTooManyRequests, Message: "rate limited"}
	if !cartesia.IsRateLimited(err) {
		t.Error("expected IsRateLimited=true")
	}
	if cartesia.IsRateLimited(&cartesia.APIError{StatusCode: 500}) {
		t.Error("expected IsRateLimited=false for 500")
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"500", &cartesia.APIError{StatusCode: 500}, true},
		{"502", &cartesia.APIError{StatusCode: 502}, true},
		{"503", &cartesia.APIError{StatusCode: 503}, true},
		{"499", &cartesia.APIError{StatusCode: 499}, false},
		{"400", &cartesia.APIError{StatusCode: 400}, false},
		{"non-APIError", fmt.Errorf("timeout"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cartesia.IsServerError(tt.err); got != tt.expect {
				t.Errorf("IsServerError=%v, want %v", got, tt.expect)
			}
		})
	}
}

func TestConnectionError_Wraps(t *testing.T) {
	inner := fmt.Errorf("dial tcp: connection refused")
	connErr := &cartesia.ConnectionError{Err: inner}

	if !errors.Is(connErr, inner) {
		t.Error("expected errors.Is to find inner error")
	}

	got := connErr.Error()
	expected := "cartesia: connection error: dial tcp: connection refused"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestConnectionError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("network unreachable")
	connErr := &cartesia.ConnectionError{Err: inner}
	if connErr.Unwrap() != inner {
		t.Error("Unwrap should return inner error")
	}
}

func TestIsConnectionError(t *testing.T) {
	connErr := &cartesia.ConnectionError{Err: fmt.Errorf("refused")}
	if !cartesia.IsConnectionError(connErr) {
		t.Error("expected IsConnectionError=true")
	}
	if cartesia.IsConnectionError(fmt.Errorf("other")) {
		t.Error("expected IsConnectionError=false for plain error")
	}
}

func TestErrorsAs_APIError(t *testing.T) {
	var apiErr *cartesia.APIError
	err := error(&cartesia.APIError{StatusCode: 422, Message: "invalid"})

	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should succeed")
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "invalid" {
		t.Errorf("expected 'invalid', got %q", apiErr.Message)
	}
}

func TestErrorsAs_WrappedAPIError(t *testing.T) {
	inner := &cartesia.APIError{StatusCode: 404, Message: "not found"}
	wrapped := fmt.Errorf("outer: %w", inner)

	var apiErr *cartesia.APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatal("errors.As should find wrapped APIError")
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestIsNotFound_WrappedError(t *testing.T) {
	inner := &cartesia.APIError{StatusCode: 404, Message: "not found"}
	wrapped := fmt.Errorf("outer: %w", inner)
	if !cartesia.IsNotFound(wrapped) {
		t.Error("expected IsNotFound=true for wrapped APIError")
	}
}
