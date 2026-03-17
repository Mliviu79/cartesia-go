package cartesia

import (
	"bytes"
	"io"
	"mime/multipart"
)

// MultipartForm builds a multipart/form-data request body.
type MultipartForm struct {
	buf    bytes.Buffer
	writer *multipart.Writer
	closed bool
}

// NewMultipartForm creates a new multipart form builder.
func NewMultipartForm() *MultipartForm {
	f := &MultipartForm{}
	f.writer = multipart.NewWriter(&f.buf)
	return f
}

// WriteField adds a text field to the form.
func (f *MultipartForm) WriteField(name, value string) error {
	return f.writer.WriteField(name, value)
}

// WriteFile adds a file field to the form.
func (f *MultipartForm) WriteFile(fieldName, fileName string, r io.Reader) error {
	fw, err := f.writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, r)
	return err
}

// Close finalizes the multipart form. Must be called before Bytes() or ContentType().
func (f *MultipartForm) Close() error {
	f.closed = true
	return f.writer.Close()
}

// ContentType returns the Content-Type header value including the boundary.
func (f *MultipartForm) ContentType() string {
	return f.writer.FormDataContentType()
}

// Bytes returns the complete form body. Close() must be called first.
func (f *MultipartForm) Bytes() []byte {
	return f.buf.Bytes()
}
