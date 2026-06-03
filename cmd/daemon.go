package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/daemon"
	"github.com/starclaw/starclaw/internal/schedule"
	"github.com/starclaw/starclaw/internal/tools"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Background daemon server with HTTP API and cron scheduler",
}

const daemonPort = 7533

var daemonWebURL = daemonWebURLForPort(daemonPort)
var daemonHealthURL = fmt.Sprintf("http://127.0.0.1:%d/health", daemonPort)
var daemonDiagnosticsURL = fmt.Sprintf("http://127.0.0.1:%d/diagnostics", daemonPort)
var daemonEnsureTimeout = 5 * time.Second
var daemonHealthPollInterval = 120 * time.Millisecond

func daemonWebURLForPort(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/app/", port)
}

var isDaemonHealthy = func(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, daemonHealthURL, nil)
	if err != nil {
		return false
	}
	resp, err := (&http.Client{Timeout: 500 * time.Millisecond}).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

var startDaemonBackground = func() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open dev null: %w", err)
	}
	cmd := exec.Command(exe, "daemon", "start")
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	if err := cmd.Start(); err != nil {
		_ = devNull.Close()
		return fmt.Errorf("start process: %w", err)
	}
	go func() {
		_ = cmd.Wait()
		_ = devNull.Close()
	}()
	return nil
}

func waitForDaemonHealth(ctx context.Context) error {
	timer := time.NewTimer(daemonEnsureTimeout)
	defer timer.Stop()
	ticker := time.NewTicker(daemonHealthPollInterval)
	defer ticker.Stop()

	for {
		if isDaemonHealthy(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for daemon health: %w", ctx.Err())
		case <-timer.C:
			return fmt.Errorf("daemon did not become healthy within %s", daemonEnsureTimeout)
		case <-ticker.C:
		}
	}
}

func ensureDaemonRunning(ctx context.Context) (bool, error) {
	if isDaemonHealthy(ctx) {
		return false, nil
	}
	if err := startDaemonBackground(); err != nil {
		return false, fmt.Errorf("start background daemon: %w", err)
	}
	if err := waitForDaemonHealth(ctx); err != nil {
		return true, err
	}
	return true, nil
}

func formatDaemonLaunchError(err error) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf("daemon: %v", err)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		message += fmt.Sprintf("; timed out waiting for %s", daemonHealthURL)
	case strings.Contains(err.Error(), "address already in use"):
		message += fmt.Sprintf("; port %d appears to be in use", daemonPort)
	}
	message += fmt.Sprintf("; run `starclaw daemon status`, check whether port %d is free, or inspect %s", daemonPort, daemonDiagnosticsURL)
	return errors.New(message)
}

func openDaemonWebUI(cmd *cobra.Command, ensure bool) error {
	return launchDaemonWebUI(cmd, ensure, true)
}

func launchDaemonWebUI(cmd *cobra.Command, ensure bool, openBrowser bool) error {
	started := false
	if ensure {
		ctx, cancel := context.WithTimeout(context.Background(), daemonEnsureTimeout+time.Second)
		defer cancel()
		var err error
		started, err = ensureDaemonRunning(ctx)
		if err != nil {
			return formatDaemonLaunchError(err)
		}
	}
	if openBrowser {
		if err := openURLInBrowser(daemonWebURL); err != nil {
			return fmt.Errorf("daemon: open web UI: %w; daemon is reachable, open %s manually", err, daemonWebURL)
		}
	}
	switch {
	case ensure && started && openBrowser:
		fmt.Fprintf(cmd.OutOrStdout(), "Started daemon and opened %s\n", daemonWebURL)
	case ensure && openBrowser:
		fmt.Fprintf(cmd.OutOrStdout(), "Daemon already running. Opened %s\n", daemonWebURL)
	case ensure && started:
		fmt.Fprintf(cmd.OutOrStdout(), "Started daemon. Web UI: %s\n", daemonWebURL)
	case ensure:
		fmt.Fprintf(cmd.OutOrStdout(), "Daemon already running. Web UI: %s\n", daemonWebURL)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Opened %s\n", daemonWebURL)
	}
	return nil
}

func printAppLaunchReadiness(cmd *cobra.Command) error {
	status := "not running"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if isDaemonHealthy(ctx) {
		status = "running"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "StarClaw app launch readiness\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Version:       %s\n", Version)
	fmt.Fprintf(cmd.OutOrStdout(), "Launch:        starclaw app\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Daemon:        %s\n", status)
	fmt.Fprintf(cmd.OutOrStdout(), "Web UI:        %s\n", daemonWebURL)
	fmt.Fprintf(cmd.OutOrStdout(), "Diagnostics:   %s\n", daemonDiagnosticsURL)
	fmt.Fprintf(cmd.OutOrStdout(), "Data:          %s\n", config.StarclawDir())
	return nil
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		starclawDir := config.StarclawDir()
		agentsDir := filepath.Join(starclawDir, "agents")
		skillsDir := filepath.Join(starclawDir, "skills")
		instructionsDir := filepath.Join(starclawDir, "instructions")

		// Create LLM client.
		llmClient := newLLMClient(cfg)

		// Create tool registry.
		registry := tools.RegisterLocalTools()
		tools.RegisterVersionTool(registry, Version)

		// Create schedule manager.
		scheduleMgr := schedule.NewManager(filepath.Join(starclawDir, "schedules.json"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle SIGINT/SIGTERM for graceful shutdown.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			log.Println("daemon: shutting down...")
			cancel()
		}()

		deps := &daemon.ServerDeps{
			StarclawDir:     starclawDir,
			ConfigPath:      filepath.Join(starclawDir, "config.yaml"),
			Config:          cfg,
			AgentsDir:       agentsDir,
			SkillsDir:       skillsDir,
			InstructionsDir: instructionsDir,
			LLMClient:       llmClient,
			Registry:        registry,
			ScheduleManager: scheduleMgr,
		}

		// Start cron scheduler.
		sched := daemon.NewScheduler(scheduleMgr, deps)
		go sched.Start(ctx)
		log.Println("daemon: cron scheduler started")

		// Start HTTP server (blocks until ctx is cancelled).
		srv := daemon.NewServer(daemonPort, deps, Version)
		srv.SetCancelFunc(cancel)
		log.Printf("daemon: starting server on localhost:%d", daemonPort)
		log.Printf("daemon: web UI available at %s", daemonWebURL)
		if err := srv.Start(ctx); err != nil {
			return fmt.Errorf("daemon: server error: %w", err)
		}
		return nil
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon server",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/shutdown", daemonPort), "application/json", nil)
		if err != nil {
			return fmt.Errorf("daemon: not reachable: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("daemon: unexpected response: %s", resp.Status)
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("daemon: failed to parse response: %w", err)
		}
		fmt.Printf("Daemon: %s\n", result["status"])
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/status", daemonPort))
		if err != nil {
			fmt.Println("Daemon is not running.")
			return nil
		}
		defer resp.Body.Close()

		var status struct {
			Uptime       int    `json:"uptime"`
			Version      string `json:"version"`
			ActiveAgents int    `json:"active_agents"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return fmt.Errorf("daemon: failed to parse status: %w", err)
		}

		fmt.Printf("Status:        running\n")
		if status.Version != "" {
			fmt.Printf("Version:       %s\n", status.Version)
		}
		fmt.Printf("Active agents: %d\n", status.ActiveAgents)
		uptime := time.Duration(status.Uptime) * time.Second
		fmt.Printf("Uptime:        %s\n", uptime)
		fmt.Printf("Web UI:        %s\n", daemonWebURL)
		return nil
	},
}

var openURLInBrowser = func(url string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
		args = []string{url}
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		command = "xdg-open"
		args = []string{url}
	}
	return exec.Command(command, args...).Start()
}

func newDaemonOpenCmd() *cobra.Command {
	var start bool
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open the daemon Web UI in the default browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return openDaemonWebUI(cmd, start)
		},
	}
	cmd.Flags().BoolVar(&start, "start", false, "Start the daemon first if it is not running")
	return cmd
}

var daemonOpenCmd = newDaemonOpenCmd()

func newAppCmd() *cobra.Command {
	var check bool
	var noOpen bool
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Start the daemon if needed and open the Web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if check {
				return printAppLaunchReadiness(cmd)
			}
			return launchDaemonWebUI(cmd, true, !noOpen)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "Print app launch readiness without starting the daemon or opening a browser")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Start or reuse the daemon and print the Web UI URL without opening a browser")
	return cmd
}

var appCmd = newAppCmd()

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonOpenCmd)
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(appCmd)
}
