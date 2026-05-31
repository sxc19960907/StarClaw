package cmd

import (
	"strings"
	"testing"

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
