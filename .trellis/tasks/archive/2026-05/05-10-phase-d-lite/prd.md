# Phase D-lite: MCP Server + Process Tool Enhancement

## Goal

Enhance the existing MCP server and process tool to full production quality — adding process spawning with output capture, signal control, and MCP server resource/prompt capabilities.

## Requirements

### 1. Process Tool Enhancement (`internal/tools/process.go`)

Add to existing tool:
- `start` action: spawn a background process, capture stdout/stderr, return PID
- `signal` action: send specific signal (SIGTERM, SIGKILL, SIGINT, SIGHUP) to PID
- `status` action: check if a PID is still running
- Output capture: store last N lines of stdout/stderr per spawned process
- Timeout: auto-kill spawned processes after configurable timeout

### 2. MCP Server Enhancement (`internal/tools/mcp_server.go`)

Add to existing server:
- Resource listing: expose available resources (project files, config)
- Prompt listing: expose available prompt templates
- Tool filtering: allow config to specify which tools to expose via MCP
- Server info: version, capabilities, tool count

### 3. MCP Server Config

- Add `tools.server_tool_timeout` to ToolsConfig (timeout for MCP tool calls)
- Add `tools.mcp_expose` list to ToolsConfig (which tools to expose, empty = all)

## Acceptance Criteria

- [ ] `process start` spawns a background process and returns PID + initial output
- [ ] `process signal` sends specified signal to a PID
- [ ] `process status` reports whether a PID is running
- [ ] Spawned processes auto-terminate after timeout
- [ ] MCP server exposes resources and prompts
- [ ] MCP server respects tool filtering config
- [ ] `tools.server_tool_timeout` is honored
- [ ] Unit tests for new process actions
- [ ] `go build ./...` and `go test ./...` pass

## Technical Notes

- Process tool already has list/kill — extend with start/signal/status
- MCP server already works via mcp-go library — add resource/prompt handlers
- Use `os/exec.Cmd.Start()` for background process spawning
- Store process state in a sync.Map keyed by PID
- Auto-cleanup goroutine for timed-out processes
