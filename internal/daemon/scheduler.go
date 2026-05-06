package daemon

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/schedule"
)

const maxConcurrentSchedules = 5

// Scheduler evaluates cron schedules each minute and fires agent runs for due entries.
type Scheduler struct {
	manager   *schedule.Manager
	deps      *ServerDeps
	mu        sync.Mutex
	lastFired map[string]time.Time // scheduleID -> last fired minute (truncated)
	sem       chan struct{}         // bounded concurrency
}

// NewScheduler creates a Scheduler that evaluates schedules from mgr.
func NewScheduler(mgr *schedule.Manager, deps *ServerDeps) *Scheduler {
	return &Scheduler{
		manager:   mgr,
		deps:      deps,
		lastFired: make(map[string]time.Time),
		sem:       make(chan struct{}, maxConcurrentSchedules),
	}
}

// Start blocks until ctx is cancelled. Fires immediately, then aligns to the
// next wall-clock minute boundary and evaluates every minute thereafter.
func (s *Scheduler) Start(ctx context.Context) {
	// Catch-up: evaluate immediately on startup.
	s.tick(ctx)

	// Align to next wall-clock minute boundary.
	now := time.Now()
	next := now.Truncate(time.Minute).Add(time.Minute)
	select {
	case <-time.After(next.Sub(now)):
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Also fire at the first aligned minute boundary.
	s.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick evaluates due schedules and fires goroutines for each.
// Non-blocking: if all concurrency slots are full, the schedule is logged and
// dropped rather than blocking and potentially missing the next minute boundary.
func (s *Scheduler) tick(ctx context.Context) {
	due := s.EvaluateDue(time.Now())
	for _, sc := range due {
		select {
		case s.sem <- struct{}{}:
			go func(sc schedule.Schedule) {
				defer func() { <-s.sem }()
				s.runSchedule(ctx, sc)
			}(sc)
		default:
			log.Printf("scheduler: skipping schedule %s (all %d slots busy)", sc.ID, maxConcurrentSchedules)
		}
	}
}

// EvaluateDue returns schedules that are due at the given time.
// Exported for testing.
func (s *Scheduler) EvaluateDue(now time.Time) []schedule.Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedules, err := s.manager.List()
	if err != nil {
		log.Printf("scheduler: failed to list schedules: %v", err)
		return nil
	}

	// Build set of active IDs for pruning.
	activeIDs := make(map[string]struct{}, len(schedules))
	for _, sc := range schedules {
		activeIDs[sc.ID] = struct{}{}
	}
	// Prune lastFired entries for deleted schedules.
	for id := range s.lastFired {
		if _, ok := activeIDs[id]; !ok {
			delete(s.lastFired, id)
		}
	}

	truncated := now.Truncate(time.Minute)
	var due []schedule.Schedule
	for _, sc := range schedules {
		if !sc.Enabled {
			continue
		}
		if !cronIsDue(sc.Cron, now) {
			continue
		}
		// Dedup: skip if already fired this minute.
		if last, ok := s.lastFired[sc.ID]; ok && last.Equal(truncated) {
			continue
		}
		s.lastFired[sc.ID] = truncated
		due = append(due, sc)
	}
	return due
}

// runSchedule fires a single scheduled agent run.
func (s *Scheduler) runSchedule(ctx context.Context, sc schedule.Schedule) {
	req := RunAgentRequest{
		Text:    sc.Prompt,
		Agent:   sc.Agent,
		Source:  ChannelSchedule,
		Channel: ChannelSchedule + "-" + sc.ID,
		Sender:  "scheduler",
		// Named agents resume their single long-lived session.
		// Default agent (no name) gets a fresh session per run.
		NewSession: sc.Agent == "",
	}

	// Use a fire-and-forget event handler that logs results.
	handler := &scheduleEventHandler{}
	_, err := RunAgent(ctx, s.deps, req, handler)
	if err != nil {
		log.Printf("scheduler: run agent for schedule %s failed: %v", sc.ID, err)
		return
	}
	log.Printf("scheduler: completed agent run for schedule %s (agent=%q)", sc.ID, sc.Agent)
}

// scheduleEventHandler is a silent EventHandler for scheduled agent runs.
type scheduleEventHandler struct{}

func (h *scheduleEventHandler) OnToolCall(name string, args string)    {}
func (h *scheduleEventHandler) OnToolResult(name string, result agent.ToolResult) {
	if result.IsError {
		log.Printf("scheduler: tool %s returned error: %s", name, result.Content)
	}
}
func (h *scheduleEventHandler) OnText(text string)            {}
func (h *scheduleEventHandler) OnUsage(usage client.Usage)    {}

// cronIsDue evaluates a 5-field cron expression against the given time.
// This is a minimal implementation that supports:
//   - * (wildcard)
//   - N (specific value)
//   - */N (step)
//   - A-B (range)
//   - A-B/N (range with step)
//   - A,B,C (list)
//   - Named expressions: @yearly, @annually, @monthly, @weekly, @daily, @hourly, @reboot
//
// Month names (JAN, FEB, etc.) and weekday names (SUN, MON, etc.) are not supported.
// @reboot always returns true (simplification: it fires on every tick).
func cronIsDue(expr string, now time.Time) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}

	switch expr {
	case "@yearly", "@annually":
		return now.Month() == time.January && now.Day() == 1 && now.Hour() == 0 && now.Minute() == 0
	case "@monthly":
		return now.Day() == 1 && now.Hour() == 0 && now.Minute() == 0
	case "@weekly":
		return now.Weekday() == time.Sunday && now.Hour() == 0 && now.Minute() == 0
	case "@daily":
		return now.Hour() == 0 && now.Minute() == 0
	case "@hourly":
		return now.Minute() == 0
	case "@reboot":
		return true
	}

	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}

	// minute (0-59), hour (0-23), day-of-month (1-31), month (1-12), day-of-week (0-6, Sun=0)
	if !cronFieldMatches(fields[0], now.Minute(), 0, 59) {
		return false
	}
	if !cronFieldMatches(fields[1], now.Hour(), 0, 23) {
		return false
	}
	if !cronFieldMatches(fields[2], now.Day(), 1, 31) {
		return false
	}
	if !cronFieldMatches(fields[3], int(now.Month()), 1, 12) {
		return false
	}
	if !cronFieldMatches(fields[4], int(now.Weekday()), 0, 6) {
		return false
	}
	return true
}

// cronFieldMatches checks whether a single cron field matches the given value.
func cronFieldMatches(field string, value int, min, max int) bool {
	// List: A,B,C
	if strings.Contains(field, ",") {
		for _, part := range strings.Split(field, ",") {
			if cronFieldMatches(strings.TrimSpace(part), value, min, max) {
				return true
			}
		}
		return false
	}

	// Step: */N or A-B/N
	if strings.Contains(field, "/") {
		parts := strings.SplitN(field, "/", 2)
		base := strings.TrimSpace(parts[0])
		step, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || step <= 0 {
			return false
		}

		if base == "*" {
			return value%step == 0
		}

		if strings.Contains(base, "-") {
			rangeParts := strings.SplitN(base, "-", 2)
			low, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			high, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 != nil || err2 != nil {
				return false
			}
			if value < low || value > high {
				return false
			}
			return (value-low)%step == 0
		}

		return false
	}

	// Range: A-B
	if strings.Contains(field, "-") {
		parts := strings.SplitN(field, "-", 2)
		low, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		high, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return false
		}
		return value >= low && value <= high
	}

	// Wildcard
	if field == "*" {
		return true
	}

	// Specific number
	n, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return n == value
}
