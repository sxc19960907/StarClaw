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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/daemon"
	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
	"github.com/starclaw/starclaw/internal/schedule"
	"github.com/starclaw/starclaw/internal/tools"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Background daemon server with HTTP API and cron scheduler",
}

const defaultDaemonPort = 7533

var daemonPort = daemonPortFromEnv()
var daemonWebURL = daemonWebURLForPort(daemonPort)
var daemonHealthURL = fmt.Sprintf("http://127.0.0.1:%d/health", daemonPort)
var daemonStatusURL = fmt.Sprintf("http://127.0.0.1:%d/status", daemonPort)
var daemonVersionURL = fmt.Sprintf("http://127.0.0.1:%d/version", daemonPort)
var daemonDiagnosticsURL = fmt.Sprintf("http://127.0.0.1:%d/diagnostics", daemonPort)
var daemonEnsureTimeout = 5 * time.Second
var daemonHealthPollInterval = 120 * time.Millisecond
var daemonRPCSocketPath string
var daemonRPCPidfilePath string

func daemonWebURLForPort(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/app/", port)
}

func daemonPortFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("STARCLAW_DAEMON_PORT"))
	if raw == "" {
		return defaultDaemonPort
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return defaultDaemonPort
	}
	return port
}

type daemonIdentity struct {
	Version       string `json:"version"`
	WebURL        string `json:"web_url"`
	HealthURL     string `json:"health_url"`
	LaunchCommand string `json:"launch_command"`
}

func isExpectedDaemonIdentity(identity daemonIdentity) bool {
	return strings.TrimSpace(identity.LaunchCommand) == "starclaw app" &&
		strings.TrimSpace(identity.WebURL) == daemonWebURL &&
		strings.TrimSpace(identity.HealthURL) == daemonHealthURL
}

var fetchDaemonIdentity = func(ctx context.Context) (daemonIdentity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, daemonVersionURL, nil)
	if err != nil {
		return daemonIdentity{}, err
	}
	resp, err := (&http.Client{Timeout: 500 * time.Millisecond}).Do(req)
	if err != nil {
		return daemonIdentity{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return daemonIdentity{}, fmt.Errorf("unexpected version status: %s", resp.Status)
	}
	var identity daemonIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return daemonIdentity{}, fmt.Errorf("decode daemon identity: %w", err)
	}
	if !isExpectedDaemonIdentity(identity) {
		return daemonIdentity{}, errors.New("local service is not a StarClaw daemon")
	}
	return identity, nil
}

var isDaemonHealthy = func(ctx context.Context) bool {
	_, err := fetchDaemonIdentity(ctx)
	return err == nil
}

func isDaemonPortOccupied(ctx context.Context) bool {
	if isDaemonHealthy(ctx) {
		return false
	}
	for _, url := range []string{daemonVersionURL, daemonHealthURL} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		resp, err := (&http.Client{Timeout: 500 * time.Millisecond}).Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		return true
	}
	return false
}

func requireOwnedDaemon(ctx context.Context) error {
	if isDaemonHealthy(ctx) {
		return nil
	}
	if isDaemonPortOccupied(ctx) {
		return fmt.Errorf("port %d is in use by another local service, not StarClaw", daemonPort)
	}
	return fmt.Errorf("daemon is not reachable at %s", daemonVersionURL)
}

func postDaemonShutdown(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/shutdown", daemonPort), nil)
	if err != nil {
		return nil, err
	}
	return (&http.Client{Timeout: 2 * time.Second}).Do(req)
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
	if isDaemonPortOccupied(ctx) {
		return false, fmt.Errorf("port %d is in use by another local service, not StarClaw", daemonPort)
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
	case strings.Contains(err.Error(), "another local service"):
		message += "; close the service using this port or set STARCLAW_DAEMON_PORT to an unused port"
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
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started daemon and opened %s\n", daemonWebURL)
	case ensure && openBrowser:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Daemon already running. Opened %s\n", daemonWebURL)
	case ensure && started:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started daemon. Web UI: %s\n", daemonWebURL)
	case ensure:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Daemon already running. Web UI: %s\n", daemonWebURL)
	default:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Opened %s\n", daemonWebURL)
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
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "StarClaw app launch readiness\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Version:       %s\n", Version)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Launch:        starclaw app\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Daemon:        %s\n", status)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Web UI:        %s\n", daemonWebURL)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Health:        %s\n", daemonHealthURL)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status API:    %s\n", daemonStatusURL)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Diagnostics:   %s\n", daemonDiagnosticsURL)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Data:          %s\n", config.StarclawDir())
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config:        %s\n", filepath.Join(config.StarclawDir(), "config.yaml"))
	return nil
}

func validateDesktopRPCLaunchFlags(sockPath, pidfilePath string) (bool, error) {
	sockPath = strings.TrimSpace(sockPath)
	pidfilePath = strings.TrimSpace(pidfilePath)
	switch {
	case sockPath == "" && pidfilePath == "":
		return false, nil
	case sockPath == "":
		return false, errors.New("--rpc-socket is required when --rpc-pidfile is set")
	case pidfilePath == "":
		return false, errors.New("--rpc-pidfile is required when --rpc-socket is set")
	default:
		return true, nil
	}
}

func startDaemonDesktopRPCListener(ctx context.Context, srv *daemon.Server, sockPath, pidfilePath string, cancel context.CancelFunc) error {
	enabled, err := validateDesktopRPCLaunchFlags(sockPath, pidfilePath)
	if err != nil || !enabled {
		return err
	}
	readyCh := make(chan struct{})
	listener := desktop_rpc.NewListener(desktop_rpc.ListenerConfig{
		SockPath:    sockPath,
		PidfilePath: pidfilePath,
		Platform:    desktop_rpc.DefaultPlatform(Version),
		Broker:      srv.DesktopRPCBroker(),
		EventSink:   srv.RecordDesktopEvent,
		ReadyCh:     readyCh,
	})
	srv.SetDesktopRPCListener(listener)

	errCh := make(chan error, 1)
	go func() {
		errCh <- listener.Run(ctx)
	}()

	select {
	case <-readyCh:
		go func() {
			if err := <-errCh; err != nil && ctx.Err() == nil {
				log.Printf("daemon: desktop rpc listener stopped: %v", err)
				cancel()
			}
		}()
		return nil
	case err := <-errCh:
		if err == nil {
			return errors.New("desktop rpc listener stopped before ready")
		}
		return fmt.Errorf("desktop rpc listener: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("desktop rpc listener: %w", ctx.Err())
	}
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon server",
	RunE: func(cmd *cobra.Command, args []string) error {
		desktopRPCEnabled, err := validateDesktopRPCLaunchFlags(daemonRPCSocketPath, daemonRPCPidfilePath)
		if err != nil {
			return err
		}

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
			StarclawDir:      starclawDir,
			ConfigPath:       filepath.Join(starclawDir, "config.yaml"),
			Config:           cfg,
			AgentsDir:        agentsDir,
			SkillsDir:        skillsDir,
			InstructionsDir:  instructionsDir,
			LLMClient:        llmClient,
			LLMClientFactory: newLLMClient,
			Registry:         registry,
			ScheduleManager:  scheduleMgr,
		}

		srv := daemon.NewServer(daemonPort, deps, Version)
		srv.SetCancelFunc(cancel)
		if desktopRPCEnabled {
			if err := startDaemonDesktopRPCListener(ctx, srv, daemonRPCSocketPath, daemonRPCPidfilePath, cancel); err != nil {
				return err
			}
			log.Printf("daemon: desktop rpc listening")
		}

		// Start cron scheduler.
		sched := daemon.NewScheduler(scheduleMgr, deps)
		go sched.Start(ctx)
		log.Println("daemon: cron scheduler started")

		// Start HTTP server (blocks until ctx is cancelled).
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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := requireOwnedDaemon(ctx); err != nil {
			return fmt.Errorf("daemon: %w", err)
		}
		resp, err := postDaemonShutdown(ctx)
		if err != nil {
			return fmt.Errorf("daemon: not reachable: %w", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := requireOwnedDaemon(ctx); err != nil {
			fmt.Println("Daemon is not running.")
			fmt.Printf("Detail:        %v\n", err)
			return nil
		}
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/status", daemonPort))
		if err != nil {
			fmt.Println("Daemon is not running.")
			return nil
		}
		defer func() {
			_ = resp.Body.Close()
		}()

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
	daemonStartCmd.Flags().StringVar(&daemonRPCSocketPath, "rpc-socket", "", "Desktop RPC Unix socket path; requires --rpc-pidfile")
	daemonStartCmd.Flags().StringVar(&daemonRPCPidfilePath, "rpc-pidfile", "", "Desktop RPC pidfile path; requires --rpc-socket")
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonOpenCmd)
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(appCmd)
}
