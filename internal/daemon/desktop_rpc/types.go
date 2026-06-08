package desktop_rpc

import "encoding/json"

const ProtocolVersion = "1.0.0"

var ProtocolMethods = []string{
	MethodSystemPing,
	MethodSystemCapabilities,
}

const (
	MethodSystemPing         = "system.ping"
	MethodSystemCapabilities = "system.capabilities"
)

const (
	FrameDesktopRPCRequest = "desktop_rpc_request"
	FrameDesktopRPCResult  = "desktop_rpc_result"
	FrameDesktopEvent      = "desktop_event"
)

const (
	ErrCodeInvalidArgument     = "invalid_argument"
	ErrCodeInternal            = "internal_error"
	ErrCodeTimeout             = "timeout"
	ErrCodeDesktopDisconnected = "desktop_disconnected"
)

const MaxFrameBodyBytes = 4 * 1024 * 1024

type Frame struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type RPCRequest struct {
	RequestID string          `json:"request_id"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params,omitempty"`
	TimeoutMs int             `json:"timeout_ms,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Agent     string          `json:"agent,omitempty"`
	Source    string          `json:"source,omitempty"`
	TS        string          `json:"ts,omitempty"`
}

type RPCResult struct {
	RequestID string          `json:"request_id"`
	OK        bool            `json:"ok"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Retriable bool            `json:"retriable"`
	Details   json.RawMessage `json:"details,omitempty"`
}

type DesktopEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
	TS    string          `json:"ts,omitempty"`
}

type SystemCapabilitiesResult struct {
	Version  string   `json:"version"`
	Methods  []string `json:"methods"`
	Platform Platform `json:"platform"`
}

type Platform struct {
	OS         string `json:"os"`
	OSVersion  string `json:"os_version,omitempty"`
	AppVersion string `json:"app_version"`
}

type SystemPingParams struct {
	Echo string `json:"echo,omitempty"`
}

type SystemPingResult struct {
	Pong       string `json:"pong"`
	ServerTime string `json:"server_time"`
}
