# Add GUI direct agent test runner

## Goal

Let users test a selected agent directly from the GUI with a prompt and see the one-off result without manually switching to Chat setup.

## Requirements

- GUI:
  - Add a direct test runner area in the Agents panel.
  - User can choose an agent, enter a prompt, and run once.
  - The current editor's `Test run` button should populate the runner with that agent.
  - Show run status, result messages, usage, session id, and request id.
  - Provide actions to open the recorded run detail and session when available.
  - Prevent duplicate test submits while a test is running.
- Backend:
  - Reuse existing `POST /message`, `/runs`, and `/sessions` APIs.
  - No new backend endpoint unless needed.
- Compatibility:
  - Existing chat composer and agent editor behavior must continue working.
  - Keep embedded static GUI with no frontend build step.

## Acceptance Criteria

- [ ] Agents panel exposes a direct agent test runner.
- [ ] `Test run` on the agent editor selects the edited agent in the test runner.
- [ ] Running the test calls `/message` with the selected agent and prompt.
- [ ] Result, usage, session id, and request id are visible after completion.
- [ ] `Open run` opens the corresponding Runs detail panel.
- [ ] Web UI smoke covers the direct test runner workflow.

## Notes

- This is a GUI run-experience task. Run history persistence was completed in the previous task and should be reused.
