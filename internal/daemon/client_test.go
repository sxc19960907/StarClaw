package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---- Health ----

func TestClientHealth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/health" {
			t.Errorf("path = %q, want /health", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
}

// ---- Status ----

func TestClientStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/status" {
			t.Errorf("path = %q, want /status", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"running_agents":2,"active_agents":2,"uptime":"5m3s","version":"0.1.0","desktop_rpc":{"listening":true,"connected":false,"pending":3}}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() returned error: %v", err)
	}
	if status.RunningAgents != 2 {
		t.Errorf("RunningAgents = %d, want 2", status.RunningAgents)
	}
	if status.Uptime != "5m3s" {
		t.Errorf("Uptime = %q, want %q", status.Uptime, "5m3s")
	}
	if status.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", status.Version, "0.1.0")
	}
	if !status.DesktopRPC.Listening {
		t.Errorf("DesktopRPC.Listening = false, want true")
	}
	if status.DesktopRPC.Connected {
		t.Errorf("DesktopRPC.Connected = true, want false")
	}
	if status.DesktopRPC.Pending != 3 {
		t.Errorf("DesktopRPC.Pending = %d, want 3", status.DesktopRPC.Pending)
	}
}

func TestClientStatus_Minimal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"running_agents":0,"uptime":"0s"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() returned error: %v", err)
	}
	if status.RunningAgents != 0 {
		t.Errorf("RunningAgents = %d, want 0", status.RunningAgents)
	}
	if status.Uptime != "0s" {
		t.Errorf("Uptime = %q, want %q", status.Uptime, "0s")
	}
	if status.Version != "" {
		t.Errorf("Version = %q, want empty", status.Version)
	}
}

// ---- Message ----

func TestClientMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/message" {
			t.Errorf("path = %q, want /message", r.URL.Path)
		}

		var req RunAgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Text != "hello" {
			t.Errorf("Text = %q, want %q", req.Text, "hello")
		}
		if req.Agent != "assistant" {
			t.Errorf("Agent = %q, want %q", req.Agent, "assistant")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"session_id":"sess_001","messages":["response text"],"usage":{"prompt_tokens":10}}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Message(context.Background(), RunAgentRequest{
		Text:  "hello",
		Agent: "assistant",
	})
	if err != nil {
		t.Fatalf("Message() returned error: %v", err)
	}
	if resp.SessionID != "sess_001" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "sess_001")
	}
	if len(resp.Messages) != 1 || resp.Messages[0] != "response text" {
		t.Errorf("Messages = %v, want [response text]", resp.Messages)
	}
	if resp.Usage["prompt_tokens"] != 10 {
		t.Errorf("prompt_tokens = %d, want 10", resp.Usage["prompt_tokens"])
	}
}

// ---- Cancel ----

func TestClientCancel_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/cancel" {
			t.Errorf("path = %q, want /cancel", r.URL.Path)
		}

		var req CancelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.SessionID != "sess_001" {
			t.Errorf("SessionID = %q, want %q", req.SessionID, "sess_001")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Cancel(context.Background(), CancelRequest{SessionID: "sess_001"}); err != nil {
		t.Fatalf("Cancel() returned error: %v", err)
	}
}

// ---- Shutdown ----

func TestClientShutdown_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/shutdown" {
			t.Errorf("path = %q, want /shutdown", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() returned error: %v", err)
	}
}

// ---- Schedules ----

func TestClientListSchedules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/schedules" {
			t.Errorf("path = %q, want /schedules", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `[{"id":"s1","agent":"daily-agent","cron":"0 9 * * *","prompt":"run daily","enabled":true,"sync_status":"ok","created_at":"2025-01-01T00:00:00Z"}]`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	schedules, err := client.ListSchedules(context.Background())
	if err != nil {
		t.Fatalf("ListSchedules() returned error: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("got %d schedules, want 1", len(schedules))
	}
	if schedules[0].ID != "s1" {
		t.Errorf("ID = %q, want %q", schedules[0].ID, "s1")
	}
	if schedules[0].Agent != "daily-agent" {
		t.Errorf("Agent = %q, want %q", schedules[0].Agent, "daily-agent")
	}
	if schedules[0].Cron != "0 9 * * *" {
		t.Errorf("Cron = %q, want %q", schedules[0].Cron, "0 9 * * *")
	}
}

func TestClientListSchedules_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `[]`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	schedules, err := client.ListSchedules(context.Background())
	if err != nil {
		t.Fatalf("ListSchedules() returned error: %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("got %d schedules, want 0", len(schedules))
	}
}

func TestClientGetSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/schedules/s1" {
			t.Errorf("path = %q, want /schedules/s1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"id":"s1","agent":"daily-agent","cron":"0 9 * * *","prompt":"run daily","enabled":true,"sync_status":"ok","created_at":"2025-01-01T00:00:00Z"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	s, err := client.GetSchedule(context.Background(), "s1")
	if err != nil {
		t.Fatalf("GetSchedule() returned error: %v", err)
	}
	if s.ID != "s1" {
		t.Errorf("ID = %q, want %q", s.ID, "s1")
	}
	if s.Agent != "daily-agent" {
		t.Errorf("Agent = %q, want %q", s.Agent, "daily-agent")
	}
}

func TestClientCreateSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/schedules" {
			t.Errorf("path = %q, want /schedules", r.URL.Path)
		}

		var req CreateScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Agent != "my-agent" {
			t.Errorf("Agent = %q, want %q", req.Agent, "my-agent")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, `{"id":"schedule-abc123"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	id, err := client.CreateSchedule(context.Background(), CreateScheduleRequest{
		Agent:  "my-agent",
		Cron:   "0 9 * * *",
		Prompt: "run daily report",
	})
	if err != nil {
		t.Fatalf("CreateSchedule() returned error: %v", err)
	}
	if id != "schedule-abc123" {
		t.Errorf("id = %q, want %q", id, "schedule-abc123")
	}
}

func TestClientPatchSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/schedules/s1" {
			t.Errorf("path = %q, want /schedules/s1", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	enabled := false
	client := NewClient(server.URL)
	err := client.PatchSchedule(context.Background(), "s1", PatchScheduleRequest{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PatchSchedule() returned error: %v", err)
	}
}

func TestClientDeleteSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/schedules/s1" {
			t.Errorf("path = %q, want /schedules/s1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.DeleteSchedule(context.Background(), "s1"); err != nil {
		t.Fatalf("DeleteSchedule() returned error: %v", err)
	}
}

// ---- Error responses ----

func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"error":"invalid request"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `unexpected status 400: {"error":"invalid request"}` {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestClient_ErrorResponseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `not found`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `unexpected status 404: not found` {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestClient_ErrorResponseInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `internal error`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `unexpected status 500: internal error` {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// ---- Connection refused ----

func TestClient_ConnectionRefused(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Close() // close immediately so connection is refused

	client := NewClient(server.URL)
	err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---- JSON round-trip marshaling ----

func TestStatusResponseJSON(t *testing.T) {
	s := StatusResponse{
		RunningAgents: 3,
		Uptime:        "1h",
		Version:       "1.0.0",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded StatusResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.RunningAgents != 3 {
		t.Errorf("RunningAgents = %d, want 3", decoded.RunningAgents)
	}
	if decoded.Uptime != "1h" {
		t.Errorf("Uptime = %q, want %q", decoded.Uptime, "1h")
	}
	if decoded.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", decoded.Version, "1.0.0")
	}
}

func TestStatusResponseJSON_OmitEmpty(t *testing.T) {
	s := StatusResponse{
		RunningAgents: 0,
		Uptime:        "0s",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Version should be omitted since it's empty
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["version"]; ok {
		t.Error("expected version to be omitted")
	}
}

func TestCancelRequestJSON(t *testing.T) {
	req := CancelRequest{
		SessionID: "sess_001",
		RequestID: "req_001",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CancelRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.SessionID != "sess_001" {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, "sess_001")
	}
	if decoded.RequestID != "req_001" {
		t.Errorf("RequestID = %q, want %q", decoded.RequestID, "req_001")
	}
}

func TestCreateScheduleRequestJSON(t *testing.T) {
	req := CreateScheduleRequest{
		Agent:  "my-agent",
		Cron:   "0 9 * * *",
		Prompt: "daily report",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CreateScheduleRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Agent != "my-agent" {
		t.Errorf("Agent = %q, want %q", decoded.Agent, "my-agent")
	}
	if decoded.Cron != "0 9 * * *" {
		t.Errorf("Cron = %q, want %q", decoded.Cron, "0 9 * * *")
	}
	if decoded.Prompt != "daily report" {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, "daily report")
	}
}

func TestPatchScheduleRequestJSON(t *testing.T) {
	enabled := false
	req := PatchScheduleRequest{
		Cron:    nil,
		Prompt:  strPtr("new prompt"),
		Enabled: &enabled,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded PatchScheduleRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Cron != nil {
		t.Error("expected Cron to be nil")
	}
	if decoded.Prompt == nil || *decoded.Prompt != "new prompt" {
		t.Errorf("Prompt = %v, want %q", decoded.Prompt, "new prompt")
	}
	if decoded.Enabled == nil || *decoded.Enabled != false {
		t.Errorf("Enabled = %v, want false", decoded.Enabled)
	}
}

func TestPatchScheduleRequestJSON_OmitUnset(t *testing.T) {
	req := PatchScheduleRequest{}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if string(data) != "{}" {
		t.Errorf("expected {}, got %s", string(data))
	}
}

// ---- Default timeout ----

func TestClient_DefaultTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// This should succeed because the server is fast enough with 30s default timeout
	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
}

// ---- Helper ----

func strPtr(s string) *string {
	return &s
}
