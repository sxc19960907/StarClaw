# Desktop RPC launch contract

## Goal

Expose a daemon launch contract for Astria Desktop RPC mode by adding paired
socket and pidfile flags to `starclaw daemon start`, wiring them into the
existing `internal/daemon/desktop_rpc` listener, and preserving current
HTTP-only daemon behavior when no desktop flags are provided.

## Requirements

- Add `--rpc-socket <path>` and `--rpc-pidfile <path>` to `starclaw daemon
  start`.
- Require both flags together; passing only one must fail before starting the
  HTTP server.
- When both flags are present, start a Desktop RPC listener using the existing
  daemon broker before or alongside HTTP server startup.
- Write the pidfile only after the socket is listening.
- Clean up listener artifacts on daemon shutdown.
- Keep `/status` Desktop RPC fields visible without exposing socket or pidfile
  paths.
- Keep ordinary `starclaw daemon start`, `starclaw app`, and `starclaw app
  --check` compatible.

## Acceptance Criteria

- [ ] `starclaw daemon start` without flags behaves as before.
- [ ] `starclaw daemon start --rpc-socket <path>` fails with an actionable
      missing-pidfile error.
- [ ] `starclaw daemon start --rpc-pidfile <path>` fails with an actionable
      missing-socket error.
- [ ] `starclaw daemon start --rpc-socket <path> --rpc-pidfile <path>` starts
      Desktop RPC listener, writes pidfile after listen succeeds, and reports
      `desktop_rpc.listening=true` through `/status`.
- [ ] Listener startup failure prevents daemon readiness and cleans scoped
      artifacts.
- [ ] Tests cover flag validation, listener wiring, pidfile creation, status
      redaction, and existing app launch compatibility.

## Notes

This child is daemon-side only. Swift socket client and `system.capabilities`
handshake belong to the next child.
