// Package cwdctx provides a CWDContext that stores the working directory per request.
package cwdctx

import "sync"

// CWDContext stores a working directory that can be associated with a request.
type CWDContext struct {
	mu  sync.Mutex
	dir string
}

// New creates a new CWDContext with an empty working directory.
func New() *CWDContext {
	return &CWDContext{}
}

// Set sets the working directory.
func (c *CWDContext) Set(dir string) {
	c.mu.Lock()
	c.dir = dir
	c.mu.Unlock()
}

// Get returns the current working directory. Returns an empty string if not set.
func (c *CWDContext) Get() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dir
}
