# Desktop RPC fallback recovery

## Goal

Harden Astria Desktop RPC lifecycle recovery around stale pidfiles, stale
sockets, disconnects, and fallback to HTTP health without weakening local-first
or file-safety boundaries.

## Requirements

- Detect stale pidfile and socket states under Astria/StarClaw runtime paths.
- Clean up only files that are deterministic runtime artifacts or were created
  by the current app launch.
- Surface broken socket, stale pidfile, disconnected Desktop RPC client, and
  fallback HTTP-only states in the shell.
- Preserve daemon usability through HTTP when Desktop RPC disconnects.
- Ensure broker pending requests are cancelled on Desktop disconnect.
- Document support/debug workflow through `/status`, `starclaw doctor`, and
  shell diagnostics.

## Acceptance Criteria

- [ ] Stale pidfile with dead process does not block a fresh app launch.
- [ ] Stale socket under runtime dir is cleaned safely before relaunch.
- [ ] Socket/pidfile outside allowed runtime boundary is not deleted silently.
- [ ] Desktop disconnect keeps HTTP daemon usable and reports degraded state.
- [ ] Tests/smoke cover stale artifacts, disconnect, retry, and fallback.

## Notes

This child depends on the launch contract and capabilities reconciliation.
