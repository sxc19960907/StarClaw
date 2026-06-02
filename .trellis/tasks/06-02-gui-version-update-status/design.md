# Design

## API

Add two read-only daemon routes:

- `GET /version`
  - Returns current build metadata and static update capability.
  - Does not perform network calls.
- `GET /update/check`
  - Performs a manual update check using the existing `internal/update.CheckForUpdate`.
  - For non-semver versions such as `dev`, returns a no-update response with a development-build message and no network call.

Response shape:

```json
{
  "version": "dev",
  "platform": "darwin/arm64",
  "web_url": "http://127.0.0.1:7533/app/",
  "update_supported": false,
  "update_command": "starclaw update --check",
  "status": "development",
  "message": "Development build - update checks require a release version."
}
```

For release builds, `GET /update/check` can return:

- `status: "available"` with `latest_version`, `release_url`, and `published_at`
- `status: "current"` when already up to date
- `status: "error"` only through HTTP 500 when the check fails

## Web UI

Add a `Version` navigation item and `panel-version`.

The panel displays:

- current version
- platform
- Web UI URL
- whether update checks are supported
- CLI update command
- manual check result

The check button calls `/update/check` and updates the panel state. No automatic update check runs from page load beyond reading `/version`.

## Compatibility

- Existing `/status` remains unchanged.
- No new frontend dependencies.
- Existing CLI update behavior remains unchanged.

## Rollback

Revert the new daemon route file, router entries, Web UI panel changes, and smoke assertions if the API/UI introduces instability.
