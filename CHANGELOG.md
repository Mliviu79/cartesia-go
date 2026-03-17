# Changelog

## v1.0.0 (2025-03-17)

Initial release. Complete Go SDK for the Cartesia AI API.

### Features

- Full REST API coverage for all Cartesia endpoints
- WebSocket TTS streaming via gorilla/websocket
- SSE (Server-Sent Events) streaming for TTS and Voice Changer
- Multipart file uploads for STT, voice cloning, voice changer, dataset files, and TTS infill
- Automatic retry with exponential backoff on transient errors (408, 409, 429, 5xx)
- OpenTelemetry tracing integration
- Zap structured logging integration
- Generic cursor-based pagination with `PageIterator[T]`
- Typed error handling with status-code helpers (`IsNotFound`, `IsRateLimited`, etc.)
- Functional options pattern for client configuration
- Token-based and API key authentication
- Context support on all methods
- 100 tests passing with race detector

### API Resources

- Access Token
- Agents (with Calls, Deployments, Metrics, Metrics Results sub-resources)
- Datasets (with Files sub-resource)
- Fine-Tunes
- Pronunciation Dictionaries
- Speech-to-Text (STT)
- Text-to-Speech (TTS) with bytes, SSE, infill, and WebSocket
- Voice Changer
- Voices (with clone and localize)
