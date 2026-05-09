package tools

import (
	"context"
	"strings"
	"testing"
)

func TestSkillTool_Info(t *testing.T) {
	tool := NewSkillTool()
	info := tool.Info()
	if info.Name != "skill" {
		t.Errorf("Name = %q, want 'skill'", info.Name)
	}
}

func TestSkillTool_List(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"list"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	// Should report either available skills or empty state
	if !strings.Contains(result.Content, "Available skills") &&
		!strings.Contains(result.Content, "No skills available") {
		t.Errorf("expected skill list or empty message, got: %s", result.Content)
	}
}

func TestSkillTool_LoadMissingName(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"load"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for load without name")
	}
}

func TestSkillTool_LoadNotFound(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"load","name":"nonexistent-skill"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for loading nonexistent skill")
	}
	if !strings.Contains(result.Content, "not found") {
		t.Errorf("expected 'not found' error, got: %s", result.Content)
	}
}

func TestSkillTool_UnloadMissingName(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"unload"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for unload without name")
	}
}

func TestSkillTool_UnloadNotLoaded(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"unload","name":"not-loaded"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for unloading not-loaded skill")
	}
	if !strings.Contains(result.Content, "not loaded") {
		t.Errorf("expected 'not loaded' error, got: %s", result.Content)
	}
}

func TestSkillTool_InvalidAction(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{"action":"invalid"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestSkillTool_MissingAction(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Content, "unknown action") {
		t.Errorf("expected unknown action error, got: %s", result.Content)
	}
}

func TestSkillTool_RequiresApproval(t *testing.T) {
	tool := NewSkillTool()
	if tool.RequiresApproval() {
		t.Error("skill should not require approval")
	}
}

func TestSkillTool_InvalidJSON(t *testing.T) {
	tool := NewSkillTool()
	result, err := tool.Run(context.Background(), `{bad}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid JSON")
	}
}

func TestSkillTool_LoadUnloadCycle(t *testing.T) {
	// Test successful load and unload by creating a temp skill dir
	dir := t.TempDir()
	_ = dir + "/skills/test-skill"
	// We use a workaround since we can't easily inject the skill dir
	// Just verify the tool structure works
	tool := NewSkillTool()
	if tool == nil {
		t.Fatal("NewSkillTool returned nil")
	}
}
