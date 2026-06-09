# Design: Streaming and Long-Run UX

## Architecture

Phase 23 builds on the existing runtime surfaces:

- Provider streaming: `agent.StreamingLLMClient` and client `StreamChat`.
- Per-run stream: `POST /message` with `Accept: text/event-stream`.
- Daemon stream: `GET /events` with replay.
- OpenAI gateway stream: `POST /v1/chat/completions` with `stream:true`.
- Control APIs: `POST /cancel` and `POST /runs/{id}/control`.

The implementation should avoid new protocols. Instead, it should make the
already emitted stream and control events more visible in Astria.

## Web UI Data Flow

1. Chat submission opens a per-run `/message` SSE request.
2. `session_started`, `text`/`delta`, `tool`, `usage`, `done`, and `error`
   events update the live chat renderer.
3. The active chat surface keeps a compact runtime strip with:
   - run id when known,
   - session id when known,
   - stream state,
   - usage when emitted,
   - last tool/control event summary,
   - stop/cancel result.
4. The final `done` payload still renders the existing run summary and refreshes
   persisted run/session state.

## API Compatibility

- `/message` SSE event names and Kocoro aliases remain unchanged.
- `/events` replay behavior remains unchanged.
- `/v1/chat/completions` continues to stream OpenAI-style
  `chat.completion.chunk` frames and exactly one `[DONE]` on success.
- Streaming error behavior must remain terminal and must not emit success
  completion markers after an error.

## Privacy

The live UI may display local prompt/result text in the same places it already
does today. New event/status summaries must not expose raw tool args, provider
payloads, secrets, or hidden memory/preflight content.

## Trade-Offs

- Prefer enhancing the existing chat stream renderer over adding a separate
  "stream monitor" panel. The user already watches Chat during a run.
- Keep daemon contract changes minimal. If a field can be derived from existing
  stream events, derive it client-side instead of adding a new event type.
- Use smoke tests for product-visible behavior and unit tests only where daemon
  contracts change.

## Rollback

The Web UI changes should be CSS/JS-local and removable without changing daemon
contracts. If runtime event handling regresses, revert the UI renderer changes
  while keeping existing backend streaming tests as the source of truth.
