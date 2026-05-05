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
