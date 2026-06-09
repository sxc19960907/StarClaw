# Daemon supervision app launcher design

## Decision

Implement daemon supervision inside the macOS Astria shell as a staged
launcher:

1. Check the fixed local daemon endpoint through HTTP.
2. If the daemon is healthy, attach to it and load `/app/`.
3. If the daemon is not healthy, resolve a `starclaw` binary and spawn
   `starclaw daemon start` as a child process.
4. Poll `/health` until ready or timeout.
5. Surface actionable shell states for launch failure, timeout, port conflict,
   and post-launch crash.

This intentionally uses StarClaw's existing local HTTP readiness contract first
instead of introducing `--rpc-socket` / `--rpc-pidfile` in this child. The
Desktop RPC reconciliation model remains the Kocoro parity target, but current
StarClaw already has reliable `/health`, `/status`, `/diagnostics`, and
`starclaw app --check` surfaces. Using those first closes the user-visible
standalone app gap without destabilizing daemon-only CLI operation.

## Scope

In scope:

- Swift-side `DaemonSupervisor` state machine.
- Swift-side binary resolution for local development and future bundled builds.
- Shell UI states for starting, attached, timeout, launch failure, and daemon
  crash.
- Build/smoke script updates that exercise the supervisor without signing.
- Documentation updates for local app launch behavior.

Out of scope:

- Production signing/notarization.
- Auto-update.
- Durable pidfile/socket reconciliation.
- Calendar/Desktop RPC feature registration.
- Replacing `starclaw app` or daemon-only CLI paths.

## State Machine

States:

- `checking`: app launched and probes `/health`.
- `starting`: no healthy daemon was found; shell spawned `starclaw daemon start`.
- `attached`: daemon is healthy and Web UI can load.
- `failed`: binary resolution, process spawn, timeout, or port conflict failed.
- `crashed`: a daemon child started by the shell exited after readiness.

Transitions:

1. `checking -> attached`: `/health` returns 200.
2. `checking -> starting`: `/health` is unavailable.
3. `starting -> attached`: daemon becomes healthy before timeout.
4. `starting -> failed`: spawn fails or timeout expires.
5. `attached -> crashed`: child daemon process exits while app is open.

If the shell attaches to an already-running daemon, it does not own that
process and must not stop it on exit.

## Binary Resolution

Resolution order:

1. `ASTRIA_STARCLAW_BIN` environment variable for local development and smoke
   tests.
2. `Contents/Resources/starclaw` inside a future packaged `.app`.
3. `starclaw` found through `/usr/bin/env` as a developer fallback.

The smoke script should use `ASTRIA_STARCLAW_BIN` against a locally built
binary so supervision can be tested without packaging a release asset.

## HTTP Contracts

Use existing endpoints:

- `GET /health`: readiness probe.
- `GET /status`: runtime/version summary for diagnostics.
- `GET /diagnostics`: user-openable diagnostics when launch fails.
- `GET /app/`: Web UI route hosted in `WKWebView`.

Use the same default timeout shape as CLI app launch unless implementation
finds a macOS-specific reason to adjust it:

- health request timeout: short, sub-second probe.
- startup timeout: approximately 5 seconds.
- poll interval: approximately 120 ms.

## Compatibility

- Existing `starclaw app`, `starclaw app --no-open`, and `starclaw app
  --check` behavior must remain unchanged.
- Existing Go daemon start behavior must remain unchanged.
- The app shell remains optional. Local integrations can continue to use the
  daemon HTTP API without the app.
- No cloud auth, cloud lifecycle transport, or telemetry is added.

## Kocoro Divergence

This child deliberately diverges from Kocoro's full pidfile/socket
reconciliation. Kocoro's model remains the target for deeper native parity, but
StarClaw should first land the standalone app's user-visible launch behavior on
top of the existing stable daemon readiness contract.

Follow-up convergence points:

- Add explicit app-spawn daemon flags if the HTTP contract is insufficient.
- Add pidfile or Desktop RPC `system.capabilities` version reconciliation in a
  later child or phase.
- Preserve the already-implemented `internal/daemon/desktop_rpc` protocol for
  native feature expansion.
