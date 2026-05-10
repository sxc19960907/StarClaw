package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	spillThreshold    = 50000
	spillPreviewChars = 2000
)

// spillToDisk writes content to a temp file under the given directory and returns
// a short preview for in-context use. The caller should use the preview as the
// tool result instead of the full content.
func spillToDisk(configDir, sessionID, callID, content string) (preview string, err error) {
	dir := filepath.Join(configDir, "tmp")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("spill mkdir: %w", err)
	}

	filename := fmt.Sprintf("tool_result_%s_%s.txt", sanitizeSpillID(sessionID), sanitizeSpillID(callID))
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("spill write: %w", err)
	}

	runes := []rune(content)
	previewRunes := runes
	if len(previewRunes) > spillPreviewChars {
		previewRunes = previewRunes[:spillPreviewChars]
	}

	preview = fmt.Sprintf("[Output saved to disk: %s (%s chars)]\n\nPreview (first %d chars):\n%s",
		path, strconv.Itoa(len(runes)), len(previewRunes), string(previewRunes))
	return preview, nil
}

// cleanupSpills removes all spill files for a given session ID.
func cleanupSpills(configDir, sessionID string) {
	dir := filepath.Join(configDir, "tmp")
	pattern := filepath.Join(dir, fmt.Sprintf("tool_result_%s_*.txt", sanitizeSpillID(sessionID)))
	matches, _ := filepath.Glob(pattern)
	for _, m := range matches {
		os.Remove(m)
	}
	os.Remove(dir) // best-effort: remove tmp dir if empty
}

// sanitizeSpillID strips path separators and traversal sequences from an ID
// to prevent directory escape when used in filenames.
func sanitizeSpillID(id string) string {
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "\\", "_")
	id = strings.ReplaceAll(id, "..", "_")
	if id == "" {
		return "_"
	}
	return id
}
