package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/tui"
)

type doctorDaemonStatus struct {
	Uptime       int    `json:"uptime"`
	Version      string `json:"version"`
	ActiveAgents int    `json:"active_agents"`
}

type doctorDaemonDiagnostics struct {
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Checks  []struct {
		Label  string `json:"label"`
		Status string `json:"status"`
		Detail string `json:"detail"`
		Action string `json:"action,omitempty"`
	} `json:"checks"`
}

var doctorHTTPClient = &http.Client{Timeout: 700 * time.Millisecond}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local StarClaw readiness and support context",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor(cmd)
	},
}

func runDoctor(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	starclawDir := config.StarclawDir()
	configPath := filepath.Join(starclawDir, "config.yaml")

	fmt.Fprintln(out, "StarClaw doctor")
	fmt.Fprintf(out, "Version:       %s\n", Version)
	fmt.Fprintf(out, "Launch:        starclaw app\n")
	fmt.Fprintf(out, "Web UI:        %s\n", daemonWebURL)
	fmt.Fprintf(out, "Diagnostics:   %s\n", daemonDiagnosticsURL)
	fmt.Fprintf(out, "Data:          %s\n", starclawDir)
	fmt.Fprintf(out, "Config:        %s\n", configPath)
	fmt.Fprintln(out)

	printLocalDoctorChecks(out, tui.NewDoctor().RunChecks())
	fmt.Fprintln(out)

	printDaemonDoctorStatus(out)
	return nil
}

func printLocalDoctorChecks(out interface {
	Write([]byte) (int, error)
}, checks []tui.CheckResult) {
	fmt.Fprintln(out, "Local checks:")
	for _, check := range checks {
		fmt.Fprintf(out, "  [%s] %s: %s\n", doctorLocalStatusLabel(check.Status), check.Name, check.Message)
	}
}

func printDaemonDoctorStatus(out interface {
	Write([]byte) (int, error)
}) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if !isDaemonHealthy(ctx) {
		fmt.Fprintln(out, "Daemon:        not running")
		fmt.Fprintln(out, "Next steps:    run `starclaw app` to start the GUI, or `starclaw daemon start` for daemon-only mode")
		return
	}

	fmt.Fprintln(out, "Daemon:        running")
	if status, err := doctorFetchJSON[doctorDaemonStatus](ctx, daemonStatusURL); err != nil {
		fmt.Fprintf(out, "Status API:    warning: %v\n", err)
	} else {
		if status.Version != "" {
			fmt.Fprintf(out, "Daemon version:%s\n", doctorPaddedValue(status.Version))
		}
		fmt.Fprintf(out, "Active agents: %d\n", status.ActiveAgents)
		fmt.Fprintf(out, "Uptime:        %s\n", (time.Duration(status.Uptime) * time.Second).String())
	}

	diagnostics, err := doctorFetchJSON[doctorDaemonDiagnostics](ctx, daemonDiagnosticsURL)
	if err != nil {
		fmt.Fprintf(out, "Diagnostics:   warning: %v\n", err)
		return
	}
	fmt.Fprintf(out, "Runtime:       %s\n", diagnostics.Status)
	if diagnostics.Summary != "" {
		fmt.Fprintf(out, "Summary:       %s\n", diagnostics.Summary)
	}
	if len(diagnostics.Checks) == 0 {
		return
	}
	fmt.Fprintln(out, "Daemon checks:")
	for _, check := range diagnostics.Checks {
		fmt.Fprintf(out, "  [%s] %s: %s\n", check.Status, check.Label, check.Detail)
		if check.Action != "" {
			fmt.Fprintf(out, "       Action: %s\n", check.Action)
		}
	}
}

func doctorFetchJSON[T any](ctx context.Context, url string) (T, error) {
	var value T
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return value, fmt.Errorf("create request: %w", err)
	}
	resp, err := doctorHTTPClient.Do(req)
	if err != nil {
		return value, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return value, fmt.Errorf("request %s returned %s", url, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		return value, fmt.Errorf("decode %s: %w", url, err)
	}
	return value, nil
}

func doctorLocalStatusLabel(status tui.CheckStatus) string {
	switch status {
	case tui.CheckPass:
		return "pass"
	case tui.CheckWarn:
		return "warn"
	default:
		return "fail"
	}
}

func doctorPaddedValue(value string) string {
	if value == "" {
		return "-"
	}
	return " " + value
}
