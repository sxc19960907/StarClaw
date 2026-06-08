# Image tool provider boundary implementation plan

## Steps

1. Read task artifacts and applicable backend specs.
2. Add `internal/images/client.go` and focused client tests.
3. Add `internal/tools/generate_image.go` and `internal/tools/edit_image.go`.
4. Add `RegisterImageTools(reg, client)` without changing `RegisterLocalTools`.
5. Add tool tests for validation, error mapping, safe checker, approval, output formatting, and registration boundaries.
6. Run:
   - `gofmt` on touched Go files
   - `go test ./internal/images ./internal/tools`
   - `go test ./internal/daemon ./internal/images ./internal/tools`
7. Run Trellis validation.
8. Commit the work changes, archive the child task, and verify git status is clean.

## Review gates

- No default provider credentials or config changes.
- No provider-backed image tools in default local registry.
- No real network tests; use `httptest` and fake clients.
- No public upload behavior beyond returned provider URLs.
