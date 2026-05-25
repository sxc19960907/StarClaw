package agent

import (
	"sync"
	"time"
)

// Watchdog provides a timer that fires a callback when a timeout is exceeded.
// Supports start, reset, and stop operations. Safe for concurrent use.
type Watchdog struct {
	mu         sync.Mutex
	timer      *time.Timer
	timeout    time.Duration
	callback   func()
	stopped    bool
	generation uint64
}

// NewWatchdog creates a new watchdog with the given callback function.
// The callback is called when the watchdog timer expires.
func NewWatchdog(callback func()) *Watchdog {
	return &Watchdog{callback: callback}
}

// Start starts the watchdog timer with the given timeout duration.
// If the watchdog is already running, it is stopped and restarted.
func (w *Watchdog) Start(timeout time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timeout = timeout
	w.stopped = false
	w.generation++
	generation := w.generation
	w.timer = time.AfterFunc(timeout, func() {
		w.mu.Lock()
		if w.stopped || generation != w.generation {
			w.mu.Unlock()
			return
		}
		cb := w.callback
		w.mu.Unlock()
		if cb != nil {
			cb()
		}
	})
}

// Reset restarts the watchdog timer with the current timeout duration.
// Has no effect if the watchdog has been stopped or was never started.
func (w *Watchdog) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil && !w.stopped {
		w.generation++
		generation := w.generation
		w.timer.Stop()
		w.timer = time.AfterFunc(w.timeout, func() {
			w.mu.Lock()
			if w.stopped || generation != w.generation {
				w.mu.Unlock()
				return
			}
			cb := w.callback
			w.mu.Unlock()
			if cb != nil {
				cb()
			}
		})
	}
}

// Stop stops the watchdog timer and prevents the callback from firing.
// Safe to call multiple times.
func (w *Watchdog) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.stopped = true
	w.generation++
}
