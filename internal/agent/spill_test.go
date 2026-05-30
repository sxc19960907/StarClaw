package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpillToDisk(t *testing.T) {
	tmpDir := t.TempDir()

	// Content larger than spillThreshold
	largeContent := strings.Repeat("abcdefghij", 6000) // 60,000 chars
	sessionID := "test-session-123"
	callID := "call-1"

	preview, err := spillToDisk(tmpDir, sessionID, callID, largeContent)
	if err != nil {
		t.Fatalf("spillToDisk failed: %v", err)
	}

	if preview == "" {
		t.Error("Expected non-empty preview")
	}
	if !strings.Contains(preview, "Output saved to disk") {
		t.Error("Preview should mention output saved to disk")
	}
	if !strings.Contains(preview, "Preview") {
		t.Error("Preview should contain preview section")
	}

	// Verify file was written
	expectedFile := filepath.Join(tmpDir, "tmp", "tool_result_test-session-123_call-1.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected spill file at %s", expectedFile)
	}

	// Verify file content
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read spill file: %v", err)
	}
	if string(data) != largeContent {
		t.Error("Spill file content doesn't match original")
	}

	// Preview should be shorter than full content
	if len(preview) >= len(largeContent) {
		t.Error("Preview should be shorter than full content")
	}
}

func TestSpillToDisk_SmallContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Content below spillThreshold
	smallContent := "hello world"
	_, err := spillToDisk(tmpDir, "sess", "call", smallContent)
	if err != nil {
		t.Fatalf("spillToDisk should not error for small content: %v", err)
	}
	// Small content is still written to disk (caller should skip if not needed)
}

func TestSpillToDisk_MkdirError(t *testing.T) {
	// Create a file where the tmp directory should be
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "tmp"), []byte("block"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := spillToDisk(tmpDir, "sess", "call", strings.Repeat("x", 60000))
	if err == nil {
		t.Error("Expected error when tmp is a file not a directory")
	}
}

func TestCleanupSpills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some spill files
	sessionID := "cleanup-test"
	for i := 0; i < 3; i++ {
		if _, err := spillToDisk(tmpDir, sessionID, "call-"+string(rune('0'+i)), strings.Repeat("x", 60000)); err != nil {
			t.Fatal(err)
		}
	}

	// Create a file for a DIFFERENT session (should not be cleaned up)
	otherSession := "other-session"
	if _, err := spillToDisk(tmpDir, otherSession, "call-0", strings.Repeat("x", 60000)); err != nil {
		t.Fatal(err)
	}

	// Cleanup the first session
	cleanupSpills(tmpDir, sessionID)

	// Session files should be gone
	pattern := filepath.Join(tmpDir, "tmp", "tool_result_cleanup-test_*.txt")
	matches, _ := filepath.Glob(pattern)
	if len(matches) != 0 {
		t.Errorf("Expected 0 remaining files, got %d", len(matches))
	}

	// Other session files should still exist
	otherPattern := filepath.Join(tmpDir, "tmp", "tool_result_other-session_*.txt")
	otherMatches, _ := filepath.Glob(otherPattern)
	if len(otherMatches) != 1 {
		t.Errorf("Expected 1 remaining file for other session, got %d", len(otherMatches))
	}
}

func TestSpillPreviewSize(t *testing.T) {
	tmpDir := t.TempDir()
	content := strings.Repeat("hello世界", 10000) // includes unicode

	preview, err := spillToDisk(tmpDir, "sess", "call", content)
	if err != nil {
		t.Fatalf("spillToDisk failed: %v", err)
	}

	previewRunes := []rune(preview)
	if len(previewRunes) > len(content)+500 {
		t.Errorf("Preview too long: %d runes (full content: %d)", len(previewRunes), len([]rune(content)))
	}

	// Preview should contain at most spillPreviewChars chars of actual content + header
	// Just verify it doesn't panic with unicode
	t.Logf("Preview length: %d chars (header + %d limit)", len(previewRunes), spillPreviewChars)
}
