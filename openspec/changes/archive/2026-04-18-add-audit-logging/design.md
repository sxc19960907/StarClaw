# Design: Audit Logging

## Overview

Audit logging captures all tool calls to an append-only JSON-lines file. Each entry includes metadata about the call, with automatic redaction of sensitive data.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Audit Logging                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │  Agent      │────▶│  executeTool│────▶│ AuditLogger │   │
│  │  Loop       │     │             │     │             │   │
│  └─────────────┘     └─────────────┘     └──────┬──────┘   │
│                                                 │           │
│                                                 ▼           │
│                                          ┌─────────────┐    │
│                                          │ RedactSecrets│   │
│                                          │             │    │
│                                          │ - AWS keys  │    │
│                                          │ - JWT       │    │
│                                          │ - API keys  │    │
│                                          │ - Passwords │    │
│                                          └──────┬──────┘    │
│                                                 │           │
│                                                 ▼           │
│                                          ┌─────────────┐    │
│                                          │  audit.log  │    │
│                                          │ (JSON-lines)│    │
│                                          └─────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Implementation

### Audit Entry Structure

```go
// internal/audit/audit.go
package audit

import (
    "encoding/json"
    "os"
    "path/filepath"
    "regexp"
    "sync"
    "time"
)

// AuditEntry represents a single audited tool call event.
type AuditEntry struct {
    Timestamp     time.Time `json:"timestamp"`
    SessionID     string    `json:"session_id"`
    ToolName      string    `json:"tool_name"`
    InputSummary  string    `json:"input_summary"`
    OutputSummary string    `json:"output_summary"`
    Decision      string    `json:"decision"`     // "approved", "denied", "auto-approved"
    Approved      bool      `json:"approved"`
    DurationMs    int64     `json:"duration_ms"`
}

// AuditLogger writes audit entries as JSON lines.
type AuditLogger struct {
    mu     sync.Mutex
    file   *os.File
    closed bool
}
```

### Logger Creation

```go
const maxSummaryLen = 500

// NewAuditLogger creates a logger that writes to logDir/audit.log.
// Creates the directory if it does not exist.
func NewAuditLogger(logDir string) (*AuditLogger, error) {
    if logDir == "" {
        return nil, fmt.Errorf("logDir must not be empty")
    }

    if err := os.MkdirAll(logDir, 0700); err != nil {
        return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
    }

    logPath := filepath.Join(logDir, "audit.log")
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        return nil, fmt.Errorf("failed to open audit log %s: %w", logPath, err)
    }

    return &AuditLogger{file: f}, nil
}
```

### Log Method

```go
// Log records a tool call event with automatic redaction.
func (a *AuditLogger) Log(entry AuditEntry) {
    // Truncate and redact
    entry.InputSummary = RedactSecrets(truncate(entry.InputSummary, maxSummaryLen))
    entry.OutputSummary = RedactSecrets(truncate(entry.OutputSummary, maxSummaryLen))

    data, err := json.Marshal(entry)
    if err != nil {
        // Silently drop on marshal error - don't break the flow
        return
    }

    a.mu.Lock()
    defer a.mu.Unlock()

    if a.closed {
        return
    }

    a.file.Write(data)
    a.file.Write([]byte("\n"))
}

// Close closes the underlying log file.
func (a *AuditLogger) Close() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.closed {
        return nil
    }
    a.closed = true
    return a.file.Close()
}
```

### Secret Redaction

```go
// Redaction patterns compiled once at package init
var redactPatterns []*regexp.Regexp

func init() {
    patterns := []string{
        // AWS access key IDs
        `AKIA[0-9A-Z]{16}`,
        // JWT tokens (header.payload.signature format)
        `eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,
        // sk- style API keys (OpenAI, Stripe, etc.)
        `sk-[a-zA-Z0-9]{20,}`,
        // key- style API keys
        `key-[a-zA-Z0-9]{20,}`,
        // Bearer tokens
        `Bearer\s+[A-Za-z0-9_\-\.]+`,
        // PEM content markers
        `-----BEGIN[A-Z\s]*-----`,
        `-----END[A-Z\s]*-----`,
        // Env var assignments with secret-like names
        `(?i)[A-Z_]*(?:KEY|SECRET|TOKEN|PASSWORD)\s*=\s*\S+`,
        // GitHub tokens (ghp_, gho_, ghs_, etc.)
        `gh[pousr]_[A-Za-z0-9]{36,}`,
        // Generic API key patterns
        `(?i)api[_-]?key["\s]*[:=]["\s]*[A-Za-z0-9]{16,}`,
    }

    for _, p := range patterns {
        if re, err := regexp.Compile(p); err == nil {
            redactPatterns = append(redactPatterns, re)
        }
    }
}

// RedactSecrets replaces known secret patterns with [REDACTED].
func RedactSecrets(text string) string {
    result := text
    for _, re := range redactPatterns {
        result = re.ReplaceAllString(result, "[REDACTED]")
    }
    return result
}

// truncate shortens text to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

## Integration with Agent Loop

### Modified Agent Loop

```go
// internal/agent/loop.go

type AgentLoop struct {
    // ... existing fields ...
    auditLogger *audit.AuditLogger
}

// SetAuditLogger sets the audit logger (optional)
func (a *AgentLoop) SetAuditLogger(logger *audit.AuditLogger) {
    a.auditLogger = logger
}

// executeTool executes a tool and logs the call
func (a *AgentLoop) executeTool(ctx context.Context, toolUse client.ToolUse) ToolResult {
    tool, ok := a.registry.Get(toolUse.Name)
    if !ok {
        return ValidationError(fmt.Sprintf("unknown tool: %s", toolUse.Name))
    }

    // Report tool call
    if a.handler != nil {
        a.handler.OnToolCall(toolUse.Name, string(toolUse.Input))
    }

    // Execute with timing
    start := time.Now()
    result, err := tool.Run(ctx, string(toolUse.Input))
    duration := time.Since(start)

    if err != nil {
        result = ToolResult{
            Content: fmt.Sprintf("error: %v", err),
            IsError: true,
        }
    }

    // Audit log
    if a.auditLogger != nil {
        a.auditLogger.Log(audit.AuditEntry{
            Timestamp:     time.Now(),
            SessionID:     a.sessionID,  // Would need to add this field
            ToolName:      toolUse.Name,
            InputSummary:  string(toolUse.Input),
            OutputSummary: result.Content,
            Decision:      "approved",  // Or determine from approval flow
            Approved:      true,
            DurationMs:    duration.Milliseconds(),
        })
    }

    // ... rest of method ...
}
```

### Configuration

```go
// internal/config/config.go

type Config struct {
    // ... existing fields ...
    Audit AuditConfig `mapstructure:"audit" yaml:"audit"`
}

type AuditConfig struct {
    Enabled bool `mapstructure:"enabled" yaml:"enabled"`
}

// Default: enabled = true
viper.SetDefault("audit.enabled", true)
```

### Initialization

```go
// cmd/root.go or main.go

func setupAuditLogger(cfg *config.Config) (*audit.AuditLogger, error) {
    if !cfg.Audit.Enabled {
        return nil, nil
    }
    
    logDir := filepath.Join(config.StarclawDir(), "logs")
    return audit.NewAuditLogger(logDir)
}
```

## Testing Strategy

### Unit Tests

```go
// internal/audit/audit_test.go

func TestRedactSecrets_AWSKey(t *testing.T) {
    input := "Access key: AKIAIOSFODNN7EXAMPLE"
    expected := "Access key: [REDACTED]"
    assert.Equal(t, expected, RedactSecrets(input))
}

func TestRedactSecrets_JWT(t *testing.T) {
    input := "Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
    assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_BearerToken(t *testing.T) {
    input := "Authorization: Bearer abc123def456"
    assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_EnvVar(t *testing.T) {
    input := "export API_KEY=secret123"
    assert.Contains(t, RedactSecrets(input), "[REDACTED]")
}

func TestRedactSecrets_NoSecrets(t *testing.T) {
    input := "This is normal text without secrets"
    assert.Equal(t, input, RedactSecrets(input))
}

func TestAuditLogger_Log(t *testing.T) {
    tmpDir := t.TempDir()
    logger, err := NewAuditLogger(tmpDir)
    require.NoError(t, err)
    defer logger.Close()

    entry := AuditEntry{
        Timestamp:     time.Now(),
        SessionID:     "test-session",
        ToolName:      "file_read",
        InputSummary:  `{"file_path": "/tmp/test"}`,
        OutputSummary: "file contents",
        Approved:      true,
        DurationMs:    100,
    }

    logger.Log(entry)

    // Read and verify log file
    content, err := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
    require.NoError(t, err)
    
    var loggedEntry AuditEntry
    err = json.Unmarshal(content, &loggedEntry)
    require.NoError(t, err)
    
    assert.Equal(t, "file_read", loggedEntry.ToolName)
    assert.Equal(t, "test-session", loggedEntry.SessionID)
}

func TestAuditLogger_ThreadSafe(t *testing.T) {
    tmpDir := t.TempDir()
    logger, err := NewAuditLogger(tmpDir)
    require.NoError(t, err)
    defer logger.Close()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            logger.Log(AuditEntry{
                SessionID: fmt.Sprintf("session-%d", n),
                ToolName:  "test",
            })
        }(i)
    }
    wg.Wait()

    // Verify all entries were written
    content, _ := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
    lines := strings.Split(strings.TrimSpace(string(content)), "\n")
    assert.Len(t, lines, 100)
}
```

## Log File Format

### Example Entries

```json
{"timestamp":"2026-04-16T10:30:00Z","session_id":"sess-abc123","tool_name":"file_read","input_summary":"{\"file_path\":\"/tmp/test.txt\"}","output_summary":"Hello world content","decision":"approved","approved":true,"duration_ms":5}
{"timestamp":"2026-04-16T10:30:05Z","session_id":"sess-abc123","tool_name":"bash","input_summary":"{\"command\":\"ls -la\"}","output_summary":"total 16\ndrwxr-xr-x  3 user user 4096 Apr 16 10:30 .","decision":"auto-approved","approved":true,"duration_ms":50}
```

### Querying Logs

```bash
# View recent entries
tail -f ~/.starclaw/logs/audit.log

# Pretty print
jq . ~/.starclaw/logs/audit.log

# Filter by tool
grep '"tool_name":"bash"' ~/.starclaw/logs/audit.log | jq .

# Filter by session
grep '"session_id":"sess-abc123"' ~/.starclaw/logs/audit.log | jq .
```

## Security Considerations

1. **File permissions**: Log file created with 0600 (owner read/write only)
2. **Directory permissions**: Log directory created with 0700
3. **Secret redaction**: Multiple patterns to catch common secret formats
4. **Truncation**: Limits log entry size to prevent abuse
5. **No external exposure**: Local file only, no network transmission

## Future Enhancements

- Log rotation by size or date
- Structured query interface
- Log compression for old entries
- Integration with external SIEM systems
