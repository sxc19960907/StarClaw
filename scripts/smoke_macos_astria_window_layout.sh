#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
BUILD_DIR="$TMP_DIR/build/desktop/macos"
BASE_URL="${APP_SMOKE_BASE_URL:-http://127.0.0.1:7533}"
APP_PATH=""

cleanup() {
  osascript -e 'tell application "Astria" to quit' >/dev/null 2>&1 || true
  curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_macos_astria_window_layout: $*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "smoke_macos_astria_window_layout: skipped (macOS required)"
  exit 0
fi

command -v osascript >/dev/null 2>&1 || fail "missing required command: osascript"

echo "==> building StarClaw for Astria window layout smoke"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

APP_PATH="$(ASTRIA_BUILD_DIR="$BUILD_DIR" ASTRIA_BUNDLED_STARCLAW_BIN="$BIN" "$ROOT_DIR/scripts/build_macos_astria_shell.sh")"

osascript -e 'tell application "Astria" to quit' >/dev/null 2>&1 || true
curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
sleep 1

echo "==> launching Astria"
open -n "$APP_PATH"

for _ in {1..80}; do
  if curl -fsS "$BASE_URL/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done
curl -fsS "$BASE_URL/health" >/dev/null || fail "Astria did not start a healthy daemon"

for _ in {1..80}; do
  if osascript <<'APPLESCRIPT' >/dev/null 2>&1
tell application "System Events"
  tell process "Astria"
    if (count of windows) < 1 then error "Astria window not ready"
  end tell
end tell
APPLESCRIPT
  then
    break
  fi
  sleep 0.2
done

echo "==> checking minimum desktop window size"
window_size="$(
  osascript <<'APPLESCRIPT'
tell application "System Events"
  tell process "Astria"
    set frontmost to true
    set size of window 1 to {900, 600}
    delay 0.2
    set windowSize to size of window 1
  end tell
end tell
return (item 1 of windowSize as text) & "x" & (item 2 of windowSize as text)
APPLESCRIPT
)" || {
  if [[ "${ASTRIA_WINDOW_LAYOUT_REQUIRE_GUI:-}" == "1" ]]; then
    fail "could not inspect Astria window size"
  fi
  echo "smoke_macos_astria_window_layout: skipped (GUI scripting unavailable)"
  exit 0
}

width="${window_size%x*}"
height="${window_size#*x}"

[[ "$width" =~ ^[0-9]+$ ]] || fail "unexpected window width: $window_size"
[[ "$height" =~ ^[0-9]+$ ]] || fail "unexpected window height: $window_size"
(( width >= 1400 )) || fail "window width shrank below desktop layout minimum: $window_size"
(( height >= 860 )) || fail "window height shrank below desktop layout minimum: $window_size"

echo "==> checking window can still expand"
expanded_size="$(
  osascript <<'APPLESCRIPT'
tell application "System Events"
  tell process "Astria"
    set size of window 1 to {1480, 920}
    delay 0.2
    set windowSize to size of window 1
  end tell
end tell
return (item 1 of windowSize as text) & "x" & (item 2 of windowSize as text)
APPLESCRIPT
)" || fail "could not resize Astria window upward"

expanded_width="${expanded_size%x*}"
expanded_height="${expanded_size#*x}"
[[ "$expanded_width" =~ ^[0-9]+$ ]] || fail "unexpected expanded window width: $expanded_size"
[[ "$expanded_height" =~ ^[0-9]+$ ]] || fail "unexpected expanded window height: $expanded_size"
(( expanded_width > width )) || fail "window did not expand beyond minimum desktop width: $expanded_size"
(( expanded_height > height )) || fail "window did not expand beyond minimum desktop height: $expanded_size"

echo "smoke_macos_astria_window_layout: ok"
