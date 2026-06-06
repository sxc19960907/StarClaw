# Council Workflow Handoff Design

## Goal

Make Council output actionable without making it autonomous. Users explicitly decide when a synthesis becomes a normal Astria run.

## API Surface

- `POST /council/{id}/run`

Request:

```json
{
  "agent": "optional-agent-name"
}
```

Behavior:

- Fetch the Council run by ID.
- Build a prompt from the council goal and synthesis.
- Start a normal daemon run using existing run store/recording flow.
- Return `{ "run": RunAgentResponse, "run_id": "...", "council": CouncilRun }`.

## Metadata

Run request fields:

- `Source`: `council:<council_id>`
- `Channel`: `council_handoff`
- `Sender`: `agent-council`
- `Text`: handoff prompt containing goal and synthesis

This keeps the run auditable without adding a new persistence model.

## UI

Council detail adds a `Start run` action beside copy/send. The action is explicit and user-triggered. On success, Astria refreshes runs and opens the new run detail.

## Non-Goals

- No autonomous parallel agents.
- No task board.
- No background execution without click.
