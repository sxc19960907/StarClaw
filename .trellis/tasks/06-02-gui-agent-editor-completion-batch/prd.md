# Complete agent editor GUI workflows

## Goal

Finish the Agent editor as a coherent GUI workflow: command editing, config import/export, permission preview, and unsaved-change protection.

## Requirements

- Command editor UX:
  - Provide explicit `New command` and `Cancel edit` actions.
  - Keep command rename behavior clear: editing an existing command and changing its name replaces the old staged command.
  - Surface when command changes are staged but not yet saved to the agent.
- Unsaved changes:
  - Detect dirty agent form state after edits to fields, command list, or imported config.
  - Warn before discarding unsaved edits when starting a new agent, opening another agent, deleting the current agent, or jumping to Test run.
  - Clear the dirty state after successful save or after resetting/loading an agent.
- Agent config import/export:
  - Export the current editor state as JSON, including name, prompt, memory, model, permissions, heartbeat, and commands.
  - Import a JSON config into the editor without immediately saving.
  - Imported config must mark the editor dirty so the user explicitly saves.
- Permissions preview:
  - Show a read-only preview of the effective agent-level allow, deny, and auto-approve values before saving.
  - Update the preview as the editor changes.
- Existing backend agent create/update APIs must continue to be used. No new backend API is required for this batch.

## Acceptance Criteria

- [ ] Agent editor shows `New command` and `Cancel edit` controls with clear behavior.
- [ ] Changing an edited command name stages a rename and removes the old staged command name.
- [ ] Unsaved agent edits show a dirty state and prompt before destructive navigation/actions.
- [ ] Export produces JSON containing the current editor state and commands.
- [ ] Import populates the editor, stages command/config changes, and requires Save agent.
- [ ] Permission preview reflects allow/deny/auto-approve from the current editor.
- [ ] Web UI smoke covers command UX, import/export, permission preview, and dirty warning behavior.

## Notes

- This is intentionally a larger batch task. Do not split these into separate small Trellis tasks.
