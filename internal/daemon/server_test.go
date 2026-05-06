package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/schedule"
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
		StarclawDir:      t.TempDir(),
		ConfigPath:       filepath.Join(t.TempDir(), "config.json"),
		AgentsDir:        t.TempDir(),
		InstructionsDir:  t.TempDir(),
		LLMClient:        &mockLLMClient{t: t},
		Registry:         agent.NewToolRegistry(),
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify deletion
	resp, err = http.Get(ts.URL + "/schedules/" + scheduleID)
	if err != nil {
		t.Fatalf("GET deleted: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", resp.StatusCode)
	}
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify instructions were written.
	resp, err = http.Get(ts.URL + "/instructions")
	if err != nil {
		t.Fatalf("GET /instructions after PUT: %v", err)
	}
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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

// ---------------------------------------------------------------------------
// Permissions
// ---------------------------------------------------------------------------

func TestHandlePermissions(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/permissions")
	if err != nil {
		t.Fatalf("GET /permissions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Not Implemented (501)
// ---------------------------------------------------------------------------

func TestHandleCreateAgentNotImplemented(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := `{"name":"test","prompt":"You are a test."}`
	resp, err := http.Post(ts.URL+"/agents", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /agents: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", resp.StatusCode)
	}
}

func TestHandleUpdateAgentNotImplemented(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/agents/test-agent", strings.NewReader(`{"prompt":"updated"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /agents/{name}: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", resp.StatusCode)
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
