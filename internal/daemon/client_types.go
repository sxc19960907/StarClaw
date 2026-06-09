package daemon

// StatusResponse contains daemon server status information.
type StatusResponse struct {
	RunningAgents int              `json:"running_agents"`
	ActiveAgents  int              `json:"active_agents,omitempty"`
	Uptime        string           `json:"uptime"`
	Version       string           `json:"version,omitempty"`
	DesktopRPC    DesktopRPCStatus `json:"desktop_rpc,omitempty"`
}

type DesktopRPCStatus struct {
	Listening bool                  `json:"listening"`
	Connected bool                  `json:"connected"`
	Pending   int                   `json:"pending"`
	Events    DesktopRPCEventStatus `json:"events,omitempty"`
}

type DesktopRPCEventStatus struct {
	Retained  int    `json:"retained"`
	LastEvent string `json:"last_event,omitempty"`
	LastTS    string `json:"last_ts,omitempty"`
}

// CancelRequest cancels a running agent execution.
type CancelRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id,omitempty"`
}

// CreateScheduleRequest is the request to create a new schedule.
type CreateScheduleRequest struct {
	Agent  string `json:"agent"`
	Cron   string `json:"cron"`
	Prompt string `json:"prompt"`
}

// PatchScheduleRequest carries optional fields to update on a schedule.
// Pointer fields distinguish "not set" from zero values.
type PatchScheduleRequest struct {
	Cron    *string `json:"cron,omitempty"`
	Prompt  *string `json:"prompt,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}
