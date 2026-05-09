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

// ComputerTool provides mouse and keyboard control for macOS via osascript and cliclick.
type ComputerTool struct{}

type computerArgs struct {
	Action string `json:"action"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Key    string `json:"key,omitempty"`
	Text   string `json:"text,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *ComputerTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "computer",
		Description: "Control mouse and keyboard on macOS. Actions: mouse_move (move cursor to coordinates), mouse_click (click at coordinates), mouse_doubleclick (double-click at coordinates), key_press (press a key or key combination like 'return', 'command+c'), type_text (type text string), screenshot (capture screen). Requires cliclick (brew install cliclick) for mouse operations.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: mouse_move, mouse_click, mouse_doubleclick, key_press, type_text, screenshot",
				},
				"x": map[string]any{
					"type":        "integer",
					"description": "Screen X coordinate (for mouse_move, mouse_click, mouse_doubleclick)",
				},
				"y": map[string]any{
					"type":        "integer",
					"description": "Screen Y coordinate (for mouse_move, mouse_click, mouse_doubleclick)",
				},
				"key": map[string]any{
					"type":        "string",
					"description": "Key or key combination to press (for key_press). Examples: 'return', 'tab', 'escape', 'command+c', 'cmd+shift+z'",
				},
				"text": map[string]any{
					"type":        "string",
					"description": "Text to type (for type_text action). Supports Unicode/CJK characters.",
				},
			},
		},
		Required: []string{"action"},
	}
}

// Run executes the computer tool.
func (t *ComputerTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "computer tool is only available on macOS",
			IsError: true,
		}, nil
	}

	var args computerArgs
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
	case "mouse_move":
		return t.mouseMove(ctx, args.X, args.Y)
	case "mouse_click":
		return t.mouseClick(ctx, args.X, args.Y, 1)
	case "mouse_doubleclick":
		return t.mouseClick(ctx, args.X, args.Y, 2)
	case "key_press":
		return t.keyPress(ctx, args.Key)
	case "type_text":
		return t.typeText(ctx, args.Text)
	case "screenshot":
		return t.screenshot(ctx)
	default:
		return agent.ToolResult{
			Content: fmt.Sprintf("unknown action: %q (valid: mouse_move, mouse_click, mouse_doubleclick, key_press, type_text, screenshot)", args.Action),
			IsError: true,
		}, nil
	}
}

// RequiresApproval returns true because mouse/keyboard control modifies system state.
func (t *ComputerTool) RequiresApproval() bool { return true }

// IsReadOnlyCall returns false because all actions modify the system.
func (t *ComputerTool) IsReadOnlyCall(string) bool { return false }

// cliclickAvailable returns true if the cliclick command is installed.
func cliclickAvailable() bool {
	err := exec.Command("which", "cliclick").Run()
	return err == nil
}

// mouseMove moves the mouse cursor to (x, y) using cliclick.
func (t *ComputerTool) mouseMove(ctx context.Context, x, y int) (agent.ToolResult, error) {
	if !cliclickAvailable() {
		return agent.ToolResult{
			Content: "mouse_move requires cliclick. Install with: brew install cliclick",
			IsError: true,
		}, nil
	}

	cmd := exec.CommandContext(ctx, "cliclick", fmt.Sprintf("m:%d,%d", x, y))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("cliclick error: %v\n%s", err, string(output)),
			IsError: true,
		}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Moved cursor to (%d, %d)", x, y)}, nil
}

// mouseClick performs a click or double-click at (x, y) using cliclick.
func (t *ComputerTool) mouseClick(ctx context.Context, x, y int, clicks int) (agent.ToolResult, error) {
	if !cliclickAvailable() {
		// Fall back to osascript click at (moves and clicks together)
		script := fmt.Sprintf(`tell application "System Events" to click at {%d, %d}`, x, y)
		if clicks > 1 {
			script = fmt.Sprintf(`tell application "System Events"
				click at {%d, %d}
				delay 0.1
				click at {%d, %d}
			end tell`, x, y, x, y)
		}
		out, err := execOsascript(ctx, script)
		if err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("click error: %v", err), IsError: true}, nil
		}
		_ = out
		action := "Clicked"
		if clicks > 1 {
			action = "Double-clicked"
		}
		return agent.ToolResult{Content: fmt.Sprintf("%s at (%d, %d) via osascript", action, x, y)}, nil
	}

	// Use cliclick
	action := "c"
	if clicks > 1 {
		action = "dc"
	}
	cmd := exec.CommandContext(ctx, "cliclick", fmt.Sprintf("%s:%d,%d", action, x, y))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("cliclick error: %v\n%s", err, string(output)),
			IsError: true,
		}, nil
	}
	_ = output
	actionName := "Clicked"
	if clicks > 1 {
		actionName = "Double-clicked"
	}
	return agent.ToolResult{Content: fmt.Sprintf("%s at (%d, %d)", actionName, x, y)}, nil
}

// keyPress presses a key or key combination using osascript.
func (t *ComputerTool) keyPress(ctx context.Context, key string) (agent.ToolResult, error) {
	if key == "" {
		return agent.ToolResult{
			Content: "key_press requires 'key' parameter",
			IsError: true,
		}, nil
	}

	script := buildKeyPressScript(key)
	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("key press error: %v", err), IsError: true}, nil
	}
	_ = out
	return agent.ToolResult{Content: fmt.Sprintf("Pressed: %s", key)}, nil
}

// typeText types a text string using osascript keystroke.
func (t *ComputerTool) typeText(ctx context.Context, text string) (agent.ToolResult, error) {
	if text == "" {
		return agent.ToolResult{
			Content: "type_text requires 'text' parameter",
			IsError: true,
		}, nil
	}

	escaped := escapeAppleScriptArg(text)
	script := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escaped)
	out, err := execOsascript(ctx, script)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("type error: %v", err), IsError: true}, nil
	}
	_ = out
	return agent.ToolResult{Content: fmt.Sprintf("Typed: %s", text)}, nil
}

// screenshot captures the screen using screencapture.
func (t *ComputerTool) screenshot(ctx context.Context) (agent.ToolResult, error) {
	st := &ScreenshotTool{}
	return st.Run(ctx, "{}")
}

// macOS key code mapping for special keys
var keyCodeMap = map[string]int{
	"return":      36,
	"enter":       36,
	"tab":         48,
	"space":       49,
	"delete":      51,
	"backspace":   51,
	"escape":      53,
	"esc":         53,
	"up":          126,
	"down":        125,
	"left":        123,
	"right":       124,
	"home":        115,
	"end":         119,
	"pageup":      116,
	"pagedown":    121,
	"forwarddel":  117,
	"f1":          122,
	"f2":          120,
	"f3":          99,
	"f4":          118,
	"f5":          96,
	"f6":          97,
	"f7":          98,
	"f8":          100,
	"f9":          101,
	"f10":         109,
	"f11":         103,
	"f12":         111,
	"help":        114,
	"capslock":    57,
}

// buildKeyPressScript generates an AppleScript for pressing a key or key combo.
func buildKeyPressScript(key string) string {
	parts := strings.Split(strings.ToLower(key), "+")

	// Single key (no modifier)
	if len(parts) == 1 {
		keyName := parts[0]
		if code, ok := keyCodeMap[keyName]; ok {
			return fmt.Sprintf(`tell application "System Events" to key code %d`, code)
		}
		if len(keyName) == 1 {
			return fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escapeAppleScriptArg(keyName))
		}
		return fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escapeAppleScriptArg(keyName))
	}

	// Key with modifiers: last part is the key, rest are modifiers
	keyName := parts[len(parts)-1]
	mods := parts[:len(parts)-1]

	var modClause []string
	for _, m := range mods {
		switch m {
		case "command", "cmd":
			modClause = append(modClause, "command down")
		case "shift":
			modClause = append(modClause, "shift down")
		case "option", "alt":
			modClause = append(modClause, "option down")
		case "control", "ctrl":
			modClause = append(modClause, "control down")
		default:
			modClause = append(modClause, "command down")
		}
	}

	// Handle modifier + key code (e.g., command+up)
	keyIsKeyCode := false
	keyCodeVal := 0
	if code, ok := keyCodeMap[keyName]; ok {
		keyIsKeyCode = true
		keyCodeVal = code
	}

	if len(modClause) == 1 {
		if keyIsKeyCode {
			return fmt.Sprintf(`tell application "System Events" to key code %d using {%s}`, keyCodeVal, modClause[0])
		}
		return fmt.Sprintf(`tell application "System Events" to keystroke "%s" using %s`, escapeAppleScriptArg(keyName), modClause[0])
	}

	if keyIsKeyCode {
		return fmt.Sprintf(`tell application "System Events" to key code %d using {%s}`, keyCodeVal, strings.Join(modClause, ", "))
	}
	return fmt.Sprintf(`tell application "System Events" to keystroke "%s" using {%s}`, escapeAppleScriptArg(keyName), strings.Join(modClause, ", "))
}
