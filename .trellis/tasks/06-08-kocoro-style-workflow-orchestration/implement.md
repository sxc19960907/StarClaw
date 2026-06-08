# Implementation Plan

## Scope

Implement Kocoro-style local workflow commands over StarClaw daemon `POST /message` without adding external cloud/channel transport or Web UI redesign.

## Checklist

- [x] Read backend specs and shared thinking guides before editing.
- [x] Inspect existing `/message`, SSE, run-store workflow step, routing, and council API patterns.
- [x] Add workflow command parsing for `/research` and `/swarm`.
- [x] Reject blank `/research` and `/swarm` goals before `RunStore.Start`.
- [x] Add a local workflow runner that records ordered `WorkflowStepState` entries.
- [x] Integrate workflow dispatch into JSON and SSE `/message` paths without changing ordinary prompt behavior.
- [x] Add daemon tests for slash-command parsing, bad/empty goals, JSON execution, SSE execution, run-store workflow steps, trace events, metrics, and sanitization.
- [x] Update backend quality spec for the new workflow command API contract.
- [x] Run focused validation, then full project validation.

## Validation Commands

```bash
go test ./internal/daemon
go test ./internal/agent
go test ./internal/daemon ./internal/agent
go test ./...
```

## Completion Notes

- Added `/research` and `/swarm` workflow command parsing for daemon `POST /message`.
- Command workflows transform the request into local orchestration prompts, then execute through the existing `s.runAgent` path for both JSON and SSE.
- Workflow command runs record sanitized run-store workflow steps and structured `workflow_step` trace events.
- Blank `/research` or `/swarm` goals return HTTP 400 before run-store start.
- Existing OpenAI-compatible gateway behavior was not changed by this child.

## Rollback Point

Remove `workflow_command.go`, `workflow_runner.go`, and the workflow branch in `handleMessage` to restore direct-only `/message` behavior.
