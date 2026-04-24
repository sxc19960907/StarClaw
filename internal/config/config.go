package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/starclaw/starclaw/internal/mcp"
	"gopkg.in/yaml.v3"
)

// MCPServerConfig is an alias for backward compatibility.
// The actual type is defined in the mcp package.
type MCPServerConfig = mcp.MCPServerConfig

// Config holds all configuration for StarClaw
type Config struct {
	Endpoint   string                       `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	APIKey     string                       `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	ModelTier  string                       `mapstructure:"model_tier" yaml:"model_tier" json:"model_tier"`
	Agent      AgentConfig                  `mapstructure:"agent" yaml:"agent" json:"agent"`
	Tools      ToolsConfig                  `mapstructure:"tools" yaml:"tools" json:"tools"`
	Audit      AuditConfig                  `mapstructure:"audit" yaml:"audit" json:"audit"`
	MCPServers map[string]mcp.MCPServerConfig `mapstructure:"mcp_servers" yaml:"mcp_servers,omitempty" json:"mcp_servers,omitempty"`
	Update     UpdateConfig                 `mapstructure:"update" yaml:"update,omitempty" json:"update,omitempty"`
}

// AgentConfig holds agent-specific settings
type AgentConfig struct {
	MaxIterations int     `mapstructure:"max_iterations" yaml:"max_iterations"`
	Temperature   float64 `mapstructure:"temperature" yaml:"temperature"`
	MaxTokens     int     `mapstructure:"max_tokens" yaml:"max_tokens"`
}

// ToolsConfig holds tool-specific settings
type ToolsConfig struct {
	BashTimeout      int      `mapstructure:"bash_timeout" yaml:"bash_timeout"`
	BashMaxOutput    int      `mapstructure:"bash_max_output" yaml:"bash_max_output"`
	ResultTruncation int      `mapstructure:"result_truncation" yaml:"result_truncation"`
	ArgsTruncation   int      `mapstructure:"args_truncation" yaml:"args_truncation"`
	Allowed          []string `mapstructure:"allowed" yaml:"allowed"`
	Denied           []string `mapstructure:"denied" yaml:"denied"`
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
	viper.SetDefault("model_tier", "medium")
	viper.SetDefault("agent.max_iterations", 25)
	viper.SetDefault("agent.temperature", 0)
	viper.SetDefault("agent.max_tokens", 8192)
	viper.SetDefault("tools.bash_timeout", 120)
	viper.SetDefault("tools.bash_max_output", 30000)
	viper.SetDefault("tools.result_truncation", 30000)
	viper.SetDefault("tools.args_truncation", 200)
	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("update.auto_check", true)
	viper.SetDefault("update.auto_install", false)
	viper.SetDefault("update.channel", "stable")
	viper.SetDefault("update.cache_ttl", "24h")

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
model_tier: "medium"

agent:
  max_iterations: 25
  temperature: 0
  max_tokens: 8192

tools:
  bash_timeout: 120
  bash_max_output: 30000
  result_truncation: 30000
  args_truncation: 200

audit:
  enabled: true

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
#   auto_install: false
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
	if cfg.Tools.ResultTruncation == 0 {
		cfg.Tools.ResultTruncation = 30000
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

// NeedsSetup returns true if configuration is incomplete
func NeedsSetup(cfg *Config) bool {
	return strings.TrimSpace(cfg.APIKey) == ""
}
