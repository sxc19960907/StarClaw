# Astria macOS Shell

This directory contains the first standalone Astria desktop shell skeleton.

The shell is intentionally thin in Phase13:

- It hosts the existing daemon-served Web UI at `http://127.0.0.1:7533/app/`.
- It does not replace the Astria Web UI.
- It starts or attaches to the local StarClaw daemon through the existing
  `/health` readiness contract.
- When Astria starts the daemon itself, it passes explicit Desktop RPC socket
  and pidfile paths, then validates `system.capabilities` before treating the
  desktop handshake as ready.
- It restores the last same-origin `/app` route and reloads the WebView after
  daemon health recovers.
- It does not require signing, notarization, or cloud credentials for local
  development builds.

Deeper Kocoro-style pidfile/socket reconciliation is intentionally deferred to
a later parity step.

## Local Build

From the repository root:

```bash
scripts/build_macos_astria_shell.sh
```

The script builds an unsigned app bundle under `build/desktop/macos/` and prints
the resulting `.app` path.

Build with a bundled local daemon binary:

```bash
go build -o build/starclaw ./main.go
ASTRIA_BUNDLED_STARCLAW_BIN="$PWD/build/starclaw" scripts/build_macos_astria_shell.sh
```

The bundled daemon is copied to `Astria.app/Contents/Resources/starclaw`. At
runtime, `ASTRIA_STARCLAW_BIN` still takes precedence for development overrides;
otherwise Astria tries the bundled daemon before falling back to `PATH`.

Set app version metadata for release-candidate smoke builds:

```bash
ASTRIA_APP_VERSION=0.1.0 ASTRIA_APP_BUILD=1 scripts/build_macos_astria_shell.sh
```

For local testing:

```bash
open build/desktop/macos/Astria.app
```

Override the hosted URL during development:

```bash
ASTRIA_WEB_URL=http://127.0.0.1:7533/app/ open build/desktop/macos/Astria.app
```

Point the app at a development daemon binary:

```bash
ASTRIA_STARCLAW_BIN=/path/to/starclaw open build/desktop/macos/Astria.app
```

Override the Desktop RPC runtime directory during smoke testing:

```bash
ASTRIA_RUNTIME_DIR=/tmp/astria-runtime open build/desktop/macos/Astria.app
```

By default, Astria uses
`~/Library/Application Support/dev.starclaw.astria/daemon.sock` and
`~/Library/Application Support/dev.starclaw.astria/daemon.pid`. Existing
HTTP-only daemons can still be attached as a fallback; daemon instances launched
by Astria must pass Desktop RPC capability reconciliation.

Before launching its own daemon, Astria removes stale `daemon.sock` and
`daemon.pid` only inside this runtime directory. If a healthy daemon was started
outside Astria and has no usable Desktop RPC socket, the shell keeps the Web UI
open over HTTP and shows a degraded fallback banner.

The route recovery store persists only relative `/app` routes under
`astria.lastWebRoute`; unsafe or external stored routes fall back to `/app/`.

## Signing and Updates

Local builds are intentionally unsigned. A distributable release build must use
a Developer ID Application identity, Hardened Runtime, notarization, and
stapling outside the default Linux release workflow. Do not commit signing
identities, notarization credentials, keychain profiles, or update private keys.

Astria does not auto-update itself in this phase. Future updater metadata must
be checksum/signature verified, unavailable-safe, and must not replace the app
with a daemon version that is incompatible with the app release.

Run the local distribution boundary validator on macOS before release work:

```bash
scripts/validate_release_artifacts.sh --npm-only --astria-local
```

This does not require Apple credentials. It confirms the unsigned app smoke
passes, private signing/notarization material is absent from the repository, and
Astria updater metadata is either absent or conforms to the signed JSON boundary:
`version`, `artifact_url`, `checksum_sha256`, `signature`,
`signature_algorithm`, `public_key_id`, `min_app_version`,
`min_daemon_version`, and `unavailable_safe=true`. Metadata must not enable app
replacement until a verified updater implementation exists.

To validate the updater decision path without building release artifacts:

```bash
scripts/validate_release_artifacts.sh --updater-dry-run-smoke
```

The dry-run decision keeps `replacement="disabled"` even for valid metadata.
