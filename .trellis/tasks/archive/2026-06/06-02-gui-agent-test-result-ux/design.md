# Design

## UI Behavior

- Keep the test runner in the Agents panel as the primary completion surface.
- After a test completes, refresh Runs and Sessions in the background but do not navigate away.
- Keep explicit navigation through existing result buttons.

## Result Summary

- Build a single summary string from payload and result:
  - agent;
  - prompt;
  - request/run ID;
  - session ID;
  - usage;
  - messages or error.
- Use the existing `copyText` and `markButtonCopied` helpers via event delegation.

## Error State

- Replace generic error rendering for agent tests with a result-like error card that includes agent, prompt, request ID, and error message.
