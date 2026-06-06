# Real Channel Provider Implementation Plan

## Checklist

1. Add GitHub webhook request parsing and optional HMAC verification.
2. Add daemon routes:
   - `GET /inbox/providers`
   - `POST /inbox/github`
3. Convert supported GitHub events into `InboxItem`.
4. Add backend tests for:
   - issue event ingest
   - issue_comment event ingest
   - duplicate delivery/item handling
   - signature verification success/failure
5. Add Inbox UI provider setup card.
6. Add Web UI smoke coverage for provider status visibility.
7. Run validation.

## Validation Commands

```bash
gofmt -w internal/daemon/inbox_api.go internal/daemon/router.go internal/daemon/router_test.go internal/daemon/server_test.go
go test ./internal/daemon
node --check internal/daemon/webui/assets/app.js
git diff --check
./scripts/smoke_webui_core.sh
go test ./...
```
