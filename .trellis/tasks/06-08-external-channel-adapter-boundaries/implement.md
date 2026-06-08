# Implementation Plan

## Steps

1. Add channel adapter types, fake adapter, and registry in `internal/daemon/channel_adapter.go`.
2. Add registry field and default fake adapter registration in `Server`.
3. Add `GET /channel/adapters` handler and route.
4. Add unit tests for registry and fake install lifecycle.
5. Add API test for adapter list metadata.
6. Run:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-external-channel-adapter-boundaries`
7. Commit and archive the child task.

## Review Gates

- No real network calls.
- No credentials or plaintext secrets.
- API must not imply external transports are active.
- Metadata and install lists must use defensive copies.

## Completion Notes

TBD.
