package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	ctxwin "github.com/starclaw/starclaw/internal/context"
	"github.com/starclaw/starclaw/internal/hooks"
	"github.com/starclaw/starclaw/internal/permissions"
	"github.com/starclaw/starclaw/internal/session"
)

// EventHandler handles events from the agent loop
type EventHandler interface {
	OnToolCall(name string, args string)
	OnToolResult(name string, result ToolResult)
	OnText(text string)
	OnUsage(usage client.Usage)
	OnStreamDelta(delta string)
	OnPreamble(preamble string)
}

// RunStatus holds status information from the most recent Run call.
type RunStatus struct {
	Code   string // "context_bloat" or ""
	Detail string // human-readable detail
}

// RunStatusHandler is an optional interface a handler may implement to receive
// turn-level status updates. The agent loop checks for it via a type assertion,
// so handlers that do not implement it simply miss these events with no breakage.
//
// Known codes:
//
//	"context_bloat" — large tool results are dominating context; informational
type RunStatusHandler interface {
	OnRunStatus(code string, detail string)
}

// ApprovalRequest describes a tool call that needs explicit human approval.
type ApprovalRequest struct {
	Tool   string
	Args   string
	Reason string
}

// ApprovalDecision is the response returned by an ApprovalRequester.
type ApprovalDecision string

const (
	ApprovalAllow ApprovalDecision = "allow"
	ApprovalDeny  ApprovalDecision = "deny"
)

// ApprovalRequester handles tool calls that require explicit approval.
type ApprovalRequester interface {
	RequestApproval(ctx context.Context, req ApprovalRequest) (ApprovalDecision, error)
}

// PauseController blocks the loop at cooperative pause points.
type PauseController interface {
	WaitIfPaused(ctx context.Context) error
}

// MemoryPreflightProvider returns private, per-turn memory context to inject
// into the model-facing user message. The returned text must not be persisted
// to the session transcript.
type MemoryPreflightProvider interface {
	PreflightMemory(ctx context.Context, query string) (MemoryPreflightResult, error)
}

// MemoryPreflightResult is the content and content-free metadata from a memory
// preflight attempt.
type MemoryPreflightResult struct {
	Block           string
	Attempted       bool
	Provider        string
	Outcome         string
	Reason          string
	ResultsCount    int
	ContextInjected bool
}

// MemoryPreflightHandler is implemented by event handlers that record
// content-free memory preflight telemetry.
type MemoryPreflightHandler interface {
	OnMemoryPreflight(result MemoryPreflightResult)
}

// StreamingLLMClient is an optional interface for LLM clients that support streaming.
type StreamingLLMClient interface {
	StreamChat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions, onDelta func(delta string)) (*client.Response, error)
}

// AgentLoop manages the conversation with the LLM
type AgentLoop struct {
	llmClient       client.LLMClient
	registry        *ToolRegistry
	maxIter         int
	maxTokens       int
	resultTrunc     int
	handler         EventHandler
	systemPrompt    string
	auditLogger     *audit.AuditLogger
	sessionID       string
	session         *session.Session
	sessionMgr      *session.Manager
	memory          string // agent memory content
	memoryDir       string // directory for persistent memory
	configDir       string // starclaw config dir (~/.starclaw)
	loopDetector    *LoopDetector
	contextWindow   int                 // max context window in tokens (0 = disabled)
	permsConfig     *permissions.Config // tool permission rules
	hookRunner      *hooks.Runner       // lifecycle hook runner
	approver        ApprovalRequester
	pause           PauseController
	memoryPreflight MemoryPreflightProvider

	thinking        *client.ThinkingConfig
	reasoningEffort string
	specificModel   string
	enableStreaming bool
	lastRunStatus   RunStatus
	tokenBudget     TokenBudget
	budgetTracker   *tokenBudgetTracker
}

// NewAgentLoop creates a new agent loop
func NewAgentLoop(llmClient client.LLMClient, registry *ToolRegistry) *AgentLoop {
	return &AgentLoop{
		llmClient:   llmClient,
		registry:    registry,
		maxIter:     25,
		maxTokens:   8192,
		resultTrunc: 30000,
	}
}

// SetMaxIterations sets the maximum number of iterations
func (a *AgentLoop) SetMaxIterations(n int) {
	a.maxIter = n
}

// GetMaxIterations gets the maximum number of iterations
func (a *AgentLoop) GetMaxIterations() int {
	return a.maxIter
}

// SetMaxTokens sets the maximum tokens for responses
func (a *AgentLoop) SetMaxTokens(n int) {
	a.maxTokens = n
}

// SetResultTruncation sets the result truncation limit
func (a *AgentLoop) SetResultTruncation(n int) {
	a.resultTrunc = n
}

// SetEventHandler sets the event handler
func (a *AgentLoop) SetEventHandler(h EventHandler) {
	a.handler = h
}

// SetSystemPrompt sets the system prompt
func (a *AgentLoop) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// SetAuditLogger sets the audit logger (optional)
func (a *AgentLoop) SetAuditLogger(logger *audit.AuditLogger) {
	a.auditLogger = logger
}

// SetSessionID sets the session ID for audit correlation
func (a *AgentLoop) SetSessionID(id string) {
	a.sessionID = id
}

// SetSession sets the current session
func (a *AgentLoop) SetSession(sess *session.Session) {
	a.session = sess
}

// SetSessionManager sets the session manager for auto-save
func (a *AgentLoop) SetSessionManager(mgr *session.Manager) {
	a.sessionMgr = mgr
}

// SwitchAgent injects agent instructions into the system prompt and loads memory.
// agentDir is the path to the agent's directory for memory persistence.
func (a *AgentLoop) SwitchAgent(prompt, agentDir string) {
	if prompt != "" {
		// Prepend agent instructions to the existing system prompt
		a.systemPrompt = prompt + "\n\n" + a.systemPrompt
	}
	if agentDir != "" {
		a.memoryDir = filepath.Join(agentDir, "memory")
		// Load existing memory if present
		a.loadMemory()
	}
}

// SetMemory sets the agent memory content directly.
func (a *AgentLoop) SetMemory(memory string) {
	a.memory = memory
}

// SetMemoryDir sets the memory directory and loads existing memory.
func (a *AgentLoop) SetMemoryDir(dir string) {
	a.memoryDir = dir
	a.loadMemory()
}

// SetConfigDir sets the StarClaw config directory for spill and other features.
func (a *AgentLoop) SetConfigDir(dir string) {
	a.configDir = dir
	a.loopDetector = NewLoopDetector()
}

// SetContextWindow sets the context window size in tokens (0 = disabled).
func (a *AgentLoop) SetContextWindow(tokens int) {
	a.contextWindow = tokens
}

// SetPermissions sets the tool permission rules for this loop.
func (a *AgentLoop) SetPermissions(cfg *permissions.Config) {
	a.permsConfig = cfg
}

// SetApprovalRequester sets the handler for tool calls that require approval.
func (a *AgentLoop) SetApprovalRequester(requester ApprovalRequester) {
	a.approver = requester
}

// SetPauseController sets the cooperative pause controller for this loop.
func (a *AgentLoop) SetPauseController(controller PauseController) {
	a.pause = controller
}

// SetMemoryPreflightProvider sets the optional private memory preflight source.
func (a *AgentLoop) SetMemoryPreflightProvider(provider MemoryPreflightProvider) {
	a.memoryPreflight = provider
}

// SetHookRunner sets the lifecycle hook runner for this loop.
func (a *AgentLoop) SetHookRunner(runner *hooks.Runner) {
	a.hookRunner = runner
}

// SetThinking sets the thinking configuration for LLM requests.
func (a *AgentLoop) SetThinking(cfg *client.ThinkingConfig) {
	a.thinking = cfg
}

// SetReasoningEffort sets the reasoning effort for the LLM.
func (a *AgentLoop) SetReasoningEffort(effort string) {
	a.reasoningEffort = effort
}

// SetSpecificModel sets a specific model override for LLM requests.
func (a *AgentLoop) SetSpecificModel(model string) {
	a.specificModel = model
}

// SetEnableStreaming enables or disables streaming for LLM responses.
func (a *AgentLoop) SetEnableStreaming(enable bool) {
	a.enableStreaming = enable
}

// SetTokenBudget configures per-run token budget tracking and hard-stop behavior.
func (a *AgentLoop) SetTokenBudget(budget TokenBudget) {
	a.tokenBudget = budget
	a.budgetTracker = newTokenBudgetTracker(budget)
}

// LastRunStatus returns the status from the most recent Run call.
// Callers should read it in the same goroutine immediately after Run returns.
func (a *AgentLoop) LastRunStatus() RunStatus {
	return a.lastRunStatus
}

// LastBudgetStatus returns the current budget status from the most recent Run.
func (a *AgentLoop) LastBudgetStatus() TokenBudgetUsage {
	if a.budgetTracker == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	return a.budgetTracker.Status()
}

// SpillCleanupFunc returns a function that cleans up spill files for the current session.
func (a *AgentLoop) SpillCleanupFunc() func() {
	return func() {
		if a.configDir != "" && a.sessionID != "" {
			cleanupSpills(a.configDir, a.sessionID)
		}
	}
}

func (a *AgentLoop) loadMemory() {
	if a.memoryDir == "" {
		return
	}
	files, _ := filepath.Glob(filepath.Join(a.memoryDir, "*.md"))
	if len(files) == 0 {
		return
	}
	var parts []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		parts = append(parts, string(data))
	}
	if len(parts) > 0 {
		a.memory = strings.Join(parts, "\n---\n")
	}
}

// ThinkingConfigFromAgent converts a config.AgentConfig into a client.ThinkingConfig.
// Returns nil if thinking is disabled.
func ThinkingConfigFromAgent(cfg config.AgentConfig) *client.ThinkingConfig {
	if !cfg.Thinking {
		return nil
	}
	tc := &client.ThinkingConfig{}
	switch cfg.ThinkingMode {
	case "adaptive":
		tc.Type = "adaptive"
	case "enabled":
		tc.Type = "enabled"
		tc.BudgetTokens = cfg.ThinkingBudget
	}
	return tc
}

// Run executes the agent loop with the given query
func (a *AgentLoop) Run(ctx context.Context, query string) (*client.Response, error) {
	// Reset run status
	a.lastRunStatus = RunStatus{}
	a.budgetTracker = newTokenBudgetTracker(a.tokenBudget)

	// Inject memory directory into context for memory_append tool
	if a.memoryDir != "" {
		ctx = ctxwin.WithMemoryDir(ctx, a.memoryDir)
	}

	// Initialize messages from session if resuming, or start fresh
	messages := []client.Message{}
	if a.session != nil {
		messages = append(messages, a.session.Messages...)
	}
	modelQuery := query
	if a.memoryPreflight != nil {
		if preflight, err := a.memoryPreflight.PreflightMemory(ctx, query); err == nil {
			if preflight.Block != "" {
				modelQuery = query + "\n\n" + preflight.Block
				preflight.ContextInjected = true
			}
			if a.handler != nil {
				if h, ok := a.handler.(MemoryPreflightHandler); ok {
					h.OnMemoryPreflight(preflight)
				}
			}
		} else if a.handler != nil {
			if h, ok := a.handler.(MemoryPreflightHandler); ok {
				h.OnMemoryPreflight(MemoryPreflightResult{Attempted: true, Outcome: "error", Reason: err.Error()})
			}
		}
	}
	messages = append(messages, client.Message{Role: "user", Content: modelQuery})

	// Update session title if this is the first message
	if a.session != nil && len(a.session.Messages) == 0 {
		a.session.Title = session.GenerateTitle(query)
	}

	// Build ChatOptions from AgentLoop fields
	chatOpts := &client.ChatOptions{
		Thinking:        a.thinking,
		ReasoningEffort: a.reasoningEffort,
	}
	if a.specificModel != "" {
		chatOpts.SpecificModel = a.specificModel
	}

	for i := 0; i < a.maxIter; i++ {
		// Build tools for LLM
		tools := a.buildTools()

		// Build effective system prompt with memory
		effectivePrompt := a.systemPrompt
		if a.memory != "" {
			effectivePrompt = effectivePrompt + "\n\n<agent_memory>\n" + a.memory + "\n</agent_memory>"
		}

		// Compress old tool results to save context
		ctxwin.CompressOldToolResults(messages, 3, 300)

		// Sanitize history: fix malformed messages before compaction
		if len(messages) > ctxwin.MinShapeable() {
			messages = ctxwin.SanitizeHistory(messages)
		}

		// Context window compaction: when messages exceed 85% of contextWindow,
		// persist learnings, generate an LLM summary, and reshape history.
		if a.contextWindow > 0 && len(messages) > ctxwin.MinShapeable() {
			est := ctxwin.EstimateTokens(messages)
			if ctxwin.ShouldCompact(est, 0, a.contextWindow) {
				// Write-before-compact: persist durable learnings to MEMORY.md
				if a.memoryDir != "" {
					if completer, ok := interface{}(a.llmClient).(ctxwin.Completer); ok {
						if pErr := ctxwin.PersistLearnings(ctx, completer, messages, a.memoryDir); pErr != nil {
							fmt.Fprintf(os.Stderr, "[context] persist learnings failed: %v\n", pErr)
						}
					}
				}

				// Generate LLM summary of the conversation so far
				var summary string
				if completer, ok := interface{}(a.llmClient).(ctxwin.Completer); ok {
					var sumErr error
					summary, sumErr = ctxwin.GenerateSummary(ctx, completer, messages)
					if sumErr != nil {
						fmt.Fprintf(os.Stderr, "[context] compaction summary failed: %v\n", sumErr)
					}
				}

				// Shape history with summary injection
				before := len(messages)
				messages = ctxwin.ShapeHistory(messages, summary, a.contextWindow)
				if len(messages) < before {
					fmt.Fprintf(os.Stderr, "[context] compacted: %d → %d messages\n", before, len(messages))
				}
			}
		}

		if status := a.projectBudgetInput(messages); status.Status == TokenBudgetStatusExhausted {
			decision := a.enforceTokenBudget(a.maxTokens)
			if decision.Stop {
				return a.budgetExhaustedResponse(&client.Response{}, decision.Status), nil
			}
		}

		if err := a.waitIfPaused(ctx); err != nil {
			return nil, fmt.Errorf("paused run cancelled: %w", err)
		}

		// Call LLM with retry
		resp, err := a.chatWithRetry(ctx, effectivePrompt, messages, tools, chatOpts)
		if err != nil {
			if isContextTooLargeError(err) && len(messages) > ctxwin.MinShapeable() {
				fmt.Fprintf(os.Stderr, "[agent] context too large, compacting and retrying\n")
				var summary string
				if completer, ok := interface{}(a.llmClient).(ctxwin.Completer); ok {
					summary, _ = ctxwin.GenerateSummary(ctx, completer, messages)
				}
				messages = ctxwin.ShapeHistory(messages, summary, a.contextWindow)
				resp, err = a.chatWithRetry(ctx, effectivePrompt, messages, tools, chatOpts)
			}
			if err != nil {
				return nil, fmt.Errorf("LLM error: %w", err)
			}
		}

		// Report usage
		budgetStatus := a.recordBudgetUsage(resp.Usage)
		if budgetStatus.Status == TokenBudgetStatusExhausted {
			a.lastRunStatus = RunStatus{Code: RunStatusBudgetExhausted, Detail: budgetStatus.Detail}
		}
		if a.handler != nil {
			a.handler.OnUsage(resp.Usage)
		}

		// Handle text response (no tool calls)
		if len(resp.ToolUses) == 0 {
			if a.handler != nil {
				a.handler.OnText(resp.Content)
			}

			// Update session with final messages
			if a.session != nil {
				a.session.Messages = stripPrivateMemoryFromMessages(messages)
				a.session.UpdatedAt = time.Now()
				if a.sessionMgr != nil {
					if err := a.sessionMgr.Save(); err != nil {
						a.lastRunStatus = RunStatus{Code: "session_save_failed", Detail: err.Error()}
					}
				}
			}

			return resp, nil
		}

		// Handle tool calls
		messages = append(messages, client.Message{
			Role:    "assistant",
			Content: a.buildAssistantContent(resp),
		})

		var forceStop bool
		for _, toolUse := range resp.ToolUses {
			if err := a.waitIfPaused(ctx); err != nil {
				return nil, fmt.Errorf("paused run cancelled: %w", err)
			}
			result := a.executeTool(ctx, toolUse)

			messages = append(messages, client.Message{
				Role:    "user",
				Content: a.buildToolResultContent(toolUse, result),
			})

			// Loop detection: check after each tool execution
			if a.loopDetector != nil {
				action, msg := a.loopDetector.Check(toolUse.Name)
				switch action {
				case LoopNudge:
					messages = append(messages, client.Message{Role: "user", Content: msg})
				case LoopForceStop:
					messages = append(messages, client.Message{Role: "user", Content: msg + "\n\nPlease provide your final answer now."})
					forceStop = true
				}
			}
		}

		if forceStop {
			if decision := a.enforceTokenBudget(a.maxTokens); decision.Stop {
				return a.budgetExhaustedResponse(resp, decision.Status), nil
			}
			// Give the LLM one more chance to produce a text-only response
			continue
		}

		// Context bloat detection: check if tool results dominate the context
		if detail := detectContextBloat(messages); detail != "" {
			a.lastRunStatus = RunStatus{Code: "context_bloat", Detail: detail}
			if rs, ok := a.handler.(RunStatusHandler); ok {
				rs.OnRunStatus("context_bloat", detail)
			}
		}

		// Update session after each turn and auto-save
		if a.session != nil {
			a.session.Messages = stripPrivateMemoryFromMessages(messages)
			a.session.UpdatedAt = time.Now()
			if a.sessionMgr != nil {
				if err := a.sessionMgr.Save(); err != nil {
					a.lastRunStatus = RunStatus{Code: "session_save_failed", Detail: err.Error()}
				}
			}
		}

		if decision := a.enforceTokenBudget(a.maxTokens); decision.Stop {
			decision.Status = mergeBudgetDetail(decision.Status, budgetStatus)
			return a.budgetExhaustedResponse(resp, decision.Status), nil
		}
	}

	return nil, fmt.Errorf("reached maximum iterations (%d)", a.maxIter)
}

func stripPrivateMemoryFromMessages(messages []client.Message) []client.Message {
	out := make([]client.Message, len(messages))
	copy(out, messages)
	for i := range out {
		if out[i].Role != "user" {
			continue
		}
		out[i].Content = stripPrivateMemoryBlock(out[i].Content)
	}
	return out
}

func stripPrivateMemoryBlock(content string) string {
	start := strings.Index(content, "<private_memory>")
	if start < 0 {
		return content
	}
	end := strings.Index(content[start:], "</private_memory>")
	if end < 0 {
		return strings.TrimSpace(content[:start])
	}
	end += start + len("</private_memory>")
	return strings.TrimSpace(content[:start] + content[end:])
}

func (a *AgentLoop) waitIfPaused(ctx context.Context) error {
	if a.pause == nil {
		return nil
	}
	return a.pause.WaitIfPaused(ctx)
}

func (a *AgentLoop) recordBudgetUsage(usage client.Usage) TokenBudgetUsage {
	if a.budgetTracker == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	status := a.budgetTracker.AddUsage(usage)
	a.emitBudgetStatus(status)
	return status
}

func (a *AgentLoop) projectBudgetInput(messages []client.Message) TokenBudgetUsage {
	if a.budgetTracker == nil {
		return TokenBudgetUsage{Status: TokenBudgetStatusDisabled}
	}
	status := a.budgetTracker.SetProjectedInput(ctxwin.EstimateTokens(messages))
	a.emitBudgetStatus(status)
	return status
}

func (a *AgentLoop) enforceTokenBudget(projectedOutput int) budgetDecision {
	if a.budgetTracker == nil {
		return budgetDecision{Status: TokenBudgetUsage{Status: TokenBudgetStatusDisabled}}
	}
	decision := a.budgetTracker.Decision(projectedOutput)
	if decision.Stop {
		a.lastRunStatus = RunStatus{Code: RunStatusBudgetExhausted, Detail: decision.Status.Detail}
		a.emitRunStatus(RunStatusBudgetExhausted, decision.Status.Detail)
		a.emitBudgetStatus(decision.Status)
	}
	return decision
}

func (a *AgentLoop) emitRunStatus(code, detail string) {
	if rs, ok := a.handler.(RunStatusHandler); ok {
		rs.OnRunStatus(code, detail)
	}
}

func (a *AgentLoop) emitBudgetStatus(status TokenBudgetUsage) {
	if bh, ok := a.handler.(BudgetStatusHandler); ok {
		bh.OnBudgetStatus(status)
	}
}

// BudgetStatusHandler is optionally implemented by handlers that surface budget state.
type BudgetStatusHandler interface {
	OnBudgetStatus(status TokenBudgetUsage)
}

func (a *AgentLoop) budgetExhaustedResponse(resp *client.Response, status TokenBudgetUsage) *client.Response {
	content := "Token budget exhausted; stopping before the next model call."
	if status.Detail != "" {
		content = content + " " + status.Detail + "."
	}
	return &client.Response{
		Content:    content,
		Usage:      resp.Usage,
		StopReason: RunStatusBudgetExhausted,
	}
}

func mergeBudgetDetail(current, previous TokenBudgetUsage) TokenBudgetUsage {
	if current.Detail == "" {
		current.Detail = previous.Detail
	}
	return current
}

// buildTools converts registry tools to client ToolDef
func (a *AgentLoop) buildTools() []client.ToolDef {
	tools := a.registry.List()
	defs := make([]client.ToolDef, len(tools))
	for i, tool := range tools {
		info := tool.Info()
		defs[i] = client.ToolDef{
			Name:        info.Name,
			Description: info.Description,
			InputSchema: info.Parameters,
		}
	}
	return defs
}

// executeTool executes a tool and returns the result
func (a *AgentLoop) executeTool(ctx context.Context, toolUse client.ToolUse) ToolResult {
	tool, ok := a.registry.Get(toolUse.Name)
	if !ok {
		return ValidationError(fmt.Sprintf("unknown tool: %s", toolUse.Name))
	}

	// Permission check
	needsApproval := a.approver != nil && tool.RequiresApproval()
	approvalReason := "tool requires approval"
	if a.permsConfig != nil {
		decision, reason := permissions.CheckToolCall(toolUse.Name, string(toolUse.Input), a.permsConfig)
		if decision == permissions.Deny {
			return PermissionError(fmt.Sprintf("%s: blocked (%s)", toolUse.Name, reason))
		}
		if decision == permissions.Allow {
			needsApproval = false
		}
		if decision == permissions.Ask {
			needsApproval = true
			approvalReason = reason
		}
	}
	if needsApproval {
		if a.approver == nil {
			return PermissionError(fmt.Sprintf("%s: approval required (%s)", toolUse.Name, approvalReason))
		}
		decision, err := a.approver.RequestApproval(ctx, ApprovalRequest{
			Tool:   toolUse.Name,
			Args:   string(toolUse.Input),
			Reason: approvalReason,
		})
		if err != nil {
			return PermissionError(fmt.Sprintf("%s: approval failed (%v)", toolUse.Name, err))
		}
		if decision != ApprovalAllow {
			return PermissionError(fmt.Sprintf("%s: denied by user", toolUse.Name))
		}
	}

	// Pre-tool hook
	if a.hookRunner != nil {
		if decision, reason := a.hookRunner.RunPreToolUse(ctx, toolUse.Name, string(toolUse.Input), a.sessionID); decision == "deny" {
			return PermissionError(fmt.Sprintf("%s: hook denied (%s)", toolUse.Name, reason))
		}
	}

	// Report tool call
	if a.handler != nil {
		a.handler.OnToolCall(toolUse.Name, string(toolUse.Input))
	}

	// Execute with timing
	start := time.Now()
	result, err := tool.Run(ctx, string(toolUse.Input))
	duration := time.Since(start)

	if err != nil {
		result = ToolResult{
			Content: fmt.Sprintf("error: %v", err),
			IsError: true,
		}
	}

	// Spill large results to disk
	if len(result.Content) > spillThreshold && a.configDir != "" && a.sessionID != "" {
		callID := toolUse.ID
		if preview, err := spillToDisk(a.configDir, a.sessionID, callID, result.Content); err == nil {
			result.Content = preview
		}
	} else if len(result.Content) > a.resultTrunc {
		// Truncate result if needed
		keepHead := a.resultTrunc * 3 / 4
		keepTail := a.resultTrunc / 4
		result.Content = result.Content[:keepHead] +
			fmt.Sprintf("\n\n[... truncated %d chars ...]\n\n", len(result.Content)-a.resultTrunc) +
			result.Content[len(result.Content)-keepTail:]
	}

	// Loop detection: record this tool call
	if a.loopDetector != nil {
		errMsg := ""
		if result.IsError {
			errMsg = result.Content
		}
		a.loopDetector.Record(toolUse.Name, string(toolUse.Input), result.IsError, errMsg, "", false, false)
	}

	// Audit log
	if a.auditLogger != nil {
		a.auditLogger.Log(audit.AuditEntry{
			Timestamp:     time.Now(),
			SessionID:     a.sessionID,
			ToolName:      toolUse.Name,
			InputSummary:  string(toolUse.Input),
			OutputSummary: result.Content,
			Decision:      "approved",
			Approved:      true,
			DurationMs:    duration.Milliseconds(),
		})
	}

	// Report tool result
	if a.handler != nil {
		a.handler.OnToolResult(toolUse.Name, result)
	}

	// Post-tool hook
	if a.hookRunner != nil {
		a.hookRunner.RunPostToolUse(ctx, toolUse.Name, string(toolUse.Input), result.Content, a.sessionID)
	}

	return result
}

// buildAssistantContent builds the assistant message content
func (a *AgentLoop) buildAssistantContent(resp *client.Response) string {
	var parts []string
	if resp.Content != "" {
		parts = append(parts, resp.Content)
	}

	for _, toolUse := range resp.ToolUses {
		toolJSON, _ := json.Marshal(map[string]any{
			"type":  "tool_use",
			"id":    toolUse.ID,
			"name":  toolUse.Name,
			"input": json.RawMessage(toolUse.Input),
		})
		parts = append(parts, string(toolJSON))
	}

	return strings.Join(parts, "\n")
}

// buildToolResultContent builds the tool result content
func (a *AgentLoop) buildToolResultContent(toolUse client.ToolUse, result ToolResult) string {
	content := result.Content
	if result.IsError {
		content = fmt.Sprintf("[error] %s", content)
	}

	toolResult := map[string]any{
		"type":        "tool_result",
		"tool_use_id": toolUse.ID,
		"content":     content,
	}

	if result.IsError {
		toolResult["is_error"] = true
	}

	resultJSON, _ := json.Marshal(toolResult)
	return string(resultJSON)
}

// chatWithRetry calls the LLM with retry+backoff for transient errors.
// opts contains optional thinking/config fields for the LLM request.
func (a *AgentLoop) chatWithRetry(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, opts *client.ChatOptions) (*client.Response, error) {
	retryCfg := client.DefaultRetryConfig()
	var lastErr error

	for attempt := 0; attempt < retryCfg.MaxRetries; attempt++ {
		if a.enableStreaming {
			if streamer, ok := a.llmClient.(StreamingLLMClient); ok {
				resp, err := streamer.StreamChat(ctx, systemPrompt, messages, tools, a.maxTokens, opts, func(delta string) {
					if a.handler != nil {
						a.handler.OnStreamDelta(delta)
					}
				})
				if err == nil {
					return resp, nil
				}
				if errors.Is(err, client.ErrStreamIdleTimeout) {
					return resp, err
				}
				if !client.IsRetryableError(err) {
					return nil, err
				}
				lastErr = err
				if attempt < retryCfg.MaxRetries-1 {
					if err := a.retryWait(ctx, attempt, retryCfg); err != nil {
						return nil, fmt.Errorf("retry wait cancelled: %w", err)
					}
				}
				continue
			}
		}

		resp, err := a.llmClient.Chat(ctx, systemPrompt, messages, tools, a.maxTokens, opts)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !client.IsRetryableError(err) || attempt >= retryCfg.MaxRetries-1 {
			break
		}

		if err := a.retryWait(ctx, attempt, retryCfg); err != nil {
			return nil, fmt.Errorf("retry wait cancelled: %w", err)
		}
	}

	return nil, fmt.Errorf("LLM call failed after %d attempts: %w", retryCfg.MaxRetries, lastErr)
}

func (a *AgentLoop) retryWait(ctx context.Context, attempt int, cfg client.RetryConfig) error {
	backoff := client.BackoffDelay(attempt, cfg)
	fmt.Fprintf(os.Stderr, "[agent] LLM call failed (attempt %d/%d), retrying in %v\n", attempt+1, cfg.MaxRetries, backoff)
	timer := time.NewTimer(backoff)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// detectContextBloat scans messages for tool-result blocks and returns a
// human-readable detail string when tool-result text exceeds 50% of total
// message content. Returns empty string when bloat is not detected.
func detectContextBloat(messages []client.Message) string {
	var totalSize, toolResultSize int

	for _, msg := range messages {
		totalSize += len(msg.Content)

		// Parse user messages as JSON to find tool_result blocks
		var parsed map[string]any
		if err := json.Unmarshal([]byte(msg.Content), &parsed); err == nil {
			if typ, _ := parsed["type"].(string); typ == "tool_result" {
				if content, _ := parsed["content"].(string); content != "" {
					toolResultSize += len(content)
				}
			}
		}
	}

	if totalSize > 0 && float64(toolResultSize)/float64(totalSize) > 0.5 {
		pct := int(float64(toolResultSize) / float64(totalSize) * 100)
		return fmt.Sprintf("Tool result content dominates context: %d/%d chars (%d%%). Try narrowing tool calls to reduce output.",
			toolResultSize, totalSize, pct)
	}

	return ""
}

// isContextTooLargeError returns true if the error indicates the context
// exceeded the model's maximum input length.
func isContextTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "context_length_exceeded") ||
		strings.Contains(s, "maximum context length") ||
		strings.Contains(s, "too many tokens") ||
		strings.Contains(s, "request too large") ||
		strings.Contains(s, "max_tokens")
}
