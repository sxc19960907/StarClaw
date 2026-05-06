package daemon

import (
	"sync"
)

// Event types emitted by the daemon.
const (
	EventToolCall         = "tool_call"
	EventToolResult       = "tool_result"
	EventText             = "text"
	EventApprovalNeeded   = "approval_needed"
	EventApprovalResolved = "approval_resolved"
	EventError            = "error"
)

// Event is a daemon lifecycle event pushed to SSE subscribers.
type Event struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// EventBus is a simple pub/sub bus for daemon events.
type EventBus struct {
	mu         sync.RWMutex
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
