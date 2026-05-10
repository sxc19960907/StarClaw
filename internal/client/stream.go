package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

func newStreamState() *streamState {
	return &streamState{
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
		for i := 0; i < len(s.toolCalls); i++ {
			tc := s.toolCalls[i]
			if tc == nil {
				continue
			}
			resp.ToolUses = append(resp.ToolUses, ToolUse{
				ID:    tc.ID,
				Name:  tc.Name,
				Input: json.RawMessage(tc.ArgsJSON.String()),
			})
		}
	}
	return resp
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
