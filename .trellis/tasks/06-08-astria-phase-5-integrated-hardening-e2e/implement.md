# Astria Phase 5 integrated hardening and E2E validation implementation plan

## Checklist

1. Create Phase 5 parent task and planning artifacts.
2. Create child tasks in the specified order.
3. For each child:
   - write PRD/design/implement as needed
   - start only after planning review
   - run `trellis-before-dev` before edits
   - run targeted validation and `go test ./...` when code changes land
   - commit, archive, and journal independently
4. After all validation/hardening children finish, complete `phase5-kocoro-gap-audit`.
5. Update parent acceptance criteria, archive parent, and record final Phase 5 journal.

## Initial Child Task Order

1. `phase5-runtime-e2e-smoke`
2. `phase5-api-observability-smoke`
3. `phase5-secret-leakage-regression`
4. `phase5-webui-bug-bash`
5. `phase5-docs-current-capabilities`
6. `phase5-kocoro-gap-audit`

## Validation Commands

Baseline commands expected across children:

```bash
go test ./internal/daemon ./cmd
go test ./...
git diff --check
```

Browser smoke children should use existing local Web UI smoke patterns or Playwright where applicable.

## Risk Files

- `internal/daemon/server.go`
- `internal/daemon/openai_api.go`
- `internal/daemon/run_store.go`
- `internal/daemon/trace_export.go`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/*_test.go`
- `README.md`
- `docs/`

## Non-Goals

- Do not implement Phase 6 features in Phase 5.
- Do not introduce remote telemetry.
- Do not add a frontend build chain.
