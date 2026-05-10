package daemon

import (
	"encoding/json"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

const maxPreviewLen = 200

// BusHandler implements agent.EventHandler by dispatching daemon events to
// an EventBus. Thread-safe via a mutex-protected EventBus reference.
type BusHandler struct {
	mu  sync.Mutex
	bus *EventBus
}

// NewBusHandler creates a BusHandler that publishes events to the given bus.
func NewBusHandler(bus *EventBus) *BusHandler {
	return &BusHandler{bus: bus}
}

// OnToolCall emits an EventToolStatus with status "running".
func (h *BusHandler) OnToolCall(name, args string) {
	h.emit(EventToolStatus, map[string]any{
		"tool":   name,
		"status": "running",
	})
}

// OnToolResult emits an EventToolStatus with status "completed".
func (h *BusHandler) OnToolResult(name string, result agent.ToolResult) {
	h.emit(EventToolStatus, map[string]any{
		"tool":     name,
		"status":   "completed",
		"is_error": result.IsError,
		"preview":  truncateString(result.Content, maxPreviewLen),
	})
}

// OnText emits an EventText with the text content.
func (h *BusHandler) OnText(text string) {
	h.emit(EventText, map[string]any{
		"text": text,
	})
}

// OnPreamble emits an EventPreamble with the preamble text.
// Empty text is silently dropped.
func (h *BusHandler) OnPreamble(text string) {
	if text == "" {
		return
	}
	h.emit(EventPreamble, map[string]any{
		"text": text,
	})
}

// OnStreamDelta emits an EventStreamDelta with the delta content.
func (h *BusHandler) OnStreamDelta(delta string) {
	h.emit(EventStreamDelta, map[string]any{
		"delta": delta,
	})
}

// OnUsage emits an EventUsage with token usage information.
func (h *BusHandler) OnUsage(usage client.Usage) {
	h.emit(EventUsage, map[string]any{
		"input_tokens":  usage.InputTokens,
		"output_tokens": usage.OutputTokens,
	})
}

// emit marshals the payload as JSON and publishes it to the bus.
// If the bus is nil or marshalling fails, the event is silently dropped.
func (h *BusHandler) emit(eventType string, payload any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.bus == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.bus.Publish(Event{Type: eventType, Data: string(data)})
}

// nowISO returns the current wall time in RFC3339 format.
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// truncateString truncates s to at most max bytes, never slicing mid-rune.
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max]
}
