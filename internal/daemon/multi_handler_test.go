package daemon

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

// multiSpy implements agent.EventHandler for testing multi_handler fan-out.
type multiSpy struct {
	mu          sync.Mutex
	toolCalls   []string
	toolResults []string
	texts       []string
	preambles   []string
	deltas      []string
	usageCount  int
	memoryCount int
}

func (s *multiSpy) OnToolCall(name, args string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolCalls = append(s.toolCalls, name)
}
func (s *multiSpy) OnToolResult(name string, result agent.ToolResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolResults = append(s.toolResults, name)
}
func (s *multiSpy) OnText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.texts = append(s.texts, text)
}
func (s *multiSpy) OnUsage(usage client.Usage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.usageCount++
}
func (s *multiSpy) OnStreamDelta(delta string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deltas = append(s.deltas, delta)
}
func (s *multiSpy) OnPreamble(preamble string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preambles = append(s.preambles, preamble)
}
func (s *multiSpy) OnMemoryPreflight(result agent.MemoryPreflightResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memoryCount++
}

func (s *multiSpy) ToolCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.toolCalls)
}

// TestNewMultiHandler verifies that NewMultiHandler accepts initial handlers.
func TestNewMultiHandler(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1, h2)

	mh.OnToolCall("test", "args")

	if h1.ToolCallCount() != 1 {
		t.Errorf("h1 expected 1 call, got %d", h1.ToolCallCount())
	}
	if h2.ToolCallCount() != 1 {
		t.Errorf("h2 expected 1 call, got %d", h2.ToolCallCount())
	}
}

// TestMultiHandlerFanOut verifies that all methods fan out to all handlers.
func TestMultiHandlerFanOut(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1, h2)

	mh.OnToolCall("tool", "args")
	mh.OnToolResult("tool", agent.ToolResult{Content: "ok"})
	mh.OnText("hello")
	mh.OnPreamble("pre")
	mh.OnStreamDelta("delta")
	mh.OnUsage(client.Usage{InputTokens: 10, OutputTokens: 20})
	mh.OnMemoryPreflight(agent.MemoryPreflightResult{Attempted: true})

	for i, s := range []*multiSpy{h1, h2} {
		s.mu.Lock()
		if len(s.toolCalls) != 1 {
			t.Errorf("handler %d: expected 1 tool call, got %d", i, len(s.toolCalls))
		}
		if len(s.toolResults) != 1 {
			t.Errorf("handler %d: expected 1 tool result, got %d", i, len(s.toolResults))
		}
		if len(s.texts) != 1 {
			t.Errorf("handler %d: expected 1 text, got %d", i, len(s.texts))
		}
		if len(s.preambles) != 1 {
			t.Errorf("handler %d: expected 1 preamble, got %d", i, len(s.preambles))
		}
		if len(s.deltas) != 1 {
			t.Errorf("handler %d: expected 1 delta, got %d", i, len(s.deltas))
		}
		if s.usageCount != 1 {
			t.Errorf("handler %d: expected 1 usage, got %d", i, s.usageCount)
		}
		if s.memoryCount != 1 {
			t.Errorf("handler %d: expected 1 memory preflight, got %d", i, s.memoryCount)
		}
		s.mu.Unlock()
	}
}

// TestMultiHandlerAdd verifies that Add registers a new handler.
func TestMultiHandlerAdd(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1)

	mh.Add(h2)
	mh.OnToolCall("tool", "args")

	if h1.ToolCallCount() != 1 {
		t.Errorf("h1 expected 1 call, got %d", h1.ToolCallCount())
	}
	if h2.ToolCallCount() != 1 {
		t.Errorf("h2 expected 1 call, got %d", h2.ToolCallCount())
	}
}

// TestMultiHandlerRemove verifies that Remove unregisters a handler.
func TestMultiHandlerRemove(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1, h2)

	mh.Remove(h1)
	mh.OnToolCall("tool", "args")

	if h1.ToolCallCount() != 0 {
		t.Errorf("removed handler should receive no calls, got %d", h1.ToolCallCount())
	}
	if h2.ToolCallCount() != 1 {
		t.Errorf("remaining handler should receive 1 call, got %d", h2.ToolCallCount())
	}
}

// TestMultiHandlerEmpty verifies that an empty handler list does not panic.
func TestMultiHandlerEmpty(t *testing.T) {
	mh := NewMultiHandler()
	mh.OnToolCall("tool", "args")
	mh.OnToolResult("tool", agent.ToolResult{})
	mh.OnText("text")
	mh.OnPreamble("pre")
	mh.OnStreamDelta("delta")
	mh.OnUsage(client.Usage{})
	// No panic = success
}

// TestMultiHandlerRemoveNonExistent verifies that removing a non-existent
// handler does not panic.
func TestMultiHandlerRemoveNonExistent(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1)
	mh.Remove(h2) // h2 is not registered
	// No panic = success
}

// TestMultiHandlerThreadSafety verifies that MultiHandler is safe for concurrent use.
func TestMultiHandlerThreadSafety(t *testing.T) {
	h1 := &multiSpy{}
	h2 := &multiSpy{}
	mh := NewMultiHandler(h1, h2)

	var done atomic.Bool
	var wg sync.WaitGroup

	// Writer goroutines that add/remove handlers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !done.Load() {
				h := &multiSpy{}
				mh.Add(h)
				mh.Remove(h)
			}
		}()
	}

	// Reader goroutines that call handler methods
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !done.Load() {
				mh.OnToolCall("tool", "args")
				mh.OnText("text")
				mh.OnPreamble("pre")
			}
		}()
	}

	// Run for a bit then stop
	_ = done.Swap(true) // mark done
	wg.Wait()
	// No race = success
}
