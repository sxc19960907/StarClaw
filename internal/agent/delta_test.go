package agent

import (
	"sync"
	"testing"
)

func TestDeltaBufferAddAndText(t *testing.T) {
	b := &DeltaBuffer{}
	b.Add("Hello, ")
	b.Add("world!")
	if got := b.Text(); got != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", got)
	}
}

func TestDeltaBufferClear(t *testing.T) {
	b := &DeltaBuffer{}
	b.Add("some text")
	if b.Len() == 0 {
		t.Error("expected non-zero length before clear")
	}
	b.Clear()
	if got := b.Text(); got != "" {
		t.Errorf("expected empty after clear, got %q", got)
	}
	if b.Len() != 0 {
		t.Errorf("expected 0 length after clear, got %d", b.Len())
	}
}

func TestDeltaBufferEmpty(t *testing.T) {
	b := &DeltaBuffer{}
	if got := b.Text(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	if b.Len() != 0 {
		t.Errorf("expected 0 length, got %d", b.Len())
	}
}

func TestDeltaBufferConcurrent(t *testing.T) {
	b := &DeltaBuffer{}
	var wg sync.WaitGroup
	n := 100

	// Concurrent writes.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Add("x")
		}()
	}
	wg.Wait()

	if b.Len() != n {
		t.Errorf("expected %d bytes, got %d", n, b.Len())
	}
}

func TestDeltaBufferMultipleAdds(t *testing.T) {
	b := &DeltaBuffer{}
	b.Add("a")
	b.Add("b")
	b.Add("c")
	if got := b.Text(); got != "abc" {
		t.Errorf("expected 'abc', got %q", got)
	}
}

func TestDeltaBufferClearAndReuse(t *testing.T) {
	b := &DeltaBuffer{}
	b.Add("old data")
	b.Clear()
	b.Add("new data")
	if got := b.Text(); got != "new data" {
		t.Errorf("expected 'new data', got %q", got)
	}
}

// testDeltaHandler is a simple DeltaHandler implementation for testing.
type testDeltaHandler struct {
	mu       sync.Mutex
	deltas   []string
}

func (h *testDeltaHandler) OnDelta(text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.deltas = append(h.deltas, text)
}

func TestDeltaHandlerInterface(t *testing.T) {
	var handler DeltaHandler = &testDeltaHandler{}
	handler.OnDelta("via interface")

	h := &testDeltaHandler{}
	h.OnDelta("first")
	h.OnDelta("second")

	h.mu.Lock()
	if len(h.deltas) != 2 {
		t.Errorf("expected 2 deltas, got %d", len(h.deltas))
	}
	if h.deltas[0] != "first" || h.deltas[1] != "second" {
		t.Errorf("unexpected deltas: %v", h.deltas)
	}
	h.mu.Unlock()
}

func TestDeltaBufferImplements(t *testing.T) {
	// This is a compile-time check that DeltaBuffer satisfies an expected interface.
	var _ interface {
		Add(string)
		Text() string
		Clear()
		Len() int
	} = &DeltaBuffer{}
}
