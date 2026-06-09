# Packaging signing update boundary implementation plan

## Checklist

1. Extend local macOS shell packaging support.
   - Add a deterministic way to include an optional `starclaw` daemon binary
     under `Astria.app/Contents/Resources`.
   - Preserve current unsigned development builds when no bundled daemon is
     requested.
   - Keep `ASTRIA_STARCLAW_BIN` and `PATH` fallback behavior intact.
2. Tighten smoke validation.
   - Verify app bundle structure and metadata.
   - Verify route recovery and daemon supervision still pass.
   - Add bundled-daemon smoke coverage when a local daemon binary is built.
   - Keep non-macOS smoke behavior as a skip.
3. Update release artifact validation boundary.
   - Preserve existing CLI/npm/archive checks.
   - Document or add optional macOS-only validation for unsigned Astria builds.
   - Do not require Apple signing credentials in default CI.
4. Update documentation.
   - `desktop/macos/Astria/README.md`
   - `docs/INSTALL.md`
   - `docs/RELEASE.md`
   - root `README.md` if user-facing commands change
5. Update code-spec.
   - Record app artifact layout.
   - Record signing/notarization/update boundaries.
   - Record version compatibility and failure modes.
6. Validate.
   - Trellis validation.
   - macOS shell smoke.
   - targeted app/doctor Go tests.
   - full Go tests.
   - diff check.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-packaging-signing-update-boundary`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

If release scripts change:

- `scripts/validate_release_artifacts.sh --npm-only`

## Risk Points

- Do not add committed signing identities, notarization credentials, Apple
  passwords, keychain profiles, or update private keys.
- Do not make Linux CI require macOS-only signing tools.
- Do not make unsigned local app builds look like signed release artifacts.
- Do not remove existing CLI/npm/archive release validation.
- Do not add automatic remote update checks or telemetry defaults.
- Do not bundle a daemon version that conflicts with app release metadata.

## Review Gate

After planning is committed, start the task only after review approval, then
implement the smallest useful packaging boundary that makes Astria's local app
artifact and future signed release contract testable.
