package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestRunStoreStructuredEventsRedactPayloads(t *testing.T) {
	store := NewRunStore(10)
	req := RunAgentRequest{
		RequestID: "run-redact",
		Text:      "secret prompt body",
		Channel:   ChannelHTTP,
		Source:    "test",
	}
	store.Start(req)
	store.AddEvent("run-redact", EventToolCall, map[string]any{
		"tool":       "http",
		"args":       `{"api_key":"sk-secret","url":"https://example.com"}`,
		"safe_value": "ok",
	})
	store.AddEvent("run-redact", EventText, map[string]any{"text": "assistant body"})
	store.Complete("run-redact", RunAgentResponse{
		SessionID: "sess",
		Usage:     map[string]int{"input_tokens": 10, "output_tokens": 20, "total_tokens": 30},
	}, nil)

	record, ok := store.Get("run-redact")
	if !ok {
		t.Fatal("expected run record")
	}
	if len(record.StructuredEvents) < 4 {
		t.Fatalf("structured events = %d, want at least 4", len(record.StructuredEvents))
	}
	for idx, evt := range record.StructuredEvents {
		if evt.SchemaVersion != structuredEventSchemaVersion {
			t.Fatalf("schema version = %q, want %q", evt.SchemaVersion, structuredEventSchemaVersion)
		}
		wantID := fmt.Sprintf("run-redact-%06d", idx+1)
		if evt.ID != wantID {
			t.Fatalf("event id = %q, want %q", evt.ID, wantID)
		}
		if evt.RunID != "run-redact" {
			t.Fatalf("run id = %q, want run-redact", evt.RunID)
		}
		encoded, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal event: %v", err)
		}
		body := string(encoded)
		for _, forbidden := range []string{"sk-secret", "secret prompt body", "assistant body"} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("structured event leaked %q: %s", forbidden, body)
			}
		}
	}
}

func TestRunStoreMetricsShape(t *testing.T) {
	store := NewRunStore(10)
	store.Start(RunAgentRequest{RequestID: "ok-run", Channel: ChannelHTTP})
	store.AddEvent("ok-run", EventUsage, map[string]any{"input_tokens": 1, "output_tokens": 2})
	store.Complete("ok-run", RunAgentResponse{
		SessionID:    "sess",
		Usage:        map[string]int{"input_tokens": 10, "output_tokens": 20, "total_tokens": 30},
		BudgetStatus: &agent.TokenBudgetUsage{Status: agent.TokenBudgetStatusOK, InputTokens: 10, OutputTokens: 20, TotalTokens: 30},
		Routing:      &agent.RouteRecommendation{Complexity: agent.ComplexitySimple, Route: agent.RouteDirect, ModelTier: "small", Reason: "test"},
	}, nil)

	metrics := store.Metrics()
	if metrics["runs_total"] != 1 {
		t.Fatalf("runs_total = %v, want 1", metrics["runs_total"])
	}
	statuses := metrics["runs_by_status"].(map[string]int)
	if statuses["completed"] != 1 {
		t.Fatalf("completed count = %d, want 1", statuses["completed"])
	}
	events := metrics["events_by_type"].(map[string]int)
	if events["run_started"] != 1 || events["run_completed"] != 1 || events["budget_status"] != 1 || events["routing_selected"] != 1 || events[EventUsage] != 1 {
		t.Fatalf("event counts = %#v, want run start/completed/budget/routing/usage", events)
	}
	if metrics["tokens_input_total"] != 10 || metrics["tokens_output_total"] != 20 {
		t.Fatalf("token metrics = %#v, want 10/20", metrics)
	}
	if metrics["schema_version"] != structuredEventSchemaVersion {
		t.Fatalf("schema = %v, want %s", metrics["schema_version"], structuredEventSchemaVersion)
	}
}

func TestHandleMetrics(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"text":"hello","request_id":"metrics-smoke"}`
	resp, err := http.Post(ts.URL+"/message", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /message: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /message status = %d", resp.StatusCode)
	}

	var got struct {
		Metrics map[string]any `json:"metrics"`
	}
	getJSON(t, ts.URL+"/metrics", http.StatusOK, &got)
	if got.Metrics["schema_version"] != structuredEventSchemaVersion {
		t.Fatalf("schema = %v, want %s", got.Metrics["schema_version"], structuredEventSchemaVersion)
	}
	if got.Metrics["runs_total"].(float64) < 1 {
		t.Fatalf("runs_total = %v, want at least 1", got.Metrics["runs_total"])
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal metrics: %v", err)
	}
	if strings.Contains(string(encoded), "hello") {
		t.Fatalf("metrics leaked prompt body: %s", encoded)
	}
}

func TestRunStoreExportTracesJSONLRedactsPayloads(t *testing.T) {
	store := NewRunStore(10)
	store.Start(RunAgentRequest{RequestID: "trace-redact", Text: "secret prompt body", Channel: ChannelHTTP})
	store.AddEvent("trace-redact", EventToolCall, map[string]any{
		"tool": "http",
		"args": `{"api_key":"sk-secret"}`,
		"metadata": map[string]any{
			"request": "secret prompt body",
			"token":   "sk-secret",
			"safe":    "ok",
		},
	})
	store.Complete("trace-redact", RunAgentResponse{SessionID: "sess"}, nil)

	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := store.ExportTracesJSONL(path); err != nil {
		t.Fatalf("export traces: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace export: %v", err)
	}
	body := string(data)
	for _, forbidden := range []string{"secret prompt body", "sk-secret", `"args":`, `"request":`, `"token":`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("trace export leaked %q: %s", forbidden, body)
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	count := 0
	for scanner.Scan() {
		count++
		var record TraceExportRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("decode line %d: %v", count, err)
		}
		if record.SchemaVersion != structuredEventSchemaVersion || record.TraceID != "trace-redact" || record.EventID == "" || record.Name == "" || record.Phase == "" || record.Timestamp.IsZero() {
			t.Fatalf("record %d = %#v, want OTel-ready fields", count, record)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan trace export: %v", err)
	}
	if count == 0 {
		t.Fatal("expected trace records")
	}
}

func TestRunStoreExportRunTraceJSONLMissingRun(t *testing.T) {
	store := NewRunStore(10)
	err := store.ExportRunTraceJSONL("missing", filepath.Join(t.TempDir(), "trace.jsonl"))
	if err == nil || !strings.Contains(err.Error(), "run not found") {
		t.Fatalf("ExportRunTraceJSONL error = %v, want run not found", err)
	}
}

func TestHandleRunTrace(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	s.runStore.Start(RunAgentRequest{RequestID: "trace-api", Text: "secret prompt body", Channel: ChannelHTTP})
	s.runStore.AddEvent("trace-api", EventText, map[string]any{"text": "assistant body"})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var got struct {
		Trace []TraceExportRecord `json:"trace"`
	}
	getJSON(t, ts.URL+"/runs/trace-api/trace", http.StatusOK, &got)
	if len(got.Trace) != 2 {
		t.Fatalf("trace length = %d, want 2", len(got.Trace))
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal trace response: %v", err)
	}
	if strings.Contains(string(encoded), "secret prompt body") || strings.Contains(string(encoded), "assistant body") {
		t.Fatalf("trace response leaked prompt/text: %s", encoded)
	}
	getJSON(t, ts.URL+"/runs/missing/trace", http.StatusNotFound, &map[string]any{})
}

func TestHandleExportTraces(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	s.runStore.Start(RunAgentRequest{RequestID: "trace-export-api", Text: "secret prompt body", Channel: ChannelHTTP})
	s.runStore.AddEvent("trace-export-api", EventUsage, map[string]any{"input_tokens": 1})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var missing map[string]any
	getJSON(t, ts.URL+"/traces/export", http.StatusBadRequest, &missing)

	path := filepath.Join(t.TempDir(), "trace.jsonl")
	var got struct {
		Path   string `json:"path"`
		Events int    `json:"events"`
	}
	getJSON(t, ts.URL+"/traces/export?path="+path, http.StatusOK, &got)
	if got.Path != path || got.Events != 2 {
		t.Fatalf("export response = %#v, want path and 2 events", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace export: %v", err)
	}
	if strings.Contains(string(data), "secret prompt body") {
		t.Fatalf("trace export leaked prompt: %s", data)
	}
}
