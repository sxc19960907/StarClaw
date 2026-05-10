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
