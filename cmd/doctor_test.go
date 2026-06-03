package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
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
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

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
	})

	return stub
}

func (s *doctorDaemonStub) applyURLs() {
	daemonStatusURL = s.statusURL
	daemonDiagnosticsURL = s.diagnosticsURL
}
