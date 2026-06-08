# Web UI delta consumption parity

## Goal

Verify and align the embedded Astria Web UI streaming consumer with the Kocoro-compatible daemon SSE vocabulary added in Phase11. The UI must render streamed assistant deltas as they arrive, not depend on the final `done` payload.

## Requirements

- Consume Kocoro-compatible `delta` SSE events from `POST /message`.
- Consume Kocoro-compatible `assistant_text` SSE events for mid-turn narration.
- Continue supporting StarClaw legacy `text` and `preamble` events.
- Consume or preserve metadata events such as `usage`, `tool`, and `session_started` without throwing or breaking the active stream.
- Avoid duplicate rendering when the daemon dual-emits legacy and Kocoro-compatible events for the same text.
- Preserve existing chat and agent-test streaming UX.

## Acceptance Criteria

- [ ] `streamMessage` handles `delta`, `assistant_text`, `tool`, `usage`, and `session_started` events.
- [ ] `text` + `delta` dual emission does not append duplicate assistant content.
- [ ] `preamble` + `assistant_text` dual emission does not append duplicate narration.
- [ ] Agent-test stream event timeline includes Kocoro-compatible metadata events.
- [ ] Focused Web UI regression tests or static contract checks cover the new event vocabulary.
- [ ] `go test ./internal/daemon` and `go test ./...` pass.

## Notes

- The daemon intentionally dual-emits legacy and Kocoro-compatible aliases. This child is scoped to client-side consumption and duplicate suppression.
