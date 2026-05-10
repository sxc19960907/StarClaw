package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type readTrackerKey struct{}

// ReadTrackerKey returns the context key used to store a ReadTracker.
func ReadTrackerKey() any { return readTrackerKey{} }

// ReadTracker tracks which files have been read during the current agent turn.
type ReadTracker struct {
	mu   sync.Mutex
	read map[string]bool
}

// NewReadTracker creates a new ReadTracker.
func NewReadTracker() *ReadTracker {
	return &ReadTracker{read: make(map[string]bool)}
}

// MarkRead records that a file has been read.
func (rt *ReadTracker) MarkRead(path string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	norm := normalizePath(path)
	if norm != "" {
		rt.read[norm] = true
	}
}

// HasRead returns true if the file has been read in this turn.
func (rt *ReadTracker) HasRead(path string) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	norm := normalizePath(path)
	if norm == "" {
		return false
	}
	return rt.read[norm]
}

// CheckReadBeforeWrite extracts the ReadTracker from context and returns an error
// if the given path has not been read. Returns nil if the tracker is absent.
func CheckReadBeforeWrite(ctx context.Context, path string) error {
	rt, ok := ctx.Value(readTrackerKey{}).(*ReadTracker)
	if !ok || rt == nil {
		return nil
	}
	if !rt.HasRead(path) {
		return fmt.Errorf("You must read this file with file_read before editing it. Path: %s", path)
	}
	return nil
}

// normalizePath resolves a path to an absolute, clean, symlink-resolved form.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return filepath.Clean(path)
		}
		path = filepath.Join(cwd, path)
	}
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return path
}

// isMemoryPath checks if path resolves to a MEMORY.md inside the given memory dir.
func isMemoryPath(memoryDir, path string) bool {
	if memoryDir == "" || path == "" {
		return false
	}
	return strings.EqualFold(normalizePath(path), normalizePath(filepath.Join(memoryDir, "MEMORY.md")))
}
