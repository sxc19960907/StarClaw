package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMultiLevel_GlobalOnly(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
endpoint: "https://custom.api.com"
api_key: "test-key"
agent:
  max_iterations: 50
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}

	if cfg.Endpoint != "https://custom.api.com" {
		t.Errorf("endpoint = %q, want custom", cfg.Endpoint)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("api_key = %q, want test-key", cfg.APIKey)
	}
	if cfg.Agent.MaxIterations != 50 {
		t.Errorf("max_iterations = %d, want 50", cfg.Agent.MaxIterations)
	}
	if source.Endpoint != LayerGlobal {
		t.Errorf("endpoint source = %v, want global", source.Endpoint)
	}
}

func TestLoadMultiLevel_ProjectOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Global config
	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
endpoint: "https://global.api.com"
api_key: "global-key"
agent:
  max_iterations: 25
`), 0600); err != nil {
		t.Fatal(err)
	}

	// Project config in cwd
	projectDir := filepath.Join(tmpDir, "project", ".starclaw")
	if err := os.MkdirAll(projectDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(`
agent:
  max_iterations: 100
  model: "claude-opus"
`), 0600); err != nil {
		t.Fatal(err)
	}

	// Change to project dir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "project")); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}

	// Global values preserved
	if cfg.Endpoint != "https://global.api.com" {
		t.Errorf("endpoint = %q, want global", cfg.Endpoint)
	}
	// Project overrides
	if cfg.Agent.MaxIterations != 100 {
		t.Errorf("max_iterations = %d, want 100", cfg.Agent.MaxIterations)
	}
	if cfg.Agent.Model != "claude-opus" {
		t.Errorf("model = %q, want claude-opus", cfg.Agent.Model)
	}
	if source.Agent != LayerProject {
		t.Errorf("agent source = %v, want project", source.Agent)
	}
}

func TestLoadMultiLevel_ToolsMCPServeOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
api_key: "global-key"
tools:
  server_tool_timeout: 10
  mcp_expose:
    - file_read
    - grep
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}

	if cfg.Tools.ServerToolTimeout != 10 {
		t.Errorf("server_tool_timeout = %d, want 10", cfg.Tools.ServerToolTimeout)
	}
	if len(cfg.Tools.MCPExpose) != 2 || cfg.Tools.MCPExpose[0] != "file_read" || cfg.Tools.MCPExpose[1] != "grep" {
		t.Errorf("mcp_expose = %#v, want [file_read grep]", cfg.Tools.MCPExpose)
	}
	if source.Tools != LayerGlobal {
		t.Errorf("tools source = %v, want global", source.Tools)
	}
}

func TestLoadMultiLevel_LocalOverridesProject(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
api_key: "global-key"
`), 0600); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "project", ".starclaw")
	if err := os.MkdirAll(projectDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte(`
agent:
  max_tokens: 4096
`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "config.local.yaml"), []byte(`
agent:
  max_tokens: 16384
`), 0600); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(tmpDir, "project")); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}

	if cfg.Agent.MaxTokens != 16384 {
		t.Errorf("max_tokens = %d, want 16384", cfg.Agent.MaxTokens)
	}
	if source.Agent != LayerLocal {
		t.Errorf("agent source = %v, want local", source.Agent)
	}
}

func TestLoadMultiLevel_TokenBudgetOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
api_key: "global-key"
agent:
  token_budget:
    max_input_tokens: 100
    max_total_tokens: 300
    hard_stop: true
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}
	if cfg.Agent.TokenBudget.MaxInputTokens != 100 {
		t.Fatalf("MaxInputTokens = %d, want 100", cfg.Agent.TokenBudget.MaxInputTokens)
	}
	if cfg.Agent.TokenBudget.MaxTotalTokens != 300 {
		t.Fatalf("MaxTotalTokens = %d, want 300", cfg.Agent.TokenBudget.MaxTotalTokens)
	}
	if !cfg.Agent.TokenBudget.HardStop {
		t.Fatal("HardStop = false, want true")
	}
	if source.Agent != LayerGlobal {
		t.Fatalf("agent source = %v, want global", source.Agent)
	}
}

func TestLoadMultiLevel_SyncDefaultsAndOverlay(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
sync:
  enabled: true
  dry_run: true
  endpoint: "http://127.0.0.1/sync"
  exclude_agents: ["helper"]
  batch_max_sessions: 4
  lock_timeout: "250ms"
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}
	if !cfg.Sync.Enabled || !cfg.Sync.DryRun {
		t.Fatalf("sync enabled/dry_run = %v/%v, want true/true", cfg.Sync.Enabled, cfg.Sync.DryRun)
	}
	if cfg.Sync.Endpoint != "http://127.0.0.1/sync" {
		t.Fatalf("sync endpoint = %q", cfg.Sync.Endpoint)
	}
	if len(cfg.Sync.ExcludeAgents) != 1 || cfg.Sync.ExcludeAgents[0] != "helper" {
		t.Fatalf("sync exclude agents = %v", cfg.Sync.ExcludeAgents)
	}
	if cfg.Sync.BatchMaxSessions != 4 {
		t.Fatalf("sync batch max sessions = %d, want 4", cfg.Sync.BatchMaxSessions)
	}
	if cfg.Sync.BatchMaxBytes != 5*1024*1024 {
		t.Fatalf("sync batch max bytes = %d, want default", cfg.Sync.BatchMaxBytes)
	}
	if cfg.Sync.LockTimeout != "250ms" {
		t.Fatalf("sync lock timeout = %q", cfg.Sync.LockTimeout)
	}
	if source.Sync != LayerGlobal {
		t.Fatalf("sync source = %v, want global", source.Sync)
	}
}

func TestLoadMultiLevel_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("ANTHROPIC_API_KEY", "env-key")

	globalDir := filepath.Join(tmpDir, ".starclaw")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(`
api_key: "file-key"
`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, source, err := LoadMultiLevel()
	if err != nil {
		t.Fatalf("LoadMultiLevel: %v", err)
	}

	if cfg.APIKey != "env-key" {
		t.Errorf("api_key = %q, want env-key", cfg.APIKey)
	}
	if source.APIKey != LayerEnv {
		t.Errorf("api_key source = %v, want env", source.APIKey)
	}
}

func TestConfigLayer_String(t *testing.T) {
	tests := []struct {
		layer ConfigLayer
		want  string
	}{
		{LayerDefault, "default"},
		{LayerGlobal, "global"},
		{LayerProject, "project"},
		{LayerLocal, "local"},
		{LayerEnv, "env"},
		{ConfigLayer(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.layer.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.layer, got, tt.want)
		}
	}
}
