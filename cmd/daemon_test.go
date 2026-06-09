package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/starclaw/starclaw/internal/daemon"
	"github.com/starclaw/starclaw/internal/tools"
)

func TestDaemonOpenCmd(t *testing.T) {
	origOpen := openURLInBrowser
	defer func() { openURLInBrowser = origOpen }()

	var openedURL string
	openURLInBrowser = func(url string) error {
		openedURL = url
		return nil
	}

	root := &cobra.Command{Use: "starclaw"}
	daemon := &cobra.Command{Use: "daemon"}
	daemon.AddCommand(newDaemonOpenCmd())
	root.AddCommand(daemon)

	output, err := executeCommand(root, "daemon", "open")
	if err != nil {
		t.Fatalf("daemon open failed: %v", err)
	}
	if openedURL != daemonWebURL {
		t.Fatalf("opened URL = %q, want %q", openedURL, daemonWebURL)
	}
	if !strings.Contains(output, daemonWebURL) {
		t.Fatalf("output %q does not contain web URL %q", output, daemonWebURL)
	}
}

func TestAppCmdReusesRunningDaemon(t *testing.T) {
	restore := stubDaemonLaunch(t)

	var openedURL string
	restore.openURL = func(url string) error {
		openedURL = url
		return nil
	}
	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	output, err := executeCommand(root, "app")
	if err != nil {
		t.Fatalf("app failed: %v", err)
	}
	if restore.started {
		t.Fatal("app should not start daemon when health check succeeds")
	}
	if openedURL != daemonWebURL {
		t.Fatalf("opened URL = %q, want %q", openedURL, daemonWebURL)
	}
	if !strings.Contains(output, "Daemon already running") {
		t.Fatalf("output %q should report reused daemon", output)
	}
}

func TestAppCmdStartsDaemonBeforeOpen(t *testing.T) {
	restore := stubDaemonLaunch(t)

	healthChecks := 0
	var openedURL string
	restore.openURL = func(url string) error {
		openedURL = url
		return nil
	}
	restore.isHealthy = func(ctx context.Context) bool {
		healthChecks++
		return restore.started && healthChecks >= 2
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	output, err := executeCommand(root, "app")
	if err != nil {
		t.Fatalf("app failed: %v", err)
	}
	if !restore.started {
		t.Fatal("app should start daemon when health check fails")
	}
	if openedURL != daemonWebURL {
		t.Fatalf("opened URL = %q, want %q", openedURL, daemonWebURL)
	}
	if !strings.Contains(output, "Started daemon") {
		t.Fatalf("output %q should report daemon startup", output)
	}
}

func TestAppCmdCheckPrintsLaunchReadiness(t *testing.T) {
	restore := stubDaemonLaunch(t)

	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}
	restore.openURL = func(url string) error {
		t.Fatalf("app --check should not open browser, got %s", url)
		return nil
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	output, err := executeCommand(root, "app", "--check")
	if err != nil {
		t.Fatalf("app --check failed: %v", err)
	}
	if restore.started {
		t.Fatal("app --check should not start daemon")
	}
	for _, want := range []string{
		"StarClaw app launch readiness",
		"Launch:        starclaw app",
		"Daemon:        running",
		"Web UI:        " + daemonWebURL,
		"Health:        " + daemonHealthURL,
		"Status API:    " + daemonStatusURL,
		"Diagnostics:   " + daemonDiagnosticsURL,
		"Config:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestAppCmdNoOpenStartsDaemonWithoutBrowser(t *testing.T) {
	restore := stubDaemonLaunch(t)

	healthChecks := 0
	restore.isHealthy = func(ctx context.Context) bool {
		healthChecks++
		return restore.started && healthChecks >= 2
	}
	restore.openURL = func(url string) error {
		t.Fatalf("app --no-open should not open browser, got %s", url)
		return nil
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	output, err := executeCommand(root, "app", "--no-open")
	if err != nil {
		t.Fatalf("app --no-open failed: %v", err)
	}
	if !restore.started {
		t.Fatal("app --no-open should start daemon when health check fails")
	}
	if !strings.Contains(output, "Started daemon. Web UI: "+daemonWebURL) {
		t.Fatalf("output %q should report daemon startup without browser", output)
	}
}

func TestDaemonOpenStartFlagStartsDaemonBeforeOpen(t *testing.T) {
	restore := stubDaemonLaunch(t)

	restore.openURL = func(url string) error {
		return nil
	}
	restore.isHealthy = func(ctx context.Context) bool {
		return restore.started
	}

	root := &cobra.Command{Use: "starclaw"}
	daemon := &cobra.Command{Use: "daemon"}
	daemon.AddCommand(newDaemonOpenCmd())
	root.AddCommand(daemon)

	output, err := executeCommand(root, "daemon", "open", "--start")
	if err != nil {
		t.Fatalf("daemon open --start failed: %v", err)
	}
	if !restore.started {
		t.Fatal("daemon open --start should start daemon when health check fails")
	}
	if !strings.Contains(output, "Started daemon") {
		t.Fatalf("output %q should report daemon startup", output)
	}
}

func TestAppCmdStartupFailureIncludesDiagnosticsHint(t *testing.T) {
	restore := stubDaemonLaunch(t)

	restore.isHealthy = func(ctx context.Context) bool {
		return false
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	_, err := executeCommand(root, "app")
	if err == nil {
		t.Fatal("app should fail when daemon never becomes healthy")
	}
	message := err.Error()
	if !strings.Contains(message, "starclaw daemon status") {
		t.Fatalf("error %q should include daemon status hint", message)
	}
	if !strings.Contains(message, daemonDiagnosticsURL) {
		t.Fatalf("error %q should include diagnostics URL %q", message, daemonDiagnosticsURL)
	}
	if !strings.Contains(message, "port 7533") {
		t.Fatalf("error %q should include port hint", message)
	}
}

func TestAppCmdStartProcessFailureIncludesActionableHint(t *testing.T) {
	restore := stubDaemonLaunch(t)

	restore.startDaemon = func() error {
		return errors.New("listen tcp 127.0.0.1:7533: bind: address already in use")
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	_, err := executeCommand(root, "app")
	if err == nil {
		t.Fatal("app should fail when daemon process cannot start")
	}
	message := err.Error()
	for _, want := range []string{
		"address already in use",
		"port 7533 appears to be in use",
		"starclaw daemon status",
		daemonDiagnosticsURL,
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q should contain %q", message, want)
		}
	}
}

func TestAppCmdBrowserOpenFailureKeepsManualURL(t *testing.T) {
	restore := stubDaemonLaunch(t)

	restore.isHealthy = func(ctx context.Context) bool {
		return true
	}
	restore.openURL = func(url string) error {
		return errors.New("browser unavailable")
	}

	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(newAppCmd())

	_, err := executeCommand(root, "app")
	if err == nil {
		t.Fatal("app should report browser open failure")
	}
	message := err.Error()
	for _, want := range []string{
		"browser unavailable",
		"daemon is reachable",
		"open " + daemonWebURL + " manually",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q should contain %q", message, want)
		}
	}
}

func TestValidateDesktopRPCLaunchFlags(t *testing.T) {
	tests := []struct {
		name    string
		sock    string
		pidfile string
		enabled bool
		wantErr string
	}{
		{
			name:    "disabled",
			enabled: false,
		},
		{
			name:    "missing pidfile",
			sock:    "/tmp/daemon.sock",
			wantErr: "--rpc-pidfile is required",
		},
		{
			name:    "missing socket",
			pidfile: "/tmp/daemon.pid",
			wantErr: "--rpc-socket is required",
		},
		{
			name:    "enabled",
			sock:    "/tmp/daemon.sock",
			pidfile: "/tmp/daemon.pid",
			enabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled, err := validateDesktopRPCLaunchFlags(tt.sock, tt.pidfile)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if enabled != tt.enabled {
				t.Fatalf("enabled = %v, want %v", enabled, tt.enabled)
			}
		})
	}
}

func TestStartDaemonDesktopRPCListenerWritesPidfileAndStatus(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "sc-drpc-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	sockPath := filepath.Join(dir, "daemon.sock")
	pidfilePath := filepath.Join(dir, "daemon.pid")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := daemon.NewServer(0, &daemon.ServerDeps{Registry: tools.RegisterLocalTools()}, "test-version")
	if err := startDaemonDesktopRPCListener(ctx, srv, sockPath, pidfilePath, cancel); err != nil {
		t.Fatalf("startDaemonDesktopRPCListener: %v", err)
	}

	pidfile, err := os.ReadFile(pidfilePath)
	if err != nil {
		t.Fatalf("read pidfile: %v", err)
	}
	if strings.TrimSpace(string(pidfile)) == "" {
		t.Fatal("pidfile is empty")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	desktopRPC, ok := body["desktop_rpc"].(map[string]any)
	if !ok {
		t.Fatalf("desktop_rpc missing in %#v", body)
	}
	if desktopRPC["listening"] != true {
		t.Fatalf("desktop_rpc.listening = %#v, want true", desktopRPC["listening"])
	}
	if _, ok := desktopRPC["sock_path"]; ok {
		t.Fatalf("desktop_rpc status exposed sock_path: %#v", desktopRPC)
	}
	if _, ok := desktopRPC["pidfile_path"]; ok {
		t.Fatalf("desktop_rpc status exposed pidfile_path: %#v", desktopRPC)
	}

	cancel()
	waitForMissingPath(t, sockPath)
	waitForMissingPath(t, pidfilePath)
}

type daemonLaunchStub struct {
	started     bool
	openURL     func(string) error
	isHealthy   func(context.Context) bool
	startDaemon func() error
}

func stubDaemonLaunch(t *testing.T) *daemonLaunchStub {
	t.Helper()
	stub := &daemonLaunchStub{}

	origOpen := openURLInBrowser
	origHealthy := isDaemonHealthy
	origStart := startDaemonBackground
	origTimeout := daemonEnsureTimeout
	origPoll := daemonHealthPollInterval

	openURLInBrowser = func(url string) error {
		if stub.openURL != nil {
			return stub.openURL(url)
		}
		return nil
	}
	isDaemonHealthy = func(ctx context.Context) bool {
		if stub.isHealthy != nil {
			return stub.isHealthy(ctx)
		}
		return false
	}
	startDaemonBackground = func() error {
		stub.started = true
		if stub.startDaemon != nil {
			return stub.startDaemon()
		}
		return nil
	}
	daemonEnsureTimeout = 250 * time.Millisecond
	daemonHealthPollInterval = time.Millisecond

	t.Cleanup(func() {
		openURLInBrowser = origOpen
		isDaemonHealthy = origHealthy
		startDaemonBackground = origStart
		daemonEnsureTimeout = origTimeout
		daemonHealthPollInterval = origPoll
	})

	return stub
}

func waitForMissingPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("path still exists: %s", path)
}
