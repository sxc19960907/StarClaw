# Add GUI agent heartbeat editor

## Goal

Let users configure per-agent heartbeat settings from the daemon Web UI instead of hand-editing `config.yaml`.

## Requirements

- Extend the agent create/update API request with heartbeat fields:
  - heartbeat interval (`every`)
  - active hours (`active_hours`)
  - heartbeat model override (`model`)
- Persist heartbeat settings under the existing agent `config.yaml` `heartbeat` block.
- Treat an empty heartbeat interval as disabled and omit the heartbeat config.
- Extend the Agents editor with heartbeat controls.
- Load existing heartbeat config into the editor.
- Preserve existing agent prompt, memory, model, permission, create/update/delete behavior.
- Keep the embedded static GUI with no frontend build step.
- Add backend and smoke coverage for heartbeat round trip.

## Acceptance Criteria

- [ ] `POST /agents` can create an agent with heartbeat config.
- [ ] `PUT /agents/{name}` can update and clear heartbeat config.
- [ ] Web UI can save and reload heartbeat interval, active hours, and model.
- [ ] Empty heartbeat interval removes/omits the heartbeat block.
- [ ] Backend tests cover heartbeat persistence and clearing.
- [ ] Smoke script verifies heartbeat create/edit/reload/delete flow.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
