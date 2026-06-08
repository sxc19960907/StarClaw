# Design

## Current State

StarClaw has daemon-local `/research` and `/swarm` workflow parsing in `internal/daemon/workflow_command.go`. Kocoro has a separate `internal/cloudflow` package for slash parsing, display formatting, and Gateway dispatch. StarClaw lacks that package boundary and does not support `/dag`.

## Proposed Architecture

Add `internal/cloudflow`:

- `SlashCommand`
  - `Type`: `research`, `swarm`, `auto`
  - `Strategy`: research strategy
  - `Query`
- `ParseSlash(text string) *SlashCommand`
- `CloudStatusLine(agentID, status, message string) string`
- `Provider` interface:
  - `Dispatch(ctx, Request, EventSink) (Result, error)`
- `LocalProvider`
  - deterministic no-network provider for now
  - returns a message that cloudflow is not configured, preserving local-first behavior

Update daemon:

- `parseWorkflowInvocation` calls `cloudflow.ParseSlash`.
- `/dag` maps to new workflow type `auto`.
- Research strategy is preserved in workflow metadata and prompt.

## Boundaries

- No cloud Gateway client.
- No SSE streaming from cloud.
- No credentials or account config.
- No external transport.

