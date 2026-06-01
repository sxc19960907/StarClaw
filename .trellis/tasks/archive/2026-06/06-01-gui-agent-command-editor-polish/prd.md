# Polish GUI agent command editor

## Goal

Make the Agent command editor less error-prone for normal create/edit/delete workflows.

## Requirements

- Add a clear/cancel command editing action so users can leave edit mode without saving.
- Allow renaming a staged command by editing the command name and saving.
- Keep command name validation consistent with the backend contract.
- Preserve the existing command persistence API and file layout.
- Keep the embedded static GUI with no frontend build step.
- Extend smoke coverage for rename and cancel/clear behavior.

## Acceptance Criteria

- [ ] Selecting a command loads name/body into editable controls.
- [ ] Users can rename a staged command and the old name is removed from the payload.
- [ ] Users can clear/cancel the command editor and then add another command.
- [ ] Invalid command names are rejected before staging.
- [ ] Smoke verifies rename + clear behavior and final reload state.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
