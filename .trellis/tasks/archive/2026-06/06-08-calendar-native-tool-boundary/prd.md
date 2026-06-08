# Calendar native tool boundary

## Goal

Add StarClaw's Desktop-RPC-backed calendar tool boundary so agents can call calendar tools through the local Desktop broker without direct EventKit access or cloud behavior.

## Confirmed Facts

- `desktop-rpc-calendar-protocol` added calendar v1 method constants, error/permission enums, and payload structs.
- Kocoro's calendar tools live under `/Users/timmy/PycharmProjects/Kocoro/internal/tools/calendar_*.go` and route through Desktop RPC.
- StarClaw has `internal/tools/register.go` for local tool registration and `internal/daemon/server.go` exposes `DesktopRPCBroker()`.
- StarClaw currently has local schedule tools, but no system calendar tools.

## Requirements

- Add Desktop-RPC-backed tools:
  - `calendar_check_permission`
  - `calendar_request_permission`
  - `calendar_list_sources`
  - `calendar_list_events`
  - `calendar_get_event`
  - `calendar_create_event`
  - `calendar_update_event`
  - `calendar_delete_event`
- Add shared Desktop RPC call/error mapping helpers for calendar tools.
- Validate required arguments and RFC3339 timestamps before RPC.
- Preserve Kocoro-compatible scope rules: update supports `this` and `this_and_future`; delete supports `this`, `this_and_future`, and `all`.
- Strip model-facing approval `description` before sending mutation RPC params to Desktop.
- Register calendar tools only when a Desktop RPC broker is available; default local CLI registry without broker must not expose broken calendar tools.
- Do not call EventKit, cloud, sync, keychain, or external transport from the daemon.

## Acceptance Criteria

- [x] Calendar tools are implemented with focused unit tests for no-broker, validation, method routing, timeout, error mapping, and description stripping.
- [x] The daemon local tool registry wires calendar tools when a Desktop RPC broker is available.
- [x] Existing no-arg `RegisterLocalTools()` callers continue to work and do not expose calendar tools without a broker.
- [x] `go test ./internal/tools ./internal/daemon` passes.
- [x] No real calendar data access occurs in tests or implementation.

## Out of Scope

- Desktop app implementation.
- EventKit integration.
- Calendar UI.
- Cloud-backed calendar providers.
