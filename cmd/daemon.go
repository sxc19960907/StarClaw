package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
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

var daemonWebURL = fmt.Sprintf("http://127.0.0.1:%d/app/", daemonPort)

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
	return &cobra.Command{
		Use:   "open",
		Short: "Open the daemon Web UI in the default browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := openURLInBrowser(daemonWebURL); err != nil {
				return fmt.Errorf("daemon: open web UI: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Opened %s\n", daemonWebURL)
			return nil
		},
	}
}

var daemonOpenCmd = newDaemonOpenCmd()

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonOpenCmd)
	rootCmd.AddCommand(daemonCmd)
}
