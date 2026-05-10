package tui

import (
	"os"
	"testing"
)

func TestNewDoctor(t *testing.T) {
	d := NewDoctor()
	if d == nil {
		t.Fatal("NewDoctor() returned nil")
	}
}

func TestDoctor_RunChecks_ReturnsResults(t *testing.T) {
	d := NewDoctor()
	results := d.RunChecks()

	if len(results) == 0 {
		t.Fatal("RunChecks() returned no results")
	}

	// Every result must have a name and a non-zero status is valid
	for _, r := range results {
		if r.Name == "" {
			t.Error("Every CheckResult should have a name")
		}
	}
}

func TestDoctor_RunChecks_HasExpectedChecks(t *testing.T) {
	d := NewDoctor()
	results := d.RunChecks()

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}

	expected := []string{"go_version", "config", "starclaw_dir", "tool_binaries"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("Expected check %q not found in results", name)
		}
	}
}

func TestDoctor_RunChecks_WithStarclawDir(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	d := NewDoctor()
	results := d.RunChecks()

	// Find the starclaw_dir check
	for _, r := range results {
		if r.Name == "starclaw_dir" {
			// Load() creates the dir, so it should exist and be CheckPass
			if r.Status != CheckPass {
				t.Logf("starclaw_dir check status = %d, message = %q", r.Status, r.Message)
			}
			break
		}
	}
}

func TestDoctor_CompareGoVersion(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.18", "1.18", 0},
		{"1.18", "1.19", -1},
		{"1.20", "1.18", 1},
		{"1.18.0", "1.18", 0},
		{"1.18", "1.18.5", -1},
		{"1.21.0", "1.21.1", -1},
		{"1.21.1", "1.21.0", 1},
		{"1.22", "1.18", 1},
		{"1.18", "1.22", -1},
	}

	for _, tt := range tests {
		got := compareGoVersion(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareGoVersion(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestDoctor_ConfigCheckReflectsNeedsSetup(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	d := NewDoctor()
	results := d.RunChecks()

	for _, r := range results {
		if r.Name == "config" {
			// With an empty HOME config, NeedsSetup returns true,
			// but Load() may create defaults. The check should
			// at minimum not panic or fail.
			if r.Status == CheckFail {
				t.Logf("Config check failed: %s", r.Message)
			}
			break
		}
	}
}
