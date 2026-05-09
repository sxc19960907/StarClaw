package daemon

import (
	"sync"
)

// ReadTrackerCache provides an in-memory cache for tracking which files
// have been read during a session.  This avoids duplicate file reads
// and can be cleared between agent runs.
type ReadTrackerCache struct {
	mu    sync.RWMutex
	items map[string]struct{}
}

// NewReadTrackerCache creates a new ReadTrackerCache with an empty set.
func NewReadTrackerCache() *ReadTrackerCache {
	return &ReadTrackerCache{
		items: make(map[string]struct{}),
	}
}

// MarkRead records that the given file path has been read.
func (c *ReadTrackerCache) MarkRead(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[path] = struct{}{}
}

// IsRead returns true if the given file path has been marked as read.
func (c *ReadTrackerCache) IsRead(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.items[path]
	return ok
}

// Clear removes all entries from the cache.
func (c *ReadTrackerCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]struct{})
}
