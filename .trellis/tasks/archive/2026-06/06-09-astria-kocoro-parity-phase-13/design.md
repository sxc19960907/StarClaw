# Astria Kocoro parity phase 13 design

## Architecture Boundary

Phase13 should treat the standalone desktop app as a thin native product shell
around the existing StarClaw daemon and Astria Web UI:

- CLI remains the cross-platform control surface.
- The daemon remains the local runtime owner for sessions, runs, tools,
  approvals, metrics, traces, event replay, and workflow control.
- Astria Web UI remains the primary in-app experience and continues to be
  served by the daemon at `/app/`.
- The desktop shell owns process supervision, window lifecycle, local app
  affordances, and packaging.

The recommended first shell is a macOS-native app boundary, implemented as a
small app target that hosts the local Web UI in a web view and spawns or
attaches to the existing Go daemon. This most closely matches the Kocoro
Desktop lifecycle model and avoids duplicating the Web UI in Electron/Tauri
before the supervision contract is proven.

## Kocoro Parity Reference

Kocoro's relevant model:

- Desktop starts or attaches to the daemon.
- Desktop and daemon communicate through a Unix socket Desktop RPC protocol.
- Daemon writes an explicit pidfile after the socket is listening.
- Desktop performs reconciliation through `system.capabilities` before
  declaring `desktop_online`.
- Version mismatch and broken socket state are fail-fast and user-visible.
- The daemon can survive Desktop UI exit in a semi-bound lifecycle model.

StarClaw currently has:

- `starclaw app` browser launch with daemon health polling.
- `internal/daemon/desktop_rpc` listener, broker, frame codec, protocol
  constants, and tests.
- Local runtime diagnostics through `/health`, `/status`, `/diagnostics`, and
  the Web UI Version page.
- Release docs and smoke scripts for CLI/npm/binary delivery.

The main missing layers are the native shell process, daemon supervision
contract, window recovery UX, and desktop packaging boundary.

## Data and Control Flow

1. User launches the standalone app.
2. App resolves bundled `starclaw` binary and local app data paths.
3. App checks for an existing daemon through pidfile/socket/health signals.
4. App either attaches to a compatible daemon or starts a new child daemon.
5. App waits for health/readiness and opens `/app/` in the shell window.
6. App monitors daemon health and Web UI load/reconnect status.
7. On daemon crash, stale pidfile, version mismatch, or socket failure, app
   surfaces an actionable state and follows the child task's restart policy.

## Compatibility

- Existing CLI launch commands remain supported.
- Existing browser-based Astria launch remains supported.
- Native shell must not require cloud credentials.
- Native shell must not disable local HTTP APIs used by tests and local
  integrations.
- Any new flags for desktop supervision should be optional and paired with
  tests so daemon-only mode remains valid.

## Trade-offs

- SwiftUI/native shell: best fit for macOS lifecycle, signing, TCC, and Kocoro
  parity. Higher platform-specific implementation cost.
- Tauri: attractive cross-platform packaging, but adds Rust/WebView lifecycle
  and packaging complexity before StarClaw has a proven desktop supervision
  contract.
- Electron: broad ecosystem and Web UI fit, but heavier runtime footprint and
  less aligned with local-first/native macOS parity.
- CLI-only browser launch: already works, but does not close the Kocoro
  standalone app gap.

## Rollout

Phase13 should land in child-sized increments:

1. Document the app shell boundary and create the native project skeleton.
2. Wire daemon supervision and readiness checks.
3. Add window/reconnect recovery UX.
4. Add packaging, signing, updater, and release smoke boundaries.

Each increment should leave `starclaw app` usable as a fallback.
