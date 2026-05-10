package daemon

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

// TestBusHandlerOnToolCall verifies that OnToolCall emits EventToolStatus
// with "running" status.
func TestBusHandlerOnToolCall(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	h.OnToolCall("test_tool", `{"key":"val"}`)

	evt := readEvent(t, ch)
	if evt.Type != EventToolStatus {
		t.Errorf("Type = %q, want %q", evt.Type, EventToolStatus)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["tool"] != "test_tool" {
		t.Errorf("tool = %v, want test_tool", payload["tool"])
	}
	if payload["status"] != "running" {
		t.Errorf("status = %v, want running", payload["status"])
	}
}

// TestBusHandlerOnToolResult verifies that OnToolResult emits EventToolStatus
// with "completed" status and preview.
func TestBusHandlerOnToolResult(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	result := agent.ToolResult{Content: "operation succeeded", IsError: false}
	h.OnToolResult("test_tool", result)

	evt := readEvent(t, ch)
	if evt.Type != EventToolStatus {
		t.Errorf("Type = %q, want %q", evt.Type, EventToolStatus)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["tool"] != "test_tool" {
		t.Errorf("tool = %v, want test_tool", payload["tool"])
	}
	if payload["status"] != "completed" {
		t.Errorf("status = %v, want completed", payload["status"])
	}
	if payload["preview"] != "operation succeeded" {
		t.Errorf("preview = %v, want 'operation succeeded'", payload["preview"])
	}
	if payload["is_error"] != false {
		t.Errorf("is_error = %v, want false", payload["is_error"])
	}
}

// TestBusHandlerOnText verifies that OnText emits EventText.
func TestBusHandlerOnText(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	h.OnText("hello world")

	evt := readEvent(t, ch)
	if evt.Type != EventText {
		t.Errorf("Type = %q, want %q", evt.Type, EventText)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["text"] != "hello world" {
		t.Errorf("text = %v, want 'hello world'", payload["text"])
	}
}

// TestBusHandlerOnPreambleEmpty verifies that empty preamble is silently dropped.
func TestBusHandlerOnPreambleEmpty(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	h.OnPreamble("")

	select {
	case <-ch:
		t.Error("expected no event for empty preamble")
	default:
	}
}

// TestBusHandlerOnPreambleNonEmpty verifies that non-empty preamble emits EventPreamble.
func TestBusHandlerOnPreambleNonEmpty(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	h.OnPreamble("thinking...")

	evt := readEvent(t, ch)
	if evt.Type != EventPreamble {
		t.Errorf("Type = %q, want %q", evt.Type, EventPreamble)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["text"] != "thinking..." {
		t.Errorf("text = %v, want 'thinking...'", payload["text"])
	}
}

// TestBusHandlerOnStreamDelta verifies OnStreamDelta emits EventStreamDelta.
func TestBusHandlerOnStreamDelta(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	h.OnStreamDelta("partial response chunk")

	evt := readEvent(t, ch)
	if evt.Type != EventStreamDelta {
		t.Errorf("Type = %q, want %q", evt.Type, EventStreamDelta)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload["delta"] != "partial response chunk" {
		t.Errorf("delta = %v, want 'partial response chunk'", payload["delta"])
	}
}

// TestBusHandlerNilBus verifies that a nil bus does not cause a panic.
func TestBusHandlerNilBus(t *testing.T) {
	h := NewBusHandler(nil)
	h.OnToolCall("tool", "args")
	h.OnToolResult("tool", agent.ToolResult{Content: "ok"})
	h.OnText("text")
	h.OnPreamble("preamble")
	h.OnStreamDelta("delta")
	// No panic = success
}

// TestBusHandlerThreadSafety verifies that the BusHandler is safe for concurrent use.
func TestBusHandlerThreadSafety(t *testing.T) {
	bus := NewEventBus()
	_ = bus.Subscribe("reader")
	h := NewBusHandler(bus)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.OnToolCall("tool", "args")
			h.OnToolResult("tool", agent.ToolResult{Content: "ok"})
			h.OnText("text")
		}()
	}
	wg.Wait()
	// No race = success
}

// TestBusHandlerTruncate verifies that long previews are truncated on a rune boundary.
func TestBusHandlerTruncate(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test")
	h := NewBusHandler(bus)

	long := ""
	for i := 0; i < 250; i++ {
		long += "a"
	}
	result := agent.ToolResult{Content: long}
	h.OnToolResult("tool", result)

	evt := readEvent(t, ch)
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	preview, _ := payload["preview"].(string)
	if len(preview) > 200 {
		t.Errorf("preview length = %d, want <= 200", len(preview))
	}
}

// readEvent reads a single event from the channel with a timeout.
func readEvent(t *testing.T, ch <-chan Event) Event {
	t.Helper()
	select {
	case evt := <-ch:
		return evt
	default:
		t.Fatal("expected event, got none")
		return Event{}
	}
}
