# Design

## Backend

Reuse the existing `/config` endpoint family, but make it YAML-aware and safer for GUI consumption.

### `GET /config`

Return a structured response:

```json
{
  "config": {
    "provider": "anthropic",
    "endpoint": "https://api.anthropic.com",
    "model_tier": "medium",
    "openai_endpoint": "https://api.openai.com/v1",
    "openai_model": "gpt-4o",
    "ollama_endpoint": "http://localhost:11434",
    "ollama_model": "llama3.1",
    "api_key_set": false,
    "openai_api_key_set": false
  }
}
```

Do not return `api_key` or `openai_api_key` values. The response only reports whether a key exists.

### `PATCH /config`

Accept a JSON patch for the same provider-level scalar fields. Merge into the YAML config file and write YAML back to `ConfigPath` with `0600` permissions.

After writing, reload the config from `ConfigPath` into `s.deps.Config` so `/diagnostics` reflects changes in the same daemon process.

Validation:

- reject unsupported provider values
- ignore or reject unknown config keys; prefer rejection so GUI bugs are visible
- preserve existing API keys when omitted from the patch
- allow API key updates only when non-empty values are supplied

## Frontend

Add a `Config` nav item/panel and make Diagnostics actions route users there.

Panel behavior:

- load `/config` during `refreshAll()`
- provider select toggles provider-specific field groups
- API key fields use password inputs and placeholder text that says whether a key is already set
- submitting sends only fields relevant to the selected provider
- blank key fields are omitted from the patch
- after save, reload `/config` and `/diagnostics`

## Compatibility

- Existing non-GUI clients calling `PATCH /config` with top-level JSON fields continue to work for provider-level scalar fields.
- Config comments may not be preserved when the daemon writes YAML. This is acceptable for the focused repair path but should be kept to scalar provider fields only.
- No paid provider probes are introduced.

## Rollback

The backend config handler can be reverted independently from the GUI panel. The GUI panel depends on structured `/config`, so rollback should remove the panel if the backend behavior is reverted.
