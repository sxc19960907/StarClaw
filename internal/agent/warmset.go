package agent

import "sync"

// WarmSet is a pre-loaded cache of commonly used tools, allowing fast
// lookup without re-reading the registry. Tools are loaded via Warm()
// calls and accessed via Get().
type WarmSet struct {
	mu       sync.RWMutex
	registry *ToolRegistry
	tools    map[string]Tool
}

// NewWarmSet creates a WarmSet backed by the given registry. Tools must
// be explicitly warmed before they are available.
func NewWarmSet(registry *ToolRegistry) *WarmSet {
	return &WarmSet{
		registry: registry,
		tools:    make(map[string]Tool),
	}
}

// Get returns a previously warmed tool by name. Returns nil when the
// tool has not been warmed.
func (ws *WarmSet) Get(name string) Tool {
	if ws == nil {
		return nil
	}
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.tools[name]
}

// Warm pre-loads the named tools from the registry into the warm set.
// Unknown tool names are silently skipped.
func (ws *WarmSet) Warm(names ...string) {
	if ws == nil || ws.registry == nil {
		return
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, name := range names {
		if t, ok := ws.registry.Get(name); ok {
			ws.tools[name] = t
		}
	}
}

// WarmAll pre-loads every tool from the registry.
func (ws *WarmSet) WarmAll() {
	if ws == nil || ws.registry == nil {
		return
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, name := range ws.registry.Names() {
		if t, ok := ws.registry.Get(name); ok {
			ws.tools[name] = t
		}
	}
}

// Contains reports whether the given tool has been warmed.
func (ws *WarmSet) Contains(name string) bool {
	if ws == nil {
		return false
	}
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	_, ok := ws.tools[name]
	return ok
}

// Len returns the number of warmed tools.
func (ws *WarmSet) Len() int {
	if ws == nil {
		return 0
	}
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.tools)
}

// Clear removes all tools from the warm set.
func (ws *WarmSet) Clear() {
	if ws == nil {
		return
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.tools = make(map[string]Tool)
}
