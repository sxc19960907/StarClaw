#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_cli: $*" >&2
  exit 1
}

run_capture() {
  local name="$1"
  shift
  local output
  local status

  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  printf '%s\n' "$output" > "$TMP_DIR/$name.out"
  printf '%s' "$status" > "$TMP_DIR/$name.status"
}

expect_status() {
  local name="$1"
  local want="$2"
  local got
  got="$(cat "$TMP_DIR/$name.status")"
  if [[ "$got" != "$want" ]]; then
    echo "---- $name output ----" >&2
    cat "$TMP_DIR/$name.out" >&2
    fail "$name exited with $got, want $want"
  fi
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

echo "==> building StarClaw"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

EMPTY_HOME="$TMP_DIR/home-empty"
CONFIG_HOME="$TMP_DIR/home-config"
mkdir -p "$EMPTY_HOME" "$CONFIG_HOME/.starclaw"

echo "==> checking version"
run_capture version env HOME="$EMPTY_HOME" "$BIN" version
expect_status version 0
expect_contains version "starclaw version"

echo "==> checking help"
run_capture help env HOME="$EMPTY_HOME" "$BIN" --help
expect_status help 0
expect_contains help "Available Commands:"
expect_contains help "app"
expect_contains help "chat"
expect_contains help "doctor"
expect_contains help "sessions"

echo "==> checking app command help"
run_capture app_help env HOME="$EMPTY_HOME" "$BIN" app --help
expect_status app_help 0
expect_contains app_help "Start the daemon if needed and open the Web UI"
expect_contains app_help "--check"
expect_contains app_help "--no-open"

echo "==> checking app launch readiness"
run_capture app_check env HOME="$EMPTY_HOME" "$BIN" app --check
expect_status app_check 0
expect_contains app_check "StarClaw app launch readiness"
expect_contains app_check "Launch:        starclaw app"
expect_contains app_check "Web UI:        http://127.0.0.1:7533/app/"
expect_contains app_check "Diagnostics:   http://127.0.0.1:7533/diagnostics"

echo "==> checking doctor"
run_capture doctor env HOME="$EMPTY_HOME" "$BIN" doctor
expect_status doctor 0
expect_contains doctor "StarClaw doctor"
expect_contains doctor "Launch:        starclaw app"
expect_contains doctor "Web UI:        http://127.0.0.1:7533/app/"
expect_contains doctor "Diagnostics:   http://127.0.0.1:7533/diagnostics"
expect_contains doctor "Local checks:"
expect_contains doctor "Daemon:        not running"

echo "==> checking daemon open start flag help"
run_capture daemon_open_help env HOME="$EMPTY_HOME" "$BIN" daemon open --help
expect_status daemon_open_help 0
expect_contains daemon_open_help "--start"

echo "==> checking missing-config chat failure"
run_capture chat_missing env HOME="$EMPTY_HOME" "$BIN" chat "hello"
expect_status chat_missing 1
expect_contains chat_missing "Configuration required. Run 'starclaw setup'"

if [[ ! -f "$EMPTY_HOME/.starclaw/config.yaml" ]]; then
  fail "expected default config to be created under isolated HOME"
fi

echo "==> checking sessions command"
run_capture sessions env HOME="$EMPTY_HOME" "$BIN" sessions
expect_status sessions 0
expect_contains sessions "No saved sessions found."

echo "==> checking mcp list command"
run_capture mcp_list env HOME="$EMPTY_HOME" "$BIN" mcp list
expect_status mcp_list 0
expect_contains mcp_list "No MCP servers configured."

echo "==> checking isolated existing config"
cat > "$CONFIG_HOME/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
audit:
  enabled: false
YAML
run_capture sessions_config env HOME="$CONFIG_HOME" "$BIN" sessions
expect_status sessions_config 0
expect_contains sessions_config "No saved sessions found."

echo "==> checking shell completion"
run_capture completion env HOME="$EMPTY_HOME" "$BIN" completion bash
expect_status completion 0
expect_contains completion "# bash completion"

echo "smoke_cli: ok"
