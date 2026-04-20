# Specification: Audit Logging

## Overview

| Field | Value |
|-------|-------|
| Feature | Audit Logging |
| Type | Infrastructure |
| Optional | Yes (configurable) |
| Default | Enabled |

## Purpose

Record all tool invocations for security auditing, debugging, and compliance. Each tool call is logged as a structured JSON entry with automatic redaction of sensitive data.

## Log Location

```
~/.starclaw/logs/audit.log
```

## Log Format

### Entry Structure

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | ISO8601 | When the call was made |
| `session_id` | string | Session identifier for correlation |
| `tool_name` | string | Name of the invoked tool |
| `input_summary` | string | Tool arguments (redacted, truncated) |
| `output_summary` | string | Tool result (redacted, truncated) |
| `decision` | string | "approved", "denied", "auto-approved" |
| `approved` | boolean | Whether the call was approved |
| `duration_ms` | integer | Execution time in milliseconds |

### JSON Lines Format

Each line is a complete JSON object:

```json
{"timestamp":"2026-04-16T10:30:00Z","session_id":"sess-abc123","tool_name":"file_read","input_summary":"{\"file_path\":\"/tmp/test.txt\"}","output_summary":"Hello world","decision":"approved","approved":true,"duration_ms":5}
```

### Example Entries

#### File Read
```json
{
  "timestamp": "2026-04-16T10:30:00Z",
  "session_id": "sess-abc123",
  "tool_name": "file_read",
  "input_summary": "{\"file_path\":\"/tmp/test.txt\"}",
  "output_summary": "Hello world content...",
  "decision": "approved",
  "approved": true,
  "duration_ms": 5
}
```

#### Bash with Auto-Approval
```json
{
  "timestamp": "2026-04-16T10:30:05Z",
  "session_id": "sess-abc123",
  "tool_name": "bash",
  "input_summary": "{\"command\":\"ls -la\"}",
  "output_summary": "total 16\ndrwxr-xr-x 3 user...",
  "decision": "auto-approved",
  "approved": true,
  "duration_ms": 50
}
```

#### Denied Tool
```json
{
  "timestamp": "2026-04-16T10:30:10Z",
  "session_id": "sess-abc123",
  "tool_name": "bash",
  "input_summary": "{\"command\":\"rm -rf /\"}",
  "output_summary": "",
  "decision": "denied",
  "approved": false,
  "duration_ms": 0
}
```

## Secret Redaction

### Redaction Patterns

Secrets are automatically replaced with `[REDACTED]`:

| Pattern | Example |
|---------|---------|
| AWS Access Key | `AKIAIOSFODNN7EXAMPLE` |
| JWT Token | `eyJhbGciOiJIUzI1NiIs...` |
| OpenAI/Stripe Key | `sk-abc123def456` |
| Generic API Key | `key-secret123` |
| Bearer Token | `Bearer abc123...` |
| PEM Block | `-----BEGIN PRIVATE KEY-----` |
| Environment Variable | `API_KEY=secret123` |
| GitHub Token | `ghp_xxx`, `gho_xxx` |

### Before Redaction
```
Input: {"api_key": "sk-abc123secret", "authorization": "Bearer eyJhbG..."}
```

### After Redaction
```
Input: {"api_key": "[REDACTED]", "authorization": "[REDACTED]"}
```

## Configuration

```yaml
# ~/.starclaw/config.yaml
audit:
  enabled: true  # Set to false to disable
```

Default: `enabled: true`

## Security

### File Permissions

- Log file: `0600` (owner read/write only)
- Log directory: `0700` (owner access only)

### Privacy

- Summaries truncated to 500 characters
- Secrets automatically redacted
- Local file only (no network transmission)
- No user PII collected

## Querying Logs

### View Recent Entries

```bash
tail -f ~/.starclaw/logs/audit.log
```

### Pretty Print

```bash
jq . ~/.starclaw/logs/audit.log
```

### Filter by Tool

```bash
grep '"tool_name":"bash"' ~/.starclaw/logs/audit.log | jq .
```

### Filter by Session

```bash
grep '"session_id":"sess-abc123"' ~/.starclaw/logs/audit.log | jq .
```

### Time Range

```bash
# Using jq
jq 'select(.timestamp >= "2026-04-16T10:00:00Z")' ~/.starclaw/logs/audit.log
```

### Statistics

```bash
# Count by tool
jq -r '.tool_name' ~/.starclaw/logs/audit.log | sort | uniq -c | sort -rn

# Count by decision
jq -r '.decision' ~/.starclaw/logs/audit.log | sort | uniq -c
```

## Design Rationale

- **JSON Lines**: Append-only, easy to parse, resilient to corruption
- **Automatic redaction**: Protects secrets without user action
- **Truncation**: Prevents log abuse from large outputs
- **Thread-safe**: Concurrent tool calls are logged correctly
- **Optional**: Can be disabled if not needed
- **Local only**: No external dependencies or privacy concerns
