package cartesia

import (
	"bufio"
	"io"
	"strings"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  []byte
	ID    string
}

// SSEReader reads Server-Sent Events from an io.ReadCloser.
type SSEReader struct {
	scanner *bufio.Scanner
	closer  io.Closer
}

// NewSSEReader creates a new SSE event reader.
func NewSSEReader(rc io.ReadCloser) *SSEReader {
	return &SSEReader{
		scanner: bufio.NewScanner(rc),
		closer:  rc,
	}
}

// Next reads the next SSE event. Returns io.EOF when the stream ends.
func (r *SSEReader) Next() (*SSEEvent, error) {
	var event SSEEvent
	var dataParts []string
	hasData := false

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Blank line signals end of event.
		if line == "" {
			if hasData {
				event.Data = []byte(strings.Join(dataParts, "\n"))
				return &event, nil
			}
			continue
		}

		// Comments start with colon.
		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "event":
			event.Event = value
		case "data":
			dataParts = append(dataParts, value)
			hasData = true
		case "id":
			event.ID = value
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	// Stream ended; return any pending event.
	if hasData {
		event.Data = []byte(strings.Join(dataParts, "\n"))
		return &event, nil
	}

	return nil, io.EOF
}

// Close closes the underlying reader.
func (r *SSEReader) Close() error {
	return r.closer.Close()
}
