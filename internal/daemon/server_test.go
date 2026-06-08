package daemon

import (
	"archive/zip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/mcp"
	"github.com/starclaw/starclaw/internal/permissions"
	"github.com/starclaw/starclaw/internal/schedule"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/skills"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Test setup
// ---------------------------------------------------------------------------

func newTestServer(t *testing.T, deps *ServerDeps) *Server {
	t.Helper()
	return NewServer(0, deps, "test-version")
}

func newTestServerDeps(t *testing.T) *ServerDeps {
	t.Helper()
	return &ServerDeps{
		StarclawDir:     t.TempDir(),
		ConfigPath:      filepath.Join(t.TempDir(), "config.json"),
		AgentsDir:       t.TempDir(),
		SkillsDir:       t.TempDir(),
		InstructionsDir: t.TempDir(),
		LLMClient:       &mockLLMClient{t: t},
		Registry:        agent.NewToolRegistry(),
	}
}

// ---------------------------------------------------------------------------
// Health endpoint
// ---------------------------------------------------------------------------

func TestHandleHealth(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
	if body["version"] != "test-version" {
		t.Errorf("expected version=test-version, got %q", body["version"])
	}
}

// ---------------------------------------------------------------------------
// Status endpoint
// ---------------------------------------------------------------------------

func TestHandleStatus(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["version"] != "test-version" {
		t.Errorf("expected version=test-version, got %q", body["version"])
	}
	if _, ok := body["uptime"]; !ok {
		t.Errorf("expected uptime in response")
	}
	if _, ok := body["active_agents"]; !ok {
		t.Errorf("expected active_agents in response")
	}
}

// ---------------------------------------------------------------------------
// Schedule CRUD
// ---------------------------------------------------------------------------

func TestScheduleCRUD(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))

	deps := newTestServerDeps(t)
	deps.ScheduleManager = mgr

	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// CREATE
	createBody := `{"agent":"","cron":"* * * * *","prompt":"run every minute"}`
	resp, err := http.Post(ts.URL+"/schedules", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("POST /schedules: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var created schedule.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	_ = resp.Body.Close()

	if created.ID == "" {
		t.Error("expected non-empty schedule ID")
	}
	if created.Prompt != "run every minute" {
		t.Errorf("expected prompt 'run every minute', got %q", created.Prompt)
	}
	if !created.Enabled {
		t.Error("expected schedule to be enabled by default")
	}

	scheduleID := created.ID

	// LIST
	resp, err = http.Get(ts.URL + "/schedules")
	if err != nil {
		t.Fatalf("GET /schedules: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var listResp struct {
		Schedules []schedule.Schedule `json:"schedules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(listResp.Schedules))
	}
	if listResp.Schedules[0].ID != scheduleID {
		t.Errorf("expected ID %q, got %q", scheduleID, listResp.Schedules[0].ID)
	}

	// GET by ID
	resp, err = http.Get(ts.URL + "/schedules/" + scheduleID)
	if err != nil {
		t.Fatalf("GET /schedules/{id}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var got schedule.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.ID != scheduleID {
		t.Errorf("expected ID %q, got %q", scheduleID, got.ID)
	}

	// PATCH (update)
	patchBody := `{"enabled":false}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/schedules/"+scheduleID, strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /schedules/{id}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var updated schedule.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.Enabled {
		t.Error("expected schedule to be disabled after PATCH")
	}

	// DELETE
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/schedules/"+scheduleID, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /schedules/{id}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	resp, err = http.Get(ts.URL + "/schedules/" + scheduleID)
	if err != nil {
		t.Fatalf("GET deleted: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", resp.StatusCode)
	}
}

func TestDaemonAPISmokeWorkflow(t *testing.T) {
	deps := newTestServerDeps(t)
	deps.ScheduleManager = schedule.NewManager(filepath.Join(t.TempDir(), "schedules.json"))
	writeTestAgent(t, deps.AgentsDir, "api-agent")
	writeTestSkill(t, deps.SkillsDir, "summarizer", "Summarize local evidence")

	sessionMgr := session.NewManager(filepath.Join(deps.StarclawDir, "sessions"))
	sess := sessionMgr.NewSession()
	sess.Title = "API Smoke Session"
	sess.Messages = []client.Message{
		{Role: "user", Content: "smoke question"},
		{Role: "assistant", Content: "smoke answer"},
	}
	if err := sessionMgr.Save(); err != nil {
		t.Fatalf("save session: %v", err)
	}

	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var health map[string]string
	getJSON(t, ts.URL+"/health", http.StatusOK, &health)
	if health["status"] != "ok" || health["version"] != "test-version" {
		t.Fatalf("unexpected health response: %#v", health)
	}

	var status struct {
		Uptime       int    `json:"uptime"`
		Version      string `json:"version"`
		ActiveAgents int    `json:"active_agents"`
	}
	getJSON(t, ts.URL+"/status", http.StatusOK, &status)
	if status.Version != "test-version" {
		t.Fatalf("status version = %q, want test-version", status.Version)
	}

	var agentsResp struct {
		Agents []struct {
			Name        string `json:"Name"`
			Description string `json:"Description"`
		} `json:"agents"`
	}
	getJSON(t, ts.URL+"/agents", http.StatusOK, &agentsResp)
	if len(agentsResp.Agents) != 1 || agentsResp.Agents[0].Name != "api-agent" {
		t.Fatalf("unexpected agents response: %#v", agentsResp.Agents)
	}

	var agentResp map[string]interface{}
	getJSON(t, ts.URL+"/agents/api-agent", http.StatusOK, &agentResp)
	if agentResp["Name"] != "api-agent" {
		t.Fatalf("agent response missing Name: %#v", agentResp)
	}

	var skillsResp struct {
		Skills []skills.SkillMeta `json:"skills"`
	}
	getJSON(t, ts.URL+"/skills", http.StatusOK, &skillsResp)
	if len(skillsResp.Skills) != 1 || skillsResp.Skills[0].Name != "summarizer" {
		t.Fatalf("unexpected skills response: %#v", skillsResp.Skills)
	}

	var sessionsResp struct {
		Sessions []session.SessionSummary `json:"sessions"`
	}
	getJSON(t, ts.URL+"/sessions", http.StatusOK, &sessionsResp)
	if len(sessionsResp.Sessions) != 1 || sessionsResp.Sessions[0].ID != sess.ID {
		t.Fatalf("unexpected sessions response: %#v", sessionsResp.Sessions)
	}

	var sessionDetail session.Session
	getJSON(t, ts.URL+"/sessions/"+sess.ID, http.StatusOK, &sessionDetail)
	if sessionDetail.ID != sess.ID {
		t.Fatalf("session detail id = %q, want %q", sessionDetail.ID, sess.ID)
	}
	if len(sessionDetail.Messages) != 2 || sessionDetail.Messages[1].Content != "smoke answer" {
		t.Fatalf("unexpected session detail messages: %#v", sessionDetail.Messages)
	}

	var searchResp struct {
		Results []session.SessionSummary `json:"results"`
	}
	getJSON(t, ts.URL+"/sessions/search?q=smoke", http.StatusOK, &searchResp)
	if len(searchResp.Results) != 1 || searchResp.Results[0].ID != sess.ID {
		t.Fatalf("unexpected session search response: %#v", searchResp.Results)
	}

	var created schedule.Schedule
	postJSON(t, ts.URL+"/schedules", `{"agent":"","cron":"* * * * *","prompt":"run smoke"}`, http.StatusCreated, &created)
	if created.ID == "" || created.Prompt != "run smoke" || !created.Enabled {
		t.Fatalf("unexpected created schedule: %#v", created)
	}

	var schedulesResp struct {
		Schedules []schedule.Schedule `json:"schedules"`
	}
	getJSON(t, ts.URL+"/schedules", http.StatusOK, &schedulesResp)
	if len(schedulesResp.Schedules) != 1 || schedulesResp.Schedules[0].ID != created.ID {
		t.Fatalf("unexpected schedules response: %#v", schedulesResp.Schedules)
	}

	var gotSchedule schedule.Schedule
	getJSON(t, ts.URL+"/schedules/"+created.ID, http.StatusOK, &gotSchedule)
	if gotSchedule.ID != created.ID {
		t.Fatalf("schedule id = %q, want %q", gotSchedule.ID, created.ID)
	}

	var updated schedule.Schedule
	patchJSON(t, ts.URL+"/schedules/"+created.ID, `{"enabled":false}`, http.StatusOK, &updated)
	if updated.Enabled {
		t.Fatal("expected schedule to be disabled")
	}

	deleteJSON(t, ts.URL+"/schedules/"+created.ID, http.StatusOK)
	getJSON(t, ts.URL+"/schedules/"+created.ID, http.StatusNotFound, &map[string]string{})
	getJSON(t, ts.URL+"/agents/missing-agent", http.StatusNotFound, &map[string]string{})
	getJSON(t, ts.URL+"/sessions/search", http.StatusBadRequest, &map[string]string{})
}

// ---------------------------------------------------------------------------
// POST /message
// ---------------------------------------------------------------------------

func TestHandleMessage(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"text":"hello"}`
	resp, err := http.Post(ts.URL+"/message", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /message: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result RunAgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if len(result.Messages) == 0 {
		t.Error("expected at least one message")
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
}

func TestRunHistoryAPI(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"text":"hello","request_id":"run-smoke"}`
	resp, err := http.Post(ts.URL+"/message", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /message: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /message status = %d", resp.StatusCode)
	}

	var list struct {
		Runs []RunSummary `json:"runs"`
	}
	getJSON(t, ts.URL+"/runs", http.StatusOK, &list)
	if len(list.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(list.Runs))
	}
	if list.Runs[0].ID != "run-smoke" {
		t.Fatalf("run id = %q, want run-smoke", list.Runs[0].ID)
	}
	if list.Runs[0].Status != "completed" {
		t.Fatalf("run status = %q, want completed", list.Runs[0].Status)
	}
	if list.Runs[0].Prompt != "hello" {
		t.Fatalf("run prompt = %q, want hello", list.Runs[0].Prompt)
	}

	var detail RunRecord
	getJSON(t, ts.URL+"/runs/run-smoke", http.StatusOK, &detail)
	if detail.ID != "run-smoke" {
		t.Fatalf("detail id = %q, want run-smoke", detail.ID)
	}
	if detail.Status != "completed" {
		t.Fatalf("detail status = %q, want completed", detail.Status)
	}
	if detail.Channel != ChannelHTTP {
		t.Fatalf("detail channel = %q, want %q", detail.Channel, ChannelHTTP)
	}
	if detail.Request.Text != "hello" {
		t.Fatalf("detail request text = %q, want hello", detail.Request.Text)
	}
	if detail.Response == nil || detail.Response.SessionID == "" {
		t.Fatalf("detail response missing session id: %#v", detail.Response)
	}
	getJSON(t, ts.URL+"/runs/missing-run", http.StatusNotFound, &map[string]string{})
}

func TestSSEEventHandlerToolPayloads(t *testing.T) {
	rec := httptest.NewRecorder()
	h := &sseEventHandler{w: rec, flusher: rec}

	h.OnToolCall("file_read", `{"path":"README.md"}`)
	h.OnToolResult("file_read", agent.ToolResult{
		Content:       "read failed",
		IsError:       true,
		ErrorCategory: agent.ErrCategoryPermission,
	})

	toolCall := decodeSSEEventData(t, rec.Body.String(), "tool_call")
	if toolCall["tool"] != "file_read" {
		t.Fatalf("tool_call tool = %#v, want file_read", toolCall["tool"])
	}
	if toolCall["status"] != "running" {
		t.Fatalf("tool_call status = %#v, want running", toolCall["status"])
	}
	if toolCall["args"] != `{"path":"README.md"}` {
		t.Fatalf("tool_call args = %#v", toolCall["args"])
	}

	toolResult := decodeSSEEventData(t, rec.Body.String(), "tool_result")
	if toolResult["tool"] != "file_read" {
		t.Fatalf("tool_result tool = %#v, want file_read", toolResult["tool"])
	}
	if toolResult["status"] != "error" {
		t.Fatalf("tool_result status = %#v, want error", toolResult["status"])
	}
	if toolResult["content"] != "read failed" {
		t.Fatalf("tool_result content = %#v, want read failed", toolResult["content"])
	}
	if toolResult["is_error"] != true {
		t.Fatalf("tool_result is_error = %#v, want true", toolResult["is_error"])
	}
	if toolResult["error_category"] != string(agent.ErrCategoryPermission) {
		t.Fatalf("tool_result error_category = %#v, want %q", toolResult["error_category"], agent.ErrCategoryPermission)
	}
}

func TestSSEEventHandlerStreamsDeltasAsTextWithoutFinalDuplicate(t *testing.T) {
	rec := httptest.NewRecorder()
	h := &sseEventHandler{w: rec, flusher: rec}

	h.OnStreamDelta("hello ")
	h.OnStreamDelta("world")
	h.OnText("hello world")

	events := decodeSSEEvents(t, rec.Body.String(), "text")
	if len(events) != 2 {
		t.Fatalf("text event count = %d, want 2; stream:\n%s", len(events), rec.Body.String())
	}
	if events[0]["text"] != "hello " {
		t.Fatalf("first delta text = %#v, want hello ", events[0]["text"])
	}
	if events[1]["text"] != "world" {
		t.Fatalf("second delta text = %#v, want world", events[1]["text"])
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestHandleMessageMissingText(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"text":""}`
	resp, err := http.Post(ts.URL+"/message", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /message: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for empty text, got %d", resp.StatusCode)
	}
}

func TestHandleMessageInvalidBody(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Invalid JSON
	resp, err := http.Post(ts.URL+"/message", "application/json", strings.NewReader(`{not json}`))
	if err != nil {
		t.Fatalf("POST /message: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestHandleScheduleNotFound(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))

	deps := newTestServerDeps(t)
	deps.ScheduleManager = mgr

	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/schedules/nonexistent")
	if err != nil {
		t.Fatalf("GET /schedules/nonexistent: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent schedule, got %d", resp.StatusCode)
	}
}

func TestHandleScheduleCreateInvalid(t *testing.T) {
	dir := t.TempDir()
	mgr := schedule.NewManager(filepath.Join(dir, "schedules.json"))

	deps := newTestServerDeps(t)
	deps.ScheduleManager = mgr

	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Empty prompt should be rejected.
	body := `{"agent":"","cron":"* * * * *","prompt":""}`
	resp, err := http.Post(ts.URL+"/schedules", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /schedules: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty prompt, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Agent listing
// ---------------------------------------------------------------------------

func TestHandleAgents(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/agents")
	if err != nil {
		t.Fatalf("GET /agents: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Agents []interface{} `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode agents: %v", err)
	}
	if body.Agents == nil {
		t.Error("expected empty agents list, not nil")
	}
}

// ---------------------------------------------------------------------------
// Skills
// ---------------------------------------------------------------------------

func TestHandleSkills(t *testing.T) {
	deps := newTestServerDeps(t)
	writeTestSkill(t, deps.SkillsDir, "summarizer", "Summarize local evidence")

	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/skills")
	if err != nil {
		t.Fatalf("GET /skills: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Skills []skills.SkillMeta `json:"skills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if len(body.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(body.Skills))
	}
	if body.Skills[0].Name != "summarizer" {
		t.Fatalf("skill name = %q, want summarizer", body.Skills[0].Name)
	}
	if body.Skills[0].Description != "Summarize local evidence" {
		t.Fatalf("skill description = %q", body.Skills[0].Description)
	}
	if body.Skills[0].Source != skills.SourceGlobal {
		t.Fatalf("skill source = %q, want global", body.Skills[0].Source)
	}
}

func TestHandleSkillsEmpty(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/skills")
	if err != nil {
		t.Fatalf("GET /skills: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Skills []skills.SkillMeta `json:"skills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if body.Skills == nil {
		t.Fatal("expected empty skills list, not nil")
	}
	if len(body.Skills) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(body.Skills))
	}
}

// ---------------------------------------------------------------------------
// Config / Instructions
// ---------------------------------------------------------------------------

func TestHandleConfigGetPatch(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// GET config (file does not exist yet)
	resp, err := http.Get(ts.URL + "/config")
	if err != nil {
		t.Fatalf("GET /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// PATCH config
	patchBody := `{"endpoint":"https://api.example.com"}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleConfigGetRedactsYAMLSecrets(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
provider: openai
openai_api_key: sk-secret
openai_endpoint: https://api.openai.test/v1
openai_model: gpt-test
api_key: anthropic-secret
mcp_servers:
  browser:
    command: npx
    args: ["@playwright/mcp@latest"]
    env:
      BROWSER_TOKEN: browser-secret
    context: browser context
    keep_alive: true
  disabled_http:
    type: http
    url: http://127.0.0.1:8888/mcp
    disabled: true
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body struct {
		Config daemonConfigView `json:"config"`
	}
	getJSON(t, ts.URL+"/config", http.StatusOK, &body)
	if body.Config.Provider != "openai" {
		t.Fatalf("provider = %q, want openai", body.Config.Provider)
	}
	if !body.Config.OpenAIAPIKeySet || !body.Config.APIKeySet {
		t.Fatalf("expected key presence booleans, got %+v", body.Config)
	}
	if len(body.Config.MCPServers) != 2 {
		t.Fatalf("mcp server count = %d, want 2: %+v", len(body.Config.MCPServers), body.Config.MCPServers)
	}
	if body.Config.MCPServers[0].Name != "browser" || body.Config.MCPServers[0].Type != "stdio" {
		t.Fatalf("unexpected first mcp server: %+v", body.Config.MCPServers[0])
	}
	if len(body.Config.MCPServers[0].EnvKeys) != 1 || body.Config.MCPServers[0].EnvKeys[0] != "BROWSER_TOKEN" {
		t.Fatalf("unexpected env key redaction: %+v", body.Config.MCPServers[0].EnvKeys)
	}
	if !body.Config.MCPServers[0].Context || !body.Config.MCPServers[0].KeepAlive {
		t.Fatalf("expected context and keep_alive flags: %+v", body.Config.MCPServers[0])
	}
	if !body.Config.MCPServers[1].Disabled || body.Config.MCPServers[1].URL == "" {
		t.Fatalf("unexpected disabled http mcp server: %+v", body.Config.MCPServers[1])
	}
	raw, err := http.Get(ts.URL + "/config")
	if err != nil {
		t.Fatalf("GET /config: %v", err)
	}
	defer func() {
		_ = raw.Body.Close()
	}()
	var rawBody map[string]interface{}
	if err := json.NewDecoder(raw.Body).Decode(&rawBody); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	encoded, _ := json.Marshal(rawBody)
	if strings.Contains(string(encoded), "sk-secret") || strings.Contains(string(encoded), "anthropic-secret") || strings.Contains(string(encoded), "browser-secret") {
		t.Fatalf("GET /config leaked secret: %s", encoded)
	}
}

func TestHandleConfigPatchWritesYAMLAndPreservesBlankSecrets(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
provider: openai
openai_api_key: existing-secret
openai_endpoint: https://old.example/v1
openai_model: old-model
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{Provider: "openai", OpenAIAPIKey: "existing-secret"}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	patchBody := `{"provider":"ollama","ollama_endpoint":"http://127.0.0.1:11434","ollama_model":"llama3.2","openai_api_key":""}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := yaml.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved config is not valid YAML: %v\n%s", err, data)
	}
	if saved.Provider != "ollama" || saved.OllamaModel != "llama3.2" {
		t.Fatalf("unexpected saved config: %+v", saved)
	}
	if saved.OpenAIAPIKey != "existing-secret" {
		t.Fatalf("blank patch overwrote OpenAI key: %q", saved.OpenAIAPIKey)
	}
	if deps.Config.Provider != "ollama" || deps.Config.OllamaModel != "llama3.2" {
		t.Fatalf("in-memory config not refreshed: %+v", deps.Config)
	}
}

func TestHandleConfigPatchUpdatesMCPServersAndPreservesBlankEnv(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
provider: openai
mcp_servers:
  browser:
    command: npx
    args: ["@playwright/mcp@latest"]
    env:
      BROWSER_TOKEN: existing-token
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	patchBody := `{"mcp_servers":[{"name":"browser","type":"stdio","command":"npx","args":["@playwright/mcp@latest","--headless"],"env":{"BROWSER_TOKEN":"","NEW_TOKEN":"new-secret"},"context":"browser context","keep_alive":true},{"name":"docs","type":"http","url":"http://127.0.0.1:3000/mcp","disabled":true}]}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := yaml.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved config is not valid YAML: %v\n%s", err, data)
	}
	browser := saved.MCPServers["browser"]
	if browser.Command != "npx" || len(browser.Args) != 2 || browser.Args[1] != "--headless" {
		t.Fatalf("unexpected browser config: %+v", browser)
	}
	if browser.Env["BROWSER_TOKEN"] != "existing-token" || browser.Env["NEW_TOKEN"] != "new-secret" {
		t.Fatalf("unexpected browser env preservation: %+v", browser.Env)
	}
	if browser.Context != "browser context" || !browser.KeepAlive {
		t.Fatalf("unexpected browser metadata: %+v", browser)
	}
	docs := saved.MCPServers["docs"]
	if docs.Type != "http" || docs.URL != "http://127.0.0.1:3000/mcp" || !docs.Disabled {
		t.Fatalf("unexpected docs config: %+v", docs)
	}
	if deps.Config.MCPServers["browser"].Env["BROWSER_TOKEN"] != "existing-token" {
		t.Fatalf("in-memory MCP config not refreshed: %+v", deps.Config.MCPServers)
	}

	var body struct {
		Config daemonConfigView `json:"config"`
	}
	getJSON(t, ts.URL+"/config", http.StatusOK, &body)
	encoded, _ := json.Marshal(body)
	if strings.Contains(string(encoded), "existing-token") || strings.Contains(string(encoded), "new-secret") {
		t.Fatalf("GET /config leaked MCP env secret: %s", encoded)
	}
	if len(body.Config.MCPServers) != 2 || len(body.Config.MCPServers[0].EnvKeys) == 0 {
		t.Fatalf("expected redacted MCP env keys in view: %+v", body.Config.MCPServers)
	}
}

func TestHandleConfigPatchRejectsInvalidMCPServer(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	tests := []struct {
		name string
		body string
	}{
		{"bad name", `{"mcp_servers":[{"name":"../bad","type":"stdio","command":"npx"}]}`},
		{"missing command", `{"mcp_servers":[{"name":"browser","type":"stdio"}]}`},
		{"missing url", `{"mcp_servers":[{"name":"docs","type":"http"}]}`},
		{"unsupported type", `{"mcp_servers":[{"name":"docs","type":"sse","url":"http://127.0.0.1"}]}`},
		{"blank env key", `{"mcp_servers":[{"name":"browser","type":"stdio","command":"npx","env":{" ":"secret"}}]}`},
		{"duplicate", `{"mcp_servers":[{"name":"browser","type":"stdio","command":"npx"},{"name":"browser","type":"stdio","command":"node"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("PATCH /config: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
		})
	}
}

func TestHandleConfigPatchRejectsUnsupportedProvider(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(`{"provider":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleMCPTestDisabledAndMissing(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
mcp_servers:
  disabled:
    command: echo
    disabled: true
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp/test", strings.NewReader(`{"name":"missing"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp/test missing: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", resp.StatusCode)
	}

	var body mcpTestResponse
	postJSON(t, ts.URL+"/mcp/test", `{"name":"disabled"}`, http.StatusOK, &body)
	if body.Status != "disabled" || body.Error == "" {
		t.Fatalf("disabled response = %+v", body)
	}
}

func TestHandleMCPTestDiscoversTools(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
mcp_servers:
  local:
    command: local-mcp
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{}
	deps.MCPTester = func(name string, server mcp.MCPServerConfig) ([]mcp.RemoteTool, error) {
		if name != "local" || server.Command != "local-mcp" {
			t.Fatalf("unexpected MCP tester input: %s %+v", name, server)
		}
		return []mcp.RemoteTool{
			{ServerName: name, Tool: mcpproto.Tool{Name: "search_docs", Description: "Search local docs"}},
		}, nil
	}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body mcpTestResponse
	postJSON(t, ts.URL+"/mcp/test", `{"name":"local"}`, http.StatusOK, &body)
	if body.Status != "ok" || body.ToolCount != 1 {
		t.Fatalf("unexpected MCP test response: %+v", body)
	}
	if len(body.Tools) != 1 || body.Tools[0].Name != "search_docs" || body.Tools[0].Description != "Search local docs" {
		t.Fatalf("unexpected MCP tools: %+v", body.Tools)
	}
}

func TestHandleFileIntakeDocumentText(t *testing.T) {
	dir := makeDaemonWorkspaceFixtureDir(t)
	docPath := filepath.Join(dir, "report.docx")
	writeDaemonZipFixture(t, docPath, map[string]string{
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Astria intake report</w:t></w:r></w:p></w:body></w:document>`,
	})
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body fileIntakeResponse
	postJSON(t, ts.URL+"/intake/file", `{"path":`+strconv.Quote(docPath)+`,"mode":"document_text","max_chars":2000}`, http.StatusOK, &body)
	if body.Mode != "document_text" || body.Status != "ok" || body.IsError {
		t.Fatalf("unexpected intake response: %+v", body)
	}
	if !strings.Contains(body.Content, "Astria intake report") {
		t.Fatalf("document intake content missing fixture text: %s", body.Content)
	}
}

func TestHandleFileIntakeArchiveInspectAuto(t *testing.T) {
	dir := makeDaemonWorkspaceFixtureDir(t)
	archivePath := filepath.Join(dir, "bundle.zip")
	writeDaemonZipFixture(t, archivePath, map[string]string{
		"README.md": "Astria archive intake",
		"notes.txt": "second file",
	})
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body fileIntakeResponse
	postJSON(t, ts.URL+"/intake/file", `{"path":`+strconv.Quote(archivePath)+`,"mode":"auto","max_entries":10}`, http.StatusOK, &body)
	if body.Mode != "archive_inspect" || body.Status != "ok" || body.IsError {
		t.Fatalf("unexpected intake response: %+v", body)
	}
	if !strings.Contains(body.Content, "README.md") || !strings.Contains(body.Content, `"format": "zip"`) {
		t.Fatalf("archive intake content missing expected entries: %s", body.Content)
	}
}

func TestHandleFileIntakeRejectsInvalidMode(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/intake/file", strings.NewReader(`{"path":"README.md","mode":"archive_extract"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /intake/file: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleFileIntakeToolErrorVisible(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body fileIntakeResponse
	postJSON(t, ts.URL+"/intake/file", `{"path":"testdata/missing-astria-doc.docx","mode":"document_text"}`, http.StatusOK, &body)
	if !body.IsError || body.Status != "error" || !strings.Contains(body.Content, "file not found") {
		t.Fatalf("expected visible tool error, got %+v", body)
	}
}

func makeDaemonWorkspaceFixtureDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("testdata", "tmp-"+strings.ReplaceAll(t.Name(), "/", "-"))
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("remove fixture dir: %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

func writeDaemonZipFixture(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip fixture: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip fixture: %v", err)
	}
}

func TestHandleMemoryLifecycle(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var initial memoryView
	getJSON(t, ts.URL+"/memory", http.StatusOK, &initial)
	if initial.MemoryDir == "" {
		t.Fatalf("memory_dir should be set")
	}
	if len(initial.Entries) != 0 {
		t.Fatalf("initial entries = %+v, want empty", initial.Entries)
	}

	var appended memoryView
	postJSON(t, ts.URL+"/memory", `{"content":"- User prefers Astria UI to stay desktop-like."}`, http.StatusOK, &appended)
	if !strings.Contains(appended.Content, "desktop-like") {
		t.Fatalf("memory content missing appended entry: %q", appended.Content)
	}
	if len(appended.Entries) != 1 || !appended.Entries[0].Primary {
		t.Fatalf("unexpected entries after append: %+v", appended.Entries)
	}

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/memory/MEMORY.md", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE memory: %v", err)
	}
	var afterDelete memoryView
	decodeJSONResponse(t, resp, http.StatusOK, &afterDelete)
	if afterDelete.Content != "" || len(afterDelete.Entries) != 0 {
		t.Fatalf("memory should be empty after delete: %+v", afterDelete)
	}
}

func TestMemoryTaxonomyParsesCategoriesAndWarnings(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	content := strings.Join([]string{
		"## Decisions",
		"- Use embedded Web UI assets.",
		"- Use embedded Web UI assets.",
		"## Risks",
		"- Provider setup: GitHub secret is missing.",
		"- Provider setup: GitHub secret is configured.",
		"- [command] go test ./internal/daemon",
	}, "\n")
	var view memoryView
	postJSON(t, ts.URL+"/memory", `{"content":`+strconv.Quote(content)+`}`, http.StatusOK, &view)
	if view.Categories["decisions"] != 2 || view.Categories["risks"] != 2 || view.Categories["commands"] != 1 {
		t.Fatalf("unexpected categories: %+v", view.Categories)
	}
	if len(view.Facts) != 5 {
		t.Fatalf("expected 5 parsed facts, got %d: %+v", len(view.Facts), view.Facts)
	}
	var duplicate, conflict bool
	for _, warning := range view.Warnings {
		if warning.Type == "duplicate" {
			duplicate = true
		}
		if warning.Type == "conflict" && warning.Subject == "provider setup" {
			conflict = true
		}
	}
	if !duplicate || !conflict {
		t.Fatalf("expected duplicate and conflict warnings, got %+v", view.Warnings)
	}
}

func TestHandleMemoryRejectsTraversal(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/memory/..%2Fsecret", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE traversal: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestInboxWebhookApproveAndDeduplicate(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var created struct {
		Item      InboxItem `json:"item"`
		Duplicate bool      `json:"duplicate"`
	}
	postJSON(t, ts.URL+"/inbox/webhook", `{"external_id":"evt-1","sender":"alice","text":"summarize this","metadata":{"thread":"t1"}}`, http.StatusCreated, &created)
	if created.Duplicate {
		t.Fatal("first inbox event should not be duplicate")
	}
	if created.Item.ID == "" || created.Item.Status != "pending" || created.Item.Provider != "webhook" {
		t.Fatalf("unexpected created inbox item: %+v", created.Item)
	}
	if created.Item.Metadata["thread"] != "t1" {
		t.Fatalf("metadata not preserved: %+v", created.Item.Metadata)
	}

	var duplicate struct {
		Item      InboxItem `json:"item"`
		Duplicate bool      `json:"duplicate"`
	}
	postJSON(t, ts.URL+"/inbox/webhook", `{"external_id":"evt-1","sender":"alice","text":"summarize this again"}`, http.StatusOK, &duplicate)
	if !duplicate.Duplicate || duplicate.Item.ID != created.Item.ID || duplicate.Item.Text != "summarize this" {
		t.Fatalf("unexpected duplicate response: %+v", duplicate)
	}

	var list struct {
		Items []InboxItem `json:"items"`
	}
	getJSON(t, ts.URL+"/inbox", http.StatusOK, &list)
	if len(list.Items) != 1 || list.Items[0].ID != created.Item.ID {
		t.Fatalf("unexpected inbox list: %+v", list.Items)
	}

	var approved struct {
		Item InboxItem        `json:"item"`
		Run  RunAgentResponse `json:"run"`
	}
	postJSON(t, ts.URL+"/inbox/"+created.Item.ID+"/approve", `{}`, http.StatusOK, &approved)
	if approved.Item.Status != "completed" || approved.Item.RunID == "" {
		t.Fatalf("unexpected approved item: %+v", approved.Item)
	}
	if approved.Run.SessionID == "" || len(approved.Run.Messages) == 0 {
		t.Fatalf("expected run response after approve: %+v", approved.Run)
	}

	var run RunRecord
	getJSON(t, ts.URL+"/runs/"+approved.Item.RunID, http.StatusOK, &run)
	if run.Channel != ChannelInbox || run.Prompt != "summarize this" {
		t.Fatalf("unexpected inbox run record: %+v", run)
	}
	if run.Request.Sender != "alice" {
		t.Fatalf("run request sender = %q, want alice", run.Request.Sender)
	}
}

func TestInboxRejectAndRetryValidation(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var created struct {
		Item InboxItem `json:"item"`
	}
	postJSON(t, ts.URL+"/inbox/webhook", `{"external_id":"evt-2","text":"reject me"}`, http.StatusCreated, &created)

	var rejected InboxItem
	postJSON(t, ts.URL+"/inbox/"+created.Item.ID+"/reject", `{}`, http.StatusOK, &rejected)
	if rejected.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", rejected.Status)
	}
	postJSON(t, ts.URL+"/inbox/"+created.Item.ID+"/approve", `{}`, http.StatusBadRequest, &map[string]string{})
	postJSON(t, ts.URL+"/inbox/"+created.Item.ID+"/retry", `{}`, http.StatusBadRequest, &map[string]string{})
}

func TestInboxGitHubIssueWebhook(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{
		"action":"opened",
		"repository":{"full_name":"acme/roadmap","html_url":"https://github.com/acme/roadmap"},
		"sender":{"login":"octo"},
		"issue":{"id":42,"number":7,"title":"Plan Astria channels","body":"Please wire this into Astria.","html_url":"https://github.com/acme/roadmap/issues/7","user":{"login":"alice"}}
	}`
	var created struct {
		Item      InboxItem `json:"item"`
		Duplicate bool      `json:"duplicate"`
	}
	postGitHubJSON(t, ts.URL+"/inbox/github", "issues", "delivery-1", body, "", http.StatusCreated, &created)
	if created.Duplicate || created.Item.Provider != "github" || created.Item.Status != "pending" {
		t.Fatalf("unexpected github inbox item: %+v", created)
	}
	if created.Item.ExternalID != "issue:acme/roadmap:42:opened" {
		t.Fatalf("external id = %q", created.Item.ExternalID)
	}
	if created.Item.Sender != "alice" || !strings.Contains(created.Item.Text, "Plan Astria channels") {
		t.Fatalf("unexpected sender/text: %+v", created.Item)
	}
	if created.Item.Metadata["delivery"] != "delivery-1" || created.Item.Metadata["html_url"] == "" {
		t.Fatalf("metadata not preserved: %+v", created.Item.Metadata)
	}

	var duplicate struct {
		Item      InboxItem `json:"item"`
		Duplicate bool      `json:"duplicate"`
	}
	postGitHubJSON(t, ts.URL+"/inbox/github", "issues", "delivery-2", body, "", http.StatusOK, &duplicate)
	if !duplicate.Duplicate || duplicate.Item.ID != created.Item.ID {
		t.Fatalf("expected duplicate GitHub item, got %+v", duplicate)
	}
}

func TestInboxGitHubIssueCommentWebhook(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{
		"action":"created",
		"repository":{"full_name":"acme/roadmap","html_url":"https://github.com/acme/roadmap"},
		"sender":{"login":"octo"},
		"issue":{"id":42,"number":7,"title":"Plan Astria channels"},
		"comment":{"id":99,"body":"Can Astria take this?","html_url":"https://github.com/acme/roadmap/issues/7#issuecomment-99","user":{"login":"bob"}}
	}`
	var created struct {
		Item InboxItem `json:"item"`
	}
	postGitHubJSON(t, ts.URL+"/inbox/github", "issue_comment", "delivery-comment", body, "", http.StatusCreated, &created)
	if created.Item.ExternalID != "issue_comment:acme/roadmap:99:created" {
		t.Fatalf("external id = %q", created.Item.ExternalID)
	}
	if created.Item.Sender != "bob" || !strings.Contains(created.Item.Text, "Can Astria take this?") {
		t.Fatalf("unexpected comment item: %+v", created.Item)
	}
}

func TestInboxProvidersExposeGitHubSetup(t *testing.T) {
	t.Setenv("STARCLAW_GITHUB_WEBHOOK_SECRET", "secret")
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body struct {
		Providers []inboxProviderView `json:"providers"`
	}
	getJSON(t, ts.URL+"/inbox/providers", http.StatusOK, &body)
	var github inboxProviderView
	for _, provider := range body.Providers {
		if provider.Kind == "github" {
			github = provider
		}
	}
	if github.Endpoint != "/inbox/github" || !github.SecretConfigured {
		t.Fatalf("unexpected github provider view: %+v", github)
	}
}

func TestInboxGitHubSignatureVerification(t *testing.T) {
	t.Setenv("STARCLAW_GITHUB_WEBHOOK_SECRET", "top-secret")
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"action":"opened","repository":{"full_name":"acme/roadmap"},"issue":{"id":42,"number":7,"title":"Signed"}}`
	var failed map[string]string
	postGitHubJSON(t, ts.URL+"/inbox/github", "issues", "delivery-bad", body, "sha256=bad", http.StatusUnauthorized, &failed)

	var created struct {
		Item InboxItem `json:"item"`
	}
	postGitHubJSON(t, ts.URL+"/inbox/github", "issues", "delivery-good", body, githubSignature("top-secret", body), http.StatusCreated, &created)
	if created.Item.ID == "" || created.Item.Provider != "github" {
		t.Fatalf("unexpected signed response: %+v", created)
	}
}

func TestHandleInstructions(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// GET instructions (file does not exist)
	resp, err := http.Get(ts.URL + "/instructions")
	if err != nil {
		t.Fatalf("GET /instructions: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// PUT instructions
	putBody := `{"content":"You are a test assistant."}`
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/instructions", strings.NewReader(putBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /instructions: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify instructions were written.
	resp, err = http.Get(ts.URL + "/instructions")
	if err != nil {
		t.Fatalf("GET /instructions after PUT: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode instructions: %v", err)
	}
	if body.Content != "You are a test assistant." {
		t.Errorf("expected 'You are a test assistant.', got %q", body.Content)
	}
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func TestHandleSessions(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sessions")
	if err != nil {
		t.Fatalf("GET /sessions: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Sessions []interface{} `json:"sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if body.Sessions == nil {
		t.Error("expected empty sessions list, not nil")
	}
}

func TestHandlePatchSession(t *testing.T) {
	root := t.TempDir()
	mgr := session.NewManager(filepath.Join(root, "sessions"))
	sess := mgr.NewSession()
	sess.Title = "Original title"
	if err := mgr.Save(); err != nil {
		t.Fatalf("save session: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = root
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"title":"Renamed session","favorite":true}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/sessions/"+sess.ID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /sessions/{id}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var updated session.Session
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if updated.Title != "Renamed session" || !updated.Favorite {
		t.Fatalf("unexpected updated session: %+v", updated)
	}
	reloaded, err := mgr.Resume(sess.ID)
	if err != nil {
		t.Fatalf("resume patched session: %v", err)
	}
	if reloaded.Title != "Renamed session" || !reloaded.Favorite {
		t.Fatalf("session not persisted: %+v", reloaded)
	}
}

func TestHandlePatchSessionRejectsEmptyTitle(t *testing.T) {
	root := t.TempDir()
	mgr := session.NewManager(filepath.Join(root, "sessions"))
	sess := mgr.NewSession()
	if err := mgr.Save(); err != nil {
		t.Fatalf("save session: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = root
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/sessions/"+sess.ID, strings.NewReader(`{"title":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /sessions/{id}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Permissions
// ---------------------------------------------------------------------------

func TestHandlePermissions(t *testing.T) {
	deps := newTestServerDeps(t)
	deps.Config = &config.Config{Permissions: &permissions.Config{
		AllowedDirs:       []string{"~", "."},
		AllowedCommands:   []string{"go test"},
		DeniedCommands:    []string{"shutdown"},
		NetworkAllowlist:  []string{"api.github.com"},
		SensitivePatterns: []string{"*.secret"},
	}}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/permissions")
	if err != nil {
		t.Fatalf("GET /permissions: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var body struct {
		Permissions struct {
			Configured        bool     `json:"configured"`
			AllowedDirs       []string `json:"allowed_dirs"`
			AllowedCommands   []string `json:"allowed_commands"`
			DeniedCommands    []string `json:"denied_commands"`
			NetworkAllowlist  []string `json:"network_allowlist"`
			SensitivePatterns []string `json:"sensitive_patterns"`
		} `json:"permissions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode permissions: %v", err)
	}
	if !body.Permissions.Configured {
		t.Fatal("expected configured permissions")
	}
	if len(body.Permissions.AllowedDirs) != 2 || body.Permissions.AllowedDirs[0] != "~" {
		t.Fatalf("unexpected allowed dirs: %+v", body.Permissions.AllowedDirs)
	}
}

func TestHandleConfigPatchUpdatesPermissions(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
provider: ollama
ollama_model: smoke
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{Provider: "ollama"}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	patchBody := `{"permissions":{"allowed_dirs":[" ~ ","."],"allowed_commands":["go test",""],"denied_commands":["shutdown"],"network_allowlist":["api.github.com"],"sensitive_patterns":["*.secret"]}}`
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := yaml.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved config is not valid YAML: %v\n%s", err, data)
	}
	if saved.Permissions == nil {
		t.Fatal("expected saved permissions")
	}
	if strings.Join(saved.Permissions.AllowedDirs, ",") != "~,." {
		t.Fatalf("allowed dirs = %+v", saved.Permissions.AllowedDirs)
	}
	if strings.Join(saved.Permissions.AllowedCommands, ",") != "go test" {
		t.Fatalf("allowed commands = %+v", saved.Permissions.AllowedCommands)
	}
	if deps.Config.Permissions == nil || strings.Join(deps.Config.Permissions.NetworkAllowlist, ",") != "api.github.com" {
		t.Fatalf("in-memory permissions not refreshed: %+v", deps.Config.Permissions)
	}

	var body struct {
		Permissions struct {
			Configured        bool     `json:"configured"`
			AllowedDirs       []string `json:"allowed_dirs"`
			AllowedCommands   []string `json:"allowed_commands"`
			DeniedCommands    []string `json:"denied_commands"`
			NetworkAllowlist  []string `json:"network_allowlist"`
			SensitivePatterns []string `json:"sensitive_patterns"`
		} `json:"permissions"`
	}
	getJSON(t, ts.URL+"/permissions", http.StatusOK, &body)
	if !body.Permissions.Configured {
		t.Fatal("expected configured permissions after patch")
	}
	if strings.Join(body.Permissions.AllowedCommands, ",") != "go test" {
		t.Fatalf("GET permissions allowed commands = %+v", body.Permissions.AllowedCommands)
	}
}

func TestHandleConfigPatchClearsPermissions(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`
provider: ollama
permissions:
  allowed_dirs:
    - "~"
  denied_commands:
    - shutdown
`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := newTestServerDeps(t)
	deps.StarclawDir = dir
	deps.ConfigPath = configPath
	deps.Config = &config.Config{Provider: "ollama", Permissions: &permissions.Config{AllowedDirs: []string{"~"}}}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/config", strings.NewReader(`{"permissions":{}}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /config: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := yaml.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved config is not valid YAML: %v\n%s", err, data)
	}
	if saved.Permissions != nil {
		t.Fatalf("expected permissions to be cleared, got %+v", saved.Permissions)
	}
	if deps.Config.Permissions != nil {
		t.Fatalf("expected in-memory permissions to be cleared, got %+v", deps.Config.Permissions)
	}

	var body struct {
		Permissions struct {
			Configured bool `json:"configured"`
		} `json:"permissions"`
	}
	getJSON(t, ts.URL+"/permissions", http.StatusOK, &body)
	if body.Permissions.Configured {
		t.Fatal("expected permissions to be unconfigured after clear")
	}
}

// ---------------------------------------------------------------------------
// Approval
// ---------------------------------------------------------------------------

func TestHandleApproval(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))

	// Register a pending approval.
	apr := ApprovalRequest{
		RequestID: "test-approve-1",
		Tool:      "bash",
		Args:      "echo hello",
	}

	var decision ApprovalDecision
	var waitErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		decision, waitErr = s.approvalBroker.WaitForApproval(context.Background(), apr)
	}()

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Resolve via API.
	body := `{"request_id":"test-approve-1","decision":"allow"}`
	resp, postErr := http.Post(ts.URL+"/approval", "application/json", strings.NewReader(body))
	if postErr != nil {
		t.Fatalf("POST /approval: %v", postErr)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	<-done

	if waitErr != nil {
		t.Fatalf("WaitForApproval returned error: %v", waitErr)
	}
	if decision != DecisionAllow {
		t.Errorf("expected DecisionAllow, got %q", decision)
	}
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

func TestHandleCancel(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))

	// Register a fake running request.
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	s.running.Store("test-request-1", cancel)

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"request_id":"test-request-1"}`
	resp, err := http.Post(ts.URL+"/cancel", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /cancel: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleCancelNotFound(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"request_id":"nonexistent"}`
	resp, err := http.Post(ts.URL+"/cancel", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /cancel: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleCancelMissingRequestID(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{}`
	resp, err := http.Post(ts.URL+"/cancel", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /cancel: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleCreateAgent(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"name":"test-agent","prompt":"You are a test agent.","memory":"Remember this.","model":"gpt-test","reasoning_effort":"low","tools_allow":["file_read","grep"],"tools_deny":["bash"],"auto_approve":true,"heartbeat_every":"15m","heartbeat_active_hours":"09:00-17:00","heartbeat_model":"gpt-heartbeat","commands":{"review":"Review recent changes.","deploy":"Deploy safely."}}`
	resp, err := http.Post(ts.URL+"/agents", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /agents: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var created agents.Agent
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created agent: %v", err)
	}
	if created.Name != "test-agent" || !strings.Contains(created.Prompt, "test agent") {
		t.Fatalf("unexpected created agent: %+v", created)
	}
	if created.Config == nil || created.Config.Agent == nil || created.Config.Agent.Model == nil || *created.Config.Agent.Model != "gpt-test" {
		t.Fatalf("agent model config not persisted: %+v", created.Config)
	}
	if created.Config.Tools == nil || len(created.Config.Tools.Allow) != 2 || created.Config.Tools.Deny[0] != "bash" {
		t.Fatalf("agent tools config not persisted: %+v", created.Config.Tools)
	}
	if created.Config.AutoApprove == nil || !*created.Config.AutoApprove {
		t.Fatalf("auto_approve not persisted: %+v", created.Config)
	}
	if created.Config.Heartbeat == nil || created.Config.Heartbeat.Every != "15m" || created.Config.Heartbeat.ActiveHours != "09:00-17:00" || created.Config.Heartbeat.Model != "gpt-heartbeat" {
		t.Fatalf("heartbeat config not persisted: %+v", created.Config.Heartbeat)
	}
	if len(created.Commands) != 2 || !strings.Contains(created.Commands["review"], "Review recent changes") {
		t.Fatalf("commands not persisted: %+v", created.Commands)
	}
	var list struct {
		Agents []agents.AgentInfo `json:"agents"`
	}
	getJSON(t, ts.URL+"/agents", http.StatusOK, &list)
	if len(list.Agents) != 1 || list.Agents[0].Name != "test-agent" {
		t.Fatalf("agent not listed: %+v", list.Agents)
	}
	info := list.Agents[0]
	if info.Model != "gpt-test" || info.ReasoningEffort != "low" || !info.AutoApprove || info.HeartbeatEvery != "15m" || info.HeartbeatHours != "09:00-17:00" || info.HeartbeatModel != "gpt-heartbeat" || info.CommandCount != 2 || !info.HasMemory {
		t.Fatalf("agent capability summary not listed: %+v", info)
	}
	if len(info.CommandNames) != 2 || info.CommandNames[0] != "deploy" || info.CommandNames[1] != "review" {
		t.Fatalf("agent command names not listed: %+v", info.CommandNames)
	}
	if len(info.ToolsAllow) != 2 || info.ToolsAllow[0] != "file_read" || len(info.ToolsDeny) != 1 || info.ToolsDeny[0] != "bash" {
		t.Fatalf("agent tool summary not listed: %+v", info)
	}
}

func TestHandleUpdateAgent(t *testing.T) {
	deps := newTestServerDeps(t)
	writeTestAgent(t, deps.AgentsDir, "test-agent")
	writeTestAgentCommand(t, deps.AgentsDir, "test-agent", "old", "Old command.")
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/agents/test-agent", strings.NewReader(`{"prompt":"Updated prompt","model":"gpt-updated","tools_allow":["version","file_read"],"tools_deny":["bash","http"],"auto_approve":false,"heartbeat_every":"30m","heartbeat_active_hours":"10:00-18:00","heartbeat_model":"gpt-heartbeat-updated","commands":{"review":"Updated review command."}}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /agents/{name}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated agents.Agent
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated agent: %v", err)
	}
	if updated.Prompt != "Updated prompt\n" {
		t.Fatalf("prompt = %q", updated.Prompt)
	}
	if updated.Config == nil || updated.Config.Agent == nil || updated.Config.Agent.Model == nil || *updated.Config.Agent.Model != "gpt-updated" {
		t.Fatalf("model config not updated: %+v", updated.Config)
	}
	if updated.Config.Tools == nil || len(updated.Config.Tools.Allow) != 2 || updated.Config.Tools.Allow[0] != "version" || len(updated.Config.Tools.Deny) != 2 || updated.Config.Tools.Deny[1] != "http" {
		t.Fatalf("tools config not updated: %+v", updated.Config.Tools)
	}
	if updated.Config.AutoApprove == nil || *updated.Config.AutoApprove {
		t.Fatalf("auto_approve not updated: %+v", updated.Config)
	}
	if updated.Config.Heartbeat == nil || updated.Config.Heartbeat.Every != "30m" || updated.Config.Heartbeat.ActiveHours != "10:00-18:00" || updated.Config.Heartbeat.Model != "gpt-heartbeat-updated" {
		t.Fatalf("heartbeat config not updated: %+v", updated.Config.Heartbeat)
	}
	if len(updated.Commands) != 1 || !strings.Contains(updated.Commands["review"], "Updated review command") {
		t.Fatalf("commands not updated: %+v", updated.Commands)
	}
	if _, ok := updated.Commands["old"]; ok {
		t.Fatalf("deleted command still present: %+v", updated.Commands)
	}
}

func TestHandleUpdateAgentClearsHeartbeat(t *testing.T) {
	deps := newTestServerDeps(t)
	writeTestAgent(t, deps.AgentsDir, "test-agent")
	agentDir := filepath.Join(deps.AgentsDir, "test-agent")
	if err := os.WriteFile(filepath.Join(agentDir, "config.yaml"), []byte("heartbeat:\n  every: 15m\n  active_hours: 09:00-17:00\n"), 0600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/agents/test-agent", strings.NewReader(`{"prompt":"Updated prompt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /agents/{name}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated agents.Agent
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated agent: %v", err)
	}
	if updated.Config != nil && updated.Config.Heartbeat != nil {
		t.Fatalf("heartbeat config should be cleared: %+v", updated.Config.Heartbeat)
	}
}

func TestHandleUpdateAgentPreservesCommandsWhenOmitted(t *testing.T) {
	deps := newTestServerDeps(t)
	writeTestAgent(t, deps.AgentsDir, "test-agent")
	writeTestAgentCommand(t, deps.AgentsDir, "test-agent", "review", "Review command.")
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/agents/test-agent", strings.NewReader(`{"prompt":"Updated prompt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /agents/{name}: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated agents.Agent
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated agent: %v", err)
	}
	if len(updated.Commands) != 1 || !strings.Contains(updated.Commands["review"], "Review command") {
		t.Fatalf("commands should be preserved when omitted: %+v", updated.Commands)
	}
}

func TestHandleCreateAgentRejectsInvalidCommandName(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/agents", "application/json", strings.NewReader(`{"name":"test-agent","prompt":"Prompt","commands":{"../escape":"Nope"}}`))
	if err != nil {
		t.Fatalf("POST /agents: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleCreateAgentRejectsInvalidName(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/agents", "application/json", strings.NewReader(`{"name":"BadName","prompt":"Prompt"}`))
	if err != nil {
		t.Fatalf("POST /agents: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Shutdown
// ---------------------------------------------------------------------------

func TestHandleShutdown(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/shutdown", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /shutdown: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func writeTestAgent(t *testing.T, agentsDir, name string) {
	t.Helper()
	agentDir := filepath.Join(agentsDir, name)
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatalf("create agent dir: %v", err)
	}
	content := "# API Agent\n\nAgent available through daemon API."
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write AGENT.md: %v", err)
	}
}

func writeTestAgentCommand(t *testing.T, agentsDir, agentName, commandName, content string) {
	t.Helper()
	commandsDir := filepath.Join(agentsDir, agentName, "commands")
	if err := os.MkdirAll(commandsDir, 0700); err != nil {
		t.Fatalf("create commands dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, commandName+".md"), []byte(content), 0644); err != nil {
		t.Fatalf("write command: %v", err)
	}
}

func writeTestSkill(t *testing.T, skillsDir, name, description string) {
	t.Helper()
	skillDir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(skillDir, 0700); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	content := `---
name: ` + name + `
description: ` + description + `
---

Use this skill in daemon API tests.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

func getJSON(t *testing.T, url string, wantStatus int, out interface{}) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	decodeJSONResponse(t, resp, wantStatus, out)
}

func postJSON(t *testing.T, url string, body string, wantStatus int, out interface{}) {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	decodeJSONResponse(t, resp, wantStatus, out)
}

func postGitHubJSON(t *testing.T, url, event, delivery, body, signature string, wantStatus int, out interface{}) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new GitHub POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", event)
	req.Header.Set("X-GitHub-Delivery", delivery)
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	decodeJSONResponse(t, resp, wantStatus, out)
}

func githubSignature(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func patchJSON(t *testing.T, url string, body string, wantStatus int, out interface{}) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new PATCH request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s: %v", url, err)
	}
	decodeJSONResponse(t, resp, wantStatus, out)
}

func deleteJSON(t *testing.T, url string, wantStatus int) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("new DELETE request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", url, err)
	}
	decodeJSONResponse(t, resp, wantStatus, &map[string]string{})
}

func decodeSSEEventData(t *testing.T, stream string, eventName string) map[string]interface{} {
	t.Helper()
	events := decodeSSEEvents(t, stream, eventName)
	if len(events) == 0 {
		t.Fatalf("SSE event %q not found in stream:\n%s", eventName, stream)
	}
	return events[0]
}

func decodeSSEEvents(t *testing.T, stream string, eventName string) []map[string]interface{} {
	t.Helper()
	var events []map[string]interface{}
	blocks := strings.Split(stream, "\n\n")
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		if len(lines) < 2 || strings.TrimSpace(lines[0]) != "event: "+eventName {
			continue
		}
		dataLine := strings.TrimPrefix(lines[1], "data: ")
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataLine), &data); err != nil {
			t.Fatalf("decode SSE %s data: %v", eventName, err)
		}
		events = append(events, data)
	}
	return events
}

func decodeJSONResponse(t *testing.T, resp *http.Response, wantStatus int, out interface{}) {
	t.Helper()
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != wantStatus {
		t.Fatalf("expected HTTP %d, got %d", wantStatus, resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}
