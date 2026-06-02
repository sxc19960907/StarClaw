# Design

This is a validation-only task. Use existing smoke automation as the test harness:

- `scripts/smoke_webui_agents.sh` for agent editor/import/export/test runner.
- `scripts/smoke_webui_runs.sh` for run detail and session workflow.
- `scripts/smoke_webui.sh` for combined GUI readiness.

No code changes are planned unless a test failure exposes a product or smoke blocker.
