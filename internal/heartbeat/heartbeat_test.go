package heartbeat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeAgent(t *testing.T, agentsDir, name string) {
	t.Helper()
	agentDir := filepath.Join(agentsDir, name)
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("test agent "+name), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeHeartbeatConfig(t *testing.T, agentsDir, name, interval, activeHours string) {
	t.Helper()
	var buf strings.Builder
	buf.WriteString("heartbeat:\n")
	buf.WriteString(fmt.Sprintf("  every: %s\n", interval))
	if activeHours != "" {
		buf.WriteString(fmt.Sprintf("  active_hours: %s\n", activeHours))
	}
	path := filepath.Join(agentsDir, name, "config.yaml")
	if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// New / Manager construction
// ---------------------------------------------------------------------------

func TestNew_CreatesEntriesForAgentsWithHeartbeat(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")
	writeAgent(t, agentsDir, "agent2")
	// agent2 deliberately has no heartbeat config.

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 1 {
		t.Fatalf("expected 1 agent entry, got %d", len(m.agents))
	}
	if m.agents[0].name != "agent1" {
		t.Errorf("expected agent1, got %s", m.agents[0].name)
	}
	if m.agents[0].interval != 5*time.Minute {
		t.Errorf("expected interval 5m, got %s", m.agents[0].interval)
	}
}

func TestNew_ReturnsEmptyForNoHeartbeatAgents(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeAgent(t, agentsDir, "agent2")
	// Neither agent has a heartbeat config.

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 0 {
		t.Errorf("expected 0 entries, got %d", len(m.agents))
	}
}

func TestNew_RejectsNilDeps(t *testing.T) {
	_, err := New("/tmp", nil)
	if err == nil {
		t.Fatal("expected error for nil deps")
	}
}

func TestNew_SkipsAgentWhenIntervalTooShort(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "500ms", "")

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 0 {
		t.Errorf("expected 0 entries when interval is too short, got %d", len(m.agents))
	}
}

func TestNew_SkipsAgentWhenInvalidInterval(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "not-a-duration", "")

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 0 {
		t.Errorf("expected 0 entries for invalid interval, got %d", len(m.agents))
	}
}

func TestNew_HandlesMissingAgentsDir(t *testing.T) {
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New("/nonexistent/agents", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 0 {
		t.Errorf("expected 0 entries for missing dir, got %d", len(m.agents))
	}
}

// ---------------------------------------------------------------------------
// IsHeartbeatOK
// ---------------------------------------------------------------------------

func TestIsHeartbeatOK(t *testing.T) {
	tests := []struct {
		reply string
		want  bool
	}{
		{"HEARTBEAT_OK", true},
		{"heartbeat_ok", true},
		{"  HEARTBEAT_OK  ", true},
		{"\nHEARTBEAT_OK\n", true},
		{"HEARTBEAT_OK and some extra text", false},
		{"Everything looks fine", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.reply, func(t *testing.T) {
			if got := IsHeartbeatOK(tt.reply); got != tt.want {
				t.Errorf("IsHeartbeatOK(%q) = %v, want %v", tt.reply, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FormatPrompt / FormatGoalPrompt
// ---------------------------------------------------------------------------

func TestFormatPrompt(t *testing.T) {
	checklist := "- Check disk\n- Check memory"
	got := FormatPrompt(checklist)
	if got == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !strings.Contains(got, "HEARTBEAT_OK") {
		t.Error("prompt should mention HEARTBEAT_OK")
	}
	if !strings.Contains(got, checklist) {
		t.Error("prompt should contain checklist")
	}
}

func TestFormatGoalPrompt(t *testing.T) {
	got := FormatGoalPrompt("## Goals\n- Do stuff")
	if !strings.Contains(got, "periodic check-in") {
		t.Error("goal prompt missing check-in text")
	}
	if !strings.Contains(got, "## Goals") {
		t.Error("goal prompt missing goals content")
	}
}

// ---------------------------------------------------------------------------
// ReadChecklist
// ---------------------------------------------------------------------------

func TestReadChecklist(t *testing.T) {
	dir := t.TempDir()

	// Missing file.
	content, err := ReadChecklist(filepath.Join(dir, "HEARTBEAT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		t.Errorf("expected empty for missing file, got %q", content)
	}

	// Empty / whitespace-only file.
	os.WriteFile(filepath.Join(dir, "HEARTBEAT.md"), []byte("   \n\n  "), 0o644)
	content, err = ReadChecklist(filepath.Join(dir, "HEARTBEAT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		t.Errorf("expected empty for whitespace-only file, got %q", content)
	}

	// Valid file.
	if err := os.WriteFile(filepath.Join(dir, "HEARTBEAT.md"), []byte("- Check disk\n- Check memory"), 0o644); err != nil {
		t.Fatal(err)
	}
	content, err = ReadChecklist(filepath.Join(dir, "HEARTBEAT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if content != "- Check disk\n- Check memory" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestReadChecklist_PermissionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "HEARTBEAT.md")
	if err := os.WriteFile(path, []byte("- Check disk"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(path, 0o644) // restore so temp dir cleanup can proceed
	}()

	content, err := ReadChecklist(path)
	if err == nil {
		t.Fatal("expected error for unreadable file")
	}
	if content != "" {
		t.Errorf("expected empty content on error, got %q", content)
	}
}

func TestReadChecklist_MaxSize(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("x", maxChecklistChars+1000)
	os.WriteFile(filepath.Join(dir, "HEARTBEAT.md"), []byte(big), 0o644)

	content, err := ReadChecklist(filepath.Join(dir, "HEARTBEAT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != maxChecklistChars {
		t.Errorf("expected %d chars (truncated), got %d", maxChecklistChars, len(content))
	}
}

// ---------------------------------------------------------------------------
// tick
// ---------------------------------------------------------------------------

func TestTick_CallsRunAgentOnHappyPath(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")

	agentDir := filepath.Join(agentsDir, "agent1")
	os.WriteFile(filepath.Join(agentDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	called := make(chan struct{}, 1)
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			called <- struct{}{}
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	m.tick(context.Background(), m.agents[0])

	select {
	case <-called:
		// OK
	case <-time.After(time.Second):
		t.Fatal("RunAgent was not called")
	}
}

func TestTick_RespectsTryLock(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")

	agentDir := filepath.Join(agentsDir, "agent1")
	os.WriteFile(filepath.Join(agentDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	var mu sync.Mutex
	calls := 0
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			mu.Lock()
			calls++
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	// Lock the agent heartbeat before calling tick.
	ah := m.agents[0]
	ah.mu.Lock()

	// Call tick while the lock is held — it should return immediately
	// because TryLock will fail.
	m.tick(context.Background(), ah)
	ah.mu.Unlock()

	if calls != 0 {
		t.Errorf("expected 0 RunAgent calls due to lock contention, got %d", calls)
	}
}

func TestTick_RespectsActiveHours(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	// Use a window that can never be active: 00:00-00:00
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "00:00-00:00")

	agentDir := filepath.Join(agentsDir, "agent1")
	os.WriteFile(filepath.Join(agentDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	var called bool
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			called = true
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	m.tick(context.Background(), m.agents[0])

	if called {
		t.Error("RunAgent was called despite being outside active hours")
	}
}

func TestTick_SkipsWhenNoChecklist(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")
	// No HEARTBEAT.md written.

	var called bool
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			called = true
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	m.tick(context.Background(), m.agents[0])

	if called {
		t.Error("RunAgent was called despite no HEARTBEAT.md")
	}
}

func TestTick_LogsNonOKReply(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")

	agentDir := filepath.Join(agentsDir, "agent1")
	os.WriteFile(filepath.Join(agentDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	called := make(chan struct{}, 1)
	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			called <- struct{}{}
			return "I found an issue with the disk space", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	m.tick(context.Background(), m.agents[0])

	select {
	case <-called:
		// OK
	case <-time.After(time.Second):
		t.Fatal("RunAgent was not called")
	}
}

func TestTick_PropagatesContextCancel(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "5m", "")

	agentDir := filepath.Join(agentsDir, "agent1")
	os.WriteFile(filepath.Join(agentDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Pre-cancel the context.

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			// Should not be reached because the context is already
			// cancelled and the tick returns immediately.
			return "", ctx.Err()
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	// This should not panic or hang.
	m.tick(ctx, m.agents[0])
}

// ---------------------------------------------------------------------------
// Start / Close lifecycle
// ---------------------------------------------------------------------------

func TestClose_StopsAllGoroutines(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "1h", "")

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(m.agents))
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	cancel()  // Cancel the external context.
	m.Close() // Should return without hanging.

	// Double-check that calling Close again is a no-op.
	m.Close()

	// If we got here without hanging, all goroutines stopped successfully.
}

func TestClose_BeforeStart_DoesNotHang(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "1h", "")

	m, err := New(agentsDir, &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		m.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close before Start hung")
	}
}

func TestStartClose_ConcurrentCalls(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "1h", "")

	m, err := New(agentsDir, &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			m.Start(ctx)
		}()
		go func() {
			defer wg.Done()
			m.Close()
		}()
	}
	wg.Wait()
}

func TestStart_WithMultipleAgents_StartsAndStops(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")
	writeHeartbeatConfig(t, agentsDir, "agent1", "1h", "")
	writeAgent(t, agentsDir, "agent2")
	writeHeartbeatConfig(t, agentsDir, "agent2", "1h", "")

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}

	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(m.agents))
	}

	// Start with both agents configured. We can't verify ticker firing
	// without waiting an hour, but we can verify that Start and Close
	// complete without error.
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	cancel()
	m.Close()
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestNew_WithEmptyHeartbeatEvery(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	writeAgent(t, agentsDir, "agent1")

	// Write config with heartbeat block but empty every.
	cfg := "heartbeat:\n  active_hours: 09:00-17:00\n"
	os.WriteFile(filepath.Join(agentsDir, "agent1", "config.yaml"), []byte(cfg), 0o644)

	deps := &Deps{
		RunAgent: func(ctx context.Context, agent, prompt string) (string, error) {
			return "HEARTBEAT_OK", nil
		},
	}
	m, err := New(agentsDir, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.agents) != 0 {
		t.Errorf("expected 0 entries when every is empty, got %d", len(m.agents))
	}
}

func TestReadChecklist_ReturnsEmptyOnNilDir(t *testing.T) {
	// Path with nonexistent directory.
	content, err := ReadChecklist("/nonexistent/heartbeat/HEARTBEAT.md")
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		t.Errorf("expected empty for non-existent path, got %q", content)
	}
}
