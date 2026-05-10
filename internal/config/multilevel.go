package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigLayer identifies where a config value came from.
type ConfigLayer int

const (
	LayerDefault ConfigLayer = iota
	LayerGlobal
	LayerProject
	LayerLocal
	LayerEnv
)

func (l ConfigLayer) String() string {
	switch l {
	case LayerDefault:
		return "default"
	case LayerGlobal:
		return "global"
	case LayerProject:
		return "project"
	case LayerLocal:
		return "local"
	case LayerEnv:
		return "env"
	}
	return "unknown"
}

// ConfigSource tracks which layer set each top-level config section.
type ConfigSource struct {
	Endpoint   ConfigLayer
	APIKey     ConfigLayer
	Provider   ConfigLayer
	Agent      ConfigLayer
	Tools      ConfigLayer
	MCPServers ConfigLayer
}

// LoadMultiLevel loads config from all three levels and merges them.
// Precedence: local > project > global > defaults.
// Returns the merged config and source tracking.
func LoadMultiLevel() (*Config, *ConfigSource, error) {
	source := &ConfigSource{}

	globalDir := StarclawDir()
	if globalDir == "" {
		return nil, nil, os.ErrNotExist
	}

	base := defaultConfig()

	globalPath := filepath.Join(globalDir, "config.yaml")
	if global, err := loadYAMLFile(globalPath); err == nil {
		overlayConfig(base, global, source, LayerGlobal)
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		projectPath := filepath.Join(cwd, ".starclaw", "config.yaml")
		if project, err := loadYAMLFile(projectPath); err == nil {
			overlayConfig(base, project, source, LayerProject)
		}

		localPath := filepath.Join(cwd, ".starclaw", "config.local.yaml")
		if local, err := loadYAMLFile(localPath); err == nil {
			overlayConfig(base, local, source, LayerLocal)
		}
	}

	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		base.OpenAIAPIKey = v
		source.APIKey = LayerEnv
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		base.OpenAIEndpoint = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		base.APIKey = v
		source.APIKey = LayerEnv
	}

	return base, source, nil
}

func defaultConfig() *Config {
	return &Config{
		Endpoint:       "https://api.anthropic.com",
		Provider:       "anthropic",
		OpenAIEndpoint: "https://api.openai.com/v1",
		OpenAIModel:    "gpt-4o",
		OllamaEndpoint: "http://localhost:11434",
		OllamaModel:    "llama3.1",
		ModelTier:      "medium",
		Agent: AgentConfig{
			MaxIterations:  25,
			MaxTokens:      8192,
			Thinking:       true,
			ThinkingMode:   "adaptive",
			ThinkingBudget: 10000,
		},
		Tools: ToolsConfig{
			BashTimeout:      120,
			BashMaxOutput:    30000,
			ResultTruncation: 30000,
			ArgsTruncation:   200,
			GrepMaxResults:   100,
		},
		Audit:  AuditConfig{Enabled: true},
		Update: UpdateConfig{AutoCheck: true, Channel: "stable", CacheTTL: "24h"},
		Cloud: CloudConfig{
			Timeout:       3600,
			MaxConcurrent: 3,
		},
	}
}

func loadYAMLFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func overlayConfig(base, overlay *Config, source *ConfigSource, layer ConfigLayer) {
	if overlay.Endpoint != "" {
		base.Endpoint = overlay.Endpoint
		source.Endpoint = layer
	}
	if overlay.APIKey != "" {
		base.APIKey = overlay.APIKey
		source.APIKey = layer
	}
	if overlay.Provider != "" {
		base.Provider = overlay.Provider
		source.Provider = layer
	}
	if overlay.OpenAIAPIKey != "" {
		base.OpenAIAPIKey = overlay.OpenAIAPIKey
	}
	if overlay.OpenAIEndpoint != "" {
		base.OpenAIEndpoint = overlay.OpenAIEndpoint
	}
	if overlay.OpenAIModel != "" {
		base.OpenAIModel = overlay.OpenAIModel
	}
	if overlay.OllamaEndpoint != "" {
		base.OllamaEndpoint = overlay.OllamaEndpoint
	}
	if overlay.OllamaModel != "" {
		base.OllamaModel = overlay.OllamaModel
	}
	if overlay.ModelTier != "" {
		base.ModelTier = overlay.ModelTier
	}

	overlayAgent(&base.Agent, &overlay.Agent, source, layer)
	overlayTools(&base.Tools, &overlay.Tools, source, layer)
	overlayCloud(&base.Cloud, &overlay.Cloud, layer)

	if len(overlay.MCPServers) > 0 {
		if base.MCPServers == nil {
			base.MCPServers = make(map[string]MCPServerConfig)
		}
		for k, v := range overlay.MCPServers {
			base.MCPServers[k] = v
		}
		source.MCPServers = layer
	}
}

func overlayAgent(base, overlay *AgentConfig, source *ConfigSource, layer ConfigLayer) {
	changed := false
	if overlay.MaxIterations != 0 {
		base.MaxIterations = overlay.MaxIterations
		changed = true
	}
	if overlay.Temperature != 0 {
		base.Temperature = overlay.Temperature
		changed = true
	}
	if overlay.MaxTokens != 0 {
		base.MaxTokens = overlay.MaxTokens
		changed = true
	}
	if overlay.ContextWindow != 0 {
		base.ContextWindow = overlay.ContextWindow
		changed = true
	}
	if overlay.ThinkingMode != "" {
		base.ThinkingMode = overlay.ThinkingMode
		base.Thinking = true
		changed = true
	}
	if overlay.ThinkingBudget != 0 {
		base.ThinkingBudget = overlay.ThinkingBudget
		changed = true
	}
	if overlay.ReasoningEffort != "" {
		base.ReasoningEffort = overlay.ReasoningEffort
		changed = true
	}
	if overlay.Model != "" {
		base.Model = overlay.Model
		changed = true
	}
	if changed {
		source.Agent = layer
	}
}

func overlayTools(base, overlay *ToolsConfig, source *ConfigSource, layer ConfigLayer) {
	changed := false
	if overlay.BashTimeout != 0 {
		base.BashTimeout = overlay.BashTimeout
		changed = true
	}
	if overlay.BashMaxOutput != 0 {
		base.BashMaxOutput = overlay.BashMaxOutput
		changed = true
	}
	if overlay.ResultTruncation != 0 {
		base.ResultTruncation = overlay.ResultTruncation
		changed = true
	}
	if overlay.ArgsTruncation != 0 {
		base.ArgsTruncation = overlay.ArgsTruncation
		changed = true
	}
	if overlay.GrepMaxResults != 0 {
		base.GrepMaxResults = overlay.GrepMaxResults
		changed = true
	}
	if len(overlay.Allowed) > 0 {
		base.Allowed = overlay.Allowed
		changed = true
	}
	if len(overlay.Denied) > 0 {
		base.Denied = overlay.Denied
		changed = true
	}
	if changed {
		source.Tools = layer
	}
}

func overlayCloud(base, overlay *CloudConfig, layer ConfigLayer) {
	if overlay.Enabled {
		base.Enabled = true
	}
	if overlay.Endpoint != "" {
		base.Endpoint = overlay.Endpoint
	}
	if overlay.APIKey != "" {
		base.APIKey = overlay.APIKey
	}
	if overlay.Timeout != 0 {
		base.Timeout = overlay.Timeout
	}
	if overlay.MaxConcurrent != 0 {
		base.MaxConcurrent = overlay.MaxConcurrent
	}
}
