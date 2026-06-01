# Add GUI agent permission rule editor

## Goal

Make agent permission configuration easier to edit safely from the daemon Web UI, without requiring users to hand-edit YAML.

## Requirements

- Extend the existing Agents panel editor with permission rule controls for:
  - tool allow list
  - tool deny list
  - auto approve
- Preserve the existing agent create/update API contract and file layout.
- Normalize comma/newline separated entries into clean string arrays before save.
- Show permission rule state clearly when editing an existing agent.
- Keep the embedded static GUI with no frontend build step.
- Add smoke coverage for saving and reloading permission rules.

## Acceptance Criteria

- [ ] Existing agents load their permission config into editable controls.
- [ ] Saving an agent persists tool allow/deny lists and auto approve in `config.yaml`.
- [ ] Reloading the agent shows the saved permission rules.
- [ ] Empty permission lists are persisted as absent/empty according to existing backend behavior.
- [ ] Smoke script verifies create/edit/reload/delete for permission rules.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Out Of Scope

- Full permissions policy editor for global config.
- Per-tool argument policy editing.
- Command file editor under `commands/`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
