package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/starclaw/starclaw/internal/agents"
	"gopkg.in/yaml.v3"
)

type agentEditRequest struct {
	Name            string             `json:"name"`
	Prompt          string             `json:"prompt"`
	Memory          string             `json:"memory"`
	Model           string             `json:"model"`
	ReasoningEffort string             `json:"reasoning_effort"`
	ToolsAllow      []string           `json:"tools_allow"`
	ToolsDeny       []string           `json:"tools_deny"`
	AutoApprove     *bool              `json:"auto_approve"`
	HeartbeatEvery  string             `json:"heartbeat_every"`
	HeartbeatHours  string             `json:"heartbeat_active_hours"`
	HeartbeatModel  string             `json:"heartbeat_model"`
	Commands        *map[string]string `json:"commands"`
}

var agentCommandNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func saveAgentDefinition(agentsDir, name string, req agentEditRequest, create bool) (*agents.Agent, error) {
	if err := agents.ValidateAgentName(name); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	agentDir := filepath.Join(agentsDir, name)
	agentFile := filepath.Join(agentDir, "AGENT.md")
	if create {
		if _, err := os.Stat(agentFile); err == nil {
			return nil, fmt.Errorf("agent %q already exists", name)
		} else if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat agent: %w", err)
		}
	} else if _, err := os.Stat(agentFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent %q not found", name)
		}
		return nil, fmt.Errorf("stat agent: %w", err)
	}
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		return nil, fmt.Errorf("create agent directory: %w", err)
	}
	if err := os.WriteFile(agentFile, []byte(prompt+"\n"), 0600); err != nil {
		return nil, fmt.Errorf("write AGENT.md: %w", err)
	}

	memory := strings.TrimSpace(req.Memory)
	memoryFile := filepath.Join(agentDir, "MEMORY.md")
	if memory == "" {
		if err := os.Remove(memoryFile); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove MEMORY.md: %w", err)
		}
	} else if err := os.WriteFile(memoryFile, []byte(memory+"\n"), 0600); err != nil {
		return nil, fmt.Errorf("write MEMORY.md: %w", err)
	}

	cfg := buildAgentConfig(req)
	configFile := filepath.Join(agentDir, "config.yaml")
	if isEmptyAgentConfig(cfg) {
		if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove config.yaml: %w", err)
		}
	} else {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("marshal config.yaml: %w", err)
		}
		if err := os.WriteFile(configFile, data, 0600); err != nil {
			return nil, fmt.Errorf("write config.yaml: %w", err)
		}
	}
	if req.Commands != nil {
		if err := saveAgentCommands(agentDir, *req.Commands); err != nil {
			return nil, err
		}
	}
	return agents.LoadAgent(agentsDir, name)
}

func buildAgentConfig(req agentEditRequest) *agents.AgentConfig {
	cfg := &agents.AgentConfig{}
	modelCfg := &agents.AgentModelConfig{}
	if value := strings.TrimSpace(req.Model); value != "" {
		modelCfg.Model = &value
	}
	if value := strings.TrimSpace(req.ReasoningEffort); value != "" {
		modelCfg.ReasoningEffort = &value
	}
	if modelCfg.Model != nil || modelCfg.ReasoningEffort != nil {
		cfg.Agent = modelCfg
	}
	allow := trimStringList(req.ToolsAllow)
	deny := trimStringList(req.ToolsDeny)
	if len(allow) > 0 || len(deny) > 0 {
		cfg.Tools = &agents.AgentToolsFilter{Allow: allow, Deny: deny}
	}
	if req.AutoApprove != nil {
		cfg.AutoApprove = req.AutoApprove
	}
	if every := strings.TrimSpace(req.HeartbeatEvery); every != "" {
		cfg.Heartbeat = &agents.HeartbeatConfig{
			Every:       every,
			ActiveHours: strings.TrimSpace(req.HeartbeatHours),
			Model:       strings.TrimSpace(req.HeartbeatModel),
		}
	}
	return cfg
}

func isEmptyAgentConfig(cfg *agents.AgentConfig) bool {
	return cfg == nil || (cfg.Tools == nil && cfg.Agent == nil && cfg.AutoApprove == nil && cfg.Heartbeat == nil)
}

func trimStringList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func saveAgentCommands(agentDir string, commands map[string]string) error {
	commandsDir := filepath.Join(agentDir, "commands")
	cleaned := make(map[string]string, len(commands))
	for name, content := range commands {
		trimmedName := strings.TrimSpace(name)
		if !agentCommandNameRe.MatchString(trimmedName) {
			return fmt.Errorf("invalid command name %q: must match %s", name, agentCommandNameRe.String())
		}
		trimmedContent := strings.TrimSpace(content)
		if trimmedContent == "" {
			continue
		}
		cleaned[trimmedName] = trimmedContent
	}
	if len(cleaned) == 0 {
		if err := os.RemoveAll(commandsDir); err != nil {
			return fmt.Errorf("remove commands directory: %w", err)
		}
		return nil
	}
	if err := os.RemoveAll(commandsDir); err != nil {
		return fmt.Errorf("replace commands directory: %w", err)
	}
	if err := os.MkdirAll(commandsDir, 0700); err != nil {
		return fmt.Errorf("create commands directory: %w", err)
	}
	for name, content := range cleaned {
		path := filepath.Join(commandsDir, name+".md")
		if err := os.WriteFile(path, []byte(content+"\n"), 0600); err != nil {
			return fmt.Errorf("write command %q: %w", name, err)
		}
	}
	return nil
}
