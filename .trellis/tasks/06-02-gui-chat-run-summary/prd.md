# Add GUI chat run summary

## Goal

Show a concise run summary in the Chat transcript after a daemon chat run completes.

## Requirements

- Render a summary card after successful chat completion.
- Include session id, selected agent, and token usage when available.
- Include the request id to correlate UI and daemon logs.
- Preserve existing message streaming, tool event, approval, and session behavior.
- Keep the embedded static GUI with no frontend build step.
- Add smoke coverage for the summary rendering path.

## Acceptance Criteria

- [ ] Completed chat runs render a summary card in the transcript.
- [ ] Summary includes session id when the response provides it.
- [ ] Summary includes selected agent or "default" agent.
- [ ] Summary includes usage values when available.
- [ ] Summary includes request id.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
