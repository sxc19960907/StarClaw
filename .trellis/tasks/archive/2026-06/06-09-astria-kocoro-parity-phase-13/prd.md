# Astria Kocoro parity phase 13: standalone desktop app shell

## Goal

Close the next Kocoro parity gap after Phase12 by moving Astria from a
browser-opened embedded Web UI toward a real standalone desktop application
shell that can own daemon startup, attach/reuse behavior, window recovery, and
release boundaries.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase12 closed local event replay, run lifecycle events, Web UI recovery
  after reconnect, and event contract documentation.
- StarClaw currently exposes Astria through `starclaw app`, which starts or
  reuses the daemon and opens `http://127.0.0.1:7533/app/` in the user's
  browser.
- `starclaw app --check`, `starclaw doctor`, and the Web UI Version page
  already expose launch readiness, local URLs, diagnostics, data path, and
  config path.
- StarClaw already has a local `internal/daemon/desktop_rpc` Unix-socket
  protocol surface for native desktop integration, but no bundled `.app`
  process that connects as the desktop client.
- Kocoro's Desktop model uses a native app as the user-facing shell. The
  Desktop app spawns or attaches to the daemon, passes `--rpc-socket` and
  `--rpc-pidfile`, performs `system.capabilities` version reconciliation, emits
  `desktop_online`, and surfaces daemon lifecycle failures to the user.
- StarClaw must preserve local-first defaults. Phase13 must not add Shannon
  Cloud auth, off-machine telemetry, or cloud-backed lifecycle transport by
  default.

## Child Plan

1. `standalone-desktop-shell-plan`: choose and document the native shell
   boundary, repository layout, daemon attach/start model, and first runnable
   MVP scope.
2. `daemon-supervision-app-launcher`: implement the app-side launcher contract
   needed to start or attach to the local daemon, monitor health, and open
   Astria inside the shell.
3. `desktop-window-recovery`: restore windows and Web UI state across shell
   restart, daemon reconnect, and daemon crash/restart scenarios.
4. `packaging-signing-update-boundary`: define packaging, signing,
   notarization/update boundaries, release smoke checks, and explicit
   local-first safety gates.

## Requirements

- Build toward a standalone app shell while keeping existing CLI and browser
  launch paths working.
- Keep Astria's current embedded Web UI as the primary product surface unless
  a child task explicitly scopes a native replacement for a narrow control.
- Prefer the smallest native shell that can supervise the daemon and host the
  Web UI; do not fork the UI into an unrelated frontend stack.
- Use Kocoro's Desktop/daemon lifecycle model as the parity reference, but
  preserve StarClaw naming, paths, and local-first privacy boundaries.
- Do not enable real cloud sync, cloud lifecycle routing, or remote telemetry
  as part of Phase13.
- Make each child independently testable with CLI checks, unit tests, and
  documented manual smoke steps where native packaging cannot be fully
  exercised in CI.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and
      testable acceptance criteria before implementation starts.
- [ ] The selected standalone shell architecture is documented with explicit
      trade-offs against Electron/Tauri/SwiftUI and current CLI-only launch.
- [ ] Daemon supervision requirements cover fresh launch, reuse existing
      daemon, version mismatch, unhealthy daemon, port conflict, and crash
      recovery.
- [ ] Window recovery requirements cover shell restart, Web UI reload,
      EventSource reconnect, and user-visible daemon health/crash states.
- [ ] Packaging requirements cover local development builds, release artifact
      shape, signing/notarization boundaries, updater behavior, and smoke
      validation.
- [ ] Existing `starclaw app`, `starclaw app --no-open`, and `starclaw app
      --check` behavior remains compatible unless explicitly changed and
      documented.
- [ ] Phase can close only after all children are archived and a final Kocoro
      gap review is recorded.

## Out of Scope

- Replacing Astria Web UI with a full native UI.
- Adding Shannon Cloud auth, cloud sync, or off-machine telemetry by default.
- Implementing macOS Calendar/Contacts/Reminders tools unless needed only as
  smoke fixtures for the existing Desktop RPC boundary.
- Shipping production signing credentials or private updater keys in the
  repository.

## Notes

Parent task only. Start child tasks for implementation.
