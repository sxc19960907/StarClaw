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

## Scenario: OpenAI-Compatible Local Gateway

### 1. Scope / Trigger

- Trigger: adding or changing daemon endpoints that mimic OpenAI API shapes for local StarClaw/Astria execution.
- Scope: local daemon HTTP API only. The gateway adapts external tool requests into existing `RunAgentRequest` execution and must not bypass daemon permissions, approval, session, run-store, or local-only assumptions.

### 2. Signatures

- Route: `POST /v1/chat/completions`.
- Handler: `(*Server).handleOpenAIChatCompletions`.
- Minimum request:
  ```json
  {
    "model": "local-model-or-agent-model",
    "messages": [
      {"role": "user", "content": "hello"}
    ]
  }
  ```
- Optional local extension fields:
  - `request_id`: reuse as run id and response id suffix.
  - `session_id`: resume an existing StarClaw session.
  - `agent`: run a named StarClaw agent.
  - `user`: copied to `RunAgentRequest.Sender`.
- Response envelope:
  - `id`: `chatcmpl-<request_id>`.
  - `object`: `chat.completion`.
  - `created`: Unix timestamp.
  - `model`: request model value.
  - `choices[0].message.role`: `assistant`.
  - `choices[0].message.content`: joined local run messages.
  - `choices[0].finish_reason`: `stop`.
  - `usage.prompt_tokens`, `completion_tokens`, `total_tokens`: mapped from local usage when available.
  - `starclaw_run_id`: local run id for `/runs/{id}` lookup.

### 3. Contracts

- The endpoint must call the same `s.runAgent` path used by `/message`; do not create a direct LLM client path.
- The gateway must create a run record through `RunStore.Start` / `Complete` so `/runs` and `/runs/{id}` include gateway runs.
- Supported message roles are `system`, `user`, and `assistant`. `user` content is passed as-is; non-user context is prefixed as `<role>: <content>` in the local prompt.
- The request `model` may override the effective agent config model through `RunAgentRequest.Model`, but other model parameters are intentionally unsupported unless a future task defines their local contract.
- Unsupported fields must return an OpenAI-style error envelope with `error.message` and `error.type="invalid_request_error"`; do not silently ignore them.
- `stream=true`, OpenAI tool/function calling fields, `response_format`, metadata, and `n > 1` are unsupported.

### 4. Validation & Error Matrix

- Missing `model` -> HTTP 400.
- Missing `messages` -> HTTP 400.
- Empty message content -> HTTP 400.
- Unsupported role -> HTTP 400.
- Unknown JSON field -> HTTP 400 naming the unsupported field.
- `stream=true` -> HTTP 400.
- `tools`, `functions`, `function_call`, or `tool_choice` present -> HTTP 400.
- Local run failure -> OpenAI-style HTTP 500 error envelope.

### 5. Good/Base/Bad Cases

- Good: a minimal chat-completions request returns one assistant choice, usage, and a run id discoverable via `/runs/{id}`.
- Base: request id omitted; daemon generates one and still returns a valid `chatcmpl-*` id.
- Bad: accepting `parallel_tool_calls`, `stream`, or `response_format` while doing nothing with them.

### 6. Tests Required

- Route registration test for `POST /v1/chat/completions`.
- Handler success test covering response envelope, usage mapping, run source, sender, and prompt conversion.
- Validation tests for required fields, unsupported fields, roles, streaming, tool/function fields, and multi-choice requests.
- Runner test proving request model overrides config model via `ChatOptions.SpecificModel`.

### 7. Wrong vs Correct

#### Wrong

```go
resp, err := s.deps.LLMClient.Chat(ctx, "", messages, nil, maxTokens, nil)
```

This bypasses daemon permissions, run records, approval, sessions, and configured agent overlays.

#### Correct

```go
result, err := s.runAgent(ctx, RunAgentRequest{
    Text: prompt,
    Source: "openai-compatible",
    Channel: ChannelHTTP,
}, handler)
```

## Scenario: Structured Runtime Observability

### 1. Scope / Trigger

- Trigger: adding or changing daemon run lifecycle events, runtime metrics, tracing export foundations, or any API that exposes run observability.
- Scope: local daemon HTTP API and in-memory run store. This is an observability surface, not a prompt archive or raw provider transcript.

### 2. Signatures

- Structured event type: `StructuredRunEvent`.
- Schema fields:
  ```json
  {
    "id": "run-id-000001",
    "schema_version": "2026-06-08",
    "run_id": "run-id",
    "type": "run_started",
    "phase": "start",
    "at": "2026-06-08T00:00:00Z",
    "data": {}
  }
  ```
- Run record field: `RunRecord.StructuredEvents`, serialized as `structured_events`.
- Metrics route: `GET /metrics`.
- Response shape:
  ```json
  {
    "metrics": {
      "runs_total": 1,
      "runs_by_status": {"completed": 1},
      "events_by_type": {"run_started": 1},
      "tokens_input_total": 10,
      "tokens_output_total": 20,
      "schema_version": "2026-06-08",
      "stored_run_limit": 100
    }
  }
  ```

### 3. Contracts

- Every structured event must include `id`, `schema_version`, `run_id`, `type`, `phase`, and `at`.
- Event ids are deterministic within a run using a monotonically increasing sequence suffix.
- Runtime must emit structured events for run start, run completion, run error, tool events, usage, budget status, routing decisions, and fallback decisions where those concepts are present.
- `GET /metrics` returns aggregate counters/gauges only. It must not include prompts, tool arguments, provider payloads, raw responses, or user text.
- Existing unstructured run events and SSE/Web UI behavior must remain compatible when structured observability is added.
- Structured observability is local-first. Detailed payload export remains opt-in and must not be introduced implicitly through metrics.

### 4. Validation & Error Matrix

- Prompt or assistant text in event data -> redact field to a boolean marker such as `text_redacted`.
- Tool args, request, response, delta, content, preamble, or prompt bodies -> redact by default.
- Values or keys containing API key, token, secret, password, or bearer credentials -> replace with `[REDACTED]`.
- Missing optional routing/fallback/budget response data -> omit that specific structured decision event.
- Unknown event type -> keep the type and use a generic runtime phase.

### 5. Good/Base/Bad Cases

- Good: a completed run records `run_started`, usage/tool events, optional `budget_status`, optional `routing_selected`, optional `fallback_decision`, and `run_completed`.
- Base: `/metrics` reports run counts, event counts, token totals, schema version, and stored run limit.
- Bad: metrics or structured events expose a prompt body, tool args JSON, bearer token, provider response text, or API key.

### 6. Tests Required

- Unit test structured event schema fields and per-run sequence behavior.
- Unit test redaction for prompt text, assistant text, tool args/content, and secret-like values.
- Unit test metric shape and stable counter names.
- HTTP handler test for `GET /metrics`.
- Route registration test for `GET /metrics`.
- Existing SSE/Web UI smoke tests must continue to pass when observability changes touch shared run/event code.

### 7. Wrong vs Correct

#### Wrong

```go
store.AddEvent(id, EventToolCall, map[string]any{"args": rawArgsJSON})
record.StructuredEvents = append(record.StructuredEvents, StructuredRunEvent{Data: map[string]any{"args": rawArgsJSON}})
```

#### Correct

```go
store.AddEvent(id, EventToolCall, map[string]any{"args": rawArgsJSON})
// The structured copy redacts args while the legacy run event remains compatible.
s.addStructuredEventLocked(id, EventToolCall, map[string]any{"args": rawArgsJSON})
```

## Scenario: Workflow Control API

### 1. Scope / Trigger

- Trigger: adding or changing daemon APIs that control active or historical runs, including cancel, pause, resume, replay, or run-state inspection.
- Scope: local daemon workflow control. This must preserve existing Web UI stop behavior and must not bypass run records, approval boundaries, or structured observability.

### 2. Signatures

- Compatibility route: `POST /cancel`.
- Run control route: `POST /runs/{id}/control`.
- Request shape:
  ```json
  {
    "action": "cancel",
    "reason": "operator stop",
    "approved": false
  }
  ```
- Supported action values:
  - `cancel`
  - `pause`
  - `resume`
  - `replay`
- Run metadata type: `RunControlDecision`.
- Run record field: `RunRecord.Control`, serialized as `control`.
- Structured event type: `control_decision`.

### 3. Contracts

- `POST /cancel` remains compatible with existing callers that send `request_id`.
- `POST /runs/{id}/control` is the canonical per-run control surface for new callers.
- `cancel` must call the active run cancel function when the run is active and record a `cancelled` control decision.
- `pause` and `resume` return HTTP `409` until runtime support exists. Do not simulate pause/resume by mutating run status alone.
- `replay` returns an approval-required replay plan only. It must not launch a new run or repeat tool calls in this slice.
- Replay plans must redact prompt text by default and must not include tool arguments, raw provider payloads, or external-delivery side effects.
- Control decisions must be visible on `GET /runs/{id}` and as structured `control_decision` events.

### 4. Validation & Error Matrix

- Missing `request_id` on `POST /cancel` -> HTTP `400`.
- Unknown request id on `POST /cancel` -> HTTP `404`.
- Missing `action` on `POST /runs/{id}/control` -> HTTP `400`.
- Unknown action -> HTTP `400`.
- Missing run -> HTTP `404`.
- Cancel for a known but inactive run -> HTTP `409`.
- Pause/resume for a known run -> HTTP `409`.
- Replay for a known run -> HTTP `200` with `status="approval_required"`.

### 5. Good/Base/Bad Cases

- Good: current Web UI stop button still posts to `/cancel`, active context is cancelled, run detail records the decision, and structured events record `control_decision`.
- Base: `POST /runs/{id}/control` with `replay` returns a redacted approval-required plan for the stored run.
- Bad: replay immediately calls `runAgent`, repeats tool calls, or returns the original prompt body.

### 6. Tests Required

- Route registration test for `POST /runs/{id}/control`.
- Cancel handler test preserving `POST /cancel` compatibility and recording run metadata.
- Run control cancel test covering metadata and structured control event.
- Pause/resume tests proving staged `409` behavior.
- Replay test proving approval-required response and prompt redaction.
- Validation tests for missing action, unknown action, missing run, and inactive cancel.
- Existing run smoke tests must continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
// Replays immediately and may repeat destructive tool calls.
result, err := s.runAgent(ctx, record.Request, handler)
```

#### Correct

```go
writeJSON(w, http.StatusOK, map[string]any{
    "status": "approval_required",
    "replay": replayPlan,
})
```

## Scenario: Runtime Complexity Routing and Fallback

### 1. Scope / Trigger

- Trigger: adding or changing deterministic route/model recommendations, fallback decisions, or daemon run metadata for runtime routing.
- Scope: backend runtime only. Classification must be local and deterministic; it must not issue provider calls or tool calls.

### 2. Signatures

- Classifier input: `agent.RoutingInput`.
- Classifier output: `agent.RouteRecommendation`.
- Fallback input: `agent.FallbackInput`.
- Fallback output: `*agent.FallbackDecision`.
- Daemon response fields:
  - `RunAgentResponse.Routing` serialized as `routing`.
  - `RunAgentResponse.Fallback` serialized as `fallback`.
- Run record fields:
  - `RunRecord.Routing` serialized as `routing`.
  - `RunRecord.Fallback` serialized as `fallback`.

### 3. Contracts

- `RecommendRoute` must be pure and deterministic from local signals: prompt text, requested tools, token budget, and local failure counts.
- Route recommendations are advisory metadata. They must not bypass permissions, approval, session handling, or existing `RunAgent` execution.
- Budget-constrained routes take precedence when a hard token budget is configured.
- Delivery-sensitive prompts must be routed to a review-oriented route; do not auto-deliver externally.
- `RecommendFallback` must expose the reason for provider errors, budget exhaustion, and repeated same-class failures.
- `RunStore.Get` must defensively copy routing and fallback pointers, as it does for other pointer metadata.

### 4. Validation & Error Matrix

- Simple prompt -> direct route, small model tier.
- Evidence-heavy prompt or evidence tools -> research route, medium model tier.
- Council/tradeoff prompt -> council route, high model tier.
- External delivery/risk prompt -> delivery review route, medium model tier.
- Hard token budget -> budget guard route, small model tier.
- Provider error -> provider-error fallback.
- Budget exhausted -> budget-exhausted fallback.
- Repeated failures >= 2 -> repeated-failure fallback.

### 5. Good/Base/Bad Cases

- Good: a run response includes routing metadata before any remote provider call is needed.
- Base: no fallback is returned for a successful run without budget exhaustion.
- Bad: classifier calls an LLM, executes a tool, or silently routes external delivery to direct execution.

### 6. Tests Required

- Classifier unit tests for simple, evidence-heavy, council-worthy, delivery-sensitive, budget-constrained, and tool-requested evidence prompts.
- Fallback unit tests for provider error, budget exhaustion, repeated failure, and no-fallback cases.
- Daemon tests proving run responses expose routing and budget fallback metadata.
- Existing daemon and full project tests must pass.

### 7. Wrong vs Correct

#### Wrong

```go
resp, err := llm.Chat(ctx, "", []client.Message{{Role: "user", Content: prompt}}, nil, 256, nil)
```

This makes classification paid, non-deterministic, and dependent on provider availability.

#### Correct

```go
routing := agent.RecommendRoute(agent.RoutingInput{
    Prompt: prompt,
    TokenBudget: budget,
})
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
