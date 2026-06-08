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

### Browser Smoke Tests

- When a UI click hydrates an editor from an API request, wait for the matching response before editing form fields. Otherwise the late response can overwrite test input and make save assertions flaky.

## Scenario: Agent Command Editor

### 1. Scope / Trigger

- Trigger: daemon API and Web UI create/update named agent custom command files.

### 2. Signatures

- `POST /agents` and `PUT /agents/{name}` accept optional `commands`.
- Request shape:
  ```json
  {
    "commands": {
      "review": "Review recent changes."
    }
  }
  ```

### 3. Contracts

- Command names map to `<agent>/commands/<name>.md`.
- Command names must match `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`.
- Omitted `commands` preserves existing command files.
- Present `commands` replaces the managed command directory with non-empty entries.
- Empty command content is omitted; an empty object removes the command directory.

### 4. Validation & Error Matrix

- Invalid command name -> `400`.
- Filesystem write/remove failure -> `400` from create/update path.
- Missing command body in Web UI -> client-side toast, no API call.

### 5. Good/Base/Bad Cases

- Good: add `review`, save agent, reload detail, `Commands.review` returns the saved body.
- Base: update agent prompt with no `commands` field, existing command files remain.
- Bad: command name `../escape` is rejected.

### 6. Tests Required

- Backend create persists commands.
- Backend update replaces commands and removes deleted entries.
- Backend update without `commands` preserves existing command files.
- Backend rejects invalid command names.
- Web UI smoke covers add/edit/delete command round trip.

### 7. Wrong vs Correct

#### Wrong

```go
path := filepath.Join(commandsDir, name+".md")
```

without validating `name`.

#### Correct

```go
if !agentCommandNameRe.MatchString(name) {
    return fmt.Errorf("invalid command name")
}
```

## Scenario: Runtime Token Budget Enforcement

### 1. Scope / Trigger

- Trigger: adding or changing local token budget limits, runtime usage tracking, daemon run response/status fields, or run event persistence.
- Scope: backend runtime only; the budget guard UI can configure or display this state, but enforcement belongs in `internal/agent`.

### 2. Signatures

- Config path: `agent.token_budget`.
- Fields:
  - `max_input_tokens`: integer; `0` disables the input limit.
  - `max_output_tokens`: integer; `0` disables the output limit.
  - `max_total_tokens`: integer; `0` disables the total limit.
  - `hard_stop`: boolean; when true, stop before the next model call once a concrete or projected limit is exhausted.
- Runtime conversion: `agent.TokenBudgetFromAgent(config.AgentConfig) agent.TokenBudget`.
- Loop API:
  - `(*AgentLoop).SetTokenBudget(agent.TokenBudget)`
  - `(*AgentLoop).LastBudgetStatus() agent.TokenBudgetUsage`
- Daemon response field: `RunAgentResponse.BudgetStatus` serialized as `budget_status`.
- Run record field: `RunRecord.Budget` serialized as `budget_status`.
- Run event type: `budget_status`.

### 3. Contracts

- Budget tracking is per run. `AgentLoop.Run` must reset the tracker at run start from the configured budget.
- Provider usage is authoritative when `client.Response.Usage.InputTokens` or `OutputTokens` is non-zero.
- Missing provider usage must produce `status="unknown"` and increment `unknown_turns`; do not invent precise totals.
- Before an initial or follow-up model call, the loop may use `context.EstimateTokens(messages)` plus the next request `max_tokens` as a conservative projection.
- Hard-stop returns a normal `client.Response` with `StopReason="budget_exhausted"` and a clear content message instead of continuing tool/model loops.
- Daemon responses and run records surface only counts/status/detail. They must not include prompts, API keys, provider headers, or raw request bodies inside `budget_status`.

### 4. Validation & Error Matrix

- No budget configured -> `LastBudgetStatus().Status == "disabled"` and no daemon `budget_status`.
- Usage under all configured limits -> `status="ok"`.
- Usage at or over a configured limit -> `status="exhausted"`.
- `hard_stop=false` with exhausted budget -> status is exhausted but the loop does not stop early.
- `hard_stop=true` with exhausted/projected budget -> stop before the next model call.
- Provider response has zero usage -> `status="unknown"` unless an existing exhausted state already applies.

### 5. Good/Base/Bad Cases

- Good: a tool loop reaches projected total budget and returns a `budget_exhausted` response without issuing another LLM call.
- Base: a simple one-shot run records provider usage and daemon `budget_status` with `ok`.
- Bad: code continues a follow-up model call after `max_total_tokens` is exhausted, or reports exact token totals when provider usage was missing.

### 6. Tests Required

- Unit tests for budget decision cases: under budget, at budget, over budget, soft budget, and unknown usage.
- Agent loop tests proving hard-stop before initial and follow-up model calls.
- Config tests for global defaults, multi-level overlay, and per-agent pointer overrides.
- Daemon tests proving `RunAgentResponse` surfaces `budget_status` and run records copy it defensively.

### 7. Wrong vs Correct

#### Wrong

```go
// Missing usage treated as zero, so the run falsely looks under budget.
tracker.AddUsage(client.Usage{})
status := tracker.Status() // ok
```

#### Correct

```go
status := tracker.AddUsage(client.Usage{})
// status.Status == "unknown"; callers know the provider did not return usage.
```

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
