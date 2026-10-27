package cartesia_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// TestTTSWebSocketDialReturnsWhenCancelledDuringTheUpgrade verifies that a dial
// whose server accepted the connection and never answered the upgrade returns
// when its ctx is cancelled, with an error wrapping the ctx's cause, and that
// the connection is closed rather than left to the handshake timeout.
func TestTTSWebSocketDialReturnsWhenCancelledDuringTheUpgrade(t *testing.T) {
	const watchdog = 5 * time.Second
	reached := make(chan struct{}, 1)
	cut := make(chan struct{}, 1)
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached <- struct{}{}
		<-r.Context().Done()
		cut <- struct{}{}
	}))
	cause := errors.New("cartesia test: the caller stopped the dial")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	done := make(chan error, 1)
	go func() {
		ws, err := client.TTS.WebSocket(ctx)
		if ws != nil {
			_ = ws.Close()
		}
		done <- err
	}()

	select {
	case <-reached:
	case <-time.After(watchdog):
		t.Fatal("the upgrade request never reached the server")
	}
	cancel(cause)
	select {
	case err := <-done:
		if !errors.Is(err, cause) {
			t.Errorf("WebSocket() = %v, want an error wrapping the ctx's cause", err)
		}
	case <-time.After(watchdog):
		t.Fatal("WebSocket() still blocked after its ctx was cancelled")
	}
	select {
	case <-cut:
	case <-time.After(watchdog):
		t.Error("the connection is still open after WebSocket() returned")
	}
}
