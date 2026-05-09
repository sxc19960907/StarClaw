package agent

import "sync"

// Stats holds session usage statistics.
type Stats struct {
	Turns        int
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64
}

// UsageTracker tracks token usage and cost across turns. Safe for concurrent use.
type UsageTracker struct {
	mu    sync.Mutex
	stats Stats
}

// NewUsageTracker creates a new UsageTracker.
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{}
}

// AddUsage records usage for a single turn. Turns counter increments on each call.
// Input and output tokens are added to running totals.
func (ut *UsageTracker) AddUsage(turn, input, output int) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	ut.stats.Turns++
	ut.stats.InputTokens += input
	ut.stats.OutputTokens += output
	ut.stats.TotalTokens += input + output
}

// TotalCost calculates total cost based on aggregated token usage.
// Uses standard Claude 4 Sonnet pricing: $3 per million input tokens,
// $15 per million output tokens.
func (ut *UsageTracker) TotalCost() float64 {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	inputCost := float64(ut.stats.InputTokens) * 3.0 / 1_000_000
	outputCost := float64(ut.stats.OutputTokens) * 15.0 / 1_000_000
	return inputCost + outputCost
}

// SessionStats returns a snapshot of current session statistics.
func (ut *UsageTracker) SessionStats() Stats {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	return ut.stats
}
