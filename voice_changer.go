package cartesia

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// VoiceChangerService handles voice transformation.
type VoiceChangerService struct {
	client *Client
}

// VoiceChangerParams are the parameters for changing a voice.
type VoiceChangerParams struct {
	Clip         FileParam
	VoiceID      string
	OutputFormat OutputFormat
}

// ChangeVoiceBytes transforms the voice in an audio clip and returns raw audio bytes.
func (s *VoiceChangerService) ChangeVoiceBytes(ctx context.Context, params VoiceChangerParams) ([]byte, error) {
	form, err := buildVoiceChangerForm(params)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.requestMultipart(ctx, http.MethodPost, "/voice-changer/bytes", form, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ChangeVoiceSSE transforms the voice in an audio clip and returns an SSE stream.
// The caller must close the returned SSEReader when done.
func (s *VoiceChangerService) ChangeVoiceSSE(ctx context.Context, params VoiceChangerParams) (*SSEReader, error) {
	form, err := buildVoiceChangerForm(params)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.requestMultipart(ctx, http.MethodPost, "/voice-changer/sse", form, nil)
	if err != nil {
		return nil, err
	}
	return NewSSEReader(resp.Body), nil
}

func buildVoiceChangerForm(params VoiceChangerParams) (*MultipartForm, error) {
	form := NewMultipartForm()

	if params.Clip.Reader != nil {
		if err := form.WriteFile("clip", params.Clip.FileName, params.Clip.Reader); err != nil {
			return nil, err
		}
	}
	if params.VoiceID != "" {
		_ = form.WriteField("voice[id]", params.VoiceID)
	}
	_ = form.WriteField("output_format[container]", params.OutputFormat.Container)
	if params.OutputFormat.Encoding != "" {
		_ = form.WriteField("output_format[encoding]", params.OutputFormat.Encoding)
	}
	_ = form.WriteField("output_format[sample_rate]", fmt.Sprintf("%d", params.OutputFormat.SampleRate))
	if params.OutputFormat.BitRate > 0 {
		_ = form.WriteField("output_format[bit_rate]", fmt.Sprintf("%d", params.OutputFormat.BitRate))
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	return form, nil
}
