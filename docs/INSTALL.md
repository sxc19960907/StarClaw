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

npm distribution is not published yet. The repository contains npm packaging scaffolding, but its installer is not a supported release path yet.

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
starclaw app --check
```

`starclaw app --check` prints the GUI launch command, Web UI URL, diagnostics URL, data directory, and whether the daemon is already running.

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
starclaw app --check
```

## Uninstallation

### Binary

```bash
rm $(which starclaw)
rm -rf ~/.starclaw
```

### Homebrew

No Homebrew formula is published yet, so there is nothing to uninstall through Homebrew.

### npm

npm distribution is not published yet. If you installed from local npm packaging during development, uninstall that package from the same npm environment.

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
