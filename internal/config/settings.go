package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings holds non-sensitive UX settings for the StarClaw CLI.
type Settings struct {
	Spinner          string `json:"spinner"`
	MaxResponseLines int    `json:"max_response_lines"`
	ShowTips         bool   `json:"show_tips"`
}

// DefaultSettings returns Settings with sensible defaults.
func DefaultSettings() *Settings {
	return &Settings{
		Spinner:          "dots",
		MaxResponseLines: 40,
		ShowTips:         true,
	}
}

// Save writes settings to ~/.starclaw/settings.json.
func (s *Settings) Save() error {
	dir := StarclawDir()
	if dir == "" {
		return fmt.Errorf("failed to resolve home directory")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}
	path := filepath.Join(dir, "settings.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

// LoadSettings reads settings from ~/.starclaw/settings.json.
// Returns default settings if the file does not exist, or if parsing fails.
func LoadSettings() (*Settings, error) {
	dir := StarclawDir()
	if dir == "" {
		return DefaultSettings(), nil
	}
	path := filepath.Join(dir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), fmt.Errorf("read settings: %w", err)
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), fmt.Errorf("parse settings: %w", err)
	}
	// Fill zero values with defaults so partial files work.
	if s.Spinner == "" {
		s.Spinner = "dots"
	}
	if s.MaxResponseLines <= 0 {
		s.MaxResponseLines = 40
	}
	return &s, nil
}
