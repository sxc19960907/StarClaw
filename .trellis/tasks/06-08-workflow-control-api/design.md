# Workflow control API design

## Scope

This slice adds a backend workflow-control contract around existing daemon runs. The first implementation must be compatible with the current Web UI stop button and the existing `POST /cancel` client method.

The MVP includes:

- Keep `POST /cancel` working for current UI/client callers.
- Add route-level run control endpoints for state inspection and cancel.
- Define pause/resume/replay semantics without pretending unsupported runtime behavior exists.
- Record control decisions in run metadata and structured events.

Out of scope for this slice:

- Durable process orchestration.
- True pause/resume of an in-flight Go context.
- Automatic replay that repeats tool calls or external side effects.
- New Web UI controls beyond preserving current stop behavior.

## API contracts

### Existing compatibility route

`POST /cancel`

Request:

```json
{"request_id": "run-id"}
```

Response:

```json
{"status": "cancelled", "run_id": "run-id", "action": "cancel"}
```

The route remains accepted because existing Web UI and client code call it.

### Run control route

`POST /runs/{id}/control`

Request:

```json
{
  "action": "cancel",
  "reason": "operator stop",
  "approved": false
}
```

Supported actions:

- `cancel`: cancel an active run context.
- `pause`: staged; returns `409` because the runtime has no pause primitive yet.
- `resume`: staged; returns `409` because the runtime has no pause primitive yet.
- `replay`: returns a replay plan that requires explicit approval before any new run is launched.

Cancel success response:

```json
{
  "status": "cancelled",
  "run_id": "run-id",
  "action": "cancel"
}
```

Unsupported staged response:

```json
{
  "error": "pause is not supported for this runtime"
}
```

Replay plan response:

```json
{
  "status": "approval_required",
  "run_id": "run-id",
  "action": "replay",
  "replay": {
    "source_run_id": "run-id",
    "requires_approval": true,
    "reason": "Replay can repeat tool calls or external effects.",
    "request": {
      "text": "...",
      "agent": "agent-name",
      "channel": "http",
      "session_id": "..."
    }
  }
}
```

The replay response is intentionally a plan, not execution. A later task can add a separate approved launch endpoint.

## Data flow

1. Message/OpenAI execution stores active run cancel functions in `Server.running`.
2. Control handlers validate action and run id.
3. Cancel looks up the active cancel function, calls it, and records a run control decision.
4. Replay looks up the stored run, builds a redacted/restricted plan, records the replay decision, and returns `approval_required`.
5. Run detail continues to expose existing run fields plus new control metadata.

## Run metadata

Add `RunControlDecision` to run records:

```go
type RunControlDecision struct {
    Action string    `json:"action"`
    Status string    `json:"status"`
    Reason string    `json:"reason,omitempty"`
    At     time.Time `json:"at"`
}
```

`RunRecord.Control` is serialized as `control,omitempty` and appended each time a control action is accepted or staged.

Structured event type: `control_decision`.

Event data must include only action, status, and reason. Do not include prompt bodies or raw replay request bodies in structured events.

## Compatibility

- Existing `/cancel` route remains valid.
- Existing Web UI stop behavior remains valid because it still posts `request_id` to `/cancel` and aborts the browser stream.
- Existing `/runs` summary shape remains unchanged.
- `/runs/{id}` can gain `control` metadata because run details already include extra optional fields.

## Permission and replay boundary

Cancel is local daemon control and does not launch tools. Replay is potentially dangerous because it can repeat tool calls, file writes, or external delivery. Therefore this slice must return a plan requiring explicit approval and must not execute replay automatically.

## Rollback

The change is isolated to daemon control handlers, run-store metadata, router registration, and tests. Rollback removes the new route and control metadata while leaving existing `/cancel` behavior intact.
