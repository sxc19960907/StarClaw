package agent

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWatchdog(t *testing.T) {
	w := NewWatchdog(nil)
	if w == nil {
		t.Fatal("NewWatchdog returned nil")
	}
}

func TestWatchdog_StartStop(t *testing.T) {
	var fired atomic.Bool
	w := NewWatchdog(func() {
		fired.Store(true)
	})

	w.Start(100 * time.Millisecond)
	w.Stop()

	// Wait long enough to ensure the timer would have fired.
	time.Sleep(200 * time.Millisecond)
	if fired.Load() {
		t.Error("watchdog should not have fired after Stop")
	}
}

func TestWatchdog_Timeout(t *testing.T) {
	fired := make(chan struct{}, 1)
	w := NewWatchdog(func() {
		fired <- struct{}{}
	})

	w.Start(50 * time.Millisecond)

	select {
	case <-fired:
		// Expected.
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watchdog did not fire within timeout")
	}
}

func TestWatchdog_Reset_PreventsFire(t *testing.T) {
	var fired atomic.Bool
	w := NewWatchdog(func() {
		fired.Store(true)
	})

	w.Start(200 * time.Millisecond)
	// Reset before timeout.
	time.Sleep(50 * time.Millisecond)
	w.Reset()

	// Wait past the original deadline, but not long enough for the reset timer.
	time.Sleep(120 * time.Millisecond)
	if fired.Load() {
		t.Error("watchdog should not have fired after Reset")
	}

	// Now wait for the reset timeout.
	time.Sleep(120 * time.Millisecond)
	if !fired.Load() {
		t.Error("watchdog should have fired after reset timeout elapsed")
	}
}

func TestWatchdog_MultipleStart(t *testing.T) {
	var count atomic.Int32
	w := NewWatchdog(func() {
		count.Add(1)
	})

	w.Start(50 * time.Millisecond)
	// Second start replaces first timer.
	w.Start(200 * time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	// Only the second timer should fire, and only once.
	if count.Load() != 0 {
		t.Errorf("should not have fired yet, count=%d", count.Load())
	}

	time.Sleep(200 * time.Millisecond)
	if count.Load() != 1 {
		t.Errorf("expected exactly 1 fire, got %d", count.Load())
	}
}

func TestWatchdog_Stop_Idempotent(t *testing.T) {
	w := NewWatchdog(nil)
	w.Start(100 * time.Millisecond)
	w.Stop()
	w.Stop() // second stop should not panic
}

func TestWatchdog_Reset_AfterStop(t *testing.T) {
	var fired atomic.Bool
	w := NewWatchdog(func() {
		fired.Store(true)
	})

	w.Start(50 * time.Millisecond)
	w.Stop()
	w.Reset() // should be no-op

	time.Sleep(100 * time.Millisecond)
	if fired.Load() {
		t.Error("watchdog should not fire after Stop+Reset")
	}
}

func TestWatchdog_NilCallback(t *testing.T) {
	w := NewWatchdog(nil)
	w.Start(50 * time.Millisecond)
	// Should not panic when nil callback fires.
	time.Sleep(100 * time.Millisecond)
}

func TestWatchdog_Reset_RestartsTimer(t *testing.T) {
	var fired atomic.Bool
	w := NewWatchdog(func() {
		fired.Store(true)
	})

	w.Start(150 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	w.Reset() // resets to full 150ms again
	time.Sleep(100 * time.Millisecond)

	if fired.Load() {
		t.Error("should not have fired after reset extended the window")
	}

	time.Sleep(100 * time.Millisecond)
	if !fired.Load() {
		t.Error("should have fired after reset timeout elapsed")
	}
}

func TestWatchdog_MultipleResets(t *testing.T) {
	var fired atomic.Bool
	w := NewWatchdog(func() {
		fired.Store(true)
	})

	w.Start(200 * time.Millisecond)
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		w.Reset()
	}

	// After 150ms of resets, we should be well within the 200ms window.
	if fired.Load() {
		t.Error("should not have fired during resets")
	}

	// Now let it expire.
	time.Sleep(250 * time.Millisecond)
	if !fired.Load() {
		t.Error("should have fired after last reset expired")
	}
}

func TestWatchdog_ZeroTimeout(t *testing.T) {
	fired := make(chan struct{}, 1)
	w := NewWatchdog(func() {
		fired <- struct{}{}
	})

	w.Start(0)

	select {
	case <-fired:
		// Zero-duration timer fires immediately.
	case <-time.After(50 * time.Millisecond):
		t.Fatal("zero timeout should fire immediately")
	}
}
