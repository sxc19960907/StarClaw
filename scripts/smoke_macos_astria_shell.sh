#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
SMOKE_HOME="$TMP_DIR/home"
BIN="$TMP_DIR/starclaw"
BASE_URL="${APP_SMOKE_BASE_URL:-http://127.0.0.1:7533}"

cleanup() {
  curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_macos_astria_shell: $*" >&2
  exit 1
}

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "smoke_macos_astria_shell: skipped (macOS required)"
  exit 0
fi

echo "==> building StarClaw for supervision smoke"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

app_path="$(ASTRIA_BUILD_DIR="$TMP_DIR/build/desktop/macos" ASTRIA_BUNDLED_STARCLAW_BIN="$BIN" "$ROOT_DIR/scripts/build_macos_astria_shell.sh")"

[[ -d "$app_path" ]] || fail "missing app bundle: $app_path"
[[ -x "$app_path/Contents/MacOS/Astria" ]] || fail "missing executable"
[[ -f "$app_path/Contents/Info.plist" ]] || fail "missing Info.plist"
[[ -x "$app_path/Contents/Resources/starclaw" ]] || fail "missing bundled daemon"

/usr/libexec/PlistBuddy -c "Print :CFBundleIdentifier" "$app_path/Contents/Info.plist" | grep -Fxq "dev.starclaw.astria" || fail "unexpected bundle identifier"
/usr/libexec/PlistBuddy -c "Print :CFBundleShortVersionString" "$app_path/Contents/Info.plist" >/dev/null || fail "missing bundle version"
/usr/libexec/PlistBuddy -c "Print :CFBundleVersion" "$app_path/Contents/Info.plist" >/dev/null || fail "missing bundle build"
/usr/libexec/PlistBuddy -c "Print :NSAppTransportSecurity:NSAllowsLocalNetworking" "$app_path/Contents/Info.plist" | grep -Fxq "true" || fail "local networking not enabled"

echo "==> checking Astria route recovery"
"$app_path/Contents/MacOS/Astria" --route-recovery-smoke

echo "==> checking Astria native commands"
"$app_path/Contents/MacOS/Astria" --native-command-smoke

echo "==> checking Astria diagnostics export"
"$app_path/Contents/MacOS/Astria" --diagnostics-export-smoke

echo "==> checking Astria permission helper"
"$app_path/Contents/MacOS/Astria" --permission-helper-smoke

echo "==> checking Astria Desktop RPC validation"
"$app_path/Contents/MacOS/Astria" --desktop-rpc-smoke

echo "==> checking Astria Desktop RPC fallback recovery"
"$app_path/Contents/MacOS/Astria" --desktop-rpc-fallback-smoke

echo "==> checking Astria Desktop RPC session lifecycle"
"$app_path/Contents/MacOS/Astria" --desktop-rpc-session-smoke

mkdir -p "$SMOKE_HOME/.starclaw"
cat > "$SMOKE_HOME/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
audit:
  enabled: false
YAML

curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true

echo "==> checking Astria bundled daemon supervision"
env HOME="$SMOKE_HOME" ASTRIA_RUNTIME_DIR="$TMP_DIR/runtime" "$app_path/Contents/MacOS/Astria" --supervision-smoke --startup-timeout 8
curl -fsS "$BASE_URL/health" >/dev/null || fail "daemon did not stay healthy after app supervision"
[[ -S "$TMP_DIR/runtime/daemon.sock" ]] || fail "Desktop RPC socket was not created"
[[ -f "$TMP_DIR/runtime/daemon.pid" ]] || fail "Desktop RPC pidfile was not created"

echo "==> checking Astria daemon attach"
env HOME="$SMOKE_HOME" ASTRIA_RUNTIME_DIR="$TMP_DIR/runtime" "$app_path/Contents/MacOS/Astria" --supervision-smoke --startup-timeout 8

echo "==> checking Astria daemon launch failure"
curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
if env HOME="$SMOKE_HOME" ASTRIA_RUNTIME_DIR="$TMP_DIR/missing-runtime" ASTRIA_STARCLAW_BIN="$TMP_DIR/missing-starclaw" "$app_path/Contents/MacOS/Astria" --supervision-smoke --startup-timeout 1 2>"$TMP_DIR/failed-launch.err"; then
  fail "supervision smoke should fail with a missing daemon binary"
fi
grep -Fq "failed to launch daemon" "$TMP_DIR/failed-launch.err" || {
  cat "$TMP_DIR/failed-launch.err" >&2
  fail "missing launch failure diagnostic"
}

echo "smoke_macos_astria_shell: ok"
