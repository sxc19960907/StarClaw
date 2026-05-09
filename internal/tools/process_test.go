package tools

import (
	"context"
	"strings"
	"testing"
)

func TestProcessTool_Info(t *testing.T) {
	tool := &ProcessTool{}
	info := tool.Info()
	if info.Name != "process" {
		t.Errorf("Name = %q, want 'process'", info.Name)
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	if info.Parameters == nil {
		t.Error("Parameters should not be nil")
	}
	if len(info.Required) != 1 || info.Required[0] != "action" {
		t.Error("Expected required parameter 'action'")
	}
	props, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	for _, key := range []string{"action", "pid", "name"} {
		if _, ok := props[key]; !ok {
			t.Errorf("Expected '%s' parameter", key)
		}
	}
}

func TestProcessTool_InvalidArgs(t *testing.T) {
	tool := &ProcessTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestProcessTool_UnknownAction(t *testing.T) {
	tool := &ProcessTool{}
	result, err := tool.Run(context.Background(), `{"action":"unknown"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for unknown action")
	}
	if !strings.Contains(result.Content, "unknown action") {
		t.Errorf("Expected unknown action error, got: %s", result.Content)
	}
}

func TestProcessTool_List(t *testing.T) {
	tool := &ProcessTool{}
	result, err := tool.Run(context.Background(), `{"action":"list"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "PID") && !strings.Contains(result.Content, "USER") {
		t.Errorf("Expected ps aux output with headers, got: %s", result.Content)
	}
}

func TestProcessTool_KillNoPID(t *testing.T) {
	tool := &ProcessTool{}
	result, err := tool.Run(context.Background(), `{"action":"kill"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for kill without pid or name")
	}
	if !strings.Contains(result.Content, "pid or name is required") {
		t.Errorf("Expected pid/name required error, got: %s", result.Content)
	}
}

func TestProcessTool_KillInvalidPID(t *testing.T) {
	tool := &ProcessTool{}
	result, err := tool.Run(context.Background(), `{"action":"kill","pid":999999999}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid PID")
	}
	if !strings.Contains(result.Content, "kill error") {
		t.Errorf("Expected kill error, got: %s", result.Content)
	}
}

func TestProcessTool_RequiresApproval(t *testing.T) {
	tool := &ProcessTool{}
	if tool.RequiresApproval() {
		t.Error("process RequiresApproval should return false (checked per-call)")
	}
}

func TestProcessTool_IsSafeArgs(t *testing.T) {
	tool := &ProcessTool{}
	if !tool.IsSafeArgs(`{"action":"list"}`) {
		t.Error("list should be safe")
	}
	if tool.IsSafeArgs(`{"action":"kill","pid":123}`) {
		t.Error("kill should not be safe")
	}
	if tool.IsSafeArgs(`{"action":"kill","name":"bash"}`) {
		t.Error("kill by name should not be safe")
	}
	if tool.IsSafeArgs(`invalid`) {
		t.Error("invalid JSON should not be safe")
	}
}

func TestProcessTool_IsReadOnlyCall(t *testing.T) {
	tool := &ProcessTool{}
	if !tool.IsReadOnlyCall(`{"action":"list"}`) {
		t.Error("list should be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"kill","pid":123}`) {
		t.Error("kill should not be read-only")
	}
}
