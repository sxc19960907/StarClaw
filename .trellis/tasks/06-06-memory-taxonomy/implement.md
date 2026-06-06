# Memory Taxonomy Implementation Plan

## Checklist

1. Extend memory API response with parsed facts, category counts, and warnings.
2. Add unit tests for category parsing, duplicate detection, and conflict detection.
3. Add Memory Map UI category filter, warning list, and candidate classifier.
4. Update smoke coverage for category UI visibility.
5. Run validation.

## Validation Commands

```bash
gofmt -w internal/daemon/memory_api.go internal/daemon/server_test.go
go test ./internal/daemon
node --check internal/daemon/webui/assets/app.js
git diff --check
./scripts/smoke_webui_core.sh
go test ./...
```
