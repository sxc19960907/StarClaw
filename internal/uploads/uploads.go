// Package uploads provides file upload management with unique ID tracking.
package uploads

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UploadManager manages file uploads with unique IDs.
type UploadManager struct {
	uploadDir string
}

// New creates a new UploadManager that stores files in the given directory.
// The directory is created if it does not exist.
func New(uploadDir string) (*UploadManager, error) {
	if err := os.MkdirAll(uploadDir, 0700); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &UploadManager{uploadDir: uploadDir}, nil
}

// Save reads from the provided reader and saves the content as a new upload.
// It returns a unique ID that can be used to retrieve or delete the file later.
func (m *UploadManager) Save(reader io.Reader, filename string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}

	// Sanitise the filename: strip any directory components.
	safeName := filepath.Base(filename)

	destPath := filepath.Join(m.uploadDir, id+"_"+safeName)
	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	if _, err := io.Copy(f, reader); err != nil {
		// Clean up partial write on error.
		_ = f.Close()
		_ = os.Remove(destPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("close file: %w", err)
	}

	return id, nil
}

// Get returns the full file system path for the given upload ID.
// It returns an error if no matching upload is found.
func (m *UploadManager) Get(id string) (string, error) {
	path, err := m.findByID(id)
	if err != nil {
		return "", err
	}
	return path, nil
}

// Delete removes the uploaded file identified by the given ID.
func (m *UploadManager) Delete(id string) error {
	path, err := m.findByID(id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// findByID searches the upload directory for a file whose name starts with id+"_".
func (m *UploadManager) findByID(id string) (string, error) {
	entries, err := os.ReadDir(m.uploadDir)
	if err != nil {
		return "", fmt.Errorf("read upload dir: %w", err)
	}

	prefix := id + "_"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(m.uploadDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("upload %q not found", id)
}

// generateID creates a 32-character hex string suitable for use as an upload ID.
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
