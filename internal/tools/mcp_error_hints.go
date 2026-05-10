package tools

import "strings"

// MCPErrorHints maps common MCP error message substrings to user-friendly hints.
var MCPErrorHints = map[string]string{
	"connection refused":     "The MCP server is not running or not accepting connections. Start the server and try again.",
	"no such host":           "The MCP server hostname could not be resolved. Check the server URL or host configuration.",
	"i/o timeout":            "The MCP server did not respond in time. Check network connectivity and server load.",
	"broken pipe":            "The connection to the MCP server was lost. The server may have been restarted or crashed.",
	"closed network connection": "The MCP server connection was closed unexpectedly. Reconnect and try again.",
	"eof": "The MCP server closed the connection unexpectedly. The server may have crashed or been shut down.",
	"process already finished":  "The MCP server process has exited. Restart the server to use it again.",
	"signal: killed":            "The MCP server process was terminated (killed). Check system resources and restart.",
	"no child processes":        "The MCP server process could not be started. Check the server command and arguments.",
	"permission denied":         "The MCP server does not have permission to access the requested resource or command.",
	"not found":                 "The MCP server command was not found. Verify the server binary is installed and in PATH.",
	"failed to start":           "The MCP server process failed to start. Check command, arguments, and environment variables.",
	"initialize failed":         "The MCP server initialization handshake failed. Check that the server supports the MCP protocol.",
	"tools/list failed":         "Failed to list tools from the MCP server. The server may be in an unexpected state.",
	"tools/call failed":         "Failed to call a tool on the MCP server. Check tool arguments and server state.",
	"invalid tool name":         "The requested tool does not exist on the MCP server. Check the tool name and try again.",
}

// GetHint returns a user-friendly hint for a given error message.
// It iterates over MCPErrorHints and returns the first matching hint.
// If no hint matches, it returns a generic fallback message.
func GetHint(errMsg string) string {
	lower := strings.ToLower(errMsg)
	for substr, hint := range MCPErrorHints {
		if strings.Contains(lower, substr) {
			return hint
		}
	}
	return "An unknown MCP error occurred. Check MCP server logs and configuration for details."
}
