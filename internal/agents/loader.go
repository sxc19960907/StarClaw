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
	Model                 *string                 `yaml:"model,omitempty" json:"model,omitempty"`
	MaxIterations         *int                    `yaml:"max_iterations,omitempty" json:"max_iterations,omitempty"`
	Temperature           *float64                `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxTokens             *int                    `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
	ContextWindow         *int                    `yaml:"context_window,omitempty" json:"context_window,omitempty"`
	StreamIdleTimeoutSecs *int                    `yaml:"stream_idle_timeout_secs,omitempty" json:"stream_idle_timeout_secs,omitempty"`
	TokenBudget           *AgentTokenBudgetConfig `yaml:"token_budget,omitempty" json:"token_budget,omitempty"`
	Thinking              *bool                   `yaml:"thinking,omitempty" json:"thinking,omitempty"`
	ThinkingMode          *string                 `yaml:"thinking_mode,omitempty" json:"thinking_mode,omitempty"`
	ThinkingBudget        *int                    `yaml:"thinking_budget,omitempty" json:"thinking_budget,omitempty"`
	ReasoningEffort       *string                 `yaml:"reasoning_effort,omitempty" json:"reasoning_effort,omitempty"`
}

// AgentTokenBudgetConfig holds optional per-agent token budget overrides.
type AgentTokenBudgetConfig struct {
	MaxInputTokens  *int  `yaml:"max_input_tokens,omitempty" json:"max_input_tokens,omitempty"`
	MaxOutputTokens *int  `yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`
	MaxTotalTokens  *int  `yaml:"max_total_tokens,omitempty" json:"max_total_tokens,omitempty"`
	HardStop        *bool `yaml:"hard_stop,omitempty" json:"hard_stop,omitempty"`
}

// AgentConfig is the per-agent config overlay.
type AgentConfig struct {
	Tools       *AgentToolsFilter `yaml:"tools,omitempty"`
	Agent       *AgentModelConfig `yaml:"agent,omitempty"`
	AutoApprove *bool             `yaml:"auto_approve,omitempty"`
	Heartbeat   *HeartbeatConfig  `yaml:"heartbeat,omitempty" json:"heartbeat,omitempty"`
}

// HeartbeatConfig configures periodic heartbeat checks for an agent.
type HeartbeatConfig struct {
	Every       string `yaml:"every" json:"every"`
	ActiveHours string `yaml:"active_hours,omitempty" json:"active_hours,omitempty"`
	Model       string `yaml:"model,omitempty" json:"model,omitempty"`
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
	Name            string
	Description     string // First line of AGENT.md
	Model           string
	ReasoningEffort string
	ToolsAllow      []string
	ToolsDeny       []string
	AutoApprove     bool
	HeartbeatEvery  string
	HeartbeatHours  string
	HeartbeatModel  string
	CommandCount    int
	CommandNames    []string
	HasMemory       bool
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

		info := AgentInfo{
			Name:        name,
			Description: description,
		}
		if agent, err := LoadAgent(agentsDir, name); err == nil {
			enrichAgentInfo(&info, agent)
		}
		agents = append(agents, info)
	}

	// Sort by name
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	return agents, nil
}

func enrichAgentInfo(info *AgentInfo, agent *Agent) {
	if agent == nil {
		return
	}
	info.CommandCount = len(agent.Commands)
	info.CommandNames = make([]string, 0, len(agent.Commands))
	for name := range agent.Commands {
		info.CommandNames = append(info.CommandNames, name)
	}
	sort.Strings(info.CommandNames)
	info.HasMemory = strings.TrimSpace(agent.Memory) != ""
	if agent.Config == nil {
		return
	}
	if agent.Config.Agent != nil {
		if agent.Config.Agent.Model != nil {
			info.Model = *agent.Config.Agent.Model
		}
		if agent.Config.Agent.ReasoningEffort != nil {
			info.ReasoningEffort = *agent.Config.Agent.ReasoningEffort
		}
	}
	if agent.Config.Tools != nil {
		info.ToolsAllow = append([]string(nil), agent.Config.Tools.Allow...)
		info.ToolsDeny = append([]string(nil), agent.Config.Tools.Deny...)
	}
	if agent.Config.AutoApprove != nil {
		info.AutoApprove = *agent.Config.AutoApprove
	}
	if agent.Config.Heartbeat != nil {
		info.HeartbeatEvery = agent.Config.Heartbeat.Every
		info.HeartbeatHours = agent.Config.Heartbeat.ActiveHours
		info.HeartbeatModel = agent.Config.Heartbeat.Model
	}
}

// AgentDir returns the path to an agent's directory.
func AgentDir(baseDir, name string) string {
	return filepath.Join(baseDir, "agents", name)
}

// AutoApproveEnabled returns whether the agent has auto-approve enabled.
func (a *Agent) AutoApproveEnabled() bool {
	if a == nil || a.Config == nil || a.Config.AutoApprove == nil {
		return false
	}
	return *a.Config.AutoApprove
}
