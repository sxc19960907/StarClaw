# Design

## Frontend

Reuse the existing Agents panel editor form. Add a compact permissions section with:

- `tools_allow` textarea using comma or newline separated values.
- `tools_deny` textarea using comma or newline separated values.
- `auto_approve` checkbox.

The frontend already sends these fields to `POST /agents` and `PUT /agents/{name}`. This task focuses on making the controls explicit and validating the round trip in smoke coverage.

## Backend

No new route is required. Existing create/update request fields are the source of truth:

```json
{
  "tools_allow": ["bash", "file_read"],
  "tools_deny": ["http"],
  "auto_approve": true
}
```

If backend persistence has gaps discovered during implementation, fix them in the existing agent API helper without changing the public request shape.

## Compatibility

Existing agent config files remain compatible. Unknown config keys are outside this task.
