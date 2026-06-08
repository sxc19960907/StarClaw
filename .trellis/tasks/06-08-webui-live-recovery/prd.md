# WebUI live recovery

## Goal

Make the embedded Astria Web UI recover live run state after `/events`
reconnects, replay, or daemon-side run lifecycle updates. This continues
Phase12 by connecting the EventBus replay/lifecycle work to the browser
client without introducing a standalone desktop shell or remote transport.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- StarClaw Web UI already opens `new EventSource("/events")`.
- The Web UI tracks `state.eventStream.lastEventID`, `status`,
  `reconnects`, and `reconnecting`, and marks daemon status as
  reconnecting/recovered.
- The backend now publishes replayable `run_started`, `run_completed`, and
  `run_error` lifecycle events on `/events`.
- The Web UI already has `/runs`, `/runs/{id}`, and `/runs/{id}/trace`
  rendering for run list, run detail, recovery badges, timeline, and Mission
  Control counters.
- Current EventSource handling only consumes approval events. It does not
  react to run lifecycle events or refresh `/runs` after recovery.

## Requirements

- Consume `run_started`, `run_completed`, and `run_error` EventSource events.
- Update the run list state from lifecycle payloads without leaking prompt or
  raw response content.
- Mark runs recovered when they arrive through EventBus lifecycle replay or
  reconnect recovery.
- Refresh `/runs` after EventSource recovery so durable run detail, timeline,
  and pending/terminal status converge with backend state.
- Preserve existing approval event handling, chat streaming behavior, and
  manual refresh behavior.
- Keep this scoped to embedded Web UI static assets and static/Go tests.

## Out of Scope

- Standalone desktop app shell, packaging, menu bar, daemon launcher, or
  installer behavior.
- New backend routes unless Web UI recovery cannot be achieved through
  existing `/events` and `/runs`.
- Kocoro/Shannon Cloud transport or IM `MESSAGE_LIFECYCLE` protocol.
- Full deterministic reconstruction of a streamed assistant response from
  historical deltas.

## Acceptance Criteria

- [ ] Web UI has explicit handlers for `run_started`, `run_completed`, and
      `run_error`.
- [ ] A lifecycle event upserts `state.runs`, updates Mission Control counters,
      and keeps the active run selection valid.
- [ ] EventSource recovery triggers a guarded `/runs` refresh after reconnect.
- [ ] Lifecycle-updated runs can be identified as recovered in the existing
      recovered filter/card path.
- [ ] Static Web UI tests assert the recovery markers and guard against prompt
      or raw payload leakage in lifecycle handling.
