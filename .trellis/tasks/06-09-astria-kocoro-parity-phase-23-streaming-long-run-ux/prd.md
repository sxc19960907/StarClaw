# Astria Kocoro parity phase 23: streaming and long-run UX

## Goal

Close the remaining user-visible gap between StarClaw/Astria's existing local
streaming runtime and Kocoro-style long-running assistant UX. StarClaw already
has provider streaming, `/message` SSE, daemon-wide `/events`, an
OpenAI-compatible streaming gateway, and run control APIs. This phase should
make those capabilities feel continuous, inspectable, and controllable in the
Web UI and local API surfaces without introducing Shannon Cloud transport.

## Confirmed Facts

- `agent.StreamingLLMClient` and provider `StreamChat` implementations already
  exist.
- `/message` uses per-request SSE when the client sends
  `Accept: text/event-stream`, emitting `text`, `delta`, `usage`, `tool`,
  `done`, and `error` events.
- `/v1/chat/completions` supports blocking and `stream:true` OpenAI-compatible
  responses.
- `/events` provides daemon-wide SSE with replay and lifecycle events.
- `POST /cancel` and `POST /runs/{id}/control` exist; pause/resume/cancel
  runtime contracts are already specified and covered at the daemon layer.
- The Web UI already has streaming smoke coverage, but the active run UX still
  hides too much of the runtime state behind final summaries and run details.

## Requirements

- Preserve all existing local-first streaming contracts and Kocoro-compatible
  event aliases.
- Improve the Web UI's active run experience so streaming progress, usage,
  tool/control status, and cancellation state are visible while a run is still
  in flight.
- Keep `/message`, `/events`, and `/v1/chat/completions` behavior compatible
  with existing tests and documented examples.
- Ensure OpenAI-compatible streaming errors still emit an error frame without a
  success stop chunk or `[DONE]`.
- Keep event payloads local and redacted according to existing daemon event
  contracts; do not add cloud telemetry, account sync, or external channel
  transport.
- Add or update smoke/unit tests around any newly surfaced runtime state.

## Acceptance Criteria

- [ ] Web UI chat streaming shows incremental text and a clear active run
      status before the final run summary appears.
- [ ] Web UI surfaces current run id/session id when available, usage deltas,
      and cancellation/control state without requiring the user to open run
      details.
- [ ] Stop/cancel UX remains responsive for streaming runs and records control
      decisions in run detail.
- [ ] `/message` streaming, daemon-wide `/events`, and OpenAI-compatible
      streaming tests continue to pass.
- [ ] Web UI smoke covers the improved active streaming/long-run status.
- [ ] Documentation remains accurate for local streaming and long-run control
      boundaries.

## Out Of Scope

- Shannon Cloud auth, cloud sync, off-machine telemetry, or IM
  `MESSAGE_LIFECYCLE` transport.
- Replacing the existing provider streaming parser stack.
- Full native desktop notification/control UX beyond the current daemon-served
  Web UI.
- Multi-choice OpenAI streaming, tool-call compatibility in the local gateway,
  or OpenAI response formats unless a future task scopes them explicitly.

## Notes

- This is a complex task. Keep `design.md` and `implement.md` in sync before
  implementation starts.
