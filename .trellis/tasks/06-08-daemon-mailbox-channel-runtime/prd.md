# Daemon mailbox channel runtime

## Goal

Implement the next Phase6 Kocoro parity slice: add a local daemon mailbox/channel runtime foundation that can accept typed queued messages by route key, preserve priority/FIFO ordering, expose safe queue state, and prepare inbox/channel delivery for durable claim/ack worker semantics.

## Requirements

- Use the local Kocoro checkout as parity evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/agenttypes/mailbox.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/mailbox_store.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/queue_enqueue_handler.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/channel_state_*`
- Keep the first implementation local-first. Do not add Slack/Feishu/Telegram/Shannon Cloud transport in this child.
- Add typed queued message models under StarClaw daemon/runtime code with route key, optional session id, source, sender, text, priority, enqueue timestamp, claim status, and safe metadata.
- Add an in-memory mailbox store with per-route capacity, priority ascending ordering, FIFO within a priority, defensive snapshots, claim/ack/release behavior, and dedup by provider/external id where present.
- Add daemon HTTP APIs for local clients/tests to enqueue and inspect mailbox messages.
- Preserve existing `/inbox` behavior unless explicitly migrated in the implementation plan; this child may add bridge helpers but should not remove the guarded inbox approval workflow.
- Queue state and metrics must not expose raw secrets, provider payloads, tool args, or hidden channel credentials.
- Keep future persistent storage design explicit, but do not introduce SQLite or a DB dependency unless the design and tests prove it is needed for this child.

## Acceptance Criteria

- [ ] `POST /queue` accepts a local queued message with `route_key` or `session_id`, `text`, optional `source`, `sender`, `external_id`, `agent`, and `priority`.
- [ ] `POST /queue` rejects missing route/session, empty text, oversized text, and full mailbox capacity with appropriate HTTP status codes.
- [ ] Duplicate source/external id on the same route is idempotent and does not enqueue a second item.
- [ ] `GET /queue` lists aggregate-safe queued message summaries grouped or filterable by route.
- [ ] `GET /queue/{id}` returns a safe single-message view.
- [ ] Claim/ack/release APIs or internal store methods exist and are tested so later workers can drain messages without losing them.
- [ ] Queue ordering tests prove priority ascending and FIFO within priority.
- [ ] Existing `/inbox`, `/message`, SSE, run-store, and workflow command tests remain compatible.
- [ ] Backend quality spec is updated with the mailbox/channel runtime contract if the API signatures are added or changed.

## Notes

- Kocoro's SQLite persistence and cloud replay semantics are the long-term parity target. This child should establish the local contract first so persistence can be added without changing clients.
