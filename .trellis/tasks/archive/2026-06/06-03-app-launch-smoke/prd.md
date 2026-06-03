# Add app launch smoke coverage

## Goal

Add a fast CLI-level smoke test for the app launch path so startup regressions are caught before the heavier browser Web UI smoke.

## Requirements

- Build a temporary StarClaw binary in an isolated home directory.
- Verify `starclaw app --check` prints launch readiness without starting the daemon.
- Verify `starclaw app --no-open` starts or reuses the daemon and prints the Web UI URL.
- Verify daemon `/version` and `/diagnostics` expose launch/runtime fields.
- Stop the daemon cleanly after the smoke.
- Add the smoke to CI before the browser Web UI smoke.

## Acceptance Criteria

- [x] New smoke script exists and passes locally.
- [x] Smoke validates `app --check`, `app --no-open`, `/version`, and `/diagnostics`.
- [x] CI runs the new smoke before Web UI browser smoke.
- [x] Diff check passes.
