package agent

import (
	"sync"
	"testing"
	"time"
)

func TestPhase_String(t *testing.T) {
	tests := []struct {
		p    Phase
		want string
	}{
		{PhaseAnalyze, "analyze"},
		{PhaseExecute, "execute"},
		{PhaseVerify, "verify"},
		{Phase(99), "phase(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("Phase.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewPhaseTracker(t *testing.T) {
	pt := NewPhaseTracker()
	if pt == nil {
		t.Fatal("NewPhaseTracker returned nil")
	}
	phase, _ := pt.Current()
	if phase != PhaseAnalyze {
		t.Errorf("initial phase should be PhaseAnalyze, got %v", phase)
	}
}

func TestPhaseTracker_Enter(t *testing.T) {
	pt := NewPhaseTracker()

	pt.Enter(PhaseExecute)
	phase, _ := pt.Current()
	if phase != PhaseExecute {
		t.Errorf("expected PhaseExecute, got %v", phase)
	}

	pt.Enter(PhaseVerify)
	phase, _ = pt.Current()
	if phase != PhaseVerify {
		t.Errorf("expected PhaseVerify, got %v", phase)
	}
}

func TestPhaseTracker_Enter_BackToAnalyze(t *testing.T) {
	pt := NewPhaseTracker()
	pt.Enter(PhaseExecute)
	pt.Enter(PhaseAnalyze)
	phase, _ := pt.Current()
	if phase != PhaseAnalyze {
		t.Errorf("expected PhaseAnalyze, got %v", phase)
	}
}

func TestPhaseTracker_Duration(t *testing.T) {
	pt := NewPhaseTracker()
	time.Sleep(5 * time.Millisecond)
	_, dur := pt.Current()
	if dur < 5*time.Millisecond {
		t.Errorf("duration should be >= 5ms, got %v", dur)
	}
}

func TestPhaseTracker_DurationResets(t *testing.T) {
	pt := NewPhaseTracker()
	time.Sleep(2 * time.Millisecond)
	pt.Enter(PhaseExecute)
	_, dur := pt.Current()
	if dur > 50*time.Millisecond {
		t.Errorf("duration should reset after Enter, got %v", dur)
	}
}

func TestPhaseTracker_Seq(t *testing.T) {
	pt := NewPhaseTracker()
	s0 := pt.Seq()
	pt.Enter(PhaseExecute)
	s1 := pt.Seq()
	if s1 <= s0 {
		t.Errorf("seq should increase after Enter, was %d now %d", s0, s1)
	}
}

func TestPhaseTracker_EnterTransient(t *testing.T) {
	pt := NewPhaseTracker()
	pt.Enter(PhaseExecute)

	restore := pt.EnterTransient(PhaseAnalyze)
	phase, _ := pt.Current()
	if phase != PhaseAnalyze {
		t.Errorf("transient should set phase to PhaseAnalyze, got %v", phase)
	}

	restore()
	phase, _ = pt.Current()
	if phase != PhaseExecute {
		t.Errorf("restore should return to PhaseExecute, got %v", phase)
	}
}

func TestPhaseTracker_EnterTransient_Idempotent(t *testing.T) {
	pt := NewPhaseTracker()
	pt.Enter(PhaseExecute)
	restore := pt.EnterTransient(PhaseAnalyze)

	// Call restore twice — second call should be no-op.
	restore()
	restore()
	phase, _ := pt.Current()
	if phase != PhaseExecute {
		t.Errorf("after idempotent restore, expected PhaseExecute, got %v", phase)
	}
}

func TestPhaseTracker_EnterTransient_Defer(t *testing.T) {
	pt := NewPhaseTracker()
	pt.Enter(PhaseExecute)

	func() {
		defer pt.EnterTransient(PhaseAnalyze)()
		phase, _ := pt.Current()
		if phase != PhaseAnalyze {
			t.Errorf("inside transient, expected PhaseAnalyze, got %v", phase)
		}
	}()

	phase, _ := pt.Current()
	if phase != PhaseExecute {
		t.Errorf("after deferred restore, expected PhaseExecute, got %v", phase)
	}
}

func TestPhaseTracker_MarkDirty(t *testing.T) {
	pt := NewPhaseTracker()
	if pt.IsDirty() {
		t.Error("should not be dirty initially")
	}
	pt.MarkDirty()
	if !pt.IsDirty() {
		t.Error("should be dirty after MarkDirty")
	}
}

func TestPhaseTracker_TakeDirty(t *testing.T) {
	pt := NewPhaseTracker()
	pt.MarkDirty()
	if !pt.TakeDirty() {
		t.Error("TakeDirty should return true when dirty")
	}
	if pt.IsDirty() {
		t.Error("should not be dirty after TakeDirty")
	}
}

func TestPhaseTracker_TakeDirty_Cleared(t *testing.T) {
	pt := NewPhaseTracker()
	if pt.TakeDirty() {
		t.Error("TakeDirty should return false when not dirty")
	}
}

func TestPhaseTracker_Concurrent(t *testing.T) {
	pt := NewPhaseTracker()
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pt.Enter(PhaseExecute)
			pt.Enter(PhaseAnalyze)
		}()
	}

	// Concurrent reads.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pt.Current()
			pt.Seq()
			pt.IsDirty()
		}()
	}

	wg.Wait()
}

func TestPhaseTracker_AssertClean(t *testing.T) {
	// This should not panic when clean.
	pt := NewPhaseTracker()
	pt.AssertClean()
}

func TestPhaseTracker_InvalidInitially(t *testing.T) {
	pt := NewPhaseTracker()
	if pt.Invalid() {
		t.Error("tracker should not be invalid initially")
	}
}

func TestPhaseTracker_DirtyReads(t *testing.T) {
	pt := NewPhaseTracker()
	if pt.IsDirty() {
		t.Error("should not start dirty")
	}
	pt.MarkDirty()
	if !pt.IsDirty() {
		t.Error("should be dirty after mark")
	}
	d := pt.TakeDirty()
	if !d {
		t.Error("TakeDirty should return true")
	}
	if pt.IsDirty() {
		t.Error("should be clean after take")
	}
}
