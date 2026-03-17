package cartesia_test

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"

	cartesia "github.com/Mliviu79/cartesia-go"
)

func TestWriteField(t *testing.T) {
	form := cartesia.NewMultipartForm()
	if err := form.WriteField("name", "test-value"); err != nil {
		t.Fatalf("WriteField error: %v", err)
	}
	if err := form.WriteField("description", "a description"); err != nil {
		t.Fatalf("WriteField error: %v", err)
	}
	if err := form.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Parse back.
	reader := multipart.NewReader(bytes.NewReader(form.Bytes()), extractBoundary(t, form.ContentType()))

	fields := make(map[string]string)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("NextPart error: %v", err)
		}
		data, _ := io.ReadAll(part)
		fields[part.FormName()] = string(data)
		part.Close()
	}

	if fields["name"] != "test-value" {
		t.Errorf("expected name='test-value', got %q", fields["name"])
	}
	if fields["description"] != "a description" {
		t.Errorf("expected description='a description', got %q", fields["description"])
	}
}

func TestWriteFile(t *testing.T) {
	form := cartesia.NewMultipartForm()
	fileContent := "file contents here"
	if err := form.WriteFile("upload", "test.txt", strings.NewReader(fileContent)); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	if err := form.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(form.Bytes()), extractBoundary(t, form.ContentType()))

	part, err := reader.NextPart()
	if err != nil {
		t.Fatalf("NextPart error: %v", err)
	}
	defer part.Close()

	if part.FormName() != "upload" {
		t.Errorf("expected field name 'upload', got %q", part.FormName())
	}
	if part.FileName() != "test.txt" {
		t.Errorf("expected filename 'test.txt', got %q", part.FileName())
	}
	data, _ := io.ReadAll(part)
	if string(data) != fileContent {
		t.Errorf("expected file content %q, got %q", fileContent, string(data))
	}
}

func TestContentType(t *testing.T) {
	form := cartesia.NewMultipartForm()
	ct := form.ContentType()

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		t.Fatalf("ParseMediaType error: %v", err)
	}
	if mediaType != "multipart/form-data" {
		t.Errorf("expected media type 'multipart/form-data', got %q", mediaType)
	}
	if params["boundary"] == "" {
		t.Error("expected non-empty boundary parameter")
	}
}

func TestBytes_CompleteBody(t *testing.T) {
	form := cartesia.NewMultipartForm()
	_ = form.WriteField("key", "value")
	_ = form.WriteFile("file", "data.bin", strings.NewReader("\x00\x01\x02"))
	_ = form.Close()

	body := form.Bytes()
	if len(body) == 0 {
		t.Fatal("expected non-empty body after Close")
	}

	// Verify it can be parsed.
	reader := multipart.NewReader(bytes.NewReader(body), extractBoundary(t, form.ContentType()))

	partCount := 0
	for {
		_, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		partCount++
	}
	if partCount != 2 {
		t.Errorf("expected 2 parts, got %d", partCount)
	}
}

func TestWriteField_MultipleFields(t *testing.T) {
	form := cartesia.NewMultipartForm()
	for i := 0; i < 5; i++ {
		if err := form.WriteField("field", "value"); err != nil {
			t.Fatalf("WriteField error: %v", err)
		}
	}
	_ = form.Close()

	reader := multipart.NewReader(bytes.NewReader(form.Bytes()), extractBoundary(t, form.ContentType()))
	count := 0
	for {
		_, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 parts, got %d", count)
	}
}

// extractBoundary extracts the boundary from a Content-Type header value.
func extractBoundary(t *testing.T, contentType string) string {
	t.Helper()
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("failed to parse content type: %v", err)
	}
	return params["boundary"]
}
