package agent

import (
	"strings"
	"sync"
	"testing"
)

func TestToolResultBudgetDefault(t *testing.T) {
	b := NewToolResultBudget()
	if limit := b.Limit("unknown_tool"); limit != BudgetDefault {
		t.Errorf("expected default budget %d, got %d", BudgetDefault, limit)
	}
}

func TestToolResultBudgetPerToolLimits(t *testing.T) {
	b := NewToolResultBudget()
	tests := []struct {
		name      string
		tool      string
		expected  int
	}{
		{"bash", "bash", BudgetBash},
		{"file_read", "file_read", BudgetFileRead},
		{"grep", "grep", BudgetGrep},
		{"http", "http", BudgetHTTP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if limit := b.Limit(tt.tool); limit != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, limit)
			}
		})
	}
}

func TestToolResultBudgetWithinLimit(t *testing.T) {
	b := NewToolResultBudget()
	content := "short content"
	accepted, truncated := b.Check("bash", content)
	if !accepted {
		t.Error("expected content within budget to be accepted")
	}
	if truncated != content {
		t.Error("expected original content when within budget")
	}
}

func TestToolResultBudgetExceedsLimit(t *testing.T) {
	b := NewToolResultBudget()
	// Create content that exceeds the bash limit (10K).
	content := strings.Repeat("a", BudgetBash+100)
	accepted, truncated := b.Check("bash", content)
	if accepted {
		t.Error("expected content exceeding budget to be rejected")
	}
	if len(truncated) > BudgetBash+len(truncationNotice) {
		t.Errorf("truncated content too long: %d > %d", len(truncated), BudgetBash+len(truncationNotice))
	}
	if !strings.HasSuffix(truncated, truncationNotice) {
		t.Error("expected truncated content to end with truncation notice")
	}
	// Verify the truncated content contains the right prefix.
	if !strings.HasPrefix(truncated, content[:BudgetBash]) {
		t.Error("expected truncated content to start with original prefix")
	}
}

func TestToolResultBudgetDefaultTool(t *testing.T) {
	b := NewToolResultBudget()
	// Default budget is 8K - a 9K string should be truncated.
	content := strings.Repeat("b", BudgetDefault+500)
	accepted, truncated := b.Check("custom_tool", content)
	if accepted {
		t.Error("expected default budget enforcement")
	}
	if len(truncated) > BudgetDefault+len(truncationNotice) {
		t.Errorf("expected truncated to be <= %d, got %d", BudgetDefault+len(truncationNotice), len(truncated))
	}
}

func TestToolResultBudgetSetLimit(t *testing.T) {
	b := NewToolResultBudget()
	b.SetLimit("bash", 500)

	// Content within new limit.
	accepted, _ := b.Check("bash", "a")
	if !accepted {
		t.Error("expected content within custom limit to be accepted")
	}

	// Content exceeding new limit.
	content := strings.Repeat("c", 600)
	accepted, truncated := b.Check("bash", content)
	if accepted {
		t.Error("expected content exceeding custom limit to be rejected")
	}
	if !strings.HasSuffix(truncated, truncationNotice) {
		t.Error("expected truncation notice")
	}

	// Verify the limit was updated.
	if limit := b.Limit("bash"); limit != 500 {
		t.Errorf("expected limit 500, got %d", limit)
	}
}

func TestToolResultBudgetRemoveLimit(t *testing.T) {
	b := NewToolResultBudget()
	b.SetLimit("bash", 0) // zero removes limit, falling back to default

	if limit := b.Limit("bash"); limit != BudgetDefault {
		t.Errorf("expected default budget %d after removal, got %d", BudgetDefault, limit)
	}
}

func TestToolResultBudgetConcurrent(t *testing.T) {
	b := NewToolResultBudget()
	var wg sync.WaitGroup
	n := 50

	// Concurrent reads and writes.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Limit("bash")
			b.Check("bash", "test content")
		}()
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b.SetLimit("tool_x", i*100)
		}(i)
	}

	wg.Wait()
	// Just ensure no race or deadlock occurred.
}

func TestToolResultBudgetExactlyAtLimit(t *testing.T) {
	b := NewToolResultBudget()
	content := strings.Repeat("d", BudgetGrep)
	accepted, truncated := b.Check("grep", content)
	if !accepted {
		t.Error("expected content exactly at limit to be accepted")
	}
	if truncated != content {
		t.Error("expected original content when at limit")
	}
}
