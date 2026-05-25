package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// streamState accumulates incremental deltas from an OpenAI-compatible SSE stream
// into a final Response.
type streamState struct {
	content    strings.Builder
	toolCalls  map[int]*streamToolCall // keyed by index
	usage      Usage
	stopReason string
}

type streamToolCall struct {
	ID       string
	Name     string
	ArgsJSON strings.Builder
}

type anthropicStreamState struct {
	content    strings.Builder
	toolCalls  map[int]*streamToolCall
	usage      Usage
	stopReason string
}

func newStreamState() *streamState {
	return &streamState{
		toolCalls: make(map[int]*streamToolCall),
	}
}

func newAnthropicStreamState() *anthropicStreamState {
	return &anthropicStreamState{
		toolCalls: make(map[int]*streamToolCall),
	}
}

// finalize assembles the accumulated state into a Response.
func (s *streamState) finalize() *Response {
	resp := &Response{
		Content:    s.content.String(),
		Usage:      s.usage,
		StopReason: s.stopReason,
	}
	if len(s.toolCalls) > 0 {
		resp.ToolUses = finalizeStreamToolCalls(s.toolCalls)
	}
	return resp
}

func (s *anthropicStreamState) finalize() *Response {
	resp := &Response{
		Content:    s.content.String(),
		Usage:      s.usage,
		StopReason: s.stopReason,
	}
	if len(s.toolCalls) > 0 {
		resp.ToolUses = finalizeStreamToolCalls(s.toolCalls)
	}
	return resp
}

func finalizeStreamToolCalls(toolCalls map[int]*streamToolCall) []ToolUse {
	keys := make([]int, 0, len(toolCalls))
	for key := range toolCalls {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	result := make([]ToolUse, 0, len(keys))
	for _, key := range keys {
		tc := toolCalls[key]
		if tc == nil {
			continue
		}
		result = append(result, ToolUse{
			ID:    tc.ID,
			Name:  tc.Name,
			Input: json.RawMessage(tc.ArgsJSON.String()),
		})
	}
	return result
}

// ParseOpenAIStream reads an OpenAI-compatible SSE stream from reader, calling
// onDelta for each text content chunk. Returns the fully assembled Response.
func ParseOpenAIStream(reader io.Reader, onDelta func(delta string)) (*Response, error) {
	state := newStreamState()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		if err := processStreamChunk(data, state, onDelta); err != nil {
			return state.finalize(), fmt.Errorf("stream chunk parse error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return state.finalize(), fmt.Errorf("stream read error: %w", err)
	}

	return state.finalize(), nil
}

func processStreamChunk(data string, state *streamState, onDelta func(delta string)) error {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Index    int    `json:"index"`
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"delta"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return fmt.Errorf("unmarshal chunk: %w", err)
	}

	if chunk.Usage != nil {
		state.usage.InputTokens = chunk.Usage.PromptTokens
		state.usage.OutputTokens = chunk.Usage.CompletionTokens
	}

	if len(chunk.Choices) == 0 {
		return nil
	}

	choice := chunk.Choices[0]
	if choice.FinishReason != "" {
		state.stopReason = choice.FinishReason
	}

	if choice.Delta.Content != "" {
		state.content.WriteString(choice.Delta.Content)
		if onDelta != nil {
			onDelta(choice.Delta.Content)
		}
	}

	for _, tc := range choice.Delta.ToolCalls {
		existing, ok := state.toolCalls[tc.Index]
		if !ok {
			existing = &streamToolCall{}
			state.toolCalls[tc.Index] = existing
		}
		if tc.ID != "" {
			existing.ID = tc.ID
		}
		if tc.Function.Name != "" {
			existing.Name = tc.Function.Name
		}
		if tc.Function.Arguments != "" {
			existing.ArgsJSON.WriteString(tc.Function.Arguments)
		}
	}

	return nil
}

// ParseAnthropicStream reads an Anthropic Messages SSE stream, calling onDelta
// for text deltas and returning the fully assembled final response.
func ParseAnthropicStream(reader io.Reader, onDelta func(delta string)) (*Response, error) {
	state := newAnthropicStreamState()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		if err := processAnthropicStreamEvent(data, state, onDelta); err != nil {
			return state.finalize(), fmt.Errorf("stream event parse error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return state.finalize(), fmt.Errorf("stream read error: %w", err)
	}

	return state.finalize(), nil
}

func processAnthropicStreamEvent(data string, state *anthropicStreamState, onDelta func(delta string)) error {
	var event struct {
		Type    string `json:"type"`
		Index   int    `json:"index"`
		Message *struct {
			StopReason string `json:"stop_reason"`
			Usage      *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		} `json:"message"`
		ContentBlock *struct {
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content_block"`
		Delta *struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			PartialJSON string `json:"partial_json"`
			StopReason  string `json:"stop_reason"`
		} `json:"delta"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if event.Message != nil {
		if event.Message.StopReason != "" {
			state.stopReason = event.Message.StopReason
		}
		if event.Message.Usage != nil {
			setAnthropicUsage(&state.usage, event.Message.Usage.InputTokens, event.Message.Usage.OutputTokens)
		}
	}
	if event.Usage != nil {
		setAnthropicUsage(&state.usage, event.Usage.InputTokens, event.Usage.OutputTokens)
	}

	switch event.Type {
	case "content_block_start":
		if event.ContentBlock == nil {
			return nil
		}
		switch event.ContentBlock.Type {
		case "text":
			if event.ContentBlock.Text != "" {
				state.content.WriteString(event.ContentBlock.Text)
				if onDelta != nil {
					onDelta(event.ContentBlock.Text)
				}
			}
		case "tool_use":
			tc := state.anthropicToolCall(event.Index)
			tc.ID = event.ContentBlock.ID
			tc.Name = event.ContentBlock.Name
			if len(event.ContentBlock.Input) > 0 && string(event.ContentBlock.Input) != "null" && string(event.ContentBlock.Input) != "{}" {
				tc.ArgsJSON.Write(event.ContentBlock.Input)
			}
		}
	case "content_block_delta":
		if event.Delta == nil {
			return nil
		}
		switch event.Delta.Type {
		case "text_delta":
			if event.Delta.Text != "" {
				state.content.WriteString(event.Delta.Text)
				if onDelta != nil {
					onDelta(event.Delta.Text)
				}
			}
		case "input_json_delta":
			if event.Delta.PartialJSON != "" {
				state.anthropicToolCall(event.Index).ArgsJSON.WriteString(event.Delta.PartialJSON)
			}
		}
	case "message_delta":
		if event.Delta != nil && event.Delta.StopReason != "" {
			state.stopReason = event.Delta.StopReason
		}
	}

	return nil
}

func (s *anthropicStreamState) anthropicToolCall(index int) *streamToolCall {
	tc, ok := s.toolCalls[index]
	if !ok {
		tc = &streamToolCall{}
		s.toolCalls[index] = tc
	}
	return tc
}

func setAnthropicUsage(usage *Usage, inputTokens, outputTokens int) {
	if inputTokens != 0 {
		usage.InputTokens = inputTokens
	}
	if outputTokens != 0 {
		usage.OutputTokens = outputTokens
	}
}
