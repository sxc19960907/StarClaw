# Design

## Current State

- Chat already uses `/message` with `Accept: text/event-stream`.
- Agent Test Runner currently posts `/message` synchronously through `api()`.
- Run history already records request IDs and supports `/runs/{id}` detail.
- Agent Test Runner already renders a final result card with Open run / Open session actions.

## GUI Changes

- Reuse the existing streaming parser path from chat by extracting shared behavior where useful:
  - create an `AbortController`;
  - call `streamMessage(payload, outputElement, signal)`;
  - render stream delta events into the agent-test output area;
  - keep synchronous fallback through the existing `streamMessage()` fallback.
- Add a Stop button to the Agent Test Runner form.
- Track the active test request with `state.activeAgentTestRequestID` and `state.activeAgentTestAbort`.
- On completion:
  - render the final Agent test result;
  - refresh runs and sessions;
  - select the matching run detail using the test request ID.
- On cancellation:
  - POST `/cancel` with the request ID;
  - abort the active fetch controller;
  - show a cancelled state without leaving disabled controls.

## Compatibility

- Existing chat behavior must remain unchanged.
- Existing run summary and run detail actions continue to use the same `data-run-summary-*` handlers.
- If SSE is not available, `streamMessage()` falls back to synchronous JSON and the test result still renders.

## Testing

- Extend Web UI agents smoke to mock an SSE `/message` response and verify:
  - request payload includes agent, text, `new_session`, and `request_id`;
  - streamed text is visible before done;
  - final result card includes session, usage, and Open run;
  - Run detail is selected after completion.
- Add/adjust cancellation smoke coverage with a controlled route if practical without making the smoke flaky.
