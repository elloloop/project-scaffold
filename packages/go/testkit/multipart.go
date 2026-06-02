package testkit

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MultipartRequest builds a multipart/form-data POST with text fields and file
// parts - for file-upload handler tests. files maps a form-field name to its
// bytes (the filename is reused as the field name).
func MultipartRequest(t testing.TB, procedure string, fields map[string]string, files map[string][]byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("testkit: multipart field %q: %v", k, err)
		}
	}
	for name, content := range files {
		fw, err := w.CreateFormFile(name, name)
		if err != nil {
			t.Fatalf("testkit: multipart file %q: %v", name, err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatalf("testkit: multipart write %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("testkit: multipart close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, procedure, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}
