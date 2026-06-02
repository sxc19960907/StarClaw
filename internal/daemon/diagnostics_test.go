package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/permissions"
	"github.com/starclaw/starclaw/internal/schedule"
)

type diagnosticTestTool struct{}

func (diagnosticTestTool) Info() agent.ToolInfo {
	return agent.ToolInfo{Name: "diagnostic_test", Description: "diagnostic test tool"}
}

func (diagnosticTestTool) Run(_ context.Context, _ string) (agent.ToolResult, error) {
	return agent.ToolResult{}, nil
}

func (diagnosticTestTool) RequiresApproval() bool {
	return false
}

func TestHandleDiagnosticsNeedsSetup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("provider: anthropic\napi_key: \"\"\n"), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := &ServerDeps{
		StarclawDir:     dir,
		ConfigPath:      configPath,
		Config:          &config.Config{Provider: "anthropic", Endpoint: "https://api.anthropic.com", ModelTier: "medium"},
		Registry:        agent.NewToolRegistry(),
		ScheduleManager: schedule.NewManager(filepath.Join(dir, "schedules.json")),
	}
	deps.Registry.Register(diagnosticTestTool{})
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body DiagnosticsResponse
	getJSON(t, ts.URL+"/diagnostics", http.StatusOK, &body)
	if body.Status != DiagnosticNeedsSetup {
		t.Fatalf("status = %q, want %q; body = %+v", body.Status, DiagnosticNeedsSetup, body)
	}
	check := findDiagnosticCheck(body.Checks, "provider")
	if check == nil {
		t.Fatalf("provider check missing: %+v", body.Checks)
	}
	if check.Status != DiagnosticNeedsSetup {
		t.Fatalf("provider status = %q, want %q", check.Status, DiagnosticNeedsSetup)
	}
	if check.Action == "" {
		t.Fatal("provider action should be present")
	}
}

func TestHandleDiagnosticsReady(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("provider: anthropic\napi_key: test\n"), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	deps := &ServerDeps{
		StarclawDir: dir,
		ConfigPath:  configPath,
		AgentsDir:   filepath.Join(dir, "agents"),
		Config: &config.Config{
			Provider:    "anthropic",
			Endpoint:    "https://api.anthropic.com",
			APIKey:      "test",
			ModelTier:   "medium",
			Permissions: permissions.DefaultConfig(),
		},
		Registry:        agent.NewToolRegistry(),
		ScheduleManager: schedule.NewManager(filepath.Join(dir, "schedules.json")),
	}
	deps.Registry.Register(diagnosticTestTool{})
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body DiagnosticsResponse
	getJSON(t, ts.URL+"/diagnostics", http.StatusOK, &body)
	if body.Status != DiagnosticReady {
		t.Fatalf("status = %q, want %q; checks = %+v", body.Status, DiagnosticReady, body.Checks)
	}
	if body.WebURL != "http://127.0.0.1:0/app/" {
		t.Fatalf("web_url = %q, want port-specific URL", body.WebURL)
	}
	if body.LaunchCommand != "starclaw app" {
		t.Fatalf("launch_command = %q, want starclaw app", body.LaunchCommand)
	}
	if body.StarclawDir != dir {
		t.Fatalf("starclaw_dir = %q, want %q", body.StarclawDir, dir)
	}
	if body.ConfigPath != configPath {
		t.Fatalf("config_path = %q, want %q", body.ConfigPath, configPath)
	}
	if body.AgentsDir != filepath.Join(dir, "agents") {
		t.Fatalf("agents_dir = %q, want %q", body.AgentsDir, filepath.Join(dir, "agents"))
	}
	if body.SessionsDir != filepath.Join(dir, "sessions") {
		t.Fatalf("sessions_dir = %q, want %q", body.SessionsDir, filepath.Join(dir, "sessions"))
	}
	if body.ExecutablePath == "" {
		t.Fatal("executable_path should be present when executable resolution succeeds")
	}
	if len(body.Checks) < 7 {
		t.Fatalf("expected at least 7 checks, got %d", len(body.Checks))
	}
}

func TestHandleDiagnosticsRouteReturnsStructuredChecks(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	deps := &ServerDeps{
		StarclawDir:     dir,
		ConfigPath:      configPath,
		Config:          &config.Config{Provider: "openai", OpenAIEndpoint: "https://api.openai.com/v1", OpenAIModel: "gpt-4o"},
		Registry:        agent.NewToolRegistry(),
		ScheduleManager: schedule.NewManager(filepath.Join(dir, "schedules.json")),
	}
	s := newTestServer(t, deps)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/diagnostics")
	if err != nil {
		t.Fatalf("GET /diagnostics: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body DiagnosticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != DiagnosticNeedsSetup {
		t.Fatalf("status = %q, want %q", body.Status, DiagnosticNeedsSetup)
	}
	if findDiagnosticCheck(body.Checks, "config") == nil {
		t.Fatal("config check missing")
	}
	if findDiagnosticCheck(body.Checks, "tools") == nil {
		t.Fatal("tools check missing")
	}
	if body.LaunchCommand != "starclaw app" {
		t.Fatalf("launch_command = %q, want starclaw app", body.LaunchCommand)
	}
	if body.WebURL == "" {
		t.Fatal("web_url should be present")
	}
}

func findDiagnosticCheck(checks []DiagnosticCheck, id string) *DiagnosticCheck {
	for i := range checks {
		if checks[i].ID == id {
			return &checks[i]
		}
	}
	return nil
}
