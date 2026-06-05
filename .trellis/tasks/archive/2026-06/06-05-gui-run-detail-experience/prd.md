# Improve GUI Run Detail Experience

## Problem

The GUI can already run chats, test agents, open sessions, inspect runs, and show grouped tool events in Run detail. The remaining friction is that users still have to manually select result/tool text for common diagnostics, and the Agent Test streaming view does not use the same grouped tool timeline presentation as Run detail.

## User Value

When testing an agent or inspecting a run, a user should be able to quickly copy the important output and understand tool activity without switching panels or parsing raw event rows.

## Confirmed Facts

- Run detail already supports `Copy summary`, `Copy prompt`, `Open session`, and `Re-run`.
- Run detail already groups persisted `tool_call` and `tool_result` events into one tool card.
- Agent Test streams events inline but currently renders each stream event as a raw event row.
- Existing Web UI smoke covers agent test, run detail, copied summaries, re-run, sessions, streaming provider, and a targeted real tool-call path.

## Requirements

- Add Run detail action for copying the formatted run result.
- Add copy action on each grouped Run detail tool card for the tool result payload.
- Reuse the Run detail timeline grouping/rendering logic for Agent Test streaming events, so live test output and persisted run detail present tools consistently.
- Keep changes frontend-only unless a backend contract gap is discovered.
- Extend smoke coverage without adding the targeted real tool-call script to default CI.

## Out of Scope

- New backend run APIs.
- New visual design system or navigation restructure.
- Adding targeted tool-call smoke to CI.
- Changing agent execution semantics or permissions.

## Acceptance Criteria

- [x] Run detail has a `Copy result` action that copies `formatRunResponse(run.response)`.
- [x] Grouped tool cards in Run detail have a `Copy result` action when a result payload exists.
- [x] Agent Test streaming uses grouped tool cards for tool call/result events instead of separate raw rows.
- [x] Existing Run detail actions still work: copy summary, copy prompt, open session, re-run.
- [x] Web UI smoke covers copy result and tool-result copy.
- [x] `scripts/smoke_webui_core.sh` passes locally.
- [x] `scripts/smoke_webui_tool_call.sh` passes locally.
- [x] `go test ./...`, `go vet ./...`, and `git diff --check` pass locally.
