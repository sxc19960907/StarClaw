package daemon

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	state := []byte("agent state data")
	if err := cp.Save("my-agent", state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := cp.Load("my-agent")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if string(loaded) != string(state) {
		t.Errorf("Loaded state mismatch:\ngot:  %q\nwant: %q", string(loaded), string(state))
	}
}

func TestCheckpoint_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	_, err := cp.Load("nonexistent")
	if err == nil {
		t.Fatal("Load() should return error for nonexistent checkpoint")
	}
}

func TestCheckpoint_List(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "checkpoints")
	cp := NewCheckpoint(dir)

	ids := []string{"agent-1", "agent-2", "agent-3"}
	for _, id := range ids {
		if err := cp.Save(id, []byte("state")); err != nil {
			t.Fatalf("Save(%q) error = %v", id, err)
		}
	}

	got := cp.List()
	if len(got) != len(ids) {
		t.Fatalf("List() returned %d items, want %d", len(got), len(ids))
	}

	sort.Strings(got)
	for i, id := range ids {
		if got[i] != id {
			t.Errorf("List()[%d] = %q, want %q", i, got[i], id)
		}
	}
}

func TestCheckpoint_List_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	ids := cp.List()
	if ids == nil {
		t.Fatal("List() should return empty slice, not nil")
	}
	if len(ids) != 0 {
		t.Errorf("List() should be empty, got %d items", len(ids))
	}
}

func TestCheckpoint_List_NonExistentDir(t *testing.T) {
	cp := NewCheckpoint("/tmp/nonexistent-checkpoint-dir-12345")

	ids := cp.List()
	if ids == nil {
		t.Fatal("List() should return empty slice, not nil")
	}
	if len(ids) != 0 {
		t.Errorf("List() should be empty for non-existent dir, got %d items", len(ids))
	}
}

func TestCheckpoint_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	if err := cp.Save("agent", []byte("original")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := cp.Save("agent", []byte("updated")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := cp.Load("agent")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(loaded) != "updated" {
		t.Errorf("State should be 'updated', got %q", string(loaded))
	}
}

func TestCheckpoint_SanitizeID(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	// Start with ./ prefix which filepath.Clean resolves to "."
	err := cp.Save("./state", []byte("data"))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// After cleaning, "./state" -> "state", so we should find it as "state"
	loaded, err := cp.Load("./state")
	if err != nil {
		t.Fatalf("Load('./state') error = %v", err)
	}
	if string(loaded) != "data" {
		t.Errorf("Loaded data mismatch: got %q, want %q", string(loaded), "data")
	}
}

func TestCheckpoint_SanitizeID_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	// ID with path traversal should be sanitized
	err := cp.Save("../../malicious/path", []byte("data"))
	if err != nil {
		t.Fatalf("Save() with traversal ID error = %v", err)
	}

	// Verify there is exactly one file with a sanitized name
	entries, err := os.ReadDir(filepath.Join(tmpDir, "checkpoints"))
	if err != nil {
		t.Fatalf("ReadDir error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(entries))
	}
	name := entries[0].Name()
	// The name should not contain path separators
	if strings.ContainsRune(name, filepath.Separator) {
		t.Errorf("Checkpoint file name contains path separator: %q", name)
	}
}

func TestCheckpoint_SanitizeID_BackslashTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	if err := cp.Save(`..\..\outside`, []byte("data")); err != nil {
		t.Fatalf("Save() with backslash traversal ID error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmpDir, "checkpoints"))
	if err != nil {
		t.Fatalf("ReadDir error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(entries))
	}
	name := entries[0].Name()
	if strings.Contains(name, `\`) || strings.Contains(name, "/") || strings.Contains(name, "..") {
		t.Fatalf("Checkpoint file name was not safely sanitized: %q", name)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "outside")); !os.IsNotExist(err) {
		t.Fatalf("Traversal created file outside checkpoint dir or stat failed unexpectedly: %v", err)
	}
}

func TestCheckpoint_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	if err := cp.Save("perm-test", []byte("data")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "checkpoints", "perm-test"))
	if err != nil {
		t.Fatalf("Stat checkpoint file error = %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("Checkpoint file permissions = %o, want 0600", mode)
	}
}

func TestCheckpoint_EmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpoint(filepath.Join(tmpDir, "checkpoints"))

	if err := cp.Save("", []byte("state")); err != nil {
		t.Fatalf("Save() with empty ID error = %v", err)
	}

	loaded, err := cp.Load("")
	if err != nil {
		t.Fatalf("Load() with empty ID error = %v", err)
	}
	if string(loaded) != "state" {
		t.Errorf("Loaded data mismatch: got %q, want %q", string(loaded), "state")
	}
}

func TestCheckpoint_DirCreatedOnSave(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "deeply", "nested", "checkpoints")

	// Dir does not exist yet
	cp := NewCheckpoint(dir)

	if err := cp.Save("test", []byte("data")); err != nil {
		t.Fatalf("Save() to non-existent dir error = %v", err)
	}

	// Dir should now exist
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Checkpoint dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Checkpoint path is not a directory")
	}
}
