// Package agents provides named agent loading and management.
package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var agentNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// AgentToolsFilter controls which local tools an agent can access.
type AgentToolsFilter struct {
	Allow []string `yaml:"allow,omitempty" json:"allow,omitempty"`
	Deny  []string `yaml:"deny,omitempty" json:"deny,omitempty"`
}

// AgentModelConfig holds per-agent model/iteration overrides.
type AgentModelConfig struct {
	Model         *string  `yaml:"model,omitempty" json:"model,omitempty"`
	MaxIterations *int     `yaml:"max_iterations,omitempty" json:"max_iterations,omitempty"`
	Temperature   *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxTokens     *int     `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
}

// AgentConfig is the per-agent config overlay.
type AgentConfig struct {
	Tools       *AgentToolsFilter `yaml:"tools,omitempty"`
	Agent       *AgentModelConfig `yaml:"agent,omitempty"`
	AutoApprove *bool             `yaml:"auto_approve,omitempty"`
}

// Agent represents a loaded agent definition.
type Agent struct {
	Name     string
	Prompt   string
	Memory   string
	Config   *AgentConfig
	Commands map[string]string // Custom slash commands
}

// AgentInfo provides basic info about an agent for listing.
type AgentInfo struct {
	Name        string
	Description string // First line of AGENT.md
}

// ValidateAgentName validates an agent name format.
func ValidateAgentName(name string) error {
	if !agentNameRe.MatchString(name) {
		return fmt.Errorf("invalid agent name %q: must match %s", name, agentNameRe.String())
	}
	return nil
}

// LoadAgent loads an agent by name from the agents directory.
func LoadAgent(agentsDir, name string) (*Agent, error) {
	if err := ValidateAgentName(name); err != nil {
		return nil, err
	}

	dir := filepath.Join(agentsDir, name)

	// Read AGENT.md (required)
	promptData, err := os.ReadFile(filepath.Join(dir, "AGENT.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent %q not found: AGENT.md missing", name)
		}
		return nil, fmt.Errorf("agent %q: failed to read AGENT.md: %w", name, err)
	}

	agent := &Agent{
		Name:     name,
		Prompt:   string(promptData),
		Commands: make(map[string]string),
	}

	// Read MEMORY.md (optional)
	memoryData, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err == nil {
		agent.Memory = string(memoryData)
	}

	// Read config.yaml (optional)
	configData, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err == nil {
		var cfg AgentConfig
		if err := yaml.Unmarshal(configData, &cfg); err != nil {
			return nil, fmt.Errorf("agent %q: failed to parse config.yaml: %w", name, err)
		}
		agent.Config = &cfg
	}

	// Load custom commands from commands/ directory (optional)
	commandsDir := filepath.Join(dir, "commands")
	if entries, err := os.ReadDir(commandsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Skip non-.md files
			if !strings.HasSuffix(name, ".md") {
				continue
			}
			content, err := os.ReadFile(filepath.Join(commandsDir, name))
			if err != nil {
				continue
			}
			// Strip .md extension for command name
			cmdName := strings.TrimSuffix(name, ".md")
			agent.Commands[cmdName] = string(content)
		}
	}

	return agent, nil
}

// ListAgents lists all available agents in the agents directory.
func ListAgents(agentsDir string) ([]AgentInfo, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []AgentInfo{}, nil
		}
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var agents []AgentInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Validate name format
		if err := ValidateAgentName(name); err != nil {
			continue
		}

		// Check if AGENT.md exists
		agentPath := filepath.Join(agentsDir, name)
		agentFile := filepath.Join(agentPath, "AGENT.md")
		if _, err := os.Stat(agentFile); err != nil {
			continue
		}

		// Read first line for description
		promptData, _ := os.ReadFile(agentFile)
		description := ""
		if len(promptData) > 0 {
			lines := strings.Split(string(promptData), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
					description = trimmed
					if len(description) > 50 {
						description = description[:47] + "..."
					}
					break
				}
			}
		}

		agents = append(agents, AgentInfo{
			Name:        name,
			Description: description,
		})
	}

	// Sort by name
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	return agents, nil
}

// AgentDir returns the path to an agent's directory.
func AgentDir(baseDir, name string) string {
	return filepath.Join(baseDir, "agents", name)
}
