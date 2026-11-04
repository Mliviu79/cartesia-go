package cartesia_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	cartesia "github.com/Mliviu79/cartesia-go"
)

// heldPeerBytes is how much of a frame the held peer reads before it stops.
const heldPeerBytes = 64 << 10

// heldFrameTranscript is the transcript of a frame far larger than the socket
// buffers hold, so its write cannot complete while the peer reads no more.
var heldFrameTranscript = strings.Repeat("a", 32<<20)

// dialSocketWithHeldPeer dials a socket to a peer that reads the first
// heldPeerBytes of a frame and then stops reading until the test ends. The
// returned channel receives once those bytes are read: from then a write of
// heldFrameTranscript is in flight on the socket.
func dialSocketWithHeldPeer(t *testing.T) (*cartesia.TTSWebSocket, <-chan struct{}) {
	t.Helper()
	upgrader := websocket.Upgrader{}
	held := make(chan struct{}, 1)
	release := make(chan struct{})
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if _, err := io.ReadFull(conn.UnderlyingConn(), make([]byte, heldPeerBytes)); err != nil {
			return
		}
		held <- struct{}{}
		<-release
	}))
	// Registered after newTestClient's server close, so it runs before it.
	t.Cleanup(func() { close(release) })
	ws, err := client.TTS.WebSocket(t.Context())
	if err != nil {
		t.Fatalf("WebSocket() = %v, want nil", err)
	}
	return ws, held
}

// startHeldSend sends heldFrameTranscript on ws with ctx and returns the send's
// result.
func startHeldSend(ctx context.Context, ws *cartesia.TTSWebSocket) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- ws.Send(ctx, cartesia.WSGenerationRequest{ModelID: "held", Transcript: heldFrameTranscript, ContextID: "held"})
	}()
	return done
}

// awaitHeldPeer fails the test unless the held peer has read the frame's first
// bytes within watchdog.
func awaitHeldPeer(t *testing.T, held <-chan struct{}, watchdog time.Duration) {
	t.Helper()
	select {
	case <-held:
	case <-time.After(watchdog):
		t.Fatal("the peer never read the first bytes of the frame")
	}
}

// TestTTSWebSocketCloseReturnsAtOnceWhileAPeerHoldsASend holds that a Close whose
// ctx has ended does not wait behind a send the peer holds: the send's write is
// in flight, and the closing caller's ctx cuts the connection at once.
func TestTTSWebSocketCloseReturnsAtOnceWhileAPeerHoldsASend(t *testing.T) {
	const watchdog = 5 * time.Second
	ws, held := dialSocketWithHeldPeer(t)
	startHeldSend(t.Context(), ws)
	awaitHeldPeer(t, held, watchdog)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- ws.Close(ctx) }()
	select {
	case <-closed:
	case <-time.After(watchdog):
		t.Fatal("Close with an ended ctx still blocked behind the held send")
	}
}

// TestTTSWebSocketSendReturnsItsCtxCauseWhenItsCtxEnds holds that a send whose
// write the peer holds returns its ctx's cause when that ctx ends.
func TestTTSWebSocketSendReturnsItsCtxCauseWhenItsCtxEnds(t *testing.T) {
	const watchdog = 5 * time.Second
	ws, held := dialSocketWithHeldPeer(t)
	cause := errors.New("cartesia test: the caller stopped the send")
	ctx, cancel := context.WithCancelCause(t.Context())
	sent := startHeldSend(ctx, ws)
	awaitHeldPeer(t, held, watchdog)

	cancel(cause)
	select {
	case err := <-sent:
		if !errors.Is(err, cause) {
			t.Errorf("Send() = %v, want an error wrapping the ctx's cause", err)
		}
	case <-time.After(watchdog):
		t.Fatal("Send still blocked after its ctx ended")
	}
}

// TestTTSWebSocketCancelReturnsWhenItsCtxEndsBehindAHeldSend holds that a Cancel
// waiting its turn behind a send the peer holds returns its ctx's cause when its
// own ctx ends.
func TestTTSWebSocketCancelReturnsWhenItsCtxEndsBehindAHeldSend(t *testing.T) {
	const watchdog = 5 * time.Second
	ws, held := dialSocketWithHeldPeer(t)
	startHeldSend(t.Context(), ws)
	awaitHeldPeer(t, held, watchdog)

	cause := errors.New("cartesia test: the caller stopped the cancel")
	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(cause)
	cancelled := make(chan error, 1)
	go func() { cancelled <- ws.Cancel(ctx, "held") }()
	select {
	case err := <-cancelled:
		if !errors.Is(err, cause) {
			t.Errorf("Cancel() = %v, want an error wrapping the ctx's cause", err)
		}
	case <-time.After(watchdog):
		t.Fatal("Cancel still blocked behind the held send after its ctx ended")
	}
}

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
