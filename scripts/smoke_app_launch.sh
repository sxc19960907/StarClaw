#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
SMOKE_HOME="$TMP_DIR/home"
BASE_URL="${APP_SMOKE_BASE_URL:-http://127.0.0.1:7533}"

cleanup() {
  curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_app_launch: $*" >&2
  exit 1
}

expect_contains() {
  local name="$1"
  local needle="$2"
  if ! grep -Fq -- "$needle" "$TMP_DIR/$name.out"; then
    echo "---- $name output ----" >&2
    cat "$TMP_DIR/$name.out" >&2
    fail "$name output did not contain: $needle"
  fi
}

expect_json_contains() {
  local url="$1"
  local needle="$2"
  local output="$TMP_DIR/route.out"
  curl -fsS "$url" > "$output" || fail "route failed: $url"
  if ! grep -Fq -- "$needle" "$output"; then
    echo "---- $url output ----" >&2
    cat "$output" >&2
    fail "$url did not contain: $needle"
  fi
}

wait_for_health() {
  for _ in {1..80}; do
    if curl -fsS "$BASE_URL/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done
  fail "daemon did not become healthy"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

require_cmd curl

echo "==> building StarClaw"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

mkdir -p "$SMOKE_HOME/.starclaw"
cat > "$SMOKE_HOME/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
audit:
  enabled: false
YAML

echo "==> checking launch readiness"
env HOME="$SMOKE_HOME" "$BIN" app --check > "$TMP_DIR/app-check.out"
expect_contains app-check "StarClaw app launch readiness"
expect_contains app-check "Daemon:        not running"
expect_contains app-check "Launch:        starclaw app"
expect_contains app-check "Web UI:        $BASE_URL/app/"
expect_contains app-check "Diagnostics:   $BASE_URL/diagnostics"
expect_contains app-check "Data:          $SMOKE_HOME/.starclaw"

echo "==> starting app without browser"
env HOME="$SMOKE_HOME" "$BIN" app --no-open > "$TMP_DIR/app-no-open.out"
expect_contains app-no-open "Web UI: $BASE_URL/app/"
wait_for_health

echo "==> checking daemon reuse"
env HOME="$SMOKE_HOME" "$BIN" app --no-open > "$TMP_DIR/app-reuse.out"
expect_contains app-reuse "Daemon already running. Web UI: $BASE_URL/app/"

echo "==> checking daemon routes"
expect_json_contains "$BASE_URL/version" '"launch_command":"starclaw app"'
expect_json_contains "$BASE_URL/version" '"web_url":"'"$BASE_URL"'/app/"'
expect_json_contains "$BASE_URL/version" '"health_url":"'"$BASE_URL"'/health"'
expect_json_contains "$BASE_URL/version" '"status_url":"'"$BASE_URL"'/status"'
expect_json_contains "$BASE_URL/version" '"diagnostics_url":"'"$BASE_URL"'/diagnostics"'
expect_json_contains "$BASE_URL/version" '"starclaw_dir":"'"$SMOKE_HOME"'/.starclaw"'
expect_json_contains "$BASE_URL/version" '"config_path":"'"$SMOKE_HOME"'/.starclaw/config.yaml"'
expect_json_contains "$BASE_URL/diagnostics" '"launch_command":"starclaw app"'
expect_json_contains "$BASE_URL/diagnostics" '"web_url":"'"$BASE_URL"'/app/"'
expect_json_contains "$BASE_URL/diagnostics" '"config_path":"'"$SMOKE_HOME"'/.starclaw/config.yaml"'

echo "==> checking doctor JSON"
env HOME="$SMOKE_HOME" "$BIN" doctor --json > "$TMP_DIR/doctor-json.out"
expect_contains doctor-json '"launch_command":"starclaw app"'
expect_contains doctor-json '"web_url":"'"$BASE_URL"'/app/"'
expect_contains doctor-json '"diagnostics_url":"'"$BASE_URL"'/diagnostics"'
expect_contains doctor-json '"starclaw_dir":"'"$SMOKE_HOME"'/.starclaw"'
expect_contains doctor-json '"config_path":"'"$SMOKE_HOME"'/.starclaw/config.yaml"'
expect_contains doctor-json '"daemon":{"running":true'
expect_contains doctor-json '"diagnostics":{"status":'

echo "==> stopping daemon"
env HOME="$SMOKE_HOME" "$BIN" daemon stop > "$TMP_DIR/daemon-stop.out"
expect_contains daemon-stop "Daemon: shutting_down"

echo "smoke_app_launch: ok"
