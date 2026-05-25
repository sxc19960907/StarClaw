package config

import "github.com/starclaw/starclaw/internal/agents"

// MergeAgentConfig applies agent-specific config overrides on top of the global config.
// Agent config values take precedence when non-nil.
// Returns a copy so the original is unchanged.
func MergeAgentConfig(global *Config, ag *agents.Agent) *Config {
	if ag == nil || ag.Config == nil {
		return global
	}

	merged := *global // shallow copy

	ac := ag.Config

	// Agent model overrides
	if ac.Agent != nil {
		if ac.Agent.MaxIterations != nil {
			merged.Agent.MaxIterations = *ac.Agent.MaxIterations
		}
		if ac.Agent.Temperature != nil {
			merged.Agent.Temperature = *ac.Agent.Temperature
		}
		if ac.Agent.MaxTokens != nil {
			merged.Agent.MaxTokens = *ac.Agent.MaxTokens
		}
		if ac.Agent.ContextWindow != nil {
			merged.Agent.ContextWindow = *ac.Agent.ContextWindow
		}
		if ac.Agent.Model != nil {
			merged.Agent.Model = *ac.Agent.Model
		}
		if ac.Agent.Thinking != nil {
			merged.Agent.Thinking = *ac.Agent.Thinking
		}
		if ac.Agent.ThinkingMode != nil {
			merged.Agent.ThinkingMode = *ac.Agent.ThinkingMode
		}
		if ac.Agent.ThinkingBudget != nil {
			merged.Agent.ThinkingBudget = *ac.Agent.ThinkingBudget
		}
		if ac.Agent.ReasoningEffort != nil {
			merged.Agent.ReasoningEffort = *ac.Agent.ReasoningEffort
		}
	}

	return &merged
}
