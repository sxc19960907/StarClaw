package agent

import (
	"sync"
	"testing"
)

func TestNewToolBudget(t *testing.T) {
	tb := NewToolBudget(100)
	if tb == nil {
		t.Fatal("NewToolBudget returned nil")
	}
}

func TestToolBudget_Consume_Under(t *testing.T) {
	tb := NewToolBudget(100)
	if !tb.Consume("web_search", 50) {
		t.Error("expected Consume to return true when under budget")
	}
}

func TestToolBudget_Consume_Exact(t *testing.T) {
	tb := NewToolBudget(100)
	if !tb.Consume("web_search", 100) {
		t.Error("expected Consume to return true when exactly at budget")
	}
}

func TestToolBudget_Consume_Over(t *testing.T) {
	tb := NewToolBudget(100)
	if !tb.Consume("web_search", 50) {
		t.Fatal("first consume should succeed")
	}
	if tb.Consume("web_search", 60) {
		t.Error("expected Consume to return false when over budget")
	}
}

func TestToolBudget_Consume_Negative(t *testing.T) {
	tb := NewToolBudget(100)
	if !tb.Consume("web_search", -1) {
		t.Error("negative chars should be allowed")
	}
}

func TestToolBudget_Consume_Zero(t *testing.T) {
	tb := NewToolBudget(100)
	if !tb.Consume("web_search", 0) {
		t.Error("zero chars should be allowed")
	}
}

func TestToolBudget_PerToolBudget(t *testing.T) {
	tb := NewToolBudget(100)

	// Two different tools each consume up to their own limit.
	if !tb.Consume("web_search", 100) {
		t.Error("web_search at limit should succeed")
	}
	if !tb.Consume("file_read", 100) {
		t.Error("file_read at limit should succeed independently")
	}

	// Each tool's individual budget is independent.
	if tb.Consume("web_search", 1) {
		t.Error("web_search over limit should fail")
	}
	if tb.Consume("file_read", 1) {
		t.Error("file_read over limit should fail")
	}
}

func TestToolBudget_Remaining_Basic(t *testing.T) {
	tb := NewToolBudget(100)
	if r := tb.Remaining(); r != 100 {
		t.Errorf("initial remaining should be 100, got %d", r)
	}
	tb.Consume("web_search", 30)
	if r := tb.Remaining(); r != 70 {
		t.Errorf("remaining should be 70, got %d", r)
	}
}

func TestToolBudget_Remaining_MultipleTools(t *testing.T) {
	tb := NewToolBudget(100)
	tb.Consume("web_search", 30)
	tb.Consume("file_read", 40)
	if r := tb.Remaining(); r != 30 {
		t.Errorf("remaining should be 30, got %d", r)
	}
}

func TestToolBudget_Remaining_Exhausted(t *testing.T) {
	tb := NewToolBudget(100)
	tb.Consume("web_search", 100)
	tb.Consume("file_read", 100)
	if r := tb.Remaining(); r != 0 {
		t.Errorf("remaining should be 0 when exhausted, got %d", r)
	}
}

func TestToolBudget_Unlimited(t *testing.T) {
	tb := NewToolBudget(0)
	if !tb.Consume("web_search", 1000000) {
		t.Error("unlimited budget should allow large consumption")
	}
	if r := tb.Remaining(); r != -1 {
		t.Errorf("unlimited remaining should be -1, got %d", r)
	}
}

func TestToolBudget_NegativeMax(t *testing.T) {
	tb := NewToolBudget(-1)
	if !tb.Consume("web_search", 1000000) {
		t.Error("negative max should behave as unlimited")
	}
}

func TestToolBudget_Concurrent(t *testing.T) {
	tb := NewToolBudget(1000)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tb.Consume("web_search", 10)
		}()
	}
	wg.Wait()
	if r := tb.Remaining(); r != 900 {
		t.Errorf("expected remaining 900 after concurrent usage, got %d", r)
	}
}
