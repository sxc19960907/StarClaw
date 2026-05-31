package config

import (
	"testing"

	"github.com/starclaw/starclaw/internal/agents"
)

func TestMergeAgentConfig_NilAgent(t *testing.T) {
	global := &Config{
		Agent: AgentConfig{
			MaxIterations: 25,
			MaxTokens:     8192,
		},
	}

	result := MergeAgentConfig(global, nil)
	if result != global {
		t.Error("MergeAgentConfig with nil agent should return original config")
	}
}

func TestMergeAgentConfig_NilConfig(t *testing.T) {
	global := &Config{
		Agent: AgentConfig{
			MaxIterations: 25,
		},
	}

	ag := &agents.Agent{Name: "test", Config: nil}
	result := MergeAgentConfig(global, ag)
	if result != global {
		t.Error("MergeAgentConfig with nil agent.Config should return original config")
	}
}

func TestMergeAgentConfig_AgentOverrides(t *testing.T) {
	global := &Config{
		ModelTier: "medium",
		Agent: AgentConfig{
			MaxIterations: 25,
			Temperature:   0.0,
			MaxTokens:     8192,
			Model:         "",
		},
	}

	maxIter := 50
	temp := 0.7
	maxTok := 16000

	ag := &agents.Agent{
		Name: "coder",
		Config: &agents.AgentConfig{
			Agent: &agents.AgentModelConfig{
				Model:         strPtr("opus"),
				MaxIterations: &maxIter,
				Temperature:   &temp,
				MaxTokens:     &maxTok,
			},
		},
	}

	result := MergeAgentConfig(global, ag)

	if result.Agent.Model != "opus" {
		t.Errorf("Agent.Model = %q, want 'opus'", result.Agent.Model)
	}
	if result.ModelTier != "medium" {
		t.Errorf("ModelTier = %q, want 'medium'", result.ModelTier)
	}
	if result.Agent.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want 50", result.Agent.MaxIterations)
	}
	if result.Agent.Temperature != 0.7 {
		t.Errorf("Temperature = %f, want 0.7", result.Agent.Temperature)
	}
	if result.Agent.MaxTokens != 16000 {
		t.Errorf("MaxTokens = %d, want 16000", result.Agent.MaxTokens)
	}

	// Global config should be unchanged
	if global.ModelTier != "medium" {
		t.Error("Global config should not be modified")
	}
	if global.Agent.Model != "" {
		t.Error("Global config should not be modified")
	}
	if global.Agent.MaxIterations != 25 {
		t.Error("Global config should not be modified")
	}
}

func TestMergeAgentConfig_PartialOverrides(t *testing.T) {
	global := &Config{
		ModelTier: "medium",
		Agent: AgentConfig{
			MaxIterations: 25,
			Temperature:   0.0,
			MaxTokens:     8192,
		},
	}

	maxIter := 100
	ag := &agents.Agent{
		Name: "writer",
		Config: &agents.AgentConfig{
			Agent: &agents.AgentModelConfig{
				MaxIterations: &maxIter,
				// Temperature and MaxTokens are nil — should keep global values
			},
		},
	}

	result := MergeAgentConfig(global, ag)

	if result.Agent.MaxIterations != 100 {
		t.Errorf("MaxIterations = %d, want 100", result.Agent.MaxIterations)
	}
	if result.Agent.Temperature != 0.0 {
		t.Errorf("Temperature should keep global value 0.0, got %f", result.Agent.Temperature)
	}
	if result.Agent.MaxTokens != 8192 {
		t.Errorf("MaxTokens should keep global value 8192, got %d", result.Agent.MaxTokens)
	}
	if result.ModelTier != "medium" {
		t.Errorf("ModelTier should keep global value 'medium', got %q", result.ModelTier)
	}
}

func TestMergeAgentConfig_ToolFilters(t *testing.T) {
	global := &Config{
		Tools: ToolsConfig{
			Allowed: []string{"file_read", "grep", "bash"},
			Denied:  []string{"browser"},
		},
	}

	ag := &agents.Agent{
		Name: "restricted",
		Config: &agents.AgentConfig{
			Tools: &agents.AgentToolsFilter{
				Allow: []string{"file_read", "grep"},
				Deny:  []string{"bash"},
			},
		},
	}

	result := MergeAgentConfig(global, ag)

	if got := result.Tools.Allowed; len(got) != 2 || got[0] != "file_read" || got[1] != "grep" {
		t.Fatalf("Tools.Allowed = %#v, want [file_read grep]", got)
	}
	if got := result.Tools.Denied; len(got) != 1 || got[0] != "bash" {
		t.Fatalf("Tools.Denied = %#v, want [bash]", got)
	}

	result.Tools.Allowed[0] = "mutated"
	if global.Tools.Allowed[0] != "file_read" {
		t.Fatal("global allowed tools should not be modified through merged config")
	}
}

func TestMergeAgentConfig_AdvancedAgentOptions(t *testing.T) {
	global := &Config{
		ModelTier: "medium",
		Agent: AgentConfig{
			Thinking:        true,
			ThinkingMode:    "adaptive",
			ThinkingBudget:  10000,
			ReasoningEffort: "",
		},
	}

	thinking := false
	mode := "enabled"
	budget := 2048
	effort := "high"
	model := "claude-opus-test"
	ag := &agents.Agent{
		Name: "reasoner",
		Config: &agents.AgentConfig{
			Agent: &agents.AgentModelConfig{
				Model:           &model,
				Thinking:        &thinking,
				ThinkingMode:    &mode,
				ThinkingBudget:  &budget,
				ReasoningEffort: &effort,
			},
		},
	}

	result := MergeAgentConfig(global, ag)

	if result.Agent.Model != "claude-opus-test" {
		t.Fatalf("Agent.Model = %q, want claude-opus-test", result.Agent.Model)
	}
	if result.Agent.Thinking {
		t.Fatal("Agent.Thinking = true, want false")
	}
	if result.Agent.ThinkingMode != "enabled" {
		t.Fatalf("Agent.ThinkingMode = %q, want enabled", result.Agent.ThinkingMode)
	}
	if result.Agent.ThinkingBudget != 2048 {
		t.Fatalf("Agent.ThinkingBudget = %d, want 2048", result.Agent.ThinkingBudget)
	}
	if result.Agent.ReasoningEffort != "high" {
		t.Fatalf("Agent.ReasoningEffort = %q, want high", result.Agent.ReasoningEffort)
	}
	if global.Agent.ReasoningEffort != "" {
		t.Fatal("Global config should not be modified")
	}
}

func TestMergeAgentConfig_EmptyAgentConfig(t *testing.T) {
	global := &Config{
		ModelTier: "medium",
		Agent: AgentConfig{
			MaxIterations: 25,
			MaxTokens:     8192,
		},
	}

	ag := &agents.Agent{
		Name:   "empty",
		Config: &agents.AgentConfig{},
	}

	result := MergeAgentConfig(global, ag)

	// Should be identical to global
	if result.ModelTier != global.ModelTier {
		t.Error("All values should match global")
	}
	if result.Agent.MaxIterations != global.Agent.MaxIterations {
		t.Error("All values should match global")
	}
}

func strPtr(s string) *string {
	return &s
}
