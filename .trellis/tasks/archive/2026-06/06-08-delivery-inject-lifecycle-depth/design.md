# Design

## Boundary

This task adds the local daemon-side consumer for reply delivery result events. It does not add the Cloud/WebSocket producer.

Data flow:

`delivery result payload -> ReplyRouteIndex lookup -> SystemEventStore enqueue -> future next-turn injection`

## Payload Contract

Add `ReplyDeliveryResultPayload` to `internal/daemon/types.go`:

- `ok`
- `channel`
- `thread_id`
- `platform_msg_id`
- `error`
- `reason`
- `class`

Add constants:

- `ClassPermanent = "permanent"`
- `ClassTransient = "transient"`

## Formatting

Add `internal/daemon/delivery_inject.go`:

- `channelLabel` renders `channel + " " + channel-head` when `thread_id` has Slack-style `<channel>-<ts>` shape.
- `formatDeliveryFailure`:
  - defaults empty reason to `"delivery failed"`.
  - permanent: says `FAILED`, says the user did not see it, and says the bot will not receive/deliver until re-added/re-authorized.
  - transient: says it may not have been delivered and a retry is in progress.

## Consumer

`newDeliveryResultConsumer(store, idx)` returns `func(ReplyDeliveryResultPayload, string)`:

- `OK=true`: no-op.
- missing route index hit: no-op.
- failure with route: enqueue trusted `agent.SystemEvent` with:
  - formatted text
  - context key `delivery-fail:<channel>:<thread_id>`
  - current timestamp

`HandleReplyDeliveryResult` exposes a single-call entry point for future command/client wiring.

## Trade-offs

- Unknown message ids are dropped for now. Later channel-state reconciliation can surface broader binding issues without risking cross-route leakage.
- This task deliberately does not wire a daemon API endpoint for synthetic delivery results; tests call the consumer directly.
