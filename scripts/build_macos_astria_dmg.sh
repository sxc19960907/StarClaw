#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="Astria"
BUILD_DIR="${ASTRIA_BUILD_DIR:-$ROOT_DIR/build/desktop/macos}"
DMG_DIR="${ASTRIA_DMG_DIR:-$BUILD_DIR}"
DMG_NAME="${ASTRIA_DMG_NAME:-$APP_NAME}"
DMG_VOLUME_NAME="${ASTRIA_DMG_VOLUME_NAME:-$APP_NAME}"
DMG_PATH="${ASTRIA_DMG_PATH:-$DMG_DIR/$DMG_NAME.dmg}"
APP_DIR="$BUILD_DIR/$APP_NAME.app"
SKIP_APP_BUILD="${ASTRIA_SKIP_APP_BUILD:-}"
STAGING_DIR=""

fail() {
  echo "build_macos_astria_dmg: $*" >&2
  exit 1
}

cleanup() {
  if [[ -n "$STAGING_DIR" ]]; then
    rm -rf "$STAGING_DIR"
  fi
}
trap cleanup EXIT

if [[ "$(uname -s)" != "Darwin" ]]; then
  fail "macOS is required to build the Astria DMG"
fi

command -v hdiutil >/dev/null 2>&1 || fail "missing required command: hdiutil"

if [[ "$SKIP_APP_BUILD" != "1" ]]; then
  APP_DIR="$(ASTRIA_BUILD_DIR="$BUILD_DIR" "$ROOT_DIR/scripts/build_macos_astria_shell.sh")"
fi

[[ -d "$APP_DIR" ]] || fail "missing app bundle: $APP_DIR"
[[ -x "$APP_DIR/Contents/MacOS/$APP_NAME" ]] || fail "missing app executable: $APP_DIR/Contents/MacOS/$APP_NAME"

STAGING_DIR="$(mktemp -d)"
mkdir -p "$DMG_DIR"

cp -R "$APP_DIR" "$STAGING_DIR/$APP_NAME.app"
ln -s /Applications "$STAGING_DIR/Applications"

rm -f "$DMG_PATH"
hdiutil create \
  -volname "$DMG_VOLUME_NAME" \
  -srcfolder "$STAGING_DIR" \
  -ov \
  -format UDZO \
  "$DMG_PATH" >/dev/null

echo "$DMG_PATH"
