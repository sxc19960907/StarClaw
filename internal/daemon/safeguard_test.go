package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSafeguard(t *testing.T) {
	s := NewSafeguard("/tmp/starclaw")
	if s == nil {
		t.Fatal("NewSafeguard returned nil")
	}
	if s.starclawDir != "/tmp/starclaw" {
		t.Errorf("expected starclawDir %q, got %q", "/tmp/starclaw", s.starclawDir)
	}
}

func TestCheckCommand_Safe(t *testing.T) {
	s := NewSafeguard("/tmp/starclaw")

	safeCommands := []string{
		"ls -la",
		"echo hello",
		"cat /tmp/file.txt",
		"git status",
		"go build ./...",
		"rm -rf /tmp/cache", // rm -rf on non-root path is fine
		"",
		"cd /home/user/project",
	}

	for _, cmd := range safeCommands {
		t.Run(cmd, func(t *testing.T) {
			blocked, reason := s.CheckCommand(cmd)
			if blocked {
				t.Errorf("expected safe command %q to not be blocked, reason: %s", cmd, reason)
			}
		})
	}
}

func TestCheckCommand_Blocked(t *testing.T) {
	s := NewSafeguard("/tmp/starclaw")

	blockedCommands := []struct {
		cmd    string
		reason string
	}{
		{"rm -rf /", "recursive root deletion"},
		{"rm -rf --no-preserve-root /", "recursive root deletion"},
		{"mkfs.ext4 /dev/sda", "filesystem creation"},
		{"dd if=/dev/zero of=/dev/sda", "raw disk write"},
		{":(){ :|:& };:", "fork bomb"},
		{"chmod -R 000 /etc", "permission removal"},
		{"chown -R root:root /var", "ownership change"},
	}

	for _, tc := range blockedCommands {
		t.Run(tc.cmd, func(t *testing.T) {
			blocked, reason := s.CheckCommand(tc.cmd)
			if !blocked {
				t.Errorf("expected command %q to be blocked", tc.cmd)
			}
			if reason != tc.reason {
				t.Errorf("expected reason %q, got %q", tc.reason, reason)
			}
		})
	}
}

func TestCheckCommand_CaseInsensitive(t *testing.T) {
	s := NewSafeguard("/tmp/starclaw")

	// Should block regardless of case.
	variants := []string{
		"RM -RF /",
		"MKFS.ext4 /dev/sda",
	}

	for _, cmd := range variants {
		t.Run(cmd, func(t *testing.T) {
			blocked, _ := s.CheckCommand(cmd)
			if !blocked {
				t.Errorf("expected command %q to be blocked (case-insensitive)", cmd)
			}
		})
	}
}

func TestCheckCommand_Whitespace(t *testing.T) {
	s := NewSafeguard("/tmp/starclaw")

	// Leading/trailing whitespace should be trimmed.
	blocked, reason := s.CheckCommand("  rm -rf /  ")
	if !blocked {
		t.Error("expected command with whitespace to be blocked")
	}
	if reason != "recursive root deletion" {
		t.Errorf("unexpected reason: %q", reason)
	}
}

func TestCheckPath_SafePath(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	safePath := filepath.Join(dir, "sessions", "agent1", "session.json")
	if !s.CheckPath(safePath) {
		t.Errorf("expected path %q to be safe", safePath)
	}
}

func TestCheckPath_RootPath(t *testing.T) {
	// Root is not within starclawDir.
	dir := t.TempDir()
	s := NewSafeguard(dir)

	if s.CheckPath("/") {
		t.Error("expected root path to be unsafe")
	}
}

func TestCheckPath_OutsideStarclaw(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	// A path that's a sibling of starclawDir.
	outsidePath := filepath.Join(dir, "..", "outside")
	if s.CheckPath(outsidePath) {
		t.Error("expected path outside starclawDir to be unsafe")
	}
}

func TestCheckPath_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)
	if s.CheckPath("") {
		t.Error("expected empty path to be unsafe")
	}
}

func TestCheckPath_EmptyStarclawDir(t *testing.T) {
	s := NewSafeguard("")
	if s.CheckPath("/some/path") {
		t.Error("expected path to be unsafe when starclawDir is empty")
	}
}

func TestCheckPath_NonExistentButSafe(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	// Non-existent path within starclawDir should still be safe.
	safePath := filepath.Join(dir, "nonexistent", "file.txt")
	if !s.CheckPath(safePath) {
		t.Errorf("expected non-existent path within starclawDir to be safe")
	}
}

func TestCheckPath_ActualStarclawDir(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	// Write a file actually within starclawDir.
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	if !s.CheckPath(testFile) {
		t.Errorf("expected path %q to be safe", testFile)
	}
}

func TestCheckPath_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	// Attempt to use path traversal to escape.
	traversalPath := filepath.Join(dir, "sessions", "..", "..", "etc", "passwd")
	if s.CheckPath(traversalPath) {
		t.Error("expected traversal path to be unsafe")
	}
}

func TestCheckPath_SymlinkInside(t *testing.T) {
	dir := t.TempDir()
	s := NewSafeguard(dir)

	// Create a symlink inside starclawDir pointing to a file inside.
	insideFile := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(insideFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(dir, "link.txt")
	if err := os.Symlink(insideFile, linkPath); err != nil {
		t.Skip("symlink not supported on this platform")
	}

	if !s.CheckPath(linkPath) {
		t.Errorf("expected symlink within starclawDir to be safe")
	}
}
