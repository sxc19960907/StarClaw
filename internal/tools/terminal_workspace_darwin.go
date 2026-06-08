//go:build darwin

package tools

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func terminalWorkspaceAvailabilityStatus() terminalWorkspaceAvailability {
	version, err := terminalWorkspaceGhosttyVersion()
	if err != nil {
		return terminalWorkspaceAvailability{
			supported: true,
			message:   fmt.Sprintf("Ghostty >= %s is required but was not found.", minTerminalWorkspaceGhosttyVersion),
		}
	}
	if compareSemverLike(version, minTerminalWorkspaceGhosttyVersion) < 0 {
		return terminalWorkspaceAvailability{
			supported: true,
			version:   version,
			message:   fmt.Sprintf("Ghostty %s is installed, but StarClaw requires Ghostty >= %s.", version, minTerminalWorkspaceGhosttyVersion),
		}
	}
	return terminalWorkspaceAvailability{
		supported: true,
		available: true,
		version:   version,
	}
}

func terminalWorkspaceEnsureAvailable() error {
	status := terminalWorkspaceAvailabilityStatus()
	if status.available {
		return nil
	}
	if status.message != "" {
		return errors.New(status.message)
	}
	return fmt.Errorf("terminal workspace backend is unavailable")
}

func terminalWorkspaceGhosttyVersion() (string, error) {
	out, err := exec.Command("mdfind", "kMDItemCFBundleIdentifier == 'com.mitchellh.ghostty'").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("ghostty app not found")
	}
	appPath := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	ver, err := exec.Command("defaults", "read", appPath+"/Contents/Info.plist", "CFBundleShortVersionString").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read ghostty version: %w", err)
	}
	return strings.TrimSpace(string(ver)), nil
}

func terminalWorkspaceExecScript(ctx context.Context, script string) (string, error) {
	var args []string
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			args = append(args, "-e", trimmed)
		}
	}
	out, err := exec.CommandContext(ctx, "osascript", args...).CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return "", fmt.Errorf("osascript error: %w\n%s", err, result)
	}
	return result, nil
}

func terminalWorkspaceNewTab(ctx context.Context, command, title, color string) (int, int, error) {
	_ = color
	script := `tell application "Ghostty"
	activate
	set win to front window
	set cfg to new surface configuration
	set newTab to new tab in win with configuration cfg
	set t to focused terminal of selected tab of win
end tell`
	if _, err := terminalWorkspaceExecScript(ctx, script); err != nil {
		return 0, 0, err
	}
	if title != "" {
		_ = terminalWorkspaceSetTabTitle(ctx, title)
	}
	if command != "" {
		_ = terminalWorkspaceSendCommand(ctx, command)
	}

	tabIdx := 1
	idxScript := `tell application "Ghostty"
	tell front window
		set tabIdx to count of tabs
	end tell
	return tabIdx as text
end tell`
	if result, err := terminalWorkspaceExecScript(ctx, idxScript); err == nil {
		_, _ = fmt.Sscanf(result, "%d", &tabIdx)
	}
	return 1, tabIdx, nil
}

func terminalWorkspaceNewSplit(ctx context.Context, direction, command, title, color string) (int, int, error) {
	_ = color
	script := fmt.Sprintf(`tell application "Ghostty"
	activate
	set win to front window
	set t1 to focused terminal of selected tab of win
	set cfg to new surface configuration
	set t2 to split t1 direction %s with configuration cfg
end tell`, direction)
	if _, err := terminalWorkspaceExecScript(ctx, script); err != nil {
		return 0, 0, err
	}
	if title != "" {
		_ = terminalWorkspaceSetTabTitle(ctx, title)
	}
	if command != "" {
		_ = terminalWorkspaceSendCommand(ctx, command)
	}
	return 1, 1, nil
}

func terminalWorkspaceSendInput(ctx context.Context, windowIdx, tabIdx int, text string) error {
	_ = windowIdx
	escaped := terminalWorkspaceEscapeAppleScript(text)
	script := fmt.Sprintf(`tell application "Ghostty"
	set win to window 1
	set targetTab to tab %d of win
	set t to focused terminal of targetTab
	input text "%s" to t
end tell`, tabIdx, escaped)
	_, err := terminalWorkspaceExecScript(ctx, script)
	return err
}

func terminalWorkspaceSetTabTitle(ctx context.Context, title string) error {
	escaped := terminalWorkspaceEscapeAppleScript(title)
	script := fmt.Sprintf(`tell application "Ghostty"
	tell selected tab of front window
		set title to "%s"
	end tell
end tell`, escaped)
	_, err := terminalWorkspaceExecScript(ctx, script)
	return err
}

func terminalWorkspaceSendCommand(ctx context.Context, command string) error {
	escaped := terminalWorkspaceEscapeAppleScript(command)
	script := fmt.Sprintf(`tell application "Ghostty"
	set t to focused terminal of selected tab of front window
	input text "%s" to t
	send key "enter" to t
end tell`, escaped)
	_, err := terminalWorkspaceExecScript(ctx, script)
	return err
}

func terminalWorkspaceEscapeAppleScript(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(escaped, `"`, `\"`)
}
