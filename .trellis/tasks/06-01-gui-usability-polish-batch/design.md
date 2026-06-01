# Design

## Backend

### Sessions

Add `PATCH /sessions/{id}` for small metadata updates:

```json
{
  "title": "New title",
  "favorite": true
}
```

Rules:

- validate session id with the same path safety as get/delete
- title is optional, trimmed, must be non-empty when supplied
- favorite is optional pointer bool
- update only supplied fields
- return updated session summary or full session

Implementation can load the session through `session.Manager.Resume`, mutate fields, and call `Save`. This is acceptable for the daemon API because it creates a short-lived manager per request.

### Permissions

Change `GET /permissions` from an empty placeholder to a read-only summary:

```json
{
  "permissions": {
    "configured": true,
    "allowed_dirs": ["~", "."],
    "allowed_commands": [],
    "denied_commands": [],
    "network_allowlist": [],
    "sensitive_patterns": []
  }
}
```

Do not add write behavior in this batch.

## Frontend

### Activity readability

Keep the current transcript DOM but improve row metadata and visual states:

- show active session label near composer
- tool events use consistent status tags
- approval cards remain inline and actionable
- error/system messages use distinct styling

### Session operations

Add controls to each session row:

- `Rename` prompts for a new title and calls `PATCH /sessions/{id}`
- favorite toggle calls `PATCH /sessions/{id}` with `favorite`
- delete asks `confirm()` before calling DELETE

### Permissions panel

Add nav item and panel. Render permission categories as compact rows and use empty states when config has no policy.

### Diagnostics actions

Render check actions as buttons when a check has a clear destination:

- `provider` -> Config panel
- `permissions` -> Permissions panel
- other checks -> text-only action

## Compatibility

All additions are same-origin JSON and keep existing endpoints compatible. `DELETE /sessions/{id}` and current list/get/search behavior are unchanged.
