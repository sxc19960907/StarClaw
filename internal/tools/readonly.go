package tools

import (
	"context"
	"fmt"

	"github.com/starclaw/starclaw/internal/agent"
)

// ReadOnlyMode wraps an agent.Tool and enforces read-only access. When the
// wrapped tool implements the ReadOnlyChecker interface, ReadOnlyMode delegates
// to the inner tool for read-only operations and blocks writes. When the inner
// tool does not implement ReadOnlyChecker, ReadOnlyMode blocks all calls.
//
// Use this middleware when a session or configuration requires read-only file
// access — it prevents any tool invocation that could modify filesystem state.
type ReadOnlyMode struct {
	inner agent.Tool
}

// NewReadOnlyMode creates a ReadOnlyMode wrapper around the given tool.
// If inner is nil, calls are always blocked.
func NewReadOnlyMode(inner agent.Tool) *ReadOnlyMode {
	return &ReadOnlyMode{inner: inner}
}

// Info returns the inner tool's metadata unchanged.
func (r *ReadOnlyMode) Info() agent.ToolInfo {
	if r.inner == nil {
		return agent.ToolInfo{
			Name:        "readonly",
			Description: "Read-only enforcement wrapper",
		}
	}
	return r.inner.Info()
}

// Run executes the inner tool only if the call is read-only (or the tool
// does not implement ReadOnlyChecker, in which case it is blocked).
func (r *ReadOnlyMode) Run(ctx context.Context, args string) (agent.ToolResult, error) {
	if r.inner == nil {
		return agent.PermissionError("read-only mode: no tool configured"), nil
	}

	if checker, ok := r.inner.(agent.ReadOnlyChecker); ok {
		if checker.IsReadOnlyCall(args) {
			return r.inner.Run(ctx, args)
		}
		return agent.PermissionError(
			fmt.Sprintf("read-only mode: tool %q is not allowed for this operation", r.inner.Info().Name)), nil
	}

	// No read-only checker interface — block all calls to be safe.
	return agent.PermissionError(
		fmt.Sprintf("read-only mode: tool %q cannot determine read-only safety, blocked", r.inner.Info().Name)), nil
}

// RequiresApproval delegates to the inner tool.
func (r *ReadOnlyMode) RequiresApproval() bool {
	if r.inner == nil {
		return false
	}
	return r.inner.RequiresApproval()
}
