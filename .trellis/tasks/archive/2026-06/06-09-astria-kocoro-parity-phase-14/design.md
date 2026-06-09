# Astria Kocoro parity phase 14 design

## Architecture Boundary

Phase14 should use StarClaw's existing Desktop RPC subsystem instead of adding
a new protocol. The macOS Astria shell remains a thin native shell around the
daemon-served Web UI, but readiness should become:

1. daemon HTTP health is reachable;
2. Desktop RPC socket is listening;
3. Astria connects as the desktop client;
4. `system.capabilities` returns the expected protocol and method set;
5. app/daemon version compatibility is accepted.

Only after those checks should Astria report the desktop handshake as ready.
HTTP health remains the fallback for CLI/browser launch and degraded recovery.

## Data and Control Flow

1. Astria resolves local app-data paths for:
   - `daemon.sock`
   - `daemon.pid`
2. Astria starts bundled/configured daemon with:
   - `starclaw daemon start --rpc-socket <socket> --rpc-pidfile <pidfile>`
3. Daemon validates that socket and pidfile are provided together.
4. Daemon starts Desktop RPC listener before writing the pidfile.
5. Daemon starts HTTP server as it does today.
6. Astria waits for HTTP health and Desktop RPC socket availability.
7. Astria connects to the socket and calls `system.capabilities`.
8. Astria validates:
   - protocol version equals supported Desktop RPC protocol;
   - `system.capabilities` is present;
   - daemon/app version is compatible for release builds.
9. Astria opens/restores `/app/`.
10. On disconnect or mismatch, Astria surfaces a visible state and offers
    retry/fallback according to the child task policy.

## Compatibility

- Existing daemon-only mode remains valid when neither socket nor pidfile is
  provided.
- Existing web/browser mode remains valid.
- Calendar tools may remain registered against the broker; if no desktop client
  is connected, tool calls should keep returning `desktop_disconnected`.
- `/status` should continue redacting socket paths.
- The shell must not delete arbitrary paths. Cleanup must be scoped to
  deterministic StarClaw/Astria runtime paths or files it created.

## Failure Model

- Missing socket or pidfile pair: daemon exits before listening.
- Socket bind failure: daemon exits with a clear Desktop RPC listener error.
- Pidfile write failure: daemon exits after cleaning listener artifacts.
- Stale pidfile: Astria verifies PID ownership/process liveness before trusting
  it.
- Stale socket: Astria attempts safe cleanup only under its runtime directory.
- Protocol mismatch: Astria blocks desktop-ready and surfaces upgrade/downgrade
  guidance.
- Version mismatch: dev builds may warn; release builds must fail visibly.
- Desktop disconnect: daemon remains usable through HTTP, broker cancels
  pending Desktop RPC requests, and Astria offers reconnect/retry.

## Kocoro Parity

Kocoro's public repo uses explicit Desktop RPC socket and pidfile flags,
`system.capabilities`, protocol versioning, pidfile-after-listen semantics, and
desktop disconnect cancellation. StarClaw already has much of the daemon-side
RPC package. Phase14 should close the integration gap between that package and
the Phase13 Astria shell.

## Rollout

1. Land daemon flag pair and listener lifecycle tests.
2. Land Astria socket client/capabilities handshake and smoke mode.
3. Land stale artifact handling, diagnostics, and fallback hardening.

Each child must leave HTTP/browser launch usable.
