# Design

## Scope

Focus on CLI startup behavior and user-facing launch diagnostics. Avoid introducing platform installers or auto-update mechanisms in this task.

## Approach

- Inspect the existing `starclaw app` command and daemon startup implementation.
- Preserve the current daemon HTTP API and Web UI routes.
- Prefer small CLI/service changes over adding a new process manager.
- Add tests around command behavior where the existing command test patterns make this practical.

## Compatibility

- Existing `starclaw daemon start` semantics should remain unchanged unless a bug is found.
- `starclaw app` should continue to print the Web UI URL for terminal users.
- Browser opening failure should not imply daemon startup failure if the daemon is usable.
