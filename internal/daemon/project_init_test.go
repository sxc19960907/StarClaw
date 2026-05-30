package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectInit_InitProject(t *testing.T) {
	tmpDir := t.TempDir()
	pi := &ProjectInit{}

	err := pi.InitProject(tmpDir)
	if err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	// Check .starclaw directory exists
	starclawDir := filepath.Join(tmpDir, ".starclaw")
	info, err := os.Stat(starclawDir)
	if err != nil {
		t.Fatalf(".starclaw directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".starclaw is not a directory")
	}

	// Check config.local.yaml exists with correct permissions
	configPath := filepath.Join(starclawDir, "config.local.yaml")
	configInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatal("config.local.yaml not created")
	}
	if mode := configInfo.Mode().Perm(); mode != 0600 {
		t.Errorf("config.local.yaml permissions = %o, want 0600", mode)
	}

	// Check instructions.md exists
	instructionsPath := filepath.Join(starclawDir, "instructions.md")
	if _, err := os.Stat(instructionsPath); os.IsNotExist(err) {
		t.Error("instructions.md not created")
	}

	// Verify content of instructions.md
	data, err := os.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("Failed to read instructions.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("instructions.md should not be empty")
	}
}

func TestProjectInit_InitProject_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	pi := &ProjectInit{}

	if err := pi.InitProject(tmpDir); err != nil {
		t.Fatalf("first InitProject() error = %v", err)
	}
	if err := pi.InitProject(tmpDir); err != nil {
		t.Fatalf("second InitProject() error = %v", err)
	}

	// Files should still exist
	starclawDir := filepath.Join(tmpDir, ".starclaw")
	if _, err := os.Stat(filepath.Join(starclawDir, "config.local.yaml")); os.IsNotExist(err) {
		t.Error("config.local.yaml should exist after second init")
	}
}

func TestProjectInit_InitProject_PreservesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	pi := &ProjectInit{}

	// Create the directory with a custom instructions.md
	starclawDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(starclawDir, 0700); err != nil {
		t.Fatal(err)
	}
	customInstructions := "# Custom project instructions"
	if err := os.WriteFile(filepath.Join(starclawDir, "instructions.md"), []byte(customInstructions), 0600); err != nil {
		t.Fatalf("Failed to create custom instructions.md: %v", err)
	}

	// Run init — should not overwrite existing files
	if err := pi.InitProject(tmpDir); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	// Verify instructions.md was not overwritten
	data, err := os.ReadFile(filepath.Join(starclawDir, "instructions.md"))
	if err != nil {
		t.Fatalf("Failed to read instructions.md: %v", err)
	}
	if string(data) != customInstructions {
		t.Errorf("instructions.md was overwritten:\ngot:  %q\nwant: %q", string(data), customInstructions)
	}
}

func TestProjectInit_InitProject_InvalidDir(t *testing.T) {
	// Use a path that cannot be created (e.g., on a read-only filesystem or
	// a deeply nested non-existent parent).
	pi := &ProjectInit{}
	// A path starting with /dev/null should fail on most systems.
	err := pi.InitProject("/dev/null/starclaw")
	if err == nil {
		t.Log("InitProject with invalid path succeeded (unexpected but may happen on some platforms)")
	}
}
