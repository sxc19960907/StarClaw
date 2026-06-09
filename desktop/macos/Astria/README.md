# Astria macOS Shell

This directory contains the first standalone Astria desktop shell skeleton.

The shell is intentionally thin in Phase13:

- It hosts the existing daemon-served Web UI at `http://127.0.0.1:7533/app/`.
- It does not replace the Astria Web UI.
- It does not supervise or start the daemon yet.
- It does not require signing, notarization, or cloud credentials for local
  development builds.

The next Phase13 child task adds daemon supervision and launch/attach behavior.

## Local Build

From the repository root:

```bash
scripts/build_macos_astria_shell.sh
```

The script builds an unsigned app bundle under `build/desktop/macos/` and prints
the resulting `.app` path. The app expects a StarClaw daemon to be reachable.

For local testing:

```bash
starclaw app --no-open
open build/desktop/macos/Astria.app
```

Override the hosted URL during development:

```bash
ASTRIA_WEB_URL=http://127.0.0.1:7533/app/ open build/desktop/macos/Astria.app
```
