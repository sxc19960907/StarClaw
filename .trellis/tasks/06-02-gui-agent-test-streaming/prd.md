# Improve agent test runner streaming and cancellation

## Goal

Make the GUI Agent Test Runner behave like the main chat runner: stream output as it arrives, expose cancellation, and connect completed tests directly to Run detail.

## Requirements

- Agent Test Runner must use the existing `/message` SSE path when available.
- Streaming text deltas must appear in the test output while the run is active.
- Tool calls, tool results, preamble, usage, and final response must be represented clearly enough for a focused agent test.
- Agent Test Runner must expose a Stop action while a test is running and cancel the active request through `/cancel`.
- A completed test run must refresh run history and automatically open its Run detail.
- Existing synchronous fallback behavior must remain usable if streaming is unavailable.
- The current Chat run flow must not regress.
- Browser smoke coverage must verify streaming test output, cancellation controls, payload shape, and Run detail linkage.

## Acceptance Criteria

- [x] Agent Test Runner streams assistant output before completion.
- [x] Agent Test Runner disables input while running and shows Stop instead of Run test.
- [x] Stop cancels the active test request and resets the runner state.
- [x] Completed Agent Test Runner runs refresh Runs and select the matching Run detail.
- [x] The test result still shows session, usage, request ID, and result text.
- [x] Existing chat streaming and run summary behavior still pass smoke.
- [x] Targeted Go tests and Web UI smoke pass.

## Notes

- Scope is limited to GUI behavior and existing daemon endpoints.
- Do not add a new backend endpoint unless the existing `/message`, `/cancel`, and `/runs/{id}` contracts are insufficient.
