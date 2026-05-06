package tools

import (
	"context"
	"strings"
	"testing"
)

func TestVersionTool_Info(t *testing.T) {
	tool := NewVersionTool("1.0.0")
	info := tool.Info()
	if info.Name != "version" {
		t.Errorf("Name = %q, want 'version'", info.Name)
	}
}

func TestVersionTool_Run(t *testing.T) {
	tool := NewVersionTool("v2.0.0-dev")
	result, err := tool.Run(context.Background(), "{}")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Content, "StarClaw v2.0.0-dev") {
		t.Errorf("Expected version info, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Go ") {
		t.Error("Should contain Go version")
	}
}

func TestVersionTool_RequiresApproval(t *testing.T) {
	tool := NewVersionTool("1.0")
	if tool.RequiresApproval() {
		t.Error("version should not require approval")
	}
}
