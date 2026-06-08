package tools

import (
	"context"
	"sync"
)

type browserUseTracker struct {
	mu     sync.Mutex
	count  int
	owners map[*BrowserTool]int
}

var globalBrowserTracker = &browserUseTracker{}

type BrowserUseLease struct {
	tracker  *browserUseTracker
	acquired bool
	released bool
	owner    *BrowserTool
}

type browserLeaseKey struct{}

func NewBrowserUseLease() *BrowserUseLease {
	return newBrowserUseLeaseWithTracker(globalBrowserTracker)
}

func newBrowserUseLeaseWithTracker(tracker *browserUseTracker) *BrowserUseLease {
	return &BrowserUseLease{tracker: tracker}
}

func WithBrowserUseLease(ctx context.Context) context.Context {
	return context.WithValue(ctx, browserLeaseKey{}, NewBrowserUseLease())
}

func BrowserUseLeaseFrom(ctx context.Context) *BrowserUseLease {
	if ctx == nil {
		return nil
	}
	lease, _ := ctx.Value(browserLeaseKey{}).(*BrowserUseLease)
	return lease
}

func MarkBrowserUsed(ctx context.Context, owner *BrowserTool) {
	BrowserUseLeaseFrom(ctx).MarkUsedWith(owner)
}

func (l *BrowserUseLease) MarkUsedWith(owner *BrowserTool) {
	if l == nil || l.tracker == nil {
		return
	}
	l.tracker.mu.Lock()
	defer l.tracker.mu.Unlock()
	if l.acquired || l.released {
		return
	}
	l.tracker.count++
	if owner != nil {
		if l.tracker.owners == nil {
			l.tracker.owners = make(map[*BrowserTool]int)
		}
		l.tracker.owners[owner]++
	}
	l.acquired = true
	l.owner = owner
}

func (l *BrowserUseLease) ReleaseOnly() {
	_, _ = l.ReleaseAndMaybeTeardown(nil)
}

func (l *BrowserUseLease) ReleaseAndMaybeTeardown(teardown func() error) (bool, error) {
	if l == nil || l.tracker == nil {
		return false, nil
	}
	l.tracker.mu.Lock()
	defer l.tracker.mu.Unlock()
	if l.released {
		return false, nil
	}
	l.released = true
	if !l.acquired {
		return false, nil
	}
	l.tracker.count--
	owner := l.owner
	if owner != nil {
		l.tracker.owners[owner]--
		if l.tracker.owners[owner] == 0 {
			delete(l.tracker.owners, owner)
		}
		if l.tracker.owners[owner] > 0 {
			return false, nil
		}
	} else if l.tracker.count > 0 {
		return false, nil
	}
	if teardown != nil {
		return true, teardown()
	}
	return true, nil
}

func (t *browserUseTracker) activeCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.count
}

func GlobalBrowserTrackerActiveCountForTest() int {
	return globalBrowserTracker.activeCount()
}

func BrowserOwnerActiveCount(owner *BrowserTool) int {
	globalBrowserTracker.mu.Lock()
	defer globalBrowserTracker.mu.Unlock()
	if globalBrowserTracker.owners == nil {
		return 0
	}
	return globalBrowserTracker.owners[owner]
}
