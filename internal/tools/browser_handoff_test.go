package tools

import (
	"testing"
	"time"
)

func TestHandBrowserOffNoLeasesCleansImmediately(t *testing.T) {
	t.Parallel()
	bt := &BrowserTool{}
	calls := 0
	HandBrowserOff(bt, time.Second, func() error {
		calls++
		return bt.CleanupForHandoff()
	})
	if !bt.IsDeprecated() {
		t.Fatal("browser should be deprecated")
	}
	if calls != 1 || bt.CleanupCalledForTest() != 1 {
		t.Fatalf("cleanup calls = %d internal=%d, want 1/1", calls, bt.CleanupCalledForTest())
	}
}

func TestHandBrowserOffDefersWithActiveLease(t *testing.T) {
	t.Parallel()
	bt := &BrowserTool{}
	lease := newBrowserUseLeaseWithTracker(globalBrowserTracker)
	lease.MarkUsedWith(bt)
	defer lease.ReleaseOnly()

	calls := 0
	HandBrowserOff(bt, 25*time.Millisecond, func() error {
		calls++
		return bt.CleanupForHandoff()
	})
	if calls != 0 {
		t.Fatalf("cleanup should be deferred while lease is active, got %d", calls)
	}
	time.Sleep(75 * time.Millisecond)
	if calls != 0 {
		t.Fatalf("watchdog should not clean while lease remains active, got %d", calls)
	}
	if !bt.IsDeprecated() {
		t.Fatal("browser should be deprecated")
	}
}

func TestHandBrowserOffCleansAfterLeaseRelease(t *testing.T) {
	t.Parallel()
	bt := &BrowserTool{}
	lease := newBrowserUseLeaseWithTracker(globalBrowserTracker)
	lease.MarkUsedWith(bt)
	HandBrowserOff(bt, time.Second, func() error {
		return bt.CleanupForHandoff()
	})

	torndown, err := lease.ReleaseAndMaybeTeardown(func() error {
		return bt.CleanupForHandoff()
	})
	if err != nil {
		t.Fatalf("ReleaseAndMaybeTeardown: %v", err)
	}
	if !torndown {
		t.Fatal("expected teardown after owner lease release")
	}
	if bt.CleanupCalledForTest() != 1 {
		t.Fatalf("cleanup calls = %d, want 1", bt.CleanupCalledForTest())
	}
}

func TestHandBrowserOffNilNoOp(t *testing.T) {
	t.Parallel()
	HandBrowserOff(nil, time.Millisecond, nil)
}
