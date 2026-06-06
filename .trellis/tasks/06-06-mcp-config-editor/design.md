# MCP Config Editor Design

## Scope

Add MCP server management to the existing Astria daemon UI and `/config` persistence path. The product behavior is:

- List configured MCP servers without leaking env values.
- Add a stdio server.
- Edit an existing server's transport fields, args, env keys, context, keep-alive, and disabled state.
- Preserve existing env secret values when the submitted env value is blank.
- Keep `POST /mcp/test` as the connection test endpoint after saves.

## Backend Contract

Extend `PATCH /config` with optional `mcp_servers`.

```json
{
  "mcp_servers": [
    {
      "name": "browser",
      "type": "stdio",
      "command": "npx",
      "args": ["@playwright/mcp@latest"],
      "env": {"BROWSER_TOKEN": ""},
      "context": "browser automation",
      "keep_alive": true,
      "disabled": false
    }
  ]
}
```

Semantics:

- Omitted `mcp_servers` preserves existing MCP config.
- Present `mcp_servers` replaces the managed MCP server map.
- Server names are normalized by trimming whitespace and must be simple config keys.
- `type` defaults to `stdio`; allowed values are `stdio` and `http`.
- `stdio` requires `command`.
- `http` requires `url`.
- Env keys are trimmed and empty keys are rejected.
- Blank env values preserve the existing value for the same server/key; if no existing value exists, the blank env value is omitted.
- GET `/config` continues returning `env_keys`, never env values.

## UI Contract

The MCP Starport detail pane owns the editor. The list cards expose:

- Test connection.
- Edit.
- Disable/Enable.

The detail pane exposes:

- Add stdio server.
- Save dock.
- Clear form.

The form posts through `PATCH /config` and then re-renders from the returned redacted config. UI errors use the existing toast/inline status path and should show backend validation messages.

## Compatibility

No new frontend build tool or dependency is introduced. The existing static `index.html`, `app.js`, and `styles.css` carry the UI.

## Validation

- Backend tests cover add/edit/disable, env preservation, redaction, and invalid config.
- Web UI smoke covers add/edit/disable controls without requiring a real MCP server process.
