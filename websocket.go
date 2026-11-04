package cartesia

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TTSWebSocket is a persistent WebSocket connection for streaming TTS.
type TTSWebSocket struct {
	conn            *websocket.Conn
	client          *Client
	closeWriteBound time.Duration
	mu              sync.Mutex
	closed          bool
}

// defaultCloseWriteBound bounds a WebSocket's close frame write when the client
// sets none: 5 s, the value ai-runtime's realtime connections default to.
const defaultCloseWriteBound = 5 * time.Second

// WSGenerationRequest is a TTS generation request sent over WebSocket.
type WSGenerationRequest struct {
	ModelID              string            `json:"model_id"`
	Transcript           string            `json:"transcript"`
	Voice                VoiceSpecifier    `json:"voice"`
	OutputFormat         OutputFormat      `json:"output_format"`
	ContextID            string            `json:"context_id,omitempty"`
	Continue             bool              `json:"continue,omitempty"`
	Flush                bool              `json:"flush,omitempty"`
	AddTimestamps        bool              `json:"add_timestamps,omitempty"`
	AddPhonemeTimestamps bool              `json:"add_phoneme_timestamps,omitempty"`
	Language             string            `json:"language,omitempty"`
	Duration             *float64          `json:"duration,omitempty"`
	GenerationConfig     *GenerationConfig `json:"generation_config,omitempty"`
}

// WSCancelRequest cancels generation for a specific context.
type WSCancelRequest struct {
	Cancel    bool   `json:"cancel"`
	ContextID string `json:"context_id"`
}

// WSResponse is a message received from the TTS WebSocket.
type WSResponse struct {
	Type       string  `json:"type"`
	StatusCode int     `json:"status_code"`
	Done       bool    `json:"done"`
	ContextID  *string `json:"context_id,omitempty"`
	FlushID    *int    `json:"flush_id,omitempty"`
	Data       string  `json:"data,omitempty"`
	StepTime   float64 `json:"step_time,omitempty"`
	FlushDone  bool    `json:"flush_done,omitempty"`
	Error      string  `json:"error,omitempty"`

	WordTimestamps    *WSTimestamps `json:"word_timestamps,omitempty"`
	PhonemeTimestamps *WSTimestamps `json:"phoneme_timestamps,omitempty"`
}

// WSTimestamps contains timing data for words or phonemes.
type WSTimestamps struct {
	Words    []string  `json:"words,omitempty"`
	Phonemes []string  `json:"phonemes,omitempty"`
	Start    []float64 `json:"start"`
	End      []float64 `json:"end"`
}

// newTTSWebSocket establishes a WebSocket connection to the TTS endpoint.
//
// ctx bounds the whole dial: the TCP connect, the TLS handshake and the
// upgrade exchange. The dialer stops honouring a cancel once the connection is
// open (only a deadline reaches the upgrade read after that), so the end of
// ctx closes the connection under the upgrade, and the dial returns an error
// wrapping context.Cause(ctx). The close is joined before the dial returns.
func newTTSWebSocket(ctx context.Context, c *Client) (*TTSWebSocket, error) {
	wsURL := *c.baseURL
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	wsURL.Path = "/tts/websocket"

	q := url.Values{}
	if c.token != "" {
		q.Set("api_key", c.token)
	} else if c.apiKey != "" {
		q.Set("api_key", c.apiKey)
	}
	q.Set("cartesia_version", c.version)
	wsURL.RawQuery = q.Encode()

	header := http.Header{}
	header.Set("User-Agent", fmt.Sprintf("cartesia-go/%s", SDKVersion))

	c.logger.Debug("connecting websocket", zap.String("url", wsURL.Host+wsURL.Path))

	// cut is the close tied to ctx, registered once the connection is open;
	// cutDone closes when its callback has closed the connection.
	var cut func() bool
	cutDone := make(chan struct{})
	dialer := *websocket.DefaultDialer
	dialer.NetDialContext = func(dialCtx context.Context, network, addr string) (net.Conn, error) {
		netConn, err := (&net.Dialer{}).DialContext(dialCtx, network, addr)
		if err != nil {
			return nil, err
		}
		cut = context.AfterFunc(ctx, func() {
			defer close(cutDone)
			_ = netConn.Close() // the dial reports ctx's cause, not the close
		})
		return netConn, nil
	}
	conn, _, err := dialer.DialContext(ctx, wsURL.String(), header)
	if cut != nil && !cut() {
		// ctx ended during the dial and its callback closed the connection.
		<-cutDone
		if err == nil {
			_ = conn.Close() // already closed underneath; this releases the rest
		}
		return nil, fmt.Errorf("cartesia: websocket dial: %w", context.Cause(ctx))
	}
	if err != nil {
		return nil, fmt.Errorf("cartesia: websocket dial: %w", err)
	}

	c.logger.Debug("websocket connected")

	return &TTSWebSocket{
		conn:            conn,
		client:          c,
		closeWriteBound: c.closeWriteBound,
	}, nil
}

// Send sends a generation request over the WebSocket.
func (ws *TTSWebSocket) Send(ctx context.Context, req WSGenerationRequest) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("cartesia: websocket is closed")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cartesia: marshal ws request: %w", err)
	}

	ws.client.logger.Debug("ws send", zap.String("context_id", req.ContextID))
	return ws.conn.WriteMessage(websocket.TextMessage, data)
}

// Cancel cancels generation for a specific context.
func (ws *TTSWebSocket) Cancel(ctx context.Context, contextID string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("cartesia: websocket is closed")
	}

	data, err := json.Marshal(WSCancelRequest{Cancel: true, ContextID: contextID})
	if err != nil {
		return fmt.Errorf("cartesia: marshal cancel request: %w", err)
	}

	return ws.conn.WriteMessage(websocket.TextMessage, data)
}

// Receive reads the next message from the WebSocket.
// Returns the parsed response or an error if the connection is closed.
//
// Concurrency note: Close() may be called concurrently from another goroutine
// to unblock a pending ReadMessage. The closed flag is therefore read under
// the mutex so the race detector stays happy; the subsequent ReadMessage call
// is not held under the mutex — if Close races with us, conn.Close()
// invalidates the underlying net.Conn and ReadMessage returns a net error,
// which is the intended shutdown path.
func (ws *TTSWebSocket) Receive(ctx context.Context) (*WSResponse, error) {
	ws.mu.Lock()
	if ws.closed {
		ws.mu.Unlock()
		return nil, fmt.Errorf("cartesia: websocket is closed")
	}
	ws.mu.Unlock()

	_, msg, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("cartesia: ws read: %w", err)
	}

	var resp WSResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, fmt.Errorf("cartesia: ws unmarshal: %w", err)
	}

	if resp.Type == "error" {
		ws.client.logger.Warn("ws error", zap.String("error", resp.Error), zap.Int("status", resp.StatusCode))
	}

	return &resp, nil
}

// Close sends the close frame once, best-effort, and then closes the connection.
// It never waits for the peer's close. ctx is the closing caller's: when it ends,
// the frame's write is cut at once and the transport closes, so a peer that holds
// the frame does not delay the caller. A live ctx leaves the frame's write to the
// connection's close write bound. The transport closes after the write, whatever
// its result.
func (ws *TTSWebSocket) Close(ctx context.Context) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}
	ws.closed = true
	ws.client.logger.Debug("closing websocket")

	// cutDone closes once the cut has run, so none of its work outlives Close.
	cutDone := make(chan struct{})
	stopCut := context.AfterFunc(ctx, func() {
		defer close(cutDone)
		_ = ws.conn.Close() // fails the frame's write in flight, which then ends at once
	})
	_ = ws.conn.WriteControl(websocket.CloseMessage, // best-effort: the transport closes next either way
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(ws.closeWriteBound))
	if !stopCut() {
		<-cutDone // the cut ran and closed the transport
		return nil
	}
	return ws.conn.Close()
}
