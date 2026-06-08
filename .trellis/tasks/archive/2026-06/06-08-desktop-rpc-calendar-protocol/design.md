# Desktop RPC calendar protocol design

## Boundary

This task changes only the Desktop RPC protocol package and tests under `internal/daemon/desktop_rpc`. It prepares the daemon side for Desktop-backed calendar tools without adding those tools yet.

## Existing Shape

StarClaw currently has:

- `Frame`
- `RPCRequest`
- `RPCResult`
- `RPCError`
- `DesktopEvent`
- `SystemCapabilitiesResult`
- `SystemPingParams`
- `SystemPingResult`
- `ProtocolMethods` containing system methods only

The broker/listener already handle request IDs, pending requests, timeouts, context cancellation, disconnect cleanup, and frame routing.

## New Protocol Surface

Add constants for:

- Calendar methods.
- Calendar-specific error codes.
- Permission status.
- Calendar account type.
- Attendee participation status.
- Event scope.
- Desktop event types.
- Recurrence frequency.

Add typed payload structs for representative calendar protocol bodies so tests and future tools do not need ad hoc map shapes.

## Compatibility

- Keep existing JSON field names stable.
- Keep `ProtocolVersion = "1.0.0"`.
- Keep `MaxFrameBodyBytes` unchanged.
- Keep `system.ping` and `system.capabilities` first in `ProtocolMethods`, followed by calendar methods in deterministic Kocoro-compatible order.

## Safety

The protocol package must not call system calendar APIs or read/write user calendar data. It only defines wire-level contracts and local broker behavior.

## Rollback

Rollback is limited to reverting changes in `internal/daemon/desktop_rpc/types.go` and the new/updated tests.
