# End-to-end test general agent workflow

## Goal

Validate StarClaw as a general agent workflow from the GUI using the existing smoke automation and report whether the main flow is usable.

## Test Scope

- Agent creation.
- Agent allow/deny/auto-approve preview.
- Agent command creation, rename, delete, and persistence.
- Agent config export/import.
- Agent test run and result summary.
- Run history/detail.
- Session open/copy/rename/favorite/delete controls.
- Full Web UI smoke route checks and screenshots.

## Acceptance Criteria

- [x] Full Web UI smoke passes.
- [x] Agents smoke passes.
- [x] Runs smoke passes.
- [x] Smoke artifacts are available under `output/playwright/`.
- [x] Test conclusion is recorded in session journal.
