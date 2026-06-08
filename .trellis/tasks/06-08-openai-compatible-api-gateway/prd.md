# OpenAI compatible API gateway

## Goal

Add an OpenAI-compatible local API gateway surface so external tools can call Astria/StarClaw through familiar chat-completions style endpoints.

## Requirements

- Provide a scoped local endpoint compatible with the core shape of OpenAI chat completions requests.
- Map compatible requests to existing StarClaw agent/session/run execution without bypassing permissions.
- Return OpenAI-style response envelopes for non-streaming requests.
- Preserve existing daemon API behavior and authentication/local-only assumptions.
- Document unsupported fields clearly instead of silently accepting unsafe behavior.

## Acceptance Criteria

- [x] Endpoint accepts a minimal chat-completions request with model/messages.
- [x] Endpoint rejects unsupported or unsafe request shapes with clear errors.
- [x] Endpoint returns an OpenAI-style response envelope for successful local execution.
- [x] Existing `/message`, `/runs`, and Web UI flows continue to pass.
- [x] Tests cover request validation, response shape, and unsupported-field behavior.

## Notes

- This is a backend/API slice, not UI polish.
