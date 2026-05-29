# Installation Guide

## System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.22 or later (for building from source)
- **Memory**: 50MB RAM minimum
- **Disk**: 20MB free space

## Installation Methods

### Method 1: Pre-built Binary (Recommended)

#### Linux/macOS

```bash
curl -sSL https://get.starclaw.dev | sh
```

#### Windows (PowerShell)

```powershell
iwr -useb https://get.starclaw.dev/windows | iex
```

### Method 2: Package Managers

#### Homebrew (macOS/Linux)

Homebrew distribution is not available yet. Use the pre-built release binaries, Go, or npm for now.

#### npm (Cross-platform)

```bash
npm install -g @starclaw/cli
```

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
```

## Configuration

After installation, run the setup wizard:

```bash
starclaw setup
```

You'll be prompted for:
- API Endpoint (default: https://api.anthropic.com)
- API Key (from your LLM provider)
- Model preferences

## Uninstallation

### Binary

```bash
rm $(which starclaw)
rm -rf ~/.starclaw
```

### Homebrew

No Homebrew formula is published yet, so there is nothing to uninstall through Homebrew.

### npm

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

### Windows Defender Warning

Click "More info" → "Run anyway" if Windows Defender blocks the executable.
