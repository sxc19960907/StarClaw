package daemon

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestHandleOpenAIChatCompletionsStreaming(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{
		"model":"request-model",
		"request_id":"oa-stream",
		"user":"stream-user",
		"stream":true,
		"messages":[{"role":"user","content":"hello stream"}]
	}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/v1/chat/completions", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /v1/chat/completions stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, data)
	}
	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", got)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read stream: %v", err)
	}
	stream := string(data)
	if !strings.Contains(stream, "data: [DONE]") {
		t.Fatalf("stream missing DONE:\n%s", stream)
	}
	chunks := decodeOpenAIStreamChunks(t, stream)
	if len(chunks) < 3 {
		t.Fatalf("chunks len = %d, want at least 3: %#v\n%s", len(chunks), chunks, stream)
	}
	if chunks[0].Object != openAIChatCompletionChunkObject || chunks[0].Choices[0].Delta["role"] != "assistant" {
		t.Fatalf("role chunk = %#v", chunks[0])
	}
	foundContent := false
	foundStop := false
	for _, chunk := range chunks {
		if chunk.ID != "chatcmpl-oa-stream" || chunk.Model != "request-model" || chunk.RunID != "oa-stream" {
			t.Fatalf("chunk identity = %#v", chunk)
		}
		if chunk.Choices[0].Delta["content"] != "" {
			foundContent = true
		}
		if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason == "stop" {
			foundStop = true
		}
	}
	if !foundContent {
		t.Fatalf("stream missing content chunk: %#v", chunks)
	}
	if !foundStop {
		t.Fatalf("stream missing stop chunk: %#v", chunks)
	}

	var run RunRecord
	getJSON(t, ts.URL+"/runs/oa-stream", http.StatusOK, &run)
	if run.ID != "oa-stream" || run.Request.Source != "openai-compatible" || run.Request.Sender != "stream-user" {
		t.Fatalf("run = %#v", run)
	}
	if len(run.StructuredEvents) == 0 {
		t.Fatal("expected streaming run structured events")
	}
}

func TestPhase5APIObservabilitySmoke(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	const (
		runID      = "phase5-api-smoke"
		promptText = "phase5 secret prompt body"
	)

	body := `{
		"model":"phase5-local-model",
		"request_id":"phase5-api-smoke",
		"user":"phase5-user",
		"messages":[
			{"role":"system","content":"Keep observability aggregate-safe."},
			{"role":"user","content":"phase5 secret prompt body"}
		]
	}`
	var completion openAIChatCompletionResponse
	postJSON(t, ts.URL+"/v1/chat/completions", body, http.StatusOK, &completion)
	if completion.ID != "chatcmpl-"+runID {
		t.Fatalf("id = %q, want chatcmpl-%s", completion.ID, runID)
	}
	if completion.Object != openAIChatCompletionObject {
		t.Fatalf("object = %q, want %q", completion.Object, openAIChatCompletionObject)
	}
	if completion.Model != "phase5-local-model" {
		t.Fatalf("model = %q, want phase5-local-model", completion.Model)
	}
	if completion.RunID != runID {
		t.Fatalf("starclaw_run_id = %q, want %s", completion.RunID, runID)
	}
	if len(completion.Choices) != 1 || completion.Choices[0].Message.Role != "assistant" || completion.Choices[0].FinishReason != "stop" {
		t.Fatalf("choices = %#v, want one stopped assistant message", completion.Choices)
	}
	if completion.Usage.PromptTokens != 10 || completion.Usage.CompletionTokens != 20 || completion.Usage.TotalTokens != 30 {
		t.Fatalf("usage = %+v, want 10/20/30", completion.Usage)
	}

	var run RunRecord
	getJSON(t, ts.URL+"/runs/"+runID, http.StatusOK, &run)
	if run.ID != runID || run.Request.Source != "openai-compatible" || run.Channel != ChannelHTTP {
		t.Fatalf("run = %#v, want OpenAI-compatible HTTP run %s", run, runID)
	}
	if run.Request.Sender != "phase5-user" {
		t.Fatalf("run sender = %q, want phase5-user", run.Request.Sender)
	}
	if len(run.StructuredEvents) == 0 {
		t.Fatal("expected structured events")
	}
	if run.Usage["input_tokens"] != 10 || run.Usage["output_tokens"] != 20 || run.Usage["total_tokens"] != 30 {
		t.Fatalf("run usage = %#v, want 10/20/30", run.Usage)
	}

	var metrics struct {
		Metrics map[string]any `json:"metrics"`
	}
	getJSON(t, ts.URL+"/metrics", http.StatusOK, &metrics)
	if metrics.Metrics["schema_version"] != structuredEventSchemaVersion {
		t.Fatalf("metrics schema = %v, want %s", metrics.Metrics["schema_version"], structuredEventSchemaVersion)
	}
	if metrics.Metrics["runs_total"].(float64) < 1 {
		t.Fatalf("runs_total = %v, want at least 1", metrics.Metrics["runs_total"])
	}
	assertJSONDoesNotContain(t, metrics, promptText, "metrics")

	var traceResp struct {
		Trace []TraceExportRecord `json:"trace"`
	}
	getJSON(t, ts.URL+"/runs/"+runID+"/trace", http.StatusOK, &traceResp)
	if len(traceResp.Trace) == 0 {
		t.Fatal("expected trace records")
	}
	for idx, record := range traceResp.Trace {
		if record.SchemaVersion != structuredEventSchemaVersion ||
			record.TraceID != runID ||
			record.RunID != runID ||
			record.EventID == "" ||
			record.Name == "" ||
			record.Phase == "" ||
			record.Timestamp.IsZero() {
			t.Fatalf("trace record %d = %#v, want OTel-ready fields", idx, record)
		}
	}
	assertJSONDoesNotContain(t, traceResp, promptText, "trace response")

	exportPath := filepath.Join(t.TempDir(), "phase5-traces.jsonl")
	var exportResp struct {
		Path   string `json:"path"`
		Events int    `json:"events"`
	}
	getJSON(t, ts.URL+"/traces/export?path="+exportPath, http.StatusOK, &exportResp)
	if exportResp.Path != exportPath || exportResp.Events < len(traceResp.Trace) {
		t.Fatalf("export response = %#v, want path and at least %d events", exportResp, len(traceResp.Trace))
	}
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read trace export: %v", err)
	}
	if strings.Contains(string(data), promptText) {
		t.Fatalf("trace export leaked prompt: %s", data)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var record TraceExportRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("decode trace export line %d: %v", lineCount, err)
		}
		if record.SchemaVersion != structuredEventSchemaVersion || record.TraceID == "" || record.EventID == "" || record.Name == "" || record.Phase == "" || record.Timestamp.IsZero() {
			t.Fatalf("export record %d = %#v, want OTel-ready fields", lineCount, record)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan trace export: %v", err)
	}
	if lineCount == 0 {
		t.Fatal("expected exported trace records")
	}

	var control struct {
		Status string `json:"status"`
		Action string `json:"action"`
		Replay struct {
			RequiresApproval bool           `json:"requires_approval"`
			Request          map[string]any `json:"request"`
		} `json:"replay"`
	}
	postJSON(t, ts.URL+"/runs/"+runID+"/control", `{"action":"replay","approved":false}`, http.StatusOK, &control)
	if control.Status != "approval_required" || control.Action != "replay" || !control.Replay.RequiresApproval {
		t.Fatalf("control response = %#v, want replay approval requirement", control)
	}
	if control.Replay.Request["text_redacted"] != true || control.Replay.Request["source"] != "openai-compatible" {
		t.Fatalf("replay request = %#v, want redacted OpenAI-compatible request", control.Replay.Request)
	}
	assertJSONDoesNotContain(t, control, promptText, "control response")

	var controlledRun RunRecord
	getJSON(t, ts.URL+"/runs/"+runID, http.StatusOK, &controlledRun)
	if len(controlledRun.Control) != 1 || controlledRun.Control[0].Action != "replay" || controlledRun.Control[0].Status != "approval_required" {
		t.Fatalf("control decisions = %#v, want replay approval_required", controlledRun.Control)
	}
	if len(controlledRun.Steps) != 1 || controlledRun.Steps[0].ID != "replay-approval" || controlledRun.Steps[0].Status != WorkflowStepWaitingApproval {
		t.Fatalf("workflow steps = %#v, want waiting replay approval", controlledRun.Steps)
	}
}

func assertJSONDoesNotContain(t *testing.T, value any, forbidden, label string) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", label, err)
	}
	if strings.Contains(string(encoded), forbidden) {
		t.Fatalf("%s leaked %q: %s", label, forbidden, encoded)
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

func decodeOpenAIStreamChunks(t *testing.T, stream string) []openAIChatCompletionChunk {
	t.Helper()
	var chunks []openAIChatCompletionChunk
	for _, block := range strings.Split(stream, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if !strings.HasPrefix(block, "data: ") {
			t.Fatalf("unexpected stream block %q in stream:\n%s", block, stream)
		}
		payload := strings.TrimPrefix(block, "data: ")
		if payload == "[DONE]" {
			continue
		}
		var chunk openAIChatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			t.Fatalf("decode stream chunk %q: %v", payload, err)
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}
