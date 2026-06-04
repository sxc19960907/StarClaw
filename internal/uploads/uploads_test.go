package uploads

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()
	m, err := New(filepath.Join(dir, "uploads"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if m == nil {
		t.Fatal("New() returned nil")
	}
	// Verify directory was created.
	if _, err := os.Stat(filepath.Join(dir, "uploads")); os.IsNotExist(err) {
		t.Error("upload directory was not created")
	}
}

func TestSaveAndGet(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	content := "hello world"
	id, err := m.Save(strings.NewReader(content), "test.txt")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if id == "" {
		t.Fatal("Save() returned empty id")
	}

	path, err := m.Get(id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !strings.HasSuffix(path, "_test.txt") {
		t.Errorf("Get() path = %q, expected suffix '_test.txt'", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestSave_WithDirectoryPathInFilename(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	id, err := m.Save(strings.NewReader("data"), "../../etc/passwd")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, err := m.Get(id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	// The BaseName of the filename should be used; directory separators are stripped.
	if !strings.HasSuffix(path, "_passwd") {
		t.Errorf("Get() path = %q, expected suffix '_passwd'", path)
	}
}

func TestGet_NotFound(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = m.Get("nonexistent-id")
	if err == nil {
		t.Fatal("Get() expected error for unknown id")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	id, err := m.Save(strings.NewReader("data"), "delete_me.txt")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := m.Delete(id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// File should no longer exist.
	_, err = m.Get(id)
	if err == nil {
		t.Error("Get() should fail after Delete()")
	}
}

func TestDelete_NotFound(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = m.Delete("does-not-exist")
	if err == nil {
		t.Fatal("Delete() expected error for unknown id")
	}
}

func TestMultipleUploads(t *testing.T) {
	dir := t.TempDir()
	m, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	id1, _ := m.Save(strings.NewReader("file one"), "a.txt")
	id2, _ := m.Save(strings.NewReader("file two"), "b.txt")

	if id1 == id2 {
		t.Error("expected different IDs for different uploads")
	}

	p1, _ := m.Get(id1)
	p2, _ := m.Get(id2)

	if p1 == p2 {
		t.Error("expected different paths for different uploads")
	}

	data1, _ := os.ReadFile(p1)
	data2, _ := os.ReadFile(p2)
	if string(data1) != "file one" {
		t.Errorf("file1 content = %q, want %q", string(data1), "file one")
	}
	if string(data2) != "file two" {
		t.Errorf("file2 content = %q, want %q", string(data2), "file two")
	}
}

func TestNew_InvalidDir(t *testing.T) {
	// Cannot create directory at a path that is a file.
	tmpFile := filepath.Join(t.TempDir(), "existing_file")
	if err := os.WriteFile(tmpFile, []byte("x"), 0600); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	_, err := New(tmpFile)
	if err == nil {
		t.Error("New() expected error when path is a file")
	}
}

func TestGenerateID_Format(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID() error = %v", err)
	}
	if len(id) != 32 {
		t.Errorf("generateID() length = %d, want 32", len(id))
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("generateID() contains non-hex character %c", c)
		}
	}
}
