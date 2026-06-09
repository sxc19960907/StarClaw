# Phase 24 final gap audit

## Baseline

- StarClaw/Astria workspace: `/Users/timmy/PycharmProjects/StarClaw`
- Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro`
- Kocoro commit checked locally:
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`
- Audit date: 2026-06-09

This audit follows Phase23. It separates two targets that should not be
collapsed into one number:

1. **Local-first Kocoro platform parity**: daemon, desktop shell, local APIs,
   streaming, recovery, observability, Desktop RPC, release/updater safety.
2. **Full Kocoro/Shannon production parity**: Shannon Cloud auth/WebSocket,
   cloud channel delivery, IM message lifecycle, remote uploads/sync, team and
   distribution surfaces.

## Executive decision

Astria should treat the local-first Kocoro parity track as effectively complete
and should not claim production Shannon Cloud/channel parity.

Recommended next phase: **Phase25 cloud/channel contract simulation**, only if
the product direction is to pursue cloud/channel parity. Phase25 should add
local contracts, fake cloud transport, lifecycle simulations, and Web UI
readiness states before any real Shannon Cloud credentials, OAuth flows, or
external WebSocket transport are introduced.

If the product direction remains local-first, Phase25 should instead be a
cleanup phase: docs capability alignment, route/API reference cleanup, and
archiving old completed Trellis tasks.

## Parity estimate

- **Local-first Kocoro platform parity**: approximately **98-99%**.
  Phase23 closed the last obvious user-visible streaming/long-run gap on top
  of earlier Desktop RPC, native shell, release/updater safety, workflow
  control, structured observability, and recovery work.
- **Full Kocoro/Shannon production parity**: approximately **65-75%**.
  The local runtime is strong, but Kocoro's production Cloud path still includes
  real WebSocket channel delivery, auth, IM lifecycle, proactive targeted
  delivery, Cloud uploads, Cloud memory/session sync, and team/distribution
  surfaces that Astria intentionally keeps disabled or simulated.

## Capability matrix

| Area | Astria / StarClaw status | Kocoro baseline evidence | Remaining gap | Recommendation |
|---|---|---|---|---|
| Local daemon and Web UI | Strongly aligned. Astria serves `/app/`, local `/message`, `/events`, `/runs`, `/metrics`, run trace/export, workflow control, schedules, agents, skills, memory, inbox, queue, and channel state surfaces. | Kocoro also centers production on the daemon and Desktop. `AGENTS.md` describes `internal/daemon` as the primary production path. | No major local-first gap. | Keep local daemon APIs stable. |
| Streaming and long-running run UX | Strongly aligned for local clients. Phase23 added live run status with stream state, run id, session id, usage, latest event, cancel/error/complete state. | Kocoro streams through gateway/SSE/WS and exposes tool/use-id events and IM timeline capabilities. | Astria does not stream over Shannon Cloud WS because cloud transport is disabled. | No local-first work needed; cloud stream transport belongs to future Phase25. |
| Daemon events and replay | Strongly aligned locally. `docs/DAEMON_EVENTS.md` documents replayable `/events`, `Last-Event-ID`, Kocoro-compatible `/message` aliases, and redaction. | Kocoro has EventBus plus WebSocket lifecycle and Cloud message routing. | Astria explicitly lacks Kocoro IM `MESSAGE_LIFECYCLE` by default. | Keep local `/events`; add simulated IM lifecycle only in a future cloud/channel phase. |
| Workflow control and runtime recovery | Strongly aligned. Astria has cancel, pause, resume, replay controls, durable run state, trace export, Mission Control recovery UI, and streaming live status. | Kocoro has route cancellation, injection, mailbox, run lifecycle, and workflow progress. | Full Cloud route cancellation/replay semantics are not present without Cloud messages. | No local-first gap. Future cloud phase should test route cancellation with fake channel messages. |
| Desktop/native app shell | Strongly aligned for local-first lifecycle. Previous phases added standalone macOS shell, daemon supervision, Desktop RPC socket/pidfile, capability reconciliation, native diagnostics, menu/Dock/window behavior, crash summaries, notification readiness, and updater gates. | Kocoro Desktop owns Desktop RPC, daemon spawn/reconciliation, native EventKit bridge, and continues Cloud channel operation while Desktop may be absent. | Astria does not have full production Kocoro Desktop + Cloud lifecycle coupling. | Keep local shell stable; do not expand native work unless Cloud channel operation becomes a product goal. |
| Desktop RPC native tools | Partially aligned. Calendar Desktop RPC boundary exists; broader native tool coverage remains intentionally narrow. | Kocoro documents Calendar RPC v0.5.1 and broader Desktop RPC protocol; additional native surfaces are product dependent. | Contacts/Reminders/file-provider style tools remain unscoped. | Defer unless user asks for native OS depth beyond Calendar. |
| Release/updater | Strong local gate alignment. Astria has CLI/npm-only updater decision, signed metadata validation boundaries, transaction/rollback/health manifests, sandbox updater rehearsal, and replacement disabled by policy. | Kocoro has CLI self-update via GitHub releases; public repo does not by itself solve signed Desktop app replacement. | Astria does not do real installed app replacement, signing, notarization, or public release execution. | Keep current decision. Do not start real app replacement without Apple signing/notarization scope. |
| Session sync | Foundation only. Astria has local session sync config, scanner, batching, marker/lock, thinking stripping, dry-run outbox, and no Cloud uploader by default. | Kocoro `internal/sync` includes uploader/backoff and README documents opt-in daily session upload to Shannon Cloud. | No real Cloud sync upload or account-backed memory training pipeline. | If cloud/channel parity starts, make sync upload a separate opt-in child after auth and privacy policy. |
| Session share / published files | Local-only equivalent. Astria `publish_to_web` records local web artifacts and retracts local manifests. | Kocoro uploads session shares and published files to Shannon Cloud `/api/v1/uploads`, returns public CDN URLs, supports list/retract and share progress events. | Astria lacks public Cloud uploads, CDN URLs, owner-scoped remote retract, and async Cloud share progress. | Keep local share by default. Future cloud phase should start with fake upload client and explicit approval rules. |
| Cloud tools | Partial/local boundary. Astria has `cloud_delegate` and cloud lifecycle surfaces, but current lifecycle note says no external transport is active. | Kocoro registers `cloud_delegate`, publish/list/retract uploads, image generation/editing when Cloud is enabled and API key exists. | Real Shannon gateway tools, remote research/swarm, image generation via Cloud, and public upload execution are absent. | Do not enable by default. Future work requires credential/auth/product consent. |
| Channel adapters and inbox | Contract/simulation level. Astria has fake channel adapters for Slack/Feishu/Telegram/webhook metadata, local webhook/GitHub inbox, `/queue`, `/channel/routes/{message_id}`, and `/channel/state`. | Kocoro daemon connects to Cloud WebSocket `/v1/ws/messages`, receives Slack/LINE/Feishu/Telegram/webhook messages, uses channel state events, and sends replies/proactive payloads. | No real external channel install/OAuth, inbound Cloud delivery, outbound reply delivery, or provider-specific lifecycle. | Phase25 should begin here if cloud/channel parity is approved. |
| IM message lifecycle | Not implemented by design. Astria docs explicitly say no IM `MESSAGE_LIFECYCLE` by default. | Kocoro has `CapIMMessageLifecycleV1`, `MESSAGE_LIFECYCLE`, `received`, `processing`, `done`, and `cleared`, plus `CloudMessageID` and `IMStatusContext` plumbing. | This is the main remaining protocol parity gap. | Future Phase25 child: local fake lifecycle protocol and redacted status UI, no real Cloud. |
| Proactive delivery | Local schedules exist. Astria schedules run locally and can surface in UI. | Kocoro schedules can broadcast to Cloud channels with `broadcast` and `thread` modes, using `IMStatusContext` for precise routing. | No Cloud channel proactive delivery or thread anchoring. | Future Phase25 child after fake channel route/lifecycle simulation. |
| Auth/account/team surfaces | Mostly absent by product choice. Astria uses local config/API keys and local-only runtime. | Kocoro has Cloud API key, macOS Keychain auth state, `/local/auth/*`, WS controller, marketplace and team/distribution-adjacent Cloud surfaces. | No account login, team install, remote marketplace backed by Cloud, billing, or distribution controls. | Treat as a product decision, not a missing local platform bug. |
| Skill/API reference surface | Astria has local docs and Trellis specs. | Kocoro requires bundled `kocoro` skill references to mirror daemon HTTP APIs. | Astria does not have an equivalent always-synced API assistant skill reference. | Optional cleanup if agents will operate Astria APIs heavily. Not a Kocoro runtime blocker. |

## Product decision boundary

The remaining delta is mostly **product scope**, not missing implementation
inside the local-first runtime.

Astria can honestly say:

- local daemon/runtime parity is effectively complete;
- local streaming, observability, workflow control, Desktop RPC, and release
  safety are robust;
- cloud/channel surfaces exist only as local contracts or simulations;
- Shannon Cloud behavior is intentionally not enabled.

Astria should not yet say:

- it supports Kocoro/Shannon Cloud production messaging;
- it handles real Slack/Feishu/Telegram/LINE lifecycle;
- it uploads sessions or shares to a remote service;
- it has account/team/distribution parity.

## Recommended Phase25 if cloud/channel parity proceeds

**Phase25: cloud/channel contract simulation foundation**

Child slices:

1. `cloud-channel-wire-contracts`
   - Define local structs and docs for fake Cloud envelopes, message ids,
     route keys, delivery acknowledgements, and provider-neutral channel state.
   - Keep all tests local and credential-free.

2. `im-message-lifecycle-simulation`
   - Add a fake/local lifecycle harness for `received`, `processing`, `done`,
     and `cleared` states using redacted metadata.
   - Do not connect to Shannon Cloud.

3. `channel-route-proactive-simulation`
   - Exercise queued inbound messages, route injection, schedule broadcast
     intent, and thread/top-level delivery decisions against fake providers.

4. `cloud-channel-ui-readiness`
   - Surface channel readiness/degraded states in Astria without promising real
     provider connectivity.

Exit criteria before real transport:

- no production endpoint or credentials required;
- no external WebSocket opened by tests;
- no OAuth provider setup;
- lifecycle and route simulation covered by focused daemon tests;
- explicit user/product approval before Phase26 real Cloud connector.

## Recommended Phase25 if local-first remains the product target

**Phase25: local-first completion cleanup**

Child slices:

1. capability documentation alignment in README/docs;
2. API/reference cleanup for queue/channel/cloud lifecycle local boundaries;
3. archive or retire old completed Trellis tasks that still appear active;
4. release a current capability snapshot with known out-of-scope Cloud gaps.

## Validation performed

- Inspected StarClaw current context and confirmed a clean working tree before
  task creation.
- Inspected Kocoro baseline commit with `git -C /Users/timmy/PycharmProjects/Kocoro rev-parse HEAD`.
- Compared Kocoro and StarClaw route/API surfaces with `rg`.
- Read StarClaw `docs/DAEMON_EVENTS.md`, README local runtime docs, and backend
  quality guidelines for observability and queue/channel boundaries.
- Read Kocoro `AGENTS.md`, `cmd/daemon.go`, `internal/daemon/lifecycle.go`,
  `internal/daemon/client.go`, and README Cloud/sync/share sections.
