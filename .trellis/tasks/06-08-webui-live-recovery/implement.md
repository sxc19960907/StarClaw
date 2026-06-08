# WebUI live recovery implementation plan

## Checklist

1. Extend Web UI event stream state.
   - Add `lastRecoveredAt` and `refreshingRuns`.
   - Keep existing reconnect status behavior.
2. Add safe lifecycle event mapping.
   - Parse `run_id` / `id`, status, agent, channel, source, session_id,
     timestamps, usage.
   - Omit unsafe keys: prompt, text, content, delta, request, response, args.
   - Mark mapped runs as recovered from `event_stream`.
3. Add lifecycle EventSource handlers.
   - `run_started`
   - `run_completed`
   - `run_error`
   - Reuse the same render refresh function as `loadRuns()`.
4. Add guarded recovery refresh.
   - On recovered EventSource open, call a guarded `refreshRunsAfterEventStreamRecovery()`.
   - Avoid overlapping `loadRuns()` calls.
5. Add static tests.
   - Check listeners and helper names exist.
   - Check unsafe payload keys are explicitly omitted.
   - Check recovered reconnect schedules `/runs` refresh.
6. Validate.
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-webui-live-recovery`
   - `go test ./internal/daemon -run 'TestWebUI' -count=1 -timeout=90s`
   - `go test ./internal/daemon -count=1 -timeout=90s`
   - `go test ./...`

## Risk Points

- `loadRuns()` renders many dependent panels; lifecycle upserts should reuse a
  small helper instead of duplicating render calls.
- Reconnect can fire multiple times. Guard refresh calls so the browser does
  not flood `/runs`.
- Do not use lifecycle payloads as prompt/result sources.
