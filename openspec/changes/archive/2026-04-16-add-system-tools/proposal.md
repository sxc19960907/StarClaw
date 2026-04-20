# Proposal: Add System Tools

## Summary

Add two new tools to enhance StarClaw's system interaction capabilities:
- `system_info`: Get OS, architecture, hostname, CPU, memory, and disk information
- `http`: Make HTTP requests with configurable method, headers, body, and timeout

## Motivation

Currently, StarClaw has limited visibility into the system environment. Users often need to:
1. Check system information before running commands
2. Make HTTP requests to APIs or local services

The bash tool can do these, but dedicated tools provide:
- Structured output (easier for AI to parse)
- Cross-platform compatibility (no shell dependency)
- Safety controls (http has SafeChecker for localhost GET)
- Better UX (no need to parse command output)

## Scope

### In Scope
- `system_info` tool with cross-platform support
- `http` tool with GET/POST/PUT/DELETE methods
- SafeChecker for http (auto-approve localhost GET)
- Unit tests with mocking
- Integration tests

### Out of Scope
- Authentication handling for HTTP (Bearer tokens in headers only)
- HTTP/2 specific features
- WebSocket support
- System metrics collection (CPU usage over time)

## Success Criteria

- [ ] `system_info` returns OS, arch, hostname, CPUs on all platforms
- [ ] `system_info` returns memory info on Linux/macOS
- [ ] `system_info` returns disk info where supported
- [ ] `http` tool supports GET, POST, PUT, DELETE
- [ ] `http` tool supports custom headers and request body
- [ ] `http` tool has configurable timeout
- [ ] `http` GET to localhost is auto-approved (SafeChecker)
- [ ] Both tools have comprehensive unit tests
- [ ] Integration tests verify tool registration

## Reference

Based on ShanClaw's implementation:
- `internal/tools/system_info.go`
- `internal/tools/system_info_darwin.go`
- `internal/tools/system_info_other.go`
- `internal/tools/http.go`
