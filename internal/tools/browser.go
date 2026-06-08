// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/starclaw/starclaw/internal/agent"
)

// BrowserTool opens URLs in the default browser and retrieves page titles.
type BrowserTool struct {
	mu             sync.Mutex
	deprecated     bool
	cleanupCallCnt int
}

type browserArgs struct {
	Action string `json:"action"`
	URL    string `json:"url,omitempty"`
}

type browserStatus struct {
	SupportedBrowsers []string `json:"supported_browsers"`
	Actions           []string `json:"actions"`
	Platform          string   `json:"platform"`
	AutomationBackend string   `json:"automation_backend"`
	CanNavigate       bool     `json:"can_navigate"`
	CanInspect        bool     `json:"can_inspect"`
}

type browserSnapshot struct {
	Supported    bool   `json:"supported"`
	Platform     string `json:"platform"`
	Browser      string `json:"browser,omitempty"`
	Title        string `json:"title,omitempty"`
	URL          string `json:"url,omitempty"`
	FrontmostApp string `json:"frontmost_app,omitempty"`
	WindowTitle  string `json:"window_title,omitempty"`
	Source       string `json:"source"`
	Message      string `json:"message,omitempty"`
}

var supportedBrowserApps = []string{"Safari", "Google Chrome", "Chromium", "Brave Browser"}
var browserActions = []string{"navigate", "get_title", "status", "snapshot"}

// Info returns the tool definition for the LLM.
func (t *BrowserTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "browser",
		Description: "Open URLs in the default browser or inspect current browser state. Actions: navigate (open URL), get_title (compatible prose title lookup), status (structured capability metadata), snapshot (structured current browser URL/title metadata on macOS). Uses 'open' on macOS, 'xdg-open' on Linux, 'start' on Windows.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: navigate, get_title, status, snapshot",
				},
				"url": map[string]any{
					"type":        "string",
					"description": "URL to open in the default browser (required for navigate action)",
				},
			},
		},
		Required: []string{"action"},
	}
}

// Run executes the browser tool.
func (t *BrowserTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	MarkBrowserUsed(ctx, t)

	var args browserArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	if args.Action == "" {
		return agent.ToolResult{
			Content: "missing required parameter: action",
			IsError: true,
		}, nil
	}

	switch args.Action {
	case "navigate":
		if args.URL == "" {
			return agent.ToolResult{
				Content: "navigate action requires 'url' parameter",
				IsError: true,
			}, nil
		}
		return t.navigate(ctx, args.URL)
	case "get_title":
		return t.getTitle(ctx)
	case "status":
		return t.status()
	case "snapshot":
		return t.snapshot(ctx)
	default:
		return agent.ToolResult{
			Content: fmt.Sprintf("unknown action: %q (valid: %s)", args.Action, strings.Join(browserActions, ", ")),
			IsError: true,
		}, nil
	}
}

func (t *BrowserTool) MarkDeprecated() {
	t.mu.Lock()
	t.deprecated = true
	t.mu.Unlock()
}

func (t *BrowserTool) IsDeprecated() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.deprecated
}

func (t *BrowserTool) CleanupForHandoff() error {
	t.mu.Lock()
	t.cleanupCallCnt++
	t.mu.Unlock()
	return nil
}

func (t *BrowserTool) CleanupCalledForTest() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cleanupCallCnt
}

// RequiresApproval returns false for navigate (opening a URL is low-risk),
// true for other actions like get_title (reads browser state).
func (t *BrowserTool) RequiresApproval() bool { return false }

// IsReadOnlyCall returns the read-only status based on the action.
func (t *BrowserTool) IsReadOnlyCall(argsJSON string) bool {
	var args struct {
		Action string `json:"action"`
	}
	if json.Unmarshal([]byte(argsJSON), &args) != nil {
		return false
	}
	switch args.Action {
	case "get_title", "status", "snapshot":
		return true
	default:
		return false
	}
}

// navigate opens a URL in the default browser.
func (t *BrowserTool) navigate(ctx context.Context, url string) (agent.ToolResult, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", url)
	default:
		return agent.ToolResult{
			Content: fmt.Sprintf("unsupported platform: %s", runtime.GOOS),
			IsError: true,
		}, nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("failed to open URL: %v\n%s", err, string(output)),
			IsError: true,
		}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("Opened URL in default browser: %s", url)}, nil
}

func (t *BrowserTool) status() (agent.ToolResult, error) {
	status := browserStatus{
		SupportedBrowsers: append([]string(nil), supportedBrowserApps...),
		Actions:           append([]string(nil), browserActions...),
		Platform:          runtime.GOOS,
		AutomationBackend: automationBackendForPlatform(runtime.GOOS),
		CanNavigate:       runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "windows",
		CanInspect:        runtime.GOOS == "darwin",
	}
	return jsonToolResult(status)
}

func (t *BrowserTool) snapshot(ctx context.Context) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return jsonToolResult(browserSnapshot{
			Supported: false,
			Platform:  runtime.GOOS,
			Source:    "unsupported_platform",
			Message:   "snapshot is only available on macOS (requires osascript)",
		})
	}
	snap, err := currentBrowserSnapshot(ctx)
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("snapshot error: %v", err),
			IsError: true,
		}, nil
	}
	return jsonToolResult(snap)
}

// getTitle retrieves the title of the frontmost browser window via osascript.
// On non-macOS, this returns an error since get_title relies on AppleScript.
func (t *BrowserTool) getTitle(ctx context.Context) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "get_title is only available on macOS (requires osascript)",
			IsError: true,
		}, nil
	}

	snap, err := currentBrowserSnapshot(ctx)
	if err != nil {
		fallback := `tell application "System Events" to set frontApp to name of first application process whose frontmost is true`
		result, fbErr := execOsascript(ctx, fallback)
		if fbErr != nil {
			return agent.ToolResult{
				Content: fmt.Sprintf("get_title error: %v", fbErr),
				IsError: true,
			}, nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("Frontmost application: %s\n(Could not retrieve page title)", result),
		}, nil
	}
	if !snap.Supported {
		return agent.ToolResult{
			Content: "Could not determine browser title. No supported browser found.",
			IsError: false,
		}, nil
	}
	if snap.Browser != "" {
		out := fmt.Sprintf("Browser: %s", snap.Browser)
		if snap.Title != "" {
			out += "\nTitle: " + snap.Title
		}
		if snap.URL != "" {
			out += "\nURL: " + snap.URL
		}
		return agent.ToolResult{Content: out}, nil
	}
	if snap.FrontmostApp != "" {
		out := "Browser: " + snap.FrontmostApp
		if snap.WindowTitle != "" {
			out += "\nTitle: " + snap.WindowTitle
		}
		return agent.ToolResult{Content: out}, nil
	}
	return agent.ToolResult{Content: "Could not determine browser title. No supported browser found."}, nil
}

func currentBrowserSnapshot(ctx context.Context) (browserSnapshot, error) {
	out, err := execOsascript(ctx, browserSnapshotScript())
	if err != nil {
		return browserSnapshot{}, err
	}
	snap := parseBrowserSnapshotOutput(out, runtime.GOOS)
	snap.Platform = runtime.GOOS
	return snap, nil
}

func parseBrowserSnapshotOutput(out, platform string) browserSnapshot {
	out = strings.TrimSpace(out)
	if out == "" {
		return browserSnapshot{
			Supported: false,
			Platform:  platform,
			Source:    "empty_result",
			Message:   "No browser snapshot was returned.",
		}
	}
	parts := strings.Split(out, "\x1f")
	switch {
	case len(parts) >= 5 && parts[0] == "browser":
		return browserSnapshot{
			Supported: true,
			Platform:  platform,
			Browser:   parts[1],
			Title:     parts[2],
			URL:       parts[3],
			Source:    parts[4],
		}
	case len(parts) >= 4 && parts[0] == "window":
		return browserSnapshot{
			Supported:    true,
			Platform:     platform,
			FrontmostApp: parts[1],
			WindowTitle:  parts[2],
			Source:       parts[3],
		}
	case len(parts) >= 4 && parts[0] == "app":
		return browserSnapshot{
			Supported:    false,
			Platform:     platform,
			FrontmostApp: parts[1],
			Source:       parts[2],
			Message:      parts[3],
		}
	default:
		return browserSnapshot{
			Supported: false,
			Platform:  platform,
			Source:    "unrecognized_output",
			Message:   out,
		}
	}
}

func browserSnapshotScript() string {
	return `set sep to ASCII character 31
try
	tell application "Safari"
		if exists of document 1 then
			set pageTitle to name of document 1
			set pageURL to URL of document 1
			return "browser" & sep & "Safari" & sep & pageTitle & sep & pageURL & sep & "safari"
		end if
	end tell
end try
try
	tell application "Google Chrome"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "browser" & sep & "Google Chrome" & sep & pageTitle & sep & pageURL & sep & "chrome"
		end if
	end tell
end try
try
	tell application "Chromium"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "browser" & sep & "Chromium" & sep & pageTitle & sep & pageURL & sep & "chromium"
		end if
	end tell
end try
try
	tell application "Brave Browser"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "browser" & sep & "Brave Browser" & sep & pageTitle & sep & pageURL & sep & "brave"
		end if
	end tell
end try
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	try
		tell process frontApp
			if exists window 1 then
				set winTitle to title of window 1
				return "window" & sep & frontApp & sep & winTitle & sep & "frontmost_window"
			end if
		end tell
	end try
	return "app" & sep & frontApp & sep & "frontmost_app" & sep & "No window title available"
end tell`
}

func automationBackendForPlatform(goos string) string {
	switch goos {
	case "darwin":
		return "osascript"
	case "linux":
		return "xdg-open"
	case "windows":
		return "cmd-start"
	default:
		return "unsupported"
	}
}

func jsonToolResult(v any) (agent.ToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("json encode error: %v", err),
			IsError: true,
		}, nil
	}
	return agent.ToolResult{Content: string(data)}, nil
}
