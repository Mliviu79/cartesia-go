package cartesia_test

import (
	"io"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

// nopCloser wraps a reader with a no-op Close method.
type nopCloser struct {
	io.Reader
	closed bool
}

func (n *nopCloser) Close() error {
	n.closed = true
	return nil
}

func newSSEReader(data string) (*cartesia.SSEReader, *nopCloser) {
	nc := &nopCloser{Reader: strings.NewReader(data)}
	return cartesia.NewSSEReader(nc), nc
}

func TestSSEReader_SingleEvent(t *testing.T) {
	input := "event: message\ndata: hello world\n\n"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Event != "message" {
		t.Errorf("expected event 'message', got %q", ev.Event)
	}
	if string(ev.Data) != "hello world" {
		t.Errorf("expected data 'hello world', got %q", string(ev.Data))
	}

	// Should return EOF now.
	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestSSEReader_MultipleEvents(t *testing.T) {
	input := "event: start\ndata: first\n\nevent: end\ndata: second\n\n"
	reader, _ := newSSEReader(input)

	ev1, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev1.Event != "start" {
		t.Errorf("expected event 'start', got %q", ev1.Event)
	}
	if string(ev1.Data) != "first" {
		t.Errorf("expected data 'first', got %q", string(ev1.Data))
	}

	ev2, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev2.Event != "end" {
		t.Errorf("expected event 'end', got %q", ev2.Event)
	}
	if string(ev2.Data) != "second" {
		t.Errorf("expected data 'second', got %q", string(ev2.Data))
	}

	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestSSEReader_MultiLineData(t *testing.T) {
	input := "data: line one\ndata: line two\ndata: line three\n\n"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "line one\nline two\nline three"
	if string(ev.Data) != expected {
		t.Errorf("expected data %q, got %q", expected, string(ev.Data))
	}
}

func TestSSEReader_SkipsComments(t *testing.T) {
	input := ": this is a comment\nevent: ping\ndata: pong\n: another comment\n\n"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Event != "ping" {
		t.Errorf("expected event 'ping', got %q", ev.Event)
	}
	if string(ev.Data) != "pong" {
		t.Errorf("expected data 'pong', got %q", string(ev.Data))
	}
}

func TestSSEReader_EOFAtEnd(t *testing.T) {
	input := ""
	reader, _ := newSSEReader(input)

	_, err := reader.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF for empty stream, got %v", err)
	}
}

func TestSSEReader_EventWithID(t *testing.T) {
	input := "id: 42\nevent: update\ndata: payload\n\n"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.ID != "42" {
		t.Errorf("expected ID '42', got %q", ev.ID)
	}
	if ev.Event != "update" {
		t.Errorf("expected event 'update', got %q", ev.Event)
	}
	if string(ev.Data) != "payload" {
		t.Errorf("expected data 'payload', got %q", string(ev.Data))
	}
}

func TestSSEReader_DataWithoutTrailingNewlines(t *testing.T) {
	// Stream ends without a trailing blank line -- should still return the event.
	input := "data: no trailing blank"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(ev.Data) != "no trailing blank" {
		t.Errorf("expected data 'no trailing blank', got %q", string(ev.Data))
	}
}

func TestSSEReader_ConsecutiveBlankLines(t *testing.T) {
	// Multiple blank lines between events should not produce empty events.
	input := "\n\nevent: test\ndata: value\n\n\n\n"
	reader, _ := newSSEReader(input)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Event != "test" {
		t.Errorf("expected event 'test', got %q", ev.Event)
	}
	if string(ev.Data) != "value" {
		t.Errorf("expected data 'value', got %q", string(ev.Data))
	}

	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestSSEReader_Close(t *testing.T) {
	nc := &nopCloser{Reader: strings.NewReader("data: test\n\n")}
	reader := cartesia.NewSSEReader(nc)

	err := reader.Close()
	if err != nil {
		t.Fatalf("unexpected error on Close: %v", err)
	}
	if !nc.closed {
		t.Error("expected underlying reader to be closed")
	}
}
