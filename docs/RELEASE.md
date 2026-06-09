# Release Checklist

Use this checklist before tagging a StarClaw release.

## Local Preflight

```bash
go test ./...
go vet ./...
scripts/smoke_app_launch.sh
scripts/smoke_release_local.sh
scripts/smoke_npm_install.sh
scripts/smoke_webui_core.sh
```

On macOS, also validate the unsigned Astria development shell, route recovery,
and daemon supervision smoke:

```bash
scripts/smoke_macos_astria_shell.sh
scripts/validate_release_artifacts.sh --npm-only --astria-local
```

If GoReleaser is installed, validate the full artifact set:

```bash
scripts/validate_release_artifacts.sh --snapshot
```

This verifies:

- platform archives for macOS, Linux, and Windows
- Linux `.deb`, `.rpm`, and `.apk` package artifacts
- npm package contents from `npm pack --dry-run`

## Astria Desktop Boundary

The repository can build and smoke-test an unsigned local Astria macOS app:

```bash
go build -o build/starclaw ./main.go
ASTRIA_BUNDLED_STARCLAW_BIN="$PWD/build/starclaw" \
ASTRIA_APP_VERSION="${VERSION:-0.0.0}" \
ASTRIA_APP_BUILD="${BUILD_NUMBER:-0}" \
scripts/build_macos_astria_shell.sh
scripts/smoke_macos_astria_shell.sh
```

The future signed release artifact shape is `Astria.app` with the launcher at
`Contents/MacOS/Astria` and the bundled daemon at
`Contents/Resources/starclaw`. The app and bundled daemon should come from the
same StarClaw release tag.

Signing and notarization are intentionally outside the default Linux release
workflow. A distributable macOS artifact requires a Developer ID Application
identity, Hardened Runtime, notarization with `notarytool`, and stapling. Do
not commit Apple credentials, keychain profiles, signing identities, or update
private keys.

Astria does not auto-update itself in this phase. Missing update metadata must
be non-fatal, and any future updater must verify checksums/signatures before
replacing the app or bundled daemon.

## Tag Release

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

The GitHub release workflow runs GoReleaser with `GITHUB_TOKEN`.

## npm Package

The npm package is a release-backed installer. It publishes only the wrapper and installer scripts; `postinstall` downloads the matching binary from GitHub Releases.

Before publishing npm, verify the package locally:

```bash
cd npm
npm pack --dry-run
cd ..
scripts/smoke_npm_install.sh
```

Publish manually from the `npm/` directory after the GitHub Release assets are available:

```bash
cd npm
npm publish --access public
```

Do not commit npm tokens or registry credentials. Publishing requires an authenticated npm environment outside the repository.

## Post-Release Checks

After GitHub Release and npm publish:

```bash
starclaw update --check
npm install -g @starclaw/cli
starclaw version
starclaw app --check
```
