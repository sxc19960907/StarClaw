# Design

## Boundary

This task primarily changes Web UI assets:

- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

Existing daemon agent APIs are reused:

- `GET /agents`
- `GET /agents/{name}`
- `POST /agents`
- `PUT /agents/{name}`
- `DELETE /agents/{name}`

No backend contract changes are planned.

## Editor State

Add two frontend-only state fields:

- `agentDirty`: whether current editor contents differ from the last loaded/saved/reset state.
- `agentDirtyBaseline`: serialized baseline snapshot after loading/resetting/saving.

Dirty detection compares a stable JSON string of `buildAgentPayload()` against the baseline. Command staging, form input changes, import, and delete command trigger dirty recalculation.

Guarded actions:

- `startNewAgent`
- `inspectAgent`
- `testCurrentAgent`
- `deleteCurrentAgent`

Each asks for confirmation before discarding unsaved editor state.

## Command Editor

Current command behavior already supports rename by deleting `state.selectedAgentCommand` when the command name changes. This task makes the behavior explicit in the UI:

- `New command`: clears command fields and starts a new command.
- `Cancel edit`: resets command fields and exits command edit mode.
- Save button remains `Add command` / `Update command`.
- Form state indicates staged commands require Save agent.

## Import / Export

Export:

- Uses `buildAgentPayload()` as the source of truth.
- Produces a JSON download named `<agent-name-or-agent>-config.json`.

Import:

- Reads a user-selected JSON file via `FileReader`.
- Accepts the same shape as `buildAgentPayload()`.
- Populates fields and `state.agentCommands`.
- Does not save automatically.
- Marks the form dirty.

## Permission Preview

Add a compact read-only preview inside the permissions fieldset:

- Allow: normalized allow list or `None`.
- Deny: normalized deny list or `None`.
- Auto approve: `Enabled` or `Disabled`.

The preview updates on form input and after import/load/reset/save.

## Compatibility

Existing command validation remains client-side and backend-side. Existing smoke test paths keep using the same create/update/delete endpoints.

## Test Strategy

Extend the existing Web UI smoke agent block:

- verify permission preview reflects initial allow/deny/auto-approve values,
- verify staged command rename keeps old command removed,
- verify dirty state appears after edits and warns before `New agent`,
- export current config and assert JSON fields,
- import JSON into the editor and assert fields/commands populate and dirty state appears,
- save after import and verify backend update response.
