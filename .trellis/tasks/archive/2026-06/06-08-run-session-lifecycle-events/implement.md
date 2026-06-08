# Run session lifecycle events implementation plan

## Checklist

1. Add optional EventBus wiring to `RunStore`.
   - Add a bus field and `SetEventBus`.
   - Keep `NewRunStore` and `NewPersistentRunStore` signatures compatible.
   - Wire `Server` construction so daemon-owned run stores publish to the
     server EventBus.
2. Add lifecycle payload helpers.
   - Build payloads from `RunRecord` plus terminal response data.
   - Use safe fields only and reuse redaction helpers before JSON encoding.
   - Keep nil/empty optional fields omitted where possible.
3. Publish lifecycle events.
   - Publish `run_started` after start is recorded.
   - Publish `run_completed` after successful completion is recorded.
   - Publish `run_error` for both Go errors and `RunAgentResponse.Error`.
   - Do not publish terminal lifecycle for missing run IDs or cancelled runs
     that intentionally bypass terminal completion.
4. Add tests.
   - Unit tests for `RunStore` live lifecycle bus publishing.
   - Replay/SSE test proving missed lifecycle events are returned by `/events`.
   - Redaction test for lifecycle EventBus payloads.
   - Regression test that nil bus paths still do not panic.
5. Validate.
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-run-session-lifecycle-events`
   - `go test ./internal/daemon -run 'TestRunStore|TestHandleEvents|TestOpenAI|TestWorkflow|TestRunControl|TestRunLifecycle|TestEventBus' -count=1 -timeout=90s`
   - `go test ./internal/daemon -count=1 -timeout=90s`
   - `go test ./...`

## Risk Points

- `RunStore` persistence paths must not accidentally publish stale recovered
  events during load.
- EventBus payloads must not include `RunAgentRequest.Text`, response content,
  tool args, or raw errors that include secrets.
- `/events` replay ordering must stay deterministic: replayed events first,
  then live events from the same subscription.

## Follow-Up Boundary

If implementation reveals Web UI state gaps, record them for the Phase12
`webui-live-recovery` child instead of broadening this task into UI recovery.
