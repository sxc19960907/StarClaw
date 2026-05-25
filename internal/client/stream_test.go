package client

import (
	"strings"
	"testing"
)

func TestParseOpenAIStream_TextOnly(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"choices":[{"delta":{"content":" world"},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"stop"}]}

data: [DONE]

`
	var deltas []string
	resp, err := ParseOpenAIStream(strings.NewReader(stream), func(delta string) {
		deltas = append(deltas, delta)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello world" {
		t.Errorf("content = %q, want %q", resp.Content, "Hello world")
	}
	if resp.StopReason != "stop" {
		t.Errorf("stop_reason = %q, want %q", resp.StopReason, "stop")
	}
	if len(deltas) != 2 {
		t.Errorf("got %d deltas, want 2", len(deltas))
	}
}

func TestParseOpenAIStream_ToolCalls(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"read_file","arguments":""}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"path\":"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"test.go\"}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`
	resp, err := ParseOpenAIStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolUses) != 1 {
		t.Fatalf("got %d tool uses, want 1", len(resp.ToolUses))
	}
	tc := resp.ToolUses[0]
	if tc.ID != "call_1" {
		t.Errorf("tool ID = %q, want %q", tc.ID, "call_1")
	}
	if tc.Name != "read_file" {
		t.Errorf("tool name = %q, want %q", tc.Name, "read_file")
	}
	if string(tc.Input) != `{"path":"test.go"}` {
		t.Errorf("tool args = %q, want %q", string(tc.Input), `{"path":"test.go"}`)
	}
}

func TestParseOpenAIStream_MultipleToolCalls(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_a","type":"function","function":{"name":"foo","arguments":"{}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":1,"id":"call_b","type":"function","function":{"name":"bar","arguments":"{}"}}]},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`
	resp, err := ParseOpenAIStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolUses) != 2 {
		t.Fatalf("got %d tool uses, want 2", len(resp.ToolUses))
	}
	if resp.ToolUses[0].Name != "foo" || resp.ToolUses[1].Name != "bar" {
		t.Errorf("tool names = [%q, %q], want [foo, bar]", resp.ToolUses[0].Name, resp.ToolUses[1].Name)
	}
}

func TestParseOpenAIStream_Usage(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"content":"hi"},"finish_reason":null}]}

data: {"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":50,"completion_tokens":10}}

data: [DONE]

`
	resp, err := ParseOpenAIStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Usage.InputTokens != 50 {
		t.Errorf("input_tokens = %d, want 50", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 10 {
		t.Errorf("output_tokens = %d, want 10", resp.Usage.OutputTokens)
	}
}

func TestParseOpenAIStream_EmptyStream(t *testing.T) {
	stream := `data: [DONE]

`
	resp, err := ParseOpenAIStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "" {
		t.Errorf("content = %q, want empty", resp.Content)
	}
}

func TestParseOpenAIStream_NilOnDelta(t *testing.T) {
	stream := `data: {"choices":[{"delta":{"content":"test"},"finish_reason":"stop"}]}

data: [DONE]

`
	resp, err := ParseOpenAIStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "test" {
		t.Errorf("content = %q, want %q", resp.Content, "test")
	}
}

func TestParseAnthropicStream_TextOnly(t *testing.T) {
	stream := `event: message_start
data: {"type":"message_start","message":{"usage":{"input_tokens":12,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":4}}

event: message_stop
data: {"type":"message_stop"}

`
	var deltas []string
	resp, err := ParseAnthropicStream(strings.NewReader(stream), func(delta string) {
		deltas = append(deltas, delta)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello world" {
		t.Fatalf("content = %q, want Hello world", resp.Content)
	}
	if resp.StopReason != "end_turn" {
		t.Fatalf("stop reason = %q, want end_turn", resp.StopReason)
	}
	if resp.Usage.InputTokens != 12 || resp.Usage.OutputTokens != 4 {
		t.Fatalf("usage = %+v, want 12/4", resp.Usage)
	}
	if got := strings.Join(deltas, ""); got != "Hello world" {
		t.Fatalf("deltas = %q, want Hello world", got)
	}
}

func TestParseAnthropicStream_ToolUse(t *testing.T) {
	stream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Checking"}}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_1","name":"file_read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"path\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"\"README.md\"}"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":8}}

`
	resp, err := ParseAnthropicStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Checking" {
		t.Fatalf("content = %q, want Checking", resp.Content)
	}
	if resp.StopReason != "tool_use" {
		t.Fatalf("stop reason = %q, want tool_use", resp.StopReason)
	}
	if len(resp.ToolUses) != 1 {
		t.Fatalf("tool uses = %d, want 1", len(resp.ToolUses))
	}
	toolUse := resp.ToolUses[0]
	if toolUse.ID != "toolu_1" || toolUse.Name != "file_read" {
		t.Fatalf("tool use = %+v, want toolu_1/file_read", toolUse)
	}
	if string(toolUse.Input) != `{"path":"README.md"}` {
		t.Fatalf("tool input = %q, want accumulated JSON chunks", string(toolUse.Input))
	}
}

func TestParseAnthropicStream_ToolUseEmptyInputStartsWithDeltas(t *testing.T) {
	stream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"file_read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"path\":\"README.md\"}"}}

`
	resp, err := ParseAnthropicStream(strings.NewReader(stream), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolUses) != 1 {
		t.Fatalf("tool uses = %d, want 1", len(resp.ToolUses))
	}
	if string(resp.ToolUses[0].Input) != `{"path":"README.md"}` {
		t.Fatalf("tool input = %q, want accumulated start input plus delta", string(resp.ToolUses[0].Input))
	}
}

func TestParseAnthropicStream_InvalidJSONReturnsPartial(t *testing.T) {
	stream := `event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"partial"}}

event: content_block_delta
data: {"type":

`
	resp, err := ParseAnthropicStream(strings.NewReader(stream), nil)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if resp.Content != "partial" {
		t.Fatalf("partial content = %q, want partial", resp.Content)
	}
}
