# Kocoro style workflow orchestration design

## Architecture

Introduce a local workflow orchestration layer inside the daemon runtime, keeping `POST /message` as the public launch contract.

Recommended shape:

- A small daemon workflow package or file group owns workflow request detection, planning, and execution.
- `RunAgentRequest` gains optional workflow metadata only if needed for structured clients; slash commands remain supported for Kocoro-style parity.
- `handleMessage` starts the run store entry as it does today, then dispatches either a normal agent run or a workflow run based on parsed workflow intent.
- Workflow execution reuses the existing event handler stack and run recorder so SSE, trace events, run summaries, metrics, and redaction keep the same behavior.
- Astria launch controls seed chat input or submit through the same `/message` path.

## Contracts

### Slash Commands

`/research <goal>`:

- Kind: `research`
- Expected stages: intake, plan, gather evidence, synthesize, finalize
- Output: a concise research answer with evidence and next-step caveats.

`/swarm <goal>`:

- Kind: `swarm`
- Expected stages: intake, planner, researcher, reviewer, synthesis, finalize
- Output: a role-grounded synthesis and concrete next action.

Empty goals return `400` for HTTP JSON requests or an SSE `error` for streaming requests, matching existing daemon error behavior.

### Run Metadata

Workflow runs should expose content-safe metadata:

- workflow kind
- provider, initially `local`
- stage count and current stage
- source command, without raw goal text in aggregate telemetry

Raw user goal text remains part of the operator-facing run request, as normal `POST /message` behavior already does, but should not be copied into metrics or structured event attributes that are meant to be content-free.

### Step State

Each workflow stage maps to `WorkflowStepState`:

- Stable IDs such as `workflow.intake`, `workflow.plan`, `workflow.evidence`, `workflow.synthesis`.
- Titles are user-readable and content-safe.
- Metadata includes kind, provider, stage index, and status details after redaction.
- Terminal stages use existing completed/failed statuses.

### Event Flow

Workflow execution should publish progress through the same handler path used by agent events. If direct workflow events are needed, use existing daemon event names and structured run events rather than creating a parallel bus.

For SSE requests:

- Emit progress for workflow steps.
- Preserve existing text/tool streaming semantics for any model-backed sub-run.
- Always emit final `done` with `RunAgentResponse` or `error` with the standard error envelope.

## Data Flow

1. Client posts to `/message`.
2. Daemon decodes `RunAgentRequest`, fills source/channel/request ID defaults, and starts the run store.
3. Workflow parser checks slash command or structured workflow metadata.
4. Normal requests call `runAgent` unchanged.
5. Workflow requests call the local workflow orchestrator.
6. Orchestrator writes step state into `RunStore`, emits progress through the provided handler, calls `runAgent` for model-backed stages when needed, and returns one `RunAgentResponse`.
7. `handleMessage` or `handleMessageSSE` completes the run store and writes JSON/SSE response as today.

## Compatibility

- Existing `/message` non-workflow requests must behave unchanged.
- Existing `/council` and `/council/{id}/run` must continue to work. UI additions should not remove council planning.
- Existing OpenAI-compatible `/v1/chat/completions` does not need workflow slash-command support unless it naturally maps through `RunAgentRequest`.
- Existing redaction tests are the guardrail for workflow metadata.

## Trade-Offs

Sequential local orchestration is recommended for this slice. It gives users executable workflow shape now while preserving predictable run control and avoiding mailbox/channel concurrency work that belongs to the next Phase 6 child.

Bounded concurrent role execution would be closer to the word "swarm", but it requires durable worker semantics, cancellation fan-out, mailbox/ack behavior, and sharper UI affordances. Those are better handled after the daemon-mailbox-channel-runtime task.

## Rollback

The implementation should be easy to remove by:

- Deleting the workflow parser/orchestrator files.
- Removing the branch in `handleMessage`/SSE dispatch.
- Removing UI launch controls and tests.

Normal daemon message, run-store, council, and OpenAI routes should remain intact.
