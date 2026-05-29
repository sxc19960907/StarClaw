# Implementation

## Steps

- [x] Run full test suite.
- [x] Run targeted race tests.
- [x] Run local build.
- [x] Run cross-platform builds.
- [x] Run CLI smoke checks.
- [x] Fix or document any failures.
- [x] Run whitespace check if files change.

## Validation Commands

```bash
go test ./...
go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon ./internal/tools ./internal/heartbeat
make build
make build-all
./starclaw version
./starclaw --help
./starclaw completion zsh
./starclaw mcp --help
```
