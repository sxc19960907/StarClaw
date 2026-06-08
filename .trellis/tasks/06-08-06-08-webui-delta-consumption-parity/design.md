# Web UI delta consumption parity design

## Current State

StarClaw's daemon now emits both legacy SSE events (`text`, `preamble`, `tool_call`, `tool_result`) and Kocoro-compatible aliases (`delta`, `assistant_text`, `tool`, `usage`, `session_started`). The embedded Web UI already streams `/message` with `Accept: text/event-stream`, but the parser switch currently appends only `text` and `preamble`.

## Desired Contract

- The UI should render Kocoro-compatible `delta` as assistant text.
- The UI should render Kocoro-compatible `assistant_text` as narration.
- When legacy and alias events arrive as an adjacent dual-emission pair with the same text, render once.
- `tool`, `usage`, and `session_started` should be passed to renderers as stream events so timelines/debug views can show them.
- Unknown events remain tolerated.

## Implementation Shape

- Add small duplicate-suppression state inside `streamMessage`, scoped to one stream.
- Normalize text payload extraction:
  - `text`: `data.text`
  - `delta`: `data.text || data.delta`
  - `preamble`: `data.preamble`
  - `assistant_text`: `data.text || data.preamble`
- Call `renderer.onEvent` for metadata and progress events including `tool`, `usage`, and `session_started`.
- Add static/Node-based regression coverage for the event switch if the repo has no browser test harness for this exact helper.

## Compatibility

Legacy clients keep working because daemon events are unchanged. The UI becomes forward-compatible with the Kocoro-compatible vocabulary without requiring daemon changes.
