# Logging Guidelines

> Logging conventions for StarClaw CLI

---

## Approach

StarClaw uses **minimal, targeted logging**. This is a CLI tool, not a server — output goes to the user's terminal. Excessive logging degrades UX.

## Stderr vs Stdout

| Stream | Use |
|--------|-----|
| `os.Stderr` | Warnings, errors, status messages, update notifications |
| `os.Stdout` / `fmt.Print` | AI responses, command results, TUI rendering |

```go
// ✓ Stderr for warnings
fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)

// ✓ Stdout for results
fmt.Println(resp.Content)
```

## Module-Level Logging

Use stdlib `log` for background operations (MCP, daemon). Prefix with module tag:

```go
log.Printf("[mcp] %s: not connected, attempting on-demand connect", serverName)
```

## User-Facing Messages

```go
fmt.Printf("📂 Resuming session: %s\n\n", sess.Title)
fmt.Printf("Update available: %s — run 'starclaw update' to install", release.TagName)
```

## What to Log

- ✅ Connection failures / reconnect attempts
- ✅ Non-critical service failures (audit logger unavailable, MCP server error)
- ✅ Update available notifications

## What NOT to Log

- ❌ Full request/response bodies (privacy/security)
- ❌ Debug-level internals in production
- ❌ API keys, auth tokens, or sensitive config values
- ❌ Normal operation success messages (silent on success)

## Audit Logging

For security-sensitive operations, use the audit package:

```go
auditLogger.Log(audit.AuditEntry{
    Timestamp:    time.Now(),
    SessionID:    sessionID,
    ToolName:     toolName,
    Decision:     "approved",
    Approved:     true,
})
```

## Anti-Patterns

- ❌ Don't use `log.Fatal()` in library code — return errors instead
- ❌ Don't print sensitive data in log messages
- ❌ Don't use `fmt.Println` for errors — use `os.Stderr`
