# Kocoro native tool parity comparison and Phase 9 plan

Date: 2026-06-08

## Local Kocoro baseline

- Path: `/Users/timmy/PycharmProjects/Kocoro`
- Commit: `74cdb3c`
- Remote: `https://github.com/Kocoro-lab/Kocoro.git`
- Use this checkout for parity checks unless the user explicitly asks to refresh from GitHub.

## Current StarClaw baseline

Phase 8 closed the local API streaming/runtime compatibility slice:

- OpenAI-compatible streaming deltas for `POST /v1/chat/completions`.
- Streaming runtime edge cases around final text, partial output, and error frames.
- SSE reconnect, idle watchdog, and `Last-Event-ID` replay for local event streams.
- Local-first cloud lifecycle boundary without real cloud credentials or transport.

The remaining Kocoro gap has shifted from "can local clients stream?" to "how deep is the native desktop/tool platform?"

## Evidence summary

### Desktop RPC

Kocoro:

- `internal/daemon/desktop_rpc/types.go`
- `internal/daemon/desktop_rpc/broker.go`
- `internal/tools/calendar_*.go`

Kocoro's Desktop RPC protocol includes:

- `system.ping`
- `system.capabilities`
- `calendar.list_sources`
- `calendar.list_events`
- `calendar.get_event`
- `calendar.create_event`
- `calendar.update_event`
- `calendar.delete_event`
- `calendar.check_permission`
- `calendar.request_permission`

StarClaw:

- `internal/daemon/desktop_rpc/types.go`
- `internal/daemon/desktop_rpc/broker.go`
- `internal/daemon/desktop_rpc/listener.go`

StarClaw has the socket, codec, listener, broker, pending request cleanup, and daemon status plumbing, but the protocol method set is currently limited to:

- `system.ping`
- `system.capabilities`

Gap: medium-high. The transport exists, but the first real Desktop-backed domain protocol and user-facing tools are not implemented.

### Calendar tools

Kocoro:

- `internal/tools/calendar_common.go`
- `internal/tools/calendar_list_sources.go`
- `internal/tools/calendar_list_events.go`
- `internal/tools/calendar_get_event.go`
- `internal/tools/calendar_create_event.go`
- `internal/tools/calendar_update_event.go`
- `internal/tools/calendar_delete_event.go`
- `internal/tools/calendar_check_permission.go`
- `internal/tools/calendar_request_permission.go`
- `internal/tools/calendar_tools_test.go`

Kocoro calendar tools are Desktop-RPC-backed. They validate RFC3339 times, clamp list limits, map Desktop permission/error codes to model-facing messages, and keep permission request latency separate from normal RPC timeouts.

StarClaw:

- No `internal/tools/calendar_*.go` files.
- Existing `internal/tools/schedule*.go` is a local schedule/reminder manager, not system calendar integration.

Gap: high. This is the most direct native feature gap after Phase 8.

### Browser handoff and lease behavior

Kocoro:

- `internal/tools/browser_handoff.go`
- `internal/tools/browser_lease.go`
- `internal/tools/browser_lease_test.go`
- `internal/tools/browser_handoff_test.go`

Kocoro tracks browser ownership per run, supports reload handoff for deprecated browser instances, and defers cleanup while active browser leases still exist.

StarClaw:

- `internal/tools/browser.go`
- `internal/tools/browser_test.go`

StarClaw has a browser tool and tests, but no separate handoff/lease modules. Browser lifetime is therefore less protected across concurrent runs or daemon reload-style registry replacement.

Gap: medium. The browser capability exists, but runtime ownership semantics are shallower.

### Terminal workspace / Ghostty tools

Kocoro:

- `internal/tools/ghostty.go`
- `internal/tools/ghostty_darwin.go`
- `internal/tools/ghostty_stub.go`
- `internal/tools/ghostty_test.go`

Kocoro exposes a macOS-only Ghostty tool for visible terminal tabs/splits, tracked tab titles, sending input, listing tabs, and platform fallback messaging.

StarClaw:

- No Ghostty tool.
- Has `bash`, `applescript`, `process`, and `computer`, but no first-class visible-terminal workspace helper.

Gap: medium. This is useful for desktop workflow parity, but lower priority than calendar because it can fall back to existing tools.

### Image generation and editing

Kocoro:

- `internal/images/client.go`
- `internal/tools/generate_image.go`
- `internal/tools/edit_image.go`
- `internal/tools/generate_image_test.go`
- `internal/tools/edit_image_test.go`

Kocoro exposes paid/cloud image generation and edit tools with explicit approval, provider error classification, prompt and enum validation, and permanent public CDN warnings.

StarClaw:

- `internal/tools/imaging.go`
- `internal/tools/imaging_compress.go` is absent in this checkout; StarClaw has local image processing tests through `imaging_test.go`.
- Existing publish/list/retract web tools exist, but no `generate_image` or `edit_image` provider-backed tools.

Gap: medium-high, but credentialed/provider-backed behavior is intentionally sensitive. It should remain disabled by default and require explicit provider configuration.

### Keychain, sync, and migration depth

Kocoro:

- `internal/keychain/*`
- `internal/sync/*`
- `internal/migrate/claudecode/*`
- `internal/config/migrate.go`
- `internal/daemon/migrate_claudecode.go`

StarClaw:

- No `internal/keychain`, `internal/sync`, or `internal/migrate` package.
- Prior phases added local-first channel/cloud lifecycle boundaries, but not real off-machine sync, OS keychain storage, or migration flows.

Gap: high. This is platform-foundation work, but it touches privacy, credentials, and user data movement. It should be planned separately before implementation.

### Other Kocoro modules worth deferring

Kocoro also has:

- `internal/agenttypes/*`
- deeper `internal/memory/*`
- `internal/uploads/*`
- richer share/upload cloud paths

StarClaw has memory daemon APIs and memory tools, but not Kocoro's full memory package shape. These should not be mixed into the native desktop tool phase unless a child task specifically needs them.

## Recommended Phase 9 parent

`Astria Kocoro parity phase 9: native desktop tools and local integration depth`

Goal: close the highest-value local native integration gaps left after API streaming parity, while preserving StarClaw's local-first defaults and avoiding unapproved real cloud credentials, sync, or off-machine telemetry.

## Recommended Phase 9 child tasks

1. `desktop-rpc-calendar-protocol`
   - Expand Desktop RPC constants and protocol types from system-only to calendar v1.
   - Add shared request/result/error helpers, fake Desktop broker coverage, timeout/cancel/disconnect behavior, and capability negotiation tests.
   - Keep transport local over Unix socket only.

2. `calendar-native-tool-boundary`
   - Add `calendar_check_permission`, `calendar_request_permission`, `calendar_list_sources`, `calendar_list_events`, `calendar_get_event`, `calendar_create_event`, `calendar_update_event`, and `calendar_delete_event`.
   - Register tools only when the daemon has a Desktop RPC broker.
   - Preserve no-broker disabled behavior and friendly permission/error messages.
   - Do not access EventKit directly from the daemon.

3. `browser-handoff-lease-depth`
   - Add per-run browser lease tracking and reload handoff cleanup semantics.
   - Pin no-kill-while-in-use behavior with tests.
   - Keep existing browser tool contract stable for model callers.

4. `terminal-workspace-tool-depth`
   - Add a local Ghostty-compatible terminal workspace tool or an equivalent StarClaw-named terminal workspace boundary.
   - Include macOS implementation, non-darwin stubs, version/availability checks, visible terminal approval behavior, and tests around argument validation and fallback messages.

5. `image-tool-provider-boundary`
   - Add `generate_image` and `edit_image` tool contracts behind explicit provider configuration.
   - Reuse local publish/image-processing boundaries where possible.
   - Require approval, classify provider errors, validate prompts/enums/client-side limits, and keep permanent public URL warnings.
   - Do not enable Shannon/Kocoro cloud credentials by default.

6. `keychain-sync-migration-discovery`
   - Produce the implementation plan for OS keychain storage, local sync markers/batching, Claude Code migration, upload/share privacy boundaries, and opt-in controls.
   - This should remain a planning/discovery child unless the user explicitly approves credentialed or off-machine behavior.

## Suggested execution order

1. `desktop-rpc-calendar-protocol`
2. `calendar-native-tool-boundary`
3. `browser-handoff-lease-depth`
4. `terminal-workspace-tool-depth`
5. `image-tool-provider-boundary`
6. `keychain-sync-migration-discovery`

Rationale: calendar requires Desktop RPC protocol depth, and browser lifetime management reduces risk before more desktop automation tools are added. Terminal and image tools are valuable but can be added after the core Desktop-backed integration is stable. Keychain/sync/migration should be separated because they introduce higher privacy and credential risk.

## Updated gap estimate after Phase 8

- Local daemon/API/runtime control foundations: 85-90% aligned.
- OpenAI-compatible local streaming: 75-85% aligned, with tool/function-call streaming still intentionally deferred.
- Channel/cloud lifecycle foundations: 60-70% aligned, but real delivery and credentials remain disabled.
- Native Desktop/tool depth: 35-45% aligned.
- Calendar/system integration: 10-20% aligned.
- Browser runtime ownership depth: 45-55% aligned.
- Terminal workspace depth: 20-30% aligned.
- Image generation/editing depth: 20-30% aligned.
- Keychain/sync/migration/cloud-product depth: 20-30% aligned.

Overall, StarClaw has largely closed the Kocoro platform-control and local-streaming gap, but remains materially behind on native desktop integration depth. Phase 9 should focus there before returning to UI polish.
