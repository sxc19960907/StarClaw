# Tasks: Add Audit Logging

## Task 1: Create AuditLogger Core

**ID**: T1  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Create `AuditLogger` struct with thread-safe JSON-lines logging.

### Acceptance Criteria
- [ ] `internal/audit/audit.go` created
- [ ] `AuditEntry` struct with all required fields
- [ ] `AuditLogger` struct with mutex protection
- [ ] `NewAuditLogger(logDir)` creates directory and log file
- [ ] `Log(entry)` writes JSON line with proper formatting
- [ ] `Close()` properly closes file
- [ ] Log file created with 0600 permissions
- [ ] Directory created with 0700 permissions

### Notes
Reference: ShanClaw `internal/audit/audit.go`

---

## Task 2: Implement Secret Redaction

**ID**: T2  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Add `RedactSecrets()` function with regex patterns for common secrets.

### Acceptance Criteria
- [ ] Redaction patterns defined for:
  - AWS keys (AKIA...)
  - JWT tokens
  - sk- style API keys
  - key- style API keys
  - Bearer tokens
  - PEM markers
  - Secret env vars (KEY=, SECRET=, TOKEN=, PASSWORD=)
  - GitHub tokens (ghp_, gho_, etc.)
- [ ] Patterns compiled at init
- [ ] `RedactSecrets(text)` replaces matches with `[REDACTED]`
- [ ] `truncate()` helper for limiting summary length

---

## Task 3: Write Audit Package Unit Tests

**ID**: T3  
**Status**: completed  
**Owner**:  
**Blocked By**: T1, T2

### Description
Create comprehensive unit tests for audit logging.

### Acceptance Criteria
- [ ] `audit_test.go` created
- [ ] Test secret redaction for each pattern type
- [ ] Test `truncate()` function
- [ ] Test `NewAuditLogger` creates files with correct permissions
- [ ] Test `Log()` writes valid JSON
- [ ] Test concurrent logging (thread safety)
- [ ] Test `Close()` behavior
- [ ] All tests pass

### Test Cases
```go
// AWS key
"AKIAIOSFODNN7EXAMPLE" → "[REDACTED]"

// JWT
eyJhbG... → "[REDACTED]"

// API key
"sk-abc123def456" → "[REDACTED]"

// Normal text (unchanged)
"hello world" → "hello world"
```

---

## Task 4: Add Audit Config

**ID**: T4  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Add audit configuration to config package.

### Acceptance Criteria
- [ ] `AuditConfig` struct with `Enabled bool`
- [ ] Default value: `enabled: true`
- [ ] Config properly unmarshaled
- [ ] Test config loading

### Notes
Add to existing config structure, no breaking changes.

---

## Task 5: Integrate Audit Logger into Agent Loop

**ID**: T5  
**Status**: completed  
**Owner**:  
**Blocked By**: T1, T4

### Description
Modify `AgentLoop` to log all tool calls.

### Acceptance Criteria
- [ ] `auditLogger` field added to `AgentLoop`
- [ ] `SetAuditLogger()` method added
- [ ] `executeTool()` logs before/after tool execution
- [ ] Logs include: timestamp, session, tool, input, output, approval, duration
- [ ] Logging is non-blocking (errors don't break tool execution)
- [ ] If audit logger is nil, no-op (not an error)

### Integration Code
```go
start := time.Now()
result, err := tool.Run(ctx, string(toolUse.Input))
duration := time.Since(start)

if a.auditLogger != nil {
    a.auditLogger.Log(audit.AuditEntry{
        Timestamp:     time.Now(),
        SessionID:     a.sessionID,
        ToolName:      toolUse.Name,
        InputSummary:  string(toolUse.Input),
        OutputSummary: result.Content,
        Approved:      true,
        DurationMs:    duration.Milliseconds(),
    })
}
```

---

## Task 6: Add Session ID to Agent Loop

**ID**: T6  
**Status**: completed  
**Owner**:  
**Blocked By**: T5

### Description
Add session tracking to Agent Loop for audit correlation.

### Acceptance Criteria
- [ ] `sessionID` field added to `AgentLoop`
- [ ] `SetSessionID(id string)` method added
- [ ] If not set, use empty string or "default"

---

## Task 7: Initialize Audit Logger in CLI

**ID**: T7  
**Status**: completed  
**Owner**:  
**Blocked By**: T4, T5

### Description
Wire up audit logger in main/CLI initialization.

### Acceptance Criteria
- [ ] Create audit logger based on config
- [ ] Pass to Agent Loop via `SetAuditLogger()`
- [ ] Proper cleanup on shutdown (Close())
- [ ] Works for both TUI and one-shot modes

---

## Task 8: Write Integration Tests

**ID**: T8  
**Status**: completed  
**Owner**:  
**Blocked By**: T7

### Description
Add integration tests verifying end-to-end audit logging.

### Acceptance Criteria
- [ ] Test that tool calls create audit entries
- [ ] Test that audit file contains valid JSON lines
- [ ] Test audit entries have expected fields
- [ ] Test with multiple tool calls in sequence

---

## Task 9: Add Documentation

**ID**: T9  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Document audit logging feature.

### Acceptance Criteria
- [ ] Update README with audit log location
- [ ] Add example of querying logs with jq/grep
- [ ] Document config option to disable

---

## Dependencies

```
T1 (AuditLogger Core)
  └── T2 (Redaction)
  └── T3 (Unit Tests)

T4 (Config)
  └── T7 (CLI Integration)

T1 + T5 (Agent Integration)
  └── T6 (Session ID)

T3 + T6 + T7 ──▶ T8 (Integration Tests)

T9 (Docs) - independent
```

## Estimated Effort

- T1: 1 hour
- T2: 45 minutes
- T3: 1.5 hours
- T4: 30 minutes
- T5: 45 minutes
- T6: 30 minutes
- T7: 30 minutes
- T8: 1 hour
- T9: 30 minutes

**Total**: ~6.5 hours
