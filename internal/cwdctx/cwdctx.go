// Package cwdctx provides a CWDContext that stores the working directory per request.
package cwdctx

// CWDContext stores a working directory that can be associated with a request.
type CWDContext struct {
	dir string
}

// New creates a new CWDContext with an empty working directory.
func New() *CWDContext {
	return &CWDContext{}
}

// Set sets the working directory.
func (c *CWDContext) Set(dir string) {
	c.dir = dir
}

// Get returns the current working directory. Returns an empty string if not set.
func (c *CWDContext) Get() string {
	return c.dir
}
