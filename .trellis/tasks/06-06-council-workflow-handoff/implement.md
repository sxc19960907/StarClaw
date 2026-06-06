# Council Workflow Handoff Implementation Plan

## Checklist

1. Add handoff request type and `POST /council/{id}/run` handler.
2. Register route and update route tests.
3. Add backend test proving synthesis starts a normal run and preserves source/channel.
4. Add Council UI `Start run` action.
5. Add click handler and smoke coverage.
6. Run validation.

## Validation Commands

```bash
gofmt -w internal/daemon/council_api.go internal/daemon/router.go internal/daemon/router_test.go internal/daemon/council_api_test.go
go test ./internal/daemon
node --check internal/daemon/webui/assets/app.js
git diff --check
./scripts/smoke_webui_core.sh
go test ./...
```
