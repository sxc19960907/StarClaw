# Add GUI agent command editor

## Goal

Let users manage an agent's custom slash command files from the daemon Web UI instead of editing `commands/*.md` manually.

## Requirements

- Extend the agent create/update API with an optional `commands` object.
- Persist command entries to `<agent>/commands/<name>.md`.
- Load existing commands into the Agents editor.
- Support creating, editing, and deleting command entries in the Web UI.
- Validate command names so GUI/API writes cannot escape the `commands/` directory.
- Preserve existing prompt, memory, model, permission, heartbeat, create/update/delete behavior.
- Keep the embedded static GUI with no frontend build step.
- Add backend and smoke coverage for command round trip.

## Acceptance Criteria

- [ ] `POST /agents` can create an agent with commands.
- [ ] `PUT /agents/{name}` can update command content and remove deleted commands.
- [ ] API rejects invalid command names.
- [ ] Web UI can add/edit/delete a command and reload it from the agent detail API.
- [ ] Existing update clients that omit `commands` do not delete existing commands.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
