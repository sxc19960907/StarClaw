# Implementation

## Steps

- [x] Add per-agent debounce generation tracking to `Watcher`.
- [x] Capture generation when creating debounce timer callbacks.
- [x] Make flush generation-aware and ignore stale callbacks.
- [x] Invalidate pending generations on `Close`.
- [x] Add focused unit coverage for stale callback suppression.
- [x] Run watcher tests.
- [x] Run full repository tests.
- [x] Run whitespace check.

## Validation

```bash
go test ./internal/watcher
go test ./...
git diff --check
```
