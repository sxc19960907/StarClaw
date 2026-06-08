# Kocoro local comparison and Phase 8 plan

Date: 2026-06-08

## Local Kocoro baseline

- Path: `/Users/timmy/PycharmProjects/Kocoro`
- Commit: `74cdb3c`
- Remote: `https://github.com/Kocoro-lab/Kocoro.git`
- Use this checkout for parity checks unless a deliberate refresh is requested.

## Phase 8 status

Parent:

`Astria Kocoro parity phase 8: local API streaming and runtime compatibility`

Completed children:

1. `openai-compatible-streaming-deltas`
   - Added OpenAI-compatible SSE chunks for `POST /v1/chat/completions` with `stream=true`.
   - Evidence: `internal/daemon/openai_api.go`, `internal/daemon/openai_api_test.go`.
2. `ws-controller-cloud-lifecycle`
   - Added a local-first cloud lifecycle controller boundary and `/cloud/lifecycle` API.
   - Evidence: `internal/daemon/cloud_lifecycle.go`, `internal/daemon/cloud_lifecycle_api.go`, `internal/daemon/cloud_lifecycle_api_test.go`.

## Current StarClaw coverage

StarClaw has now closed the most visible API compatibility complaint:

- `POST /v1/chat/completions` accepts `stream=true`.
- Successful streams emit OpenAI-style `chat.completion.chunk` frames, a role chunk, content chunks, terminal stop chunk, and `data: [DONE]`.
- `RunAgentRequest.EnableStreaming` routes compatible runs through the agent streaming path.
- Existing `/message` SSE, `/events`, `/metrics`, `/runs/{id}/trace`, and `/runs/{id}/control` cover the Phase 5 platform control/observability surface.

## Remaining Kocoro evidence

Kocoro has more mature streaming and local runtime compatibility than the current StarClaw slice:

- Kocoro client SSE reconnect and idle watchdog:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/client/sse.go`
  - `StreamSSEWithOptions`, `Last-Event-ID`, idle timeout, reconnect budget.
- Kocoro runner treats streaming idle timeouts as soft partial-output failures:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/runner.go`
  - `loop.SetEnableStreaming(true)` around the runner setup.
  - `isSoftRunError` includes `client.ErrStreamIdleTimeout`.
- Kocoro daemon SSE has richer request-level events:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/server.go`
  - `sseEventHandler.OnStreamDelta`, `session_started`, approvals, usage, run status, and reconnection-oriented event replay through `GET /events`.
- Kocoro Desktop/local integration surface remains deeper:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/desktop_rpc/*`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/calendar_*.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/browser_handoff.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/browser_lease.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty_*.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/generate_image.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/edit_image.go`

## Gap summary after the first two Phase 8 children

Closed or mostly closed:

- OpenAI-compatible local streaming response basics.
- Local cloud lifecycle control boundary.
- Runtime pause/resume/cancel/replay API from Phase 5.
- Structured events, metrics, and trace export from Phase 5.
- Channel route, connection state, system events, and delivery injection from Phase 7.

Still open:

1. Streaming edge semantics
   - Error frames after SSE headers are written are not yet covered by tests.
   - Partial-output behavior on upstream stream failures is not yet pinned for the OpenAI-compatible endpoint.
   - Final event consistency is not yet specified for error vs success streams.

2. SSE reconnect/replay client compatibility
   - StarClaw has basic SSE parsing and `/events`, but lacks Kocoro-style reconnect options with `Last-Event-ID`, idle watchdog, and bounded retry semantics.

3. OpenAI-compatible tool-call streaming parity
   - StarClaw intentionally rejects OpenAI `tools`, `functions`, `tool_choice`, and `response_format`.
   - This is acceptable for local daemon tool execution, but still a compatibility gap for clients expecting OpenAI tool-call deltas.

4. Native tool and Desktop RPC depth
   - StarClaw has a basic Desktop RPC boundary, browser/computer/accessibility tools, and MCP support.
   - Kocoro has deeper native Desktop RPC usage, calendar tools, browser handoff/leasing, Ghostty helpers, image generation/editing tools, and more exhaustive guard tests.

5. Secret/keychain and sync depth
   - Kocoro has `internal/keychain`, `internal/sync`, and richer upload/share/migration modules.
   - StarClaw should not enable real cloud sync or credential storage until privacy and credential behavior are explicitly approved.

## Recommended next task order

### 3. `streaming-runtime-edge-cases`

Goal: close the runtime edge cases behind the newly added OpenAI-compatible streaming endpoint.

Scope:

- Pin success stream terminal behavior: role chunk, content chunks, stop chunk, `[DONE]`.
- Pin error behavior after SSE headers are written.
- Ensure fallback `OnText` emits one content chunk when no streaming deltas arrive.
- Ensure final full text does not duplicate content when deltas already streamed.
- Record run completion/errors consistently in `RunStore`.

Why next: the user noticed missing streaming support, and the first implementation closes the happy path. Before expanding compatibility, the runtime contract should be robust under failure and fallback.

### 4. `sse-reconnect-idle-watchdog`

Goal: bring StarClaw client/server SSE behavior closer to Kocoro's reconnectable stream contract.

Scope:

- Add client options for idle timeout, bounded reconnect, backoff, and `Last-Event-ID`.
- Ensure server `/events` supports replay from event IDs if not already fully covered.
- Treat clean EOF, terminal done, idle timeout, and context cancellation distinctly.
- Add tests mirroring Kocoro's `internal/client/sse_test.go` scenarios where applicable.

Why after streaming edge cases: reconnect behavior should build on a well-specified stream terminal/error contract.

### 5. `native-tool-parity-followup`

Goal: recompare Kocoro native tool depth after the API streaming work is stable and create the next parent phase.

Scope:

- Compare Desktop RPC, calendar tools, browser handoff/leasing, Ghostty helpers, image generation/editing, keychain, sync, and migration modules.
- Separate local-safe work from credentialed/off-machine features.
- Produce Phase 9 plan with child tasks.

Why last in Phase 8: native tool parity is larger than the local API streaming theme. It should be planned from a fresh evidence pass, not mixed into the OpenAI-compatible endpoint fixes.

## Recommended Phase 9 direction

Proposed parent:

`Astria Kocoro parity phase 9: native desktop tools and local integration depth`

Likely children:

1. `desktop-rpc-request-depth`
   - Expand Unix socket Desktop RPC beyond ping/capabilities into typed requests, timeout/cancel cleanup, and fake Desktop harness coverage.
2. `calendar-native-tool-boundary`
   - Add local Desktop-RPC-backed calendar tool contracts, permission checks, and no-broker disabled behavior.
3. `browser-handoff-lease-depth`
   - Add browser handoff/leasing semantics and stronger visual/browser state guard tests.
4. `terminal-workspace-tool-depth`
   - Add Ghostty/workspace helper boundaries where locally available, with platform fallbacks.
5. `image-tool-local-boundaries`
   - Add image generation/editing tool contracts behind explicit provider config.
6. `keychain-sync-migration-planning`
   - Plan secrets, sync, upload, and migration features separately, because they affect privacy and credential handling.

## Gap estimate

After Phase 8's first two children, StarClaw is roughly:

- 75-80% aligned on local daemon/API/runtime control foundations.
- 55-65% aligned on channel/cloud delivery foundations.
- 35-45% aligned on Kocoro's native Desktop/tool depth.
- 25-35% aligned on sync/keychain/migration/cloud-product depth.

The next highest-leverage work is still Phase 8 child 3: `streaming-runtime-edge-cases`.
