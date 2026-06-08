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

## Scenario: Observability Trace Export

### 1. Scope / Trigger

- Trigger: exporting structured run events to local trace artifacts or adding trace-read HTTP endpoints.
- Scope: local structured event export only. This is not an external collector, cloud telemetry, prompt archive, or OpenTelemetry SDK integration.

### 2. Signatures

- Trace export record type: `TraceExportRecord`.
- Store APIs:
  - `(*RunStore).TraceEvents(runID string) ([]TraceExportRecord, bool)`
  - `(*RunStore).AllTraceEvents() []TraceExportRecord`
  - `(*RunStore).ExportTracesJSONL(path string) error`
  - `(*RunStore).ExportRunTraceJSONL(runID, path string) error`
- HTTP routes:
  - `GET /runs/{id}/trace`
  - `GET /traces/export?path=/local/file.jsonl`
- JSONL line shape:
  ```json
  {
    "schema_version": "2026-06-08",
    "trace_id": "run-id",
    "span_id": "run-id-000001",
    "run_id": "run-id",
    "event_id": "run-id-000001",
    "name": "run_started",
    "phase": "start",
    "timestamp": "2026-06-08T00:00:00Z",
    "attributes": {}
  }
  ```

### 3. Contracts

- Export JSONL must contain exactly one valid JSON object per structured event.
- Export writes use temp-file plus rename semantics.
- Export must be local-only and caller-directed; do not send traces to remote services.
- Export records are derived from `StructuredRunEvent`, not legacy raw events or run request/response bodies.
- Export attributes must be recursively sanitized before writing.
- Export must not include prompt text, assistant text, tool args, raw provider payloads, request/response bodies, API keys, tokens, passwords, bearer credentials, or secret-looking values.
- Existing `/metrics`, `/runs`, `/runs/{id}`, SSE, replay, pause/resume, and persistence behavior must remain compatible.

### 4. Validation & Error Matrix

- Missing export path -> contextual error / HTTP `400`.
- Missing run for single-run export -> contextual error / HTTP `404`.
- Empty trace set -> create an empty JSONL file successfully.
- Destination parent missing -> create local parent directory with private permissions.
- Export encode/write/rename failure -> return contextual error and remove temp file when possible.

### 5. Good/Base/Bad Cases

- Good: exporting a run with tool events produces JSONL with event ids and redacted attributes.
- Base: `GET /runs/{id}/trace` returns the in-memory trace records for a known run.
- Bad: exporting from `RunRecord.Events` leaks raw tool args, or `/traces/export` uploads to a collector by default.

### 6. Tests Required

- JSONL export test for all stored runs.
- Single-run trace HTTP test.
- Missing-run trace test.
- Redaction test for prompt text, tool args, request/response fields, and secret-like values.
- Existing structured event, metrics, and persistent run-store tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
for _, evt := range record.Events {
    encoder.Encode(evt)
}
```

Legacy events may contain raw tool args or text payloads.

#### Correct

```go
records := store.AllTraceEvents()
writeTraceJSONL(path, records)
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

## Scenario: Durable Run Store

### 1. Scope / Trigger

- Trigger: adding or changing daemon run-store persistence, recovery, durable run metadata, or restart behavior for Mission Control.
- Scope: local daemon run metadata only. This is not workflow step execution, cloud sync, or replay execution.

### 2. Signatures

- In-memory constructor: `NewRunStore(limit int) *RunStore`.
- Persistent constructor: `NewPersistentRunStore(limit int, path string) (*RunStore, error)`.
- Persistence envelope:
  ```json
  {
    "schema_version": "2026-06-08",
    "records": []
  }
  ```
- Last persistence/load error: `(*RunStore).PersistError() error`.

### 3. Contracts

- `NewRunStore` remains in-memory and compatible with existing tests and server setup.
- Persistent mode writes JSON to a local file with temp-file plus rename semantics.
- Persistent mode must recover run records, summaries, structured events, legacy events, control decisions, usage, budget, routing, fallback, response, and terminal status.
- Recovery must respect the configured store limit.
- Recovery must rebuild per-run structured event sequence counters so new events continue monotonically.
- Corrupt persistence data must not panic. Constructor may return an error with a safe empty store so callers/tests can decide how to surface it.
- Persistence must not add prompts, tool args, provider responses, or secrets to metrics or trace exports.

### 4. Validation & Error Matrix

- Missing persistence file -> empty store, nil error.
- Empty persistence file -> empty store, nil error.
- Corrupt JSON -> empty store plus contextual error.
- More records than limit -> newest/order-preserved records retained up to limit.
- Duplicate or blank record ids -> ignore invalid duplicates during recovery.
- Save failure -> mutation API remains compatible; `PersistError()` exposes the last failure.

### 5. Good/Base/Bad Cases

- Good: a completed run with usage, budget, routing, structured events, and control decisions survives store reconstruction.
- Base: in-memory `NewRunStore` behavior is unchanged.
- Bad: daemon startup panics on a corrupt `runs.json`, or recovered structured events restart numbering at `000001`.

### 6. Tests Required

- Restart recovery test for completed runs and metadata.
- Control-decision and structured-event recovery test.
- Corrupt-file tolerance test.
- Limit enforcement on load test.
- Existing run/control/metrics daemon tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
data, _ := os.ReadFile(path)
_ = json.Unmarshal(data, &records)
```

This silently ignores corrupt state and hides recovery failures.

#### Correct

```go
store, err := NewPersistentRunStore(defaultRunStoreLimit, path)
// err is available to the caller, while store remains safe to use.
```

## Scenario: Durable Workflow Step State

### 1. Scope / Trigger

- Trigger: adding or changing durable step-level workflow state, mission step timelines, or future replay/pause/resume state foundations.
- Scope: local daemon run records and step metadata only. This is not workflow graph execution, cloud sync, or replay execution.

### 2. Signatures

- Step state type: `WorkflowStepState`.
- Run record field: `RunRecord.Steps`, serialized as `steps`.
- Store APIs:
  - `(*RunStore).UpsertStep(runID string, step WorkflowStepState) bool`
  - `(*RunStore).TransitionStep(runID, stepID, status string, metadata map[string]any) bool`
- Structured event type: `workflow_step`.
- Status values:
  - `planned`
  - `running`
  - `blocked`
  - `waiting_approval`
  - `completed`
  - `failed`
  - `cancelled`
  - `skipped`

### 3. Contracts

- Step state is per run and must be recoverable through the persistent run-store envelope.
- Step transitions must not mutate the run's terminal status by themselves; run status remains owned by run lifecycle/control APIs.
- First upsert defaults blank status to `planned` and attempt to `1`.
- Entering `running` sets `started_at` when absent.
- Terminal step statuses set `ended_at`.
- Step metadata stored in run detail must be redacted/sanitized before persistence.
- Structured `workflow_step` events must go through the same redaction path as other structured events.
- Metrics may count `workflow_step` events, but must not include step ids, titles, metadata, prompts, tool args, provider payloads, or secrets.

### 4. Validation & Error Matrix

- Missing run id or unknown run -> mutation returns `false`.
- Blank step id -> mutation returns `false`.
- Blank upsert status -> stored as `planned`.
- Unknown status is not interpreted as a run lifecycle transition; future stricter validation must preserve compatibility deliberately.
- Corrupt persistent run store -> safe empty store plus contextual error, same as durable run-store behavior.

### 5. Good/Base/Bad Cases

- Good: a recovered run includes step timeline and can continue appending monotonic structured step events.
- Base: run detail exposes durable step state; run summaries remain unchanged.
- Bad: setting a step to `completed` marks the whole run completed, or metrics expose step metadata.

### 6. Tests Required

- Upsert and transition tests for timestamps, attempts, ordering, and metadata merge.
- Persistence recovery test for step state.
- Redaction test proving step metadata is safe in run detail and structured events.
- Metrics test proving only aggregate `workflow_step` counts are exposed.
- Existing run/control/metrics and persistent run-store tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
record.Status = WorkflowStepCompleted
record.Steps = append(record.Steps, stepWithRawToolArgs)
```

This confuses step lifecycle with run lifecycle and can leak tool metadata through run detail.

#### Correct

```go
store.TransitionStep(runID, stepID, WorkflowStepCompleted, safeMetadata)
// The run lifecycle remains controlled by Start/Complete/control APIs.
```

## Scenario: Safe Replay Execution Boundary

### 1. Scope / Trigger

- Trigger: changing replay control behavior from plan-only to approved replay launch, or adding any API that can re-run historical work.
- Scope: local daemon run control and execution boundary. This is not deterministic replay, recorded tool-output playback, cloud sync, or frontend replay UI.

### 2. Signatures

- Route: `POST /runs/{id}/control`.
- Request shape:
  ```json
  {
    "action": "replay",
    "approved": true,
    "reason": "operator approved"
  }
  ```
- Plan helper: `replayControlPlan(sourceRunID string, req RunAgentRequest, approved bool) map[string]any`.
- Replay request helper: `replayRunRequest(source RunAgentRequest, sourceRunID string) RunAgentRequest`.
- Source control metadata: `RunControlDecision{Action:"replay", Status:"approval_required|approved"}`.
- Source/replay step metadata uses durable `WorkflowStepState` without prompt text.

### 3. Contracts

- `approved=false` or omitted must remain plan-only and must not create a replay run.
- `approved=true` is the explicit boundary required before launching replay.
- Approved replay must generate a new replay run id. It must not reuse or overwrite the source run id.
- Approved replay must execute through `s.runAgent` and `s.recordingHandler`, preserving daemon approval requester, permissions, run events, usage, routing, fallback, and run-store behavior.
- Approved replay must link source and replay ids in control/step metadata.
- Replay control responses must redact source prompt text and must not expose tool args, raw provider payloads, or external side-effect payloads.
- Replay step/control metadata must not store source prompt text.
- Approved replay must not mutate the source run terminal status.
- Metrics remain aggregate-only and must not include source prompt text or replay request metadata.

### 4. Validation & Error Matrix

- Missing run -> HTTP `404`.
- Missing action -> HTTP `400`.
- Unsupported action -> HTTP `400`.
- Replay without approval -> HTTP `200`, `status="approval_required"`, no new run.
- Replay with approval -> HTTP `200`, `status="launched"`, response includes `source_run_id` and `replay_run_id`.
- Replay run execution error -> replay run is completed as error and handler returns HTTP `500` with the execution error.

### 5. Good/Base/Bad Cases

- Good: approved replay launches a new run through the normal daemon path, and a tool requiring approval still emits `approval_needed`.
- Base: unapproved replay returns a redacted plan identical in safety to the previous behavior.
- Bad: replay calls lower-level agent code without the daemon approval requester, reuses the source run id, or returns the original prompt in the control response.

### 6. Tests Required

- Unapproved replay test proving no replay run is created and prompt is redacted.
- Approved replay test proving a new replay run is created, executed, and linked to the source run.
- Source run status preservation test.
- Control/step metadata test for approval boundary.
- Metrics/control response redaction tests.
- Existing control validation and approval tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
result, err := RunAgent(ctx, s.deps, source.Request, nil)
```

This bypasses daemon approval requester wiring and loses run recording context.

#### Correct

```go
handler := s.recordingHandler(replayRunID, &httpEventHandler{})
result, err := s.runAgent(ctx, replayReq, handler)
```

## Scenario: Runtime Pause Resume Support

### 1. Scope / Trigger

- Trigger: changing daemon pause/resume behavior for active runs, agent loop cooperative pause points, or active runtime control state.
- Scope: local daemon runtime control only. This is not OS-level process suspension, deterministic replay, frontend UI, or persisted process resurrection after daemon death.

### 2. Signatures

- Agent loop interface: `agent.PauseController`.
- Agent loop setter: `(*AgentLoop).SetPauseController(controller PauseController)`.
- Daemon runtime handle: `runtimeHandle`.
- Daemon pause controller: `runtimePauseController`.
- Run control route: `POST /runs/{id}/control`.
- Supported runtime statuses:
  - `paused`
  - `resumed`
  - `not_running`
  - `cancelled`

### 3. Contracts

- Pause/resume only succeeds for active runs present in `Server.running`.
- Known inactive runs must return HTTP `409`; missing runs must return HTTP `404`.
- Pause is cooperative: the agent loop waits before model calls and before tool calls. It must not preempt a tool already executing.
- Resume releases all waiters and allows execution to continue.
- Cancel must unblock a paused run before cancelling context so the run can terminate promptly.
- Runtime controls must preserve daemon approval requester and tool permission behavior.
- Pause/resume must not mutate run terminal status by themselves.
- Control decisions must be visible on run detail and as structured `control_decision` events.
- Runtime pause step state must be visible as durable `WorkflowStepState` without prompt/tool/provider payloads.

### 4. Validation & Error Matrix

- Active run pause -> HTTP `200`, status `paused`.
- Active paused run resume -> HTTP `200`, status `resumed`.
- Active unpaused run resume -> HTTP `200`, status `resumed`.
- Known inactive pause/resume -> HTTP `409`, control status `not_running`.
- Missing run pause/resume -> HTTP `404`.
- Context cancelled while paused -> agent loop returns a contextual cancellation error.

### 5. Good/Base/Bad Cases

- Good: an active run paused before a model call does not call the LLM until resume.
- Base: cancelling a paused run records cancelled control metadata and does not hang.
- Bad: pause only writes `RunRecord.Status="paused"` while the runtime keeps making model/tool calls.

### 6. Tests Required

- Agent loop test proving pause blocks before a model call and resumes.
- Agent loop test proving cancellation exits while paused.
- Daemon pause/resume active run test.
- Daemon pause/resume inactive and missing run tests.
- Cancel tests proving active paused handles still cancel.
- Existing replay, metrics, approval, and run control tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
record.Status = "paused"
writeJSON(w, http.StatusOK, record)
```

This changes metadata only; the runtime still continues model/tool execution.

#### Correct

```go
handle.pause.Pause()
s.runStore.AddControlDecision(runID, RunControlDecision{Action: "pause", Status: "paused"})
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

## Scenario: Kocoro-Style Workflow Commands

### 1. Scope / Trigger

- Trigger: adding or changing daemon-recognized slash workflow commands that enter through `POST /message`.
- Scope: local daemon command parsing and workflow step metadata. This is not a separate workflow endpoint, cloud channel transport, Desktop RPC, or OpenAI-compatible gateway behavior.

### 2. Signatures

- Request endpoint: `POST /message`.
- Request field: `RunAgentRequest.Text`.
- Recognized command forms:
  - `/research <goal>`
  - `/swarm <goal>`
- Internal parser: `parseWorkflowInvocation(text string) (*workflowInvocation, error)`.
- Execution path: `(*Server).runWorkflowAgent(ctx, req, invocation, handler)`.

### 3. Contracts

- Command validation must happen before `RunStore.Start`; invalid empty command goals must not create run records.
- Recognized workflow commands must still execute through `s.runAgent` / `RunAgentWithApproval`; do not call an LLM client directly.
- JSON and `Accept: text/event-stream` behavior for `/message` must remain compatible.
- Non-command prompts and unknown slash commands continue through the existing direct run path unchanged.
- Workflow commands may transform the prompt before agent execution, but run metadata, approval, pause/resume/cancel, session handling, routing/fallback, and structured observability stay under the existing daemon pipeline.
- Workflow step metadata may include `workflow`, `command`, and `route_hint`.
- Workflow step metadata must not include raw goals, prompts, provider payloads, tool args/results, secrets, attachment contents, or local file contents.

### 4. Validation & Error Matrix

- `/research <goal>` -> research workflow invocation.
- `/swarm <goal>` -> swarm workflow invocation.
- `/research` or `/swarm` with blank goal -> HTTP 400 before run-store start.
- Ordinary prompt -> no workflow invocation.
- Unknown slash command -> no workflow invocation, preserving existing compatibility.
- Workflow execution failure -> run status/error remains owned by `RunStore.Complete`; workflow execution steps become `failed`.

### 5. Good/Base/Bad Cases

- Good: `/research compare two local implementations` returns a normal `/message` response, records sanitized workflow steps, and emits aggregate-safe `workflow_step` trace events.
- Base: `hello` behaves exactly like an ordinary direct `/message` request.
- Bad: adding `POST /research`, bypassing approval/session/run store, or persisting the raw workflow goal in step metadata.

### 6. Tests Required

- Parser unit tests for recognized commands, trimming, case-insensitive command name, unknown slash commands, ordinary prompts, and blank goals.
- Handler tests for JSON `/research`, JSON `/swarm`, SSE workflow command, blank command HTTP 400 without run record, and ordinary prompt compatibility.
- Run detail/trace/metrics assertions proving workflow steps are visible and sanitized.
- Existing `/message`, SSE, run-store, routing/fallback, metrics, OpenAI gateway, and control tests must continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
if strings.HasPrefix(req.Text, "/research") {
    resp, err := s.deps.LLMClient.Chat(ctx, "", messages, nil, 0, nil)
}
```

This bypasses daemon approval, sessions, run records, SSE handlers, pause/cancel, routing/fallback, and structured observability.

#### Correct

```go
workflow, err := parseWorkflowInvocation(req.Text)
if err != nil {
    writeError(w, http.StatusBadRequest, err.Error())
    return
}
result, err := s.runWorkflowAgent(ctx, req, workflow, handler)
```

This keeps workflow commands inside the normal daemon execution boundary.

## Scenario: Local Daemon Mailbox Queue

### 1. Scope / Trigger

- Trigger: adding or changing local daemon queue/mailbox APIs or worker claim/ack lifecycle.
- Scope: local in-memory daemon queue runtime. This is not external channel transport, cloud sync, Shannon Cloud replay, or persistent SQLite mailbox storage unless a task explicitly adds persistence.

### 2. Signatures

- Store constructor: `NewMailboxStore(capacity int) *MailboxStore`.
- Store APIs:
  - `(*MailboxStore).Enqueue(QueuedMessage) (QueuedMessage, error)`
  - `(*MailboxStore).List(routeKey string) []QueuedMessage`
  - `(*MailboxStore).Get(id string) (QueuedMessage, bool)`
  - `(*MailboxStore).Claim(routeKey string, limit int) ([]QueuedMessage, error)`
  - `(*MailboxStore).Ack(id, claimID string) bool`
  - `(*MailboxStore).Release(id, claimID string) bool`
- HTTP routes:
  - `POST /queue`
  - `GET /queue?route_key=<route>`
  - `GET /queue/{id}`
  - `POST /queue/claim`
  - `POST /queue/{id}/ack`
  - `POST /queue/{id}/release`

### 3. Contracts

- `POST /queue` requires either `route_key` or `session_id`, plus non-empty `text`.
- If `route_key` is omitted and `session_id` is present, the store may derive `route_key` as `session:<session_id>`.
- Queue capacity is per route and counts non-acknowledged messages.
- Ordering is priority ascending, then enqueue time ascending, then stable id order.
- Dedup is scoped to `route_key + source + external_id`; duplicate enqueue returns the existing message and must not add a second queue item.
- Claimed messages are not claimable again until released. Acknowledged messages are not claimable again.
- Release clears the claim and increments attempt so a future worker can see retry count.
- Public queue views must use sanitized metadata and must not expose provider payloads, tool args/results, secrets, hidden channel credentials, or attachment contents.
- Existing `/inbox` approval behavior must remain compatible when queue APIs are added.

### 4. Validation & Error Matrix

- Missing route/session -> HTTP 400.
- Empty text -> HTTP 400.
- Text over the configured queue text cap -> HTTP 413.
- Negative or zero priority -> HTTP 400.
- Route capacity exceeded -> HTTP 503.
- Unknown queue id -> HTTP 404.
- Ack/release with wrong or missing claim id -> HTTP 409 or HTTP 400.
- Duplicate source/external id on same route -> HTTP 200 with `duplicate=true`.
- Successful new enqueue -> HTTP 202.

### 5. Good/Base/Bad Cases

- Good: a local webhook bridge enqueues a message with `route_key`, `source`, `external_id`, and priority; a worker claims it, processes it, and acks it after successful persistence.
- Base: `GET /queue` on an empty queue returns an empty message list.
- Bad: adding external transport credentials to queue metadata, acknowledging before the worker has safely persisted work, or letting a claimed message be claimed by two workers at the same time.

### 6. Tests Required

- Store tests for priority/FIFO ordering, capacity, per-route dedup, defensive snapshots, claim/ack/release, and wrong-claim rejection.
- HTTP tests for route registration, enqueue success, validation errors, capacity, duplicate enqueue, list/detail, claim, ack, and release.
- Regression tests proving existing `/inbox`, `/message`, SSE, run-store, workflow command, and metrics behavior remains compatible.

### 7. Wrong vs Correct

#### Wrong

```go
msg.Metadata["token"] = providerToken
store.Ack(msg.ID, msg.ClaimID)
err := session.Save()
```

This leaks credentials into public queue state and can lose a message if persistence fails after ack.

#### Correct

```go
claimed, err := store.Claim(routeKey, 1)
// Worker processes and persists the result first.
if saveErr == nil {
    store.Ack(claimed[0].ID, claimed[0].ClaimID)
} else {
    store.Release(claimed[0].ID, claimed[0].ClaimID)
}
```

This preserves queue ownership and leaves failed work retryable.

## Scenario: Episodic Memory Sidecar Preflight

### 1. Scope / Trigger

- Trigger: adding or changing memory provider status, memory recall APIs, or model-input memory preflight injection.
- Scope: local-first memory foundation. This is not cloud session sync, external telemetry, Kocoro/Shannon Cloud bundle pull, tenant account management, or real sidecar process supervision unless a task explicitly adds those pieces.

### 2. Signatures

- HTTP routes:
  - `GET /memory/status`
  - `POST /memory/recall`
- Agent loop interface:
  - `agent.MemoryPreflightProvider`
  - `(*AgentLoop).SetMemoryPreflightProvider(provider)`
- Structured event type:
  - `memory_preflight`

### 3. Contracts

- Memory status must report provider readiness and reason codes without exposing recalled content, query text, secrets, or provider payloads.
- Local bundle discovery is limited to safe paths under the StarClaw memory directory.
- Memory recall may return recalled content to the direct API caller, but memory preflight telemetry must remain content-free.
- Preflight private memory may be appended to the model-facing user message inside `<private_memory>...</private_memory>`.
- `<private_memory>` content must not be persisted to session messages, run request/prompt fields, structured events, metrics, trace export, or compaction summaries.
- Preflight is optional. If the provider is unavailable or returns no data, the run must continue without private memory injection.
- Existing `GET /memory`, `POST /memory`, `DELETE /memory/{name}`, `memory`, and `memory_append` behavior must remain compatible.

### 4. Validation & Error Matrix

- No memory dir or no local memory -> status provider `disabled`, reason `no_local_memory`.
- Valid MEMORY.md facts -> status provider `local`, ready `true`, fallback available.
- Valid current bundle manifest -> status provider `local`, ready `true`, bundle metadata present.
- Malformed or unsafe bundle -> status provider `degraded`, reason `bundle_invalid`.
- Empty recall query -> outcome `no_data`, reason `query_empty`.
- Recall with no facts -> outcome `no_data`, reason `no_local_memory`.
- Provider/preflight error -> run continues and emits content-free error telemetry.

### 5. Good/Base/Bad Cases

- Good: a run asks about a remembered local preference; preflight injects a private block into the model input, records `results_count=1`, and saves only the original user message to session storage.
- Base: no memory is configured; `/memory/status` reports disabled and ordinary runs are unchanged.
- Bad: structured events include the recalled fact text, session JSON contains `<private_memory>`, or bundle path resolution follows a symlink outside the memory directory.

### 6. Tests Required

- Status tests for empty memory, MEMORY.md facts, valid bundle, and malformed bundle.
- Recall tests for matching facts and no-data reason codes.
- Agent-loop tests proving private memory reaches model input and is stripped from saved session messages.
- Run-store tests proving `memory_preflight` events contain counts/status only and no query or recalled content.
- Regression tests proving existing memory APIs/tools and daemon tests continue to pass.

### 7. Wrong vs Correct

#### Wrong

```go
messages = append(messages, client.Message{Role: "user", Content: query + privateMemory})
session.Messages = messages
store.AddEvent(runID, "memory_preflight", map[string]any{"memory": privateMemory})
```

This persists private recall content and leaks it into structured observability.

#### Correct

```go
modelQuery := query + "\n\n" + privateMemory
messages = append(messages, client.Message{Role: "user", Content: modelQuery})
session.Messages = stripPrivateMemoryFromMessages(messages)
handler.OnMemoryPreflight(agent.MemoryPreflightResult{ResultsCount: count})
```

This sends recall context to the model while keeping persistence and telemetry content-free.

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

## Scenario: Desktop RPC Boundary

### Scope

- Trigger: adding or changing native Desktop integration between the StarClaw daemon and a future local Astria/Desktop app.
- Scope: local daemon only. The first boundary lives under `internal/daemon/desktop_rpc` and must not assume a cloud transport, calendar/EventKit implementation, or native app launcher.

### Contracts

- Transport uses Unix domain sockets with 4-byte big-endian length-prefixed JSON frames.
- Supported system methods start with `system.ping` and `system.capabilities`.
- Codec must reject zero-length, malformed, partial, and oversized frames.
- Broker pending requests must terminate on success, timeout, context cancellation, send failure, or disconnect. Pending requests must not hang after a Desktop disconnect.
- Listener may handle Desktop-originated `system.*` requests locally and route daemon-originated request results back through the broker.
- `/status` may expose only `desktop_rpc.listening`, `desktop_rpc.connected`, and `desktop_rpc.pending`. Do not expose socket paths, raw frame payloads, request params, API keys, provider headers, or user content.

### Tests Required

- Codec round-trip and invalid frame tests.
- Broker not-connected, request/result correlation, timeout, context cancel, and disconnect cancellation tests.
- Listener fake Desktop smoke tests for Desktop-originated `system.ping` / `system.capabilities`.
- Listener fake Desktop smoke test for daemon-originated request/result flow.
- Daemon `/status` test proving Desktop RPC exposes only state booleans/counts.

---

## Anti-Patterns

- ❌ Don't add comments that explain WHAT — code should be self-documenting
- ❌ Don't create premature abstractions — three similar lines > one premature helper
- ❌ Don't add error handling for impossible scenarios
- ❌ Don't ship commented-out code — git history preserves it
- ❌ Don't create `utils.go` / `helpers.go` files
- ❌ Don't add features "for future use" (YAGNI)
