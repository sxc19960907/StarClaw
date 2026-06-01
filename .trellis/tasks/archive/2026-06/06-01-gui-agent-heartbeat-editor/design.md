# Design

## Data Flow

Web UI form fields map to the existing `agents.HeartbeatConfig` YAML shape:

```json
{
  "heartbeat_every": "15m",
  "heartbeat_active_hours": "09:00-17:00",
  "heartbeat_model": "optional-model"
}
```

The daemon agent create/update request translates these fields to:

```yaml
heartbeat:
  every: 15m
  active_hours: 09:00-17:00
  model: optional-model
```

The response remains the loaded `agents.Agent`.

## Backend

- Add heartbeat fields to `agentEditRequest`.
- Update `buildAgentConfig` to create `agents.HeartbeatConfig` only when `heartbeat_every` is non-empty.
- Keep `isEmptyAgentConfig` behavior so clearing heartbeat and all other config removes `config.yaml`.
- Reuse existing agent name and prompt validation.

## Frontend

- Add a Heartbeat fieldset in the existing Agents editor.
- Fill fields from `agent.Config.Heartbeat` or `agent.config.heartbeat`.
- Send heartbeat fields in the existing create/update payload.
- Do not validate duration format in the browser; heartbeat manager already handles invalid duration by skipping/logging.

## Compatibility

Existing agents without heartbeat config load with empty heartbeat fields. Existing heartbeat configs retain the same YAML shape.
