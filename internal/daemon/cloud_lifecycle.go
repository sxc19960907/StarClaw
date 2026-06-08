package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const defaultCloudLifecycleNote = "Cloud WebSocket lifecycle boundary is local-only; no external transport is active."

type CloudLifecycleRunner func(context.Context) error

type CloudLifecycleStatus struct {
	Running      bool       `json:"running"`
	Configured   bool       `json:"configured"`
	Enabled      bool       `json:"enabled"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	StoppedAt    *time.Time `json:"stopped_at,omitempty"`
	RestartCount int        `json:"restart_count"`
	LastError    string     `json:"last_error,omitempty"`
	Note         string     `json:"note"`
}

type CloudLifecycleController struct {
	mu           sync.Mutex
	parent       context.Context
	runner       CloudLifecycleRunner
	cancel       context.CancelFunc
	done         chan struct{}
	running      bool
	configured   bool
	enabled      bool
	startedAt    *time.Time
	stoppedAt    *time.Time
	restartCount int
	lastError    string
	note         string
}

func NewCloudLifecycleController(parent context.Context, runner CloudLifecycleRunner) *CloudLifecycleController {
	if parent == nil {
		parent = context.Background()
	}
	if runner == nil {
		runner = defaultCloudLifecycleRunner
	}
	return &CloudLifecycleController{
		parent: parent,
		runner: runner,
		note:   defaultCloudLifecycleNote,
	}
}

func defaultCloudLifecycleRunner(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *CloudLifecycleController) Start(_ context.Context) {
	if c == nil {
		return
	}
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	runCtx, cancel := context.WithCancel(c.parent)
	done := make(chan struct{})
	now := time.Now().UTC()
	c.cancel = cancel
	c.done = done
	c.running = true
	c.startedAt = &now
	c.stoppedAt = nil
	c.lastError = ""
	runner := c.runner
	c.mu.Unlock()

	go func() {
		err := runner(runCtx)
		stopped := time.Now().UTC()
		c.mu.Lock()
		if err != nil && err != context.Canceled {
			c.lastError = err.Error()
		}
		c.running = false
		c.cancel = nil
		c.done = nil
		c.stoppedAt = &stopped
		c.mu.Unlock()
		close(done)
	}()
}

func (c *CloudLifecycleController) Stop() {
	if c == nil {
		return
	}
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (c *CloudLifecycleController) Restart(ctx context.Context) {
	if c == nil {
		return
	}
	c.mu.Lock()
	cancel := c.cancel
	done := c.done
	c.cancel = nil
	c.restartCount++
	c.mu.Unlock()

	if cancel == nil {
		c.Start(ctx)
		return
	}
	cancel()
	go func() {
		if done != nil {
			<-done
		}
		c.Start(ctx)
	}()
}

func (c *CloudLifecycleController) Status() CloudLifecycleStatus {
	if c == nil {
		return CloudLifecycleStatus{Note: defaultCloudLifecycleNote}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return CloudLifecycleStatus{
		Running:      c.running,
		Configured:   c.configured,
		Enabled:      c.enabled,
		StartedAt:    cloneTimePtr(c.startedAt),
		StoppedAt:    cloneTimePtr(c.stoppedAt),
		RestartCount: c.restartCount,
		LastError:    c.lastError,
		Note:         c.note,
	}
}

func (c *CloudLifecycleController) SetLocalOnlyState(configured, enabled bool, note string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configured = configured
	c.enabled = enabled
	if note != "" {
		c.note = note
	}
}

func (c *CloudLifecycleController) WaitStopped(ctx context.Context) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	done := c.done
	running := c.running
	c.mu.Unlock()
	if !running || done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait for cloud lifecycle stop: %w", ctx.Err())
	}
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}
