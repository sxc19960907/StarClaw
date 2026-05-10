package context

import (
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func makeMessages(count int) []client.Message {
	msgs := []client.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello, help me with a task."},
	}
	for i := 0; i < count; i++ {
		msgs = append(msgs,
			client.Message{Role: "assistant", Content: "Response " + strings.Repeat("x", 100)},
			client.Message{Role: "user", Content: "Tool result " + strings.Repeat("y", 200)},
		)
	}
	return msgs
}

func TestMinShapeable(t *testing.T) {
	if MinShapeable() != 7 {
		t.Errorf("MinShapeable = %d, want 7", MinShapeable())
	}
}

func TestEstimateTokens(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
	}
	est := EstimateTokens(msgs)
	if est <= 0 {
		t.Errorf("EstimateTokens should be > 0, got %d", est)
	}
	t.Logf("EstimateTokens for 2 messages: %d", est)
}

func TestEstimateTokens_EmptyContent(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: ""},
	}
	est := EstimateTokens(msgs)
	// Just overhead per message
	if est != overheadPerMessage {
		t.Errorf("EstimateTokens for empty = %d, want %d", est, overheadPerMessage)
	}
}

func TestShouldCompact(t *testing.T) {
	// Well under threshold
	if ShouldCompact(1000, 0, 100000) {
		t.Error("ShouldCompact should return false for low tokens")
	}
	// Over 85% threshold
	if !ShouldCompact(90000, 0, 100000) {
		t.Error("ShouldCompact should return true when over 85%")
	}
	// Disabled
	if ShouldCompact(90000, 0, 0) {
		t.Error("ShouldCompact should return false when contextWindow is 0")
	}
}

func TestShapeHistory_NoChange(t *testing.T) {
	msgs := makeMessages(3) // 2 + 6 = 8 messages, below min shapeable
	orig := len(msgs)
	shaped := ShapeHistory(msgs, "", 100000)
	if len(shaped) != orig {
		t.Errorf("Short history should not be shaped: %d → %d", orig, len(shaped))
	}
}

func TestShapeHistory_Compacts(t *testing.T) {
	// Create a very large history
	msgs := makeMessages(50) // 2 + 100 messages
	shaped := ShapeHistory(msgs, "", 10000)
	if len(shaped) >= len(msgs) {
		t.Errorf("Large history should be compacted: %d → %d", len(msgs), len(shaped))
	}
	// Should keep first message as anchor
	if shaped[0].Role != "system" {
		t.Error("Should preserve first message (system in this test)")
	}
}

func TestShapeHistory_KeepsSystem(t *testing.T) {
	msgs := makeMessages(30)
	shaped := ShapeHistory(msgs, "", 5000)
	if len(shaped) == 0 || shaped[0].Role != "system" {
		t.Error("Shaped history must keep first message (system in this test)")
	}
}

func TestCompressOldToolResults(t *testing.T) {
	msgs := []client.Message{
		{Role: "system", Content: "system prompt"},
		{Role: "user", Content: "query"},
		{Role: "assistant", Content: "response"},
		{Role: "user", Content: strings.Repeat("tool result data ", 500)}, // old result
		{Role: "assistant", Content: "another response"},
		{Role: "user", Content: strings.Repeat("recent result ", 100)}, // recent — should NOT be truncated
	}

	CompressOldToolResults(msgs, 1, 200)

	// Old tool result at index 3 should be truncated
	if len([]rune(msgs[3].Content)) > 250 {
		t.Errorf("Old tool result should be truncated, got %d chars", len([]rune(msgs[3].Content)))
	}
	if !strings.Contains(msgs[3].Content, "[truncated]") {
		t.Error("Old tool result should have [truncated] marker")
	}

	// Recent tool result at index 5 should NOT be truncated
	if strings.Contains(msgs[5].Content, "[truncated]") {
		t.Error("Recent tool result should NOT be truncated")
	}
}

func TestCompressOldToolResults_Empty(t *testing.T) {
	// Should not panic on empty or small message lists
	CompressOldToolResults(nil, 3, 300)
	CompressOldToolResults([]client.Message{}, 3, 300)
	CompressOldToolResults([]client.Message{{Role: "user", Content: "hi"}}, 3, 300)
}

func TestShapeHistory_WithSummary(t *testing.T) {
	msgs := makeMessages(30)
	summary := "The user asked about Go error handling patterns."
	shaped := ShapeHistory(msgs, summary, 5000)

	found := false
	for _, m := range shaped {
		if strings.Contains(m.Content, "Previous context summary:") {
			found = true
			if !strings.Contains(m.Content, summary) {
				t.Error("Summary message should contain the summary text")
			}
			break
		}
	}
	if !found {
		t.Error("Should inject summary message into shaped history")
	}
}

func TestShapeHistory_EmptySummary(t *testing.T) {
	msgs := makeMessages(30)
	shaped := ShapeHistory(msgs, "", 5000)
	for _, m := range shaped {
		if strings.Contains(m.Content, "Previous context summary:") {
			t.Error("Should NOT inject summary message when summary is empty")
		}
	}
}

func TestEstimateTokens_CJK(t *testing.T) {
	msgs := []client.Message{
		{Role: "user", Content: "你好世界！这是一个中文测试。"},
	}
	est := EstimateTokens(msgs)
	if est <= 0 {
		t.Error("EstimateTokens should handle CJK")
	}
	t.Logf("CJK estimate: %d", est)
}
