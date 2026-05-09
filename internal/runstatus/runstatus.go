// Package runstatus defines the run status constants and type used by the agent loop.
package runstatus

// RunStatus represents the exit status of an agent run.
type RunStatus struct {
	Code   int
	Detail string
}

const (
	// StatusOK indicates the run completed successfully.
	StatusOK = 0
	// StatusContextBloat indicates the run was terminated due to context window overflow.
	StatusContextBloat = 1
	// StatusLoopDetected indicates the run was terminated due to a detected loop.
	StatusLoopDetected = 2
)
