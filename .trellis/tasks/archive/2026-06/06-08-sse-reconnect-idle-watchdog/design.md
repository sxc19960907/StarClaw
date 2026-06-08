# SSE reconnect idle watchdog design

## Current Shape

`internal/client/sse.go` has `SSEClient.Connect(ctx, url)` and a private reconnect loop. It reconnects forever on errors with exponential backoff, but it does not:

- expose reconnect bounds to callers,
- send `Last-Event-ID` on reconnect,
- detect silent streams that neither emit lines nor close,
- report or classify terminal conditions.

`internal/daemon/server.go` exposes `/events` through the daemon `EventBus`. The next implementation should verify whether event replay from event IDs already exists and add it if needed.

## Client Contract

Add an options API without breaking the existing API:

```go
type SSEConnectOptions struct {
    IdleTimeout time.Duration
    MaxReconnects int
    ReconnectBackoffBase time.Duration
}

func (c *SSEClient) ConnectWithOptions(ctx context.Context, url string, opts SSEConnectOptions) (<-chan SSEEvent, error)
```

`Connect` delegates to `ConnectWithOptions` with legacy defaults. Legacy defaults should preserve automatic reconnect behavior for existing callers, but idle watchdog should remain opt-in.

When a delivered event has a non-empty ID, the client stores it. Reconnect attempts include `Last-Event-ID: <last id>`.

The scanner runs in a goroutine and the read loop uses a timer to detect idle streams. Heartbeat comments reset the idle timer because they prove the connection is alive.

## Server Contract

`GET /events` should accept a replay cursor from:

- `?last_event_id=<id>`
- `Last-Event-ID: <id>`

It should replay events with IDs greater than the cursor and then continue live streaming. If StarClaw already supports this, this task should add focused tests and preserve behavior.

## Terminal Semantics

- `done` event: clean completion.
- Clean EOF: close normally.
- Context cancellation: close promptly.
- Idle timeout/read/connect error: reconnect while the budget remains.
- Reconnect budget exhausted: close the channel.

The public channel API cannot return the final error without changing existing caller contracts, so tests should verify behavior through connection counts and channel close timing.

## Risk

The biggest compatibility risk is changing legacy reconnect behavior. To avoid that, `Connect` keeps unbounded reconnect on connect/read errors and opts out of idle timeout unless callers use options.
