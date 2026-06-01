#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
SMOKE_HOME="$TMP_DIR/home"
DAEMON_LOG="$TMP_DIR/daemon.log"
NODE_DIR="$TMP_DIR/node"
NODE_SCRIPT="$NODE_DIR/webui-smoke.mjs"
BASE_URL="http://127.0.0.1:7533"
SCREENSHOT_DIR="$ROOT_DIR/output/playwright"
SCREENSHOT="$SCREENSHOT_DIR/daemon-webui-smoke.png"
DAEMON_PID=""

cleanup() {
  if [[ -n "$DAEMON_PID" ]] && kill -0 "$DAEMON_PID" >/dev/null 2>&1; then
    curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
    for _ in {1..20}; do
      if ! kill -0 "$DAEMON_PID" >/dev/null 2>&1; then
        break
      fi
      sleep 0.1
    done
    if kill -0 "$DAEMON_PID" >/dev/null 2>&1; then
      kill "$DAEMON_PID" >/dev/null 2>&1 || true
    fi
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_webui: $*" >&2
  if [[ -f "$DAEMON_LOG" ]]; then
    echo "---- daemon log ----" >&2
    cat "$DAEMON_LOG" >&2
  fi
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
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

require_cmd curl
require_cmd npx

echo "==> building StarClaw"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

mkdir -p "$SMOKE_HOME/.starclaw" "$SCREENSHOT_DIR"
cat > "$SMOKE_HOME/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
audit:
  enabled: false
YAML

mkdir -p "$NODE_DIR"
cat > "$NODE_DIR/package.json" <<'JSON'
{"type":"module","dependencies":{"playwright":"^1.59.1"}}
JSON

echo "==> installing browser smoke dependency"
(cd "$NODE_DIR" && npm install --silent)
(cd "$NODE_DIR" && npx playwright install chromium >/dev/null)

echo "==> starting daemon"
env HOME="$SMOKE_HOME" "$BIN" daemon start >"$DAEMON_LOG" 2>&1 &
DAEMON_PID="$!"
wait_for_health

echo "==> checking daemon routes"
curl -fsS "$BASE_URL/status" >/dev/null
curl -fsSI "$BASE_URL/" | grep -F "Location: /app/" >/dev/null || fail "root redirect missing"
curl -fsSI "$BASE_URL/app" | grep -F "Location: /app/" >/dev/null || fail "app redirect missing"
curl -fsS "$BASE_URL/app/" | grep -F "StarClaw" >/dev/null || fail "app HTML missing StarClaw"
curl -fsS "$BASE_URL/app/assets/app.js" | grep -F "connectEventStream" >/dev/null || fail "app JS missing event stream code"
curl -fsS "$BASE_URL/app/assets/styles.css" | grep -F "approval-card" >/dev/null || fail "CSS missing approval styles"

cat > "$NODE_SCRIPT" <<'JS'
import { chromium } from "playwright";
import fs from "node:fs";

const baseURL = process.env.BASE_URL;
const screenshot = process.env.SCREENSHOT;

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 960 } });

try {
  await page.goto(`${baseURL}/app/`, { waitUntil: "domcontentloaded" });
  await page.getByRole("heading", { name: "Chat" }).waitFor();
  await page.getByPlaceholder("Message StarClaw").waitFor();
  await page.getByRole("button", { name: "Send" }).waitFor();
  assert(await page.locator(".sidebar").count() === 1, "sidebar missing");

  await page.getByRole("button", { name: /Schedules/ }).click();
  await page.getByLabel("Cron expression").fill("* * * * *");
  await page.getByLabel("Schedule prompt").fill("webui smoke schedule");
  await page.getByRole("button", { name: "Create schedule" }).click();
  await page.getByText("webui smoke schedule").waitFor();
  await page.getByRole("button", { name: "Pause" }).click();
  await page.getByRole("button", { name: "Enable" }).waitFor();
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByText("No schedules configured.").waitFor();

  await page.getByRole("button", { name: /Chat/ }).click();
  const approvalRendered = await page.evaluate(async () => {
    if (typeof window.renderApprovalCard !== "function") return false;
    window.renderApprovalCard({
      request_id: "apr_smoke_missing",
      thread_id: "webui-smoke",
      channel: "http",
      tool: "bash",
      args: JSON.stringify({ command: "echo smoke" }),
      reason: "smoke approval"
    });
    return true;
  });
  assert(approvalRendered, "approval renderer unavailable");
  const approvalCard = page.locator(".approval-card");
  await approvalCard.getByText("Approval required").waitFor();
  await approvalCard.getByText("smoke approval").waitFor();
  await page.getByRole("button", { name: "Allow" }).click();
  await approvalCard.getByText("allowed").waitFor();

  const eventStatus = await page.evaluate(async (url) => {
    const response = await fetch(`${url}/events`);
    await response.body.cancel();
    return response.status;
  }, baseURL);
  assert(eventStatus === 200, `events status = ${eventStatus}`);

  await page.screenshot({ path: screenshot, fullPage: true });
  assert(fs.existsSync(screenshot), "screenshot was not written");
} finally {
  await browser.close();
}
JS

echo "==> running browser smoke"
env BASE_URL="$BASE_URL" SCREENSHOT="$SCREENSHOT" node "$NODE_SCRIPT"

echo "smoke_webui: ok"
echo "screenshot: $SCREENSHOT"
