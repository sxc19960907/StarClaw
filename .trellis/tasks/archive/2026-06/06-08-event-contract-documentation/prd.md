# Event contract documentation

## Goal

Document StarClaw/Astria daemon event contracts after Phase12 so local clients
can consume `/message` streaming SSE, `/events` replay SSE, run lifecycle
events, and legacy/Kocoro-compatible aliases consistently.

## Requirements

- Add user/developer-facing documentation for daemon event surfaces.
- Cover `/events` replay behavior, cursor sources, EventBus ID semantics,
  keepalive behavior, bounded history, and slow-client drop behavior.
- Cover canonical EventBus names: approval events, run lifecycle events,
  run status/usage/tool events, budget events, and cloud delegation boundary
  events.
- Cover `/message?stream=true` per-request SSE names and legacy aliases:
  `tool_call`, `tool_result`, `tool`, `text`, `delta`, `preamble`,
  `assistant_text`, `usage`, `session_started`, `done`, and `error`.
- Describe payload privacy/redaction expectations and unsafe fields.
- Explicitly document intentional divergence from Kocoro/Shannon Cloud:
  local-first, no real cloud transport, no IM `MESSAGE_LIFECYCLE` by default.
- Link the new event contract documentation from the README Local Runtime API
  section.

## Acceptance Criteria

- [ ] `docs/DAEMON_EVENTS.md` exists and documents both `/events` and
      per-request `/message` SSE contracts.
- [ ] Documentation lists replay cursor semantics for `last_event_id` and
      `Last-Event-ID`.
- [ ] Documentation lists canonical and compatibility/legacy event names.
- [ ] Documentation states privacy boundaries and local-first divergence from
      Kocoro/Shannon Cloud.
- [ ] README links to the event contract document.
- [ ] Tests assert the documentation route/markers exist so future changes do
      not silently drop the contract.
