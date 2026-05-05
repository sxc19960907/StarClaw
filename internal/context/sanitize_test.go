package context

import (
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func TestSanitizeHistory_ToolRoleDropped(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: "hello"},
		{Role: "tool", Content: "legacy tool message"},
		{Role: "assistant", Content: "response"},
	}
	result := SanitizeHistory(msgs)
	for _, m := range result {
		if m.Role == "tool" {
			t.Error("tool role messages should be dropped")
		}
	}
}

func TestSanitizeHistory_EmptyMessageDropped(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: ""},
		{Role: "user", Content: "next"},
	}
	result := SanitizeHistory(msgs)
	for _, m := range result {
		if m.Role == "assistant" && m.Content == "" {
			t.Error("empty messages should be dropped")
		}
	}
}

func TestSanitizeHistory_ConsecutiveRoles(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "first"},
		{Role: "assistant", Content: "second"}, // should replace first
		{Role: "user", Content: "next"},
	}
	result := SanitizeHistory(msgs)
	// Count assistant messages
	assistantCount := 0
	for _, m := range result {
		if m.Role == "assistant" {
			assistantCount++
		}
	}
	if assistantCount != 1 {
		t.Errorf("Expected 1 assistant after merging, got %d", assistantCount)
	}
}

func TestSanitizeHistory_StripOrphanedTools(t *testing.T) {
	// Pure tool_use JSON without text prefix — orphaned, message should be dropped
	msgs := []client.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"abc123","name":"file_read","input":{"path":"test.go"}}`},
		{Role: "user", Content: "orphaned response without tool_result"},
	}
	result := SanitizeHistory(msgs)
	// The assistant message containing ONLY orphaned tool_use should be dropped
	for _, m := range result {
		if m.Role == "assistant" && strings.Contains(m.Content, `"type":"tool_use"`) {
			t.Error("Message with only orphaned tool_use should be dropped entirely")
		}
	}
}

func TestSanitizeHistory_ValidToolPair(t *testing.T) {
	// Valid tool_use + tool_result pair should be preserved
	msgs := []client.Message{
		{Role: "system", Content: "system"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "Let me read a file\n" + `{"type":"tool_use","id":"abc123","name":"file_read","input":{"path":"test.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"abc123","content":"file contents here"}`},
	}
	result := SanitizeHistory(msgs)
	found := false
	for _, m := range result {
		if m.Role == "assistant" && strings.Contains(m.Content, `"type":"tool_use"`) {
			found = true
		}
	}
	if !found {
		t.Error("Valid tool_use + tool_result pair should be preserved")
	}
}

func TestSanitizeHistory_EmptyInput(t *testing.T) {
	result := SanitizeHistory(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}
	result = SanitizeHistory([]client.Message{})
	if len(result) != 0 {
		t.Error("empty input should return empty")
	}
}

func TestParseJSONBlocks(t *testing.T) {
	content := `{"type":"tool_use","id":"abc","name":"file_read","input":{"path":"test.go"}}
{"type":"tool_result","tool_use_id":"abc","content":"hello"}
plain text line`

	blocks := parseJSONBlocks(content)
	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}
	if blocks[0].typ != "tool_use" || blocks[0].id != "abc" {
		t.Error("First block should be tool_use")
	}
	if blocks[1].typ != "tool_result" || blocks[1].toolUseID != "abc" {
		t.Error("Second block should be tool_result")
	}
	if blocks[2].typ != "text" {
		t.Error("Third block should be plain text")
	}
}
