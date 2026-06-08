# Phase 12 final gap review

## Baseline

- Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.
- StarClaw comparison state: after `e5d0d02 docs: document daemon event
  contracts`.

## Completed lifecycle resilience scope

- `eventbus-replay-sse-resilience`: `/events` now uses atomic
  subscribe-with-replay, accepts `last_event_id` and `Last-Event-ID`, replays
  missed events before live delivery, and the Web UI tracks stream
  reconnect/recovered state.
- `run-session-lifecycle-events`: `RunStore` publishes replayable local
  EventBus lifecycle events for `run_started`, `run_completed`, and
  `run_error` while preserving structured run records and redaction rules.
- `webui-live-recovery`: Astria Web UI consumes lifecycle events, upserts
  recovered run summaries, refreshes `/runs` after EventSource recovery, and
  keeps Mission Control/recovered filters current.
- `event-contract-documentation`: `docs/DAEMON_EVENTS.md` documents `/events`
  replay, `/message` streaming aliases, canonical/legacy event names, privacy
  boundaries, and the intentional Kocoro/Shannon Cloud divergence.

## Evidence

- Runtime/event code:
  - `internal/daemon/events.go`
  - `internal/daemon/server.go`
  - `internal/daemon/run_store.go`
  - `internal/daemon/webui/assets/app.js`
- Tests:
  - `internal/daemon/events_test.go`
  - `internal/daemon/server_test.go`
  - `internal/daemon/run_store_lifecycle_events_test.go`
  - `internal/daemon/webui_test.go`
  - `internal/daemon/event_docs_test.go`
- Documentation:
  - `docs/DAEMON_EVENTS.md`
  - README Local Runtime API link to the daemon event contract.

## Remaining Kocoro differences

- StarClaw keeps local-first defaults. It still does not enable Shannon Cloud
  auth, off-machine telemetry, or Kocoro IM `MESSAGE_LIFECYCLE` transport by
  default.
- StarClaw intentionally preserves legacy event names alongside
  Kocoro-compatible aliases. Kocoro is closer to a single canonical daemon
  event vocabulary in some paths.
- Kocoro has deeper Desktop/cloud route lifecycle machinery. StarClaw now has
  the local replay/lifecycle/recovery foundation, but not a standalone desktop
  application shell or cloud-backed lifecycle stack.
- StarClaw EventBus replay remains bounded in memory. It is sufficient for
  reconnect resilience, not durable all-history event sourcing.

## Current gap estimate

For the Phase12 target area, StarClaw is now substantially aligned with Kocoro
on local lifecycle resilience:

- Event replay and reconnect: mostly closed.
- Run lifecycle vocabulary: mostly closed for local daemon clients.
- Web UI recovery after reconnect/refresh: materially improved and adequate
  for current Astria embedded UI.
- Event contract clarity: closed for current local surfaces.

Remaining Kocoro gap has shifted away from Phase12 and toward product shell /
native app parity: a standalone desktop app, daemon lifecycle supervision,
installer/update UX, Desktop/cloud route depth, and optional cloud transport
boundaries.

## Next recommended phase

Phase13 should focus on standalone application and desktop shell parity:

1. `standalone-desktop-shell-plan`: choose Tauri/Electron/SwiftUI boundary,
   daemon startup/attach model, port discovery, and packaging constraints.
2. `daemon-supervision-app-launcher`: local app launcher starts or attaches to
   daemon, monitors health, and opens Astria.
3. `desktop-window-recovery`: app shell restores windows, reconnects Web UI,
   and surfaces daemon health/crash states.
4. `packaging-signing-update-boundary`: define packaging, signing,
   auto-update, and release safety gates without enabling cloud sync by
   default.

## Closeout

Phase12 closes the lifecycle resilience work planned after Phase11. The
remaining Kocoro parity gap is no longer basic streaming, event replay, run
lifecycle events, or Web UI reconnect recovery. The next gap is standalone
desktop productization and deeper native/cloud lifecycle boundaries.
