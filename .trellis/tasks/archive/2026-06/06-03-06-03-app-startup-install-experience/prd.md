# Improve app startup and install experience

## Goal

Make StarClaw easier to launch and diagnose from the local GUI entry point, especially for users who expect `starclaw app` to be the one command that starts the daemon and opens the Web UI.

## Requirements

- `starclaw app` should provide a reliable one-command GUI startup path for local use.
- If the daemon is already running, the command should reuse it and open/print the existing Web UI URL instead of failing noisily.
- If startup fails, the CLI should return actionable messaging that distinguishes port conflicts, daemon startup failure, and browser-open failure.
- Version/build surfaces should keep showing launch and update command hints.
- Existing daemon routes, Web UI smoke tests, and normal daemon commands should remain compatible.

## Acceptance Criteria

- [x] `starclaw app` behavior is audited and improved where needed.
- [x] Existing-running-daemon path is handled cleanly.
- [x] Startup failures include actionable messages.
- [x] Tests or smoke coverage validate the user-facing app launch behavior.
- [x] Existing Web UI smoke remains green.
