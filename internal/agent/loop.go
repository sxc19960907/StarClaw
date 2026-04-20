package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
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
	llmClient    LLMClient
	registry     *ToolRegistry
	maxIter      int
	maxTokens    int
	resultTrunc  int
	handler      EventHandler
	systemPrompt string
	auditLogger  *audit.AuditLogger
	sessionID    string
	session      *session.Session
	sessionMgr   *session.Manager
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

		// Call LLM
		resp, err := a.llmClient.Chat(ctx, a.systemPrompt, messages, tools, a.maxTokens)
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

	// Truncate result if needed
	if len(result.Content) > a.resultTrunc {
		keepHead := a.resultTrunc * 3 / 4
		keepTail := a.resultTrunc / 4
		result.Content = result.Content[:keepHead] +
			fmt.Sprintf("\n\n[... truncated %d chars ...]\n\n", len(result.Content)-a.resultTrunc) +
			result.Content[len(result.Content)-keepTail:]
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
