# Local Release Install Smoke Implementation

## Checklist

- [x] Add `scripts/smoke_release_local.sh`.
- [x] Build current-platform binary with release-like ldflags.
- [x] Package the binary into tar.gz or zip.
- [x] Invoke `scripts/smoke_release_install.sh` with the generated archive.
- [x] Document the command in `RELEASE_CHECKLIST.md`.
- [x] Validate script syntax and run the local smoke script.

## Validation

```bash
bash -n scripts/smoke_release_local.sh scripts/smoke_release_install.sh
scripts/smoke_release_local.sh
```
