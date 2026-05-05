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

	// Agent overrides should take effect
	if result.ModelTier != "opus" {
		t.Errorf("ModelTier = %q, want 'opus'", result.ModelTier)
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
