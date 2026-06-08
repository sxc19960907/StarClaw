package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleOpenAIChatCompletions(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{
		"model":"request-model",
		"request_id":"oa-smoke",
		"user":"local-user",
		"messages":[
			{"role":"system","content":"Keep answers terse."},
			{"role":"user","content":"hello"}
		]
	}`
	var got openAIChatCompletionResponse
	postJSON(t, ts.URL+"/v1/chat/completions", body, http.StatusOK, &got)

	if got.ID != "chatcmpl-oa-smoke" {
		t.Fatalf("id = %q, want chatcmpl-oa-smoke", got.ID)
	}
	if got.Object != openAIChatCompletionObject {
		t.Fatalf("object = %q, want %q", got.Object, openAIChatCompletionObject)
	}
	if got.Model != "request-model" {
		t.Fatalf("model = %q, want request-model", got.Model)
	}
	if got.RunID != "oa-smoke" {
		t.Fatalf("starclaw_run_id = %q, want oa-smoke", got.RunID)
	}
	if len(got.Choices) != 1 {
		t.Fatalf("choice count = %d, want 1", len(got.Choices))
	}
	if got.Choices[0].Message.Role != "assistant" {
		t.Fatalf("choice role = %q, want assistant", got.Choices[0].Message.Role)
	}
	if got.Choices[0].Message.Content != "This is a mock response." {
		t.Fatalf("choice content = %q, want mock response", got.Choices[0].Message.Content)
	}
	if got.Choices[0].FinishReason != "stop" {
		t.Fatalf("finish_reason = %q, want stop", got.Choices[0].FinishReason)
	}
	if got.Usage.PromptTokens != 10 || got.Usage.CompletionTokens != 20 || got.Usage.TotalTokens != 30 {
		t.Fatalf("usage = %+v, want 10/20/30", got.Usage)
	}

	var run RunRecord
	getJSON(t, ts.URL+"/runs/oa-smoke", http.StatusOK, &run)
	if run.Channel != ChannelHTTP {
		t.Fatalf("run channel = %q, want %q", run.Channel, ChannelHTTP)
	}
	if run.Request.Source != "openai-compatible" {
		t.Fatalf("run source = %q, want openai-compatible", run.Request.Source)
	}
	if run.Request.Sender != "local-user" {
		t.Fatalf("run sender = %q, want local-user", run.Request.Sender)
	}
	if !strings.Contains(run.Request.Text, "system: Keep answers terse.") ||
		!strings.Contains(run.Request.Text, "hello") {
		t.Fatalf("run prompt missing expected chat content: %q", run.Request.Text)
	}
}

func TestHandleOpenAIChatCompletionsValidation(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "missing model",
			body: `{"messages":[{"role":"user","content":"hello"}]}`,
			want: "model is required",
		},
		{
			name: "missing messages",
			body: `{"model":"local"}`,
			want: "messages is required",
		},
		{
			name: "unsupported stream",
			body: `{"model":"local","stream":true,"messages":[{"role":"user","content":"hello"}]}`,
			want: "stream=true is not supported",
		},
		{
			name: "unsupported tools",
			body: `{"model":"local","tools":[{"type":"function"}],"messages":[{"role":"user","content":"hello"}]}`,
			want: "tool/function calling fields are not supported",
		},
		{
			name: "unsupported unknown field",
			body: `{"model":"local","temperature":0.2,"messages":[{"role":"user","content":"hello"}]}`,
			want: "temperature is not supported",
		},
		{
			name: "unsupported metadata",
			body: `{"model":"local","metadata":{"trace":"x"},"messages":[{"role":"user","content":"hello"}]}`,
			want: "metadata is not supported",
		},
		{
			name: "multiple choices",
			body: `{"model":"local","n":2,"messages":[{"role":"user","content":"hello"}]}`,
			want: "n greater than 1 is not supported",
		},
		{
			name: "invalid role",
			body: `{"model":"local","messages":[{"role":"tool","content":"hello"}]}`,
			want: "messages[0].role must be system, user, or assistant",
		},
		{
			name: "empty content",
			body: `{"model":"local","messages":[{"role":"user","content":"   "}]}`,
			want: "messages[0].content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			postJSON(t, ts.URL+"/v1/chat/completions", tt.body, http.StatusBadRequest, &got)
			if !strings.Contains(got.Error.Message, tt.want) {
				t.Fatalf("error message = %q, want containing %q", got.Error.Message, tt.want)
			}
			if got.Error.Type != "invalid_request_error" {
				t.Fatalf("error type = %q, want invalid_request_error", got.Error.Type)
			}
		})
	}
}
