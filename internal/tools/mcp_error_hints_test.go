package tools

import (
	"strings"
	"testing"
)

func TestGetHint_KnownErrors(t *testing.T) {
	tests := []struct {
		errMsg   string
		wantHint string
	}{
		{"connection refused: dial tcp 127.0.0.1:8080: connect: connection refused", "The MCP server is not running or not accepting connections."},
		{"i/o timeout after 30s", "The MCP server did not respond in time."},
		{"broken pipe error", "The connection to the MCP server was lost."},
		{"read: broken pipe", "The connection to the MCP server was lost."},
		{"use of closed network connection", "The MCP server connection was closed unexpectedly."},
		{"unexpected EOF", "The MCP server closed the connection unexpectedly."},
		{"process already finished", "The MCP server process has exited."},
		{"signal: killed", "The MCP server process was terminated (killed)."},
		{"permission denied", "The MCP server does not have permission to access the requested resource or command."},
		{"command not found", "The MCP server command was not found."},
		{"failed to start MCP server", "The MCP server process failed to start."},
		{"initialize failed: handshake error", "The MCP server initialization handshake failed."},
		{"tools/call failed: invalid arguments", "Failed to call a tool on the MCP server."},
		{"invalid tool name: foobar", "The requested tool does not exist on the MCP server."},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg[:min(len(tt.errMsg), 40)], func(t *testing.T) {
			hint := GetHint(tt.errMsg)
			if !strings.HasPrefix(hint, tt.wantHint[:len(tt.wantHint)-3]) {
				t.Errorf("GetHint(%q) = %q, want hint starting with %q", tt.errMsg, hint, tt.wantHint[:len(tt.wantHint)-3])
			}
		})
	}
}

// TestGetHint_CompoundErrors tests error messages that match multiple map keys.
// Due to map iteration order, either hint is acceptable.
func TestGetHint_CompoundErrors(t *testing.T) {
	tests := []struct {
		errMsg   string
		possible []string
	}{
		{"tools/list failed: connection refused", []string{
			"Failed to list tools from the MCP server.",
			"The MCP server is not running or not accepting connections.",
		}},
		{"no child processes: connection refused", []string{
			"The MCP server process could not be started.",
			"The MCP server is not running or not accepting connections.",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg[:min(len(tt.errMsg), 40)], func(t *testing.T) {
			hint := GetHint(tt.errMsg)
			matched := false
			for _, p := range tt.possible {
				if strings.HasPrefix(hint, p[:len(p)-3]) {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("GetHint(%q) = %q, expected any of %v", tt.errMsg, hint, tt.possible)
			}
		})
	}
}

func TestGetHint_UnknownError(t *testing.T) {
	hint := GetHint("some random error that doesn't match anything")
	if hint != "An unknown MCP error occurred. Check MCP server logs and configuration for details." {
		t.Errorf("unexpected hint for unknown error: %q", hint)
	}
}

func TestGetHint_EmptyString(t *testing.T) {
	hint := GetHint("")
	if hint == "" {
		t.Error("expected non-empty hint for empty string")
	}
}

func TestGetHint_CaseInsensitive(t *testing.T) {
	// Should match regardless of case
	hint := GetHint("CONNECTION REFUSED")
	if !strings.HasPrefix(hint, "The MCP server is not running") {
		t.Errorf("case insensitive match failed, got: %q", hint)
	}

	hint = GetHint("Connection Refused")
	if !strings.HasPrefix(hint, "The MCP server is not running") {
		t.Errorf("case insensitive match failed, got: %q", hint)
	}
}

func TestMCPErrorHints_HasEntries(t *testing.T) {
	if len(MCPErrorHints) < 10 {
		t.Errorf("expected at least 10 error hint entries, got %d", len(MCPErrorHints))
	}
}

func TestMCPErrorHints_NonEmpty(t *testing.T) {
	for substr, hint := range MCPErrorHints {
		if substr == "" {
			t.Error("empty substr in MCPErrorHints")
		}
		if hint == "" {
			t.Errorf("empty hint for substr %q", substr)
		}
	}
}
