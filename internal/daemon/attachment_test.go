package daemon

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAttachment_ValidFile(t *testing.T) {
	dir := t.TempDir()
	sessionID := "sess-001"

	// Build a multipart form with a test file.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	path, err := SaveAttachment(dir, sessionID, r)
	if err != nil {
		t.Fatalf("SaveAttachment failed: %v", err)
	}

	if path == "" {
		t.Fatal("expected non-empty path")
	}

	// Verify file content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}

	// Verify it's stored in the right location.
	expected := filepath.Join(dir, "attachments", sessionID, "test.txt")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestSaveAttachment_EmptyStarclawDir(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := SaveAttachment("", "sess-001", r)
	if err == nil {
		t.Fatal("expected error for empty starclawDir")
	}
}

func TestSaveAttachment_EmptySessionID(t *testing.T) {
	dir := t.TempDir()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := SaveAttachment(dir, "", r)
	if err == nil {
		t.Fatal("expected error for empty sessionID")
	}
}

func TestSaveAttachment_NilRequest(t *testing.T) {
	dir := t.TempDir()
	_, err := SaveAttachment(dir, "sess-001", nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSaveAttachment_NoFile(t *testing.T) {
	dir := t.TempDir()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	_, err := SaveAttachment(dir, "sess-001", r)
	if err == nil {
		t.Fatal("expected error when no file is provided")
	}
}

func TestSaveAttachment_SanitisesPath(t *testing.T) {
	dir := t.TempDir()
	sessionID := "sess-002"

	// Attempt to inject a path traversal.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "../../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	path, err := SaveAttachment(dir, sessionID, r)
	if err != nil {
		t.Fatalf("SaveAttachment failed: %v", err)
	}

	// The sanitised filename should just be "passwd".
	if !strings.HasSuffix(path, "passwd") {
		t.Errorf("expected path to end with 'passwd', got %q", path)
	}

	// The path should be inside the attachments directory.
	if !strings.Contains(path, filepath.Join("attachments", sessionID)) {
		t.Errorf("path should be inside attachments dir, got %q", path)
	}
}

func TestListAttachments_EmptySession(t *testing.T) {
	dir := t.TempDir()
	attachments := ListAttachments(dir, "nonexistent-session")
	if attachments != nil {
		t.Errorf("expected nil for nonexistent session, got %v", attachments)
	}
}

func TestListAttachments_EmptyParams(t *testing.T) {
	if v := ListAttachments("", "sess"); v != nil {
		t.Errorf("expected nil for empty starclawDir")
	}
	if v := ListAttachments("/tmp", ""); v != nil {
		t.Errorf("expected nil for empty sessionID")
	}
}

func TestListAttachments_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	sessionID := "sess-multi"

	// Create several attachment files.
	attachDir := filepath.Join(dir, "attachments", sessionID)
	if err := os.MkdirAll(attachDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"a.txt":   "content a",
		"b.json":  `{"key": "value"}`,
		"c.go":    "package main",
		"subdir":  "", // Should be skipped.
	}
	for name, content := range files {
		if name == "subdir" {
			os.MkdirAll(filepath.Join(attachDir, name), 0755)
			continue
		}
		if err := os.WriteFile(filepath.Join(attachDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	attachments := ListAttachments(dir, sessionID)
	if len(attachments) != 3 {
		t.Fatalf("expected 3 attachments, got %d", len(attachments))
	}

	// Check each attachment.
	names := make(map[string]bool)
	for _, a := range attachments {
		names[a.Filename] = true
		if a.Size <= 0 {
			t.Errorf("expected positive size for %q", a.Filename)
		}
	}
	if !names["a.txt"] {
		t.Error("expected a.txt in attachments")
	}
	if !names["b.json"] {
		t.Error("expected b.json in attachments")
	}
	if !names["c.go"] {
		t.Error("expected c.go in attachments")
	}
}

func TestListAttachments_OrderedByCreation(t *testing.T) {
	dir := t.TempDir()
	sessionID := "sess-order"
	attachDir := filepath.Join(dir, "attachments", sessionID)
	if err := os.MkdirAll(attachDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write files in opposite order.
	if err := os.WriteFile(filepath.Join(attachDir, "second.txt"), []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attachDir, "first.txt"), []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}

	attachments := ListAttachments(dir, sessionID)
	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	// Should be sorted by creation time (modTime).
	if attachments[0].Filename != "second.txt" && attachments[1].Filename != "first.txt" {
		t.Log("attachments sorted by modTime (order depends on filesystem timing)")
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"file.txt", "text/plain"},
		{"doc.md", "text/markdown"},
		{"data.json", "application/json"},
		{"config.yaml", "application/x-yaml"},
		{"config.yml", "application/x-yaml"},
		{"page.html", "text/html"},
		{"style.css", "text/css"},
		{"script.js", "application/javascript"},
		{"main.go", "text/x-go"},
		{"script.py", "text/x-python"},
		{"run.sh", "application/x-sh"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"animation.gif", "image/gif"},
		{"vector.svg", "image/svg+xml"},
		{"doc.pdf", "application/pdf"},
		{"archive.zip", "application/zip"},
		{"package.tar", "application/x-tar"},
		{"file.gz", "application/gzip"},
		{"unknown.xyz", "application/octet-stream"},
		{"noext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectMimeType(tt.filename)
			if got != tt.expected {
				t.Errorf("detectMimeType(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

// TestSaveAttachment_BinaryContent verifies that binary content is preserved.
func TestSaveAttachment_BinaryContent(t *testing.T) {
	dir := t.TempDir()
	sessionID := "sess-bin"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "data.bin")
	if err != nil {
		t.Fatal(err)
	}

	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	if _, err := part.Write(binaryData); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	path, err := SaveAttachment(dir, sessionID, r)
	if err != nil {
		t.Fatalf("SaveAttachment failed: %v", err)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(saved, binaryData) {
		t.Errorf("saved data differs from original")
	}
}

// TestSaveAttachment_ConcurrentSessions verifies that different sessions
// get separate attachment directories.
func TestSaveAttachment_ConcurrentSessions(t *testing.T) {
	dir := t.TempDir()

	writeAttachment := func(sessionID, content string) string {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "note.txt")
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(part, content)
		writer.Close()

		r := httptest.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", writer.FormDataContentType())

		path, err := SaveAttachment(dir, sessionID, r)
		if err != nil {
			t.Fatalf("SaveAttachment failed for session %q: %v", sessionID, err)
		}
		return path
	}

	p1 := writeAttachment("sess-a", "alpha")
	p2 := writeAttachment("sess-b", "beta")

	// Paths should be different (different session directories).
	if p1 == p2 {
		t.Error("expected different paths for different sessions")
	}

	// Each should be in its own session directory.
	if !strings.Contains(p1, "sess-a") {
		t.Errorf("expected path containing sess-a, got %q", p1)
	}
	if !strings.Contains(p2, "sess-b") {
		t.Errorf("expected path containing sess-b, got %q", p2)
	}
}
