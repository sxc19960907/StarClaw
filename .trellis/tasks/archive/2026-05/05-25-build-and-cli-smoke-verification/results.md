# Results

## Build Verification

- `make build` passed.
  - Rebuilt ignored root binary: `starclaw`.
  - Found and fixed version injection: the binary previously printed `dev` because `main.Version` was not set by ldflags.
- `make build-all` passed.
  - Generated ignored artifacts under `dist/`:
    - `starclaw-darwin-amd64`
    - `starclaw-darwin-arm64`
    - `starclaw-linux-amd64`
    - `starclaw-linux-arm64`

## CLI Smoke Verification

- `./starclaw version` passed and now prints the injected dirty git version.
- `./starclaw --help` passed.
- `./starclaw completion zsh` passed.
- `./starclaw mcp --help` passed.
- `./dist/starclaw-darwin-arm64 version` passed.

## Fixes Made

- Updated `Makefile` ldflags to set both `main.Version` and `cmd.Version`.
- Updated `.goreleaser.yaml` ldflags to set `main.Version` as well as `cmd.Version`.
- Updated `RELEASE_CHECKLIST.md` with build and CLI smoke commands.

## Final Verification

- `go test ./...` passed.

## Artifact Notes

- Root `starclaw` and `dist/` outputs are ignored by `.gitignore`.
- No release tag was created.
