# Improve GUI agent test result UX

## Goal

Make the GUI agent test runner useful as an in-place agent editor workflow, while still linking to run history when a user wants deeper inspection.

## Requirements

- Running an agent test from the Agents panel should keep the user on the Agents panel after completion.
- The test result card should show:
  - agent;
  - prompt;
  - session ID;
  - usage;
  - request/run ID.
- The result card should offer actions:
  - open run;
  - open session;
  - copy summary.
- Error results should keep enough context to debug which agent/prompt/request failed.
- Existing cancellation behavior should keep working.
- Browser smoke should cover the in-place result and copy action.

## Acceptance Criteria

- [x] Agent test completion leaves `#panel-agents` active.
- [x] Agent test result includes prompt and request metadata.
- [x] Agent test result can copy a summary to the clipboard.
- [x] Existing Open run and Open session actions remain available.
- [x] Web UI agents smoke passes with the new assertions.
- [x] JS syntax and diff checks pass.
