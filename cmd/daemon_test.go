package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
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
		"Diagnostics:   " + daemonDiagnosticsURL,
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
