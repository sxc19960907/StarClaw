package mcp

import (
	"context"
	"fmt"
	"time"
)

// ReadinessChecker checks whether MCP servers are ready for use.
// It wraps a ClientManager to determine readiness by checking
// the connection state of each MCP server.
type ReadinessChecker struct {
	manager *ClientManager
}

// NewReadinessChecker creates a new ReadinessChecker for the given ClientManager.
func NewReadinessChecker(manager *ClientManager) *ReadinessChecker {
	return &ReadinessChecker{manager: manager}
}

// IsReady returns true if the named MCP server has an active connection.
func (r *ReadinessChecker) IsReady(serverName string) bool {
	return r.manager.IsConnected(serverName)
}

// WaitForReady blocks until the named MCP server is connected or the context is cancelled.
// It polls the connection state at 100ms intervals.
func (r *ReadinessChecker) WaitForReady(ctx context.Context, serverName string) error {
	if r.IsReady(serverName) {
		return nil
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if r.IsReady(serverName) {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("MCP server %q not ready within timeout: %w", serverName, ctx.Err())
		}
	}
}
