# Proposal: Add Audit Logging

## Summary

Add comprehensive audit logging for all tool calls. Each tool invocation is recorded as a JSON line with timestamp, tool name, input/output summaries, and approval status. Sensitive data is automatically redacted.

## Motivation

Security and debugging requirements:
1. **Audit trail** - Know what tools were called, when, and with what arguments
2. **Debugging** - Diagnose issues by reviewing tool call history
3. **Compliance** - Some environments require audit trails
4. **Security** - Detect suspicious tool usage patterns

## Scope

### In Scope
- JSON-lines audit log format
- Automatic secret redaction (AWS keys, JWT, API keys, etc.)
- Integration with Agent Loop
- Configurable enable/disable
- Concurrent-safe logging
- Log file rotation (optional for Phase 1)

### Out of Scope
- Real-time log streaming
- Log analysis/aggregation tools
- Remote audit log forwarding
- Structured query interface (use grep/jq)

## Success Criteria

- [ ] Every tool call is logged to audit file
- [ ] Log entry includes: timestamp, session, tool, input, output, approval, duration
- [ ] Secrets are redacted from logs (AWS keys, JWT, API keys, passwords)
- [ ] Audit logging is thread-safe
- [ ] Can be disabled via configuration
- [ ] Log file has restricted permissions (0600)
- [ ] Unit tests for redaction patterns
- [ ] Integration tests verify logging

## Reference

Based on ShanClaw's implementation: `internal/audit/audit.go`
