package agent

import (
	"crypto/sha256"
	"encoding/hex"
)

// ApprovalCache tracks tool calls that the user has already approved during the
// current turn. It is scoped per Run() invocation and resets each turn.
//
// The cache key is "toolName:" + SHA-256(argsJSON)[0:16], so:
//   - Same tool + same args = auto-approve (don't re-ask)
//   - Same tool + different args = ask again
//   - Different tool + same args = ask again
type ApprovalCache struct {
	approved map[string]bool
}

// NewApprovalCache creates an empty cache.
func NewApprovalCache() *ApprovalCache {
	return &ApprovalCache{approved: make(map[string]bool)}
}

// WasApproved returns true if this exact tool+args combination was previously approved.
func (c *ApprovalCache) WasApproved(toolName, argsJSON string) bool {
	return c.approved[approvalKey(toolName, argsJSON)]
}

// RecordApproval marks a tool+args combination as approved for the remainder of this turn.
func (c *ApprovalCache) RecordApproval(toolName, argsJSON string) {
	c.approved[approvalKey(toolName, argsJSON)] = true
}

func approvalKey(toolName, argsJSON string) string {
	h := sha256.Sum256([]byte(argsJSON))
	return toolName + ":" + hex.EncodeToString(h[:8])
}
