package agent

import (
	"sync"
)

// Default per-tool result character budgets.
const (
	BudgetBash     = 10000
	BudgetFileRead = 20000
	BudgetGrep     = 5000
	BudgetHTTP     = 15000
	BudgetDefault  = 8000
)

// truncationNotice is appended to truncated content.
const truncationNotice = "\n\n[Output truncated: exceeded character budget]"

// ToolResultBudget enforces per-tool-result character limits.
// Each tool type has a maximum character count; results exceeding the limit
// are truncated with a notice. Safe for concurrent use.
type ToolResultBudget struct {
	mu     sync.RWMutex
	limits map[string]int
}

// NewToolResultBudget creates a ToolResultBudget with default per-tool limits.
func NewToolResultBudget() *ToolResultBudget {
	return &ToolResultBudget{
		limits: map[string]int{
			"bash":      BudgetBash,
			"file_read": BudgetFileRead,
			"grep":      BudgetGrep,
			"http":      BudgetHTTP,
		},
	}
}

// Check evaluates tool result content against the character budget for the
// given tool name. If accepted is true the content passes without modification;
// if false, truncated contains a truncated version of the content.
func (b *ToolResultBudget) Check(name string, content string) (accepted bool, truncated string) {
	b.mu.RLock()
	limit := b.limitFor(name)
	b.mu.RUnlock()

	if len(content) <= limit {
		return true, content
	}
	trunc := content[:limit] + truncationNotice
	return false, trunc
}

// SetLimit overrides the character budget for a specific tool name.
// A limit of 0 or less disables truncation for that tool.
func (b *ToolResultBudget) SetLimit(name string, limit int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if limit <= 0 {
		delete(b.limits, name)
		return
	}
	b.limits[name] = limit
}

// Limit returns the character budget for the given tool name.
func (b *ToolResultBudget) Limit(name string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.limitFor(name)
}

// limitFor returns the budget for a tool name without locking.
// Caller must hold at least a read lock.
func (b *ToolResultBudget) limitFor(name string) int {
	if limit, ok := b.limits[name]; ok {
		return limit
	}
	return BudgetDefault
}
