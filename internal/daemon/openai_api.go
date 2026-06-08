package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openAIChatCompletionObject = "chat.completion"

type openAIChatCompletionRequest struct {
	Model          string              `json:"model"`
	Messages       []openAIChatMessage `json:"messages"`
	Stream         bool                `json:"stream,omitempty"`
	Tools          []map[string]any    `json:"tools,omitempty"`
	Functions      []map[string]any    `json:"functions,omitempty"`
	FunctionCall   any                 `json:"function_call,omitempty"`
	ToolChoice     any                 `json:"tool_choice,omitempty"`
	ResponseFormat map[string]any      `json:"response_format,omitempty"`
	N              int                 `json:"n,omitempty"`
	User           string              `json:"user,omitempty"`
	SessionID      string              `json:"session_id,omitempty"`
	RequestID      string              `json:"request_id,omitempty"`
	Agent          string              `json:"agent,omitempty"`
	Metadata       map[string]any      `json:"metadata,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type openAIChatCompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage,omitempty"`
	RunID   string         `json:"starclaw_run_id,omitempty"`
}

type openAIChoice struct {
	Index        int               `json:"index"`
	Message      openAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

func (s *Server) handleOpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil {
		writeOpenAIError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}

	req, ok := decodeOpenAIChatCompletionBody(w, r)
	if !ok {
		return
	}
	if err := validateOpenAIChatCompletionRequest(req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, err.Error())
		return
	}

	prompt := openAIChatPrompt(req.Messages)
	requestID := strings.TrimSpace(req.RequestID)
	if requestID == "" {
		requestID = generateRequestID()
	}

	runReq := RunAgentRequest{
		Text:      prompt,
		Agent:     strings.TrimSpace(req.Agent),
		Source:    "openai-compatible",
		Channel:   ChannelHTTP,
		Sender:    strings.TrimSpace(req.User),
		SessionID: strings.TrimSpace(req.SessionID),
		Model:     strings.TrimSpace(req.Model),
		RequestID: requestID,
	}
	s.runStore.Start(runReq)

	ctx, cancel := context.WithCancel(r.Context())
	s.running.Store(runReq.RequestID, cancel)
	defer s.running.Delete(runReq.RequestID)

	handler := s.recordingHandler(runReq.RequestID, &httpEventHandler{})
	result, err := s.runAgent(ctx, runReq, handler)
	s.runStore.Complete(runReq.RequestID, result, err)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.Error != "" {
		writeOpenAIError(w, http.StatusInternalServerError, result.Error)
		return
	}

	writeJSON(w, http.StatusOK, openAIResponseFromRun(req.Model, runReq.RequestID, result))
}

func decodeOpenAIChatCompletionBody(w http.ResponseWriter, r *http.Request) (openAIChatCompletionRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid request body")
		return openAIChatCompletionRequest{}, false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid request body")
		return openAIChatCompletionRequest{}, false
	}
	for field := range raw {
		if !openAIChatCompletionAllowedFields[field] {
			writeOpenAIError(w, http.StatusBadRequest, fmt.Sprintf("%s is not supported by the OpenAI-compatible gateway", field))
			return openAIChatCompletionRequest{}, false
		}
	}
	var req openAIChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid request body")
		return openAIChatCompletionRequest{}, false
	}
	return req, true
}

var openAIChatCompletionAllowedFields = map[string]bool{
	"model":           true,
	"messages":        true,
	"stream":          true,
	"tools":           true,
	"functions":       true,
	"function_call":   true,
	"tool_choice":     true,
	"response_format": true,
	"n":               true,
	"user":            true,
	"session_id":      true,
	"request_id":      true,
	"agent":           true,
	"metadata":        true,
}

func validateOpenAIChatCompletionRequest(req openAIChatCompletionRequest) error {
	if strings.TrimSpace(req.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages is required")
	}
	if req.Stream {
		return fmt.Errorf("stream=true is not supported by the OpenAI-compatible gateway")
	}
	if len(req.Tools) > 0 || len(req.Functions) > 0 || req.FunctionCall != nil || req.ToolChoice != nil {
		return fmt.Errorf("OpenAI tool/function calling fields are not supported; use StarClaw local tools through the daemon run")
	}
	if req.ResponseFormat != nil {
		return fmt.Errorf("response_format is not supported")
	}
	if req.Metadata != nil {
		return fmt.Errorf("metadata is not supported")
	}
	if req.N < 0 {
		return fmt.Errorf("n must be 1 when provided")
	}
	if req.N > 1 {
		return fmt.Errorf("n greater than 1 is not supported")
	}
	for i, msg := range req.Messages {
		role := strings.TrimSpace(msg.Role)
		if role != "system" && role != "user" && role != "assistant" {
			return fmt.Errorf("messages[%d].role must be system, user, or assistant", i)
		}
		if strings.TrimSpace(msg.Content) == "" {
			return fmt.Errorf("messages[%d].content is required", i)
		}
	}
	return nil
}

func openAIChatPrompt(messages []openAIChatMessage) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		role := strings.TrimSpace(msg.Role)
		content := strings.TrimSpace(msg.Content)
		if role == "user" {
			parts = append(parts, content)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", role, content))
	}
	return strings.Join(parts, "\n\n")
}

func openAIResponseFromRun(model, requestID string, result RunAgentResponse) openAIChatCompletionResponse {
	content := ""
	if len(result.Messages) > 0 {
		content = strings.Join(result.Messages, "\n")
	}
	usage := openAIUsage{}
	if result.Usage != nil {
		usage.PromptTokens = result.Usage["input_tokens"]
		usage.CompletionTokens = result.Usage["output_tokens"]
		usage.TotalTokens = result.Usage["total_tokens"]
	}
	return openAIChatCompletionResponse{
		ID:      "chatcmpl-" + requestID,
		Object:  openAIChatCompletionObject,
		Created: time.Now().Unix(),
		Model:   model,
		RunID:   requestID,
		Choices: []openAIChoice{{
			Index: 0,
			Message: openAIChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: "stop",
		}},
		Usage: usage,
	}
}

func writeOpenAIError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"message": msg,
			"type":    "invalid_request_error",
		},
	})
}
