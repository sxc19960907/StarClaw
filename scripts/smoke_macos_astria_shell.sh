#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

fail() {
  echo "smoke_macos_astria_shell: $*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "smoke_macos_astria_shell: skipped (macOS required)"
  exit 0
fi

app_path="$("$ROOT_DIR/scripts/build_macos_astria_shell.sh")"

[[ -d "$app_path" ]] || fail "missing app bundle: $app_path"
[[ -x "$app_path/Contents/MacOS/Astria" ]] || fail "missing executable"
[[ -f "$app_path/Contents/Info.plist" ]] || fail "missing Info.plist"

/usr/libexec/PlistBuddy -c "Print :CFBundleIdentifier" "$app_path/Contents/Info.plist" | grep -Fxq "dev.starclaw.astria" || fail "unexpected bundle identifier"
/usr/libexec/PlistBuddy -c "Print :NSAppTransportSecurity:NSAllowsLocalNetworking" "$app_path/Contents/Info.plist" | grep -Fxq "true" || fail "local networking not enabled"

echo "smoke_macos_astria_shell: ok"
