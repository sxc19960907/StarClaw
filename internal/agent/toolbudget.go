package agent

import "sync"

// ToolBudget tracks per-tool character output budgets.
// Each tool type has an independent budget of maxChars. Safe for concurrent use.
type ToolBudget struct {
	mu       sync.Mutex
	maxChars int
	used     map[string]int
}

// NewToolBudget creates a ToolBudget with a per-tool-type max character limit.
// If maxChars is 0 or negative, budgets are unlimited.
func NewToolBudget(maxChars int) *ToolBudget {
	return &ToolBudget{
		maxChars: maxChars,
		used:     make(map[string]int),
	}
}

// Consume records chars used by a tool type.
// Returns false if consuming chars would exceed the budget for this tool type,
// in which case no chars are consumed.
func (tb *ToolBudget) Consume(name string, chars int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if chars < 0 {
		return true
	}
	if tb.maxChars <= 0 {
		// Unlimited budget.
		tb.used[name] += chars
		return true
	}
	current := tb.used[name] + chars
	if current > tb.maxChars {
		return false
	}
	tb.used[name] = current
	return true
}

// Remaining returns the total remaining character budget across all tools.
// Returns -1 if the budget is unlimited (maxChars was 0 or negative).
func (tb *ToolBudget) Remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.maxChars <= 0 {
		return -1
	}
	total := 0
	for _, v := range tb.used {
		total += v
	}
	remaining := tb.maxChars - total
	if remaining < 0 {
		return 0
	}
	return remaining
}
