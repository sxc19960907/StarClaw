# Add GUI run history detail

## Goal

Add a real Run history/detail workflow in the GUI so users can inspect recent daemon runs, their request metadata, result, usage, approvals, and tool events.

## Requirements

- Backend:
  - Record recent HTTP/SSE daemon agent runs in memory.
  - Capture run id/request id, status, agent, channel, prompt preview, session id, start/end timestamps, usage, error, and run events.
  - Capture tool calls/results, text/preamble/stream events, and approval events where available.
  - Expose `GET /runs` and `GET /runs/{id}`.
- GUI:
  - Add a `Runs` navigation panel.
  - Show recent runs with status, agent, session, and timestamp.
  - Show selected run detail including request metadata, result/session, usage, and event timeline.
  - Let Chat run summary open the corresponding run detail.
- Compatibility:
  - Existing chat, session, approval, and smoke behavior must continue to work.
  - Run history can be in-memory for this iteration; persistence is out of scope.

## Acceptance Criteria

- [ ] A successful `/message` call appears in `GET /runs`.
- [ ] `GET /runs/{id}` returns full run detail and event timeline.
- [ ] GUI `Runs` panel lists recent runs and opens detail.
- [ ] Chat run summary includes an `Open run` action linked to the recorded run.
- [ ] Web UI smoke verifies run list/detail and summary-to-detail navigation.
- [ ] Go tests cover run store/API basics.

## Notes

- This is a module-level task. Do not split into small UI-only tasks.
