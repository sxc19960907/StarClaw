# Event contract documentation implementation plan

## Checklist

1. Add `docs/DAEMON_EVENTS.md`.
   - Document `/events` endpoint.
   - Document `/message` streaming SSE endpoint.
   - List canonical names, compatibility aliases, and payload rules.
   - Record local-first divergence from Kocoro/Shannon Cloud.
2. Link the doc from README Local Runtime API.
3. Add static documentation test.
   - Verify README link.
   - Verify key event names and replay terms exist in the doc.
4. Validate.
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-event-contract-documentation`
   - `go test ./internal/daemon -run 'TestDaemonEventDocumentation' -count=1 -timeout=90s`
   - `go test ./internal/daemon -count=1 -timeout=90s`
   - `go test ./...`

## Risk Points

- Keep the document factual and aligned with current code.
- Do not imply remote/cloud behavior that StarClaw intentionally does not
  enable.
- Avoid documenting raw payload content as shareable; prompts/results remain
  local operator review surfaces.
