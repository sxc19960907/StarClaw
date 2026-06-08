# Kocoro local comparison and Phase 7 plan

Date: 2026-06-08

## Local Kocoro baseline

- Path: `/Users/timmy/PycharmProjects/Kocoro`
- Commit: `74cdb3c`
- Use this checkout for parity checks before browsing GitHub again.

## Phase 7 status

Parent:

`Astria Kocoro parity phase 7: channel and cloud delivery parity`

Completed children:

1. `cloudflow-dispatch-contract`
   - Added local-first cloudflow dispatch boundary and slash workflow contract.
   - Archived.
2. `channel-route-index-connection-state`
   - Added bounded reply route index, connection state cache, server wiring, and read-only channel diagnostics.
   - Archived.

## Remaining Kocoro evidence

Kocoro still has daemon/runtime files that StarClaw does not yet mirror:

- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/system_event_store.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/suggestion_handler.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/delivery_inject.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/ws_controller.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/feishu_handler.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/system_event.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/suggestion.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/suggestion_state.go`

StarClaw currently has the route/state foundation but not Kocoro's next-turn system event injection, suggestion state APIs, delivery failure injection, WebSocket cloud-controller lifecycle, or external channel management endpoints.

## Recommended remaining Phase 7 child order

### 3. `system-event-store-suggestions`

Goal: add durable-enough local system event and suggestion foundations that can feed next-turn context and UI diagnostics without storing message content in broad observability channels.

Scope:

- Add an in-memory, per-route `SystemEventStore` with bounded FIFO, drain, forget, and consecutive context-key collapse.
- Add agent-facing system event data types and formatting hooks for trusted daemon-authored events.
- Add suggestion state storage and daemon GET/accept endpoints for session suggestions.
- Keep diagnostics content-minimized and route-scoped.
- Add unit and handler tests.

Why next: Kocoro's delivery lifecycle and channel state consumers depend on a route-scoped event store. Doing delivery injection first would duplicate or weaken that boundary.

### 4. `delivery-inject-lifecycle-depth`

Goal: deepen queue/mailbox delivery behavior around failed replies, busy routes, orphan replies, and re-enqueue flows.

Scope:

- Add Kocoro-style delivery result payload handling.
- Bind failed delivery results back to route keys through `ReplyRouteIndex`.
- Inject trusted, route-scoped failure events into `SystemEventStore`.
- Cover permanent vs transient delivery failure wording.
- Add tests for orphan replies, route miss, route hit, and bounded event behavior.

Why after system events: delivery injection is one producer of next-turn system events.

### 5. `external-channel-adapter-boundaries`

Goal: define local-first external channel adapter interfaces and management contracts without enabling real off-machine transport by default.

Scope:

- Add adapter interfaces for Feishu/Lark, Slack, Telegram, and generic webhook/channel bindings.
- Add fake adapters and tests for install/list/delete style management.
- Gate any real cloud/channel passthrough behind explicit config and credential checks.
- Document privacy and credential boundaries.

Why last in this phase: real channel transports require credentials and off-machine delivery. The code should first have local route, event, and delivery lifecycle boundaries.

## Deferred beyond Phase 7

- `ws-controller-cloud-lifecycle`: Kocoro's `WSController` maps to signed-in cloud runtime lifecycle. StarClaw should only implement this once cloud auth/sign-in semantics are explicit.
- `openai-compatible-streaming-tool-deltas`: StarClaw's OpenAI-compatible endpoint should later support streaming response deltas/tool-call deltas if local API parity becomes a product priority. This is important, but not the next Kocoro daemon parity gap.
- `astria-stellar-workbench-ui-language`: UI polish remains after core platform parity work.

## Current gap summary

Against local Kocoro commit `74cdb3c`, Phase 7 has closed roughly the first 40% of channel/cloud delivery parity:

- Closed: cloudflow dispatch contract, route index, connection state diagnostics.
- Still open: system event store, suggestions, delivery injection lifecycle, adapter boundaries, and cloud WebSocket lifecycle.

The next task should be `system-event-store-suggestions`.
