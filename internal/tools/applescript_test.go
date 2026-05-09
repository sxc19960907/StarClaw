package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestAppleScriptTool_Info(t *testing.T) {
	tool := &AppleScriptTool{}
	info := tool.Info()
	if info.Name != "applescript" {
		t.Errorf("Name = %q, want 'applescript'", info.Name)
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	if info.Parameters == nil {
		t.Error("Parameters should not be nil")
	}
	if len(info.Required) != 1 || info.Required[0] != "script" {
		t.Error("Expected required parameter 'script'")
	}
	props, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	if _, ok := props["script"]; !ok {
		t.Error("Expected 'script' parameter")
	}
}

func TestAppleScriptTool_RequiresApproval(t *testing.T) {
	tool := &AppleScriptTool{}
	if !tool.RequiresApproval() {
		t.Error("applescript should require approval")
	}
}

func TestAppleScriptTool_InvalidArgs(t *testing.T) {
	tool := &AppleScriptTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestAppleScriptTool_EmptyScript(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AppleScriptTool{}
	result, err := tool.Run(context.Background(), `{"script":""}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for empty script")
	}
}

func TestAppleScriptTool_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS")
	}

	tool := &AppleScriptTool{}
	result, err := tool.Run(context.Background(), `{"script":"return \"hello\""}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error on non-macOS")
	}
	if !strings.Contains(result.Content, "only available on macOS") {
		t.Errorf("Expected macOS-only error, got: %s", result.Content)
	}
}

func TestAppleScriptTool_IsReadOnlyCall(t *testing.T) {
	tool := &AppleScriptTool{}
	if tool.IsReadOnlyCall(`{}`) {
		t.Error("applescript should not be read-only")
	}
}

func TestBuildOsascriptArgs(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   int // expected number of args
	}{
		{"simple script", `return "hello"`, 2},
		{"multi-line", "tell app \"Finder\"\n\tactivate\nend tell", 6},
		{"empty lines", "line1\n\n\nline2", 4},
		{"whitespace only", "  \n\t", 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildOsascriptArgs(tc.script)
			if len(args) != tc.want {
				t.Errorf("buildOsascriptArgs(%q) returned %d args, want %d", tc.script, len(args), tc.want)
			}
			// All even indices should be "-e"
			for i := 0; i < len(args); i += 2 {
				if args[i] != "-e" {
					t.Errorf("Expected -e flag at index %d, got %q", i, args[i])
				}
			}
		})
	}
}
