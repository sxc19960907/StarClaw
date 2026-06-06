# Document and Archive Tools Pack

## Goal

Add practical local-file understanding tools inspired by Kocoro's document/archive capabilities. Astria should be able to inspect common documents and archives without requiring the user to manually convert files first.

## Requirements

- Add tool support for:
  - PDF text extraction.
  - DOCX text extraction.
  - XLSX sheet/text extraction.
  - PPTX slide text extraction.
  - Archive inspection.
  - Archive extraction with safe destination handling.
- Tools must follow existing StarClaw tool registration, permission, and test patterns.
- Extraction must avoid unsafe path traversal and unexpected overwrites.
- Tool outputs should be structured enough for agent use, not just raw logs.
- Large or binary files should fail with clear bounded errors.

## Acceptance Criteria

- [x] Each supported file type has at least one unit test with a fixture or generated sample.
- [x] Archive extraction rejects path traversal entries.
- [x] Tool descriptions are discoverable to the agent.
- [x] Permission behavior matches existing local file/tool conventions.
- [x] Relevant package tests pass with `go test`.

## Non-Goals

- No OCR in this task.
- No cloud document connectors.
- No full document rendering UI.
- No editing/writing Office documents.

## Dependencies

- Requires inspection of current tool registry and file permission conventions before implementation.
- May inform later Home Actions and Memory Map work by providing file-ingestion primitives.
