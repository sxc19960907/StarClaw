# Phase 5 docs current capabilities implementation plan

## Checklist

1. Load relevant specs with `trellis-before-dev`.
2. Inspect README and docs for current runtime/API/Web UI/safety descriptions.
3. Update concise docs for validated Phase 5 platform capabilities.
4. Run docs-relevant checks:
   - `go test ./internal/daemon ./cmd`
   - `go test ./...`
   - `git diff --check`
5. Update PRD acceptance criteria, commit, archive child, and record journal.

## Risk Files

- `README.md`
- `docs/`
- Any current capability/reference markdown files discovered during implementation.

## Non-Goals

- No hosted/cloud documentation.
- No broad marketing rewrite.
- No docs for unimplemented Phase 6 ideas.
