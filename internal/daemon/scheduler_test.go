package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/schedule"
)

// --- cronIsDue tests ---

func TestCronIsDueWildcard(t *testing.T) {
	// * * * * * matches everything
	now := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)
	if !cronIsDue("* * * * *", now) {
		t.Error("expected * * * * * to match any time")
	}
}

func TestCronIsDueSpecificMinute(t *testing.T) {
	// 30 * * * * matches minute 30
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if !cronIsDue("30 * * * *", now) {
		t.Error("expected 30 * * * * to match minute 30")
	}

	now2 := time.Date(2024, 6, 15, 10, 31, 0, 0, time.UTC)
	if cronIsDue("30 * * * *", now2) {
		t.Error("expected 30 * * * * not to match minute 31")
	}
}

func TestCronIsDueStep(t *testing.T) {
	// */5 * * * * matches every 5th minute
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if !cronIsDue("*/5 * * * *", now) {
		t.Error("expected */5 * * * * to match minute 30 (30 % 5 == 0)")
	}

	now2 := time.Date(2024, 6, 15, 10, 31, 0, 0, time.UTC)
	if cronIsDue("*/5 * * * *", now2) {
		t.Error("expected */5 * * * * not to match minute 31 (31 % 5 != 0)")
	}
}

func TestCronIsDueRange(t *testing.T) {
	// 0 9-17 * * * matches 9am-5pm
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if !cronIsDue("0 9-17 * * *", now) {
		t.Error("expected 0 9-17 * * * to match hour 12")
	}

	now2 := time.Date(2024, 6, 15, 18, 0, 0, 0, time.UTC)
	if cronIsDue("0 9-17 * * *", now2) {
		t.Error("expected 0 9-17 * * * not to match hour 18")
	}
}

func TestCronIsDueRangeWithStep(t *testing.T) {
	// 0 0 1-10/3 * * matches days 1,4,7,10
	now := time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC)
	if !cronIsDue("0 0 1-10/3 * *", now) {
		t.Error("expected 0 0 1-10/3 * * to match day 4")
	}

	now2 := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)
	if cronIsDue("0 0 1-10/3 * *", now2) {
		t.Error("expected 0 0 1-10/3 * * not to match day 5")
	}
}

func TestCronIsDueList(t *testing.T) {
	// 0 0 * * 1,3,5 matches Mon, Wed, Fri
	now := time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC) // Monday
	if !cronIsDue("0 0 * * 1,3,5", now) {
		t.Error("expected 0 0 * * 1,3,5 to match Monday (1)")
	}

	now2 := time.Date(2024, 6, 18, 0, 0, 0, 0, time.UTC) // Tuesday
	if cronIsDue("0 0 * * 1,3,5", now2) {
		t.Error("expected 0 0 * * 1,3,5 not to match Tuesday (2)")
	}
}

func TestCronIsDueNamed(t *testing.T) {
	midnight := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	if !cronIsDue("@daily", midnight) {
		t.Error("expected @daily to match midnight")
	}
	if !cronIsDue("@hourly", time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)) {
		t.Error("expected @hourly to match :00")
	}
	if cronIsDue("@daily", time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)) {
		t.Error("expected @daily not to match 10:00")
	}
}

func TestCronIsDueInvalid(t *testing.T) {
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	if cronIsDue("", now) {
		t.Error("expected empty expression to return false")
	}
	if cronIsDue("invalid", now) {
		t.Error("expected invalid expression to return false")
	}
	if cronIsDue("* * * *", now) { // only 4 fields
		t.Error("expected 4-field expression to return false")
	}
	if cronIsDue("* * * * * *", now) { // 6 fields
		t.Error("expected 6-field expression to return false")
	}
}

func TestCronIsDueReboot(t *testing.T) {
	// @reboot always returns true as a simplification
	if !cronIsDue("@reboot", time.Now()) {
		t.Error("expected @reboot to always return true")
	}
}

// --- Scheduler EvaluateDue tests ---

func TestSchedulerEvaluateDue(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	// Create schedules
	id1, err := mgr.Create("", "* * * * *", "every minute")
	if err != nil {
		t.Fatalf("Create id1: %v", err)
	}
	id2, err := mgr.Create("test-agent", "*/5 * * * *", "every 5 min")
	if err != nil {
		t.Fatalf("Create id2: %v", err)
	}
	id3, err := mgr.Create("", "0 0 * * *", "daily - disabled")
	if err != nil {
		t.Fatalf("Create id3: %v", err)
	}
	// Disable id3
	if err := mgr.Update(id3, &schedule.UpdateOpts{Enabled: boolPtr(false)}); err != nil {
		t.Fatalf("Update id3: %v", err)
	}

	// Choose a time where both id1 (* * * * *) and id2 (*/5 * * * *) are due.
	// 30 % 5 == 0, so both are due.
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	_ = id2

	due := s.EvaluateDue(now)

	if len(due) != 2 {
		t.Fatalf("expected 2 due schedules, got %d", len(due))
	}

	dueIDs := make(map[string]bool)
	for _, sc := range due {
		dueIDs[sc.ID] = true
	}
	if !dueIDs[id1] {
		t.Errorf("expected schedule %s (enabled, every minute) in due list", id1)
	}
	if dueIDs[id3] {
		t.Errorf("expected disabled schedule %s not in due list", id3)
	}
}

func TestSchedulerEvaluateDueDedup(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	_, err := mgr.Create("", "* * * * *", "every minute")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// First call should return the schedule.
	due1 := s.EvaluateDue(now)
	if len(due1) != 1 {
		t.Fatalf("expected 1 due schedule on first call, got %d", len(due1))
	}

	// Second call in same minute should return empty (dedup).
	due2 := s.EvaluateDue(now)
	if len(due2) != 0 {
		t.Errorf("expected 0 due schedules on second call (dedup), got %d", len(due2))
	}

	// Call at next minute should return the schedule again.
	later := now.Add(time.Minute)
	due3 := s.EvaluateDue(later)
	if len(due3) != 1 {
		t.Errorf("expected 1 due schedule at next minute, got %d", len(due3))
	}
}

func TestSchedulerEvaluateDueDisabledSkipped(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	id, err := mgr.Create("", "* * * * *", "every minute")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mgr.Update(id, &schedule.UpdateOpts{Enabled: boolPtr(false)}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	due := s.EvaluateDue(now)
	if len(due) != 0 {
		t.Errorf("expected 0 due schedules (disabled), got %d", len(due))
	}
}

func TestSchedulerEvaluateDueMalformedCron(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schedules.json")

	// Write schedules directly to bypass Create validation.
	schedules := []schedule.Schedule{
		{
			ID: "bad", Agent: "", Cron: "not-a-valid-cron",
			Prompt: "bad", Enabled: true, CreatedAt: time.Now(),
		},
		{
			ID: "good", Agent: "", Cron: "* * * * *",
			Prompt: "good", Enabled: true, CreatedAt: time.Now(),
		},
	}
	data, err := json.Marshal(schedules)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	mgr := schedule.NewManager(path)
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	due := s.EvaluateDue(now)

	// Only "good" should be due; "bad" should be silently skipped.
	if len(due) != 1 {
		t.Fatalf("expected 1 due schedule, got %d", len(due))
	}
	if due[0].ID != "good" {
		t.Errorf("expected schedule 'good' to be due, got %q", due[0].ID)
	}
}

func TestSchedulerEvaluateDuePruneDeleted(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	id, err := mgr.Create("", "* * * * *", "every minute")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// First call populates lastFired.
	due1 := s.EvaluateDue(now)
	if len(due1) != 1 {
		t.Fatalf("expected 1 due schedule, got %d", len(due1))
	}

	// Verify lastFired has the entry.
	s.mu.Lock()
	if _, ok := s.lastFired[id]; !ok {
		t.Error("expected lastFired to have entry after EvaluateDue")
	}
	s.mu.Unlock()

	// Delete the schedule from the manager.
	if err := mgr.Remove(id); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// Next EvaluateDue at a new minute should prune the entry.
	later := now.Add(time.Minute)
	due2 := s.EvaluateDue(later)
	if len(due2) != 0 {
		t.Errorf("expected 0 due schedules after deletion, got %d", len(due2))
	}

	// Verify lastFired was pruned.
	s.mu.Lock()
	if _, ok := s.lastFired[id]; ok {
		t.Error("expected lastFired entry to be pruned after deletion")
	}
	s.mu.Unlock()
}

// --- tick concurrency tests ---

func TestSchedulerTickDispatch(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	// Create more due schedules than max concurrency slots.
	for i := 0; i < maxConcurrentSchedules+3; i++ {
		_, err := mgr.Create("", "* * * * *", fmt.Sprintf("schedule %d", i))
		if err != nil {
			t.Fatalf("Create schedule %d: %v", i, err)
		}
	}

	// Tick should return without blocking and all schedules should be tracked
	// in lastFired (added by EvaluateDue before goroutines are dispatched).
	s.tick(context.Background())

	s.mu.Lock()
	count := len(s.lastFired)
	s.mu.Unlock()

	expected := maxConcurrentSchedules + 3
	if count != expected {
		t.Errorf("expected %d entries in lastFired, got %d", expected, count)
	}
}

func TestSchedulerTickEmpty(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	// Tick with no schedules should not panic.
	s.tick(context.Background())
}

// --- Start/Stop tests ---

func TestSchedulerStartRespectsCancellation(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	// Use a pre-cancelled context so Start returns immediately
	// (the alignment timer catches <-ctx.Done()).
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Start returned promptly.
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop after pre-cancelled context")
	}
}

func TestSchedulerStartCancelsDuringTick(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))
	deps := &ServerDeps{}
	s := NewScheduler(mgr, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Start(ctx)
		close(done)
	}()

	// Give Start a moment to begin (tick + alignment sleep).
	time.Sleep(10 * time.Millisecond)

	cancel()

	select {
	case <-done:
		// Start returned after cancellation.
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop after context cancellation during alignment")
	}
}

// --- helpers ---

func boolPtr(b bool) *bool {
	return &b
}
