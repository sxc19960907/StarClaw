# File Intake UI

## Goal

Expose document and archive tools from the Web UI so users can select or provide local files and ask Astria to inspect, summarize, or extract them without remembering tool names.

## Requirements

- Use existing document/archive tools first.
- Keep operations local and permission-aware.
- Provide clear states for unsupported formats, missing files, and extraction targets.
- Route generated summaries or prompts into normal chat/run workflows.
- Avoid browser-only uploads if local path selection is enough for MVP.

## Acceptance Criteria

- [x] Web UI has a first-class file intake surface or Home action.
- [x] User can provide a local document/archive path.
- [x] UI can request document text or archive inspection through existing tool/run workflow.
- [x] Errors are visible and actionable.
- [x] Tests cover the UI route and relevant tool path.

## Non-Goals

- No cloud upload storage.
- No full file manager.
- No OCR/image/PDF visual rendering unless planned separately.

## Dependencies

- Depends on document/archive tools pack.
