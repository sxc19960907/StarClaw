# Channel Inbox MVP

## Goal

Add the first external channel inbox so Astria can receive tasks or messages from outside the Web UI, route them through approvals, and hand them off to normal runs. This is inspired by Kocoro's channel messaging direction but should start narrow.

## Requirements

- Choose one initial channel provider during design, likely Feishu or Telegram depending on available credentials and user preference.
- Show inbound messages/tasks in an Inbox surface.
- Allow the user to approve, reject, or convert an inbound item into an Astria run.
- Preserve source metadata and reply status.
- Avoid unattended execution by default.
- Handle provider failures and duplicate delivery safely.

## Acceptance Criteria

- [x] One provider can ingest an inbound message into Astria state.
- [x] User can convert an inbox item into a run through an approval step.
- [x] Duplicate provider events do not create duplicate active tasks.
- [x] Failure states are visible and retryable.
- [x] Relevant daemon/provider tests pass.

## Non-Goals

- No multi-provider framework in MVP.
- No always-on automation without explicit approval.
- No public hosted relay unless separately planned.

## Dependencies

- Depends on daemon/background reliability and a clear approvals model.
- Should be lower priority than local tool and MCP work unless the user's workflow needs channels immediately.
