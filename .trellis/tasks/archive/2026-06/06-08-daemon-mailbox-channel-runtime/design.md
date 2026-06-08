# Design

## Current StarClaw State

StarClaw has a guarded `InboxStore` and `/inbox` APIs. Webhooks create pending inbox items, and approval/retry directly run an agent through `s.runAgent`. There is no route-key mailbox, priority queue, claim/ack lifecycle, or mailbox endpoint yet.

The previous Phase6 child added Kocoro-style `/research` and `/swarm` workflow commands over existing `POST /message`; this child should preserve that path.

## Kocoro Evidence

Kocoro's open daemon includes:

- In-memory priority/FIFO mailbox in `internal/agenttypes/mailbox.go`.
- SQLite mailbox persistence in `internal/daemon/mailbox_store.go`.
- HTTP enqueue path in `internal/daemon/queue_enqueue_handler.go`.
- Channel state, reply route, and delivery inject files under `internal/daemon/channel_state_*`, `reply_route_index.go`, and `delivery_inject.go`.

The Kocoro durability rule is important: persist before in-memory enqueue/ack, and mark consumed only after the message is safely saved to the session. StarClaw will first implement the local API/store contract, then persistence in a later slice if needed.

## Proposed Architecture

Add StarClaw daemon mailbox runtime:

- `internal/daemon/mailbox.go`
  - `QueuedMessage`
  - `QueuedMessageStatus`
  - `MailboxStore`
  - enqueue/snapshot/get/claim/ack/release methods

- `internal/daemon/queue_api.go`
  - `POST /queue`
  - `GET /queue`
  - `GET /queue/{id}`
  - claim/ack/release endpoints if the final implementation chooses API-level lifecycle in this child

- `internal/daemon/router.go`
  - Add queue routes under a `registerQueueRoutes` module.

- `internal/daemon/server.go`
  - Add `mailboxStore *MailboxStore`.

## Queue Model

Queued message fields:

- `id`
- `route_key`
- `session_id`
- `source`
- `external_id`
- `sender`
- `agent`
- `text`
- `priority`
- `status`: queued, claimed, acknowledged, released, failed
- `claim_id`
- `attempt`
- `enqueued_at`
- `updated_at`
- `metadata`

Safe summary views may include text previews, but must not include provider payloads, secrets, or hidden credentials. Metadata must be sanitized like inbox metadata.

## Ordering

Available queued messages drain by:

1. lower numeric priority first
2. older `enqueued_at` first
3. stable insertion order as final tie breaker

Claimed messages are not returned to new claimers until released or acknowledged.

## Compatibility

- Existing `/inbox` remains intact.
- Existing `/message`, workflow commands, SSE, run-store, metrics, and control APIs remain unchanged.
- No cloud transport is introduced.

## Future Extension

Later children can:

- persist queue rows before enqueue
- wire inbox approvals into mailbox injection
- drain mailbox messages into active sessions
- add channel state notices and reply route indexes
- add external provider transports only after explicit user approval
