package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewReadTracker(t *testing.T) {
	rt := NewReadTracker()
	if rt == nil {
		t.Fatal("NewReadTracker() returned nil")
	}
	if rt.read == nil {
		t.Error("read map not initialized")
	}
}

func TestReadTracker_MarkRead_HasRead(t *testing.T) {
	rt := NewReadTracker()

	path := "/test/path/file.go"
	if rt.HasRead(path) {
		t.Error("HasRead should return false before MarkRead")
	}

	rt.MarkRead(path)
	if !rt.HasRead(path) {
		t.Error("HasRead should return true after MarkRead")
	}
}

func TestReadTracker_MultipleFiles(t *testing.T) {
	rt := NewReadTracker()

	rt.MarkRead("/a.go")
	rt.MarkRead("/b.go")
	rt.MarkRead("/c.go")

	if !rt.HasRead("/a.go") {
		t.Error("Should remember file a.go")
	}
	if !rt.HasRead("/b.go") {
		t.Error("Should remember file b.go")
	}
	if rt.HasRead("/d.go") {
		t.Error("Should not claim unread file was read")
	}
}

func TestReadTracker_EmptyPath(t *testing.T) {
	rt := NewReadTracker()

	rt.MarkRead("")
	if rt.HasRead("") {
		t.Error("Empty path should never be marked as read")
	}
}

func TestCheckReadBeforeWrite(t *testing.T) {
	rt := NewReadTracker()
	rt.MarkRead("/existing.go")

	ctx := context.WithValue(context.Background(), readTrackerKey{}, rt)

	// Read file — should pass
	if err := CheckReadBeforeWrite(ctx, "/existing.go"); err != nil {
		t.Errorf("Expected nil for read file, got: %v", err)
	}

	// Unread file — should error
	if err := CheckReadBeforeWrite(ctx, "/unread.go"); err == nil {
		t.Error("Expected error for unread file")
	}

	// No tracker in context — should pass
	emptyCtx := context.Background()
	if err := CheckReadBeforeWrite(emptyCtx, "/anything.go"); err != nil {
		t.Errorf("Expected nil without tracker, got: %v", err)
	}
}

func TestNormalizePath(t *testing.T) {
	cwd, _ := os.Getwd()

	// Relative path becomes absolute
	norm := normalizePath("test.go")
	if !filepath.IsAbs(norm) {
		t.Errorf("Expected absolute path, got %s", norm)
	}
	if norm != filepath.Join(cwd, "test.go") {
		t.Errorf("Expected %s, got %s", filepath.Join(cwd, "test.go"), norm)
	}

	// Absolute path stays absolute
	norm = normalizePath("/absolute/path.go")
	if norm != "/absolute/path.go" {
		t.Errorf("Expected /absolute/path.go, got %s", norm)
	}

	// Empty path
	if normalizePath("") != "" {
		t.Error("Empty path should remain empty")
	}

	// Path with .. cleanup
	norm = normalizePath("/a/b/../c.go")
	if norm != "/a/c.go" {
		t.Errorf("Expected /a/c.go, got %s", norm)
	}
}

func TestReadTrackerKey(t *testing.T) {
	key := ReadTrackerKey()
	if key == nil {
		t.Error("ReadTrackerKey should return non-nil")
	}
}

func TestIsMemoryPath(t *testing.T) {
	if isMemoryPath("", "/anything") {
		t.Error("Should return false for empty memory dir")
	}
	if isMemoryPath("/memdir", "") {
		t.Error("Should return false for empty path")
	}
}
