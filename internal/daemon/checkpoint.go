package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Checkpoint provides agent checkpoint/restore functionality.
// Checkpoints are stored as individual files in a directory on disk.
type Checkpoint struct {
	dir string
}

// NewCheckpoint creates a new Checkpoint store rooted at dir.
func NewCheckpoint(dir string) *Checkpoint {
	return &Checkpoint{dir: dir}
}

// Save persists a checkpoint named id with the given state bytes.
func (cp *Checkpoint) Save(id string, state []byte) error {
	if err := os.MkdirAll(cp.dir, 0700); err != nil {
		return fmt.Errorf("create checkpoint dir: %w", err)
	}
	path := filepath.Join(cp.dir, sanitizeID(id))
	if err := os.WriteFile(path, state, 0600); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}
	return nil
}

// Load reads a checkpoint by id. Returns an error if the checkpoint does not
// exist.
func (cp *Checkpoint) Load(id string) ([]byte, error) {
	path := filepath.Join(cp.dir, sanitizeID(id))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load checkpoint: %w", err)
	}
	return data, nil
}

// List returns the IDs of all stored checkpoints. Returns an empty slice
// when no checkpoints exist.
func (cp *Checkpoint) List() []string {
	entries, err := os.ReadDir(cp.dir)
	if err != nil {
		return []string{}
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ids = append(ids, e.Name())
	}
	return ids
}

// sanitizeID ensures the checkpoint ID cannot escape the checkpoint directory.
func sanitizeID(id string) string {
	cleaned := filepath.Clean(id)
	cleaned = strings.ReplaceAll(cleaned, string(filepath.Separator), "_")
	if cleaned == "." || cleaned == "" {
		return "_default"
	}
	return cleaned
}
