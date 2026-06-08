package agent

import (
	"fmt"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
)

const (
	TokenBudgetStatusDisabled  = "disabled"
	TokenBudgetStatusOK        = "ok"
	TokenBudgetStatusUnknown   = "unknown"
	TokenBudgetStatusExhausted = "exhausted"

	RunStatusBudgetExhausted = "budget_exhausted"
)

// TokenBudget constrains token usage for a single agent run.
type TokenBudget struct {
	MaxInputTokens  int  `mapstructure:"max_input_tokens" yaml:"max_input_tokens,omitempty" json:"max_input_tokens,omitempty"`
	MaxOutputTokens int  `mapstructure:"max_output_tokens" yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`
	MaxTotalTokens  int  `mapstructure:"max_total_tokens" yaml:"max_total_tokens,omitempty" json:"max_total_tokens,omitempty"`
	HardStop        bool `mapstructure:"hard_stop" yaml:"hard_stop,omitempty" json:"hard_stop,omitempty"`
}

// TokenBudgetFromAgent converts config.AgentConfig into an agent runtime budget.
func TokenBudgetFromAgent(cfg config.AgentConfig) TokenBudget {
	return TokenBudget{
		MaxInputTokens:  cfg.TokenBudget.MaxInputTokens,
		MaxOutputTokens: cfg.TokenBudget.MaxOutputTokens,
		MaxTotalTokens:  cfg.TokenBudget.MaxTotalTokens,
		HardStop:        cfg.TokenBudget.HardStop,
	}
}

// TokenBudgetUsage is the runtime-safe status view surfaced to callers.
type TokenBudgetUsage struct {
	Status       string `json:"status"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
	UnknownTurns int    `json:"unknown_turns,omitempty"`
	Detail       string `json:"detail,omitempty"`
}

type budgetDecision struct {
	Stop   bool
	Status TokenBudgetUsage
}

type tokenBudgetTracker struct {
	budget        TokenBudget
	inputTokens   int
	outputTokens  int
	inputEstimate int
	unknownTurns  int
	lastDetail    string
	exhausted     bool
}

func newTokenBudgetTracker(budget TokenBudget) *tokenBudgetTracker {
	if !budget.Enabled() {
		return nil
	}
	return &tokenBudgetTracker{budget: budget}
}

// Enabled reports whether any token limit is configured.
func (b TokenBudget) Enabled() bool {
	return b.MaxInputTokens > 0 || b.MaxOutputTokens > 0 || b.MaxTotalTokens > 0
}

func (t *tokenBudgetTracker) AddUsage(usage client.Usage) TokenBudgetUsage {
	if t == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	if usage.InputTokens <= 0 && usage.OutputTokens <= 0 {
		t.unknownTurns++
		t.lastDetail = "provider did not return token usage for this response"
		return t.Status()
	}
	t.inputTokens += max(usage.InputTokens, 0)
	t.outputTokens += max(usage.OutputTokens, 0)
	t.inputEstimate = 0
	t.lastDetail = ""
	if detail := t.exhaustedDetail(0); detail != "" {
		t.exhausted = true
		t.lastDetail = detail
	}
	return t.Status()
}

func (t *tokenBudgetTracker) SetProjectedInput(tokens int) TokenBudgetUsage {
	if t == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	if t.inputTokens == 0 {
		t.inputEstimate = max(tokens, 0)
	}
	return t.Status()
}

func (t *tokenBudgetTracker) Status() TokenBudgetUsage {
	if t == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	status := TokenBudgetStatusOK
	detail := t.exhaustedDetail(0)
	if detail != "" || t.exhausted {
		status = TokenBudgetStatusExhausted
		if detail == "" {
			detail = t.lastDetail
		}
	} else if t.unknownTurns > 0 {
		status = TokenBudgetStatusUnknown
		detail = t.lastDetail
	}
	return TokenBudgetUsage{
		Status:       status,
		InputTokens:  t.effectiveInputTokens(),
		OutputTokens: t.outputTokens,
		TotalTokens:  t.effectiveInputTokens() + t.outputTokens,
		UnknownTurns: t.unknownTurns,
		Detail:       detail,
	}
}

func (t *tokenBudgetTracker) Decision(projectedOutput int) budgetDecision {
	status := t.Status()
	if t == nil || !t.budget.HardStop {
		return budgetDecision{Status: status}
	}
	if detail := t.exhaustedDetail(projectedOutput); detail != "" {
		t.exhausted = true
		t.lastDetail = detail
		status.Status = TokenBudgetStatusExhausted
		status.Detail = detail
		return budgetDecision{Stop: true, Status: status}
	}
	return budgetDecision{Status: status}
}

func (t *tokenBudgetTracker) effectiveInputTokens() int {
	if t == nil {
		return 0
	}
	if t.inputTokens > 0 {
		return t.inputTokens
	}
	return t.inputEstimate
}

func (t *tokenBudgetTracker) exhaustedDetail(projectedOutput int) string {
	if t == nil {
		return ""
	}
	projectedOutput = max(projectedOutput, 0)
	inputTokens := t.effectiveInputTokens()
	projectedTotal := inputTokens + t.outputTokens + projectedOutput
	if t.budget.MaxInputTokens > 0 && inputTokens >= t.budget.MaxInputTokens {
		return fmt.Sprintf("input token budget exhausted: used %d of %d", inputTokens, t.budget.MaxInputTokens)
	}
	if t.budget.MaxOutputTokens > 0 && t.outputTokens+projectedOutput >= t.budget.MaxOutputTokens {
		return fmt.Sprintf("output token budget exhausted: projected %d of %d", t.outputTokens+projectedOutput, t.budget.MaxOutputTokens)
	}
	if t.budget.MaxTotalTokens > 0 && projectedTotal >= t.budget.MaxTotalTokens {
		return fmt.Sprintf("total token budget exhausted: projected %d of %d", projectedTotal, t.budget.MaxTotalTokens)
	}
	return ""
}
