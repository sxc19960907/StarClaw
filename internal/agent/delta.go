package agent

import (
	"strings"
	"sync"
)

// DeltaHandler is notified of streaming text deltas as they arrive.
type DeltaHandler interface {
	OnDelta(text string)
}

// DeltaBuffer accumulates streaming text deltas and provides thread-safe
// access to the accumulated text. Safe for concurrent use.
type DeltaBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

// Add appends a delta string to the buffer.
func (b *DeltaBuffer) Add(delta string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.WriteString(delta)
}

// Text returns the full accumulated text.
func (b *DeltaBuffer) Text() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// Clear resets the buffer, discarding all accumulated text.
func (b *DeltaBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}

// Len returns the number of bytes accumulated.
func (b *DeltaBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}
