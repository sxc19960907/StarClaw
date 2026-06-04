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

type doctorReport struct {
	Version        string             `json:"version"`
	LaunchCommand  string             `json:"launch_command"`
	WebURL         string             `json:"web_url"`
	DiagnosticsURL string             `json:"diagnostics_url"`
	StarclawDir    string             `json:"starclaw_dir"`
	ConfigPath     string             `json:"config_path"`
	LocalChecks    []doctorLocalCheck `json:"local_checks"`
	Daemon         doctorDaemonReport `json:"daemon"`
}

type doctorLocalCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type doctorDaemonReport struct {
	Running     bool                     `json:"running"`
	Status      *doctorDaemonStatus      `json:"status,omitempty"`
	Diagnostics *doctorDaemonDiagnostics `json:"diagnostics,omitempty"`
	Errors      []string                 `json:"errors,omitempty"`
}

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

var doctorOutputJSON bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local StarClaw readiness and support context",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor(cmd)
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorOutputJSON, "json", false, "Print diagnostics as JSON")
}

func runDoctor(cmd *cobra.Command) error {
	report := buildDoctorReport()
	if doctorOutputJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	printDoctorReport(cmd.OutOrStdout(), report)
	return nil
}

func buildDoctorReport() doctorReport {
	starclawDir := config.StarclawDir()
	report := doctorReport{
		Version:        Version,
		LaunchCommand:  "starclaw app",
		WebURL:         daemonWebURL,
		DiagnosticsURL: daemonDiagnosticsURL,
		StarclawDir:    starclawDir,
		ConfigPath:     filepath.Join(starclawDir, "config.yaml"),
		LocalChecks:    doctorLocalChecks(tui.NewDoctor().RunChecks()),
		Daemon:         buildDoctorDaemonReport(),
	}
	return report
}

func printDoctorReport(out interface {
	Write([]byte) (int, error)
}, report doctorReport) {
	_, _ = fmt.Fprintln(out, "StarClaw doctor")
	_, _ = fmt.Fprintf(out, "Version:       %s\n", report.Version)
	_, _ = fmt.Fprintf(out, "Launch:        %s\n", report.LaunchCommand)
	_, _ = fmt.Fprintf(out, "Web UI:        %s\n", report.WebURL)
	_, _ = fmt.Fprintf(out, "Diagnostics:   %s\n", report.DiagnosticsURL)
	_, _ = fmt.Fprintf(out, "Data:          %s\n", report.StarclawDir)
	_, _ = fmt.Fprintf(out, "Config:        %s\n", report.ConfigPath)
	_, _ = fmt.Fprintln(out)

	printLocalDoctorChecks(out, report.LocalChecks)
	_, _ = fmt.Fprintln(out)

	printDaemonDoctorStatus(out, report.Daemon)
}

func printLocalDoctorChecks(out interface {
	Write([]byte) (int, error)
}, checks []doctorLocalCheck) {
	fmt.Fprintln(out, "Local checks:")
	for _, check := range checks {
		fmt.Fprintf(out, "  [%s] %s: %s\n", check.Status, check.Name, check.Message)
	}
}

func buildDoctorDaemonReport() doctorDaemonReport {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if !isDaemonHealthy(ctx) {
		return doctorDaemonReport{Running: false}
	}

	report := doctorDaemonReport{Running: true}
	status, err := doctorFetchJSON[doctorDaemonStatus](ctx, daemonStatusURL)
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("status API: %v", err))
	} else {
		report.Status = &status
	}

	diagnostics, err := doctorFetchJSON[doctorDaemonDiagnostics](ctx, daemonDiagnosticsURL)
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("diagnostics API: %v", err))
	} else {
		report.Diagnostics = &diagnostics
	}
	return report
}

func printDaemonDoctorStatus(out interface {
	Write([]byte) (int, error)
}, daemon doctorDaemonReport) {
	if !daemon.Running {
		fmt.Fprintln(out, "Daemon:        not running")
		fmt.Fprintln(out, "Next steps:    run `starclaw app` to start the GUI, or `starclaw daemon start` for daemon-only mode")
		return
	}

	fmt.Fprintln(out, "Daemon:        running")
	if daemon.Status != nil {
		status := *daemon.Status
		if status.Version != "" {
			fmt.Fprintf(out, "Daemon version:%s\n", doctorPaddedValue(status.Version))
		}
		fmt.Fprintf(out, "Active agents: %d\n", status.ActiveAgents)
		fmt.Fprintf(out, "Uptime:        %s\n", (time.Duration(status.Uptime) * time.Second).String())
	}

	for _, err := range daemon.Errors {
		fmt.Fprintf(out, "Warning:       %s\n", err)
	}
	if daemon.Diagnostics == nil {
		return
	}
	diagnostics := *daemon.Diagnostics
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

func doctorLocalChecks(results []tui.CheckResult) []doctorLocalCheck {
	checks := make([]doctorLocalCheck, 0, len(results))
	for _, result := range results {
		checks = append(checks, doctorLocalCheck{
			Name:    result.Name,
			Status:  doctorLocalStatusLabel(result.Status),
			Message: result.Message,
		})
	}
	return checks
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
