//go:build !darwin

package tools

import (
	"context"
	"fmt"
	"runtime"
)

func terminalWorkspaceAvailabilityStatus() terminalWorkspaceAvailability {
	return terminalWorkspaceAvailability{
		supported: false,
		message:   fmt.Sprintf("terminal_workspace Ghostty backend is only supported on macOS; current platform is %s.", runtime.GOOS),
	}
}

func terminalWorkspaceEnsureAvailable() error {
	return fmt.Errorf("terminal_workspace Ghostty backend is only supported on macOS; current platform is %s", runtime.GOOS)
}

func terminalWorkspaceNewTab(context.Context, string, string, string) (int, int, error) {
	return 0, 0, terminalWorkspaceEnsureAvailable()
}

func terminalWorkspaceNewSplit(context.Context, string, string, string, string) (int, int, error) {
	return 0, 0, terminalWorkspaceEnsureAvailable()
}

func terminalWorkspaceSendInput(context.Context, int, int, string) error {
	return terminalWorkspaceEnsureAvailable()
}
