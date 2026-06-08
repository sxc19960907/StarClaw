package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/starclaw/starclaw/internal/hooks"
	"github.com/starclaw/starclaw/internal/mcp"
	"github.com/starclaw/starclaw/internal/permissions"
	"gopkg.in/yaml.v3"
)

// MCPServerConfig is an alias for backward compatibility.
// The actual type is defined in the mcp package.
type MCPServerConfig = mcp.MCPServerConfig

// Config holds all configuration for StarClaw
type Config struct {
	Endpoint       string                         `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	APIKey         string                         `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	Provider       string                         `mapstructure:"provider"         yaml:"provider"         json:"provider"`
	OpenAIAPIKey   string                         `mapstructure:"openai_api_key"   yaml:"openai_api_key"   json:"openai_api_key"`
	OpenAIEndpoint string                         `mapstructure:"openai_endpoint"  yaml:"openai_endpoint"  json:"openai_endpoint"`
	OpenAIModel    string                         `mapstructure:"openai_model"     yaml:"openai_model"     json:"openai_model"`
	OllamaEndpoint string                         `mapstructure:"ollama_endpoint"  yaml:"ollama_endpoint"  json:"ollama_endpoint"`
	OllamaModel    string                         `mapstructure:"ollama_model"     yaml:"ollama_model"     json:"ollama_model"`
	ModelTier      string                         `mapstructure:"model_tier" yaml:"model_tier" json:"model_tier"`
	Agent          AgentConfig                    `mapstructure:"agent" yaml:"agent" json:"agent"`
	Tools          ToolsConfig                    `mapstructure:"tools" yaml:"tools" json:"tools"`
	Audit          AuditConfig                    `mapstructure:"audit" yaml:"audit" json:"audit"`
	MCPServers     map[string]mcp.MCPServerConfig `mapstructure:"mcp_servers" yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	Update         UpdateConfig                   `mapstructure:"update" yaml:"update,omitempty" json:"update,omitempty"`
	Cloud          CloudConfig                    `mapstructure:"cloud" yaml:"cloud,omitempty" json:"cloud,omitempty"`
	Sync           SyncConfig                     `mapstructure:"sync" yaml:"sync,omitempty" json:"sync,omitempty"`
	Permissions    *permissions.Config            `mapstructure:"permissions" yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Hooks          *hooks.Config                  `mapstructure:"hooks" yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

// AgentConfig holds agent-specific settings
type AgentConfig struct {
	MaxIterations         int               `mapstructure:"max_iterations" yaml:"max_iterations"`
	Temperature           float64           `mapstructure:"temperature" yaml:"temperature"`
	MaxTokens             int               `mapstructure:"max_tokens" yaml:"max_tokens"`
	ContextWindow         int               `mapstructure:"context_window" yaml:"context_window"`
	StreamIdleTimeoutSecs int               `mapstructure:"stream_idle_timeout_secs" yaml:"stream_idle_timeout_secs" json:"stream_idle_timeout_secs"`
	TokenBudget           TokenBudgetConfig `mapstructure:"token_budget" yaml:"token_budget,omitempty" json:"token_budget,omitempty"`
	Thinking              bool              `mapstructure:"thinking"         yaml:"thinking"         json:"thinking"`
	ThinkingMode          string            `mapstructure:"thinking_mode"    yaml:"thinking_mode"    json:"thinking_mode"`
	ThinkingBudget        int               `mapstructure:"thinking_budget"  yaml:"thinking_budget"  json:"thinking_budget"`
	ReasoningEffort       string            `mapstructure:"reasoning_effort" yaml:"reasoning_effort" json:"reasoning_effort"`
	Model                 string            `mapstructure:"model"            yaml:"model"            json:"model"`
}

// TokenBudgetConfig configures local per-run token budget enforcement.
type TokenBudgetConfig struct {
	MaxInputTokens  int  `mapstructure:"max_input_tokens" yaml:"max_input_tokens,omitempty" json:"max_input_tokens,omitempty"`
	MaxOutputTokens int  `mapstructure:"max_output_tokens" yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`
	MaxTotalTokens  int  `mapstructure:"max_total_tokens" yaml:"max_total_tokens,omitempty" json:"max_total_tokens,omitempty"`
	HardStop        bool `mapstructure:"hard_stop" yaml:"hard_stop,omitempty" json:"hard_stop,omitempty"`
}

// ToolsConfig holds tool-specific settings
type ToolsConfig struct {
	BashTimeout       int      `mapstructure:"bash_timeout" yaml:"bash_timeout"`
	BashMaxOutput     int      `mapstructure:"bash_max_output" yaml:"bash_max_output"`
	ResultTruncation  int      `mapstructure:"result_truncation" yaml:"result_truncation"`
	ArgsTruncation    int      `mapstructure:"args_truncation" yaml:"args_truncation"`
	GrepMaxResults    int      `mapstructure:"grep_max_results" yaml:"grep_max_results"`
	ServerToolTimeout int      `mapstructure:"server_tool_timeout" yaml:"server_tool_timeout" json:"server_tool_timeout"`
	MCPExpose         []string `mapstructure:"mcp_expose" yaml:"mcp_expose,omitempty" json:"mcp_expose,omitempty"`
	Allowed           []string `mapstructure:"allowed" yaml:"allowed"`
	Denied            []string `mapstructure:"denied" yaml:"denied"`
}

// UpdateConfig holds auto-update settings
type UpdateConfig struct {
	AutoCheck   bool   `mapstructure:"auto_check" yaml:"auto_check"`
	AutoInstall bool   `mapstructure:"auto_install" yaml:"auto_install"`
	Channel     string `mapstructure:"channel" yaml:"channel"`
	CacheTTL    string `mapstructure:"cache_ttl" yaml:"cache_ttl"`
}

// AuditConfig holds audit logging settings
type AuditConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
}

// CloudConfig holds cloud agent delegation settings
type CloudConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Endpoint      string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	APIKey        string `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	Timeout       int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	MaxConcurrent int    `mapstructure:"max_concurrent" yaml:"max_concurrent" json:"max_concurrent"`
}

// SyncConfig holds local session sync settings. Sync remains disabled by
// default and future cloud uploader work must explicitly opt in.
type SyncConfig struct {
	Enabled                    bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	DryRun                     bool     `mapstructure:"dry_run" yaml:"dry_run" json:"dry_run"`
	Endpoint                   string   `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	ExcludeAgents              []string `mapstructure:"exclude_agents" yaml:"exclude_agents,omitempty" json:"exclude_agents,omitempty"`
	ExcludeSources             []string `mapstructure:"exclude_sources" yaml:"exclude_sources,omitempty" json:"exclude_sources,omitempty"`
	BatchMaxSessions           int      `mapstructure:"batch_max_sessions" yaml:"batch_max_sessions" json:"batch_max_sessions"`
	BatchMaxBytes              int      `mapstructure:"batch_max_bytes" yaml:"batch_max_bytes" json:"batch_max_bytes"`
	SingleSessionMaxBytes      int      `mapstructure:"single_session_max_bytes" yaml:"single_session_max_bytes" json:"single_session_max_bytes"`
	DaemonInterval             string   `mapstructure:"daemon_interval" yaml:"daemon_interval" json:"daemon_interval"`
	DaemonStartupDelay         string   `mapstructure:"daemon_startup_delay" yaml:"daemon_startup_delay" json:"daemon_startup_delay"`
	FailedMaxAttemptsTransient int      `mapstructure:"failed_max_attempts_transient" yaml:"failed_max_attempts_transient" json:"failed_max_attempts_transient"`
	LockTimeout                string   `mapstructure:"lock_timeout" yaml:"lock_timeout" json:"lock_timeout"`
}

// StarclawDir returns the StarClaw configuration directory
func StarclawDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".starclaw")
}

// Load loads configuration from files
func Load() (*Config, error) {
	dir := StarclawDir()
	if dir == "" {
		return nil, fmt.Errorf("failed to resolve home directory")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	// Set defaults
	viper.SetDefault("endpoint", "https://api.anthropic.com")
	viper.SetDefault("provider", "anthropic")
	viper.SetDefault("openai_api_key", "")
	viper.SetDefault("openai_endpoint", "https://api.openai.com/v1")
	viper.SetDefault("openai_model", "gpt-4o")
	viper.SetDefault("ollama_endpoint", "http://localhost:11434")
	viper.SetDefault("ollama_model", "llama3.1")
	viper.SetDefault("model_tier", "medium")
	viper.SetDefault("agent.max_iterations", 25)
	viper.SetDefault("agent.temperature", 0)
	viper.SetDefault("agent.max_tokens", 8192)
	viper.SetDefault("agent.context_window", 0) // 0 = auto/disabled
	viper.SetDefault("agent.stream_idle_timeout_secs", 90)
	viper.SetDefault("agent.token_budget.max_input_tokens", 0)
	viper.SetDefault("agent.token_budget.max_output_tokens", 0)
	viper.SetDefault("agent.token_budget.max_total_tokens", 0)
	viper.SetDefault("agent.token_budget.hard_stop", false)
	viper.SetDefault("agent.thinking", true)
	viper.SetDefault("agent.thinking_mode", "adaptive")
	viper.SetDefault("agent.thinking_budget", 10000)
	viper.SetDefault("agent.reasoning_effort", "")
	viper.SetDefault("agent.model", "")
	viper.SetDefault("tools.bash_timeout", 120)
	viper.SetDefault("tools.bash_max_output", 30000)
	viper.SetDefault("tools.result_truncation", 30000)
	viper.SetDefault("tools.args_truncation", 200)
	viper.SetDefault("tools.grep_max_results", 100)
	viper.SetDefault("tools.server_tool_timeout", 0)
	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("update.auto_check", true)
	viper.SetDefault("update.auto_install", false)
	viper.SetDefault("update.channel", "stable")
	viper.SetDefault("update.cache_ttl", "24h")
	setSyncDefaults(viper.GetViper())

	// Bind environment variables
	if err := viper.BindEnv("openai_api_key", "OPENAI_API_KEY"); err != nil {
		return nil, fmt.Errorf("bind OPENAI_API_KEY: %w", err)
	}
	if err := viper.BindEnv("openai_endpoint", "OPENAI_BASE_URL"); err != nil {
		return nil, fmt.Errorf("bind OPENAI_BASE_URL: %w", err)
	}

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, will create default
			if err := SaveDefault(dir); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			// Re-read after creating
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read created config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Load local config if exists (project-level override)
	if err := loadLocalConfig(viper.GetViper()); err != nil {
		// Non-fatal: local config is optional
		_ = err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Trim API key
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)

	// Validate thinking_mode
	if cfg.Agent.Thinking {
		switch cfg.Agent.ThinkingMode {
		case "adaptive", "enabled":
			// valid
		default:
			return nil, fmt.Errorf("invalid agent.thinking_mode %q: must be \"adaptive\" or \"enabled\"", cfg.Agent.ThinkingMode)
		}
	}
	if cfg.Agent.StreamIdleTimeoutSecs < 0 {
		return nil, fmt.Errorf("agent.stream_idle_timeout_secs (%d) must be >= 0 (0 = disabled)", cfg.Agent.StreamIdleTimeoutSecs)
	}

	return &cfg, nil
}

// loadLocalConfig loads local config from current working directory
func loadLocalConfig(v *viper.Viper) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	localPath := filepath.Join(cwd, ".starclaw", "config.local.yaml")
	if _, err := os.Stat(localPath); err != nil {
		return err // File doesn't exist
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}

	// Merge into viper
	var localCfg map[string]interface{}
	if err := yaml.Unmarshal(data, &localCfg); err != nil {
		return err
	}

	for key, value := range localCfg {
		v.Set(key, value)
	}

	return nil
}

// SaveDefault creates a default configuration file
func SaveDefault(dir string) error {
	configPath := filepath.Join(dir, "config.yaml")

	defaultConfig := `endpoint: "https://api.anthropic.com"
api_key: ""
provider: "anthropic"  # "anthropic", "openai", or "ollama"
model_tier: "medium"

# OpenAI configuration (used when provider is "openai")
# openai_api_key: ""
# openai_endpoint: "https://api.openai.com/v1"
# openai_model: "gpt-4o"

# Ollama configuration (used when provider is "ollama")
# ollama_endpoint: "http://localhost:11434"
# ollama_model: "llama3.1"

agent:
  max_iterations: 25
  temperature: 0
  max_tokens: 8192
  context_window: 0  # 0 = disabled, set to e.g. 200000 to enable compaction
  stream_idle_timeout_secs: 90  # provider stream chunk-gap watchdog; 0 disables
  token_budget:
    max_input_tokens: 0   # 0 = disabled
    max_output_tokens: 0  # 0 = disabled
    max_total_tokens: 0   # 0 = disabled
    hard_stop: false
  thinking: true
  thinking_mode: "adaptive"  # "adaptive" or "enabled"
  thinking_budget: 10000
  reasoning_effort: ""
  model: ""  # empty = use model_tier

tools:
  bash_timeout: 120
  bash_max_output: 30000
  result_truncation: 30000
  args_truncation: 200
  grep_max_results: 100
  server_tool_timeout: 0  # seconds; 0 disables MCP server tool timeout
  # mcp_expose: ["file_read", "grep", "version"]  # optional allow-list for starclaw mcp serve

audit:
  enabled: true

# Local session sync foundation. Disabled by default; dry-run writes only local
# outbox files and no cloud upload is configured here.
sync:
  enabled: false
  dry_run: false
  endpoint: ""
  batch_max_sessions: 25
  batch_max_bytes: 5242880
  single_session_max_bytes: 4194304
  daemon_interval: "24h"
  daemon_startup_delay: "60s"
  failed_max_attempts_transient: 5
  lock_timeout: "30s"

# MCP servers configuration (optional)
# mcp_servers:
#   github:
#     command: npx
#     args: ["-y", "@modelcontextprotocol/server-github"]
#     env:
#       GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_TOKEN}
#     keep_alive: true

# Update configuration (optional)
# update:
#   auto_check: true
#   auto_install: false  # do not install automatically on startup
#   channel: stable
#   cache_ttl: 24h
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	dir := StarclawDir()
	if dir == "" {
		return fmt.Errorf("failed to resolve home directory")
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(dir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// LoadFromPath loads configuration from a specific file path (used for testing)
func LoadFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to inspect config: %w", err)
	}

	// Set defaults for missing values
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.anthropic.com"
	}
	if cfg.ModelTier == "" {
		cfg.ModelTier = "medium"
	}
	if cfg.Agent.MaxIterations == 0 {
		cfg.Agent.MaxIterations = 25
	}
	if cfg.Agent.MaxTokens == 0 {
		cfg.Agent.MaxTokens = 8192
	}
	if cfg.Agent.StreamIdleTimeoutSecs == 0 && !yamlPathExists(&root, "agent", "stream_idle_timeout_secs") {
		cfg.Agent.StreamIdleTimeoutSecs = 90
	}
	if cfg.Tools.ResultTruncation == 0 {
		cfg.Tools.ResultTruncation = 30000
	}
	applySyncDefaults(&cfg.Sync)
	// thinking is tricky since bool zero value is false — check if it was set
	// For simplicity, set default thinking to true when thinking_mode is set
	if cfg.Agent.ThinkingMode == "" {
		cfg.Agent.ThinkingMode = "adaptive"
	}
	if cfg.Agent.ThinkingBudget == 0 {
		cfg.Agent.ThinkingBudget = 10000
	}
	if cfg.Agent.Model == "" {
		cfg.Agent.Model = cfg.ModelTier
	}

	// Audit is enabled by default
	// (bool zero value is false, but we want true as default)
	// Since YAML parsing doesn't distinguish between false and missing,
	// we use a different approach: if not explicitly set, we enable it
	// For now, we assume it's enabled unless explicitly disabled

	// Trim API key
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)

	return &cfg, nil
}

func setSyncDefaults(v *viper.Viper) {
	v.SetDefault("sync.enabled", false)
	v.SetDefault("sync.dry_run", false)
	v.SetDefault("sync.endpoint", "")
	v.SetDefault("sync.exclude_agents", []string{})
	v.SetDefault("sync.exclude_sources", []string{})
	v.SetDefault("sync.batch_max_sessions", 25)
	v.SetDefault("sync.batch_max_bytes", 5*1024*1024)
	v.SetDefault("sync.single_session_max_bytes", 4*1024*1024)
	v.SetDefault("sync.daemon_interval", "24h")
	v.SetDefault("sync.daemon_startup_delay", "60s")
	v.SetDefault("sync.failed_max_attempts_transient", 5)
	v.SetDefault("sync.lock_timeout", "30s")
}

func applySyncDefaults(sync *SyncConfig) {
	if sync.BatchMaxSessions == 0 {
		sync.BatchMaxSessions = 25
	}
	if sync.BatchMaxBytes == 0 {
		sync.BatchMaxBytes = 5 * 1024 * 1024
	}
	if sync.SingleSessionMaxBytes == 0 {
		sync.SingleSessionMaxBytes = 4 * 1024 * 1024
	}
	if sync.DaemonInterval == "" {
		sync.DaemonInterval = "24h"
	}
	if sync.DaemonStartupDelay == "" {
		sync.DaemonStartupDelay = "60s"
	}
	if sync.FailedMaxAttemptsTransient == 0 {
		sync.FailedMaxAttemptsTransient = 5
	}
	if sync.LockTimeout == "" {
		sync.LockTimeout = "30s"
	}
}

// NeedsSetup returns true if configuration is incomplete
func NeedsSetup(cfg *Config) bool {
	switch cfg.Provider {
	case "openai":
		return strings.TrimSpace(cfg.OpenAIAPIKey) == ""
	case "ollama":
		return false // local server, no API key needed
	default:
		return strings.TrimSpace(cfg.APIKey) == ""
	}
}
