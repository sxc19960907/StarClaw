# Desktop RPC calendar protocol

## Goal

Expand StarClaw's Desktop RPC protocol foundation from system-only methods to the calendar v1 protocol shape needed by the next `calendar-native-tool-boundary` task.

## Confirmed Facts

- Kocoro baseline `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c` exposes Desktop RPC calendar methods and structured calendar error/permission enums.
- StarClaw already has `internal/daemon/desktop_rpc` socket codec, listener, broker, request/result envelopes, and tests.
- StarClaw's current `ProtocolMethods` only includes `system.ping` and `system.capabilities`.
- Calendar tools are intentionally out of this task; they should be added only after protocol foundations are test-covered.

## Requirements

- Add Desktop RPC calendar method constants for list sources/events, get/create/update/delete event, check permission, and request permission.
- Add calendar permission, account type, attendee status, event scope, recurrence frequency, and desktop event constants compatible with the Kocoro v1 contract.
- Add or preserve structured error code constants needed by calendar callers: permission denied, permission not determined, not found, invalid argument, read-only calendar, internal error, timeout, and desktop disconnected.
- Add calendar protocol payload structs where useful for typed tests and future tools.
- Preserve existing system RPC behavior, codec framing, listener behavior, broker timeout/cancel/disconnect behavior, and local Unix socket transport.
- Do not implement calendar tools or direct EventKit access in this task.

## Acceptance Criteria

- [x] `desktop_rpc.ProtocolMethods` includes system methods and all calendar v1 methods in a deterministic order.
- [x] Tests cover the expected method list, calendar enum/error constants, and representative calendar payload JSON tags.
- [x] Existing Desktop RPC broker/codec/listener tests still pass.
- [x] No real system calendar, cloud, credential, or sync behavior is introduced.

## Out of Scope

- Calendar tool registration.
- Desktop app implementation.
- EventKit access from the daemon.
- Credentialed/cloud-backed behavior.
