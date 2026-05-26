# Quality Guidelines

> Code quality standards for StarClaw

---

## Core Principles

1. **Read Before Write** — Understand context before coding
2. **Follow Existing Patterns** — Match the codebase, don't introduce new styles
3. **Test Everything** — Every new package needs tests
4. **Minimal Dependencies** — Justify every new import

---

## Code Organization

### One Tool Per File

Each tool is a separate file in `internal/tools/`. Pattern:

```go
type FileReadTool struct{}

type fileReadArgs struct {
    Path   string `json:"path"`
    Offset int    `json:"offset,omitempty"`
}

func (t *FileReadTool) Info() agent.ToolInfo { ... }
func (t *FileReadTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) { ... }
func (t *FileReadTool) RequiresApproval() bool { return false }
```

### Constructor Convention: `NewXxx()`

```go
func NewAgentLoop(llmClient LLMClient, registry *ToolRegistry) *AgentLoop
func NewClientManager() *ClientManager
```

Return `*Type` for simple cases, `(*Type, error)` when initialization can fail.

---

## Testing Standards

### Table-Driven Tests (preferred)

```go
tests := []struct {
    name     string
    input    string
    expected bool
}{
    {"valid", "my-agent", true},
    {"uppercase", "My-Agent", false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### Test File Placement

- Unit tests: `*_test.go` alongside source (same package)
- Integration tests: `tests/` directory (blackbox `package tests`)

### Mocking

- Define mock types in the same test file
- Implement the full interface
- Verify at compile time: `var _ Interface = (*MockType)(nil)`
- Shared test mocks must be concurrency-safe when used across goroutines. If a mock records slice or map inputs, store defensive copies and return defensive copies from getters.

### Timers and Callbacks

- For resettable/stoppable timers, guard callbacks with a generation or cancellation token so callbacks queued before `Reset` or `Stop` cannot fire after the state changed.
- For long waits that race against context cancellation, prefer `time.NewTimer` plus `Stop`/non-blocking drain over `time.After` so cancellation releases timer resources promptly.
- Retry/backoff helpers that wait on context cancellation should return `ctx.Err()` to their caller so retry loops stop instead of issuing another operation with an already-cancelled context.
- Do not share plain `bytes.Buffer` instances between `os/exec` stdout/stderr copy goroutines and status readers. Use a synchronized writer/reader wrapper or wait for the process to exit before reading captured output.

### Streaming Parsers

- For line-delimited streaming protocols such as SSE, flush any accumulated event after EOF when the scanner ended without error. Producers may legally close the connection without a trailing blank-line delimiter, and dropping that pending event loses the final status/result.
- Add a regression test for streams that end immediately after the final `data:` line.

---

## Interface Design

### Small, focused interfaces

```go
type Tool interface {
    Info() ToolInfo
    Run(ctx context.Context, args string) (ToolResult, error)
    RequiresApproval() bool
}

// Optional behaviors via separate interfaces
type SafeChecker interface { IsSafeArgs(argsJSON string) bool }
type ReadOnlyChecker interface { IsReadOnlyCall(argsJSON string) bool }
```

---

## Configuration

### Triple-tag convention

```go
type Config struct {
    Endpoint string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
}
```

### Config activation checklist

When adding or activating a config field, keep every config path in sync:

- Struct tags on the config type.
- `viper.SetDefault` in `Load` for global config defaults.
- The default YAML emitted by `SaveDefault`.
- Setup-generated config in `setup.go`, when the field belongs in first-run config.
- Multi-level defaults and overlay logic in `multilevel.go`, when project or local config should override it.
- Unit tests for parsing/overlay behavior when the field affects runtime behavior.

Missing one path can make a setting work in chat/TUI but silently fail in project-level config or server mode.

### Pointer fields for optional values

Use `*string`, `*int`, `*bool` to distinguish "not set" from zero:

```go
type AgentModelConfig struct {
    Model         *string  `yaml:"model,omitempty"`
    MaxIterations *int     `yaml:"max_iterations,omitempty"`
}
```

---

## Security

### Path validation

All file ops use `SafePath` before accessing the filesystem:
```go
safePath, err := validatePath(args.Path, ".")
```

### Sensitive data

- API keys: `strings.TrimSpace()`, stored `0600`
- Audit logs: redaction via `internal/audit/redaction.go`
- No secrets in stdout

---

## Anti-Patterns

- ❌ Don't add comments that explain WHAT — code should be self-documenting
- ❌ Don't create premature abstractions — three similar lines > one premature helper
- ❌ Don't add error handling for impossible scenarios
- ❌ Don't ship commented-out code — git history preserves it
- ❌ Don't create `utils.go` / `helpers.go` files
- ❌ Don't add features "for future use" (YAGNI)
