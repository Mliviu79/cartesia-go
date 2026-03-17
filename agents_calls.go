package cartesia

import (
	"context"
	"io"
	"net/http"
)

// AgentsCallsService manages agent calls.
type AgentsCallsService struct {
	client *Client
}

// TelephonyParams contains phone call details.
type TelephonyParams struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// TextChunk is a segment of transcribed text with timing.
type TextChunk struct {
	StartTimestamp float64 `json:"start_timestamp"`
	Text           string  `json:"text"`
}

// ToolCall represents a tool invocation during a call.
type ToolCall struct {
	ID        string            `json:"id"`
	Arguments map[string]string `json:"arguments"`
	Name      string            `json:"name"`
}

// LogEvent is a log entry from a call.
type LogEvent struct {
	Event     string            `json:"event"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp float64           `json:"timestamp"`
}

// LogMetric is a metric recorded during a call.
type LogMetric struct {
	Name      string  `json:"name"`
	Timestamp float64 `json:"timestamp"`
	Value     float64 `json:"value"`
}

// AgentTranscript is a segment of the call transcript.
type AgentTranscript struct {
	EndTimestamp   float64     `json:"end_timestamp"`
	Role           string      `json:"role"`
	StartTimestamp float64     `json:"start_timestamp"`
	EndReason      *string     `json:"end_reason,omitempty"`
	LogEvent       *LogEvent   `json:"log_event,omitempty"`
	LogMetric      *LogMetric  `json:"log_metric,omitempty"`
	Text           *string     `json:"text,omitempty"`
	TextChunks     []TextChunk `json:"text_chunks,omitempty"`
	ToolCalls      []ToolCall  `json:"tool_calls,omitempty"`
	TTSTTFB        *float64    `json:"tts_ttfb,omitempty"`
	VADBufferMs    *float64    `json:"vad_buffer_ms,omitempty"`
}

// AgentCall represents a call made by or to an agent.
type AgentCall struct {
	ID              string            `json:"id"`
	AgentID         string            `json:"agent_id"`
	Status          string            `json:"status"`
	DeploymentID    *string           `json:"deployment_id,omitempty"`
	EndTime         *string           `json:"end_time,omitempty"`
	ErrorMessage    *string           `json:"error_message,omitempty"`
	StartTime       *string           `json:"start_time,omitempty"`
	Summary         *string           `json:"summary,omitempty"`
	TelephonyParams *TelephonyParams  `json:"telephony_params,omitempty"`
	Transcript      []AgentTranscript `json:"transcript,omitempty"`
}

// AgentCallsListParams are the query parameters for listing agent calls.
type AgentCallsListParams struct {
	AgentID       string  `url:"agent_id"`
	Expand        *string `url:"expand,omitempty"`
	Limit         *int    `url:"limit,omitempty"`
	StartingAfter *string `url:"starting_after,omitempty"`
}

// AgentCallsPage is a paginated list of agent calls.
type AgentCallsPage = CursorPage[AgentCall]

// Retrieve gets a call by ID.
func (s *AgentsCallsService) Retrieve(ctx context.Context, callID string) (*AgentCall, error) {
	var res AgentCall
	_, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/calls/"+callID, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List returns agent calls matching the given parameters.
func (s *AgentsCallsService) List(ctx context.Context, params AgentCallsListParams) (*AgentCallsPage, error) {
	var res AgentCallsPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/agents/calls", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// DownloadAudio returns the raw audio data of a call.
func (s *AgentsCallsService) DownloadAudio(ctx context.Context, callID string) ([]byte, error) {
	resp, err := s.client.requestJSON(ctx, http.MethodGet, "/agents/calls/"+callID+"/audio", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
