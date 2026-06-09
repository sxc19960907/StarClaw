# Desktop RPC event monitoring design

## Current Shape

`internal/daemon/desktop_rpc` already defines:

- `FrameDesktopEvent = "desktop_event"`
- `DesktopEvent{event,data,ts}`
- `ListenerConfig.EventSink`
- `Listener.handleIncomingEvent`, which decodes event payloads and calls the
  sink when configured.

The missing piece is daemon ownership. StarClaw currently starts the listener
without a daemon event sink, so incoming events are ignored after decode.

## Proposed Shape

Add a small daemon-owned event monitor:

- `DesktopEventMonitor` stores a bounded recent-event ring or slice.
- `Record(*desktop_rpc.DesktopEvent)` normalizes missing timestamps and stores
  only local in-memory events.
- `Status()` returns safe aggregate metadata:
  - retained count;
  - last event type;
  - last event timestamp.

Wire `Server` to own the monitor and pass `EventSink` to the Desktop RPC
listener at daemon startup.

## Status Contract

Extend `desktop_rpc.Status` with an optional nested event metadata field. It
must not include raw event `data`, socket path, pidfile path, or user content.

Example:

```json
{
  "desktop_rpc": {
    "listening": true,
    "connected": true,
    "pending": 0,
    "events": {
      "retained": 3,
      "last_event": "desktop_online",
      "last_ts": "2026-06-09T..."
    }
  }
}
```

## Compatibility

- Existing `listening`, `connected`, and `pending` fields remain unchanged.
- The status extension is additive.
- Desktop clients that never send events behave exactly as before with empty
  event metadata.

## Test Strategy

- Add daemon monitor unit tests for empty status, bounded retention, missing
  timestamp normalization, and raw payload omission.
- Add listener test proving `desktop_event` frames invoke `EventSink`.
- Add daemon status test proving event metadata is present and redacted.
