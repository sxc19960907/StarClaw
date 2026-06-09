#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="Astria"
SRC_DIR="$ROOT_DIR/desktop/macos/Astria"
BUILD_DIR="${ASTRIA_BUILD_DIR:-$ROOT_DIR/build/desktop/macos}"
APP_DIR="$BUILD_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

fail() {
  echo "build_macos_astria_shell: $*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Darwin" ]]; then
  fail "macOS is required to build the Astria shell"
fi

command -v swiftc >/dev/null 2>&1 || fail "missing required command: swiftc"

rm -rf "$APP_DIR"
mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

swiftc \
  -parse-as-library \
  -target "$(uname -m)-apple-macosx13.0" \
  -framework SwiftUI \
  -framework WebKit \
  -o "$MACOS_DIR/$APP_NAME" \
  "$SRC_DIR/Sources/"*.swift

cp "$SRC_DIR/Info.plist" "$CONTENTS_DIR/Info.plist"
chmod +x "$MACOS_DIR/$APP_NAME"

echo "$APP_DIR"
