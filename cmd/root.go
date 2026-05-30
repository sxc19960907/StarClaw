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
	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/hooks"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/tools"
	"github.com/starclaw/starclaw/internal/tui"
	"github.com/starclaw/starclaw/internal/update"
)

var sessionsExportFormat string
var sessionsExportOutput string

var (
	Version       = "dev"
	autoApprove   = false
	resumeSession string
	listSessions  bool
	agentName     string // New: --agent flag
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

			// Check for updates in background (non-blocking)
			if cfg.Update.AutoCheck && update.IsSemver(Version) {
				go func() {
					cacheDir := config.StarclawDir()
					if msg := update.AutoUpdate(Version, cacheDir); msg != "" {
						fmt.Printf("\n📦 %s\n\n", msg)
					}
				}()
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
	rootCmd.AddCommand(mcpCmd)        // New: MCP subcommand
	rootCmd.AddCommand(updateCmd)     // New: Update subcommand
	rootCmd.AddCommand(completionCmd) // Shell completion

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "yes", "y", false, "Automatically approve all tool calls")
	rootCmd.PersistentFlags().StringVar(&resumeSession, "resume", "", "Resume a previous session by ID")
	rootCmd.PersistentFlags().BoolVar(&listSessions, "list-sessions", false, "List all saved sessions")
	rootCmd.PersistentFlags().StringVar(&agentName, "agent", "", "Use named agent configuration")
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
	llmClient := newLLMClient(cfg)

	// Load named agent if --agent flag is set
	var agentOverride *agents.Agent
	if agentName != "" {
		agentsDir := filepath.Join(config.StarclawDir(), "agents")
		ag, err := agents.LoadAgent(agentsDir, agentName)
		if err != nil {
			return fmt.Errorf("agent %q: %w", agentName, err)
		}
		agentOverride = ag
		// Ensure agent sessions directory exists
		if err := os.MkdirAll(filepath.Join(agentsDir, agentName, "sessions"), 0700); err != nil {
			return fmt.Errorf("create agent sessions dir: %w", err)
		}
		// Apply agent config overrides
		cfg = config.MergeAgentConfig(cfg, ag)
	}

	// Create tool registry
	registry := tools.RegisterLocalTools(cfg.Tools)
	tools.RegisterVersionTool(registry, Version)

	// Create agent loop
	loop := agent.NewAgentLoop(llmClient, registry)
	loop.SetMaxIterations(cfg.Agent.MaxIterations)
	loop.SetMaxTokens(cfg.Agent.MaxTokens)
	loop.SetResultTruncation(cfg.Tools.ResultTruncation)
	loop.SetConfigDir(config.StarclawDir())
	loop.SetContextWindow(cfg.Agent.ContextWindow)
	loop.SetPermissions(cfg.Permissions)
	applyAgentOptions(loop, cfg)
	if cfg.Hooks != nil {
		loop.SetHookRunner(hooks.NewRunner(*cfg.Hooks))
	}

	// Set up session management (agent-scoped or global)
	var baseDir string
	if agentOverride != nil {
		baseDir = filepath.Join(config.StarclawDir(), "agents", agentName)
	} else {
		baseDir = config.StarclawDir()
	}
	sessionsDir := filepath.Join(baseDir, "sessions")
	sessionMgr := session.NewManager(sessionsDir)
	tools.RegisterSessionSearch(registry, sessionMgr)

	// Determine which session to use
	var sess *session.Session
	if resumeSession != "" {
		// Resume specific session
		var err error
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
			fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)
		} else {
			loop.SetAuditLogger(auditLogger)
			loop.SetSessionID(sess.ID)
			defer auditLogger.Close()
		}
	}

	// Set system prompt
	loop.SetSystemPrompt(buildSystemPrompt(registry))

	// Inject agent if loaded
	if agentOverride != nil {
		agentDir := filepath.Join(config.StarclawDir(), "agents", agentName)
		loop.SwitchAgent(agentOverride.Prompt, agentDir)
		if agentOverride.Memory != "" {
			loop.SetMemory(agentOverride.Memory)
		}
	}

	// Create event handler
	autoApproveEffective := autoApprove || (agentOverride != nil && agentOverride.AutoApproveEnabled())
	handler := &CLIEventHandler{
		autoApprove: autoApproveEffective,
	}
	loop.SetEventHandler(handler)

	// Run the conversation
	fmt.Printf("🤔 Thinking...\n\n")
	resp, err := loop.Run(ctx, query)
	if err != nil {
		return fmt.Errorf("conversation failed: %w", err)
	}

	// Print final response
	if resp.Content != "" && !handler.streamed {
		fmt.Println(resp.Content)
	} else if handler.streamed {
		fmt.Println()
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

	// Prepend CWD context for pipe mode so the agent knows where it is working
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Only prepend CWD if input is not already a specific path reference
	if !isPathReference(query) {
		query = fmt.Sprintf("Current directory: %s\n\n%s", cwd, query)
	}

	// Increase max_iterations for pipe mode (batch processing benefits from more steps)
	cfg.Agent.MaxIterations = max(cfg.Agent.MaxIterations, 50)

	return runChat(cfg, query)
}

// isPathReference checks if the input starts with a path-like prefix
func isPathReference(s string) bool {
	return strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "~/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../")
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
	streamed    bool
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

func (h *CLIEventHandler) OnStreamDelta(delta string) {
	// Streamed text is printed inline
	h.streamed = true
	fmt.Print(delta)
}

func (h *CLIEventHandler) OnPreamble(preamble string) {
	// No-op for CLI
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

// newLLMClient creates the appropriate LLM client based on config provider.
func newLLMClient(cfg *config.Config) client.LLMClient {
	switch cfg.Provider {
	case "openai":
		return client.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIEndpoint, cfg.OpenAIModel)
	case "ollama":
		return client.NewOllamaClient(cfg.OllamaEndpoint, cfg.OllamaModel)
	default:
		return client.NewAnthropicClient(cfg.APIKey, cfg.Endpoint, cfg.ModelTier)
	}
}

func applyAgentOptions(loop *agent.AgentLoop, cfg *config.Config) {
	loop.SetThinking(agent.ThinkingConfigFromAgent(cfg.Agent))
	loop.SetReasoningEffort(cfg.Agent.ReasoningEffort)
	loop.SetSpecificModel(cfg.Agent.Model)
	loop.SetEnableStreaming(true)
}

func Execute(version string) {
	Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// sessionsCmd and its subcommands manage saved sessions
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List and manage saved sessions",
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

		fmt.Printf("%-30s  %-34s  %10s  %s\n", "ID", "Title", "Messages", "Date")
		fmt.Println(strings.Repeat("-", 110))
		for _, s := range summaries {
			id := s.ID
			if len(id) > 28 {
				id = id[:25] + "..."
			}

			// Build title with tags and favorite marker
			title := s.Title
			if len(title) > 22 {
				title = title[:19] + "..."
			}
			var suffix string
			if len(s.Tags) > 0 {
				suffix = " [" + strings.Join(s.Tags, ", ") + "]"
			}
			if s.Favorite {
				suffix = " ★" + suffix
			}
			titleDisplay := title + suffix
			if len(titleDisplay) > 34 {
				titleDisplay = titleDisplay[:31] + "..."
			}

			fmt.Printf("%-30s  %-34s  %10d  %s\n",
				id,
				titleDisplay,
				s.MsgCount,
				s.CreatedAt.Format("2006-01-02"))
		}
		return nil
	},
}

// sessionsTagCmd adds a tag to a session
var sessionsTagCmd = &cobra.Command{
	Use:   "tag <id> <tag>",
	Short: "Add a tag to a session",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		mgr := session.NewManager(sessionsDir)
		return mgr.AddTag(args[0], args[1])
	},
}

// sessionsUntagCmd removes a tag from a session
var sessionsUntagCmd = &cobra.Command{
	Use:   "untag <id> <tag>",
	Short: "Remove a tag from a session",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		mgr := session.NewManager(sessionsDir)
		return mgr.RemoveTag(args[0], args[1])
	},
}

// sessionsFavoriteCmd marks a session as favorite
var sessionsFavoriteCmd = &cobra.Command{
	Use:   "favorite <id>",
	Short: "Mark a session as favorite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		mgr := session.NewManager(sessionsDir)
		return mgr.SetFavorite(args[0], true)
	},
}

// sessionsUnfavoriteCmd unmarks a session as favorite
var sessionsUnfavoriteCmd = &cobra.Command{
	Use:   "unfavorite <id>",
	Short: "Unmark a session as favorite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		mgr := session.NewManager(sessionsDir)
		return mgr.SetFavorite(args[0], false)
	},
}

// sessionsExportCmd exports a session in a specified format
var sessionsExportCmd = &cobra.Command{
	Use:   "export <id>",
	Short: "Export a session (md or html)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.Load()
		if err != nil {
			return err
		}
		sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
		mgr := session.NewManager(sessionsDir)

		content, err := mgr.Export(args[0], sessionsExportFormat)
		if err != nil {
			return err
		}

		if sessionsExportOutput != "" {
			if err := os.WriteFile(sessionsExportOutput, []byte(content), 0600); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
			fmt.Printf("Session exported to %s\n", sessionsExportOutput)
		} else {
			fmt.Println(content)
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

		// Load named agent if --agent flag is set
		var agentOverride *agents.Agent
		if agentName != "" {
			agentsDir := filepath.Join(config.StarclawDir(), "agents")
			ag, err := agents.LoadAgent(agentsDir, agentName)
			if err != nil {
				return fmt.Errorf("agent %q: %w", agentName, err)
			}
			agentOverride = ag
			if err := os.MkdirAll(filepath.Join(agentsDir, agentName, "sessions"), 0700); err != nil {
				return fmt.Errorf("create agent sessions dir: %w", err)
			}
			cfg = config.MergeAgentConfig(cfg, ag)
		}

		// Create LLM client
		llmClient := newLLMClient(cfg)

		// Create tool registry
		registry := tools.RegisterLocalTools(cfg.Tools)
		tools.RegisterVersionTool(registry, Version)
		// Create agent loop
		loop := agent.NewAgentLoop(llmClient, registry)
		loop.SetMaxIterations(cfg.Agent.MaxIterations)
		loop.SetMaxTokens(cfg.Agent.MaxTokens)
		loop.SetResultTruncation(cfg.Tools.ResultTruncation)
		loop.SetConfigDir(config.StarclawDir())
		loop.SetContextWindow(cfg.Agent.ContextWindow)
		loop.SetPermissions(cfg.Permissions)
		applyAgentOptions(loop, cfg)
		if cfg.Hooks != nil {
			loop.SetHookRunner(hooks.NewRunner(*cfg.Hooks))
		}

		// Set up session management (agent-scoped or global)
		var baseDir string
		if agentOverride != nil {
			baseDir = filepath.Join(config.StarclawDir(), "agents", agentName)
		} else {
			baseDir = config.StarclawDir()
		}
		sessionsDir := filepath.Join(baseDir, "sessions")
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

		// Inject agent if loaded
		if agentOverride != nil {
			agentDir := filepath.Join(config.StarclawDir(), "agents", agentName)
			loop.SwitchAgent(agentOverride.Prompt, agentDir)
			if agentOverride.Memory != "" {
				loop.SetMemory(agentOverride.Memory)
			}
		}

		// Launch TUI
		return tui.Run(loop)
	},
}

// mcpCmd manages MCP servers
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) servers",
	Long:  "Commands for managing MCP servers and their configurations.",
}

// mcpListCmd lists configured MCP servers
var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.MCPServers) == 0 {
			fmt.Println("No MCP servers configured.")
			fmt.Println("Add servers to ~/.starclaw/config.yaml")
			return nil
		}

		fmt.Println("Configured MCP servers:")
		for name, server := range cfg.MCPServers {
			status := "enabled"
			if server.Disabled {
				status = "disabled"
			}
			fmt.Printf("  %s: %s [%s]\n", name, server.Command, status)
		}
		return nil
	},
}

// mcpServeCmd starts an MCP stdio server exposing all local tools.
var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP stdio server exposing local tools",
	Long: `Start an MCP (Model Context Protocol) stdio server that exposes
all StarClaw local tools to MCP consumers such as Claude Desktop or
other MCP-compatible clients. The server communicates via JSON-RPC
over stdin/stdout.

The server auto-approves all tool calls — the MCP consumer handles
its own authorization layer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		registry := tools.RegisterLocalTools(cfg.Tools)
		tools.RegisterVersionTool(registry, Version)

		srv := tools.NewMCPServer(registry, "starclaw", Version, tools.MCPServerConfig{
			ToolTimeout: cfg.Tools.ServerToolTimeout,
			ExposeTools: cfg.Tools.MCPExpose,
		})
		return srv.Serve(cmd.Context())
	},
}

// updateCmd checks for updates.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install updates",
	Long:  "Check for new versions of StarClaw and install the latest release for this platform.",
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")

		fmt.Printf("Current version: %s\n", Version)

		// Skip dev builds
		if !update.IsSemver(Version) {
			fmt.Println("Development build - skipping update check.")
			return nil
		}

		if checkOnly {
			fmt.Println("Checking for updates...")
			release, hasUpdate, err := update.CheckForUpdate(Version)
			if err != nil {
				return fmt.Errorf("update check failed: %w", err)
			}
			if !hasUpdate {
				fmt.Println("You're already on the latest version!")
				return nil
			}
			fmt.Printf("Update available: %s\n", release.TagName)
			fmt.Printf("Published: %s\n", release.PublishedAt)
			fmt.Printf("Release URL: %s\n", release.HTMLURL)
			return nil
		}

		fmt.Println("Checking for updates...")
		release, hasUpdate, err := update.CheckForUpdate(Version)
		if err != nil {
			return fmt.Errorf("update check failed: %w", err)
		}
		if !hasUpdate {
			fmt.Println("You're already on the latest version!")
			return nil
		}

		fmt.Printf("Update available: %s\n", release.TagName)
		fmt.Println("Installing update...")

		newVersion, err := update.DoUpdate(Version)
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Printf("Successfully updated to %s!\n", newVersion)
		fmt.Println("Please restart StarClaw to use the new version.")
		return nil
	},
}

func init() {
	// Add subcommands to mcp
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpServeCmd)

	// Add flags to update command
	updateCmd.Flags().BoolP("check", "c", false, "Check only, don't install")

	// Add subcommands to sessions
	sessionsCmd.AddCommand(sessionsTagCmd)
	sessionsCmd.AddCommand(sessionsUntagCmd)
	sessionsCmd.AddCommand(sessionsFavoriteCmd)
	sessionsCmd.AddCommand(sessionsUnfavoriteCmd)
	sessionsCmd.AddCommand(sessionsExportCmd)

	// Add flags to sessions export
	sessionsExportCmd.Flags().StringVarP(&sessionsExportFormat, "format", "f", "md", "Export format (md or html)")
	sessionsExportCmd.Flags().StringVarP(&sessionsExportOutput, "output", "o", "", "Output file path")
}
