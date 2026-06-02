# Add GUI agent test run entry

## Goal

Let users jump from an agent's editor directly into a one-off chat test with that agent selected.

## Requirements

- Add a test-run action to the Agents editor.
- The action should switch to the Chat panel with the current agent selected.
- The action should prefill a concise test prompt.
- The action should enable `New session` for the test run.
- Do not auto-send the prompt; users should review before sending.
- Preserve the existing chat `/message` flow and agent save behavior.
- Keep the embedded static GUI with no frontend build step.

## Acceptance Criteria

- [ ] Editing an existing agent exposes a test-run action.
- [ ] Clicking the action switches to Chat.
- [ ] The Chat agent selector is set to the edited agent.
- [ ] The Chat input is prefilled with a test prompt.
- [ ] The New session checkbox is checked.
- [ ] Smoke verifies the interaction.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, `go vet ./...`, and `scripts/smoke_webui.sh` pass.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
