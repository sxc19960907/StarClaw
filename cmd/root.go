package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/tools"
	"github.com/starclaw/starclaw/internal/tui"
)

var (
	Version       = "dev"
	autoApprove   = false
	resumeSession string
	listSessions  bool
	rootCmd       = &cobra.Command{
		Use:   "starclaw",
		Short: "StarClaw - AI Agent CLI",
		Long:  "AI-powered CLI agent with local tools, configuration management, and session support.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if setup is needed
			cfg, err := config.Load()
			if err != nil || config.NeedsSetup(cfg) {
				fmt.Println("未检测到 API Key，启动配置向导...")
				_, err := config.RunSetup()
				return err
			}

			// If no arguments and stdin is TTY, show info
			if len(args) == 0 && isTTY() {
				fmt.Println("StarClaw is configured and ready!")
				fmt.Printf("  Endpoint: %s\n", cfg.Endpoint)
				fmt.Printf("  Model: %s\n", cfg.ModelTier)
				fmt.Println()
				fmt.Println("Use 'starclaw chat <query>' for one-shot mode")
				fmt.Println("Or run 'starclaw --help' for more options")
				return nil
			}

			// Handle piped input
			if !isTTY() {
				return runPipedMode(cfg)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(sessionsCmd)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "yes", "y", false, "Automatically approve all tool calls")
	rootCmd.PersistentFlags().StringVar(&resumeSession, "resume", "", "Resume a previous session by ID")
	rootCmd.PersistentFlags().BoolVar(&listSessions, "list-sessions", false, "List all saved sessions")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "starclaw version %s\n", Version)
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run interactive setup to configure StarClaw",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.RunSetup()
		return err
	},
}

var chatCmd = &cobra.Command{
	Use:   "chat [query]",
	Short: "Chat with the AI agent",
	Long:  "Send a single query to the AI agent and get a response.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil || config.NeedsSetup(cfg) {
			fmt.Println("Configuration required. Run 'starclaw setup'")
			return fmt.Errorf("not configured")
		}

		query := strings.Join(args, " ")
		return runChat(cfg, query)
	},
}

// runChat executes a chat query
func runChat(cfg *config.Config, query string) error {
	ctx := context.Background()

	// Create LLM client
	model := os.Getenv("ANTHROPIC_MODEL")
	llmClient := client.NewLLMClient(cfg.APIKey, cfg.Endpoint, model)

	// Create tool registry
	registry := tools.RegisterLocalTools()

	// Create agent loop
	loop := agent.NewAgentLoop(llmClient, registry)
	loop.SetMaxIterations(cfg.Agent.MaxIterations)
	loop.SetMaxTokens(cfg.Agent.MaxTokens)
	loop.SetResultTruncation(cfg.Tools.ResultTruncation)

	// Set up session management
	sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
	sessionMgr := session.NewManager(sessionsDir)

	// Determine which session to use
	var sess *session.Session
	var err error
	if resumeSession != "" {
		// Resume specific session
		sess, err = sessionMgr.Resume(resumeSession)
		if err != nil {
			return fmt.Errorf("failed to resume session %s: %w", resumeSession, err)
		}
		fmt.Printf("📂 Resuming session: %s\n\n", sess.Title)
	} else {
		// Create new session
		sess = sessionMgr.NewSession()
	}

	loop.SetSession(sess)
	loop.SetSessionManager(sessionMgr)

	// Set up audit logging
	if cfg.Audit.Enabled {
		logDir := filepath.Join(config.StarclawDir(), "logs")
		auditLogger, err := audit.NewAuditLogger(logDir)
		if err != nil {
			// Log the error but don't fail - audit logging is non-critical
			fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)
		} else {
			loop.SetAuditLogger(auditLogger)
			loop.SetSessionID(sess.ID)
			defer auditLogger.Close()
		}
	}

	// Set system prompt
	loop.SetSystemPrompt(buildSystemPrompt(registry))

	// Create event handler
	handler := &CLIEventHandler{
		autoApprove: autoApprove,
	}
	loop.SetEventHandler(handler)

	// Run the conversation
	fmt.Printf("🤔 Thinking...\n\n")
	resp, err := loop.Run(ctx, query)
	if err != nil {
		return fmt.Errorf("conversation failed: %w", err)
	}

	// Print final response
	if resp.Content != "" {
		fmt.Println(resp.Content)
	}

	// Print usage
	fmt.Printf("\n📊 Usage: %d input tokens, %d output tokens\n",
		resp.Usage.InputTokens, resp.Usage.OutputTokens)

	// Save session on exit
	if err := sessionMgr.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	return nil
}

// runPipedMode handles piped input
func runPipedMode(cfg *config.Config) error {
	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	query := strings.TrimSpace(string(data))
	if query == "" {
		return fmt.Errorf("empty input")
	}

	return runChat(cfg, query)
}

// isTTY checks if stdin is a terminal
func isTTY() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fileInfo.Mode()&os.ModeCharDevice != 0
}

// CLIEventHandler handles events for CLI mode
type CLIEventHandler struct {
	autoApprove bool
}

func (h *CLIEventHandler) OnToolCall(name string, args string) {
	fmt.Printf("🔧 Tool: %s\n", name)
	if !h.autoApprove {
		fmt.Printf("   Args: %s\n", truncateString(args, 100))
	}
}

func (h *CLIEventHandler) OnToolResult(name string, result agent.ToolResult) {
	if result.IsError {
		fmt.Printf("   ❌ Error: %s\n", truncateString(result.Content, 100))
	} else {
		fmt.Printf("   ✅ Done\n")
	}
}

func (h *CLIEventHandler) OnText(text string) {
	// Text is printed at the end
}

func (h *CLIEventHandler) OnUsage(usage client.Usage) {
	// Usage is printed at the end
}

// buildSystemPrompt builds the system prompt with tool descriptions
func buildSystemPrompt(registry *agent.ToolRegistry) string {
	var sb strings.Builder
	sb.WriteString("You are StarClaw, an AI assistant with access to local tools.\n\n")
	sb.WriteString("Available tools:\n")

	tools := registry.List()
	for _, tool := range tools {
		info := tool.Info()
		sb.WriteString(fmt.Sprintf("- %s: %s\n", info.Name, info.Description))
	}

	sb.WriteString("\nWhen facing complex multi-step tasks, use the `think` tool first to plan your approach.")
	sb.WriteString("\nUse the tools when appropriate to help the user.")
	return sb.String()
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func Execute(version string) {
	Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// sessionsCmd lists all saved sessions
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List all saved sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}

		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		sessionMgr := session.NewManager(sessionsDir)

		summaries, err := sessionMgr.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(summaries) == 0 {
			fmt.Println("No saved sessions found.")
			return nil
		}

		fmt.Printf("%-30s  %-30s  %10s  %s\n", "ID", "Title", "Messages", "Date")
		fmt.Println(strings.Repeat("-", 100))
		for _, s := range summaries {
			// Truncate ID and title for display
			id := s.ID
			if len(id) > 28 {
				id = id[:25] + "..."
			}
			title := s.Title
			if len(title) > 28 {
				title = title[:25] + "..."
			}
			fmt.Printf("%-30s  %-30s  %10d  %s\n",
				id,
				title,
				s.MsgCount,
				s.CreatedAt.Format("2006-01-02"))
		}
		return nil
	},
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive TUI mode",
	Long:  "Start an interactive chat session with the AI agent using a terminal UI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil || config.NeedsSetup(cfg) {
			fmt.Println("Configuration required. Run 'starclaw setup'")
			return fmt.Errorf("not configured")
		}

		// Create LLM client
		model := os.Getenv("ANTHROPIC_MODEL")
		llmClient := client.NewLLMClient(cfg.APIKey, cfg.Endpoint, model)

		// Create tool registry
		registry := tools.RegisterLocalTools()

		// Create agent loop
		loop := agent.NewAgentLoop(llmClient, registry)
		loop.SetMaxIterations(cfg.Agent.MaxIterations)
		loop.SetMaxTokens(cfg.Agent.MaxTokens)
		loop.SetResultTruncation(cfg.Tools.ResultTruncation)

		// Set up session management
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		sessionMgr := session.NewManager(sessionsDir)

		// Determine which session to use
		var sess *session.Session
		if resumeSession != "" {
			// Resume specific session
			sess, err = sessionMgr.Resume(resumeSession)
			if err != nil {
				return fmt.Errorf("failed to resume session %s: %w", resumeSession, err)
			}
			fmt.Printf("📂 Resuming session: %s\n", sess.Title)
		} else {
			// Create new session
			sess = sessionMgr.NewSession()
		}

		loop.SetSession(sess)
		loop.SetSessionManager(sessionMgr)

		// Set up audit logging
		if cfg.Audit.Enabled {
			logDir := filepath.Join(config.StarclawDir(), "logs")
			auditLogger, err := audit.NewAuditLogger(logDir)
			if err != nil {
				// Log the error but don't fail - audit logging is non-critical
				fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)
			} else {
				loop.SetAuditLogger(auditLogger)
				loop.SetSessionID(sess.ID)
				defer auditLogger.Close()
			}
		}

		// Set system prompt
		loop.SetSystemPrompt(buildSystemPrompt(registry))

		// Launch TUI
		return tui.Run(loop)
	},
}
