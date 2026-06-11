#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
SMOKE_HOME="$TMP_DIR/home"
NODE_DIR="$TMP_DIR/node"
NODE_SCRIPT="$NODE_DIR/desktop-layout-smoke.mjs"
DAEMON_LOG="$TMP_DIR/daemon.log"
DAEMON_PID=""

pick_free_port() {
  python3 - <<'PY'
import socket

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.bind(("127.0.0.1", 0))
    print(sock.getsockname()[1])
PY
}

url_port() {
  URL_TO_PARSE="$1" python3 - <<'PY'
import os
from urllib.parse import urlparse

parsed = urlparse(os.environ["URL_TO_PARSE"])
if parsed.port is None:
    raise SystemExit("missing port")
print(parsed.port)
PY
}

if [[ -n "${WEBUI_SMOKE_BASE_URL:-}" ]]; then
  BASE_URL="$WEBUI_SMOKE_BASE_URL"
  DAEMON_PORT="$(url_port "$BASE_URL")"
else
  DAEMON_PORT="$(pick_free_port)"
  BASE_URL="http://127.0.0.1:$DAEMON_PORT"
fi

cleanup() {
  if [[ -n "$DAEMON_PID" ]] && kill -0 "$DAEMON_PID" >/dev/null 2>&1; then
    curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
    for _ in {1..20}; do
      if ! kill -0 "$DAEMON_PID" >/dev/null 2>&1; then
        break
      fi
      sleep 0.1
    done
    kill "$DAEMON_PID" >/dev/null 2>&1 || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_webui_desktop_layout: $*" >&2
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

write_config() {
  mkdir -p "$SMOKE_HOME/.starclaw"
  cat > "$SMOKE_HOME/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
permissions:
  allowed_dirs:
    - "~"
    - "."
  allowed_commands:
    - "go test"
  denied_commands:
    - "shutdown"
audit:
  enabled: false
YAML
}

write_node_package() {
  mkdir -p "$NODE_DIR"
  cat > "$NODE_DIR/package.json" <<'JSON'
{"type":"module","dependencies":{"playwright":"^1.59.1"}}
JSON
}

write_node_script() {
  cat > "$NODE_SCRIPT" <<'JS'
import { chromium } from "playwright";

const baseURL = process.env.BASE_URL;

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

function assertWithMeasure(condition, message, measure) {
  if (!condition) throw new Error(`${message}: ${JSON.stringify(measure)}`);
}

async function switchPanel(page, panel) {
  await page.evaluate((name) => {
    if (typeof window.switchPanel === "function") window.switchPanel(name);
  }, panel);
  await page.locator(`#panel-${panel}.active`).waitFor();
}

async function inspectLayout(page, panel, viewport) {
  await switchPanel(page, panel);
  return page.evaluate((panelName) => {
    const activePanel = document.querySelector(`#panel-${panelName}.active`);
    const shell = document.querySelector(".shell");
    const workspace = document.querySelector(".workspace");
    const topbar = document.querySelector(".topbar");
    const title = document.querySelector("#view-title");
    const starMap = document.querySelector("[data-star-map]");
    const starDecayField = document.querySelector("[data-star-decay-field]");
    const panelRect = activePanel?.getBoundingClientRect();
    const shellRect = shell?.getBoundingClientRect();
    const workspaceRect = workspace?.getBoundingClientRect();
    const topbarRect = topbar?.getBoundingClientRect();
    const titleRect = title?.getBoundingClientRect();
    const homeHero = document.querySelector(".home-hero");
    const homeHeroRect = homeHero?.getBoundingClientRect();
    const starMapRect = starMap?.getBoundingClientRect();
    const starDecayFieldRect = starDecayField?.getBoundingClientRect();

    return {
      viewport: { width: innerWidth, height: innerHeight },
      documentScrollWidth: document.documentElement.scrollWidth,
      documentClientWidth: document.documentElement.clientWidth,
      bodyScrollWidth: document.body.scrollWidth,
      shell: shellRect && { x: shellRect.x, y: shellRect.y, width: shellRect.width, height: shellRect.height, right: shellRect.right, bottom: shellRect.bottom },
      workspace: workspaceRect && { width: workspaceRect.width, height: workspaceRect.height, right: workspaceRect.right, bottom: workspaceRect.bottom },
      topbar: topbarRect && { width: topbarRect.width, height: topbarRect.height, right: topbarRect.right, bottom: topbarRect.bottom },
      title: titleRect && { text: title?.textContent?.trim() || "", width: titleRect.width, height: titleRect.height },
      panel: panelRect && { width: panelRect.width, height: panelRect.height, right: panelRect.right, bottom: panelRect.bottom },
      homeHero: homeHeroRect && { width: homeHeroRect.width, height: homeHeroRect.height },
      starMap: starMapRect && { width: starMapRect.width, height: starMapRect.height, state: starMap?.dataset?.starState || "" },
      starDecayField: starDecayFieldRect && { width: starDecayFieldRect.width, height: starDecayFieldRect.height },
      activePanelText: activePanel?.textContent?.trim().slice(0, 120) || "",
    };
  }, panel);
}

function assertLayout(measure, panel, viewport) {
  assert(measure.documentScrollWidth <= measure.documentClientWidth, `${panel}@${viewport.width} has document horizontal overflow`);
  assert(measure.bodyScrollWidth <= measure.documentClientWidth, `${panel}@${viewport.width} has body horizontal overflow`);
  assert(measure.shell, `${panel}@${viewport.width} missing shell`);
  assert(measure.workspace, `${panel}@${viewport.width} missing workspace`);
  assert(measure.topbar, `${panel}@${viewport.width} missing topbar`);
  assert(measure.panel, `${panel}@${viewport.width} missing active panel`);
  assert(measure.title?.text, `${panel}@${viewport.width} missing view title`);
  assert(measure.shell.width >= Math.min(940, viewport.width - 48), `${panel}@${viewport.width} shell width too small: ${measure.shell.width}`);
  assert(measure.shell.height >= Math.min(620, viewport.height - 40), `${panel}@${viewport.width} shell height too small: ${measure.shell.height}`);
  assert(measure.shell.right <= measure.viewport.width + 1, `${panel}@${viewport.width} shell exceeds viewport`);
  assert(measure.panel.right <= measure.viewport.width + 1, `${panel}@${viewport.width} panel exceeds viewport`);
  assert(measure.topbar.right <= measure.viewport.width + 1, `${panel}@${viewport.width} topbar exceeds viewport`);
  assert(measure.activePanelText.length > 20, `${panel}@${viewport.width} panel looks empty`);
  if (panel === "home") {
    assertWithMeasure(measure.homeHero?.width > 240, `${panel}@${viewport.width} home briefing width too small`, measure);
    assertWithMeasure(measure.homeHero?.height > 220, `${panel}@${viewport.width} home briefing height too small`, measure);
    assertWithMeasure(measure.starMap?.width > 260, `${panel}@${viewport.width} star map width too small`, measure);
    assertWithMeasure(measure.starMap?.height > 220, `${panel}@${viewport.width} star map height too small`, measure);
    assert(measure.starMap?.state, `${panel}@${viewport.width} star map state missing`);
    assertWithMeasure(measure.starDecayField?.width > 260, `${panel}@${viewport.width} star decay field width too small`, measure);
    assertWithMeasure(measure.starDecayField?.height > 180, `${panel}@${viewport.width} star decay field height too small`, measure);
  }
}

const browser = await chromium.launch({ headless: true });
const viewports = [
  { width: 1400, height: 860 },
  { width: 1180, height: 760 },
  { width: 1100, height: 720 },
  { width: 1000, height: 680 },
];
const panels = [
  "home",
  "runs",
  "results",
  "memory",
  "settings",
  "manage",
  "diagnostics",
  "config",
  "permissions",
  "version",
];

try {
  for (const viewport of viewports) {
    const context = await browser.newContext({ viewport });
    const page = await context.newPage();
    await page.goto(`${baseURL}/app/`, { waitUntil: "domcontentloaded" });
    await page.getByRole("heading", { name: "Astria 任务台" }).waitFor();

    for (const panel of panels) {
      const measure = await inspectLayout(page, panel, viewport);
      assertLayout(measure, panel, viewport);
    }

    await context.close();
  }
} finally {
  await browser.close();
}
JS
}

require_cmd curl
require_cmd npx

echo "==> building StarClaw"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

write_config
write_node_package
write_node_script

echo "==> installing browser smoke dependency"
(cd "$NODE_DIR" && npm install --silent)
if [[ "${CI:-}" == "true" ]]; then
  (cd "$NODE_DIR" && npx playwright install chromium --with-deps >/dev/null)
else
  (cd "$NODE_DIR" && npx playwright install chromium >/dev/null)
fi

echo "==> starting daemon"
env HOME="$SMOKE_HOME" STARCLAW_DAEMON_PORT="$DAEMON_PORT" "$BIN" daemon start >"$DAEMON_LOG" 2>&1 &
DAEMON_PID="$!"
wait_for_health

echo "==> checking desktop layout at minimum app sizes"
env BASE_URL="$BASE_URL" node "$NODE_SCRIPT"

echo "smoke_webui_desktop_layout: ok"
