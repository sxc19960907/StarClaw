package daemon

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/mcp"
	"github.com/starclaw/starclaw/internal/permissions"
)

var mcpServerNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

type daemonConfigView struct {
	Provider        string          `json:"provider"`
	Endpoint        string          `json:"endpoint"`
	ModelTier       string          `json:"model_tier"`
	OpenAIEndpoint  string          `json:"openai_endpoint"`
	OpenAIModel     string          `json:"openai_model"`
	OllamaEndpoint  string          `json:"ollama_endpoint"`
	OllamaModel     string          `json:"ollama_model"`
	APIKeySet       bool            `json:"api_key_set"`
	OpenAIAPIKeySet bool            `json:"openai_api_key_set"`
	MCPServers      []mcpServerView `json:"mcp_servers"`
}

type mcpServerView struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Command     string   `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`
	URL         string   `json:"url,omitempty"`
	Disabled    bool     `json:"disabled"`
	KeepAlive   bool     `json:"keep_alive"`
	Context     bool     `json:"context"`
	ContextText string   `json:"context_text,omitempty"`
	EnvKeys     []string `json:"env_keys,omitempty"`
}

func newDaemonConfigView(cfg *config.Config) daemonConfigView {
	if cfg == nil {
		return daemonConfigView{}
	}
	return daemonConfigView{
		Provider:        cfg.Provider,
		Endpoint:        cfg.Endpoint,
		ModelTier:       cfg.ModelTier,
		OpenAIEndpoint:  cfg.OpenAIEndpoint,
		OpenAIModel:     cfg.OpenAIModel,
		OllamaEndpoint:  cfg.OllamaEndpoint,
		OllamaModel:     cfg.OllamaModel,
		APIKeySet:       strings.TrimSpace(cfg.APIKey) != "",
		OpenAIAPIKeySet: strings.TrimSpace(cfg.OpenAIAPIKey) != "",
		MCPServers:      newMCPServerViews(cfg),
	}
}

func newMCPServerViews(cfg *config.Config) []mcpServerView {
	if cfg == nil || len(cfg.MCPServers) == 0 {
		return nil
	}
	views := make([]mcpServerView, 0, len(cfg.MCPServers))
	for name, server := range cfg.MCPServers {
		serverType := server.Type
		if serverType == "" {
			serverType = "stdio"
		}
		envKeys := make([]string, 0, len(server.Env))
		for key := range server.Env {
			envKeys = append(envKeys, key)
		}
		sort.Strings(envKeys)
		views = append(views, mcpServerView{
			Name:        name,
			Type:        serverType,
			Command:     server.Command,
			Args:        append([]string(nil), server.Args...),
			URL:         server.URL,
			Disabled:    server.Disabled,
			KeepAlive:   server.KeepAlive,
			Context:     strings.TrimSpace(server.Context) != "",
			ContextText: server.Context,
			EnvKeys:     envKeys,
		})
	}
	sort.Slice(views, func(i, j int) bool {
		return views[i].Name < views[j].Name
	})
	return views
}

func readDaemonConfig(path string, fallback *config.Config) (*config.Config, error) {
	cfg, err := config.LoadFromPath(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if fallback != nil {
				copy := *fallback
				applyProviderDefaults(&copy)
				return &copy, nil
			}
			return &config.Config{}, nil
		}
		return nil, err
	}
	applyProviderDefaults(cfg)
	return cfg, nil
}

type providerConfigPatch struct {
	Provider        *string             `json:"provider"`
	Endpoint        *string             `json:"endpoint"`
	APIKey          *string             `json:"api_key"`
	ModelTier       *string             `json:"model_tier"`
	OpenAIEndpoint  *string             `json:"openai_endpoint"`
	OpenAIAPIKey    *string             `json:"openai_api_key"`
	OpenAIModel     *string             `json:"openai_model"`
	OllamaEndpoint  *string             `json:"ollama_endpoint"`
	OllamaModel     *string             `json:"ollama_model"`
	OpenAIKeySet    *bool               `json:"openai_api_key_set"`
	AnthropicKeySet *bool               `json:"api_key_set"`
	Permissions     *permissions.Config `json:"permissions"`
	MCPServers      *[]mcpServerPatch   `json:"mcp_servers"`
}

type mcpServerPatch struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	URL       string            `json:"url"`
	Env       map[string]string `json:"env"`
	Disabled  bool              `json:"disabled"`
	KeepAlive bool              `json:"keep_alive"`
	Context   string            `json:"context"`
}

func (p providerConfigPatch) apply(cfg *config.Config) error {
	if p.Provider != nil {
		provider := strings.TrimSpace(*p.Provider)
		switch provider {
		case "anthropic", "openai", "ollama":
			cfg.Provider = provider
		default:
			return fmt.Errorf("unsupported provider %q", provider)
		}
	}
	if p.Endpoint != nil {
		cfg.Endpoint = strings.TrimSpace(*p.Endpoint)
	}
	if p.ModelTier != nil {
		cfg.ModelTier = strings.TrimSpace(*p.ModelTier)
	}
	if p.OpenAIEndpoint != nil {
		cfg.OpenAIEndpoint = strings.TrimSpace(*p.OpenAIEndpoint)
	}
	if p.OpenAIModel != nil {
		cfg.OpenAIModel = strings.TrimSpace(*p.OpenAIModel)
	}
	if p.OllamaEndpoint != nil {
		cfg.OllamaEndpoint = strings.TrimSpace(*p.OllamaEndpoint)
	}
	if p.OllamaModel != nil {
		cfg.OllamaModel = strings.TrimSpace(*p.OllamaModel)
	}
	if p.APIKey != nil && strings.TrimSpace(*p.APIKey) != "" {
		cfg.APIKey = strings.TrimSpace(*p.APIKey)
	}
	if p.OpenAIAPIKey != nil && strings.TrimSpace(*p.OpenAIAPIKey) != "" {
		cfg.OpenAIAPIKey = strings.TrimSpace(*p.OpenAIAPIKey)
	}
	if p.Permissions != nil {
		cfg.Permissions = cleanPermissions(p.Permissions)
	}
	if p.MCPServers != nil {
		servers, err := cleanMCPServers(*p.MCPServers, cfg.MCPServers)
		if err != nil {
			return err
		}
		cfg.MCPServers = servers
	}
	applyProviderDefaults(cfg)
	return nil
}

func cleanMCPServers(patches []mcpServerPatch, existing map[string]mcp.MCPServerConfig) (map[string]mcp.MCPServerConfig, error) {
	if len(patches) == 0 {
		return nil, nil
	}
	servers := make(map[string]mcp.MCPServerConfig, len(patches))
	for _, patch := range patches {
		name := strings.TrimSpace(patch.Name)
		if name == "" {
			return nil, fmt.Errorf("MCP server name is required")
		}
		if !mcpServerNameRe.MatchString(name) {
			return nil, fmt.Errorf("invalid MCP server name %q", name)
		}
		if _, exists := servers[name]; exists {
			return nil, fmt.Errorf("duplicate MCP server name %q", name)
		}

		serverType := strings.TrimSpace(patch.Type)
		if serverType == "" {
			serverType = "stdio"
		}
		server := mcp.MCPServerConfig{
			Type:      serverType,
			Disabled:  patch.Disabled,
			KeepAlive: patch.KeepAlive,
			Context:   strings.TrimSpace(patch.Context),
		}
		switch serverType {
		case "stdio":
			server.Command = strings.TrimSpace(patch.Command)
			if server.Command == "" {
				return nil, fmt.Errorf("MCP server %q command is required for stdio", name)
			}
			server.Args = cleanStringList(patch.Args)
		case "http":
			server.URL = strings.TrimSpace(patch.URL)
			if server.URL == "" {
				return nil, fmt.Errorf("MCP server %q url is required for http", name)
			}
		default:
			return nil, fmt.Errorf("MCP server %q has unsupported type %q", name, serverType)
		}

		env, err := cleanMCPEnv(name, patch.Env, existing)
		if err != nil {
			return nil, err
		}
		server.Env = env
		servers[name] = server
	}
	return servers, nil
}

func cleanMCPEnv(name string, patchEnv map[string]string, existing map[string]mcp.MCPServerConfig) (map[string]string, error) {
	if len(patchEnv) == 0 {
		return nil, nil
	}
	env := make(map[string]string, len(patchEnv))
	var existingEnv map[string]string
	if existing != nil {
		existingEnv = existing[name].Env
	}
	for rawKey, rawValue := range patchEnv {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			return nil, fmt.Errorf("MCP server %q env key is required", name)
		}
		if strings.TrimSpace(rawValue) == "" {
			if existingEnv != nil {
				if existingValue, ok := existingEnv[key]; ok {
					env[key] = existingValue
				}
			}
			continue
		}
		env[key] = rawValue
	}
	if len(env) == 0 {
		return nil, nil
	}
	return env, nil
}

func cleanPermissions(perms *permissions.Config) *permissions.Config {
	if perms == nil {
		return nil
	}
	cleaned := &permissions.Config{
		AllowedDirs:       cleanStringList(perms.AllowedDirs),
		AllowedCommands:   cleanStringList(perms.AllowedCommands),
		DeniedCommands:    cleanStringList(perms.DeniedCommands),
		SensitivePatterns: cleanStringList(perms.SensitivePatterns),
		NetworkAllowlist:  cleanStringList(perms.NetworkAllowlist),
	}
	if len(cleaned.AllowedDirs) == 0 &&
		len(cleaned.AllowedCommands) == 0 &&
		len(cleaned.DeniedCommands) == 0 &&
		len(cleaned.SensitivePatterns) == 0 &&
		len(cleaned.NetworkAllowlist) == 0 {
		return nil
	}
	return cleaned
}

func cleanStringList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		if item := strings.TrimSpace(value); item != "" {
			cleaned = append(cleaned, item)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func applyProviderDefaults(cfg *config.Config) {
	if cfg.Provider == "" {
		cfg.Provider = "anthropic"
	}
}
