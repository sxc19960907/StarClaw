package agent

import (
	"sync"
	"testing"
)

func TestNewUsageTracker(t *testing.T) {
	ut := NewUsageTracker()
	if ut == nil {
		t.Fatal("NewUsageTracker returned nil")
	}
}

func TestUsageTracker_AddUsage(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 100, 50)

	stats := ut.SessionStats()
	if stats.Turns != 1 {
		t.Errorf("expected 1 turn, got %d", stats.Turns)
	}
	if stats.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 50 {
		t.Errorf("expected 50 output tokens, got %d", stats.OutputTokens)
	}
	if stats.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got %d", stats.TotalTokens)
	}
}

func TestUsageTracker_AddUsage_MultipleTurns(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 100, 50)
	ut.AddUsage(2, 200, 30)

	stats := ut.SessionStats()
	if stats.Turns != 2 {
		t.Errorf("expected 2 turns, got %d", stats.Turns)
	}
	if stats.InputTokens != 300 {
		t.Errorf("expected 300 input tokens, got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 80 {
		t.Errorf("expected 80 output tokens, got %d", stats.OutputTokens)
	}
	if stats.TotalTokens != 380 {
		t.Errorf("expected 380 total tokens, got %d", stats.TotalTokens)
	}
}

func TestUsageTracker_AddUsage_ZeroValues(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(0, 0, 0)

	stats := ut.SessionStats()
	if stats.Turns != 1 {
		t.Errorf("expected 1 turn, got %d", stats.Turns)
	}
	if stats.TotalTokens != 0 {
		t.Errorf("expected 0 total tokens, got %d", stats.TotalTokens)
	}
}

func TestUsageTracker_TotalCost_Zero(t *testing.T) {
	ut := NewUsageTracker()
	cost := ut.TotalCost()
	if cost != 0 {
		t.Errorf("expected 0 cost, got %f", cost)
	}
}

func TestUsageTracker_TotalCost_InputOnly(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 1_000_000, 0)
	cost := ut.TotalCost()
	if cost != 3.0 {
		t.Errorf("expected $3.00 for 1M input tokens, got %f", cost)
	}
}

func TestUsageTracker_TotalCost_OutputOnly(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 0, 1_000_000)
	cost := ut.TotalCost()
	if cost != 15.0 {
		t.Errorf("expected $15.00 for 1M output tokens, got %f", cost)
	}
}

func TestUsageTracker_TotalCost_Combined(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 1000, 500)

	expectedCost := float64(1000)*3.0/1_000_000 + float64(500)*15.0/1_000_000
	cost := ut.TotalCost()
	if cost != expectedCost {
		t.Errorf("expected %f cost, got %f", expectedCost, cost)
	}
}

func TestUsageTracker_TotalCost_MultipleTurns(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 500, 200)
	ut.AddUsage(2, 1500, 800)

	expectedCost := float64(2000)*3.0/1_000_000 + float64(1000)*15.0/1_000_000
	cost := ut.TotalCost()
	if cost != expectedCost {
		t.Errorf("expected %f cost, got %f", expectedCost, cost)
	}
}

func TestUsageTracker_SessionStats_SnapshotIndependent(t *testing.T) {
	ut := NewUsageTracker()
	ut.AddUsage(1, 100, 50)

	stats := ut.SessionStats()
	stats.Turns = 999 // modify snapshot

	// Verify original is unchanged.
	stats2 := ut.SessionStats()
	if stats2.Turns != 1 {
		t.Errorf("original should be unchanged, got %d turns", stats2.Turns)
	}
}

func TestUsageTracker_Concurrent(t *testing.T) {
	ut := NewUsageTracker()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ut.AddUsage(1, 100, 50)
		}()
	}

	wg.Wait()
	stats := ut.SessionStats()
	if stats.Turns != 20 {
		t.Errorf("expected 20 turns, got %d", stats.Turns)
	}
	if stats.InputTokens != 2000 {
		t.Errorf("expected 2000 input tokens, got %d", stats.InputTokens)
	}
}
