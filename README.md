# cartesia-go

Go SDK for the [Cartesia AI](https://cartesia.ai) API. Provides typed access to all Cartesia endpoints including text-to-speech, speech-to-text, voice cloning, agents, and more.

> **Status:** Internal use. Maintained as needed. Not tested in production.

## Installation

```bash
go get github.com/Mliviu79/cartesia-go
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    cartesia "github.com/Mliviu79/cartesia-go"
)

func main() {
    client := cartesia.NewClient(os.Getenv("CARTESIA_API_KEY"))
    ctx := context.Background()

    // Check API status
    status, err := client.GetStatus(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("API OK: %v, Version: %s\n", status.OK, status.Version)
}
```

## Usage

### Text-to-Speech

```go
// Generate audio bytes
audio, err := client.TTS.Generate(ctx, cartesia.TTSRequest{
    ModelID:    "sonic-2",
    Transcript: "Hello from Cartesia!",
    Voice:      cartesia.VoiceSpecifier{Mode: "id", ID: "your-voice-id"},
    OutputFormat: cartesia.OutputFormat{
        Container:  "wav",
        Encoding:   "pcm_s16le",
        SampleRate: 44100,
    },
})

// Stream via SSE
sse, err := client.TTS.GenerateSSE(ctx, cartesia.TTSRequest{...})
defer sse.Close()
for {
    event, err := sse.Next()
    if err != nil {
        break
    }
    // process event.Data
}
```

### WebSocket TTS

```go
ws, err := client.TTS.WebSocket(ctx)
defer ws.Close()

err = ws.Send(ctx, cartesia.WSGenerationRequest{
    ModelID:    "sonic-2",
    Transcript: "Streaming hello!",
    Voice:      cartesia.VoiceSpecifier{Mode: "id", ID: "your-voice-id"},
    OutputFormat: cartesia.OutputFormat{
        Container:  "raw",
        Encoding:   "pcm_f32le",
        SampleRate: 44100,
    },
    ContextID: "ctx-1",
    Flush:     true,
})

for {
    resp, err := ws.Receive(ctx)
    if err != nil || resp.Done {
        break
    }
    // resp.Data contains base64-encoded audio chunks
}
```

### Speech-to-Text

```go
file, _ := os.Open("audio.wav")
defer file.Close()

result, err := client.STT.Transcribe(ctx, cartesia.STTTranscribeParams{
    File:     cartesia.FileParam{Reader: file, FileName: "audio.wav"},
    Language: "en",
    Model:    "ink-whisper",
})
fmt.Println(result.Text)
```

### Voices

```go
// List voices
page, err := client.Voices.List(ctx, &cartesia.VoicesListParams{
    Limit: cartesia.Int(10),
    Q:     cartesia.String("female"),
})

// Clone a voice
clip, _ := os.Open("sample.wav")
voice, err := client.Voices.Clone(ctx, cartesia.VoiceCloneParams{
    Clip:     cartesia.FileParam{Reader: clip, FileName: "sample.wav"},
    Name:     "My Clone",
    Language: "en",
})

// Localize a voice
localized, err := client.Voices.Localize(ctx, cartesia.VoiceLocalizeParams{
    VoiceID:               "voice-id",
    Language:              "es",
    Name:                  "Spanish Voice",
    Description:           "Localized to Spanish",
    OriginalSpeakerGender: "female",
})
```

### Voice Changer

```go
clip, _ := os.Open("input.wav")
audio, err := client.VoiceChanger.ChangeVoiceBytes(ctx, cartesia.VoiceChangerParams{
    Clip:    cartesia.FileParam{Reader: clip, FileName: "input.wav"},
    VoiceID: "target-voice-id",
    OutputFormat: cartesia.OutputFormat{
        Container:  "wav",
        Encoding:   "pcm_s16le",
        SampleRate: 44100,
    },
})
```

### Agents

```go
// List agents
agents, err := client.Agents.List(ctx)

// Get agent calls
calls, err := client.Agents.Calls.List(ctx, cartesia.AgentCallsListParams{
    AgentID: "agent-id",
    Limit:   cartesia.Int(25),
})

// Download call audio
audio, err := client.Agents.Calls.DownloadAudio(ctx, "call-id")
```

### Datasets & Fine-Tuning

```go
// Create dataset
ds, err := client.Datasets.Create(ctx, cartesia.DatasetCreateParams{
    Name:        "Training Data",
    Description: "Voice training samples",
})

// Upload file
file, _ := os.Open("samples.zip")
err = client.Datasets.Files.Upload(ctx, ds.ID, cartesia.FileUploadParams{
    File:     file,
    FileName: "samples.zip",
    Purpose:  "fine_tune",
})

// Start fine-tune
ft, err := client.FineTunes.Create(ctx, cartesia.FineTuneCreateParams{
    Dataset:     ds.ID,
    ModelID:     "sonic-2",
    Name:        "My Fine-Tune",
    Language:    "en",
    Description: "Custom voice model",
})
```

### Pagination

```go
// Manual pagination
page, _ := client.Voices.List(ctx, &cartesia.VoicesListParams{Limit: cartesia.Int(10)})
for _, voice := range page.Data {
    fmt.Println(voice.Name)
}
if page.HasMore {
    next, _ := client.Voices.List(ctx, &cartesia.VoicesListParams{
        Limit:         cartesia.Int(10),
        StartingAfter: page.Next,
    })
    // ...
}

// Auto-pagination with PageIterator
iter := cartesia.NewPageIterator(
    func(p cartesia.ListParams) (*cartesia.CursorPage[cartesia.Voice], error) {
        return client.Voices.List(ctx, &cartesia.VoicesListParams{
            Limit:         p.Limit,
            StartingAfter: p.StartingAfter,
        })
    },
    cartesia.ListParams{Limit: cartesia.Int(50)},
)
for iter.Next() {
    for _, voice := range iter.Current() {
        fmt.Println(voice.Name)
    }
}
if err := iter.Err(); err != nil {
    panic(err)
}
```

## Client Configuration

```go
client := cartesia.NewClient("api-key",
    cartesia.WithBaseURL("https://custom.endpoint.com"),
    cartesia.WithMaxRetries(3),
    cartesia.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
    cartesia.WithToken("short-lived-token"),   // takes precedence over API key
    cartesia.WithLogger(zapLogger),             // structured logging
    cartesia.WithTracer(otelTracer),            // OpenTelemetry tracing
    cartesia.WithVersion("2025-11-04"),         // API version header
)
```

### Access Tokens

```go
token, err := client.AccessToken.Create(ctx, cartesia.AccessTokenCreateParams{
    ExpiresIn: cartesia.Int(3600),
    Grants: &cartesia.AccessTokenGrants{
        TTS: cartesia.Bool(true),
        STT: cartesia.Bool(true),
    },
})

// Use the token for subsequent requests
tokenClient := cartesia.NewClient("", cartesia.WithToken(token.Token))
```

## Error Handling

```go
audio, err := client.TTS.Generate(ctx, params)
if err != nil {
    if cartesia.IsRateLimited(err) {
        // back off and retry
    }
    if cartesia.IsUnauthorized(err) {
        // refresh credentials
    }
    if cartesia.IsNotFound(err) {
        // resource doesn't exist
    }

    var apiErr *cartesia.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d, Body: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Features

- **Complete API coverage** -- all Cartesia REST endpoints, WebSocket TTS, SSE streaming
- **Automatic retries** -- exponential backoff on 408, 409, 429, and 5xx errors
- **OpenTelemetry tracing** -- spans for every API call with method, path, and status attributes
- **Structured logging** -- Zap logger integration with debug/warn levels
- **Multipart uploads** -- STT, voice clone, voice changer, dataset files, and TTS infill
- **Cursor pagination** -- `CursorPage[T]` generics with `PageIterator` for auto-pagination
- **Pointer helpers** -- `Int()`, `String()`, `Bool()`, `Float64()` for optional fields
- **Context support** -- all methods accept `context.Context` for cancellation and deadlines
- **Race-safe WebSocket** -- mutex-protected writes on `TTSWebSocket`

## API Coverage

| Service | Endpoints |
|---------|-----------|
| Status | `GetStatus` |
| Access Token | `Create` |
| Agents | `Retrieve`, `Update`, `List`, `Delete`, `ListPhoneNumbers`, `ListTemplates` |
| Agents > Calls | `Retrieve`, `List`, `DownloadAudio` |
| Agents > Deployments | `Retrieve`, `List` |
| Agents > Metrics | `Create`, `Retrieve`, `List`, `AddToAgent`, `RemoveFromAgent` |
| Agents > Metrics > Results | `List`, `Export` |
| Datasets | `Create`, `Retrieve`, `Update`, `List`, `Delete` |
| Datasets > Files | `List`, `Delete`, `Upload` |
| Fine-Tunes | `Create`, `Retrieve`, `List`, `Delete`, `ListVoices` |
| Pronunciation Dicts | `Create`, `Retrieve`, `Update`, `List`, `Delete` |
| STT | `Transcribe` |
| TTS | `Generate`, `GenerateSSE`, `Infill`, `WebSocket` |
| Voice Changer | `ChangeVoiceBytes`, `ChangeVoiceSSE` |
| Voices | `Update`, `List`, `Delete`, `Clone`, `Get`, `Localize` |

## License

MIT -- see [LICENSE](LICENSE).

## Code Generation

Code generated by Jules and Claude Code.
