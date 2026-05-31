# Design

## Approach

Add a focused Go integration test that drives `agent.AgentLoop` through `client.MockClient` responses. The mock client returns ordered tool calls, allowing the test to exercise the same `AgentLoop.Run` path used by real providers without external dependencies.

## Workflow Under Test

1. Create an isolated temporary working directory.
2. Seed an input file.
3. Configure a session manager and current session.
4. Use a mock LLM response sequence:
   - `file_read` reads the seed file.
   - `glob` or `grep` verifies discovery/search behavior.
   - `file_write` creates a generated output file.
   - `file_edit` modifies that output file.
   - `bash` inspects the result.
   - final text response summarizes completion.
5. Assert:
   - event handler observed all expected tools and outputs.
   - output file exists and contains edited content.
   - shell output was captured.
   - session can be resumed with saved messages.

## Location

Use `tests/` for the integration-style workflow test because it validates cross-package runtime behavior.

## Risk Controls

- Use temporary directories and restore CWD.
- Use mock client only; no network.
- Use simple shell commands with portable POSIX syntax.
- Avoid relying on tool call prompt inference. The mock explicitly returns tool calls.
