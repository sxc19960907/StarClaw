# Astria macOS Shell

This directory contains the first standalone Astria desktop shell skeleton.

The shell is intentionally thin in Phase13:

- It hosts the existing daemon-served Web UI at `http://127.0.0.1:7533/app/`.
- It does not replace the Astria Web UI.
- It starts or attaches to the local StarClaw daemon through the existing
  `/health` readiness contract.
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
