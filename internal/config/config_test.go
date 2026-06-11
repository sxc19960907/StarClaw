package config

import (
	"os"
	"path/filepath"
	"strings"
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
	t.Setenv("HOME", tmpDir)

	// First load should create default config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Check defaults. Connector URL, model names, and keys stay user-supplied
	// so compatible providers do not receive fake launch-ready values.
	if cfg.Endpoint != "" {
		t.Errorf("Endpoint = %q, want empty user-supplied endpoint", cfg.Endpoint)
	}
	if cfg.ModelTier != "" {
		t.Errorf("ModelTier = %q, want empty user-supplied model", cfg.ModelTier)
	}
	if cfg.OpenAIModel != "" {
		t.Errorf("OpenAIModel = %q, want empty user-supplied model", cfg.OpenAIModel)
	}
	if cfg.OllamaModel != "" {
		t.Errorf("OllamaModel = %q, want empty user-supplied model", cfg.OllamaModel)
	}

	// Check Agent defaults
	if cfg.Agent.MaxIterations != 25 {
		t.Errorf("Agent.MaxIterations = %d, want 25", cfg.Agent.MaxIterations)
	}

	if cfg.Agent.Temperature != 0 {
		t.Errorf("Agent.Temperature = %f, want 0", cfg.Agent.Temperature)
	}

	if cfg.Sync.Enabled {
		t.Error("Sync.Enabled should default to false")
	}
	if cfg.Sync.DryRun {
		t.Error("Sync.DryRun should default to false")
	}
	if cfg.Sync.BatchMaxSessions != 25 {
		t.Errorf("Sync.BatchMaxSessions = %d, want 25", cfg.Sync.BatchMaxSessions)
	}
	if cfg.Sync.BatchMaxBytes != 5*1024*1024 {
		t.Errorf("Sync.BatchMaxBytes = %d, want 5242880", cfg.Sync.BatchMaxBytes)
	}
	if cfg.Sync.SingleSessionMaxBytes != 4*1024*1024 {
		t.Errorf("Sync.SingleSessionMaxBytes = %d, want 4194304", cfg.Sync.SingleSessionMaxBytes)
	}
	if cfg.Sync.DaemonInterval != "24h" {
		t.Errorf("Sync.DaemonInterval = %q, want 24h", cfg.Sync.DaemonInterval)
	}
	if cfg.Sync.LockTimeout != "30s" {
		t.Errorf("Sync.LockTimeout = %q, want 30s", cfg.Sync.LockTimeout)
	}
	if cfg.Agent.StreamIdleTimeoutSecs != 90 {
		t.Errorf("Agent.StreamIdleTimeoutSecs = %d, want 90", cfg.Agent.StreamIdleTimeoutSecs)
	}
}

func TestLoadStreamIdleTimeoutOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(`
api_key: "test-key"
agent:
  stream_idle_timeout_secs: 30
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.StreamIdleTimeoutSecs != 30 {
		t.Fatalf("StreamIdleTimeoutSecs = %d, want 30", cfg.Agent.StreamIdleTimeoutSecs)
	}
}

func TestLoadStreamIdleTimeoutExplicitZeroDisables(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(`
api_key: "test-key"
agent:
  stream_idle_timeout_secs: 0
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Agent.StreamIdleTimeoutSecs != 0 {
		t.Fatalf("StreamIdleTimeoutSecs = %d, want disabled zero", cfg.Agent.StreamIdleTimeoutSecs)
	}
}

func TestLoadRejectsNegativeStreamIdleTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(`
api_key: "test-key"
agent:
  stream_idle_timeout_secs: -1
`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative stream idle timeout")
	}
	if !strings.Contains(err.Error(), "agent.stream_idle_timeout_secs") {
		t.Fatalf("error = %v, want agent.stream_idle_timeout_secs", err)
	}
}

func TestLoadFromPathStreamIdleTimeoutDefaultAndExplicitZero(t *testing.T) {
	tmpDir := t.TempDir()
	missingPath := filepath.Join(tmpDir, "missing.yaml")
	if err := os.WriteFile(missingPath, []byte(`
api_key: "test-key"
`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFromPath(missingPath)
	if err != nil {
		t.Fatalf("LoadFromPath missing value: %v", err)
	}
	if cfg.Agent.StreamIdleTimeoutSecs != 90 {
		t.Fatalf("missing stream idle timeout = %d, want default 90", cfg.Agent.StreamIdleTimeoutSecs)
	}

	zeroPath := filepath.Join(tmpDir, "zero.yaml")
	if err := os.WriteFile(zeroPath, []byte(`
api_key: "test-key"
agent:
  stream_idle_timeout_secs: 0
`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadFromPath(zeroPath)
	if err != nil {
		t.Fatalf("LoadFromPath explicit zero: %v", err)
	}
	if cfg.Agent.StreamIdleTimeoutSecs != 0 {
		t.Fatalf("explicit zero stream idle timeout = %d, want disabled zero", cfg.Agent.StreamIdleTimeoutSecs)
	}
}

func TestSave(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Override home directory
	t.Setenv("HOME", tmpDir)

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
		name string
		cfg  *Config
		want bool
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

func TestLoadFromPathSyncConfig(t *testing.T) {
	configData := `
sync:
  enabled: true
  dry_run: true
  endpoint: "http://127.0.0.1/sync"
  exclude_agents: ["helper"]
  exclude_sources: ["remote"]
  batch_max_sessions: 3
  batch_max_bytes: 1024
  single_session_max_bytes: 512
  daemon_interval: "2h"
  daemon_startup_delay: "5s"
  failed_max_attempts_transient: 7
  lock_timeout: "250ms"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}
	if !cfg.Sync.Enabled || !cfg.Sync.DryRun {
		t.Fatalf("sync enabled/dry_run = %v/%v, want true/true", cfg.Sync.Enabled, cfg.Sync.DryRun)
	}
	if cfg.Sync.Endpoint != "http://127.0.0.1/sync" {
		t.Fatalf("sync endpoint = %q", cfg.Sync.Endpoint)
	}
	if cfg.Sync.BatchMaxSessions != 3 || cfg.Sync.BatchMaxBytes != 1024 || cfg.Sync.SingleSessionMaxBytes != 512 {
		t.Fatalf("sync caps = %+v", cfg.Sync)
	}
	if cfg.Sync.DaemonInterval != "2h" || cfg.Sync.DaemonStartupDelay != "5s" || cfg.Sync.LockTimeout != "250ms" {
		t.Fatalf("sync durations = %+v", cfg.Sync)
	}
	if cfg.Sync.FailedMaxAttemptsTransient != 7 {
		t.Fatalf("failed transient cap = %d, want 7", cfg.Sync.FailedMaxAttemptsTransient)
	}
}

func TestLoadFromPathSyncDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`provider: "anthropic"`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}
	if cfg.Sync.Enabled {
		t.Fatal("sync should default disabled")
	}
	if cfg.Sync.BatchMaxSessions != 25 || cfg.Sync.BatchMaxBytes != 5*1024*1024 {
		t.Fatalf("sync defaults = %+v", cfg.Sync)
	}
	if cfg.Sync.DaemonInterval != "24h" || cfg.Sync.LockTimeout != "30s" {
		t.Fatalf("sync duration defaults = %+v", cfg.Sync)
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
	t.Setenv("HOME", tmpDir)

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
