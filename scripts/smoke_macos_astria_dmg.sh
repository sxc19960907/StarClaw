#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
MOUNT_DIR="$TMP_DIR/mount"

cleanup() {
  if [[ -d "$MOUNT_DIR" ]]; then
    hdiutil detach "$MOUNT_DIR" -quiet >/dev/null 2>&1 || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_macos_astria_dmg: $*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "smoke_macos_astria_dmg: skipped (macOS required)"
  exit 0
fi

command -v hdiutil >/dev/null 2>&1 || fail "missing required command: hdiutil"

echo "==> building StarClaw for DMG smoke"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

echo "==> building Astria DMG"
dmg_path="$(ASTRIA_BUILD_DIR="$TMP_DIR/build/desktop/macos" ASTRIA_DMG_DIR="$TMP_DIR/dist" ASTRIA_BUNDLED_STARCLAW_BIN="$BIN" "$ROOT_DIR/scripts/build_macos_astria_dmg.sh")"

[[ -f "$dmg_path" ]] || fail "missing DMG: $dmg_path"
[[ -s "$dmg_path" ]] || fail "empty DMG: $dmg_path"

hdiutil imageinfo "$dmg_path" >/dev/null || fail "invalid DMG image"

mkdir -p "$MOUNT_DIR"
hdiutil attach "$dmg_path" -readonly -nobrowse -mountpoint "$MOUNT_DIR" >/dev/null

[[ -d "$MOUNT_DIR/Astria.app" ]] || fail "mounted DMG is missing Astria.app"
[[ -x "$MOUNT_DIR/Astria.app/Contents/MacOS/Astria" ]] || fail "mounted app is missing executable"
[[ -x "$MOUNT_DIR/Astria.app/Contents/Resources/starclaw" ]] || fail "mounted app is missing bundled daemon"
[[ -L "$MOUNT_DIR/Applications" ]] || fail "mounted DMG is missing Applications symlink"
[[ "$(readlink "$MOUNT_DIR/Applications")" == "/Applications" ]] || fail "Applications symlink does not target /Applications"

hdiutil detach "$MOUNT_DIR" -quiet

echo "smoke_macos_astria_dmg: ok"
