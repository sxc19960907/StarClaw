# Standalone desktop shell plan implementation plan

## Checklist

1. Record the shell choice and trade-offs in the task artifacts.
2. Inspect current CLI app launch and release docs for compatibility points.
3. Create a minimal macOS app skeleton only if the implementation phase is
   approved.
4. Add local developer build notes or a smoke script if a skeleton is created.
5. Validate that existing CLI app launch tests still pass.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-standalone-desktop-shell-plan`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`

## Risk Points

- Do not add a second Web UI implementation.
- Do not require Xcode signing credentials to run ordinary Go tests.
- Keep the native shell optional while the CLI remains the stable fallback.
