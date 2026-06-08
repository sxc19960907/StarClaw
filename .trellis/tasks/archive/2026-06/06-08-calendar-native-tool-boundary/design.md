# Calendar native tool boundary design

## Boundary

This task adds daemon-side calendar tools under `internal/tools`. The tools are thin Desktop RPC clients. They validate model input, call the local Desktop RPC broker, and render structured Desktop errors into model-facing tool results.

The daemon must not access EventKit directly.

## Registration

Existing callers use `RegisterLocalTools(toolsConfig ...config.ToolsConfig)`. To avoid breaking them, add an optional registration option that carries a Desktop RPC broker. With no broker, calendar tools are not registered.

Expected shape:

- Preserve `RegisterLocalTools()` and `RegisterLocalTools(config.ToolsConfig{})`.
- Add an option/config path such as `WithDesktopRPCBroker(broker)` or equivalent.
- Wire daemon setup to pass `s.DesktopRPCBroker()`.

## Tool Behavior

Read-only, no approval:

- `calendar_check_permission`
- `calendar_list_sources`
- `calendar_list_events`
- `calendar_get_event`

Approval-required:

- `calendar_request_permission`
- `calendar_create_event`
- `calendar_update_event`
- `calendar_delete_event`

RPC behavior:

- Missing broker returns a tool error, not a Go error.
- RPC transport error returns a tool error.
- RPC `OK=false` maps known error codes to clear model-facing text.
- Successful RPC returns the raw Desktop JSON result.

## Validation

- List/create times must be RFC3339 with timezone.
- List limit clamps to `1..2000`, default `500`.
- Update rejects `scope=all`.
- Delete accepts `scope=all`.
- Mutation tools require `description` for approval context, but strip it from the Desktop RPC payload.
- `calendar_request_permission` uses a five-minute RPC timeout.

## Safety

- Tests use fake brokers only.
- No system calendar, network, keychain, or sync behavior.
- Calendar payloads use `internal/daemon/desktop_rpc` method constants and structs where practical.

## Rollback

Rollback is limited to removing the new calendar tools/tests and reverting registration wiring.
