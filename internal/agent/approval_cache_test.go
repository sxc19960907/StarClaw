package agent

import (
	"testing"
)

func TestNewApprovalCache(t *testing.T) {
	c := NewApprovalCache()
	if c == nil {
		t.Fatal("NewApprovalCache() returned nil")
	}
	if c.approved == nil {
		t.Error("approved map not initialized")
	}
}

func TestApprovalCache_WasApproved_New(t *testing.T) {
	c := NewApprovalCache()
	if c.WasApproved("bash", `{"command":"ls"}`) {
		t.Error("WasApproved should return false for new entry")
	}
}

func TestApprovalCache_RecordAndCheck(t *testing.T) {
	c := NewApprovalCache()

	c.RecordApproval("bash", `{"command":"ls"}`)
	if !c.WasApproved("bash", `{"command":"ls"}`) {
		t.Error("WasApproved should return true after RecordApproval")
	}
}

func TestApprovalCache_DifferentArgs(t *testing.T) {
	c := NewApprovalCache()

	c.RecordApproval("bash", `{"command":"ls"}`)
	if c.WasApproved("bash", `{"command":"rm -rf /"}`) {
		t.Error("Different args should NOT match")
	}
}

func TestApprovalCache_DifferentTools(t *testing.T) {
	c := NewApprovalCache()

	c.RecordApproval("bash", `{"command":"ls"}`)
	if c.WasApproved("file_read", `{"command":"ls"}`) {
		t.Error("Different tool with same args should NOT match")
	}
}

func TestApprovalCache_MultipleEntries(t *testing.T) {
	c := NewApprovalCache()

	c.RecordApproval("tool_a", `{"arg":1}`)
	c.RecordApproval("tool_b", `{"arg":2}`)

	if !c.WasApproved("tool_a", `{"arg":1}`) {
		t.Error("Should remember tool_a")
	}
	if !c.WasApproved("tool_b", `{"arg":2}`) {
		t.Error("Should remember tool_b")
	}
	if c.WasApproved("tool_c", `{"arg":3}`) {
		t.Error("Should not have tool_c")
	}
}
