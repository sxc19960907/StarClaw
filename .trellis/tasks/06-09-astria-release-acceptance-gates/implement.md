# Astria release acceptance gates implementation plan

## Steps

1. Add release acceptance manifest writer and validator in
   `scripts/validate_release_artifacts.sh`.
2. Add `--astria-release-acceptance-gates-smoke`.
3. Run the smoke as part of `--npm-only --astria-local`.
4. Update Astria README and backend spec.
5. Validate with focused smoke, local release validation, macOS smoke, Go tests,
   and diff checks.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-release-acceptance-gates`
- `scripts/validate_release_artifacts.sh --astria-release-acceptance-gates-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`
