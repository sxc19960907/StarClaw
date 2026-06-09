# Desktop RPC fallback recovery design

## Decision

Add Astria-side recovery for Desktop RPC runtime artifacts and degraded HTTP
fallback without extending the Desktop RPC protocol in this child.

The shell should:

1. clean stale `daemon.sock` / `daemon.pid` only inside its configured runtime
   directory;
2. remove dead-process pidfiles before launching a new daemon;
3. remove stale sockets only when scoped to the runtime directory and not
   serving a valid Desktop RPC handshake;
4. keep Web UI usable through HTTP when an already-running daemon is healthy
   but Desktop RPC is absent or broken;
5. fail visibly when a daemon launched by Astria cannot complete Desktop RPC
   reconciliation.

## Scope

In scope:

- Swift runtime artifact ownership checks.
- Stale pidfile dead-process detection.
- Stale socket cleanup under `ASTRIA_RUNTIME_DIR` / Application Support.
- User-visible degraded HTTP fallback state.
- Smoke helpers for stale artifacts and unsafe cleanup refusal.
- Docs/spec updates for recovery boundaries.

Out of scope:

- Killing existing daemon processes.
- Deleting artifacts outside Astria runtime paths.
- Persistent Desktop RPC connection monitoring.
- Calendar/EventKit native tool implementation.
- Replacing HTTP health monitoring.

## Artifact Safety Contract

Astria may delete only these exact paths:

- `<runtime-dir>/daemon.sock`
- `<runtime-dir>/daemon.pid`

The runtime directory is either:

- default Application Support path:
  `~/Library/Application Support/dev.starclaw.astria`; or
- explicit `ASTRIA_RUNTIME_DIR` / `--runtime-dir` override for smoke tests.

Astria must not delete:

- parent directories;
- symlink targets outside the runtime directory;
- arbitrary paths provided by future flags;
- files with unexpected names.

## Fallback States

- Shell-launched daemon:
  - HTTP health ok + Desktop RPC ok -> `attached`.
  - HTTP health ok + Desktop RPC failure -> `failed`.
- Existing healthy daemon:
  - no Desktop RPC socket -> `degraded`.
  - Desktop RPC socket fails reconciliation -> `degraded`.

`degraded` keeps the WebView mounted and shows a banner that Astria is using the
HTTP fallback. This preserves local usability while making the Kocoro gap
visible.

## Failure Model

- Dead pidfile PID -> remove pidfile and stale socket if both are scoped runtime
  artifacts.
- Malformed pidfile -> remove pidfile and scoped socket.
- Live pidfile PID -> do not remove artifacts.
- Stale socket under runtime dir -> remove before launching if it cannot
  complete `system.capabilities`.
- Socket/pidfile outside scoped runtime dir -> refuse cleanup.

## Kocoro Parity

This closes the practical recovery gap around stale local Desktop RPC artifacts
and visible degraded fallback. It still does not implement long-lived native
tool RPC monitoring; that belongs to later OS integration work.
