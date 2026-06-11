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
	Sync       ConfigLayer
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
	if global, presence, err := loadYAMLFile(globalPath); err == nil {
		overlayConfig(base, global, presence, source, LayerGlobal)
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		projectPath := filepath.Join(cwd, ".starclaw", "config.yaml")
		if project, presence, err := loadYAMLFile(projectPath); err == nil {
			overlayConfig(base, project, presence, source, LayerProject)
		}

		localPath := filepath.Join(cwd, ".starclaw", "config.local.yaml")
		if local, presence, err := loadYAMLFile(localPath); err == nil {
			overlayConfig(base, local, presence, source, LayerLocal)
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
		Provider: "anthropic",
		Agent: AgentConfig{
			MaxIterations:         25,
			MaxTokens:             8192,
			StreamIdleTimeoutSecs: 90,
			Thinking:              true,
			ThinkingMode:          "adaptive",
			ThinkingBudget:        10000,
		},
		Tools: ToolsConfig{
			BashTimeout:       120,
			BashMaxOutput:     30000,
			ResultTruncation:  30000,
			ArgsTruncation:    200,
			GrepMaxResults:    100,
			ServerToolTimeout: 0,
		},
		Audit:  AuditConfig{Enabled: true},
		Update: UpdateConfig{AutoCheck: true, Channel: "stable", CacheTTL: "24h"},
		Cloud: CloudConfig{
			Timeout:       3600,
			MaxConcurrent: 3,
		},
		Sync: SyncConfig{
			BatchMaxSessions:           25,
			BatchMaxBytes:              5 * 1024 * 1024,
			SingleSessionMaxBytes:      4 * 1024 * 1024,
			DaemonInterval:             "24h",
			DaemonStartupDelay:         "60s",
			FailedMaxAttemptsTransient: 5,
			LockTimeout:                "30s",
		},
	}
}

type configPresence struct {
	AgentStreamIdleTimeout bool
}

func loadYAMLFile(path string) (*Config, configPresence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, configPresence{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, configPresence{}, err
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, configPresence{}, err
	}
	presence := configPresence{
		AgentStreamIdleTimeout: yamlPathExists(&root, "agent", "stream_idle_timeout_secs"),
	}
	return &cfg, presence, nil
}

func yamlPathExists(root *yaml.Node, path ...string) bool {
	if root == nil || len(path) == 0 {
		return false
	}
	node := root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	for _, key := range path {
		if node.Kind != yaml.MappingNode {
			return false
		}
		found := false
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				node = node.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func overlayConfig(base, overlay *Config, presence configPresence, source *ConfigSource, layer ConfigLayer) {
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

	overlayAgent(&base.Agent, &overlay.Agent, presence, source, layer)
	overlayTools(&base.Tools, &overlay.Tools, source, layer)
	overlayCloud(&base.Cloud, &overlay.Cloud, layer)
	overlaySync(&base.Sync, &overlay.Sync, source, layer)

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

func overlayAgent(base, overlay *AgentConfig, presence configPresence, source *ConfigSource, layer ConfigLayer) {
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
	if overlay.StreamIdleTimeoutSecs != 0 || presence.AgentStreamIdleTimeout {
		base.StreamIdleTimeoutSecs = overlay.StreamIdleTimeoutSecs
		changed = true
	}
	if overlay.TokenBudget.MaxInputTokens != 0 {
		base.TokenBudget.MaxInputTokens = overlay.TokenBudget.MaxInputTokens
		changed = true
	}
	if overlay.TokenBudget.MaxOutputTokens != 0 {
		base.TokenBudget.MaxOutputTokens = overlay.TokenBudget.MaxOutputTokens
		changed = true
	}
	if overlay.TokenBudget.MaxTotalTokens != 0 {
		base.TokenBudget.MaxTotalTokens = overlay.TokenBudget.MaxTotalTokens
		changed = true
	}
	if overlay.TokenBudget.HardStop {
		base.TokenBudget.HardStop = overlay.TokenBudget.HardStop
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
	if overlay.ServerToolTimeout != 0 {
		base.ServerToolTimeout = overlay.ServerToolTimeout
		changed = true
	}
	if len(overlay.MCPExpose) > 0 {
		base.MCPExpose = overlay.MCPExpose
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

func overlaySync(base, overlay *SyncConfig, source *ConfigSource, layer ConfigLayer) {
	changed := false
	if overlay.Enabled {
		base.Enabled = true
		changed = true
	}
	if overlay.DryRun {
		base.DryRun = true
		changed = true
	}
	if overlay.Endpoint != "" {
		base.Endpoint = overlay.Endpoint
		changed = true
	}
	if len(overlay.ExcludeAgents) > 0 {
		base.ExcludeAgents = append([]string(nil), overlay.ExcludeAgents...)
		changed = true
	}
	if len(overlay.ExcludeSources) > 0 {
		base.ExcludeSources = append([]string(nil), overlay.ExcludeSources...)
		changed = true
	}
	if overlay.BatchMaxSessions != 0 {
		base.BatchMaxSessions = overlay.BatchMaxSessions
		changed = true
	}
	if overlay.BatchMaxBytes != 0 {
		base.BatchMaxBytes = overlay.BatchMaxBytes
		changed = true
	}
	if overlay.SingleSessionMaxBytes != 0 {
		base.SingleSessionMaxBytes = overlay.SingleSessionMaxBytes
		changed = true
	}
	if overlay.DaemonInterval != "" {
		base.DaemonInterval = overlay.DaemonInterval
		changed = true
	}
	if overlay.DaemonStartupDelay != "" {
		base.DaemonStartupDelay = overlay.DaemonStartupDelay
		changed = true
	}
	if overlay.FailedMaxAttemptsTransient != 0 {
		base.FailedMaxAttemptsTransient = overlay.FailedMaxAttemptsTransient
		changed = true
	}
	if overlay.LockTimeout != "" {
		base.LockTimeout = overlay.LockTimeout
		changed = true
	}
	if changed {
		source.Sync = layer
	}
}
