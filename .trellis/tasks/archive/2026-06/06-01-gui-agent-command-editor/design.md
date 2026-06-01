# Design

## API Contract

Extend the existing agent create/update payload with an optional `commands` object:

```json
{
  "commands": {
    "review": "Review recent changes for regressions."
  }
}
```

If `commands` is omitted, preserve existing command files. If present, replace the managed `commands/` directory with exactly the provided entries.

## Persistence

Each command is stored as `<agent>/commands/<name>.md`.

Command names must match a safe filename subset:

```text
^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$
```

Empty command content is omitted when saving. An empty `commands` object removes the `commands/` directory.

## Frontend

Add a Commands fieldset in the existing Agents editor:

- command list
- command name input
- command body textarea
- Add/update and delete controls

The frontend keeps a local `state.agentCommands` object and includes it in agent create/update payloads.

## Compatibility

Existing agents without commands load with an empty command editor. Existing clients that omit `commands` keep existing command files unchanged.
