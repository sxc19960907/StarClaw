package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CompactResult holds data for a single tool result in the compact view.
type CompactResult struct {
	Name    string
	Args    string
	Content string
	IsError bool
	Elapsed time.Duration
}

// CompactView manages a list of compacted tool results and renders them as
// a single summary line in the TUI.
type CompactView struct {
	results []CompactResult
}

// NewCompactView creates a new CompactView.
func NewCompactView() *CompactView {
	return &CompactView{}
}

// Add appends a tool result to the view.
func (cv *CompactView) Add(name, args, content string, isError bool, elapsed time.Duration) {
	cv.results = append(cv.results, CompactResult{
		Name:    name,
		Args:    args,
		Content: content,
		IsError: isError,
		Elapsed: elapsed,
	})
}

// AddEntry appends a CompactResult to the view.
func (cv *CompactView) AddEntry(entry CompactResult) {
	cv.results = append(cv.results, entry)
}

// Clear removes all results from the view.
func (cv *CompactView) Clear() {
	cv.results = cv.results[:0]
}

// Summary returns a single-line summary of all tool results.
func (cv *CompactView) Summary() string {
	if len(cv.results) == 0 {
		return ""
	}

	var errCount int
	for _, r := range cv.results {
		if r.IsError {
			errCount++
		}
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	successIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	errorIcon := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")

	var line string
	if errCount == 0 {
		line = fmt.Sprintf("⏵ %d tools used  %s", len(cv.results), successIcon)
	} else {
		okCount := len(cv.results) - errCount
		line = fmt.Sprintf("⏵ %d tools used  %s%d %s%d", len(cv.results), successIcon, okCount, errorIcon, errCount)
	}
	return dimStyle.Render(line)
}
