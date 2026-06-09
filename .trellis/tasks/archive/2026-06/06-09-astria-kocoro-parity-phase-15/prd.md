# Astria Kocoro parity phase 15: long-lived Desktop RPC session and native event monitoring

## Goal

Close the next Kocoro parity gap after Phase14 by turning Astria's Desktop RPC
integration from launch-time reconciliation into a long-lived native session
boundary. Astria should keep observing the daemon/RPC relationship after the
initial `system.capabilities` check, surface ongoing health transitions, and
prepare the native shell for event-driven Desktop RPC capabilities.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase14 added daemon `--rpc-socket` / `--rpc-pidfile` launch contract,
  app-side Unix socket connection, `system.capabilities` validation, scoped
  stale artifact cleanup, and degraded HTTP fallback.
- Phase14 final gap review estimates Astria is now roughly 85-90% aligned for
  local-first desktop lifecycle behavior.
- The remaining Kocoro gap is not basic launch; it is the depth of the native
  Desktop RPC session after launch.
- Existing browser/CLI daemon flows must stay valid without a desktop client.
- StarClaw must preserve local-first boundaries: no Shannon Cloud auth, remote
  lifecycle routing, or off-machine telemetry.

## Child Plan

1. `desktop-rpc-session-lifecycle`: maintain an app-side Desktop RPC session
   after the initial handshake, detect disconnect/reconnect transitions, and
   expose state without blocking HTTP fallback.
2. `desktop-rpc-event-monitoring`: define and implement the first local
   Desktop RPC event-monitoring path so the native shell can receive ongoing
   daemon/session signals instead of only polling launch state.
3. `native-desktop-diagnostics-recovery`: turn session and event state into
   useful native diagnostics/recovery UX and smoke-testable states.

## Requirements

- Keep Phase14 launch compatibility: `starclaw daemon start`,
  `starclaw daemon start --rpc-socket ... --rpc-pidfile ...`, `starclaw app`,
  and `starclaw app --check` must continue to work.
- Keep Desktop RPC optional for browser/CLI flows and mandatory only for the
  Astria desktop shell when reconciliation is enabled.
- Preserve HTTP WebView fallback when Desktop RPC is disconnected or degraded.
- Add durable session state after initial capabilities reconciliation:
  connected, disconnected, reconnecting, degraded, and terminal mismatch states
  must be distinguishable.
- Add local event monitoring without adding cloud routing, remote telemetry, or
  Shannon auth dependencies.
- Do not delete arbitrary socket/pidfile paths; runtime cleanup remains scoped
  to Astria-owned artifacts.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and
      testable acceptance criteria before implementation starts.
- [ ] Astria maintains long-lived Desktop RPC session state after launch and
      detects daemon/RPC disconnects without tearing down healthy HTTP fallback.
- [ ] Session reconnect/retry behavior is bounded, observable, and does not
      spin or leak native tasks.
- [ ] The daemon and/or app expose enough local state to distinguish no desktop
      client, connected desktop client, pending work, and disconnected/degraded
      desktop session.
- [ ] A first event-monitoring mechanism exists for local Desktop RPC lifecycle
      or tool/session events, with tests or smoke coverage.
- [ ] Native diagnostics/recovery surfaces session state in a user-visible way
      without exposing unsafe local paths.
- [ ] Smoke/unit coverage exercises steady connected session, daemon/RPC
      disconnect, reconnect or retry, degraded fallback, and event monitoring.

## Out of Scope

- Full native UI replacement.
- Production signing/notarization or auto-update implementation.
- Native Calendar/Contacts/Reminders tool implementation beyond event/session
  plumbing needed for this phase.
- Remote cloud lifecycle coordination, telemetry, or Shannon Cloud auth.
- Cross-platform native shells beyond the macOS Astria shell.

## Notes

Parent task only. Start child tasks for implementation.
