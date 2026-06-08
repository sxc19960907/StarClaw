# Safe replay execution boundary design

## Scope

This slice turns replay from plan-only into an explicitly approved launch path. It does not bypass tool permissions, does not replay recorded tool outputs, and does not make replay deterministic. It re-runs the source `RunAgentRequest` through the normal daemon execution path only after `approved=true`.

## API Behavior

Route remains `POST /runs/{id}/control`.

- `{"action":"replay"}` or `{"action":"replay","approved":false}`
  - Return HTTP 200 with `status="approval_required"`.
  - Include a redacted replay plan.
  - Record a `RunControlDecision{Action:"replay", Status:"approval_required"}` on the source run.
  - Do not create or execute a new run.
- `{"action":"replay","approved":true}`
  - Create a new replay request id, e.g. `replay-<source>-<suffix>`.
  - Record `RunControlDecision{Action:"replay", Status:"approved"}` on the source run.
  - Create/execute a new run through `s.runAgent`.
  - Return HTTP 200 with `status`, `source_run_id`, `replay_run_id`, `replay`, and `run`.

## Execution

Use the same flow as `/message`:

1. Clone source `RunAgentRequest`.
2. Assign a new request id.
3. Set source/channel metadata to indicate replay.
4. Start a new run record.
5. Add durable workflow steps to the source and replay records:
   - source: replay approval requested/approved
   - replay: launched/running/completed or failed
6. Run `s.runAgent` with `s.recordingHandler(replayRunID, ...)`.
7. Complete the replay run normally.

Because `s.runAgent` constructs `NewDaemonApprovalRequester`, all tool/external approval gates remain intact.

## Redaction

- Planning response uses `replayControlRequest`; it never returns source prompt text.
- Control metadata and step metadata must not store the original prompt.
- Metrics remain aggregate-only.
- Run detail may contain the replay run prompt because run detail already stores run requests; this child only prevents prompt leakage through replay control responses, structured events, and metrics.

## Compatibility

- Existing `cancel`, `pause`, `resume`, and unapproved replay behavior remain compatible.
- Existing `/runs`, `/runs/{id}`, `/metrics`, `/cancel`, and approval routes remain unchanged.
- Replay launch is synchronous in this slice, matching current `/message` non-SSE behavior.

## Rollback

Remove approved replay branch and helper functions; unapproved replay plan behavior remains.
