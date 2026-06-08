# Delivery inject lifecycle depth

## Goal

Add Kocoro-style reply delivery result handling so failed outbound channel replies can be mapped back to the originating route and queued as trusted, route-scoped system events for the next turn.

This is Phase 7 child 4 under `Astria Kocoro parity phase 7: channel and cloud delivery parity`.

## Confirmed Facts

- Local Kocoro baseline is `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- Kocoro implements delivery result formatting and event enqueueing in `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/delivery_inject.go`.
- StarClaw already has `ReplyRouteIndex` and `SystemEventStore`.
- StarClaw does not yet have Kocoro's `ReplyDeliveryResultPayload` type, classification constants, or delivery result consumer.
- Real Cloud/WebSocket delivery result intake remains out of scope for this child.

## Requirements

- Add a local reply delivery result payload contract.
- Add permanent/transient delivery classification constants.
- Format permanent failures as clear non-delivery facts and include the re-add/re-authorize implication.
- Format transient failures cautiously and avoid claiming permanent removal.
- Bind delivery failures to routes through `ReplyRouteIndex`.
- Enqueue one trusted `agent.SystemEvent` into `SystemEventStore` when a failed delivery has a known route.
- Keep success delivery results silent.
- Drop unknown-message delivery failures without panicking or leaking globally.
- Keep the implementation local-first; do not add real external transport, credentials, or Cloud WebSocket wiring.

## Acceptance Criteria

- [ ] Unit tests cover permanent and transient formatting.
- [ ] Unit tests cover Slack-style thread channel label rendering.
- [ ] Tests prove success results do not enqueue events.
- [ ] Tests prove failed results enqueue only for known message ids.
- [ ] Tests prove enqueued events are trusted and route-scoped.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Out of Scope

- Real Feishu/Lark, Slack, Telegram, LINE, or webhook transport.
- WebSocket cloud-controller lifecycle.
- External adapter management APIs.
- UI changes.

## Evidence

- `.trellis/research/kocoro-local-comparison-phase7-plan.md`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/delivery_inject.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/delivery_inject_test.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/types.go`
