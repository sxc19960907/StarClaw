package daemon

import (
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/schedule"
)

// Channel routing constants.
const (
	ChannelCLI      = "cli"
	ChannelHTTP     = "http"
	ChannelSchedule = "schedule"
)

// RunAgentRequest is a request to execute an agent.
type RunAgentRequest struct {
	Text       string   `json:"text"`
	Agent      string   `json:"agent,omitempty"`
	Source     string   `json:"source,omitempty"`
	Channel    string   `json:"channel,omitempty"`
	Sender     string   `json:"sender,omitempty"`
	NewSession bool     `json:"new_session,omitempty"`
	SessionID  string   `json:"session_id,omitempty"`
	Model      string   `json:"model,omitempty"`
	RequestID  string   `json:"request_id,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

// RunAgentResponse is the result of an agent execution.
type RunAgentResponse struct {
	SessionID string         `json:"session_id"`
	Messages  []string       `json:"messages,omitempty"`
	Usage     map[string]int `json:"usage,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// ServerDeps aggregates all dependencies the daemon server needs.
type ServerDeps struct {
	StarclawDir      string
	ConfigPath       string
	AgentsDir        string
	SkillsDir        string
	InstructionsDir  string
	LLMClient        client.LLMClient
	Registry         *agent.ToolRegistry
	ScheduleManager  *schedule.Manager
}

// ApprovalDecision represents the user's response to a tool approval request.
type ApprovalDecision string

const (
	DecisionAllow ApprovalDecision = "allow"
	DecisionDeny  ApprovalDecision = "deny"
)

// ApprovalRequest is sent by the daemon when a tool needs user approval.
type ApprovalRequest struct {
	Channel   string `json:"channel"`
	ThreadID  string `json:"thread_id"`
	RequestID string `json:"request_id"`
	Tool      string `json:"tool"`
	Args      string `json:"args"`
	Agent     string `json:"agent"`
}

// ApprovalResolvedPayload carries the resolution of an approval request.
type ApprovalResolvedPayload struct {
	RequestID  string           `json:"request_id"`
	Decision   ApprovalDecision `json:"decision"`
	ResolvedBy string           `json:"resolved_by,omitempty"`
}
