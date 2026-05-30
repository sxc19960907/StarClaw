package daemon

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

func TestNewSessionCache(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)
	if sc == nil {
		t.Fatal("NewSessionCache returned nil")
	}
	if sc.starclawDir != dir {
		t.Errorf("expected starclawDir %q, got %q", dir, sc.starclawDir)
	}
}

func TestNewSessionCache_InitialMaps(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)
	if sc.routes == nil {
		t.Error("expected non-nil routes map")
	}
	if sc.managers == nil {
		t.Error("expected non-nil managers map")
	}
}

func TestGetOrCreateManager(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	if mgr == nil {
		t.Fatal("GetOrCreateManager returned nil")
	}

	// Should have created a session.
	sess := mgr.Current()
	if sess == nil {
		t.Fatal("expected a current session after GetOrCreateManager")
	}
	if sess.ID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestGetOrCreateManager_Caches(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr1 := sc.GetOrCreateManager(sessionsDir)
	mgr2 := sc.GetOrCreateManager(sessionsDir)

	if mgr1 != mgr2 {
		t.Error("expected same manager instance for the same directory")
	}
}

func TestGetOrCreateManager_DifferentDirs(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	mgr1 := sc.GetOrCreateManager(filepath.Join(dir, "sessions", "agent1"))
	mgr2 := sc.GetOrCreateManager(filepath.Join(dir, "sessions", "agent2"))

	if mgr1 == mgr2 {
		t.Error("expected different managers for different directories")
	}
}

func TestGetOrCreate(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	mgr := sc.GetOrCreate("test-agent")
	if mgr == nil {
		t.Fatal("GetOrCreate returned nil")
	}

	sess := mgr.Current()
	if sess == nil {
		t.Fatal("expected a current session")
	}
}

func TestSessionsDir(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Default agent (empty).
	defaultDir := sc.SessionsDir("")
	expected := filepath.Join(dir, "sessions")
	if defaultDir != expected {
		t.Errorf("expected %q, got %q", expected, defaultDir)
	}

	// Named agent.
	namedDir := sc.SessionsDir("my-agent")
	expected = filepath.Join(dir, "sessions", "my-agent")
	if namedDir != expected {
		t.Errorf("expected %q, got %q", expected, namedDir)
	}
}

func TestResolveLatestSession(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	sess := mgr.NewSession()
	sess.Messages = append(sess.Messages, client.Message{Role: "user", Content: "hello"})
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	sessionID, msgs, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != nil {
		t.Fatalf("ResolveLatestSession failed: %v", err)
	}
	if sessionID != sess.ID {
		t.Errorf("expected session ID %q, got %q", sess.ID, sessionID)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Errorf("unexpected message: %+v", msgs[0])
	}
}

func TestResolveLatestSession_Empty(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	// No session exists yet — ResolveLatestSession should return empty.
	sessionID, msgs, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != nil {
		t.Fatalf("ResolveLatestSession failed: %v", err)
	}
	if sessionID != "" {
		t.Errorf("expected empty session ID, got %q", sessionID)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestResolveLatestSession_ErrRouteActive(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	// Ensure a session exists.
	mgr := sc.GetOrCreateManager(sessionsDir)
	mgr.NewSession()
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	// Lock the route (simulates an active run).
	entry := sc.LockRoute("route1")
	defer sc.UnlockRoute("route1")
	_ = entry

	// ResolveLatestSession should fail because the route is active.
	_, _, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != ErrRouteActive {
		t.Errorf("expected ErrRouteActive, got %v", err)
	}
}

func TestAppendToSession(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	sess := mgr.NewSession()
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	msgs := []client.Message{{Role: "user", Content: "test message"}}
	err := sc.AppendToSession("route1", sessionsDir, sess.ID, msgs)
	if err != nil {
		t.Fatalf("AppendToSession failed: %v", err)
	}

	// Verify messages were appended.
	_, msgs2, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != nil {
		t.Fatalf("ResolveLatestSession failed: %v", err)
	}
	if len(msgs2) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs2))
	}
	if msgs2[0].Role != "user" || msgs2[0].Content != "test message" {
		t.Errorf("unexpected message: %+v", msgs2[0])
	}
}

func TestAppendToSession_AppendMultiple(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	sess := mgr.NewSession()
	// Add an initial message.
	sess.Messages = append(sess.Messages, client.Message{Role: "user", Content: "first"})
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	msgs := []client.Message{
		{Role: "assistant", Content: "response 1"},
		{Role: "user", Content: "follow-up"},
	}
	err := sc.AppendToSession("route1", sessionsDir, sess.ID, msgs)
	if err != nil {
		t.Fatalf("AppendToSession failed: %v", err)
	}

	_, msgs2, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != nil {
		t.Fatalf("ResolveLatestSession failed: %v", err)
	}
	if len(msgs2) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs2))
	}
}

func TestAppendToSession_ErrSessionChanged(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	_ = mgr.NewSession()
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	// Use a wrong session ID.
	msgs := []client.Message{{Role: "user", Content: "test"}}
	err := sc.AppendToSession("route1", sessionsDir, "wrong-session-id", msgs)
	if err != ErrSessionChanged {
		t.Errorf("expected ErrSessionChanged, got %v", err)
	}
}

func TestAppendToSession_ErrRouteActive(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	sessionsDir := filepath.Join(dir, "sessions")
	mgr := sc.GetOrCreateManager(sessionsDir)
	sess := mgr.NewSession()
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	// Lock the route.
	entry := sc.LockRoute("route1")
	defer sc.UnlockRoute("route1")
	_ = entry

	// AppendToSession should fail because the route is active.
	msgs := []client.Message{{Role: "user", Content: "test"}}
	err := sc.AppendToSession("route1", sessionsDir, sess.ID, msgs)
	if err != ErrRouteActive {
		t.Errorf("expected ErrRouteActive, got %v", err)
	}
}

func TestAppendToSession_PathSecurity(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Try to use a sessions directory outside starclawDir.
	sessionsDir := filepath.Join(dir, "..", "other")
	err := sc.AppendToSession("route1", sessionsDir, "sess-id", nil)
	if err == nil {
		t.Error("expected error for path outside starclaw directory")
	}
}

func TestLockRouteUnlockRoute_Basic(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	entry := sc.LockRoute("route1")
	if entry == nil {
		t.Fatal("LockRoute returned nil")
	}

	// Verify the route is locked (TryLock should fail).
	if entry.mu.TryLock() {
		t.Error("expected route to be locked")
	}

	sc.UnlockRoute("route1")

	// After unlock, TryLock should succeed.
	if !entry.mu.TryLock() {
		t.Error("expected route to be unlocked")
	}
	entry.mu.Unlock()
}

func TestLockRoute_CancelsExisting(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Lock the route first (goroutine A acquires the lock).
	_ = sc.LockRoute("route1")

	// Goroutine B tries to LockRoute while A holds the lock.
	// B should block and then eventually acquire the lock after A unlocks.
	acquired := make(chan struct{})
	go func() {
		e := sc.LockRoute("route1")
		close(acquired)
		// Release the lock that B just acquired.
		sc.UnlockRoute("route1")
		_ = e
	}()

	// Give goroutine B time to start and block.
	time.Sleep(50 * time.Millisecond)

	// B should be blocked (cannot acquire lock while A holds it).
	select {
	case <-acquired:
		t.Fatal("goroutine B should be blocked waiting for the lock")
	default:
	}

	// A releases the lock — B should now acquire it.
	sc.UnlockRoute("route1")

	// B should acquire the lock within a reasonable time.
	select {
	case <-acquired:
		// Success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for goroutine B to acquire the lock")
	}
}

func TestLockRoute_MultipleKeys(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	entry1 := sc.LockRoute("route1")
	entry2 := sc.LockRoute("route2")

	// Both should be locked simultaneously since they're different keys.
	if entry1 == entry2 {
		t.Error("expected different entries for different keys")
	}

	sc.UnlockRoute("route1")
	sc.UnlockRoute("route2")
}

func TestLockRoute_Reuse(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Lock and unlock the same route multiple times.
	for i := 0; i < 5; i++ {
		entry := sc.LockRoute("reuse-route")
		sc.UnlockRoute("reuse-route")
		if entry == nil {
			t.Fatalf("iteration %d: LockRoute returned nil", i)
		}
	}
}

func TestLockRoute_ConcurrentSafety(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			entry := sc.LockRoute("concurrent-route")
			// Simulate some work.
			time.Sleep(10 * time.Millisecond)
			sc.UnlockRoute("concurrent-route")
			_ = entry
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed without deadlock.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out — possible deadlock")
	}
}

func TestResolveLatestSession_PathSecurity(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Try to use a sessions directory outside starclawDir.
	sessionsDir := filepath.Join(dir, "..", "outside")
	_, _, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err == nil {
		t.Error("expected error for path outside starclaw directory")
	}
}

func TestResolveLatestSession_InSubDir(t *testing.T) {
	dir := t.TempDir()
	sc := NewSessionCache(dir)

	// Sessions directory within starclawDir should work.
	sessionsDir := filepath.Join(dir, "sessions", "nested")
	mgr := sc.GetOrCreateManager(sessionsDir)
	s := mgr.NewSession()
	s.Messages = append(s.Messages, client.Message{Role: "user", Content: "nested test"})
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	sessionID, msgs, err := sc.ResolveLatestSession("route1", sessionsDir)
	if err != nil {
		t.Fatalf("ResolveLatestSession failed: %v", err)
	}
	if sessionID != s.ID {
		t.Errorf("expected session ID %q, got %q", s.ID, sessionID)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestNewManager(t *testing.T) {
	dir := t.TempDir()

	mgr := newManager(filepath.Join(dir, "sessions"))
	if mgr == nil {
		t.Fatal("newManager returned nil")
	}

	// Should have created a session.
	sess := mgr.Current()
	if sess == nil {
		t.Fatal("expected a current session")
	}
}

func TestNewManager_ResumesExisting(t *testing.T) {
	dir := t.TempDir()
	sessionsDir := filepath.Join(dir, "sessions")

	// Create a manager and session.
	mgr1 := newManager(sessionsDir)
	sess1 := mgr1.Current()
	sess1.Messages = append(sess1.Messages, client.Message{Role: "user", Content: "existing"})
	if err := mgr1.Save(); err != nil {
		t.Fatal(err)
	}

	// Create another manager for the same directory — should resume.
	mgr2 := newManager(sessionsDir)
	sess2 := mgr2.Current()
	if sess2 == nil {
		t.Fatal("expected resumed session")
	}
	if len(sess2.Messages) != 1 {
		t.Errorf("expected 1 message from existing session, got %d", len(sess2.Messages))
	}
}

func TestEnsureSubDir(t *testing.T) {
	dir := t.TempDir()

	// Valid subdirectory.
	if err := ensureSubDir(dir, filepath.Join(dir, "sessions")); err != nil {
		t.Errorf("expected no error for subdirectory: %v", err)
	}

	// Same directory.
	if err := ensureSubDir(dir, dir); err != nil {
		t.Errorf("expected no error for same directory: %v", err)
	}

	// Outside directory.
	if err := ensureSubDir(dir, filepath.Join(dir, "..", "outside")); err == nil {
		t.Error("expected error for outside directory")
	}
}
