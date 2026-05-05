package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAgentName(t *testing.T) {
	tests := []struct {
		name    string
		valid   bool
	}{
		{"agent1", true},
		{"my-agent", true},
		{"my_agent", true},
		{"a", true},
		{"agent-123", true},
		{"Agent", false},       // uppercase
		{"-agent", false},      // starts with hyphen
		{"_agent", false},      // starts with underscore
		{"agent!", false},      // special character
		{"", false},            // empty
		{"agent name", false},  // space
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentName(tt.name)
			if tt.valid && err != nil {
				t.Errorf("ValidateAgentName(%q) returned error: %v", tt.name, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateAgentName(%q) should have returned error", tt.name)
			}
		})
	}
}

func TestLoadAgent(t *testing.T) {
	// Create temp agents directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create test agent
	agentDir := filepath.Join(agentsDir, "test-agent")
	os.MkdirAll(agentDir, 0755)

	// Create AGENT.md
	agentPrompt := `# Test Agent

You are a test agent for unit testing.`
	os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(agentPrompt), 0644)

	// Create MEMORY.md
	agentMemory := `# Memory

Remember to be helpful.`
	os.WriteFile(filepath.Join(agentDir, "MEMORY.md"), []byte(agentMemory), 0644)

	// Create config.yaml
	configData := `agent:
  max_iterations: 50
  temperature: 0.5
tools:
  allow:
    - file_read
    - file_write
`
	os.WriteFile(filepath.Join(agentDir, "config.yaml"), []byte(configData), 0644)

	// Load the agent
	agent, err := LoadAgent(agentsDir, "test-agent")
	if err != nil {
		t.Fatalf("LoadAgent failed: %v", err)
	}

	// Verify loaded data
	if agent.Name != "test-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "test-agent")
	}

	if agent.Prompt != agentPrompt {
		t.Errorf("Prompt mismatch: got %q, want %q", agent.Prompt, agentPrompt)
	}

	if agent.Memory != agentMemory {
		t.Errorf("Memory mismatch: got %q, want %q", agent.Memory, agentMemory)
	}

	if agent.Config == nil {
		t.Fatal("Config should be loaded")
	}

	if agent.Config.Agent == nil {
		t.Fatal("Agent config should be loaded")
	}

	if agent.Config.Agent.MaxIterations == nil || *agent.Config.Agent.MaxIterations != 50 {
		t.Error("MaxIterations should be 50")
	}

	if agent.Config.Agent.Temperature == nil || *agent.Config.Agent.Temperature != 0.5 {
		t.Error("Temperature should be 0.5")
	}

	if agent.Config.Tools == nil {
		t.Fatal("Tools config should be loaded")
	}

	if len(agent.Config.Tools.Allow) != 2 {
		t.Errorf("Tools.Allow should have 2 items, got %d", len(agent.Config.Tools.Allow))
	}
}

func TestLoadAgent_MissingAgentFile(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create agent dir but no AGENT.md
	agentDir := filepath.Join(agentsDir, "incomplete")
	os.MkdirAll(agentDir, 0755)

	_, err := LoadAgent(agentsDir, "incomplete")
	if err == nil {
		t.Error("LoadAgent should return error when AGENT.md is missing")
	}
}

func TestLoadAgent_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	_, err := LoadAgent(agentsDir, "Invalid-Name")
	if err == nil {
		t.Error("LoadAgent should return error for invalid name")
	}
}

func TestListAgents(t *testing.T) {
	// Create temp agents directory
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create two test agents
	for _, name := range []string{"agent-a", "agent-b"} {
		agentDir := filepath.Join(agentsDir, name)
		os.MkdirAll(agentDir, 0755)
		os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# "+name+"\n\nDescription."), 0644)
	}

	// Create an incomplete agent (no AGENT.md)
	incompleteDir := filepath.Join(agentsDir, "incomplete")
	os.MkdirAll(incompleteDir, 0755)
	os.WriteFile(filepath.Join(incompleteDir, "README.md"), []byte("not an agent"), 0644)

	// Create an invalid agent name
	invalidDir := filepath.Join(agentsDir, "Invalid-Name")
	os.MkdirAll(invalidDir, 0755)
	os.WriteFile(filepath.Join(invalidDir, "AGENT.md"), []byte("# Invalid"), 0644)

	// List agents
	agents, err := ListAgents(agentsDir)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}

	// Should be sorted
	if len(agents) >= 2 && agents[0].Name != "agent-a" {
		t.Error("Agents should be sorted by name")
	}
}

func TestListAgents_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	agents, err := ListAgents(agentsDir)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents))
	}
}

func TestListAgents_MissingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "nonexistent")

	agents, err := ListAgents(agentsDir)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents for missing directory, got %d", len(agents))
	}
}

func TestAgentDir(t *testing.T) {
	dir := AgentDir("/home/user/.starclaw", "my-agent")
	expected := filepath.Join("/home/user/.starclaw", "agents", "my-agent")
	if dir != expected {
		t.Errorf("AgentDir = %q, want %q", dir, expected)
	}
}

func TestAutoApproveEnabled(t *testing.T) {
	// Nil agent
	var ag *Agent
	if ag.AutoApproveEnabled() {
		t.Error("nil agent should return false")
	}

	// Agent without config
	ag = &Agent{}
	if ag.AutoApproveEnabled() {
		t.Error("agent without config should return false")
	}

	// Agent with config but nil AutoApprove
	ag = &Agent{Config: &AgentConfig{}}
	if ag.AutoApproveEnabled() {
		t.Error("agent with nil AutoApprove should return false")
	}

	// Agent with AutoApprove = true
	trueVal := true
	ag = &Agent{Config: &AgentConfig{AutoApprove: &trueVal}}
	if !ag.AutoApproveEnabled() {
		t.Error("agent with AutoApprove=true should return true")
	}

	// Agent with AutoApprove = false
	falseVal := false
	ag = &Agent{Config: &AgentConfig{AutoApprove: &falseVal}}
	if ag.AutoApproveEnabled() {
		t.Error("agent with AutoApprove=false should return false")
	}
}

func TestLoadAgent_CustomCommands(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")

	agentDir := filepath.Join(agentsDir, "test-agent")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# Test\n\nTest agent."), 0644)

	// Create commands directory
	commandsDir := filepath.Join(agentDir, "commands")
	os.MkdirAll(commandsDir, 0755)
	os.WriteFile(filepath.Join(commandsDir, "deploy.md"), []byte("# Deploy Command\n\nDeploy to production."), 0644)
	os.WriteFile(filepath.Join(commandsDir, "review.md"), []byte("# Review Command\n\nReview the code."), 0644)
	os.WriteFile(filepath.Join(commandsDir, "not-a-command.txt"), []byte("ignored"), 0644) // non-.md, should skip
	os.WriteFile(filepath.Join(commandsDir, "README.md"), []byte("# README\n\nDocumentation."), 0644)

	agent, err := LoadAgent(agentsDir, "test-agent")
	if err != nil {
		t.Fatalf("LoadAgent failed: %v", err)
	}

	if len(agent.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(agent.Commands))
	}

	// Verify .md extension is stripped
	if _, ok := agent.Commands["deploy"]; !ok {
		t.Error("Expected 'deploy' command")
	}
	if _, ok := agent.Commands["review"]; !ok {
		t.Error("Expected 'review' command")
	}
	if _, ok := agent.Commands["README"]; !ok {
		t.Error("Expected 'README' command")
	}
}

func TestLoadAgent_WithoutOptionalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")

	agentDir := filepath.Join(agentsDir, "minimal")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# Minimal\n\nMinimal agent."), 0644)

	agent, err := LoadAgent(agentsDir, "minimal")
	if err != nil {
		t.Fatalf("LoadAgent failed: %v", err)
	}

	if agent.Memory != "" {
		t.Error("Memory should be empty when MEMORY.md is missing")
	}
	if agent.Config != nil {
		t.Error("Config should be nil when config.yaml is missing")
	}
	if len(agent.Commands) != 0 {
		t.Error("Commands should be empty when commands dir is missing")
	}
	if agent.Name != "minimal" {
		t.Errorf("Name = %q, want 'minimal'", agent.Name)
	}
}

func TestLoadAgent_BadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")

	agentDir := filepath.Join(agentsDir, "test-agent")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# Test\n\nTest agent."), 0644)
	os.WriteFile(filepath.Join(agentDir, "config.yaml"), []byte("invalid: [yaml: : bad"), 0644)

	_, err := LoadAgent(agentsDir, "test-agent")
	if err == nil {
		t.Error("LoadAgent should fail with bad config.yaml")
	}
}

func TestLoadAgent_AgentNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	_, err := LoadAgent(agentsDir, "nonexistent")
	if err == nil {
		t.Error("LoadAgent should fail when agent directory doesn't exist")
	}
}

func TestListAgents_OnlyValidNames(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	// Create an invalid-name agent with valid AGENT.md
	// It should be skipped by ListAgents
	invalidDir := filepath.Join(agentsDir, "Invalid!")
	os.MkdirAll(invalidDir, 0755)
	os.WriteFile(filepath.Join(invalidDir, "AGENT.md"), []byte("# Invalid"), 0644)

	agents, err := ListAgents(agentsDir)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents (invalid name filtered), got %d", len(agents))
	}
}
