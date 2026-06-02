# Improve GUI version release readiness

## Goal

Make the Version page show a clearer release/readiness summary so users can distinguish development builds, release builds, update support, and GUI launch entrypoints.

## Requirements

- Version page should render a concise readiness card above raw metadata rows.
- Readiness card should show:
  - build status;
  - update support;
  - launch command;
  - Web UI URL.
- Development builds should clearly indicate update checks require a release build.
- Existing update check behavior must not change.
- Core Web UI smoke should assert the readiness card.

## Acceptance Criteria

- [x] Version page shows a release readiness card.
- [x] Development builds show update checks as requiring a release build.
- [x] Launch command and Web UI URL remain visible.
- [x] Existing update check flow still works.
- [x] Core Web UI smoke and JS syntax check pass.
