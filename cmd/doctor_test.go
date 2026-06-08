package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDoctorCmdPrintsUnavailableDaemonReadiness(t *testing.T) {
	restore := stubDoctorDaemon(t)
	restore.isHealthy = func(ctx context.Context) bool {
		return false
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(doctorCmd)

	output, err := executeCommand(root, "doctor")
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	for _, want := range []string{
		"StarClaw doctor",
		"Version:",
		"Launch:        starclaw app",
		"Web UI:        " + daemonWebURL,
		"Diagnostics:   " + daemonDiagnosticsURL,
		"Data:          " + home + "/.starclaw",
		"Config:        " + home + "/.starclaw/config.yaml",
		"Local checks:",
		"Daemon:        not running",
		"Next steps:    run `starclaw app`",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestDoctorCmdPrintsUnavailableDaemonJSON(t *testing.T) {
	restore := stubDoctorDaemon(t)
	restore.isHealthy = func(ctx context.Context) bool {
		return false
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(doctorCmd)

	output, err := executeCommand(root, "doctor", "--json")
	if err != nil {
		t.Fatalf("doctor --json failed: %v", err)
	}
	var report doctorReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("doctor --json output is not JSON: %v\n%s", err, output)
	}
	if report.LaunchCommand != "starclaw app" {
		t.Fatalf("launch_command = %q, want starclaw app", report.LaunchCommand)
	}
	if report.WebURL != daemonWebURL {
		t.Fatalf("web_url = %q, want %q", report.WebURL, daemonWebURL)
	}
	if report.DiagnosticsURL != daemonDiagnosticsURL {
		t.Fatalf("diagnostics_url = %q, want %q", report.DiagnosticsURL, daemonDiagnosticsURL)
	}
	if report.StarclawDir != home+"/.starclaw" {
		t.Fatalf("starclaw_dir = %q, want HOME .starclaw", report.StarclawDir)
	}
	if report.ConfigPath != home+"/.starclaw/config.yaml" {
		t.Fatalf("config_path = %q, want HOME config.yaml", report.ConfigPath)
	}
	if len(report.LocalChecks) == 0 {
		t.Fatal("local_checks should not be empty")
	}
	if report.Daemon.Running {
		t.Fatal("daemon.running should be false when daemon is unavailable")
	}
}

func TestDoctorCmdPrintsReachableDaemonReadiness(t *testing.T) {
	restore := stubDoctorDaemon(t)
	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"uptime":12,"version":"test-version","active_agents":2}`))
		case "/diagnostics":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ready","summary":"StarClaw is ready.","checks":[{"label":"Config","status":"ready","detail":"Config is available."}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	restore.statusURL = server.URL + "/status"
	restore.diagnosticsURL = server.URL + "/diagnostics"
	restore.applyURLs()

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(doctorCmd)

	output, err := executeCommand(root, "doctor")
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	for _, want := range []string{
		"Daemon:        running",
		"Daemon version: test-version",
		"Active agents: 2",
		"Uptime:        12s",
		"Runtime:       ready",
		"Summary:       StarClaw is ready.",
		"Daemon checks:",
		"[ready] Config: Config is available.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestDoctorCmdPrintsReachableDaemonJSON(t *testing.T) {
	restore := stubDoctorDaemon(t)
	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"uptime":30,"version":"json-version","active_agents":3}`))
		case "/diagnostics":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"warning","summary":"Needs attention.","checks":[{"label":"Provider","status":"warning","detail":"Ollama unavailable.","action":"Start Ollama."}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	restore.statusURL = server.URL + "/status"
	restore.diagnosticsURL = server.URL + "/diagnostics"
	restore.applyURLs()

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(doctorCmd)

	output, err := executeCommand(root, "doctor", "--json")
	if err != nil {
		t.Fatalf("doctor --json failed: %v", err)
	}
	var report doctorReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("doctor --json output is not JSON: %v\n%s", err, output)
	}
	if !report.Daemon.Running {
		t.Fatal("daemon.running should be true")
	}
	if report.Daemon.Status == nil || report.Daemon.Status.Version != "json-version" {
		t.Fatalf("daemon status missing version: %#v", report.Daemon.Status)
	}
	if report.Daemon.Status.ActiveAgents != 3 {
		t.Fatalf("active_agents = %d, want 3", report.Daemon.Status.ActiveAgents)
	}
	if report.Daemon.Diagnostics == nil || report.Daemon.Diagnostics.Status != "warning" {
		t.Fatalf("daemon diagnostics missing warning status: %#v", report.Daemon.Diagnostics)
	}
	if len(report.Daemon.Diagnostics.Checks) != 1 {
		t.Fatalf("diagnostic checks len = %d, want 1", len(report.Daemon.Diagnostics.Checks))
	}
	if strings.Contains(strings.ToLower(output), "api_key") {
		t.Fatalf("doctor JSON should not include API key fields: %s", output)
	}
}

func TestDoctorCmdRedactsDaemonDiagnosticSecrets(t *testing.T) {
	restore := stubDoctorDaemon(t)
	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"uptime":30,"version":"json-version","active_agents":3}`))
		case "/diagnostics":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"warning","summary":"api_key sk-phase5-secret","checks":[{"label":"Provider token","status":"warning","detail":"Bearer phase5-token","action":"password phase5-password"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	restore.statusURL = server.URL + "/status"
	restore.diagnosticsURL = server.URL + "/diagnostics"
	restore.applyURLs()

	for _, args := range [][]string{{"doctor"}, {"doctor", "--json"}} {
		root := &cobra.Command{Use: "starclaw"}
		root.AddCommand(doctorCmd)
		output, err := executeCommand(root, args...)
		if err != nil {
			t.Fatalf("%s failed: %v", strings.Join(args, " "), err)
		}
		for _, forbidden := range []string{"api_key", "sk-phase5-secret", "Provider token", "Bearer phase5-token", "password", "phase5-password"} {
			if strings.Contains(output, forbidden) {
				t.Fatalf("%s leaked %q: %s", strings.Join(args, " "), forbidden, output)
			}
		}
		if !strings.Contains(output, "[REDACTED]") {
			t.Fatalf("%s output missing redaction marker: %s", strings.Join(args, " "), output)
		}
	}
}

type doctorDaemonStub struct {
	isHealthy      func(context.Context) bool
	statusURL      string
	diagnosticsURL string
}

func stubDoctorDaemon(t *testing.T) *doctorDaemonStub {
	t.Helper()
	stub := &doctorDaemonStub{
		statusURL:      daemonStatusURL,
		diagnosticsURL: daemonDiagnosticsURL,
	}

	origHealthy := isDaemonHealthy
	origStatusURL := daemonStatusURL
	origDiagnosticsURL := daemonDiagnosticsURL
	origHTTPClient := doctorHTTPClient
	origOutputJSON := doctorOutputJSON

	isDaemonHealthy = func(ctx context.Context) bool {
		if stub.isHealthy != nil {
			return stub.isHealthy(ctx)
		}
		return false
	}
	doctorHTTPClient = http.DefaultClient

	t.Cleanup(func() {
		isDaemonHealthy = origHealthy
		daemonStatusURL = origStatusURL
		daemonDiagnosticsURL = origDiagnosticsURL
		doctorHTTPClient = origHTTPClient
		doctorOutputJSON = origOutputJSON
	})

	return stub
}

func (s *doctorDaemonStub) applyURLs() {
	daemonStatusURL = s.statusURL
	daemonDiagnosticsURL = s.diagnosticsURL
}
