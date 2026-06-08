package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDaemonEventDocumentation(t *testing.T) {
	root := repoRoot(t)
	doc, err := os.ReadFile(filepath.Join(root, "docs", "DAEMON_EVENTS.md"))
	if err != nil {
		t.Fatalf("read daemon event docs: %v", err)
	}
	body := string(doc)
	for _, want := range []string{
		"GET /events",
		"POST /message",
		"last_event_id",
		"Last-Event-ID",
		"SubscribeWithReplay",
		"bounded replay",
		"approval_needed",
		"approval_resolved",
		"run_started",
		"run_completed",
		"run_error",
		"run_status",
		"tool_status",
		"budget_status",
		"session_started",
		"assistant_text",
		"delta",
		"tool",
		"done",
		"args",
		"content",
		"prompt",
		"[REDACTED]",
		"local-first",
		"MESSAGE_LIFECYCLE",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("daemon event docs missing %q", want)
		}
	}

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "[Daemon Event Contracts](docs/DAEMON_EVENTS.md)") {
		t.Fatalf("README missing daemon event docs link")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
