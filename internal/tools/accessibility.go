// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// AccessibilityTool queries macOS UI elements via osascript System Events.
type AccessibilityTool struct{}

type accessibilityArgs struct {
	Action string `json:"action"`
	App    string `json:"app,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *AccessibilityTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "accessibility",
		Description: "Query macOS UI elements via System Events. Actions: get_focused_element (get focused UI element info), get_element_at (get element at screen coordinates), list_windows (list all windows for an app), get_window_title (get title of frontmost window). On non-macOS this tool returns an error.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action to perform: get_focused_element, get_element_at, list_windows, get_window_title",
				},
				"app": map[string]any{
					"type":        "string",
					"description": "Application name filter (e.g. 'Safari', 'Finder'). Optional, defaults to frontmost app.",
				},
				"x": map[string]any{
					"type":        "integer",
					"description": "X screen coordinate (for get_element_at action)",
				},
				"y": map[string]any{
					"type":        "integer",
					"description": "Y screen coordinate (for get_element_at action)",
				},
			},
		},
		Required: []string{"action"},
	}
}

// Run executes the accessibility tool.
func (t *AccessibilityTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "accessibility tool is only available on macOS",
			IsError: true,
		}, nil
	}

	var args accessibilityArgs
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
	case "get_focused_element":
		return t.getFocusedElement(ctx)
	case "get_element_at":
		return t.getElementAt(ctx, args.X, args.Y)
	case "list_windows":
		return t.listWindows(ctx, args.App)
	case "get_window_title":
		return t.getWindowTitle(ctx, args.App)
	default:
		return agent.ToolResult{
			Content: fmt.Sprintf("unknown action: %q (valid: get_focused_element, get_element_at, list_windows, get_window_title)", args.Action),
			IsError: true,
		}, nil
	}
}

// RequiresApproval returns true because accessibility operations can interact with UI.
func (t *AccessibilityTool) RequiresApproval() bool { return true }

// IsReadOnlyCall returns true for all actions — accessibility queries are read-only.
func (t *AccessibilityTool) IsReadOnlyCall(string) bool { return true }

// execOsascript runs an osascript command with the given AppleScript and returns the output.
// The script should be valid AppleScript with proper quoting — it is passed directly to osascript -e.
func execOsascript(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("osascript error: %w\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// escapeAppleScriptArg escapes backslashes and double-quotes for embedding user-supplied
// values into AppleScript string literals.
func escapeAppleScriptArg(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func (t *AccessibilityTool) getFocusedElement(ctx context.Context) (agent.ToolResult, error) {
	script := `tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	tell process frontApp
		try
			set elemDesc to description of focused UI element
			set elemRole to role of focused UI element
			set elemName to name of focused UI element
			return "Application: " & frontApp & return & "Role: " & elemRole & return & "Name: " & elemName & return & "Description: " & elemDesc
		on error errMsg
			return "Application: " & frontApp & return & "No focused element could be identified." & return & "Error: " & errMsg
		end try
	end tell
end tell`

	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("accessibility error: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: out}, nil
}

func (t *AccessibilityTool) getElementAt(ctx context.Context, x, y int) (agent.ToolResult, error) {
	script := fmt.Sprintf(`tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	tell process frontApp
		set winTitle to ""
		try
			set winTitle to title of window 1
		end try
		set winCount to count of windows
		set positionInfo to "Requested coordinates: (%d, %d)" & return
		set positionInfo to positionInfo & "Application: " & frontApp & return
		set positionInfo to positionInfo & "Windows open: " & winCount & return
		if winTitle is not "" then
			set positionInfo to positionInfo & "Frontmost window: " & winTitle & return
		end if
		repeat with w in windows
			try
				set wTitle to title of w
				set wPos to position of w
				set wSize to size of w
				set px to (item 1 of wPos) as integer
				set py to (item 2 of wPos) as integer
				set wx to (item 1 of wSize) as integer
				set wy to (item 2 of wSize) as integer
				set positionInfo to positionInfo & "Window: " & wTitle & " at (" & px & ", " & py & ") size " & wx & "x" & wy & return
			end try
		end repeat
		return positionInfo
	end tell
end tell`, x, y)

	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("accessibility error: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: out}, nil
}

func (t *AccessibilityTool) listWindows(ctx context.Context, app string) (agent.ToolResult, error) {
	appClause := `name of first application process whose frontmost is true`
	if app != "" {
		safeApp := escapeAppleScriptArg(app)
		appClause = fmt.Sprintf(`"%s"`, safeApp)
	}
	script := fmt.Sprintf(`tell application "System Events"
	set targetApp to %s
	try
		tell process targetApp
			set winNames to name of every window
			set winCount to count of windows
			set resultStr to "Application: " & targetApp & return & "Window count: " & winCount & return
			repeat with w in winNames
				set resultStr to resultStr & "- " & w & return
			end repeat
			return resultStr
		end tell
	on error errMsg
		return "Application: " & targetApp & return & "Could not list windows: " & errMsg
	end try
end tell`, appClause)

	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("accessibility error: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: out}, nil
}

func (t *AccessibilityTool) getWindowTitle(ctx context.Context, app string) (agent.ToolResult, error) {
	appClause := `name of first application process whose frontmost is true`
	if app != "" {
		safeApp := escapeAppleScriptArg(app)
		appClause = fmt.Sprintf(`"%s"`, safeApp)
	}
	script := fmt.Sprintf(`tell application "System Events"
	set targetApp to %s
	try
		tell process targetApp
			if exists window 1 then
				set winTitle to title of window 1
				return "Application: " & targetApp & return & "Window title: " & winTitle
			else
				return "Application: " & targetApp & return & "No windows open"
			end if
		end tell
	on error errMsg
		return "Application: " & targetApp & return & "Could not get window title: " & errMsg
	end try
end tell`, appClause)

	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("accessibility error: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: out}, nil
}
