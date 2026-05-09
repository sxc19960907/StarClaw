package daemon

import (
	"sync"
	"testing"
	"time"
)

func TestNewSessionCWD(t *testing.T) {
	sc := NewSessionCWD()
	if sc == nil {
		t.Fatal("NewSessionCWD returned nil")
	}
	if sc.cwds == nil {
		t.Error("expected non-nil cwds map")
	}
}

func TestGetCWD_Default(t *testing.T) {
	sc := NewSessionCWD()
	cwd := sc.GetCWD("sess-default")
	if cwd != "" {
		t.Errorf("expected empty string for unknown session, got %q", cwd)
	}
}

func TestSetAndGetCWD(t *testing.T) {
	sc := NewSessionCWD()
	sessionID := "sess-001"
	expected := "/home/user/project"

	sc.SetCWD(sessionID, expected)
	got := sc.GetCWD(sessionID)

	if got != expected {
		t.Errorf("GetCWD(%q) = %q, want %q", sessionID, got, expected)
	}
}

func TestGetCWD_MultipleSessions(t *testing.T) {
	sc := NewSessionCWD()

	sc.SetCWD("sess-a", "/tmp/a")
	sc.SetCWD("sess-b", "/tmp/b")
	sc.SetCWD("sess-c", "/tmp/c")

	if got := sc.GetCWD("sess-a"); got != "/tmp/a" {
		t.Errorf("sess-a: got %q, want %q", got, "/tmp/a")
	}
	if got := sc.GetCWD("sess-b"); got != "/tmp/b" {
		t.Errorf("sess-b: got %q, want %q", got, "/tmp/b")
	}
	if got := sc.GetCWD("sess-c"); got != "/tmp/c" {
		t.Errorf("sess-c: got %q, want %q", got, "/tmp/c")
	}
}

func TestSetCWD_Overwrite(t *testing.T) {
	sc := NewSessionCWD()
	sessionID := "sess-overwrite"

	sc.SetCWD(sessionID, "/first/path")
	sc.SetCWD(sessionID, "/second/path")

	got := sc.GetCWD(sessionID)
	if got != "/second/path" {
		t.Errorf("expected overwritten value %q, got %q", "/second/path", got)
	}
}

func TestSetCWD_EmptyDir(t *testing.T) {
	sc := NewSessionCWD()
	sc.SetCWD("sess-empty", "")
	got := sc.GetCWD("sess-empty")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetCWD_UnknownSession(t *testing.T) {
	sc := NewSessionCWD()
	sc.SetCWD("sess-known", "/some/path")
	got := sc.GetCWD("sess-unknown")
	if got != "" {
		t.Errorf("expected empty string for unknown session, got %q", got)
	}
}

func TestSessionCWD_ConcurrentSafety(t *testing.T) {
	sc := NewSessionCWD()
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionID := "sess-concurrent"
			sc.SetCWD(sessionID, "/path")
			_ = sc.GetCWD(sessionID)
		}(i)
	}

	// Concurrent reads on different sessions.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionID := "sess-concurrent"
			_ = sc.GetCWD(sessionID)
			_ = sc.GetCWD("other-session")
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed without race.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrent operations")
	}
}

func TestSessionCWD_IsolateSessions(t *testing.T) {
	sc := NewSessionCWD()

	sessions := []string{"sess-1", "sess-2", "sess-3"}
	for i, s := range sessions {
		sc.SetCWD(s, "/path/to/dir")
		_ = sc.GetCWD(s)
		_ = i
	}

	// SetCWD on one should not affect others.
	sc.SetCWD("sess-1", "/new/path")

	if got := sc.GetCWD("sess-2"); got != "/path/to/dir" {
		t.Errorf("sess-2 changed unexpectedly: got %q", got)
	}
	if got := sc.GetCWD("sess-3"); got != "/path/to/dir" {
		t.Errorf("sess-3 changed unexpectedly: got %q", got)
	}
}

func TestSessionCWD_EmptySessionID(t *testing.T) {
	sc := NewSessionCWD()
	sc.SetCWD("", "/some/path")
	got := sc.GetCWD("")
	if got != "/some/path" {
		t.Errorf("expected '/some/path' for empty sessionID, got %q", got)
	}
}
