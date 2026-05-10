// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/starclaw/starclaw/internal/agent"
)

// BrowserTool opens URLs in the default browser and retrieves page titles.
type BrowserTool struct{}

type browserArgs struct {
	Action string `json:"action"`
	URL    string `json:"url,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *BrowserTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "browser",
		Description: "Open URLs in the default browser or get the current browser page title. Actions: navigate (open URL in default browser), get_title (retrieve title of frontmost browser window via macOS AppleScript). Uses 'open' on macOS, 'xdg-open' on Linux, 'start' on Windows.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: navigate (open URL), get_title (get current page title)",
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
	default:
		return agent.ToolResult{
			Content: fmt.Sprintf("unknown action: %q (valid: navigate, get_title)", args.Action),
			IsError: true,
		}, nil
	}
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
	return args.Action == "get_title"
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

// getTitle retrieves the title of the frontmost browser window via osascript.
// On non-macOS, this returns an error since get_title relies on AppleScript.
func (t *BrowserTool) getTitle(ctx context.Context) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "get_title is only available on macOS (requires osascript)",
			IsError: true,
		}, nil
	}

	// Try Safari first, then Chrome/Chromium, then fall back to frontmost window title.
	script := `try
	tell application "Safari"
		if exists of document 1 then
			set pageTitle to name of document 1
			set pageURL to URL of document 1
			return "Browser: Safari" & return & "Title: " & pageTitle & return & "URL: " & pageURL
		end if
	end tell
end try
try
	tell application "Google Chrome"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "Browser: Google Chrome" & return & "Title: " & pageTitle & return & "URL: " & pageURL
		end if
	end tell
end try
try
	tell application "Chromium"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "Browser: Chromium" & return & "Title: " & pageTitle & return & "URL: " & pageURL
		end if
	end tell
end try
try
	tell application "Brave Browser"
		if exists of window 1 then
			set pageTitle to title of active tab of front window
			set pageURL to URL of active tab of front window
			return "Browser: Brave" & return & "Title: " & pageTitle & return & "URL: " & pageURL
		end if
	end tell
end try
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	try
		tell process frontApp
			if exists window 1 then
				set winTitle to title of window 1
				return "Browser: " & frontApp & return & "Title: " & winTitle
			end if
		end tell
	end try
	return "Application: " & frontApp & return & "No window title available"
end tell`

	out, err := execOsascript(ctx, script)
	if err != nil {
		// If osascript fails entirely, try a minimal AppleScript one-liner
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

	// If all tries failed (empty result), return a meaningful message
	if out == "" {
		return agent.ToolResult{
			Content: "Could not determine browser title. No supported browser found.",
			IsError: false,
		}, nil
	}

	return agent.ToolResult{Content: out}, nil
}
