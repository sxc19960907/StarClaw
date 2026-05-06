package tools

import (
	"context"
	"testing"
)

func TestSessionSearchTool_Info(t *testing.T) {
	tool := &SessionSearchTool{}
	info := tool.Info()
	if info.Name != "session_search" {
		t.Errorf("Name = %q, want 'session_search'", info.Name)
	}
	if len(info.Required) != 1 || info.Required[0] != "query" {
		t.Error("Required should be ['query']")
	}
}

func TestSessionSearchTool_EmptyQuery(t *testing.T) {
	tool := &SessionSearchTool{}
	result, _ := tool.Run(context.Background(), `{"query":""}`)
	if !result.IsError {
		t.Error("Empty query should return error")
	}
}

func TestSessionSearchTool_NoManager(t *testing.T) {
	tool := &SessionSearchTool{} // no manager
	result, _ := tool.Run(context.Background(), `{"query":"test"}`)
	if !result.IsError {
		t.Error("No manager should return error")
	}
	if result.Content != "session_search: no session manager available" {
		t.Errorf("Unexpected message: %s", result.Content)
	}
}

func TestSessionSearchTool_RequiresApproval(t *testing.T) {
	tool := &SessionSearchTool{}
	if tool.RequiresApproval() {
		t.Error("session_search should not require approval")
	}
}
