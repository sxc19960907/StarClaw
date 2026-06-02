# Design

## Frontend

Add a runner form to the Agents detail pane near the existing editor actions:

- `#agent-test-agent`: select with the same options as chat/schedule agent selects.
- `#agent-test-prompt`: textarea.
- `#agent-test-form`: submit.
- `#agent-test-output`: result panel.

`updateAgentSelects()` updates the test select as well.

The existing editor `Test run` button changes behavior:

1. set test select to current agent
2. prefill prompt
3. focus the prompt
4. keep user in Agents panel

Submitting the runner posts to `/message` with:

```json
{
  "text": "...",
  "agent": "selected-agent",
  "new_session": true,
  "request_id": "agent-test-..."
}
```

After success:

- render status/result/usage/session/request
- call `loadRuns()` and `loadSessions()`
- allow `Open run` via existing `selectRun(requestID)`
- allow `Open session` via existing `selectSession(sessionID)`

The runner uses non-streaming JSON for simplicity and deterministic smoke behavior.

## Test Strategy

- Web UI smoke:
  - create a smoke agent
  - use editor `Test run` to populate runner
  - mock `/message` for direct agent test
  - submit prompt
  - assert payload includes selected agent
  - assert output displays result, usage, request id
  - mock `/runs/{id}` or route to existing run path when testing `Open run`

## Rollback

Remove the runner form and JS functions if it regresses existing agent editor or chat flows.
