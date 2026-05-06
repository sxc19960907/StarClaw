package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/session"
)

// SessionSearchTool searches past session messages.
type SessionSearchTool struct {
	manager *session.Manager
}

// NewSessionSearchTool creates a session search tool bound to a manager.
func NewSessionSearchTool(mgr *session.Manager) *SessionSearchTool {
	return &SessionSearchTool{manager: mgr}
}

type sessionSearchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (t *SessionSearchTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "session_search",
		Description: "Search past session messages for keyword matches. Useful for recalling previous conversations and findings.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search keywords or quoted phrase",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Max results (default 20)",
				},
			},
		},
		Required: []string{"query"},
	}
}

func (t *SessionSearchTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args sessionSearchArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	if args.Query == "" {
		return agent.ValidationError("query is required"), nil
	}
	if args.Limit <= 0 {
		args.Limit = 20
	}

	if t.manager == nil {
		return agent.ToolResult{Content: "session_search: no session manager available", IsError: true}, nil
	}

	summaries, err := t.manager.List()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("session_search: failed to list sessions: %v", err), IsError: true}, nil
	}

	query := strings.ToLower(args.Query)
	var results []string
	count := 0

	for _, s := range summaries {
		if count >= args.Limit {
			break
		}
		// Search in session titles
		if strings.Contains(strings.ToLower(s.Title), query) {
			results = append(results, fmt.Sprintf("[%s] %s (%d messages)", s.ID[:16], s.Title, s.MsgCount))
			count++
		}
	}

	if len(results) == 0 {
		return agent.ToolResult{Content: fmt.Sprintf("No sessions found matching %q", args.Query)}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("Found %d matching sessions:\n\n%s", len(results), strings.Join(results, "\n"))}, nil
}

func (t *SessionSearchTool) RequiresApproval() bool { return false }
