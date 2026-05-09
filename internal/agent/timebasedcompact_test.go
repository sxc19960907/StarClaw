package agent

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

func TestTimeBasedCompactor_Noop_RecentMessages(t *testing.T) {
	c := NewTimeBasedCompactor(10 * time.Minute)
	// Force lastCompactAt to be recent — no-op expected
	c.lastCompactAt = time.Now()

	msgs := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"old result"}`},
	}

	result := c.Compact(msgs)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if !strings.Contains(result[0].Content, "old result") {
		t.Error("recent result should not be compacted when time has not elapsed")
	}
}

func TestTimeBasedCompactor_CompactOldMessages(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	msgs := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"old result 1"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"old result 2"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"old result 3"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu4","content":"recent result 4"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu5","content":"recent result 5"}`},
	}

	result := c.Compact(msgs)

	// First 2 results (beyond defaultKeepRecent=3) should be compacted
	if !strings.Contains(result[0].Content, "omitted") {
		t.Error("old result 1 should be compacted")
	}
	if !strings.Contains(result[1].Content, "omitted") {
		t.Error("old result 2 should be compacted")
	}
	// Last 3 should be kept intact
	if !strings.Contains(result[2].Content, "old result 3") {
		t.Error("result 3 should be kept")
	}
	if !strings.Contains(result[3].Content, "recent result 4") {
		t.Error("result 4 should be kept")
	}
	if !strings.Contains(result[4].Content, "recent result 5") {
		t.Error("result 5 should be kept")
	}
}

func TestTimeBasedCompactor_EmptyMessages(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Minute)

	result := c.Compact(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}

	result = c.Compact([]client.Message{})
	if len(result) != 0 {
		t.Error("empty input should return empty")
	}
}

func TestTimeBasedCompactor_SingleMessage(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	msgs := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"only result"}`},
	}

	result := c.Compact(msgs)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if !strings.Contains(result[0].Content, "only result") {
		t.Error("single result should be kept (below keepRecent threshold)")
	}
}

func TestTimeBasedCompactor_ExactlyKeepRecent(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	msgs := make([]client.Message, defaultKeepRecent)
	for i := 0; i < defaultKeepRecent; i++ {
		id := fmt.Sprintf("tu%d", i+1)
		msgs[i] = client.Message{
			Role:    "user",
			Content: fmt.Sprintf(`{"type":"tool_result","tool_use_id":"%s","content":"result %d"}`, id, i+1),
		}
	}

	result := c.Compact(msgs)
	if len(result) != defaultKeepRecent {
		t.Fatalf("expected %d messages, got %d", defaultKeepRecent, len(result))
	}
	for i, msg := range result {
		if !strings.Contains(msg.Content, fmt.Sprintf("result %d", i+1)) {
			t.Errorf("message %d should keep its content", i)
		}
	}
}

func TestTimeBasedCompactor_NonToolMessagesPreserved(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	msgs := []client.Message{
		{Role: "user", Content: "plain text message"},
		{Role: "assistant", Content: "assistant response"},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"result 1"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"result 2"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"result 3"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu4","content":"result 4"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu5","content":"result 5"}`},
	}

	result := c.Compact(msgs)
	if len(result) != 7 {
		t.Fatalf("expected 7 messages, got %d", len(result))
	}

	// Non-tool messages should be preserved verbatim
	if result[0].Content != "plain text message" {
		t.Error("plain text should be preserved")
	}
	if result[1].Content != "assistant response" {
		t.Error("assistant response should be preserved")
	}

	// First two tool results should be compacted
	if !strings.Contains(result[2].Content, "omitted") {
		t.Error("first tool result should be compacted")
	}
	if !strings.Contains(result[3].Content, "omitted") {
		t.Error("second tool result should be compacted")
	}

	// Last three should be kept
	if !strings.Contains(result[4].Content, "result 3") {
		t.Error("third result should be kept")
	}
	if !strings.Contains(result[5].Content, "result 4") {
		t.Error("fourth result should be kept")
	}
	if !strings.Contains(result[6].Content, "result 5") {
		t.Error("fifth result should be kept")
	}
}

func TestTimeBasedCompactor_Idempotent(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	placeholder := "[tool result from 0 minutes ago omitted]"

	msgs := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"` + placeholder + `"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"result 2"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"result 3"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu4","content":"result 4"}`},
	}

	// First call: nothing to compact (all within keepRecent)
	result := c.Compact(msgs)
	if len(result) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(result))
	}
}

func TestTimeBasedCompactor_MultipleRounds(t *testing.T) {
	c := NewTimeBasedCompactor(1 * time.Nanosecond)
	c.lastCompactAt = time.Now().Add(-1 * time.Hour)

	// First round: compact old results
	msgs := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"v2"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"v3"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu4","content":"v4"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu5","content":"v5"}`},
	}

	_ = c.Compact(msgs) // first call compacts tu1, tu2

	// After compaction, lastCompactAt is now
	lastCompactAt := c.lastCompactAt

	// Second round: with new messages but no time elapsed, should be no-op
	c.lastCompactAt = time.Now()

	msgs2 := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"v2"}`},
	}

	result2 := c.Compact(msgs2)
	if !strings.Contains(result2[0].Content, "v1") {
		t.Error("no-op should preserve content when time has not elapsed")
	}

	// Third round: elapse time again
	c.lastCompactAt = lastCompactAt

	msgs3 := []client.Message{
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"v1"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu2","content":"v2"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu3","content":"v3"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu4","content":"v4"}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu5","content":"v5"}`},
	}

	result3 := c.Compact(msgs3)
	if !strings.Contains(result3[0].Content, "omitted") {
		t.Error("time elapsed should trigger compaction again")
	}
	if !strings.Contains(result3[1].Content, "omitted") {
		t.Error("time elapsed should compact second result")
	}
	if !strings.Contains(result3[2].Content, "v3") {
		t.Error("third result should be kept")
	}
}
