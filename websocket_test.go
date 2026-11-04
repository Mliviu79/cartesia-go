package cartesia_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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
			_ = ws.Close(context.Background())
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

// TestTTSWebSocketCloseSendsTheCloseFrame holds that closing a live socket sends
// the protocol's normal close frame before the transport ends, so the peer sees a
// normal closure rather than a connection that just stopped.
func TestTTSWebSocketCloseSendsTheCloseFrame(t *testing.T) {
	const watchdog = 5 * time.Second
	peerRead := make(chan error, 1)
	upgrader := websocket.Upgrader{}
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			peerRead <- err
			return
		}
		defer conn.Close()
		_, _, err = conn.ReadMessage()
		peerRead <- err
	}))
	ws, err := client.TTS.WebSocket(context.Background())
	if err != nil {
		t.Fatalf("WebSocket() = %v, want nil", err)
	}
	if err := ws.Close(context.Background()); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
	select {
	case err := <-peerRead:
		var closeErr *websocket.CloseError
		if !errors.As(err, &closeErr) || closeErr.Code != websocket.CloseNormalClosure {
			t.Errorf("the peer read %v, want a normal closure frame (code %d)", err, websocket.CloseNormalClosure)
		}
	case <-time.After(watchdog):
		t.Fatal("the peer never read from the socket after Close")
	}
}

// TestTTSWebSocketCloseReturnsAtOnceWhenItsCtxHasEnded holds that a Close whose
// ctx has already ended returns without waiting and ends the connection, so the
// peer's read stops at once.
func TestTTSWebSocketCloseReturnsAtOnceWhenItsCtxHasEnded(t *testing.T) {
	const watchdog = 5 * time.Second
	peerRead := make(chan error, 1)
	upgrader := websocket.Upgrader{}
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			peerRead <- err
			return
		}
		defer conn.Close()
		_, _, err = conn.ReadMessage()
		peerRead <- err
	}))
	ws, err := client.TTS.WebSocket(context.Background())
	if err != nil {
		t.Fatalf("WebSocket() = %v, want nil", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- ws.Close(ctx) }()
	select {
	case <-closed:
	case <-time.After(watchdog):
		t.Fatal("Close with an ended ctx still blocked")
	}
	select {
	case <-peerRead:
	case <-time.After(watchdog):
		t.Fatal("the peer's connection did not end after Close")
	}
}
