package daemon

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const structuredEventSchemaVersion = "2026-06-08"

// Event types emitted by the daemon.
const (
	EventToolCall         = "tool_call"
	EventToolResult       = "tool_result"
	EventText             = "text"
	EventApprovalNeeded   = "approval_needed"
	EventApprovalResolved = "approval_resolved"
	EventError            = "error"

	// BusHandler event types
	EventToolStatus   = "tool_status"
	EventPreamble     = "preamble"
	EventStreamDelta  = "stream_delta"
	EventUsage        = "usage"
	EventRunStatus    = "run_status"
	EventBudgetStatus = "budget_status"

	// Cloud delegation event types
	EventCloudDelegateStart    = "cloud_delegate_start"
	EventCloudDelegateProgress = "cloud_delegate_progress"
	EventCloudDelegateComplete = "cloud_delegate_complete"
)

// Event is a daemon lifecycle event pushed to SSE subscribers.
type Event struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// StructuredRunEvent is the redacted, versioned event shape persisted for
// observability and future tracing export.
type StructuredRunEvent struct {
	ID            string         `json:"id"`
	SchemaVersion string         `json:"schema_version"`
	RunID         string         `json:"run_id"`
	Type          string         `json:"type"`
	Phase         string         `json:"phase"`
	At            time.Time      `json:"at"`
	Data          map[string]any `json:"data,omitempty"`
}

func newStructuredRunEvent(runID, eventType, phase string, at time.Time, data map[string]any, seq int) StructuredRunEvent {
	return StructuredRunEvent{
		ID:            fmt.Sprintf("%s-%06d", runID, seq),
		SchemaVersion: structuredEventSchemaVersion,
		RunID:         runID,
		Type:          eventType,
		Phase:         phase,
		At:            at,
		Data:          redactEventData(eventType, data),
	}
}

func eventPhase(eventType string) string {
	switch eventType {
	case EventToolCall, EventToolResult, EventToolStatus:
		return "tool"
	case EventUsage:
		return "model"
	case EventBudgetStatus:
		return "budget"
	case EventRunStatus, EventError:
		return "error"
	case "run_started":
		return "start"
	case "routing_selected":
		return "routing"
	case "fallback_decision":
		return "fallback"
	case "control_decision":
		return "control"
	case "memory_preflight":
		return "memory"
	case "workflow_step":
		return "workflow"
	case "run_completed", "run_error":
		return "end"
	default:
		return "runtime"
	}
}

func redactEventData(eventType string, data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	redacted := make(map[string]any, len(data))
	for key, value := range data {
		switch key {
		case "args", "content", "text", "delta", "preamble", "prompt", "request", "response":
			redacted[key+"_redacted"] = true
		default:
			redacted[key] = redactScalar(value)
		}
	}
	return redacted
}

func redactScalar(value any) any {
	switch v := value.(type) {
	case string:
		if looksSensitive(v) {
			return "[REDACTED]"
		}
		return v
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, nested := range v {
			if eventPayloadKeyRedacted(key) {
				out[key+"_redacted"] = true
				continue
			}
			if looksSensitive(key) {
				out[key] = "[REDACTED]"
				continue
			}
			out[key] = redactScalar(nested)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, redactScalar(item))
		}
		return out
	default:
		return value
	}
}

func eventPayloadKeyRedacted(key string) bool {
	switch key {
	case "args", "content", "text", "delta", "preamble", "prompt", "request", "response":
		return true
	default:
		return false
	}
}

func looksSensitive(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "api_key") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password") ||
		strings.Contains(lower, "bearer ")
}

// EventBus is a simple pub/sub bus for daemon events.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]chan Event
	bufferSize  int
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan Event),
		bufferSize:  64,
	}
}

// Subscribe registers a subscriber identified by id and returns a channel
// that receives all emitted events. Caller must call Unsubscribe when done.
func (b *EventBus) Subscribe(id string) <-chan Event {
	ch := make(chan Event, b.bufferSize)
	b.mu.Lock()
	b.subscribers[id] = ch
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber. No further events will be sent to the
// subscriber's channel. The channel is not closed.
func (b *EventBus) Unsubscribe(id string) {
	b.mu.Lock()
	delete(b.subscribers, id)
	b.mu.Unlock()
}

// Publish sends an event to all subscribers. Non-blocking: if a subscriber's
// buffer is full, the event is dropped for that subscriber.
func (b *EventBus) Publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- evt:
		default:
			// subscriber too slow, drop
		}
	}
}
