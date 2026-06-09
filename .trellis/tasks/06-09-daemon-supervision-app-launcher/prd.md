# Daemon supervision app launcher

## Goal

Implement the app-side daemon launcher/supervisor contract for the standalone
Astria shell, closing the Kocoro parity gap where Desktop owns startup,
attach/reuse, health monitoring, and user-visible daemon failure states.

## Requirements

- Start or attach to the local StarClaw daemon from the desktop shell.
- Preserve daemon-only CLI operation and browser launch fallback.
- Detect and surface daemon states: not running, starting, healthy, unhealthy,
  version mismatch, port conflict, startup timeout, and process crash.
- Reuse existing local endpoints and diagnostics where possible:
  `/health`, `/status`, `/diagnostics`, `/app/`, and existing `starclaw app`
  readiness logic.
- Define how this child uses or stages Desktop RPC, pidfile, and
  `system.capabilities` relative to Kocoro.
- Avoid silent infinite restart loops; failures must produce actionable user
  states and useful logs.

## Acceptance Criteria

- [ ] The standalone shell can launch Astria from a clean no-daemon state.
- [ ] The shell can attach to an already-running compatible daemon.
- [ ] Startup timeout, port conflict, and unhealthy daemon paths are testable.
- [ ] Existing `starclaw app --check` and `starclaw doctor` remain consistent
      with app launcher readiness.
- [ ] Unit or smoke tests cover supervision logic that can be exercised without
      a signed native app bundle.
- [ ] Any intentional divergence from Kocoro's pidfile/socket reconciliation is
      documented with a follow-up path.

## Notes

Depends on `standalone-desktop-shell-plan`.
