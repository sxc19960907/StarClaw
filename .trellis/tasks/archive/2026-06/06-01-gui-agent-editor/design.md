# Design

## Backend

Use existing routes:

- `POST /agents`
- `PUT /agents/{name}`

Request shape:

```json
{
  "name": "researcher",
  "prompt": "You are...",
  "memory": "optional",
  "model": "optional",
  "reasoning_effort": "optional",
  "tools_allow": ["file_read"],
  "tools_deny": ["bash"],
  "auto_approve": false
}
```

For PUT, name comes from path; request name is ignored or must match if present.

Persist files:

- `AGENT.md` always, mode `0600`
- `MEMORY.md` only when non-empty; remove when empty
- `config.yaml` when any config field is set; remove if config is empty

Return the loaded `agents.Agent` after write.

## Frontend

Enhance existing Agents panel:

- Add form in detail pane.
- `New agent` clears form.
- `Inspect/Edit` loads agent details into form.
- Save uses POST for new agents and PUT for existing agents.
- Delete uses existing DELETE and refreshes list.

## Compatibility

Existing agent file layout stays unchanged. Existing list/get/delete APIs remain compatible.
