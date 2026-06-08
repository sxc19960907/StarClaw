package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func withTerminalWorkspaceHooks(
	t *testing.T,
	availability func() terminalWorkspaceAvailability,
	ensure func() error,
	newTab func(context.Context, string, string, string) (int, int, error),
	newSplit func(context.Context, string, string, string, string) (int, int, error),
	sendInput func(context.Context, int, int, string) error,
) {
	t.Helper()
	oldAvailability := terminalWorkspaceAvailabilityStatusFn
	oldEnsure := terminalWorkspaceEnsureAvailableFn
	oldNewTab := terminalWorkspaceNewTabFn
	oldNewSplit := terminalWorkspaceNewSplitFn
	oldSendInput := terminalWorkspaceSendInputFn
	if availability != nil {
		terminalWorkspaceAvailabilityStatusFn = availability
	}
	if ensure != nil {
		terminalWorkspaceEnsureAvailableFn = ensure
	}
	if newTab != nil {
		terminalWorkspaceNewTabFn = newTab
	}
	if newSplit != nil {
		terminalWorkspaceNewSplitFn = newSplit
	}
	if sendInput != nil {
		terminalWorkspaceSendInputFn = sendInput
	}
	t.Cleanup(func() {
		terminalWorkspaceAvailabilityStatusFn = oldAvailability
		terminalWorkspaceEnsureAvailableFn = oldEnsure
		terminalWorkspaceNewTabFn = oldNewTab
		terminalWorkspaceNewSplitFn = oldNewSplit
		terminalWorkspaceSendInputFn = oldSendInput
	})
}

func TestRegisterLocalToolsIncludesTerminalWorkspace(t *testing.T) {
	t.Parallel()
	reg := RegisterLocalTools()
	tool, ok := reg.Get("terminal_workspace")
	if !ok {
		t.Fatal("terminal_workspace not registered")
	}
	if tool.Info().Name != "terminal_workspace" {
		t.Fatalf("tool name = %q", tool.Info().Name)
	}
}

func TestTerminalWorkspaceToolInfoAndApproval(t *testing.T) {
	t.Parallel()
	tool := NewTerminalWorkspaceTool()
	info := tool.Info()
	if info.Name != "terminal_workspace" {
		t.Fatalf("name = %q", info.Name)
	}
	required := map[string]bool{}
	for _, name := range info.Required {
		required[name] = true
	}
	if !required["action"] || !required["description"] {
		t.Fatalf("required = %v", info.Required)
	}
	if !tool.RequiresApproval() {
		t.Fatal("terminal_workspace should require approval")
	}
}

func TestTerminalWorkspaceReadOnlyClassification(t *testing.T) {
	t.Parallel()
	tool := NewTerminalWorkspaceTool()
	tests := []struct {
		name string
		args string
		want bool
	}{
		{"status", `{"action":"status"}`, true},
		{"list", `{"action":"list_tabs"}`, true},
		{"new tab", `{"action":"new_tab"}`, false},
		{"send input", `{"action":"send_input"}`, false},
		{"invalid", `not json`, false},
		{"unknown", `{"action":"bogus"}`, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tool.IsReadOnlyCall(tt.args); got != tt.want {
				t.Fatalf("IsReadOnlyCall = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminalWorkspaceInvalidJSONAndUnknownAction(t *testing.T) {
	t.Parallel()
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `not json`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || res.ErrorCategory != agent.ErrCategoryValidation {
		t.Fatalf("invalid JSON result = %#v", res)
	}

	res, err = tool.Run(context.Background(), `{"action":"bogus"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "unknown action") {
		t.Fatalf("unknown action result = %#v", res)
	}
}

func TestTerminalWorkspaceStatus(t *testing.T) {
	withTerminalWorkspaceHooks(t,
		func() terminalWorkspaceAvailability {
			return terminalWorkspaceAvailability{
				supported: true,
				available: true,
				version:   "1.4.0",
			}
		},
		nil, nil, nil, nil,
	)
	tool := NewTerminalWorkspaceTool()
	tool.tabs.add("dev", terminalTabRef{windowIndex: 1, tabIndex: 2})
	res, err := tool.Run(context.Background(), `{"action":"status"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	var status terminalWorkspaceStatus
	if err := json.Unmarshal([]byte(res.Content), &status); err != nil {
		t.Fatalf("status JSON: %v\n%s", err, res.Content)
	}
	if !status.Available || !status.Supported || status.DetectedVersion != "1.4.0" {
		t.Fatalf("status = %+v", status)
	}
	if status.TrackedTabs != 1 {
		t.Fatalf("tracked tabs = %d, want 1", status.TrackedTabs)
	}
}

func TestTerminalWorkspaceUnavailableFallback(t *testing.T) {
	withTerminalWorkspaceHooks(t, nil, func() error {
		return errors.New("Ghostty unavailable")
	}, nil, nil, nil)
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `{"action":"new_tab","description":"open dev server"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || res.ErrorCategory != agent.ErrCategoryBusiness {
		t.Fatalf("result = %#v, want business error", res)
	}
	if !strings.Contains(res.Content, "Fallback") || !strings.Contains(res.Content, "bash") {
		t.Fatalf("fallback missing from result: %s", res.Content)
	}
}

func TestTerminalWorkspaceListTabs(t *testing.T) {
	t.Parallel()
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `{"action":"list_tabs"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError || !strings.Contains(res.Content, "No tracked") {
		t.Fatalf("empty list result = %#v", res)
	}

	tool.tabs.add("test", terminalTabRef{windowIndex: 1, tabIndex: 2})
	res, err = tool.Run(context.Background(), `{"action":"list_tabs"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(res.Content, "test") || !strings.Contains(res.Content, "window:1") {
		t.Fatalf("list result = %s", res.Content)
	}
}

func TestTerminalWorkspaceNewTabTracksTitle(t *testing.T) {
	withTerminalWorkspaceHooks(t,
		nil,
		func() error { return nil },
		func(_ context.Context, command, title, color string) (int, int, error) {
			if command != "npm run dev" {
				t.Fatalf("command = %q", command)
			}
			if title != "dev" {
				t.Fatalf("title = %q", title)
			}
			if !strings.HasPrefix(color, "#") {
				t.Fatalf("color = %q", color)
			}
			return 2, 3, nil
		},
		nil,
		nil,
	)
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `{"action":"new_tab","description":"dev server","command":"npm run dev","title":"dev"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	ref, ok := tool.tabs.lookup("dev")
	if !ok {
		t.Fatal("dev tab not tracked")
	}
	if ref.windowIndex != 2 || ref.tabIndex != 3 {
		t.Fatalf("ref = %+v", ref)
	}
}

func TestTerminalWorkspaceNewSplitValidationAndTracking(t *testing.T) {
	withTerminalWorkspaceHooks(t,
		nil,
		func() error { return nil },
		nil,
		func(_ context.Context, direction, command, title, color string) (int, int, error) {
			if direction != "down" {
				t.Fatalf("direction = %q", direction)
			}
			if command != "tail -f app.log" || title != "logs" || color == "" {
				t.Fatalf("args command=%q title=%q color=%q", command, title, color)
			}
			return 4, 5, nil
		},
		nil,
	)
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `{"action":"new_split","description":"logs","direction":"left"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "direction") {
		t.Fatalf("invalid direction result = %#v", res)
	}

	res, err = tool.Run(context.Background(), `{"action":"new_split","description":"logs","direction":"down","command":"tail -f app.log","title":"logs"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	ref, ok := tool.tabs.lookup("logs")
	if !ok || ref.windowIndex != 4 || ref.tabIndex != 5 {
		t.Fatalf("tracked ref ok=%v ref=%+v", ok, ref)
	}
}

func TestTerminalWorkspaceSendInputValidationAndSuccess(t *testing.T) {
	var gotWindow, gotTab int
	var gotText string
	withTerminalWorkspaceHooks(t,
		nil,
		func() error { return nil },
		nil,
		nil,
		func(_ context.Context, windowIdx, tabIdx int, text string) error {
			gotWindow = windowIdx
			gotTab = tabIdx
			gotText = text
			return nil
		},
	)
	tool := NewTerminalWorkspaceTool()
	res, err := tool.Run(context.Background(), `{"action":"send_input","description":"send","text":"echo hi"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "target") {
		t.Fatalf("missing target result = %#v", res)
	}

	res, err = tool.Run(context.Background(), `{"action":"send_input","description":"send","target":"missing","text":"echo hi"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "not found") {
		t.Fatalf("missing tracked tab result = %#v", res)
	}

	tool.tabs.add("dev", terminalTabRef{windowIndex: 7, tabIndex: 8})
	res, err = tool.Run(context.Background(), `{"action":"send_input","description":"send","target":"dev","text":"echo hi"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if gotWindow != 7 || gotTab != 8 || gotText != "echo hi" {
		t.Fatalf("got window=%d tab=%d text=%q", gotWindow, gotTab, gotText)
	}
}

func TestTerminalWorkspaceTitleColorAndVersion(t *testing.T) {
	t.Parallel()
	if got := terminalWorkspaceTitle("custom", "npm run dev"); got != "custom" {
		t.Fatalf("custom title = %q", got)
	}
	if got := terminalWorkspaceTitle("", "/usr/bin/npm run dev"); got != "npm" {
		t.Fatalf("command title = %q", got)
	}
	if got := terminalWorkspaceTitle("", ""); got != "terminal" {
		t.Fatalf("empty title = %q", got)
	}
	if c1, c2 := terminalWorkspaceColor("dev"), terminalWorkspaceColor("dev"); c1 != c2 || !strings.HasPrefix(c1, "#") {
		t.Fatalf("color c1=%q c2=%q", c1, c2)
	}
	if compareSemverLike("1.3.0", "1.3.0") != 0 {
		t.Fatal("same versions should compare equal")
	}
	if compareSemverLike("1.2.9", "1.3.0") >= 0 {
		t.Fatal("older version should compare lower")
	}
	if compareSemverLike("v1.10.0-beta", "1.3.0") <= 0 {
		t.Fatal("newer prefixed prerelease should compare higher")
	}
}
