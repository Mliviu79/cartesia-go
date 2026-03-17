// Package cartesia provides a Go client for the Cartesia AI API.
//
// Usage:
//
//	client := cartesia.NewClient("your-api-key")
//
//	// Text-to-Speech
//	audio, err := client.TTS.Generate(ctx, cartesia.TTSRequest{
//	    ModelID:    "sonic-2",
//	    Transcript: "Hello, world!",
//	    Voice:      cartesia.VoiceSpecifier{Mode: "id", ID: "voice-id"},
//	    OutputFormat: cartesia.OutputFormat{Container: "wav", Encoding: "pcm_s16le", SampleRate: 44100},
//	})
//
//	// List Voices
//	page, err := client.Voices.List(ctx, &cartesia.VoicesListParams{Limit: cartesia.Int(10)})
//
//	// WebSocket TTS
//	ws, err := client.TTS.WebSocket(ctx)
//	defer ws.Close()
package cartesia

const (
	// SDKVersion is the version of this SDK.
	SDKVersion = "1.0.0"

	// APIVersion is the default Cartesia API version header.
	APIVersion = "2025-11-04"

	// DefaultBaseURL is the default Cartesia API base URL.
	DefaultBaseURL = "https://api.cartesia.ai"
)
