# Design

## Frontend

Reuse the existing command editor state:

- `state.selectedAgentCommand`: selected command key.
- `state.agentCommands`: staged command map included in the agent save payload.

Changes:

- Keep command name input enabled while editing.
- Add a `Clear` button that resets selected command/name/body without changing `state.agentCommands`.
- On save, if `selectedAgentCommand` differs from the entered name, delete the old key and write the new key.
- Use the same command-name regex already used by the frontend and backend.

## Backend

No backend contract changes. Existing command persistence remains the source of truth.

## Compatibility

Existing command files and payload shape remain unchanged.
