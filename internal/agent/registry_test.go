package agent

import (
	"testing"
)

// MockTool is defined in loop_test.go

func TestNewToolRegistry(t *testing.T) {
	reg := NewToolRegistry()
	if reg == nil {
		t.Fatal("NewToolRegistry() returned nil")
	}
	if len(reg.tools) != 0 {
		t.Error("New registry should be empty")
	}
}

func TestToolRegistry_Register(t *testing.T) {
	reg := NewToolRegistry()
	tool := &MockTool{name: "test_tool", description: "Test tool"}

	reg.Register(tool)

	if reg.Count() != 1 {
		t.Errorf("Expected 1 tool, got %d", reg.Count())
	}

	// Registering same name should replace
	reg.Register(tool)
	if reg.Count() != 1 {
		t.Errorf("Expected still 1 tool, got %d", reg.Count())
	}
}

func TestToolRegistry_Get(t *testing.T) {
	reg := NewToolRegistry()
	tool := &MockTool{name: "test_tool", description: "Test tool"}
	reg.Register(tool)

	// Get existing tool
	got, ok := reg.Get("test_tool")
	if !ok {
		t.Error("Expected to find registered tool")
	}
	if got.Info().Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", got.Info().Name)
	}

	// Get non-existing tool
	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("Should not find non-existent tool")
	}
}

func TestToolRegistry_List(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "tool_a", description: "Tool A"})
	reg.Register(&MockTool{name: "tool_b", description: "Tool B"})
	reg.Register(&MockTool{name: "tool_c", description: "Tool C"})

	tools := reg.List()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}

	// Should be sorted
	names := []string{tools[0].Info().Name, tools[1].Info().Name, tools[2].Info().Name}
	expected := []string{"tool_a", "tool_b", "tool_c"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected tool[%d] = %s, got %s", i, expected[i], name)
		}
	}
}

func TestToolRegistry_Names(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "zebra", description: "Zebra"})
	reg.Register(&MockTool{name: "apple", description: "Apple"})

	names := reg.Names()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	// Should be sorted
	if names[0] != "apple" || names[1] != "zebra" {
		t.Errorf("Expected sorted names [apple, zebra], got %v", names)
	}
}

func TestToolRegistry_Remove(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "tool_a", description: "Tool A"})
	reg.Register(&MockTool{name: "tool_b", description: "Tool B"})

	reg.Remove("tool_a")

	if reg.Count() != 1 {
		t.Errorf("Expected 1 tool after removal, got %d", reg.Count())
	}

	_, ok := reg.Get("tool_a")
	if ok {
		t.Error("Removed tool should not be found")
	}
}

func TestToolRegistry_Clone(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "tool_a", description: "Tool A"})
	reg.Register(&MockTool{name: "tool_b", description: "Tool B"})

	clone := reg.Clone()

	// Clone should have same tools
	if clone.Count() != reg.Count() {
		t.Errorf("Clone count %d != original %d", clone.Count(), reg.Count())
	}

	// Modifying clone should not affect original
	clone.Remove("tool_a")
	if reg.Count() == clone.Count() {
		t.Error("Modifying clone should not affect original")
	}
}

func TestToolRegistry_FilterByAllow(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "tool_a", description: "Tool A"})
	reg.Register(&MockTool{name: "tool_b", description: "Tool B"})
	reg.Register(&MockTool{name: "tool_c", description: "Tool C"})

	filtered := reg.FilterByAllow([]string{"tool_a", "tool_c"})

	if filtered.Count() != 2 {
		t.Errorf("Expected 2 filtered tools, got %d", filtered.Count())
	}

	_, ok := filtered.Get("tool_a")
	if !ok {
		t.Error("Allowed tool_a should be in filtered registry")
	}

	_, ok = filtered.Get("tool_b")
	if ok {
		t.Error("Not allowed tool_b should not be in filtered registry")
	}
}

func TestToolRegistry_FilterByDeny(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "tool_a", description: "Tool A"})
	reg.Register(&MockTool{name: "tool_b", description: "Tool B"})
	reg.Register(&MockTool{name: "tool_c", description: "Tool C"})

	filtered := reg.FilterByDeny([]string{"tool_b"})

	if filtered.Count() != 2 {
		t.Errorf("Expected 2 filtered tools, got %d", filtered.Count())
	}

	_, ok := filtered.Get("tool_b")
	if ok {
		t.Error("Denied tool_b should not be in filtered registry")
	}

	_, ok = filtered.Get("tool_a")
	if !ok {
		t.Error("Non-denied tool_a should be in filtered registry")
	}
}

func TestErrorHelpers(t *testing.T) {
	// TransientError
	result := TransientError("timeout")
	if !result.IsError {
		t.Error("TransientError should set IsError")
	}
	if result.ErrorCategory != ErrCategoryTransient {
		t.Errorf("Expected category transient, got %s", result.ErrorCategory)
	}
	if !result.IsRetryable {
		t.Error("TransientError should be retryable")
	}

	// ValidationError
	result = ValidationError("invalid args")
	if result.ErrorCategory != ErrCategoryValidation {
		t.Errorf("Expected category validation, got %s", result.ErrorCategory)
	}
	if result.IsRetryable {
		t.Error("ValidationError should not be retryable")
	}

	// BusinessError
	result = BusinessError("not allowed")
	if result.ErrorCategory != ErrCategoryBusiness {
		t.Errorf("Expected category business, got %s", result.ErrorCategory)
	}

	// PermissionError
	result = PermissionError("access denied")
	if result.ErrorCategory != ErrCategoryPermission {
		t.Errorf("Expected category permission, got %s", result.ErrorCategory)
	}
}
