# Design

## Backend

Add a small diagnostics module under `internal/daemon/diagnostics.go`.

Response shape:

```json
{
  "status": "needs_setup",
  "summary": "Provider setup is incomplete.",
  "checks": [
    {
      "id": "provider",
      "label": "Provider",
      "status": "needs_setup",
      "detail": "Anthropic API key is missing.",
      "action": "Run starclaw setup or set api_key."
    }
  ]
}
```

Statuses:

- `ready`
- `warning`
- `needs_setup`
- `error`

Overall status is the highest-severity check:

`error` > `needs_setup` > `warning` > `ready`.

## Check Strategy

- Config exists: `os.Stat(deps.ConfigPath)`.
- Provider setup:
  - `anthropic`: `api_key` non-empty.
  - `openai`: `openai_api_key` non-empty and model/endpoint non-empty.
  - `ollama`: endpoint/model non-empty; short HTTP GET `/api/tags` best-effort.
- Storage: mkdir/stat write probe under StarClaw dir and sessions dir.
- Schedule manager: non-nil.
- Tools: registry non-nil and tool count > 0.
- Permissions: warn when config permissions are nil.

## Frontend

Use existing Web UI patterns:

- topbar gets compact diagnostics chip
- add a `Diagnostics` nav item/panel
- panel renders checks as rows with status tags and action text
- `refreshAll()` loads diagnostics alongside existing data

## Tests

- Unit tests against `handleDiagnostics`.
- Smoke script checks diagnostics nav/panel renders.

## Compatibility

Diagnostics is read-only. No config mutation, no paid provider calls.
