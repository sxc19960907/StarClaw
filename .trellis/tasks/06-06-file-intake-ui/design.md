# File Intake UI Design

## Scope

Add a first-class Astria file intake surface for local document and archive workflows. The MVP accepts local paths and reuses the existing document/archive tools.

## Backend Contract

Add `POST /intake/file` for read-only intake operations:

```json
{
  "path": "docs/report.docx",
  "mode": "document_text",
  "max_chars": 12000,
  "max_entries": 80
}
```

Allowed modes:

- `document_text`: runs the existing `document_text` tool.
- `archive_inspect`: runs the existing `archive_inspect` tool.
- `auto`: chooses archive inspection for `.zip`, `.tar`, `.tar.gz`, `.tgz`; otherwise chooses document text.

Response:

```json
{
  "mode": "document_text",
  "path": "docs/report.docx",
  "status": "ok",
  "content": "...",
  "is_error": false
}
```

The endpoint must not execute `archive_extract`, because extraction writes files and should continue to flow through the normal agent approval path.

## UI Contract

Add a File Intake panel in Astria and a Home/Manage entry. The panel includes:

- Local path input.
- Mode selector: Auto, Document text, Archive inspect.
- Analyze button.
- Result preview with clear error state.
- Send to Chat button that seeds a normal chat/run prompt using the selected file path and latest intake result.
- Archive extraction prompt helper that seeds a run asking the agent to call `archive_extract` with approval.

## Compatibility

- Static Web UI only: no new frontend dependency or build tool.
- Existing document/archive tools remain the source of behavior.
- Existing tool tests remain authoritative for format parsing and extraction safety.

## Validation

- Router test includes the new route.
- Server tests cover document text, archive inspect, auto mode, invalid mode, and tool errors.
- Web UI smoke covers local path analysis and sending an intake prompt to chat.
