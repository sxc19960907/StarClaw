package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettings_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	s := &Settings{
		Spinner:          "moon",
		MaxResponseLines: 20,
		ShowTips:         false,
	}

	if err := s.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists with correct permissions
	settingsPath := filepath.Join(tmpDir, ".starclaw", "settings.json")
	info, err := os.Stat(settingsPath)
	if err != nil {
		t.Fatalf("settings.json not created: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("settings.json permissions = %o, want 0600", mode)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if loaded.Spinner != "moon" {
		t.Errorf("Spinner = %q, want %q", loaded.Spinner, "moon")
	}
	if loaded.MaxResponseLines != 20 {
		t.Errorf("MaxResponseLines = %d, want %d", loaded.MaxResponseLines, 20)
	}
	if loaded.ShowTips {
		t.Error("ShowTips should be false")
	}
}

func TestSettings_DefaultWhenNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	s, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if s.Spinner != "dots" {
		t.Errorf("default Spinner = %q, want %q", s.Spinner, "dots")
	}
	if s.MaxResponseLines != 40 {
		t.Errorf("default MaxResponseLines = %d, want %d", s.MaxResponseLines, 40)
	}
	if !s.ShowTips {
		t.Error("default ShowTips should be true")
	}
}

func TestSettings_ZeroValuesGetDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	s := &Settings{
		Spinner:          "",
		MaxResponseLines: 0,
		ShowTips:         false,
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if loaded.Spinner != "dots" {
		t.Errorf("Spinner should default to 'dots', got %q", loaded.Spinner)
	}
	if loaded.MaxResponseLines != 40 {
		t.Errorf("MaxResponseLines should default to 40, got %d", loaded.MaxResponseLines)
	}
}

func TestSettings_LoadSettings_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Write invalid JSON
	settingsPath := filepath.Join(tmpDir, ".starclaw", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	// LoadSettings should return an error on parse failure
	s, err := LoadSettings()
	if err == nil {
		t.Fatal("LoadSettings() should return an error for corrupted settings.json")
	}
	if s.Spinner != "dots" {
		t.Errorf("default Spinner = %q, want %q", s.Spinner, "dots")
	}
}

func TestSettings_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	original := &Settings{
		Spinner:          "circle",
		MaxResponseLines: 10,
		ShowTips:         true,
	}
	if err := original.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}

	if original.Spinner != loaded.Spinner {
		t.Errorf("Spinner mismatch: %q vs %q", original.Spinner, loaded.Spinner)
	}
	if original.MaxResponseLines != loaded.MaxResponseLines {
		t.Errorf("MaxResponseLines mismatch: %d vs %d", original.MaxResponseLines, loaded.MaxResponseLines)
	}
	if original.ShowTips != loaded.ShowTips {
		t.Errorf("ShowTips mismatch: %v vs %v", original.ShowTips, loaded.ShowTips)
	}
}
