package cartesia

import (
	"context"
	"io"
	"net/http"
)

// TTSService handles text-to-speech generation.
type TTSService struct {
	client *Client
}

// OutputFormat specifies the audio output format.
type OutputFormat struct {
	Container  string `json:"container"`
	Encoding   string `json:"encoding,omitempty"`
	SampleRate int    `json:"sample_rate"`
	BitRate    int    `json:"bit_rate,omitempty"`
}

// VoiceSpecifier identifies a voice.
type VoiceSpecifier struct {
	Mode string `json:"mode"`
	ID   string `json:"id"`
}

// GenerationConfig controls TTS generation parameters.
type GenerationConfig struct {
	Emotion []string `json:"emotion,omitempty"`
	Speed   *float64 `json:"speed,omitempty"`
	Volume  *float64 `json:"volume,omitempty"`
}

// TTSRequest is the request body for TTS generation.
type TTSRequest struct {
	ModelID              string            `json:"model_id"`
	Transcript           string            `json:"transcript"`
	Voice                VoiceSpecifier    `json:"voice"`
	OutputFormat         OutputFormat      `json:"output_format"`
	GenerationConfig     *GenerationConfig `json:"generation_config,omitempty"`
	Language             string            `json:"language,omitempty"`
	PronunciationDictID  *string           `json:"pronunciation_dict_id,omitempty"`
	Save                 *bool             `json:"save,omitempty"`
	AddTimestamps        *bool             `json:"add_timestamps,omitempty"`
	AddPhonemeTimestamps *bool             `json:"add_phoneme_timestamps,omitempty"`
	ContextID            string            `json:"context_id,omitempty"`
}

// TTSInfillParams are the parameters for audio infill.
type TTSInfillParams struct {
	LeftAudio      FileParam
	RightAudio     FileParam
	Transcript     string
	VoiceID        string
	ModelID        string
	Language       string
	OutputFormat   *OutputFormat
}

// Generate produces audio bytes from text.
func (s *TTSService) Generate(ctx context.Context, params TTSRequest) ([]byte, error) {
	resp, err := s.client.requestJSON(ctx, http.MethodPost, "/tts/bytes", params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// GenerateSSE streams TTS audio via Server-Sent Events.
// The caller must close the returned SSEReader when done.
func (s *TTSService) GenerateSSE(ctx context.Context, params TTSRequest) (*SSEReader, error) {
	resp, err := s.client.requestJSON(ctx, http.MethodPost, "/tts/sse", params, nil)
	if err != nil {
		return nil, err
	}
	return NewSSEReader(resp.Body), nil
}

// Infill generates audio to fill between two audio segments using multipart upload.
func (s *TTSService) Infill(ctx context.Context, params TTSInfillParams) ([]byte, error) {
	form := NewMultipartForm()

	if params.LeftAudio.Reader != nil {
		if err := form.WriteFile("left_audio", params.LeftAudio.FileName, params.LeftAudio.Reader); err != nil {
			return nil, err
		}
	}
	if params.RightAudio.Reader != nil {
		if err := form.WriteFile("right_audio", params.RightAudio.FileName, params.RightAudio.Reader); err != nil {
			return nil, err
		}
	}
	if params.Transcript != "" {
		_ = form.WriteField("transcript", params.Transcript)
	}
	if params.VoiceID != "" {
		_ = form.WriteField("voice_id", params.VoiceID)
	}
	if params.ModelID != "" {
		_ = form.WriteField("model_id", params.ModelID)
	}
	if params.Language != "" {
		_ = form.WriteField("language", params.Language)
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	resp, err := s.client.requestMultipart(ctx, http.MethodPost, "/infill/bytes", form, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// WebSocket creates a new WebSocket connection for streaming TTS.
// The caller must close the returned TTSWebSocket when done.
func (s *TTSService) WebSocket(ctx context.Context) (*TTSWebSocket, error) {
	return newTTSWebSocket(ctx, s.client)
}
