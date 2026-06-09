# Desktop RPC capabilities reconciliation design

## Decision

Make the Astria macOS shell use the daemon's Desktop RPC socket as a readiness
gate when it starts a daemon in desktop mode. The shell should still use HTTP
health as the daemon liveness signal, but it should not declare the desktop
handshake ready until `system.capabilities` succeeds and the response matches
the supported protocol and required system methods.

This child intentionally keeps the first reconciliation narrow:

- use the existing Desktop RPC frame format;
- request only `system.capabilities`;
- validate protocol and required methods;
- launch daemon with deterministic socket and pidfile paths;
- preserve current HTTP fallback and retry behavior.

## Runtime Paths

Astria should resolve runtime artifacts under the user's Application Support
directory:

- `~/Library/Application Support/dev.starclaw.astria/daemon.sock`
- `~/Library/Application Support/dev.starclaw.astria/daemon.pid`

Test/smoke code may use a temporary runtime directory to avoid touching user
state.

## Desktop RPC Client

Swift client responsibilities:

1. Open a Unix domain socket with `socket(AF_UNIX, SOCK_STREAM, 0)`.
2. Write Desktop RPC frames with a 4-byte big-endian length prefix and JSON
   envelope.
3. Send a `desktop_rpc_request` payload:
   - `request_id`
   - `method: "system.capabilities"`
4. Read one `desktop_rpc_result` frame.
5. Decode:
   - `version`
   - `methods`
   - `platform.app_version`
6. Validate:
   - `version == "1.0.0"`
   - methods include `system.ping` and `system.capabilities`
   - release builds do not silently accept incompatible app/daemon versions

## Launch Flow

1. Astria checks HTTP health.
2. If no daemon is healthy, Astria starts `starclaw daemon start` with:
   - `--rpc-socket <socket path>`
   - `--rpc-pidfile <pidfile path>`
3. Astria waits for HTTP health.
4. Astria runs Desktop RPC capabilities reconciliation.
5. If reconciliation succeeds, state becomes attached.
6. If reconciliation fails after Astria launched the daemon in desktop mode,
   state becomes failed with an actionable message.
7. If an already-running daemon is healthy but has no Desktop RPC socket, keep
   HTTP attach behavior for now and leave stale/reconnect fallback hardening to
   the next child.

## Compatibility

- `starclaw app` and daemon-only mode are unchanged.
- The app keeps `ASTRIA_STARCLAW_BIN` / bundled / PATH resolution order.
- `ASTRIA_RUNTIME_DIR` may override the runtime directory for smoke tests.
- Existing `--supervision-smoke` should exercise the desktop launch path.
- The route recovery smoke stays independent.

## Failure Model

- Socket missing after desktop launch -> fail with Desktop RPC connection
  diagnostic.
- Malformed frame/result -> fail with protocol diagnostic.
- Protocol mismatch -> fail with upgrade/downgrade diagnostic.
- Missing required system method -> fail with capability diagnostic.
- Version mismatch in development -> warn through smoke validation helpers; hard
  fail policy for release builds remains documented for later enforcement.

## Kocoro Parity

This closes the immediate gap where the native shell only waits for HTTP
health. It does not yet implement stale pidfile/socket cleanup or long-lived
Desktop disconnect recovery; those remain in `desktop-rpc-fallback-recovery`.
