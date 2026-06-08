# Kocoro style workflow orchestration

## Goal

Add local-first `/research` and `/swarm` workflow entry points to StarClaw/Astria so multi-step work can be launched through the existing daemon `POST /message` contract, streamed over SSE, recorded as durable workflow steps, and surfaced as executable orchestration rather than static council planning.

## User Value

Users should be able to start a structured research or swarm-style run from Astria or local HTTP clients without leaving the current StarClaw daemon/runtime model. The first slice should make orchestration observable, replayable, and cancellable through existing run-store and control surfaces, while keeping cloud/channel transport optional for later Phase 6 children.

## Confirmed Facts

- StarClaw already exposes `POST /message` with sync JSON and SSE behavior, plus run-store recording and runtime control.
- StarClaw already records durable workflow step state through `RunStore.UpsertStep` and `RunStore.TransitionStep`.
- StarClaw already has deterministic council planning and a council-to-run handoff, but it is not a true executable `/research` or `/swarm` workflow engine.
- The saved Kocoro comparison identifies `/research` and `/swarm` over daemon `POST /message` + SSE as the first recommended Phase 6 child.
- Kocoro cloud/channel transport, memory sidecar, mailbox runtime, Desktop RPC, share/sync, and deeper browser tooling are separate Phase 6 slices and should not be pulled into this first child.

## Requirements

- Detect workflow commands sent through daemon `POST /message`:
  - `/research <goal>`
  - `/swarm <goal>`
- Support an explicit structured request form for local clients without requiring slash-command parsing. The exact shape may be added to `RunAgentRequest` or a nested workflow field if that fits existing code better.
- Keep `POST /message` as the first-class launch surface for workflow runs. New helper routes may be added only if they delegate into the same runtime path.
- Stream workflow lifecycle progress through the existing SSE event path when clients request `Accept: text/event-stream`.
- Persist workflow stage state in the existing run store using `WorkflowStepState` / structured `workflow_step` events.
- Preserve normal daemon permission, approval, session, run-store, cancellation, pause/resume, replay, metrics, trace export, and redaction behavior.
- Implement local-first providers:
  - `research` should break a goal into evidence-gathering, synthesis, and answer stages.
  - `swarm` should break a goal into planner, researcher, reviewer, and synthesis stages.
- Upgrade Astria's council/workflow surface enough that users can draft or launch `/research` and `/swarm` workflows from the UI, with run details showing their workflow steps through existing runtime panels.
- Add tests proving command parsing, SSE/workflow step emission, persisted run metadata, cancellation/control compatibility where practical, and UI/static behavior for launch affordances.

## Acceptance Criteria

- [ ] `POST /message` with `/research some goal` launches a research workflow run and returns/streams through the same daemon message contract.
- [ ] `POST /message` with `/swarm some goal` launches a swarm workflow run and returns/streams through the same daemon message contract.
- [ ] Workflow runs include stable run metadata identifying workflow kind, source command, and local provider without leaking raw prompt content into aggregate telemetry.
- [ ] Workflow runs create ordered durable workflow steps that are visible through `GET /runs/{id}` and structured trace/event surfaces.
- [ ] SSE clients receive workflow progress and a final `done` or `error` event without breaking existing text/tool event handling.
- [ ] Existing council handoff behavior still works and can launch or draft the new workflow commands from Astria.
- [ ] Existing daemon runtime controls and run summaries remain compatible with workflow runs.
- [ ] Focused daemon tests and Web UI static tests pass.

## Out of Scope

- Real Shannon Cloud, Slack, LINE, Feishu, Telegram, or other external channel transport.
- Durable mailbox claim/ack lifecycle and channel worker queues.
- Memory sidecar, cloud sync/upload, Desktop RPC, share publishing, or deep browser/AX leases.
- Replacing `POST /message` with a separate workflow-only daemon API.
- Full parallel execution of independent model workers if the first slice can provide deterministic staged orchestration.

## Open Questions

- Product decision: should `/swarm` execute role stages sequentially in the first slice, or launch bounded concurrent agent calls immediately?

Recommended answer: start sequentially with explicit step state and reserve bounded concurrency for the mailbox/channel-runtime child. This keeps the first workflow slice testable and compatible with existing run controls.
