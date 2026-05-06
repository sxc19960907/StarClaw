package prompt

import (
	"strings"
	"testing"
)

func TestBuild_Empty(t *testing.T) {
	parts := Build(Options{})
	if parts.System == "" {
		t.Error("System should not be empty even with no options")
	}
	if parts.VolatileContext == "" {
		t.Error("VolatileContext should not be empty")
	}
}

func TestBuild_WithTools(t *testing.T) {
	parts := Build(Options{
		BasePrompt: "You are a helpful assistant.",
		ToolNames:  []string{"file_read", "bash", "grep"},
	})
	if !strings.Contains(parts.System, "file_read, bash, grep") {
		t.Error("System should list tool names")
	}
}

func TestBuild_WithMemory(t *testing.T) {
	parts := Build(Options{
		Memory: "User prefers Go over Python.",
	})
	if !strings.Contains(parts.VolatileContext, "User prefers Go over Python.") {
		t.Error("VolatileContext should contain memory")
	}
}

func TestBuild_WithInstructions(t *testing.T) {
	parts := Build(Options{
		Instructions: "Always write tests before code.",
	})
	if !strings.Contains(parts.VolatileContext, "Always write tests before code.") {
		t.Error("VolatileContext should contain instructions")
	}
}

func TestBuild_WithMCPContext(t *testing.T) {
	parts := Build(Options{
		MCPContext: "GitHub MCP: provides PR and issue access.",
	})
	if !strings.Contains(parts.VolatileContext, "GitHub MCP") {
		t.Error("VolatileContext should contain MCP context")
	}
}

func TestBuild_WithModel(t *testing.T) {
	parts := Build(Options{
		ModelName: "deepseek-v4",
	})
	if !strings.Contains(parts.VolatileContext, "deepseek-v4") {
		t.Error("VolatileContext should contain model name")
	}
}

func TestBuild_WithContextWindow(t *testing.T) {
	parts := Build(Options{
		ContextWindow: 200000,
	})
	if !strings.Contains(parts.VolatileContext, "200000 tokens") {
		t.Error("VolatileContext should contain context window")
	}
	// 0 should omit
	parts = Build(Options{ContextWindow: 0})
	if strings.Contains(parts.VolatileContext, "Context window") {
		t.Error("ContextWindow=0 should omit context window line")
	}
}

func TestBuild_WithSkills(t *testing.T) {
	parts := Build(Options{
		SkillNames: []string{"go-refactoring", "python-testing"},
	})
	if !strings.Contains(parts.System, "go-refactoring") {
		t.Error("System should list skill names")
	}
	if !strings.Contains(parts.System, "use_skill") {
		t.Error("System should mention use_skill")
	}
}

func TestBuild_WithMemoryDir(t *testing.T) {
	parts := Build(Options{
		MemoryDir: "/home/user/.starclaw/memory",
	})
	if !strings.Contains(parts.System, "memory_append") {
		t.Error("System should mention memory_append when MemoryDir is set")
	}
}

func TestBuild_WithCWD(t *testing.T) {
	parts := Build(Options{
		CWD: "/home/user/project",
	})
	if !strings.Contains(parts.VolatileContext, "/home/user/project") {
		t.Error("VolatileContext should contain CWD")
	}
}

func TestTruncate(t *testing.T) {
	short := "hello"
	if truncate(short, 100) != short {
		t.Error("Short string should not be truncated")
	}

	long := strings.Repeat("x", 200)
	result := truncate(long, 100)
	if !strings.Contains(result, "[truncated]") {
		t.Error("Long string should be truncated")
	}
	if len([]rune(result)) < 100 {
		t.Error("Truncated string should keep maxChars characters")
	}

	// Empty
	if truncate("", 10) != "" {
		t.Error("Empty string should remain empty")
	}
}
