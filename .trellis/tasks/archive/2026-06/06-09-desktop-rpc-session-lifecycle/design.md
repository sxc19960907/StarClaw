# Desktop RPC session lifecycle design

## Current Shape

`DaemonSupervisor` owns daemon lifecycle and HTTP health state. Desktop RPC is
currently probed once by `reconcileDesktopRPC()`, which opens a Unix socket,
calls `system.capabilities`, validates the result, and closes the socket.
`startHealthMonitor()` is the only long-lived monitor today.

## Proposed Shape

Add a second long-lived monitor inside `DaemonSupervisor`:

- `desktopRPCMonitorTask`: background task parallel to `healthMonitorTask`.
- `DesktopRPCSessionState`: lightweight internal state for smoke tests and
  future diagnostics.
- `monitorDesktopRPC()`: periodically validates the socket with the same
  `system.capabilities` request used by Phase14, with bounded retries on
  recoverable failures.

This does not yet introduce a daemon event subscription. That is reserved for
`desktop-rpc-event-monitoring`.

## State Transitions

- Initial reconciliation succeeds: state becomes `connected`; monitor starts.
- Single transient RPC probe failure: state becomes `reconnecting`.
- Retry succeeds: state returns to `connected`; UI daemon state remains or
  returns to `.attached`.
- Bounded retries fail while HTTP is healthy: UI daemon state becomes
  `.degraded(...)`.
- Protocol/version mismatch: state becomes `mismatch`; monitor exits because
  retrying the same daemon is not useful.
- HTTP health failure: existing health monitor owns `.unavailable` and
  recovery; RPC monitor may remain degraded until daemon recovery restarts
  reconciliation.

## Compatibility

- The daemon wire protocol is unchanged.
- HTTP-only daemon mode and browser/CLI launch stay valid.
- Status redaction and runtime artifact cleanup rules from Phase14 stay intact.

## Test Strategy

Add a Swift smoke entrypoint that uses a fake local Unix socket server to verify:

- successful session probe marks the session connected;
- recoverable probe failures are retried and become degraded after the retry
  bound;
- protocol mismatch is terminal and not retried as a recoverable disconnect.
