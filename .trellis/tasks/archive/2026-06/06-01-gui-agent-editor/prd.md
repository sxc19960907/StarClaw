# Add GUI agent editor

## Goal

Let users create and edit named agents from the daemon Web UI instead of hand-editing files under `~/.starclaw/agents`.

## Requirements

- Implement daemon agent create/update APIs currently returning 501.
- Support editing core fields:
  - name for create
  - prompt stored in `AGENT.md`
  - memory stored in `MEMORY.md` when non-empty
  - model and reasoning effort in `config.yaml`
  - tool allow/deny lists in `config.yaml`
  - auto approve flag in `config.yaml`
- Validate agent names with existing `agents.ValidateAgentName`.
- Preserve existing list/get/delete behavior.
- Add Web UI controls in the Agents panel for create/edit/save/delete.
- Keep embedded static GUI; no frontend build step.
- Update smoke coverage for creating/editing/deleting an agent.

## Acceptance Criteria

- [ ] `POST /agents` creates an agent directory with `AGENT.md` and optional config files.
- [ ] `PUT /agents/{name}` updates prompt/config for an existing agent.
- [ ] Backend tests cover create, update, validation, and reload via GET/list.
- [ ] GUI Agents panel can create and edit an agent.
- [ ] Smoke script verifies create/edit/delete agent flow.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Out Of Scope

- Full command file editor under `commands/`.
- Agent heartbeat UI.
- Bulk import/export.
