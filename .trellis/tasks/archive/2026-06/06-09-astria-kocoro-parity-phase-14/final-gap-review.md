# Phase14 final gap review

## Scope closed

Phase14 moved Astria from Phase13's HTTP-health-only desktop shell toward a
semi-bound local Desktop RPC lifecycle:

- `starclaw daemon start --rpc-socket <path> --rpc-pidfile <path>` now exposes
  an explicit paired Desktop RPC launch contract while preserving HTTP-only
  daemon mode when neither flag is provided.
- The daemon starts the Desktop RPC listener before writing the pidfile and
  reports redacted Desktop RPC state through `/status`.
- The macOS Astria shell resolves deterministic runtime artifacts, launches the
  daemon with explicit socket/pidfile paths, connects over Unix socket, calls
  `system.capabilities`, and validates protocol/method compatibility before
  treating Desktop RPC as attached.
- Astria now handles stale socket/pidfile artifacts inside its own runtime
  directory, preserves live pidfile ownership, refuses unsafe cleanup paths, and
  keeps HTTP WebView fallback available in degraded Desktop RPC states.

## Evidence

- `273237e feat: add daemon desktop rpc launch contract`
- `8907e2f feat: reconcile Astria desktop rpc capabilities`
- `c58305a feat: harden Astria desktop rpc fallback recovery`
- Validation performed during the final child:
  - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-fallback-recovery`
  - `scripts/smoke_macos_astria_shell.sh`
  - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
  - `go test ./...`
  - `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 85-90% aligned with Kocoro for local-first desktop
lifecycle behavior. The remaining gap is no longer basic standalone launch or
daemon reconciliation; it is the depth and polish of native desktop integration
around the Desktop RPC session.

## Remaining Kocoro gaps

- Long-lived native Desktop RPC client session and event monitoring after the
  initial `system.capabilities` handshake.
- Native OS integration depth: menu/Dock behavior, multi-window management,
  TCC-style permission helpers, notifications, file/clipboard affordances, and
  first-class native diagnostics.
- Production desktop hardening: crash reporter, richer reconnect UX, daemon
  lifecycle observability, and operator-friendly failure reports.
- Signed, notarized, distributable app pipeline with update channel mechanics.

## Recommended next phase

Phase15 should focus on long-lived Desktop RPC session management and native
event monitoring. That closes the most direct gap left by Phase14: Astria can
currently prove that the daemon and app are compatible at launch, but it does
not yet maintain a richer native session boundary for disconnects, event
streams, native tool availability, and ongoing health transitions.
