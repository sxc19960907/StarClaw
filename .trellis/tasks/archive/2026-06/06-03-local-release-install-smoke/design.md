# Local Release Install Smoke Design

## Scope

Add:

```bash
scripts/smoke_release_local.sh
```

The script builds a temporary release-style archive for the current OS/architecture, then calls:

```bash
RELEASE_ARCHIVE=<archive> scripts/smoke_release_install.sh
```

## Archive Shape

- Unix platforms: `starclaw_<OS>_<ARCH>.tar.gz`.
- Windows: `starclaw_Windows_<ARCH>.zip`.
- Binary name is `starclaw` or `starclaw.exe`.
- The binary can live at the archive root; the install smoke already searches recursively.

## Version Injection

Use Go ldflags to set:

- `main.Version`
- `github.com/starclaw/starclaw/cmd.Version`

Default local version:

```text
v0.0.0-local
```

Allow override with `LOCAL_RELEASE_VERSION`.

## Paths

Use `mktemp -d` for build and archive output by default. This keeps artifacts outside tracked paths. If debugging is needed later, a separate opt-in preserve mode can be added.

## Dependencies

- Required: `go`, `tar`.
- Required on Windows archive mode: `zip`.

## Compatibility

This complements GoReleaser validation. It does not replace `scripts/validate_release_artifacts.sh`, which still validates cross-platform GoReleaser output.
