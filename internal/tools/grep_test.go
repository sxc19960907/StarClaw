package tools

import (
	"context"
	"testing"
)

func TestGrepTool_Info(t *testing.T) {
	tool := &GrepTool{}
	info := tool.Info()

	if info.Name != "grep" {
		t.Errorf("expected name 'grep', got %q", info.Name)
	}
	if info.Description == "" {
		t.Error("expected non-empty description")
	}

	params, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected parameters.properties to be a map")
	}

	if _, hasGlob := params["glob"]; !hasGlob {
		t.Error("expected 'glob' parameter in tool info")
	}
	if _, hasPattern := params["pattern"]; !hasPattern {
		t.Error("expected 'pattern' parameter in tool info")
	}
	if _, hasPath := params["path"]; !hasPath {
		t.Error("expected 'path' parameter in tool info")
	}
}

func TestGrepTool_Run_NoMatch(t *testing.T) {
	tool := &GrepTool{}
	result, err := tool.Run(context.Background(), `{"pattern": "zzzzz_nonexistent", "path": "grep.go"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "No matches found." {
		t.Fatalf("expected 'No matches found.', got %q", result.Content)
	}
}

func TestGrepTool_Run_Match(t *testing.T) {
	tool := &GrepTool{}
	result, err := tool.Run(context.Background(), `{"pattern": "package", "path": "grep.go"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content == "No matches found." {
		t.Fatal("expected to find matches for 'package' in grep.go")
	}
}

func TestGrepTool_Run_Directory(t *testing.T) {
	tool := &GrepTool{}
	result, err := tool.Run(context.Background(), `{"pattern": "GrepTool", "path": "."}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content == "No matches found." {
		t.Fatal("expected to find GrepTool references")
	}
}

func TestGrepTool_RequiresApproval(t *testing.T) {
	tool := &GrepTool{}
	if tool.RequiresApproval() {
		t.Error("GrepTool should not require approval")
	}
}

func TestGrepTool_IsSafeArgs(t *testing.T) {
	tool := &GrepTool{}
	if !tool.IsSafeArgs(`{"pattern": "test"}`) {
		t.Error("grep tool args should be safe")
	}
}

func TestGrepTool_Run_InvalidJSON(t *testing.T) {
	tool := &GrepTool{}
	result, err := tool.Run(context.Background(), `invalid json`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGrepTool_Run_InvalidPattern(t *testing.T) {
	tool := &GrepTool{}
	result, err := tool.Run(context.Background(), `{"pattern": "[invalid"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid regex pattern")
	}
}
