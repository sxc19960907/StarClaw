package daemon

import (
	"sync"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

// MultiHandler fans out agent.EventHandler callbacks to multiple wrapped handlers.
// All base methods (OnToolCall, OnToolResult, OnText, etc.) call every registered
// handler in order. Thread-safe for concurrent access.
type MultiHandler struct {
	mu       sync.RWMutex
	handlers []agent.EventHandler
}

// NewMultiHandler creates a MultiHandler with the given initial handlers.
func NewMultiHandler(handlers ...agent.EventHandler) *MultiHandler {
	m := make([]agent.EventHandler, len(handlers))
	copy(m, handlers)
	return &MultiHandler{handlers: m}
}

// Add registers a handler. Safe for concurrent use.
func (m *MultiHandler) Add(h agent.EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, h)
}

// Remove unregisters the first occurrence of h. Safe for concurrent use.
func (m *MultiHandler) Remove(h agent.EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, handler := range m.handlers {
		if interface{}(handler) == interface{}(h) {
			m.handlers = append(m.handlers[:i], m.handlers[i+1:]...)
			return
		}
	}
}

// OnToolCall calls OnToolCall on every registered handler.
func (m *MultiHandler) OnToolCall(name, args string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnToolCall(name, args)
	}
}

// OnToolResult calls OnToolResult on every registered handler.
func (m *MultiHandler) OnToolResult(name string, result agent.ToolResult) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnToolResult(name, result)
	}
}

// OnText calls OnText on every registered handler.
func (m *MultiHandler) OnText(text string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnText(text)
	}
}

// OnUsage calls OnUsage on every registered handler.
func (m *MultiHandler) OnUsage(usage client.Usage) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnUsage(usage)
	}
}

// OnStreamDelta calls OnStreamDelta on every registered handler.
func (m *MultiHandler) OnStreamDelta(delta string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnStreamDelta(delta)
	}
}

// OnPreamble calls OnPreamble on every registered handler.
func (m *MultiHandler) OnPreamble(preamble string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, h := range m.handlers {
		h.OnPreamble(preamble)
	}
}
