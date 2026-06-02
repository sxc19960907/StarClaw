package cmd

import (
	"context"
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

type daemonLaunchStub struct {
	started   bool
	openURL   func(string) error
	isHealthy func(context.Context) bool
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
