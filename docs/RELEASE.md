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

If GoReleaser is installed, validate the full artifact set:

```bash
scripts/validate_release_artifacts.sh --snapshot
```

This verifies:

- platform archives for macOS, Linux, and Windows
- Linux `.deb`, `.rpm`, and `.apk` package artifacts
- npm package contents from `npm pack --dry-run`

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
