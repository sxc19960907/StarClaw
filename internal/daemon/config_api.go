package daemon

import (
	"fmt"
	"os"
	"strings"

	"github.com/starclaw/starclaw/internal/config"
	"gopkg.in/yaml.v3"
)

type daemonConfigView struct {
	Provider        string `json:"provider"`
	Endpoint        string `json:"endpoint"`
	ModelTier       string `json:"model_tier"`
	OpenAIEndpoint  string `json:"openai_endpoint"`
	OpenAIModel     string `json:"openai_model"`
	OllamaEndpoint  string `json:"ollama_endpoint"`
	OllamaModel     string `json:"ollama_model"`
	APIKeySet       bool   `json:"api_key_set"`
	OpenAIAPIKeySet bool   `json:"openai_api_key_set"`
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
	}
}

func readDaemonConfig(path string, fallback *config.Config) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if fallback != nil {
				copy := *fallback
				return &copy, nil
			}
			return &config.Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}
	applyProviderDefaults(&cfg)
	return &cfg, nil
}

type providerConfigPatch struct {
	Provider        *string `json:"provider"`
	Endpoint        *string `json:"endpoint"`
	APIKey          *string `json:"api_key"`
	ModelTier       *string `json:"model_tier"`
	OpenAIEndpoint  *string `json:"openai_endpoint"`
	OpenAIAPIKey    *string `json:"openai_api_key"`
	OpenAIModel     *string `json:"openai_model"`
	OllamaEndpoint  *string `json:"ollama_endpoint"`
	OllamaModel     *string `json:"ollama_model"`
	OpenAIKeySet    *bool   `json:"openai_api_key_set"`
	AnthropicKeySet *bool   `json:"api_key_set"`
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
	applyProviderDefaults(cfg)
	return nil
}

func applyProviderDefaults(cfg *config.Config) {
	if cfg.Provider == "" {
		cfg.Provider = "anthropic"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.anthropic.com"
	}
	if cfg.ModelTier == "" {
		cfg.ModelTier = "medium"
	}
	if cfg.OpenAIEndpoint == "" {
		cfg.OpenAIEndpoint = "https://api.openai.com/v1"
	}
	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4o"
	}
	if cfg.OllamaEndpoint == "" {
		cfg.OllamaEndpoint = "http://localhost:11434"
	}
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = "llama3.1"
	}
}
