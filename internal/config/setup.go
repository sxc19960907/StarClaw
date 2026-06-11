package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SetupWizard runs an interactive configuration wizard
func SetupWizard() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║     StarClaw Configuration Setup          ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Println()

	// API Endpoint
	fmt.Print("API Endpoint / Base URL: ")
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("API endpoint is required")
	}

	// API Key
	fmt.Print("API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Model name
	fmt.Print("Model name (for Anthropic-compatible endpoint): ")
	modelTier, _ := reader.ReadString('\n')
	modelTier = strings.TrimSpace(modelTier)
	if modelTier == "" {
		return nil, fmt.Errorf("model name is required")
	}

	cfg := &Config{
		Endpoint:  endpoint,
		APIKey:    apiKey,
		ModelTier: modelTier,
		Agent: AgentConfig{
			MaxIterations: 25,
			Temperature:   0,
			MaxTokens:     8192,
		},
		Tools: ToolsConfig{
			BashTimeout:       120,
			BashMaxOutput:     30000,
			ResultTruncation:  30000,
			ArgsTruncation:    200,
			GrepMaxResults:    100,
			ServerToolTimeout: 0,
		},
	}

	// Save configuration
	if err := Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved successfully!")
	fmt.Printf("  Config location: %s/config.yaml\n", StarclawDir())
	fmt.Println()
	fmt.Println("You can now run 'starclaw' to start.")

	return cfg, nil
}

// RunSetup checks if setup is needed and runs the wizard
func RunSetup() (*Config, error) {
	// Try to load existing config
	cfg, err := Load()
	if err == nil && !NeedsSetup(cfg) {
		fmt.Println("Configuration already exists.")
		fmt.Printf("  Endpoint: %s\n", cfg.Endpoint)
		fmt.Printf("  Model: %s\n", cfg.ModelTier)
		fmt.Println()
		fmt.Print("Do you want to reconfigure? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Setup cancelled.")
			return cfg, nil
		}
	}

	// Run the wizard
	return SetupWizard()
}
