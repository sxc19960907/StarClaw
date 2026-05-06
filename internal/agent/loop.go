package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	ctxwin "github.com/starclaw/starclaw/internal/context"
	"github.com/starclaw/starclaw/internal/permissions"
	"github.com/starclaw/starclaw/internal/session"
)

// EventHandler handles events from the agent loop
type EventHandler interface {
	OnToolCall(name string, args string)
	OnToolResult(name string, result ToolResult)
	OnText(text string)
	OnUsage(usage client.Usage)
}

// LLMClient defines the interface for LLM clients
type LLMClient interface {
	Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int) (*client.Response, error)
}

// AgentLoop manages the conversation with the LLM
type AgentLoop struct {
	llmClient     LLMClient
	registry      *ToolRegistry
	maxIter       int
	maxTokens     int
	resultTrunc   int
	handler       EventHandler
	systemPrompt  string
	auditLogger   *audit.AuditLogger
	sessionID     string
	session       *session.Session
	sessionMgr    *session.Manager
	memory        string // agent memory content
	memoryDir     string // directory for persistent memory
	configDir     string // starclaw config dir (~/.starclaw)
	loopDetector  *LoopDetector
	contextWindow int                  // max context window in tokens (0 = disabled)
	permsConfig   *permissions.Config  // tool permission rules
}

// NewAgentLoop creates a new agent loop
func NewAgentLoop(llmClient LLMClient, registry *ToolRegistry) *AgentLoop {
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

// Run executes the agent loop with the given query
func (a *AgentLoop) Run(ctx context.Context, query string) (*client.Response, error) {
	// Initialize messages from session if resuming, or start fresh
	messages := []client.Message{}
	if a.session != nil {
		messages = append(messages, a.session.Messages...)
	}
	messages = append(messages, client.Message{Role: "user", Content: query})

	// Update session title if this is the first message
	if a.session != nil && len(a.session.Messages) == 0 {
		a.session.Title = session.GenerateTitle(query)
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

		// Call LLM with retry
		resp, err := a.chatWithRetry(ctx, effectivePrompt, messages, tools)
		if err != nil {
			return nil, fmt.Errorf("LLM error: %w", err)
		}

		// Report usage
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
				a.session.Messages = messages
				a.session.UpdatedAt = time.Now()
				if a.sessionMgr != nil {
					a.sessionMgr.Save()
				}
			}

			return resp, nil
		}

		// Handle tool calls
		messages = append(messages, client.Message{
			Role:    "assistant",
			Content: a.buildAssistantContent(resp),
		})

		for _, toolUse := range resp.ToolUses {
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
					// Fall through to return — handled by maxIter or next text-only response
				}
			}
		}

		// Update session after each turn and auto-save
		if a.session != nil {
			a.session.Messages = messages
			a.session.UpdatedAt = time.Now()
			if a.sessionMgr != nil {
				a.sessionMgr.Save()
			}
		}
	}

	return nil, fmt.Errorf("reached maximum iterations (%d)", a.maxIter)
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
	if a.permsConfig != nil {
		decision, reason := permissions.CheckToolCall(toolUse.Name, string(toolUse.Input), a.permsConfig)
		if decision == permissions.Deny {
			return PermissionError(fmt.Sprintf("%s: blocked (%s)", toolUse.Name, reason))
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
		a.loopDetector.Record(toolUse.Name, string(toolUse.Input), result.IsError, errMsg, "")
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
		"type":       "tool_result",
		"tool_use_id": toolUse.ID,
		"content":    content,
	}

	if result.IsError {
		toolResult["is_error"] = true
	}

	resultJSON, _ := json.Marshal(toolResult)
	return string(resultJSON)
}

// chatWithRetry calls the LLM with retry+backoff for transient errors.
func (a *AgentLoop) chatWithRetry(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef) (*client.Response, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := a.llmClient.Chat(ctx, systemPrompt, messages, tools, a.maxTokens)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableLLMError(err) || attempt >= maxRetries-1 {
			break
		}

		backoff := time.Duration(1<<attempt) * time.Second // 1s, 2s, 4s
		fmt.Fprintf(os.Stderr, "[agent] LLM call failed (attempt %d/%d), retrying in %v: %v\n", attempt+1, maxRetries, backoff, err)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, fmt.Errorf("LLM call cancelled: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries, lastErr)
}

// isRetryableLLMError returns true for transient errors that may succeed on retry.
func isRetryableLLMError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// Rate limits and server errors
	if strings.Contains(s, "429") || strings.Contains(s, "rate limit") {
		return true
	}
	if strings.Contains(s, "500") || strings.Contains(s, "502") || strings.Contains(s, "503") || strings.Contains(s, "504") {
		return true
	}
	// Timeout and connection errors
	if strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded") {
		return true
	}
	if strings.Contains(s, "connection reset") || strings.Contains(s, "connection refused") {
		return true
	}
	if strings.Contains(s, "EOF") || strings.Contains(s, "broken pipe") {
		return true
	}
	// Context cancellation is NOT retryable
	if strings.Contains(s, "context canceled") || strings.Contains(s, "context deadline exceeded") {
		return false
	}
	// Auth errors are NOT retryable
	if strings.Contains(s, "401") || strings.Contains(s, "403") {
		return false
	}
	return false
}
