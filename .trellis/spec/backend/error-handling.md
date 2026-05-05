# Error Handling

> Error handling conventions for StarClaw Go code

---

## Core Rules

### 1. Always wrap errors with context

Wrap errors from external calls so failures are traceable:

```go
// ✓ Good — wraps with operation context
data, err := os.ReadFile(path)
if err != nil {
    return nil, fmt.Errorf("failed to read config file: %w", err)
}

// ✗ Bad — loses context
data, err := os.ReadFile(path)
if err != nil {
    return nil, err
}
```

### 2. Use `%w` for error wrapping, `%v` for logging

- `%w` preserves the error chain (allows `errors.Is` / `errors.As`)
- `%v` converts to string (breaks the chain)

```go
// ✓ Good — %w preserves original error
return fmt.Errorf("agent %q: %w", name, err)

// ✗ Bad — %v breaks error unwrapping
return fmt.Errorf("agent %q: %v", name, err)
```

### 3. Format: `"action description: %w"`

Error messages follow `"<what we were doing>: <specific detail>"`:

```go
fmt.Errorf("failed to parse config: %w", err)
fmt.Errorf("agent %q not found: AGENT.md missing", name)
fmt.Errorf("MCP server %q on-demand connect failed: %w", serverName, err)
```

### 4. Non-critical errors → log and continue

Don't fail the whole operation for non-critical errors. Log to stderr and continue:

```go
// ✓ Good — audit logging is best-effort
auditLogger, err := audit.NewAuditLogger(logDir)
if err != nil {
    fmt.Fprintf(os.Stderr, "Warning: failed to create audit logger: %v\n", err)
}
```

## Tool Execution Errors

Tools return errors as `ToolResult{IsError: true}`, NOT as Go errors:

```go
func (t *FileReadTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
    // Validation errors → IsError result, nil Go error
    if args.Path == "" {
        return agent.ValidationError("path is required"), nil
    }
    // System errors → IsError result, nil Go error
    if os.IsNotExist(err) {
        return agent.ToolResult{Content: "file not found", IsError: true}, nil
    }
}
```

## Error Categories

Use `agent.ErrorCategory` for classification:

| Category | Use Case | Constructor |
|----------|----------|-------------|
| `ErrCategoryValidation` | Bad input / missing args | `agent.ValidationError(msg)` |
| `ErrCategoryPermission` | Access denied | `agent.PermissionError(msg)` |
| `ErrCategoryTransient` | Retryable error | `agent.TransientError(msg)` |
| `ErrCategoryBusiness` | Logic / business error | `agent.BusinessError(msg)` |

## Constructor Patterns

- Simple: `func NewXxx() *Xxx` (returns pointer, no error)
- Complex: `func NewXxx() (*Xxx, error)` (may fail)

## Anti-Patterns

- ❌ Don't use `panic()` for expected errors — only for programmer mistakes
- ❌ Don't silently ignore errors — at minimum add a `// intentionally ignored` comment
- ❌ Don't log AND return the same error — return it, let the caller decide
- ❌ Don't use `errors.New()` when `fmt.Errorf()` with `%w` gives more context
