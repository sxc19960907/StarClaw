# Kocoro local comparison and Phase 6 plan

Date: 2026-06-08

## Local Kocoro checkout

- Path: `/Users/timmy/PycharmProjects/Kocoro`
- Remote: `https://github.com/Kocoro-lab/Kocoro.git`
- Checked commit: `74cdb3c Merge pull request #233 from Kocoro-lab/fix/mailbox-consume-after-save`

This local checkout should be used for future Kocoro parity comparisons instead of re-fetching GitHub every turn. Refresh it deliberately with `git -C /Users/timmy/PycharmProjects/Kocoro pull --ff-only` when current upstream state matters.

## Kocoro positioning

Kocoro's README describes the open-source repo as the `shan` engine and daemon. The closed-source Kocoro Desktop app runs on top of the daemon. The engine focus is not just UI polish: it combines local Mac tools, daemon HTTP/SSE APIs, MCP, channel messaging, scheduling, named agents, memory, and Shannon Cloud workflows.

Key README evidence:

- Local Mac agent with full computer access and Slack / LINE / Feishu / Telegram via Shannon Cloud: `/Users/timmy/PycharmProjects/Kocoro/README.md:13`.
- Open-source scope is engine + daemon; Desktop is separate closed-source product: `/Users/timmy/PycharmProjects/Kocoro/README.md:15`.
- `/research` and `/swarm` commands exist and are accepted through daemon `POST /message` with SSE: `/Users/timmy/PycharmProjects/Kocoro/README.md:203`, `/Users/timmy/PycharmProjects/Kocoro/README.md:227`.
- Channel daemon architecture and behavior are documented at `/Users/timmy/PycharmProjects/Kocoro/README.md:534` and `/Users/timmy/PycharmProjects/Kocoro/README.md:553`.
- Memory sidecar / episodic preflight / session sync are documented at `/Users/timmy/PycharmProjects/Kocoro/README.md:653` and `/Users/timmy/PycharmProjects/Kocoro/README.md:692`.

## Module-level gaps

Kocoro has internal modules that StarClaw does not currently have:

| Kocoro module | StarClaw status | Planning implication |
|---|---|---|
| `internal/agenttypes` | missing | Needed for durable queued messages, mailbox lifecycle, and typed cancellation/message metadata. |
| `internal/cloudflow` | missing | Needed for `/research` and `/swarm` workflow orchestration over Gateway/SSE. |
| `internal/images` | missing | Needed for image generation/editing parity if product scope requires it. |
| `internal/keychain` | missing | Needed for safer secret storage before any cloud/channel credential feature. |
| `internal/memory` | missing | Needed for Kocoro-style episodic memory sidecar and bundle lifecycle. |
| `internal/migrate` | missing | Needed for Claude Code import parity, lower priority for runtime parity. |
| `internal/share` | missing | Needed for rendered share pages, artifact publishing, and retractable delivery flows. |
| `internal/sync` | missing | Needed for opt-in session sync / dry-run / marker / uploader lifecycle. |

## Runtime/daemon gaps

Kocoro's daemon has significantly deeper channel/runtime infrastructure:

- Mailbox and durable message queue: `internal/daemon/mailbox_store.go`, `router_mailbox.go`, `queue_*`, and `agenttypes/mailbox.go`.
- Channel state, routing, and lifecycle: `channel_state_*`, `message_origin.go`, `reply_route_index.go`, `lifecycle.go`, `delivery_inject.go`.
- WebSocket cloud controller: `ws_controller.go`.
- Desktop reverse RPC: `desktop_rpc/*`.
- Migration endpoints: `migrate_claudecode.go`.
- Share handlers: `share_handler.go`, `share_async.go`.
- Suggestion handler and system event store: `suggestion_handler.go`, `system_event_store.go`.

StarClaw Phase 5 already closed local runtime foundations: budget, routing/fallback metadata, observability, trace export, durable run metadata, pause/resume/cancel/replay, docs, and redaction. The next gap is now Kocoro-style orchestration and daemon runtime depth.

## Tool gaps

Kocoro has tool files not present in StarClaw, including:

- Deep browser/GUI: `axclient.go`, `axserver/`, `browser_handoff.go`, `browser_lease.go`, `pinchtab.go`.
- Calendar through Desktop RPC: `calendar_*`.
- Terminal/workspace: `ghostty_*`.
- Images: `generate_image.go`, `edit_image.go`, image compression helpers.
- Cloud/share lifecycle: `retract_published_file.go`, richer publish/list support.
- Document extraction consolidation: `doc_extract.go`.
- MCP server wrapper file is `server.go`; StarClaw has `mcp_server.go`, so this is naming/shape rather than complete absence.

## Phase 6 recommendation

The next parent should be:

`Astria Kocoro parity phase 6: engine orchestration and channel runtime`

Goal: move StarClaw/Astria from a hardened local runtime/workbench into Kocoro-like engine behavior: executable workflows, durable message queues, channel runtime contracts, memory sidecar foundations, Desktop RPC boundary, and deeper local tool orchestration.

### Recommended child order

1. `kocoro-style-workflow-orchestration`
   - Build `/research` and `/swarm` style workflow entry points in StarClaw.
   - Use daemon `POST /message` + SSE as the first-class contract.
   - Upgrade Astria council from planner/handoff into executable workflow orchestration.
   - Keep external Shannon Cloud optional; define a local-first provider interface first.

2. `daemon-mailbox-channel-runtime`
   - Add typed queued messages, route keys, mailbox persistence, claim/ack style lifecycle, busy/inject behavior, and bounded worker semantics.
   - Model this after Kocoro's `agenttypes`, `mailbox_store`, `queue_*`, `channel_state_*`, and `delivery_inject` families.
   - Keep actual Slack/Feishu/Telegram cloud transport out of the first slice unless explicitly approved.

3. `episodic-memory-sidecar-foundation`
   - Add a memory provider abstraction, local bundle files, sidecar readiness state, and preflight recall injection.
   - Preserve local-first behavior and content-free observability.
   - Defer cloud sync/upload until the user explicitly accepts the privacy tradeoff.

4. `desktop-rpc-boundary`
   - Add a Unix socket RPC broker/listener contract for future native Desktop integration.
   - Start with `system.capabilities` / `system.ping` and a fake-desktop smoke harness.
   - Calendar tools can follow once the boundary is stable.

5. `desktop-browser-tool-depth`
   - Deepen existing `browser`, `computer`, `accessibility`, and `screenshot` tools with Kocoro-style AX/browser handoff, leases, visual verification, and optional Ghostty workspace.

6. `share-sync-delivery-lifecycle`
   - Add local share artifact rendering and retractable publishing abstractions.
   - Add session sync dry-run/marker/uploader only after privacy and redaction boundaries are explicit.

## Priority rationale

OpenAI-compatible streaming is still valuable, but it is not the top Kocoro parity gap. Kocoro's open-source daemon emphasizes `POST /message` SSE, `/research`, `/swarm`, channel runtime, memory sidecar, and Desktop RPC. Therefore Phase 6 should prioritize engine orchestration and daemon/channel runtime before polishing the OpenAI-compatible gateway.

If product strategy shifts toward "StarClaw as a local OpenAI API endpoint", then `openai-compatible-streaming-tools` should move back to first place. For Kocoro parity, it should be a later compatibility slice.

## 2026-06-08 follow-up: local checkout baseline and next plan

Kocoro is now available locally at `/Users/timmy/PycharmProjects/Kocoro`, so future parity checks should use that checkout first instead of repeatedly reading GitHub. The checkout baseline for this comparison is still:

- Commit: `74cdb3c Merge pull request #233 from Kocoro-lab/fix/mailbox-consume-after-save`
- Branch state: `main...origin/main`
- Refresh note: `git pull --ff-only` failed once with `Error in the HTTP2 framing layer`; retrying with `http.version=HTTP/1.1` hung and was terminated. Treat this plan as based on local commit `74cdb3c` until a later explicit refresh succeeds.

Phase6 status against the recommended child order:

1. `kocoro-style-workflow-orchestration` — done and archived.
2. `daemon-mailbox-channel-runtime` — done and archived.
3. `episodic-memory-sidecar-foundation` — done and archived.
4. `desktop-rpc-boundary` — current planned child. This is still the right next task because Kocoro exposes native Desktop capabilities through `internal/daemon/desktop_rpc/*`, and StarClaw has no equivalent package yet.
5. `desktop-browser-tool-depth` — next after Desktop RPC boundary. Kocoro still has deeper browser / AX / lease / handoff / PinchTab / Ghostty-style tool surfaces than StarClaw.
6. `share-sync-delivery-lifecycle` — after tool depth. Kocoro still has dedicated `internal/share`, `internal/sync`, retractable publish, and upload lifecycle code that StarClaw only partially mirrors through current publish/upload tools.

Remaining post-Phase6 gaps likely need additional phases:

- Phase7: channel/cloud delivery parity, including richer route indexing, connection state, channel lifecycle, WebSocket controller equivalents, and opt-in external channel transports. Keep this local-first until cloud/privacy approval is explicit.
- Phase8: native desktop and tool parity, expanding the Desktop RPC boundary into calendar/native tools, deeper browser handoff, AX integration, Ghostty/workspace helpers, and visual verification.
- Phase9: share/sync/migration parity, including local share rendering, retractable publishing abstractions, sync dry-run/marker/uploader lifecycle, and Claude Code import/migration support.
- Phase10: product polish and compatibility, including OpenAI-compatible streaming/tool-call deltas if StarClaw is intended to serve as a local OpenAI-style endpoint, plus Astria stellar workbench UI polish after core platform gaps are closed.

## 2026-06-08 follow-up: Phase6 complete and Phase7 recommendation

Phase6 child tasks are now complete and archived:

1. `kocoro-style-workflow-orchestration`
2. `daemon-mailbox-channel-runtime`
3. `episodic-memory-sidecar-foundation`
4. `desktop-rpc-boundary`
5. `desktop-browser-tool-depth`
6. `share-sync-delivery-lifecycle`

Remaining module-level gaps against local Kocoro commit `74cdb3c`:

- `internal/agenttypes`
- `internal/cloudflow`
- `internal/images`
- `internal/keychain`
- `internal/memory`
- `internal/migrate`
- `internal/sync`

Remaining daemon gaps are now concentrated in channel/cloud delivery and platform lifecycle depth:

- `ws_controller.go`
- `channel_state_*`
- `connection_state_cache.go`
- `reply_route_index.go`
- `delivery_inject.go`
- `message_origin.go`
- `system_event_store.go`
- `suggestion_handler.go`
- auth / IM bindings / Feishu handlers

Recommended next parent:

`Astria Kocoro parity phase 7: channel and cloud delivery parity`

Goal: move StarClaw from local-only workflow/channel foundations toward Kocoro's channel delivery model: cloudflow dispatch contracts, connection state and route indexing, system events, suggestion events, channel lifecycle state, and optional external transport boundaries. Keep this local-first unless cloud credentials / off-machine transport are explicitly approved.

Recommended child order:

1. `cloudflow-dispatch-contract`
   - Add a Kocoro-style cloudflow package boundary for slash command parsing/display/dispatch.
   - Extend workflow parsing to support `/dag` as auto orchestration.
   - Keep provider local/null by default; no cloud Gateway call in the first slice.

2. `channel-route-index-connection-state`
   - Add route index and connection-state cache foundations.
   - Track route keys, last-seen state, and safe injection targets without external IM transports.

3. `system-event-store-suggestions`
   - Add durable system event store and suggestion records.
   - Surface content-free system events for diagnostics and UI.

4. `delivery-inject-lifecycle-depth`
   - Deepen queue/mailbox delivery injection lifecycle around busy runs, orphan replies, and re-enqueue behavior.

5. `external-channel-adapter-boundaries`
   - Define Feishu/Slack/Telegram adapter interfaces and test fakes.
   - Do not enable real external transport without explicit privacy and credential approval.
