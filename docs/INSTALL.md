# Installation Guide

## System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.22 or later (for building from source)
- **Memory**: 50MB RAM minimum
- **Disk**: 20MB free space

## Installation Methods

### Method 1: GitHub Release Binary (Recommended)

Download the archive for your platform from [GitHub Releases](https://github.com/starclaw/starclaw/releases).

#### macOS/Linux archive example

```bash
# Pick the archive that matches your OS and CPU.
# Examples:
#   starclaw_Darwin_arm64.tar.gz
#   starclaw_Darwin_x86_64.tar.gz
#   starclaw_Linux_arm64.tar.gz
#   starclaw_Linux_x86_64.tar.gz
curl -LO https://github.com/starclaw/starclaw/releases/latest/download/starclaw_Darwin_arm64.tar.gz
tar -xzf starclaw_Darwin_arm64.tar.gz
chmod +x starclaw
sudo mv starclaw /usr/local/bin/
```

#### Windows archive example

Download `starclaw_Windows_x86_64.zip` or `starclaw_Windows_arm64.zip` from GitHub Releases, unzip it, and add the extracted directory to `PATH`.

Linux package artifacts are also published as `.deb`, `.rpm`, and `.apk`.

### Method 2: Package Managers

#### Homebrew (macOS/Linux)

Homebrew distribution is not available yet. Use the pre-built release binaries, Go, or npm for now.

#### npm (Cross-platform)

```bash
npm install -g @starclaw/cli
```

The npm package installs a small wrapper and downloads the matching StarClaw binary from GitHub Releases during `postinstall`. It supports macOS, Linux, and Windows on x64/arm64 where release assets are available.

### Method 3: Build from Source

#### Prerequisites

- Go 1.22 or later
- Git

#### Build Steps

```bash
# Clone repository
git clone https://github.com/starclaw/starclaw.git
cd starclaw

# Build
go build -o starclaw .

# Install (optional)
go install .
```

#### Verify Installation

```bash
starclaw version
starclaw doctor
starclaw doctor --json
starclaw app --check
```

`starclaw doctor` prints local readiness checks, the GUI launch command, Web UI URL, diagnostics URL, data directory, config path, and daemon readiness. Use `starclaw doctor --json` for support bundles or scripted checks. `starclaw app --check` is a shorter launch-readiness check.

## Configuration

After installation, run the setup wizard:

```bash
starclaw setup
```

You'll be prompted for:
- API Endpoint (default: https://api.anthropic.com)
- API Key (from your LLM provider)
- Model preferences

## Launch the GUI

Start the daemon if needed and open the embedded Web UI:

```bash
starclaw app
```

If a daemon is already running on the default local port, `starclaw app` reuses it and prints the same Web UI URL instead of starting another daemon.

If you are on a remote or headless machine, start/reuse the daemon without opening a browser:

```bash
starclaw app --no-open
```

Then open the printed Web UI URL from a browser that can reach the machine.

Check launch readiness without starting the daemon or opening a browser:

```bash
starclaw app --check
```

This prints the launch command, daemon running state, Web UI URL, diagnostics URL, and local data directory. In the Web UI, open **Settings -> Version** to see the same runtime context plus health, status, diagnostics, data, and config paths.

If the daemon starts but the browser cannot open, use the printed Web UI URL manually. If startup fails, check whether port `7533` is already in use and run:

```bash
starclaw daemon status
starclaw doctor
starclaw app --check
```

Advanced native desktop integration can start the daemon with a paired Desktop
RPC socket and pidfile:

```bash
starclaw daemon start --rpc-socket /tmp/starclaw/daemon.sock --rpc-pidfile /tmp/starclaw/daemon.pid
```

Passing only one of these flags fails before daemon startup. Standard
`starclaw app` and browser workflows do not require them.

### macOS development shell

The repository includes an unsigned Astria macOS shell skeleton for local
development:

```bash
scripts/build_macos_astria_shell.sh
open build/desktop/macos/Astria.app
```

This development shell hosts the daemon-served Web UI. It is not yet a signed
release artifact, but it can start or reuse the local daemon through the same
HTTP readiness contract used by `starclaw app`. It restores only same-origin
`/app` routes and falls back to `/app/` for unsafe stored routes.

For bundled-app smoke testing, build a local daemon and copy it into the app
bundle:

```bash
go build -o build/starclaw ./main.go
ASTRIA_BUNDLED_STARCLAW_BIN="$PWD/build/starclaw" scripts/build_macos_astria_shell.sh
```

The bundled daemon lives at `Astria.app/Contents/Resources/starclaw`. Unsigned
development builds are not notarized release artifacts.

Validate the local Astria distribution boundary without Apple credentials:

```bash
scripts/validate_release_artifacts.sh --npm-only --astria-local
```

This checks the npm package shape, runs the Astria shell smoke, verifies that
private signing/notarization material is not committed, and confirms Astria
updater metadata is either absent or has checksum, signature, release
compatibility, and unavailable-safe fields.

A signed distributable Astria build requires a Developer ID Application
identity, Hardened Runtime, notarization with `notarytool`, stapling, and
release-matched app/daemon inputs. Do not commit signing identities, keychain
profiles, Apple credentials, notarization secrets, or updater private keys.
Missing updater metadata is intentionally non-fatal for local builds.
Updater metadata must not enable app replacement until verified checksum and
signature enforcement exists in the updater implementation.

When the shell starts the daemon itself, it also passes Desktop RPC socket and
pidfile paths under `~/Library/Application Support/dev.starclaw.astria/` and
validates `system.capabilities` before declaring the desktop handshake ready.
Set `ASTRIA_RUNTIME_DIR=/tmp/astria-runtime` to keep these artifacts isolated
for smoke testing. Stale `daemon.sock` and `daemon.pid` files are cleaned only
inside the configured Astria runtime directory; healthy HTTP-only daemons remain
usable through a visible fallback mode.

## Uninstallation

### Binary

```bash
rm $(which starclaw)
rm -rf ~/.starclaw
```

### Homebrew

No Homebrew formula is published yet, so there is nothing to uninstall through Homebrew.

### npm

If installed with npm:

```bash
npm uninstall -g @starclaw/cli
```

## Troubleshooting

### "command not found"

Ensure `~/go/bin` is in your PATH:

```bash
export PATH=$PATH:~/go/bin
```

### Permission Denied (Linux/macOS)

```bash
chmod +x $(which starclaw)
```

### GUI does not open

Use the no-browser launch mode and open the URL manually:

```bash
starclaw app --no-open
```

If the command reports a port conflict, stop the existing daemon or free port `7533`:

```bash
starclaw daemon stop
starclaw app
```

### Windows Defender Warning

Click "More info" → "Run anyway" if Windows Defender blocks the executable.
