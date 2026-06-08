package daemon

import (
	"context"
	"sync"
)

type runtimeHandle struct {
	cancel context.CancelFunc
	pause  *runtimePauseController
}

func (h *runtimeHandle) Cancel() {
	if h == nil {
		return
	}
	if h.pause != nil {
		h.pause.Cancel()
	}
	if h.cancel != nil {
		h.cancel()
	}
}

type runtimePauseController struct {
	mu        sync.Mutex
	resumeCh  chan struct{}
	paused    bool
	cancelled bool
}

func newRuntimePauseController() *runtimePauseController {
	return &runtimePauseController{resumeCh: make(chan struct{})}
}

func (c *runtimePauseController) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelled {
		return
	}
	c.paused = true
}

func (c *runtimePauseController) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.paused {
		return
	}
	c.paused = false
	close(c.resumeCh)
	c.resumeCh = make(chan struct{})
}

func (c *runtimePauseController) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelled {
		return
	}
	c.cancelled = true
	c.paused = false
	close(c.resumeCh)
}

func (c *runtimePauseController) Paused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}

func (c *runtimePauseController) WaitIfPaused(ctx context.Context) error {
	for {
		c.mu.Lock()
		if c.cancelled || !c.paused {
			c.mu.Unlock()
			return nil
		}
		resumeCh := c.resumeCh
		c.mu.Unlock()

		select {
		case <-resumeCh:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
