package tools

import (
	"context"
	"os"
	"path/filepath"
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
	home := t.TempDir()
	t.Setenv("HOME", home)

	skillDir := filepath.Join(home, ".starclaw", "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: test-skill
description: Test skill for load cycle
---

Use this skill in tests.
`), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewSkillTool()
	ctx := context.Background()

	listResult, err := tool.Run(ctx, `{"action":"list"}`)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if listResult.IsError {
		t.Fatalf("list returned error: %s", listResult.Content)
	}
	if !strings.Contains(listResult.Content, "test-skill") {
		t.Fatalf("list should include test-skill, got: %s", listResult.Content)
	}

	loadResult, err := tool.Run(ctx, `{"action":"load","name":"test-skill"}`)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loadResult.IsError {
		t.Fatalf("load returned error: %s", loadResult.Content)
	}
	if !strings.Contains(loadResult.Content, "Use this skill in tests") {
		t.Fatalf("load should return skill prompt, got: %s", loadResult.Content)
	}

	listLoaded, err := tool.Run(ctx, `{"action":"list"}`)
	if err != nil {
		t.Fatalf("list loaded failed: %v", err)
	}
	if !strings.Contains(listLoaded.Content, "test-skill [loaded]") {
		t.Fatalf("list should mark test-skill loaded, got: %s", listLoaded.Content)
	}

	unloadResult, err := tool.Run(ctx, `{"action":"unload","name":"test-skill"}`)
	if err != nil {
		t.Fatalf("unload failed: %v", err)
	}
	if unloadResult.IsError {
		t.Fatalf("unload returned error: %s", unloadResult.Content)
	}
	if !strings.Contains(unloadResult.Content, `Skill "test-skill" unloaded`) {
		t.Fatalf("unexpected unload result: %s", unloadResult.Content)
	}
}
