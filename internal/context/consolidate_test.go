package context

import (
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func TestConsolidateRedundant_EmptyInput(t *testing.T) {
	result := ConsolidateRedundant(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}

	result = ConsolidateRedundant([]client.Message{})
	if len(result) != 0 {
		t.Error("empty input should return empty")
	}
}

func TestConsolidateRedundant_SingleMessage(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "hello"},
	}
	result := ConsolidateRedundant(msgs)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0].Content != "hello" {
		t.Error("single message should be preserved")
	}
}

func TestConsolidateRedundant_FileReadSameFile(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "system prompt"},
		{Role: "user", Content: "read foo.go"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"result v1"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"result v2"}`},
	}

	result := ConsolidateRedundant(msgs)

	// The first pair (tu1) should be removed entirely
	for _, m := range result {
		if strings.Contains(m.Content, "tu1") {
			t.Error("first tool_use (tu1) should be removed")
		}
	}

	// The last pair (tu2) should be preserved
	foundV2 := false
	for _, m := range result {
		if m.Role == "user" && strings.Contains(m.Content, "result v2") {
			foundV2 = true
		}
	}
	if !foundV2 {
		t.Error("last result (v2) should be preserved")
	}

	// After dropping the first pair (2 messages), result should have 4 messages
	if len(result) != 4 {
		t.Errorf("expected 4 messages after consolidation, got %d", len(result))
	}
}

func TestConsolidateRedundant_FileReadSameFileChain(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "read same file 3 times"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"v2"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu3","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"v3"}`},
	}

	result := ConsolidateRedundant(msgs)

	// tu1 and tu2 should be removed; only tu3 (the last) should remain
	for _, m := range result {
		if strings.Contains(m.Content, "tu1") || strings.Contains(m.Content, "tu2") {
			t.Error("first two tool_use blocks should be removed")
		}
	}

	foundV3 := false
	for _, m := range result {
		if m.Role == "user" && strings.Contains(m.Content, "v3") {
			foundV3 = true
		}
	}
	if !foundV3 {
		t.Error("last result (v3) should be preserved")
	}
}

func TestConsolidateRedundant_FileReadDifferentFiles(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "read multiple files"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"foo content"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"bar.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"bar content"}`},
	}

	result := ConsolidateRedundant(msgs)

	foundFoo := false
	foundBar := false
	for _, m := range result {
		if strings.Contains(m.Content, "foo content") {
			foundFoo = true
		}
		if strings.Contains(m.Content, "bar content") {
			foundBar = true
		}
	}
	if !foundFoo || !foundBar {
		t.Error("different file reads should both be preserved")
	}
}

func TestConsolidateRedundant_GrepSamePattern(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "search foo"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"file1.go:10: foo"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"file2.go:20: foo"}`},
	}

	result := ConsolidateRedundant(msgs)

	// tu1 should be removed; tu2 should have merged content
	foundCombined := false
	for _, m := range result {
		if m.Role == "user" && strings.Contains(m.Content, "file1.go:10: foo") &&
			strings.Contains(m.Content, "file2.go:20: foo") {
			foundCombined = true
		}
	}
	if !foundCombined {
		t.Error("grep results should be merged into the last entry")
	}
}

func TestConsolidateRedundant_GrepSamePatternChain(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "search foo 3 times"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"file1.go:10: foo"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"file2.go:20: foo"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu3","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"file3.go:30: foo"}`},
	}

	result := ConsolidateRedundant(msgs)

	// All three results should be merged into tu3
	foundAll := false
	for _, m := range result {
		if m.Role == "user" &&
			strings.Contains(m.Content, "file1.go:10: foo") &&
			strings.Contains(m.Content, "file2.go:20: foo") &&
			strings.Contains(m.Content, "file3.go:30: foo") {
			foundAll = true
		}
	}
	if !foundAll {
		t.Error("all three grep results should be merged into the last entry")
	}
}

func TestConsolidateRedundant_GrepDifferentPattern(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "search multiple patterns"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"grep","input":{"pattern":"foo"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"file1.go:10: foo"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"grep","input":{"pattern":"bar"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"file2.go:20: bar"}`},
	}

	result := ConsolidateRedundant(msgs)

	// Both should be preserved
	foundFoo := false
	foundBar := false
	for _, m := range result {
		if strings.Contains(m.Content, "file1.go:10: foo") {
			foundFoo = true
		}
		if strings.Contains(m.Content, "file2.go:20: bar") {
			foundBar = true
		}
	}
	if !foundFoo || !foundBar {
		t.Error("different pattern grep results should both be preserved")
	}
}

func TestConsolidateRedundant_FileReadWithText(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "read file"},
		{Role: "assistant",
			Content: "Let me read that file.\n" +
				`{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "assistant",
			Content: "Reading again.\n" +
				`{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"v2"}`},
	}

	result := ConsolidateRedundant(msgs)

	// The second assistant message's text should be preserved (first assistant's
	// text may be lost when consecutive assistant messages are merged).
	foundSecondText := false
	for _, m := range result {
		if m.Role == "assistant" && strings.Contains(m.Content, "Reading again") {
			foundSecondText = true
		}
	}
	if !foundSecondText {
		t.Error("text from second assistant message should be preserved")
	}

	// v2 should be preserved
	foundV2 := false
	for _, m := range result {
		if strings.Contains(m.Content, "v2") {
			foundV2 = true
		}
	}
	if !foundV2 {
		t.Error("last result should be preserved")
	}
}

func TestConsolidateRedundant_MixedToolsInterleaved(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "do work"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"grep","input":{"pattern":"bar"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"grep v1"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu3","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"v3"}`},
	}

	result := ConsolidateRedundant(msgs)

	// All results should be preserved since none are consecutive (grep is interleaved)
	for _, target := range []string{"v1", "grep v1", "v3"} {
		found := false
		for _, m := range result {
			if strings.Contains(m.Content, target) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("result %q should be preserved when tools are interleaved", target)
		}
	}
}

func TestConsolidateRedundant_FileReadWithIsError(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "read file"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"result v1"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"foo.go"}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"result v2","is_error":false}`},
	}

	result := ConsolidateRedundant(msgs)

	foundV2 := false
	for _, m := range result {
		if strings.Contains(m.Content, "result v2") {
			foundV2 = true
		}
	}
	if !foundV2 {
		t.Error("last result with is_error should be preserved")
	}
}

func TestConsolidateRedundant_FileReadWithPathAndOffset(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "read file with offset"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"file_read","input":{"path":"foo.go","offset":10}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"lines 11-20"}`},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu2","name":"file_read","input":{"path":"foo.go","offset":20}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"lines 21-30"}`},
	}

	result := ConsolidateRedundant(msgs)

	// Same-path reads are consolidated regardless of offset — tu1 is removed
	found1 := false
	found2 := false
	for _, m := range result {
		if strings.Contains(m.Content, "lines 11-20") {
			found1 = true
		}
		if strings.Contains(m.Content, "lines 21-30") {
			found2 = true
		}
	}
	if found1 {
		t.Error("same-path file read with different offsets should still be consolidated (path-based)")
	}
	if !found2 {
		t.Error("last file_read result should be preserved")
	}
}

func TestConsolidateRedundant_NoToolCalls(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
		{Role: "user", Content: "again"},
	}
	result := ConsolidateRedundant(msgs)
	if len(result) != len(msgs) {
		t.Fatalf("expected %d messages, got %d", len(msgs), len(result))
	}
}
