package daemon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Attachment represents a file uploaded or referenced in an agent session.
type Attachment struct {
	Filename  string    `json:"filename"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// SaveAttachment extracts a file from an HTTP multipart form and saves it
// to <starclawDir>/attachments/<sessionID>/<filename>. It expects the file
// to be submitted under the form key "file".
func SaveAttachment(starclawDir, sessionID string, r *http.Request) (string, error) {
	if starclawDir == "" {
		return "", fmt.Errorf("starclawDir is required")
	}
	if sessionID == "" {
		return "", fmt.Errorf("sessionID is required")
	}
	if r == nil {
		return "", fmt.Errorf("request is nil")
	}

	// Limit multipart form memory to 32 MB; disk spool will be used beyond that.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return "", fmt.Errorf("parse multipart form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("get form file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	filename := header.Filename
	if filename == "" {
		return "", fmt.Errorf("empty filename")
	}

	// Sanitise filename: strip any directory components.
	filename = filepath.Base(filename)
	if filename == "." || filename == string(filepath.Separator) {
		return "", fmt.Errorf("invalid filename")
	}

	attachmentsDir := filepath.Join(starclawDir, "attachments", filepath.Clean(sessionID))
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		return "", fmt.Errorf("create attachments directory: %w", err)
	}

	dest := filepath.Join(attachmentsDir, filename)

	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create destination file: %w", err)
	}

	if _, err := io.Copy(out, file); err != nil {
		_ = out.Close()
		return "", fmt.Errorf("copy file content: %w", err)
	}

	if err := out.Close(); err != nil {
		return "", fmt.Errorf("close destination file: %w", err)
	}

	return dest, nil
}

// ListAttachments returns all attachments stored for the given session ID.
// Files are read from <starclawDir>/attachments/<sessionID>/.  Returns an
// empty slice if the directory does not exist or cannot be read.
func ListAttachments(starclawDir, sessionID string) []Attachment {
	if starclawDir == "" || sessionID == "" {
		return nil
	}

	attachmentsDir := filepath.Join(starclawDir, "attachments", filepath.Clean(sessionID))

	entries, err := os.ReadDir(attachmentsDir)
	if err != nil {
		return nil
	}

	var result []Attachment
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		mimeType := detectMimeType(info.Name())

		result = append(result, Attachment{
			Filename:  info.Name(),
			Path:      filepath.Join(attachmentsDir, info.Name()),
			Size:      info.Size(),
			MimeType:  mimeType,
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

// detectMimeType returns a best-guess MIME type based on the file extension.
func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/x-yaml"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".py":
		return "text/x-python"
	case ".go":
		return "text/x-go"
	case ".sh":
		return "application/x-sh"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
