package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStarclawDir(t *testing.T) {
	dir := StarclawDir()
	if dir == "" {
		t.Error("StarclawDir() should not return empty string")
	}

	// Should contain .starclaw
	if !contains(dir, ".starclaw") {
		t.Errorf("StarclawDir() should contain '.starclaw', got: %s", dir)
	}
}

func TestSaveDefault(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	err := SaveDefault(tmpDir)
	if err != nil {
		t.Fatalf("SaveDefault() error = %v", err)
	}

	// Check file was created
	configPath := filepath.Join(tmpDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("SaveDefault() did not create config.yaml")
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", mode)
	}
}

func TestLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Override home directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// First load should create default config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Check defaults
	if cfg.Endpoint == "" {
		t.Error("Load() returned empty endpoint")
	}

	if cfg.ModelTier == "" {
		t.Error("Load() returned empty model_tier")
	}

	// Check Agent defaults
	if cfg.Agent.MaxIterations != 25 {
		t.Errorf("Agent.MaxIterations = %d, want 25", cfg.Agent.MaxIterations)
	}

	if cfg.Agent.Temperature != 0 {
		t.Errorf("Agent.Temperature = %f, want 0", cfg.Agent.Temperature)
	}
}

func TestSave(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Override home directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		Endpoint:  "https://test.example.com",
		APIKey:    "sk-test123",
		ModelTier: "test",
		Agent: AgentConfig{
			MaxIterations: 10,
			Temperature:   0.5,
			MaxTokens:     4096,
		},
		Tools: ToolsConfig{
			BashTimeout:      60,
			BashMaxOutput:    10000,
			ResultTruncation: 10000,
			ArgsTruncation:   100,
		},
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.Endpoint != cfg.Endpoint {
		t.Errorf("Loaded Endpoint = %s, want %s", loaded.Endpoint, cfg.Endpoint)
	}

	if loaded.APIKey != cfg.APIKey {
		t.Errorf("Loaded APIKey = %s, want %s", loaded.APIKey, cfg.APIKey)
	}

	if loaded.Agent.MaxIterations != cfg.Agent.MaxIterations {
		t.Errorf("Loaded MaxIterations = %d, want %d", loaded.Agent.MaxIterations, cfg.Agent.MaxIterations)
	}
}

func TestNeedsSetup(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		want    bool
	}{
		{
			name: "Empty API key",
			cfg:  &Config{APIKey: ""},
			want: true,
		},
		{
			name: "Has API key",
			cfg:  &Config{APIKey: "sk-123"},
			want: false,
		},
		{
			name: "Whitespace API key",
			cfg:  &Config{APIKey: "   "},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsSetup(tt.cfg)
			if got != tt.want {
				t.Errorf("NeedsSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, len(s)-len(substr)))
}

func containsAt(s, substr string, start int) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for new MCP and Update config

func TestMCPServerConfig_Validation(t *testing.T) {
	configData := `
endpoint: "https://api.anthropic.com"
api_key: "test-key"
mcp_servers:
  github:
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-github"
    env:
      GITHUB_TOKEN: "secret"
    keep_alive: true
  disabled_server:
    command: echo
    disabled: true
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() failed: %v", err)
	}

	// Check MCP servers
	if len(cfg.MCPServers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(cfg.MCPServers))
	}

	github, ok := cfg.MCPServers["github"]
	if !ok {
		t.Fatal("Missing 'github' MCP server")
	}
	if github.Command != "npx" {
		t.Errorf("GitHub command wrong: got %q, want %q", github.Command, "npx")
	}
	if len(github.Args) != 2 {
		t.Errorf("GitHub args wrong length: got %d, want 2", len(github.Args))
	}
	if !github.KeepAlive {
		t.Error("GitHub KeepAlive should be true")
	}
	if github.Env["GITHUB_TOKEN"] != "secret" {
		t.Error("GitHub env not parsed correctly")
	}

	disabled := cfg.MCPServers["disabled_server"]
	if !disabled.Disabled {
		t.Error("disabled_server should be disabled")
	}
}

func TestUpdateConfig_Validation(t *testing.T) {
	configData := `
endpoint: "https://api.anthropic.com"
api_key: "test-key"
update:
  auto_check: false
  auto_install: true
  channel: beta
  cache_ttl: "48h"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() failed: %v", err)
	}

	if cfg.Update.AutoCheck {
		t.Error("Update.AutoCheck should be false")
	}
	if !cfg.Update.AutoInstall {
		t.Error("Update.AutoInstall should be true")
	}
	if cfg.Update.Channel != "beta" {
		t.Errorf("Update.Channel wrong: got %q, want %q", cfg.Update.Channel, "beta")
	}
	if cfg.Update.CacheTTL != "48h" {
		t.Errorf("Update.CacheTTL wrong: got %q, want %q", cfg.Update.CacheTTL, "48h")
	}
}

func TestConfig_UpdateDefaults(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Load config (should create defaults)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check new defaults
	if cfg.Update.Channel != "stable" {
		t.Errorf("Update.Channel default wrong: got %q, want %q", cfg.Update.Channel, "stable")
	}
	if cfg.Update.CacheTTL != "24h" {
		t.Errorf("Update.CacheTTL default wrong: got %q, want %q", cfg.Update.CacheTTL, "24h")
	}
	if cfg.Update.AutoInstall {
		t.Error("Update.AutoInstall should default to false")
	}
}

