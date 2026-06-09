# Phase15 final gap review

## Scope closed

Phase15 turned the Phase14 launch-time Desktop RPC handshake into a more
durable local desktop boundary:

- Astria now keeps a long-lived Desktop RPC session monitor after initial
  `system.capabilities` validation.
- Recoverable Desktop RPC disconnects enter bounded reconnect attempts and then
  degraded HTTP fallback instead of crashing or spinning.
- Protocol/version mismatches are classified as compatibility failures and do
  not retry against the same daemon.
- The daemon records incoming local `desktop_event` frames through the existing
  `EventSink` path with bounded in-memory retention.
- `/status.desktop_rpc.events` exposes only redacted metadata: retained count,
  last event type, and last timestamp.
- Astria surfaces reconnecting, degraded, and mismatch Desktop RPC session
  states through native banners while keeping the WebView mounted when HTTP is
  healthy.

## Evidence

- `d9b20e2 feat: add Astria desktop rpc session lifecycle`
- `05216a4 feat: monitor desktop rpc events`
- `53548bc feat: surface Astria desktop rpc diagnostics`
- Validation performed during Phase15 children:
  - `python3 ./.trellis/scripts/task.py validate ...`
  - `scripts/smoke_macos_astria_shell.sh`
  - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./...`
  - `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 88-92% aligned with Kocoro for local-first desktop
lifecycle behavior. The remaining gap has moved from daemon/RPC lifecycle to
broader native desktop product depth and release/distribution maturity.

## Remaining Kocoro gaps

- Full native OS integration depth: menus, Dock actions, multi-window behavior,
  notifications, file/clipboard affordances, and TCC-style helper flows.
- Production hardening: crash reporter, richer native diagnostics export,
  operator-friendly failure reports, and broader recovery actions.
- Signed/notarized distributable app pipeline with update channel mechanics.
- Native implementations for deeper desktop tools beyond the current RPC
  boundary and event metadata plumbing.

## Recommended next phase

Phase16 should focus on native OS integration and distribution hardening. The
highest-value split is:

1. native menu/Dock/multi-window shell behavior;
2. native diagnostics export and crash/failure reporting;
3. signed/notarized package and updater pipeline boundary.
