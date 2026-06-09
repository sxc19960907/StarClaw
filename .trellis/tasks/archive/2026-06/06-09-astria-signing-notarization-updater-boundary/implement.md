# Astria signing notarization updater boundary implementation plan

## Checklist

1. Add credential/updater boundary checks to `validate_release_artifacts.sh`.
2. Keep `--npm-only --astria-local` credential-free and smoke-backed.
3. Update `docs/INSTALL.md` with signed release prerequisites and local unsigned
   behavior.
4. Update backend/macOS shell spec with distribution boundary rules.
5. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-signing-notarization-updater-boundary`
   - `scripts/validate_release_artifacts.sh --npm-only --astria-local`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Do not require Apple credentials locally.
- Do not add real updater metadata without signature/checksum validation.
- Do not commit private keys, keychain profiles, signing identities, or
  notarization secrets.
