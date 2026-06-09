# Astria Kocoro parity phase 14: desktop RPC handshake and daemon reconciliation

## Goal

Close the next Kocoro parity gap after Phase13 by moving Astria daemon
supervision from HTTP health-only attach/start toward a semi-bound Desktop RPC
lifecycle: explicit socket and pidfile paths, `system.capabilities`
reconciliation, and app/daemon version compatibility checks.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase13 added a thin macOS Astria shell, daemon health supervision, route
  recovery, bundled daemon support, and packaging/signing/update boundaries.
- Phase13 final gap review estimates Astria is now roughly 80-85% aligned for
  local-first standalone shell behavior.
- StarClaw already has `internal/daemon/desktop_rpc` with frame codec, broker,
  listener, `system.ping`, `system.capabilities`, protocol version `1.0.0`,
  pidfile writing, and `/status` Desktop RPC visibility.
- StarClaw's `cmd daemon start` currently starts HTTP daemon mode only; it does
  not expose app-provided `--rpc-socket` / `--rpc-pidfile` launch flags.
- The Astria Swift shell currently starts `starclaw daemon start` and waits on
  `/health`; it does not connect to Desktop RPC or validate capabilities before
  declaring the desktop ready.
- Existing CLI/browser launch paths must remain valid and must not require a
  desktop client.

## Child Plan

1. `desktop-rpc-launch-contract`: add daemon launch flags and Astria app socket
   / pidfile path selection while preserving HTTP-only daemon mode.
2. `desktop-rpc-capabilities-reconciliation`: make Astria connect to the Unix
   socket, call `system.capabilities`, and fail visibly on protocol or version
   mismatch.
3. `desktop-rpc-fallback-recovery`: harden stale pidfile/socket cleanup,
   fallback behavior, status diagnostics, and smoke coverage.

## Requirements

- Add explicit desktop launch inputs for daemon socket and pidfile paths.
- Require socket and pidfile flags as a pair; partial desktop launch
  configuration must fail fast with a clear error.
- Keep standalone `starclaw daemon start`, `starclaw app`, and `starclaw app
  --check` behavior compatible.
- Make Astria choose deterministic local paths for desktop socket and pidfile
  under the StarClaw/Astria local app-data boundary.
- Make Astria verify Desktop RPC capabilities before treating the desktop
  connection as ready.
- Surface protocol mismatch, app/daemon version mismatch, stale pidfile, broken
  socket, and disconnected desktop states in a user-visible way.
- Preserve local-first privacy boundaries: no cloud lifecycle routing, remote
  telemetry, or Shannon Cloud auth.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and
      testable acceptance criteria before implementation starts.
- [ ] `starclaw daemon start --rpc-socket <path> --rpc-pidfile <path>` starts a
      Desktop RPC listener, writes the pidfile only after listening, and
      reports Desktop RPC status through `/status`.
- [ ] Passing only one of socket or pidfile fails with an actionable error and
      does not start a partial desktop listener.
- [ ] Astria launches bundled or configured daemon with explicit socket and
      pidfile paths when desktop reconciliation is enabled.
- [ ] Astria calls `system.capabilities`, validates protocol version and app /
      daemon compatibility, and only then considers desktop RPC attached.
- [ ] Stale pidfile/socket states are detected and either cleaned safely or
      surfaced clearly without deleting unrelated user files.
- [ ] HTTP health fallback remains available for CLI/browser launch and for
      explicit recovery paths.
- [ ] Smoke and unit coverage exercise successful handshake, missing flag pair,
      protocol/version mismatch, stale socket/pidfile behavior, and fallback.

## Out of Scope

- Full native UI replacement.
- Production signing/notarization or auto-update implementation.
- Calendar/Contacts/Reminders native tool implementations beyond existing
  Desktop RPC fixtures.
- Remote cloud lifecycle coordination or telemetry.
- Cross-platform Desktop RPC clients beyond the macOS Astria shell.

## Notes

Parent task only. Start child tasks for implementation.
