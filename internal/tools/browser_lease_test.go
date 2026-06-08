package tools

import (
	"context"
	"sync"
	"testing"
)

func TestBrowserUseLeaseMarkReleaseIdempotent(t *testing.T) {
	t.Parallel()
	tr := &browserUseTracker{}
	lease := newBrowserUseLeaseWithTracker(tr)
	owner := &BrowserTool{}

	lease.MarkUsedWith(owner)
	lease.MarkUsedWith(owner)
	if got := tr.activeCount(); got != 1 {
		t.Fatalf("active count = %d, want 1", got)
	}
	if got := tr.owners[owner]; got != 1 {
		t.Fatalf("owner count = %d, want 1", got)
	}

	lease.ReleaseOnly()
	lease.ReleaseOnly()
	if got := tr.activeCount(); got != 0 {
		t.Fatalf("active count after release = %d, want 0", got)
	}
	if got := tr.owners[owner]; got != 0 {
		t.Fatalf("owner count after release = %d, want 0", got)
	}
}

func TestBrowserUseLeaseReleaseBeforeAcquireDoesNotLeak(t *testing.T) {
	t.Parallel()
	tr := &browserUseTracker{}
	lease := newBrowserUseLeaseWithTracker(tr)
	lease.ReleaseOnly()
	lease.MarkUsedWith(&BrowserTool{})
	if got := tr.activeCount(); got != 0 {
		t.Fatalf("active count = %d, want 0", got)
	}
}

func TestBrowserUseLeasePerOwnerTeardown(t *testing.T) {
	t.Parallel()
	tr := &browserUseTracker{}
	ownerA := &BrowserTool{}
	ownerB := &BrowserTool{}
	leaseA := newBrowserUseLeaseWithTracker(tr)
	leaseB := newBrowserUseLeaseWithTracker(tr)
	leaseA.MarkUsedWith(ownerA)
	leaseB.MarkUsedWith(ownerB)

	callsA := 0
	torndown, err := leaseA.ReleaseAndMaybeTeardown(func() error {
		callsA++
		return nil
	})
	if err != nil {
		t.Fatalf("ReleaseAndMaybeTeardown: %v", err)
	}
	if !torndown || callsA != 1 {
		t.Fatalf("owner A teardown = %v calls=%d, want true/1", torndown, callsA)
	}
	if got := tr.activeCount(); got != 1 {
		t.Fatalf("active count = %d, want 1", got)
	}

	callsB := 0
	torndown, err = leaseB.ReleaseAndMaybeTeardown(func() error {
		callsB++
		return nil
	})
	if err != nil {
		t.Fatalf("ReleaseAndMaybeTeardown B: %v", err)
	}
	if !torndown || callsB != 1 {
		t.Fatalf("owner B teardown = %v calls=%d, want true/1", torndown, callsB)
	}
	if got := tr.activeCount(); got != 0 {
		t.Fatalf("active count = %d, want 0", got)
	}
}

func TestBrowserUseLeaseContextHelpers(t *testing.T) {
	t.Parallel()
	if BrowserUseLeaseFrom(context.Background()) != nil {
		t.Fatal("expected nil lease on plain context")
	}
	ctx := WithBrowserUseLease(context.Background())
	lease := BrowserUseLeaseFrom(ctx)
	if lease == nil {
		t.Fatal("expected lease in context")
	}
	owner := &BrowserTool{}
	MarkBrowserUsed(ctx, owner)
	if got := BrowserOwnerActiveCount(owner); got != 1 {
		t.Fatalf("owner count = %d, want 1", got)
	}
	lease.ReleaseOnly()
	if got := BrowserOwnerActiveCount(owner); got != 0 {
		t.Fatalf("owner count after release = %d, want 0", got)
	}
}

func TestBrowserUseLeaseRaceAcquireRelease(t *testing.T) {
	const leases = 100
	tr := &browserUseTracker{}
	var wg sync.WaitGroup
	for i := 0; i < leases; i++ {
		lease := newBrowserUseLeaseWithTracker(tr)
		wg.Add(2)
		go func() {
			defer wg.Done()
			lease.MarkUsedWith(&BrowserTool{})
		}()
		go func() {
			defer wg.Done()
			lease.ReleaseOnly()
		}()
	}
	wg.Wait()
	if got := tr.activeCount(); got != 0 {
		t.Fatalf("active count leaked under race: %d", got)
	}
}
