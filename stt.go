package cartesia

import (
	"context"
	"net/http"
)

// STTService handles speech-to-text transcription.
type STTService struct {
	client *Client
}

// STTWord represents a word with timing information.
type STTWord struct {
	End   float64 `json:"end"`
	Start float64 `json:"start"`
	Word  string  `json:"word"`
}

// STTTranscribeResponse is the response from a transcription request.
type STTTranscribeResponse struct {
	Text     string   `json:"text"`
	Duration *float64 `json:"duration,omitempty"`
	Language *string  `json:"language,omitempty"`
	Words    []STTWord `json:"words,omitempty"`
}

// STTTranscribeParams are the parameters for transcribing audio.
type STTTranscribeParams struct {
	File                   FileParam
	Language               string
	Model                  string
	Encoding               string
	SampleRate             *int
	TimestampGranularities []string
}

// FileParam holds a file for multipart upload.
type FileParam struct {
	Reader   interface{ Read([]byte) (int, error) }
	FileName string
}

// sttQueryParams are the URL query parameters for STT.
type sttQueryParams struct {
	Encoding   string `url:"encoding,omitempty"`
	SampleRate *int   `url:"sample_rate,omitempty"`
}

// Transcribe converts audio to text using multipart upload.
func (s *STTService) Transcribe(ctx context.Context, params STTTranscribeParams) (*STTTranscribeResponse, error) {
	form := NewMultipartForm()

	if params.File.Reader != nil {
		if err := form.WriteFile("file", params.File.FileName, params.File.Reader); err != nil {
			return nil, err
		}
	}
	if params.Language != "" {
		if err := form.WriteField("language", params.Language); err != nil {
			return nil, err
		}
	}
	if params.Model != "" {
		if err := form.WriteField("model", params.Model); err != nil {
			return nil, err
		}
	}
	for _, g := range params.TimestampGranularities {
		if err := form.WriteField("timestamp_granularities", g); err != nil {
			return nil, err
		}
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	query := &sttQueryParams{
		Encoding:   params.Encoding,
		SampleRate: params.SampleRate,
	}

	var res STTTranscribeResponse
	_, err := s.client.requestMultipartQuery(ctx, http.MethodPost, "/stt", form, query, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
