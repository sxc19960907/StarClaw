# Implementation Plan

## Checklist

1. Inspect `CHANGELOG.md`, `RELEASE_CHECKLIST.md`, and current git status.
2. Add concise release-note entries for:
   - critical bug-review fixes
   - path / publish / screenshot / grep security hardening
   - process / heartbeat / registry / readtracker concurrency coverage
3. Run final verification:
   - `go test ./...`
   - `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon ./internal/tools ./internal/heartbeat`
4. Record final results and recommended commit grouping.
5. Archive the task.

## Risk Notes

- Keep documentation factual and avoid claiming a tagged release unless one is actually created.
- Do not stage or commit product-code changes unless explicitly requested; Trellis may auto-commit task archival metadata.
