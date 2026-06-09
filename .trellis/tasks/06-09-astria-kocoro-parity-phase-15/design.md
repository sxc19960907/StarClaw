# Astria Kocoro parity phase 15 design

## Architecture Boundary

Phase15 should extend the Phase14 Desktop RPC handshake rather than replace it.
The launch-time sequence remains:

1. Astria starts or attaches to the daemon.
2. HTTP health proves the WebView fallback can load.
3. Desktop RPC connects over the Astria runtime socket.
4. `system.capabilities` validates protocol and method compatibility.

Phase15 adds a second layer after step 4: a long-lived native session monitor.
That monitor owns app-side state transitions, reconnect attempts, and the first
local event-monitoring path. HTTP remains the recovery surface; Desktop RPC
adds native depth.

## Session Model

The app-side session should distinguish:

- `idle`: daemon has not started or reconciliation is disabled.
- `connecting`: socket connection or capabilities validation is in progress.
- `connected`: Desktop RPC is attached and protocol-compatible.
- `reconnecting`: a previously connected session is trying to recover.
- `degraded`: HTTP fallback is healthy but Desktop RPC is unavailable or
  recoverable.
- `mismatch`: protocol/version compatibility failed and retrying the same
  daemon is not useful.

The daemon-side status should continue to redact local paths and expose only
coarse Desktop RPC state such as listening, connected, and pending request
count.

## Event Monitoring

The first event-monitoring slice should be local and minimal. Acceptable shapes:

- a Desktop RPC subscription method for lifecycle/tool events;
- a daemon-owned event stream surfaced through the existing broker;
- a native app monitor that consumes daemon/RPC state transitions and records
  them as structured local events.

The chosen mechanism must be testable without requiring external services.

## Compatibility

- HTTP-only daemon mode stays valid.
- Browser launch and CLI checks must not require Desktop RPC.
- Existing Desktop RPC request/response framing must remain compatible with
  Phase14 smoke tests.
- Calendar/tool calls may continue to return `desktop_disconnected` until a
  richer native event/tool client is implemented.

## Failure Model

- RPC socket disappears after launch: Astria enters `reconnecting`, then
  `degraded` if bounded retries fail.
- Daemon HTTP remains healthy but RPC is broken: WebView stays mounted and the
  user sees a recoverable Desktop RPC state.
- Daemon exits: Astria follows existing daemon supervision and route/window
  recovery, then re-runs capabilities reconciliation when the daemon returns.
- Protocol mismatch: Astria enters `mismatch` and avoids retry loops against the
  same incompatible daemon.
- Event monitor fails: session remains usable if request/response RPC is still
  healthy, but diagnostics must reflect event monitoring degradation.

## Rollout

1. Land app-side long-lived session lifecycle and bounded reconnect behavior.
2. Land local event monitoring with focused daemon/app tests.
3. Land diagnostics/recovery UX that consumes session/event state.

Each child must leave Phase14 smoke and Go tests green.
