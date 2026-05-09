package agent

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// Phase represents a discrete agent phase during a task.
type Phase int

const (
	PhaseAnalyze Phase = iota
	PhaseExecute
	PhaseVerify
)

func (p Phase) String() string {
	switch p {
	case PhaseAnalyze:
		return "analyze"
	case PhaseExecute:
		return "execute"
	case PhaseVerify:
		return "verify"
	default:
		return fmt.Sprintf("phase(%d)", int(p))
	}
}

// phaseStrictMode forces panics on violations in production builds.
// Enable with STARCLAW_PHASE_STRICT=1 for diagnostic runs.
var phaseStrictMode = os.Getenv("STARCLAW_PHASE_STRICT") == "1"

// PhaseTracker tracks agent phase transitions. Safe for concurrent use.
//
// FAIL-CLOSED DESIGN:
//   - Enter panics if called while a transient is active (under go test or
//     strict mode), otherwise it logs a warning and marks the tracker invalid.
//   - EnterTransient returns a restore closure; if the closure is forgotten,
//     a subsequent top-level Enter panics, catching the leak.
//   - Invalid() signals to observers that the tracker data is unreliable.
type PhaseTracker struct {
	mu             sync.RWMutex
	phase          Phase
	since          time.Time
	dirty          bool
	transientDepth int
	seq            int64
	invalid        bool
}

// NewPhaseTracker creates a PhaseTracker starting in PhaseAnalyze.
func NewPhaseTracker() *PhaseTracker {
	return &PhaseTracker{phase: PhaseAnalyze, since: time.Now()}
}

// Enter transitions to a new top-level phase.
// Panics if called inside an active transient (detects forgotten restores).
func (pt *PhaseTracker) Enter(p Phase) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.transientDepth != 0 {
		pt.reportViolation(fmt.Sprintf(
			"Enter(%s) called while transient is active (depth=%d, current=%s)",
			p, pt.transientDepth, pt.phase))
	}
	pt.phase = p
	pt.since = time.Now()
	pt.seq++
}

// EnterTransient enters phase p and returns a restore closure that restores
// the previous phase. The closure is idempotent.
//
// Usage:
//
//	restore := tracker.EnterTransient(PhaseExecute)
//	defer restore()
func (pt *PhaseTracker) EnterTransient(p Phase) func() {
	pt.mu.Lock()
	prev := pt.phase
	prevSince := pt.since
	pt.phase = p
	pt.since = time.Now()
	pt.seq++
	pt.transientDepth++
	pt.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			pt.mu.Lock()
			pt.phase = prev
			pt.since = prevSince
			pt.seq++
			pt.transientDepth--
			pt.mu.Unlock()
		})
	}
}

// Current returns the current phase and how long it has been active.
func (pt *PhaseTracker) Current() (Phase, time.Duration) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.phase, time.Since(pt.since)
}

// Seq returns the current transition sequence number. Observers use this to
// detect re-entry into the same phase.
func (pt *PhaseTracker) Seq() int64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.seq
}

// Invalid reports whether the tracker has recorded a structural violation.
// Observers should disable themselves when this is true.
func (pt *PhaseTracker) Invalid() bool {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.invalid
}

// MarkDirty signals that the current phase produced durable state.
func (pt *PhaseTracker) MarkDirty() {
	pt.mu.Lock()
	pt.dirty = true
	pt.mu.Unlock()
}

// TakeDirty atomically reads and clears the dirty flag.
func (pt *PhaseTracker) TakeDirty() bool {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	d := pt.dirty
	pt.dirty = false
	return d
}

// IsDirty reads the dirty flag without clearing.
func (pt *PhaseTracker) IsDirty() bool {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.dirty
}

// AssertClean reports a violation if any transient restore was forgotten.
func (pt *PhaseTracker) AssertClean() {
	pt.mu.RLock()
	depth := pt.transientDepth
	phase := pt.phase
	pt.mu.RUnlock()
	if depth != 0 {
		pt.reportViolation(fmt.Sprintf(
			"pending transient: depth=%d, stuck_in=%s", depth, phase))
	}
}

// reportViolation is the single choke point for structural phase violations.
// It marks the tracker invalid, and panics under go test or strict mode.
func (pt *PhaseTracker) reportViolation(msg string) {
	pt.invalid = true
	if testing.Testing() || phaseStrictMode {
		panic("phaseTracker: " + msg)
	}
	fmt.Fprintf(os.Stderr, "[phase] WARN %s (tracker disabled for rest of run)\n", msg)
}
