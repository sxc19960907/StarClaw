package agent

import "sync"

// StateCache provides a simple thread-safe cache for agent state across turns.
type StateCache struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewStateCache creates a new StateCache.
func NewStateCache() *StateCache {
	return &StateCache{
		data: make(map[string]any),
	}
}

// Get retrieves a value from the cache by key. Returns the value and a boolean
// indicating whether the key was found.
func (sc *StateCache) Get(key string) (any, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	v, ok := sc.data[key]
	return v, ok
}

// Set stores a value in the cache under the given key.
func (sc *StateCache) Set(key string, value any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data[key] = value
}

// Clear removes all entries from the cache.
func (sc *StateCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data = make(map[string]any)
}

// Len returns the number of entries in the cache.
func (sc *StateCache) Len() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.data)
}
