# Design

## Current State

StarClaw has a mailbox queue keyed by route, but no index from platform message IDs back to route keys and no connection-state cache. Kocoro uses these structures to reconcile replies/results and render connection health in sticky context.

## Proposed Architecture

Add daemon-local structures:

- `ReplyRouteIndex`
  - bounded FIFO map: `message_id -> route_key`
  - in-memory only
  - cap defaults to 256

- `ConnectionStateCache`
  - membership state by `platform:channel_id`
  - binding state by `platform`
  - transport state by `platform`
  - binding takes precedence over transport in platform rendering

Server wiring:

- `Server` owns `replyRoutes` and `connectionState`.
- `handleCreateQueueMessage` records `ExternalID -> RouteKey` after enqueue succeeds.
- New diagnostic endpoints:
  - `GET /channel/routes/{message_id}`
  - `GET /channel/state?platform=&channel_id=`

Scope is local only; no Slack/Feishu adapters or cloud WebSocket controller.

