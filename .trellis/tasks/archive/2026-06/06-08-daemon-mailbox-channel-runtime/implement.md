# Implementation Plan

## Scope

Add local daemon mailbox/channel runtime foundations: typed queue data, in-memory store lifecycle, local queue APIs, tests, and spec update.

## Checklist

- [x] Run `trellis-before-dev`.
- [x] Inspect current `/inbox`, `/message`, route registration, JSON helpers, metrics/redaction helpers, and Kocoro mailbox evidence.
- [x] Add mailbox store unit tests for enqueue, priority/FIFO ordering, capacity, dedup, defensive snapshots, and claim/ack/release.
- [x] Implement mailbox store and typed queued message model.
- [x] Add queue API handler tests for route registration, enqueue, validation, capacity, duplicate enqueue, list/detail, claim, ack, and release.
- [x] Implement queue API routes and server wiring.
- [x] Update backend quality spec with the queue API contract.
- [x] Run focused and broader validation.

## Completion Notes

- Added in-memory `MailboxStore` with per-route capacity, priority/FIFO ordering, per-route source/external id dedup, defensive snapshots, and claim/ack/release lifecycle.
- Added local daemon queue APIs:
  - `POST /queue`
  - `GET /queue`
  - `GET /queue/{id}`
  - `POST /queue/claim`
  - `POST /queue/{id}/ack`
  - `POST /queue/{id}/release`
- Wired the store into `Server` and registered queue routes without changing existing `/inbox`, `/message`, SSE, or workflow command behavior.
- Updated `.trellis/spec/backend/quality-guidelines.md` with the local daemon mailbox queue contract.

## Validation Commands

```bash
go test ./internal/daemon -run 'TestQueueAPI|TestMailboxStore|TestRouter'
go test ./internal/daemon
go test ./...
git diff --check
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-daemon-mailbox-channel-runtime
```

## Review Gates

- Existing inbox approval behavior remains compatible.
- Queue APIs are local-first and do not introduce external provider transport.
- Dedup and capacity behavior are deterministic.
- Claim/ack lifecycle cannot lose messages in the in-memory store.
- Public queue views are aggregate-safe and secret-safe.
