package cartesia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TTSWebSocket is a persistent WebSocket connection for streaming TTS.
type TTSWebSocket struct {
	conn   *websocket.Conn
	client *Client
	mu     sync.Mutex
	closed bool
}

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

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), header)
	if err != nil {
		return nil, fmt.Errorf("cartesia: websocket dial: %w", err)
	}

	c.logger.Debug("websocket connected")

	return &TTSWebSocket{
		conn:   conn,
		client: c,
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

// Close closes the WebSocket connection.
func (ws *TTSWebSocket) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}
	ws.closed = true
	ws.client.logger.Debug("closing websocket")
	return ws.conn.Close()
}
