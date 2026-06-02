# Run GUI full smoke and fix readiness issues

## Goal

Run the full Web UI smoke suite and fix any regressions or stability issues that block GUI release readiness.

## Requirements

- Run `scripts/smoke_webui.sh`.
- If it fails, fix only issues exposed by the full smoke run.
- Keep smoke assertions scoped to avoid Playwright strict-mode ambiguity.
- Do not expand feature scope beyond release-readiness fixes.
- Preserve smoke artifacts for inspection.

## Acceptance Criteria

- [x] Full Web UI smoke passes.
- [x] JS syntax check passes if JS changes are made.
- [x] Diff check passes.
- [x] Any code changes are committed and task is archived.
