#!/usr/bin/env bash
set -euo pipefail

SMOKE_MODE="${WEBUI_SMOKE_MODE:-full}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
BIN="$TMP_DIR/starclaw"
SMOKE_HOME="$TMP_DIR/home"
DAEMON_LOG="$TMP_DIR/daemon.log"
NODE_DIR="$TMP_DIR/node"
NODE_SCRIPT="$NODE_DIR/webui-smoke.mjs"
FAKE_PROVIDER_SCRIPT="$NODE_DIR/fake-openai-provider.mjs"
ARTIFACT_DIR="${WEBUI_SMOKE_ARTIFACT_DIR:-$ROOT_DIR/output/playwright}"
SCREENSHOT_DIR="$ARTIFACT_DIR"
SCREENSHOT="$SCREENSHOT_DIR/daemon-webui-${SMOKE_MODE}-smoke.png"
HOME_SCREENSHOT="$SCREENSHOT_DIR/astria-home-${SMOKE_MODE}-smoke.png"
INTAKE_DOC="$SCREENSHOT_DIR/intake-smoke.docx"
DAEMON_LOG_ARTIFACT="$ARTIFACT_DIR/daemon-webui-${SMOKE_MODE}-smoke.log"
FAKE_PROVIDER_LOG_ARTIFACT="$ARTIFACT_DIR/fake-provider-${SMOKE_MODE}-smoke.log"
METADATA_ARTIFACT="$ARTIFACT_DIR/daemon-webui-${SMOKE_MODE}-smoke.metadata"
DAEMON_PID=""
FAKE_PROVIDER_PID=""

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

if [[ -n "${WEBUI_FAKE_PROVIDER_URL:-}" ]]; then
  FAKE_PROVIDER_URL="$WEBUI_FAKE_PROVIDER_URL"
  FAKE_PROVIDER_PORT="$(url_port "$FAKE_PROVIDER_URL")"
else
  FAKE_PROVIDER_PORT="$(pick_free_port)"
  FAKE_PROVIDER_URL="http://127.0.0.1:$FAKE_PROVIDER_PORT"
fi

persist_artifacts() {
  mkdir -p "$ARTIFACT_DIR"
  if [[ -f "$DAEMON_LOG" ]]; then
    cp "$DAEMON_LOG" "$DAEMON_LOG_ARTIFACT"
  fi
  if [[ -f "$TMP_DIR/fake-provider.log" ]]; then
    cp "$TMP_DIR/fake-provider.log" "$FAKE_PROVIDER_LOG_ARTIFACT"
  fi
  {
    printf 'mode=%s\n' "$SMOKE_MODE"
    printf 'base_url=%s\n' "$BASE_URL"
    printf 'daemon_port=%s\n' "$DAEMON_PORT"
    printf 'fake_provider_url=%s\n' "$FAKE_PROVIDER_URL"
    printf 'fake_provider_port=%s\n' "$FAKE_PROVIDER_PORT"
    printf 'screenshot=%s\n' "$SCREENSHOT"
    printf 'home_screenshot=%s\n' "$HOME_SCREENSHOT"
    printf 'daemon_log=%s\n' "$DAEMON_LOG_ARTIFACT"
    printf 'fake_provider_log=%s\n' "$FAKE_PROVIDER_LOG_ARTIFACT"
    printf 'created_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  } > "$METADATA_ARTIFACT"
}

cleanup() {
  persist_artifacts
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
  if [[ -n "$FAKE_PROVIDER_PID" ]] && kill -0 "$FAKE_PROVIDER_PID" >/dev/null 2>&1; then
    kill "$FAKE_PROVIDER_PID" >/dev/null 2>&1 || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_webui_${SMOKE_MODE}: $*" >&2
  persist_artifacts
  if [[ -f "$DAEMON_LOG" ]]; then
    echo "---- daemon log ----" >&2
    cat "$DAEMON_LOG" >&2
  fi
  if [[ -f "$TMP_DIR/fake-provider.log" ]]; then
    echo "---- fake provider log ----" >&2
    cat "$TMP_DIR/fake-provider.log" >&2
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

write_smoke_config() {
  mkdir -p "$SMOKE_HOME/.starclaw" "$SCREENSHOT_DIR"
  if [[ "$SMOKE_MODE" == "config" ]]; then
    cat > "$SMOKE_HOME/.starclaw/config.yaml" <<'YAML'
provider: anthropic
permissions:
  allowed_dirs:
    - "~"
    - "."
  denied_commands:
    - "shutdown"
audit:
  enabled: false
YAML
    return
  fi
  if [[ "$SMOKE_MODE" == "streaming" || "$SMOKE_MODE" == "tool_call" ]]; then
    local max_iterations="1"
    if [[ "$SMOKE_MODE" == "tool_call" ]]; then
      max_iterations="2"
    fi
    cat > "$SMOKE_HOME/.starclaw/config.yaml" <<YAML
provider: openai
openai_endpoint: "$FAKE_PROVIDER_URL"
openai_model: "fake-streaming-model"
openai_api_key: "fake-key"
api_key: dummy
agent:
  max_iterations: $max_iterations
  thinking: false
permissions:
  allowed_dirs:
    - "~"
    - "."
  denied_commands:
    - "shutdown"
audit:
  enabled: false
YAML
    return
  fi
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
  network_allowlist:
    - "api.github.com"
  sensitive_patterns:
    - "*.secret"
audit:
  enabled: false
YAML
}

write_fake_provider() {
  cat > "$FAKE_PROVIDER_SCRIPT" <<'JS'
import http from "node:http";

const port = Number(process.env.FAKE_PROVIDER_PORT || "17534");
let scenario = process.env.FAKE_PROVIDER_SCENARIO || "streaming";

function writeJSON(res, status, data) {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify(data));
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    let body = "";
    req.setEncoding("utf8");
    req.on("data", (chunk) => { body += chunk; });
    req.on("end", () => resolve(body));
    req.on("error", reject);
  });
}

function writeSSE(res, data) {
  res.write(`data: ${JSON.stringify(data)}\n\n`);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

const server = http.createServer(async (req, res) => {
  if (req.method === "GET" && req.url === "/health") {
    writeJSON(res, 200, { status: "ok" });
    return;
  }
  if (req.method === "POST" && req.url === "/scenario") {
    const rawBody = await readBody(req);
    const body = JSON.parse(rawBody || "{}");
    scenario = String(body.scenario || "streaming");
    writeJSON(res, 200, { scenario });
    return;
  }
  if (req.method !== "POST" || !req.url.endsWith("/chat/completions")) {
    writeJSON(res, 404, { error: { message: `unexpected route: ${req.method} ${req.url}` } });
    return;
  }

  const rawBody = await readBody(req);
  const request = JSON.parse(rawBody || "{}");
  if (scenario === "config_auth") {
    writeJSON(res, 401, { error: { message: "bad key Bearer fake-key" } });
    return;
  }
  if (scenario === "config_model") {
    writeJSON(res, 404, { error: { message: "model not found" } });
    return;
  }
  if (scenario === "config_rate") {
    writeJSON(res, 429, { error: { message: "too many requests" } });
    return;
  }
  if (scenario === "config_invalid") {
    writeJSON(res, 200, { choices: [] });
    return;
  }
  const messages = Array.isArray(request.messages) ? request.messages : [];
  const hasToolResult = messages.some((message) => String(message.content || "").includes('"type":"tool_result"'));
  if (scenario === "tool_call" && !hasToolResult) {
    const toolCall = {
      index: 0,
      id: "call_version_smoke",
      type: "function",
      function: { name: "version", arguments: "{}" },
    };
    if (!request.stream) {
      writeJSON(res, 200, {
        id: "fake-tool-call-sync",
        object: "chat.completion",
        choices: [{
          message: { role: "assistant", content: "", tool_calls: [{ id: toolCall.id, type: toolCall.type, function: toolCall.function }] },
          finish_reason: "tool_calls",
        }],
        usage: { prompt_tokens: 13, completion_tokens: 2 },
      });
      return;
    }
    res.writeHead(200, {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    });
    writeSSE(res, { choices: [{ delta: { tool_calls: [toolCall] }, finish_reason: null }] });
    writeSSE(res, { choices: [{ delta: {}, finish_reason: "tool_calls" }], usage: { prompt_tokens: 13, completion_tokens: 2 } });
    res.write("data: [DONE]\n\n");
    res.end();
    return;
  }
  const content = "Fake provider streamed response for GUI smoke.";
  const toolContent = "Version tool call completed for GUI smoke.";
  const responseContent = scenario === "tool_call" ? toolContent : content;
  if (!request.stream) {
    writeJSON(res, 200, {
      id: "fake-chat-sync",
      object: "chat.completion",
      choices: [{ message: { role: "assistant", content: responseContent }, finish_reason: "stop" }],
      usage: { prompt_tokens: 11, completion_tokens: 7 },
    });
    return;
  }

  res.writeHead(200, {
    "Content-Type": "text/event-stream",
    "Cache-Control": "no-cache",
    "Connection": "keep-alive",
  });
  const chunks = scenario === "tool_call"
    ? ["Version tool call ", "completed for GUI smoke."]
    : ["Fake provider ", "streamed response ", "for GUI smoke."];
  for (const chunk of chunks) {
    writeSSE(res, { choices: [{ delta: { content: chunk }, finish_reason: null }] });
    await sleep(80);
  }
  writeSSE(res, { choices: [{ delta: {}, finish_reason: "stop" }], usage: { prompt_tokens: 11, completion_tokens: 7 } });
  res.write("data: [DONE]\n\n");
  res.end();
});

server.listen(port, "127.0.0.1", () => {
  console.log(`fake-openai-provider listening on http://127.0.0.1:${port}`);
});

process.on("SIGTERM", () => server.close(() => process.exit(0)));
process.on("SIGINT", () => server.close(() => process.exit(0)));
JS
}

start_fake_provider_if_needed() {
  if [[ "$SMOKE_MODE" != "streaming" && "$SMOKE_MODE" != "tool_call" && "$SMOKE_MODE" != "config" ]]; then
    return
  fi
  write_fake_provider
  echo "==> starting fake OpenAI provider"
  local provider_scenario="$SMOKE_MODE"
  if [[ "$SMOKE_MODE" == "config" ]]; then
    provider_scenario="streaming"
  fi
  FAKE_PROVIDER_PORT="$FAKE_PROVIDER_PORT" FAKE_PROVIDER_SCENARIO="$provider_scenario" node "$FAKE_PROVIDER_SCRIPT" >"$TMP_DIR/fake-provider.log" 2>&1 &
  FAKE_PROVIDER_PID="$!"
  for _ in {1..80}; do
    if curl -fsS "$FAKE_PROVIDER_URL/health" >/dev/null 2>&1; then
      return
    fi
    sleep 0.1
  done
  fail "fake provider did not become healthy"
}

write_node_package() {
  mkdir -p "$NODE_DIR"
  cat > "$NODE_DIR/package.json" <<'JSON'
{"type":"module","dependencies":{"playwright":"^1.59.1"}}
JSON
}

write_intake_fixture() {
  mkdir -p "$SCREENSHOT_DIR"
  INTAKE_DOC="$INTAKE_DOC" python3 - <<'PY'
import os
import zipfile

path = os.environ["INTAKE_DOC"]
os.makedirs(os.path.dirname(path), exist_ok=True)
with zipfile.ZipFile(path, "w") as zf:
    zf.writestr(
        "word/document.xml",
        '<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Astria smoke intake document</w:t></w:r></w:p></w:body></w:document>',
    )
PY
}

write_browser_smoke() {
  cat > "$NODE_SCRIPT" <<'JS'
import { chromium } from "playwright";
import fs from "node:fs";

const baseURL = process.env.BASE_URL;
const screenshot = process.env.SCREENSHOT;
const homeScreenshot = process.env.HOME_SCREENSHOT;
const intakeDoc = process.env.INTAKE_DOC;
const mode = process.env.WEBUI_SMOKE_MODE || "full";

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

function agentConfig(agent) {
  const config = agent.Config || agent.config || {};
  const tools = config.Tools || config.tools || {};
  const heartbeat = config.Heartbeat || config.heartbeat || {};
  return {
    allow: tools.Allow || tools.allow || [],
    deny: tools.Deny || tools.deny || [],
    autoApprove: config.AutoApprove ?? config.auto_approve,
    heartbeatEvery: heartbeat.Every || heartbeat.every || "",
    heartbeatActiveHours: heartbeat.ActiveHours || heartbeat.active_hours || "",
    heartbeatModel: heartbeat.Model || heartbeat.model || "",
  };
}

function agentCommands(agent) {
  return agent.Commands || agent.commands || {};
}

async function fulfillIfUnhandled(route, options) {
  try {
    await route.fulfill(options);
  } catch (error) {
    if (!String(error?.message || error).includes("already handled")) throw error;
  }
}

async function openHomeDisclosure(page, selector) {
  await page.evaluate((targetSelector) => {
    const workbench = document.querySelector(".home-secondary-workbench");
    if (workbench) workbench.style.display = "grid";
    const disclosure = document.querySelector(targetSelector);
    if (disclosure) {
      disclosure.hidden = false;
      disclosure.style.display = "block";
      disclosure.open = true;
    }
  }, selector);
  const disclosure = page.locator(selector);
  await disclosure.waitFor({ state: "attached" });
  if (!(await disclosure.evaluate((node) => node.open))) {
    await disclosure.evaluate((node) => node.scrollIntoView({ block: "center", inline: "nearest" }));
    await disclosure.locator("summary").click();
  }
  await disclosure.evaluate((node) => node.scrollIntoView({ block: "start", inline: "nearest" }));
}

async function closeHomeDisclosures(page) {
  await page.evaluate(() => {
    document.querySelectorAll(".home-disclosure[open]").forEach((node) => {
      node.open = false;
    });
  });
}

async function openHome(page, options = {}) {
  await openPanel(page, "home");
  await page.locator("#panel-home.active").waitFor();
  if (options.calm) {
    await closeHomeDisclosures(page);
    await page.locator("#panel-home").evaluate((node) => {
      node.scrollTo({ top: 0, left: 0, behavior: "instant" });
    });
  }
}

function homeTaskInput(page) {
  return page.locator("#home-task-input");
}

async function openCommandRoute(page, query, resultName, commandID = "") {
  await page.locator("#command-center-button").click();
  await page.locator("#command-center-input").fill(query);
  if (commandID) {
    await page.locator(`#command-center-list [data-command-id="${commandID}"]`).click();
    return;
  }
  await page.locator("#command-center-list").getByRole("button", { name: resultName }).click();
}

async function openPanel(page, panel) {
  await page.evaluate((panelName) => {
    if (typeof window.switchPanel === "function") window.switchPanel(panelName);
  }, panel);
}

async function openSession(page, sessionID) {
  await page.evaluate((id) => {
    if (typeof window.selectSession === "function") window.selectSession(id);
  }, sessionID);
}

async function boot(page) {
  await page.goto(`${baseURL}/app/`, { waitUntil: "domcontentloaded" });
  await page.getByRole("heading", { name: "Astria 任务台" }).waitFor();
  await homeTaskInput(page).waitFor();
  await page.locator("#home-brief-current").waitFor();
  const briefVisible = await page.locator(".home-hero").evaluate((node) => {
    const rect = node.getBoundingClientRect();
    const style = getComputedStyle(node);
    return style.display !== "none" && rect.width > 260 && rect.height > 260;
  });
  assert(briefVisible, "home briefing panel should be visible and substantial");
  assert(await page.locator(".home-brief-metrics button").count() === 4, "home briefing should show four metrics");
  assert(await page.locator(".home-brief-queue button").count() === 3, "home briefing should show three queued actions");
  assert(await page.locator(".home-brief-queue").evaluate((node) => node.textContent.includes("工作队列")), "home briefing should include work queue");
  await page.locator('.home-brief-queue [data-recipe="code-review"]').evaluate((button) => button.click());
  assert((await homeTaskInput(page).inputValue()).includes("Review the current working tree"), "home work queue should launch code review recipe");
  await openCommandRoute(page, "memory", /记忆星图|Memory Map/, "panel:memory");
  await page.locator("#panel-memory.active").waitFor();
  await page.keyboard.press(process.platform === "darwin" ? "Meta+K" : "Control+K");
  await page.locator("#command-center-input").fill("review");
  await page.locator("#command-center-list").getByRole("button", { name: /代码评审/ }).click();
  await page.locator("#panel-home.active").waitFor();
  await openHomeDisclosure(page, "#home-planning-disclosure");
  await page.locator('#strategy-matrix [data-strategy="research"]').evaluate((element) => element.click());
  const strategyPrompt = await homeTaskInput(page).inputValue();
  assert(strategyPrompt.includes("research brief") || strategyPrompt.includes("调研简报"), "research strategy should prefill home prompt");
  assert(await page.locator("#strategy-brief").evaluate((node) => node.textContent.includes("调研简报")), "strategy brief should show research strategy");
  assert(await page.locator("#focus-brief").evaluate((node) => node.textContent.includes("策略")), "focus brief should show strategy row");
  await page.locator('#strategy-brief [data-panel="runs"]').evaluate((element) => element.click());
  await page.locator("#panel-runs.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-planning-disclosure");
  await page.locator("#prompt-suggestion-dock").evaluate((dock) => {
    const button = [...dock.querySelectorAll("button")].find((item) => item.dataset.homePrompt?.includes("research brief") || item.textContent.includes("调研简报"));
    if (!button) throw new Error("research prompt suggestion not found");
    button.click();
  });
  const suggestionPrompt = await homeTaskInput(page).inputValue();
  assert(suggestionPrompt.includes("research brief") || suggestionPrompt.includes("调研简报"), "prompt suggestion should seed home prompt");
  await page.locator('#workflow-recipes [data-recipe="code-review"]').evaluate((element) => element.click());
  const reviewPrompt = await homeTaskInput(page).inputValue();
  assert(reviewPrompt.includes("Review the current working tree"), "code review recipe should prefill home prompt");
  assert(await page.locator("#workflow-stage-rail").evaluate((node) => node.textContent.includes("代码评审")), "stage rail should show selected workflow");
  assert(await page.locator("#workflow-stage-rail").evaluate((node) => node.textContent.includes("Daemon 执行")), "stage rail should show run stage");
  assert(await page.locator("#focus-brief").evaluate((node) => node.textContent.includes("代码评审")), "focus brief should show code review workflow");
  await openHomeDisclosure(page, "#home-review-disclosure");
  assert(await page.locator("#workspace-health-strip").evaluate((node) => node.textContent.includes("诊断")), "workspace health should include diagnostics");
  assert(await page.locator("#workspace-health-strip").evaluate((node) => node.textContent.includes("权限")), "workspace health should include permissions");
  assert(await page.locator("#workspace-health-strip").evaluate((node) => node.textContent.includes("MCP")), "workspace health should include MCP");
  await page.locator('#workspace-health-strip [data-panel="memory"]').evaluate((element) => element.click());
  await page.locator("#panel-memory.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-review-disclosure");
  await openPanel(page, "permissions");
  await page.locator("#panel-permissions.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-review-disclosure");
  await openPanel(page, "permissions");
  await page.locator("#panel-permissions.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-review-disclosure");
  await page.locator("#knowledge-curation-grid").evaluate((grid) => {
    const button = grid.querySelector('[data-panel="memory"]') || grid.querySelector("button");
    if (!button) throw new Error("knowledge curation action not found");
    button.click();
  });
  await page.locator("#panel-memory.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-review-disclosure");
  await page.locator("#tool-dock-inspector-grid button").first().evaluate((element) => element.click());
  await page.locator("#panel-mcp.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-planning-disclosure");
  assert(await page.locator("#workflow-brief").evaluate((node) => node.textContent.includes("一份按严重程度排序的评审报告")), "workflow brief should show code review outcome");
  assert(await page.locator("#workflow-brief").evaluate((node) => node.textContent.includes("当前 git diff")), "workflow brief should show code review context");
  await page.locator('#workflow-recipes [data-recipe="file-intake"]').evaluate((element) => element.click());
  assert(await page.locator("#workflow-brief").evaluate((node) => node.textContent.includes("把本地文件内容整理成可引用上下文")), "workflow brief should show file intake outcome");
  assert(await page.locator("#workflow-brief").evaluate((node) => node.textContent.includes("打开文件星舱")), "workflow brief should show file intake route");
  assert(await page.locator("#home-mode-route").evaluate((node) => node.textContent.includes("打开文件星舱")), "home route should show file intake");
  await page.locator('#workflow-brief [data-panel="intake"]').evaluate((element) => element.click());
  await page.locator("#panel-intake.active").waitFor();
  await openHome(page);
  await openHomeDisclosure(page, "#home-context-disclosure");
  assert(await page.locator("#workspace-session-hub").evaluate((node) => node.textContent.includes("会话")), "session hub should include sessions");
  assert(await page.locator("#workspace-session-hub").evaluate((node) => node.textContent.includes("运行")), "session hub should include runs");
  assert(await page.locator("#workspace-session-hub").evaluate((node) => node.textContent.includes("记忆")), "session hub should include memory");
  await page.locator('#workspace-session-hub [data-panel="intake"]').evaluate((element) => element.click());
  await page.locator("#panel-intake.active").waitFor();
  await openHome(page, { calm: true });
  if (homeScreenshot) {
    await page.screenshot({ path: homeScreenshot, fullPage: true });
    assert(fs.existsSync(homeScreenshot), "home screenshot was not written");
  }
  await openPanel(page, "chat");
  await page.locator("#panel-chat").waitFor();
  await page.getByPlaceholder("向 Astria 描述任务").waitFor();
  await page.getByRole("button", { name: "发送" }).waitFor();
  assert(await page.locator(".sidebar").count() === 1, "sidebar missing");
  await page.locator("#diagnostics-chip").waitFor();
}

async function openManagePanel(page, name) {
  const manageAliases = {
    Agents: "agents",
    Agent: "agents",
    "Agent Council": "council",
    "Browser Planner": "browser",
    "Budget Guard": "budget",
    "Citation Planner": "citation",
    "Comparison Workbench": "compare",
    "Data Planner": "data",
    "File Intake": "intake",
    "Inbox": "inbox",
    "Knowledge Reconciliation": "reconcile",
    "MCP Starport": "mcp",
    "Memory Map": "memory",
    "Playbook Library": "playbooks",
    "Proactive Delivery": "delivery",
    "Prompt Lab": "promptlab",
    "Run Quality": "quality",
    "Schedules": "schedules",
    "Share Pack": "share",
    "Source Registry": "sources",
    "Starter Kits": "starter",
    "Workspace Snapshot": "snapshot",
    "产物星库": "results",
    "复用星库": "reuse",
    "主动交付": "delivery",
    "交接包": "share",
    "启动套件": "starter",
    "实践手册": "playbooks",
    "工作区快照": "snapshot",
    "数据规划器": "data",
    "浏览器规划器": "browser",
    "路径比较台": "compare",
    "运行质量": "quality",
    "预算守卫": "budget",
  };
  const panel = manageAliases[name];
  if (panel) {
    await page.evaluate((targetPanel) => {
      if (typeof window.switchPanel === "function") window.switchPanel(targetPanel);
    }, panel);
    await page.locator(`#panel-${panel}.active`).waitFor();
    return;
  }
  const manageButton = page.getByRole("button", { name: /Manage|工作枢纽|更多功能/ });
  if (await manageButton.isVisible().catch(() => false)) {
    await manageButton.click();
  } else {
    await page.evaluate(() => {
      if (typeof window.switchPanel === "function") window.switchPanel("manage");
    });
  }
  await page.locator("#panel-manage").getByRole("heading", { name: /Manage|工作枢纽|更多功能/ }).waitFor();
  await page.locator("#panel-manage").getByRole("button", { name: new RegExp(`^${name}`) }).click();
}

async function openSettingsPanel(page, name) {
  await openPanel(page, "settings");
  await page.locator("#panel-settings").getByRole("heading", { name: /Settings|系统/ }).waitFor();
  const settingsAliases = {
    Diagnostics: "diagnostics",
    "诊断": "diagnostics",
    Config: "config",
    "连接器": "config",
    Permissions: "permissions",
    "权限": "permissions",
    Version: "version",
    "版本": "version",
  };
  const panel = settingsAliases[name];
  if (panel) {
    await page.evaluate((targetPanel) => {
      if (typeof window.switchPanel === "function") window.switchPanel(targetPanel);
    }, panel);
    await page.locator(`#panel-${panel}.active`).waitFor();
    return;
  }
  await page.locator("#panel-settings").getByRole("button", { name: new RegExp(`^${name}`) }).click();
}

async function runCore(page) {
  await page.locator("#diagnostics-chip").click();
  await page.locator("#panel-diagnostics").getByRole("heading", { name: /Diagnostics|诊断/ }).waitFor();
  await page.locator("#panel-diagnostics").getByText(/Launch readiness|启动就绪/).waitFor();
  await page.locator("#panel-diagnostics").getByText("starclaw app").waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/app/`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/health`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/status`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/diagnostics`).waitFor();
  assert(await page.locator("#panel-diagnostics").getByText(/Config|配置/).count() > 0, "diagnostics should include config path row");
  assert(await page.locator("#panel-diagnostics").getByText(/Agents|Agent 目录/).count() > 0, "diagnostics should include agents path row");
  assert(await page.locator("#panel-diagnostics").getByText(/Sessions|会话目录/).count() > 0, "diagnostics should include sessions path row");
  await page.locator("#panel-diagnostics").getByText("Provider", { exact: true }).waitFor();
  await page.getByRole("button", { name: /Fix provider setup|修复 provider 设置/ }).click();
  await page.locator("#panel-config").getByRole("heading", { name: /Config|连接器/ }).waitFor();
  await page.getByLabel("Provider").selectOption("ollama");
  await page.getByLabel("Ollama endpoint").fill("http://127.0.0.1:1");
  await page.getByLabel("Ollama model").fill("smoke-gui-model");
  await page.getByRole("button", { name: /Save provider config|保存 provider 设置/ }).click();
  await page.getByText(/Provider config saved\.|provider 设置已保存。/).waitFor();
  await openSettingsPanel(page, "Version");
  await page.locator("#panel-version").getByRole("heading", { name: /Version|版本/ }).waitFor();
  await page.locator("#version-list").getByText(/Release readiness|发布就绪/).waitFor();
  await page.locator("#version-list").getByText(/Development build|开发构建/).waitFor();
  await page.locator("#version-list").getByText(/Use a semver release build to enable update checks\.|使用 semver 发布构建后可启用更新检查。/).waitFor();
  await page.locator("#version-list").getByText(/Version|版本/).waitFor();
  await page.locator("#version-list").getByText(/Platform|平台/).waitFor();
  await page.locator("#version-list").getByText("Web UI", { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText(/Launch|启动命令/, { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText("starclaw app", { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText("starclaw update --check").waitFor();
  await page.locator("#version-list").getByText(/Runtime context|运行时上下文/).waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/health`).waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/status`).waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/diagnostics`).waitFor();
  await page.locator("#update-overview strong").getByText(/Development build|开发构建/).waitFor();
  assert(await page.getByRole("button", { name: /Check updates|检查更新/ }).isDisabled(), "development build should disable update check button");
  await page.locator("#update-check-state").getByText(/Unavailable|不可用/).waitFor();
  await page.getByRole("button", { name: /Copy support info|复制支持信息/ }).click();
  await page.getByText(/Support info copied\.|支持信息已复制。/).waitFor();
  const supportInfo = await page.evaluate(() => navigator.clipboard.readText());
  assert(supportInfo.includes("Astria support info") || supportInfo.includes("Astria 支持信息"), "support info missing heading");
  assert(supportInfo.includes("Version: dev"), "support info missing version");
  assert(supportInfo.includes(`Web UI: ${baseURL}/app/`), "support info missing web URL");
  assert(supportInfo.includes(`Diagnostics URL: ${baseURL}/diagnostics`), "support info missing diagnostics URL");
  assert(supportInfo.includes("Data dir:") || supportInfo.includes("数据目录:"), "support info missing data dir");
  assert(supportInfo.includes("Diagnostics status:"), "support info missing diagnostics status");
  assert(!supportInfo.toLowerCase().includes("api_key"), "support info should not include API key fields");
  await openHome(page);
  await openManagePanel(page, "MCP Starport");
  await page.locator("#panel-mcp.active").waitFor();
  await page.locator("#mcp-new-button").click();
  await page.getByLabel("MCP server name").fill("smoke");
  await page.getByLabel("MCP command").fill("node");
  await page.getByLabel("MCP args").fill("smoke-mcp.js");
  await page.getByLabel("MCP env").fill("SMOKE_TOKEN=fake-secret");
  await page.locator("#mcp-form").getByRole("button", { name: /Save dock|保存 dock/ }).click();
  await page.getByText(/MCP dock saved\.|MCP dock 已保存。/).waitFor();
  await page.locator("#mcp-list").getByText("smoke", { exact: true }).waitFor();
  await page.locator("#mcp-list").getByRole("button", { name: /Edit|编辑/ }).click();
  await page.getByLabel("MCP args").fill("smoke-mcp.js, --edited");
  await page.locator("#mcp-form").getByRole("button", { name: /Save dock|保存 dock/ }).click();
  await page.getByText(/MCP dock saved\.|MCP dock 已保存。/).waitFor();
  await page.locator("#mcp-list").getByRole("button", { name: /Disable|停用/ }).click();
  await page.getByText(/MCP dock disabled\.|MCP dock 已停用。/).waitFor();
  const mcpConfig = await page.evaluate(async () => {
    const response = await fetch("/config");
    return response.json();
  });
  const smokeDock = mcpConfig.config.mcp_servers.find((server) => server.name === "smoke");
  assert(smokeDock, "smoke MCP dock should be saved");
  assert(smokeDock.disabled === true, "smoke MCP dock should be disabled");
  assert(smokeDock.args.includes("--edited"), `smoke MCP args should include edit, got ${JSON.stringify(smokeDock.args)}`);
  assert(smokeDock.env_keys.includes("SMOKE_TOKEN"), `smoke MCP env key should be redacted, got ${JSON.stringify(smokeDock.env_keys)}`);
  await openHome(page);
  await openHomeDisclosure(page, "#home-context-disclosure");
  await page.locator('#workspace-session-hub [data-panel="intake"]').evaluate((button) => button.click());
  await page.locator("#panel-intake.active").waitFor();
  await page.getByLabel(/File intake path|文件星舱路径/).fill(intakeDoc);
  await page.getByLabel(/File intake mode|文件星舱模式/).selectOption("document_text");
  const intakeResponsePromise = page.waitForResponse((response) => response.url().endsWith("/intake/file") && response.request().method() === "POST");
  await page.locator("#intake-form").getByRole("button", { name: /Analyze file|分析文件/ }).click();
  const intakeResponse = await intakeResponsePromise;
  const intakePayload = await intakeResponse.json();
  assert(String(intakePayload.content || "").includes("Astria smoke intake document"), `UI file intake response failed: ${JSON.stringify(intakePayload)}`);
  const directIntake = await page.evaluate(async (path) => {
    const response = await fetch("/intake/file", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path, mode: "document_text", max_chars: 12000 }),
    });
    return response.json();
  }, intakeDoc);
  assert(String(directIntake.content || "").includes("Astria smoke intake document"), `direct file intake failed: ${JSON.stringify(directIntake)}`);
  await page.locator("#intake-result").getByText("Astria smoke intake document").waitFor();
  await page.getByRole("button", { name: /Send to Chat|发送到对话/ }).click();
  await page.locator("#panel-chat.active").waitFor();
  await page.locator("#chat-input").inputValue().then((value) => {
    assert(value.includes("Astria smoke intake document"), "chat prompt should include intake result");
  });
  await openHome(page);
  await openCommandRoute(page, "inbox", /Inbox/, "panel:inbox");
  await page.locator("#panel-inbox.active").waitFor();
  await openHome(page);
  await page.locator('.constellation-card[data-panel="memory"]').evaluate((button) => button.click());
  await page.locator("#panel-memory.active").waitFor();
  await page.locator("#memory-taxonomy-bar").getByRole("button", { name: /Decisions|决策/ }).waitFor();
  await page.getByLabel(/Memory candidate|记忆候选/).fill("- [decision] Keep Astria UI calm and native.");
  await page.locator("#memory-candidate-preview").getByText("Decisions").waitFor();
  await openHome(page);
  await openCommandRoute(page, "mcp", /MCP/, "panel:mcp");
  await page.locator("#panel-mcp.active").waitFor();
  await openHome(page);
  await openCommandRoute(page, "council", /Agent Council|智能体议会/, "panel:council");
  await page.locator("#panel-council.active").waitFor();
  await page.getByLabel(/Council goal|议会目标/).fill("webui council smoke");
  await page.getByRole("button", { name: /Start council|启动议会/ }).click();
  await page.locator("#council-detail").getByRole("heading", { name: "webui council smoke" }).waitFor();
  await page.locator("#council-detail").getByRole("heading", { name: /Council stages|议会阶段/ }).waitFor();
  await page.locator(".council-stage-card").nth(4).waitFor();
  await page.locator(".council-role-list").getByText("planner", { exact: true }).waitFor();
  await page.locator(".council-role-list").getByText("researcher", { exact: true }).waitFor();
  await page.locator(".council-role-list").getByText("reviewer", { exact: true }).waitFor();
  await page.locator(".council-stage-card").filter({ hasText: "planner" }).getByRole("button", { name: /Copy notes|复制笔记/ }).click();
  await page.getByText("Council role notes copied.").waitFor();
  const councilRoleNotes = await page.evaluate(() => navigator.clipboard.readText());
  assert(councilRoleNotes.includes("Role: planner"), "council role copy missing role");
  await page.locator(".council-stage-card").filter({ hasText: "planner" }).getByRole("button", { name: /Draft to chat|起草到对话/ }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Role: planner"), "council role draft missing role");
  await openCommandRoute(page, "council", /Agent Council|智能体议会/, "panel:council");
  await page.locator("#panel-council.active").waitFor();
  assert(await page.locator("#council-detail").getByText(/Synthesis|综合结论/).count() > 0, "council detail should include synthesis stage");
  await page.locator("#council-detail").getByText("Handoff", { exact: true }).waitFor();
  await page.locator("#council-detail").getByText(/Final synthesis|最终综合结论/).waitFor();
  const finalSynthesisSection = page.locator(".run-detail-section").filter({ has: page.getByRole("heading", { name: /Final synthesis|最终综合结论/ }) });
  await finalSynthesisSection.getByRole("button", { name: /Copy synthesis|复制综合结论/ }).click();
  await page.getByText("Council synthesis copied.").waitFor();
  const councilSynthesis = await page.evaluate(() => navigator.clipboard.readText());
  assert(councilSynthesis.includes("Council synthesis for: webui council smoke"), "council synthesis copy missing goal");
  await page.locator(".council-stage-card").filter({ hasText: "Handoff" }).getByRole("button", { name: /Start run|启动运行/ }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.locator("#run-detail .run-meta-grid").getByText("council_handoff", { exact: true }).waitFor();
  await openManagePanel(page, "Comparison Workbench");
  await page.locator("#panel-compare.active").waitFor();
  await page.locator("#panel-compare").getByRole("heading", { name: "路径比较台" }).waitFor();
  await page.locator(".comparison-lane").nth(2).waitFor();
  await page.locator('[data-compare-lane="recent-runs"]').getByText(/完成|失败|最近/).first().waitFor();
  await page.locator("#comparison-detail").getByRole("heading", { name: "证据", exact: true }).waitFor();
  await page.locator('[data-compare-lane="recent-runs"]').getByRole("button", { name: "起草比较" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Compare recent Astria runs"), "comparison draft missing run prompt");
  await openManagePanel(page, "Comparison Workbench");
  await page.locator("#panel-compare.active").waitFor();
  await page.locator('[data-compare-lane="council-synthesis"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-council.active").waitFor();
  await openManagePanel(page, "Run Quality");
  await page.locator("#panel-quality.active").waitFor();
  await page.locator("#panel-quality").getByRole("heading", { name: "运行质量" }).waitFor();
  await page.locator(".run-quality-card").nth(6).waitFor();
  await page.locator('[data-run-quality="latest-run"]').getByText("quality score", { exact: true }).waitFor();
  await page.locator('[data-run-quality="completed-output"]').getByText("已完成产物就绪度", { exact: true }).waitFor();
  await page.locator('[data-run-quality="failure-retry"]').getByText("失败与重试风险", { exact: true }).waitFor();
  await page.locator('[data-run-quality="evidence-quality"]').getByText("证据质量评分", { exact: true }).waitFor();
  await page.locator('[data-run-quality="budget-posture"]').getByText("预算与停止规则姿态", { exact: true }).waitFor();
  await page.locator('[data-run-quality="reuse-readiness"]').getByText("可复用产物就绪度", { exact: true }).waitFor();
  await page.locator('[data-run-quality="delivery-readiness"]').getByText("交付就绪度评分", { exact: true }).waitFor();
  await page.locator("#run-quality-detail").getByRole("heading", { name: "信号", exact: true }).waitFor();
  await page.locator('[data-run-quality="budget-posture"]').getByRole("button", { name: "质量简报" }).click();
  await page.locator("#run-quality-detail").getByText("预算形态、模型路线、降级阶梯和停止条件必须清楚。", { exact: true }).waitFor();
  await page.locator('[data-run-quality="budget-posture"]').getByRole("button", { name: "起草复查" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const qualityDraft = await page.locator("#chat-input").inputValue();
  assert(qualityDraft.includes("Evaluate Astria budget posture"), "run quality draft missing budget posture prompt");
  assert(qualityDraft.includes("Run Quality review"), "run quality draft missing review shape");
  await openManagePanel(page, "Run Quality");
  await page.locator("#panel-quality.active").waitFor();
  await page.locator('[data-run-quality="budget-posture"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-budget.active").waitFor();
  await openManagePanel(page, "Prompt Lab");
  await page.locator("#panel-promptlab.active").waitFor();
  await page.locator("#panel-promptlab").getByRole("heading", { name: "Prompt 实验室" }).waitFor();
  await page.getByLabel("Prompt 实验目标").fill("Ship the next Kocoro parity slice");
  await page.locator(".prompt-variant").nth(3).waitFor();
  await page.locator('[data-prompt-variant="direct"]').getByText("直接执行").waitFor();
  await page.locator('[data-prompt-variant="evidence"]').getByText("证据优先实验").waitFor();
  await page.locator('[data-prompt-variant="council"]').getByText("议会评审变体").waitFor();
  await page.locator('[data-prompt-variant="delivery"]').getByText("交付就绪变体").waitFor();
  await page.locator("#promptlab-detail").getByRole("heading", { name: "评估方式", exact: true }).waitFor();
  await page.locator('[data-prompt-variant="evidence"]').getByRole("button", { name: "起草变体" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Run an evidence-first Astria prompt experiment"), "prompt lab draft missing evidence variant");
  await openManagePanel(page, "Prompt Lab");
  await page.locator("#panel-promptlab.active").waitFor();
  await page.locator('[data-prompt-variant="council"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-council.active").waitFor();
  await openManagePanel(page, "Budget Guard");
  await page.locator("#panel-budget.active").waitFor();
  await page.locator("#panel-budget").getByRole("heading", { name: "预算守卫" }).waitFor();
  await page.locator(".budget-guard-card").nth(6).waitFor();
  await page.locator('[data-budget-guard="hard-cap"]').getByText("硬预算上限", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="model-route"]').getByText("按复杂度选择模型路线", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="context-trim"]').getByText("上下文裁剪", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="fallback"]').getByText("自动降级阶梯", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="stop-rules"]').getByText("长运行停止规则", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="schedule-limit"]').getByText("定时工作预算", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="evidence-cost"]').getByText("证据成本取舍", { exact: true }).waitFor();
  await page.locator("#budget-guard-detail").getByRole("heading", { name: "执行护栏", exact: true }).waitFor();
  await page.locator('[data-budget-guard="model-route"]').getByRole("button", { name: "预算简报" }).click();
  await page.locator("#budget-guard-detail").getByText("先分类为简单、证据密集、需要议会或交付敏感，再选择路线。", { exact: true }).waitFor();
  await page.locator('[data-budget-guard="model-route"]').getByRole("button", { name: "起草守卫" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const budgetDraft = await page.locator("#chat-input").inputValue();
  assert(budgetDraft.includes("Plan an Astria complexity-based model route"), "budget guard draft missing model route prompt");
  assert(budgetDraft.includes("Ship the next Kocoro parity slice"), "budget guard draft missing prompt lab goal");
  await openManagePanel(page, "Budget Guard");
  await page.locator("#panel-budget.active").waitFor();
  await page.locator('[data-budget-guard="model-route"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-promptlab.active").waitFor();
  await openManagePanel(page, "Source Registry");
  await page.locator("#panel-sources.active").waitFor();
  await page.locator("#panel-sources").getByRole("heading", { name: /Source Registry|来源登记/ }).waitFor();
  await page.locator(".source-row").nth(4).waitFor();
  assert(await page.locator('[data-source-row="memory"]').getByText(/Reviewed memory|已审查记忆/).count() > 0, "source registry should include memory lane");
  assert(await page.locator('[data-source-row="sessions"]').getByText(/Favorite sessions|收藏会话/).count() > 0, "source registry should include sessions lane");
  assert(await page.locator('[data-source-row="runs"]').getByText(/Execution evidence|执行证据/).count() > 0, "source registry should include runs lane");
  assert(await page.locator('[data-source-row="intake"]').getByText(/File Intake|文件星舱/).count() > 0, "source registry should include intake lane");
  assert(await page.locator('[data-source-row="council"]').getByText(/Council|议会/).count() > 0, "source registry should include council lane");
  await page.locator("#source-registry-detail").getByRole("heading", { name: "可靠性", exact: true }).waitFor();
  await page.locator('[data-source-row="runs"]').getByRole("button", { name: "起草维护" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Review recent runs as knowledge sources"), "source registry draft missing runs prompt");
  await openManagePanel(page, "Source Registry");
  await page.locator("#panel-sources.active").waitFor();
  await page.locator('[data-source-row="council"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-council.active").waitFor();
  await openManagePanel(page, "Knowledge Reconciliation");
  await page.locator("#panel-reconcile.active").waitFor();
  await page.locator("#panel-reconcile").getByRole("heading", { name: /Knowledge Reconciliation|知识校验/ }).waitFor();
  await page.locator(".reconcile-card").nth(6).waitFor();
  await page.locator('[data-reconcile-risk="source-conflict"]').getByText("Source conflict review", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="stale-memory"]').getByText("Stale memory review", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="weak-citation"]').getByText("Weak citation escalation", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="duplicate-memory"]').getByText("Duplicate or uncategorized memory", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="missing-coverage"]').getByText("Missing source coverage", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="privacy-boundary"]').getByText("Privacy and approval boundary", { exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="result-freshness"]').getByText("Result freshness review", { exact: true }).waitFor();
  await page.locator("#knowledge-reconcile-detail").getByRole("heading", { name: "可信边界", exact: true }).waitFor();
  await page.locator('[data-reconcile-risk="weak-citation"]').getByRole("button", { name: "起草解决" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const reconcileDraft = await page.locator("#chat-input").inputValue();
  assert(reconcileDraft.includes("Run an Astria weak citation escalation"), "knowledge reconciliation draft missing weak citation prompt");
  assert(reconcileDraft.includes("confidence boundary"), "knowledge reconciliation draft missing resolution shape");
  await openManagePanel(page, "Knowledge Reconciliation");
  await page.locator("#panel-reconcile.active").waitFor();
  await page.locator('[data-reconcile-risk="privacy-boundary"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-share.active").waitFor();
  await openManagePanel(page, "Citation Planner");
  await page.locator("#panel-citation.active").waitFor();
  await page.locator("#panel-citation").getByRole("heading", { name: /Citation Planner|引用校准/ }).waitFor();
  await page.getByLabel(/Citation claim scope|引用声明范围/).fill("The release decision is supported by current docs and browser evidence.");
  await page.getByLabel(/Citation source posture|引用来源姿态/).fill("official docs plus local memory, fresh browser evidence required");
  await page.getByLabel(/Citation evidence level|引用证据等级/).fill("direct quote or dated source summary with gap report");
  await page.locator(".citation-grounding-card").nth(4).waitFor();
  await page.locator('[data-citation-grounding="coverage"]').getByText("Source coverage check", { exact: true }).waitFor();
  await page.locator('[data-citation-grounding="claim-map"]').getByText("Claim-to-citation map", { exact: true }).waitFor();
  await page.locator('[data-citation-grounding="quote-capture"]').getByText("Quote and evidence capture", { exact: true }).waitFor();
  await page.locator('[data-citation-grounding="freshness"]').getByText("Freshness and version risk", { exact: true }).waitFor();
  await page.locator('[data-citation-grounding="gap-escalation"]').getByText("Evidence gap escalation", { exact: true }).waitFor();
  await page.locator("#citation-grounding-detail").getByRole("heading", { name: "Citation rule", exact: true }).waitFor();
  await page.locator('[data-citation-grounding="coverage"]').getByRole("button", { name: "起草校准" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const citationDraft = await page.locator("#chat-input").inputValue();
  assert(citationDraft.includes("Plan an Astria source coverage check"), "citation planner draft missing coverage prompt");
  assert(citationDraft.includes("release decision is supported"), "citation planner draft missing claim scope");
  assert(citationDraft.includes("fresh browser evidence required"), "citation planner draft missing source posture");
  await openManagePanel(page, "Citation Planner");
  await page.locator("#panel-citation.active").waitFor();
  await page.locator('[data-citation-grounding="coverage"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-sources.active").waitFor();
  await openManagePanel(page, "Citation Planner");
  await page.locator("#panel-citation.active").waitFor();
  await page.locator('[data-citation-grounding="quote-capture"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-browser.active").waitFor();
  await openManagePanel(page, "Citation Planner");
  await page.locator("#panel-citation.active").waitFor();
  await page.locator('[data-citation-grounding="freshness"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-data.active").waitFor();
  await openManagePanel(page, "Citation Planner");
  await page.locator("#panel-citation.active").waitFor();
  await page.locator('[data-citation-grounding="gap-escalation"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-share.active").waitFor();
  await openManagePanel(page, "复用星库");
  await page.locator("#panel-reuse.active").waitFor();
  await page.locator("#panel-reuse").getByRole("heading", { name: "复用星库" }).waitFor();
  await page.locator(".reuse-asset").nth(4).waitFor();
  await page.locator('[data-reuse-asset="prompt-direct"]').getByText("Prompt", { exact: true }).waitFor();
  await page.locator('[data-reuse-asset="agent-default"]').getByText("Agent", { exact: true }).waitFor();
  await page.locator('[data-reuse-asset="source-memory"]').getByText("知识", { exact: true }).waitFor();
  await page.locator("#reuse-gallery-detail").getByRole("heading", { name: "复用价值", exact: true }).waitFor();
  await page.locator('[data-reuse-asset="prompt-evidence"]').getByRole("button", { name: "起草任务" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Reuse this Astria prompt asset"), "reuse gallery draft missing prompt asset starter");
  await openManagePanel(page, "复用星库");
  await page.locator("#panel-reuse.active").waitFor();
  await page.locator('[data-reuse-asset="source-memory"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-memory.active").waitFor();
  await openManagePanel(page, "产物星库");
  await page.locator("#panel-results.active").waitFor();
  await page.locator("#panel-results").getByRole("heading", { name: "产物星库" }).waitFor();
  await page.locator(".result-library-card").nth(4).waitFor();
  await page.locator(".result-library-card").getByText("运行报告", { exact: true }).first().waitFor();
  await page.locator('[data-result-archive="share-result-brief"]').getByText("交接总览简报", { exact: true }).waitFor();
  await page.locator('[data-result-archive="data-result-trend"]').getByText("洞察简报", { exact: true }).waitFor();
  await page.locator('[data-result-archive="citation-result-coverage"]').getByText("引用简报", { exact: true }).waitFor();
  await page.locator('[data-result-archive="reuse-result-prompt-direct"]').getByText("可复用产物", { exact: true }).waitFor();
  await page.locator("#result-library-detail").getByRole("heading", { name: "鲜度", exact: true }).waitFor();
  await page.locator('[data-result-archive="citation-result-coverage"]').getByRole("button", { name: "起草后续" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const resultDraft = await page.locator("#chat-input").inputValue();
  assert(resultDraft.includes("Archive review: save the claim map"), "result library draft missing archive review prompt");
  assert(resultDraft.includes("source evidence"), "result library draft missing follow-up shape");
  await openManagePanel(page, "产物星库");
  await page.locator("#panel-results.active").waitFor();
  await page.locator('[data-result-archive="share-result-brief"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-share.active").waitFor();
  await openManagePanel(page, "Playbook Library");
  await page.locator("#panel-playbooks.active").waitFor();
  await page.locator("#panel-playbooks").getByRole("heading", { name: "实践手册" }).waitFor();
  await page.locator(".playbook-card").nth(7).waitFor();
  await page.locator('[data-playbook="reviewed-research"]').getByText("已审查证据调研", { exact: true }).waitFor();
  await page.locator('[data-playbook="data-insight"]').getByText("可审查数据洞察", { exact: true }).waitFor();
  await page.locator('[data-playbook="handoff-pack"]').getByText("本地交接包", { exact: true }).waitFor();
  await page.locator('[data-playbook="citation-grounding"]').getByText("结论溯源复查", { exact: true }).waitFor();
  await page.locator('[data-playbook="agent-profile"]').getByText("聚焦 Agent 配置", { exact: true }).waitFor();
  await page.locator('[data-playbook="memory-curation"]').getByText("长期记忆整理", { exact: true }).waitFor();
  await page.locator('[data-playbook="delivery-review"]').getByText("审批优先交付", { exact: true }).waitFor();
  await page.locator('[data-playbook="council-decision"]').getByText("多角色决策复查", { exact: true }).waitFor();
  await page.locator("#playbook-library-detail").getByRole("heading", { name: "证据门槛", exact: true }).waitFor();
  await page.locator('[data-playbook="reviewed-research"]').getByRole("button", { name: "起草手册" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const playbookDraft = await page.locator("#chat-input").inputValue();
  assert(playbookDraft.includes("Run the Astria reviewed evidence research playbook"), "playbook library draft missing research playbook");
  assert(playbookDraft.includes("Evidence gate"), "playbook library draft missing evidence gate");
  await openManagePanel(page, "Playbook Library");
  await page.locator("#panel-playbooks.active").waitFor();
  await page.locator('[data-playbook="data-insight"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-data.active").waitFor();
  await openManagePanel(page, "Starter Kits");
  await page.locator("#panel-starter.active").waitFor();
  await page.locator("#panel-starter").getByRole("heading", { name: "启动套件" }).waitFor();
  await page.locator(".starter-kit-card").nth(5).waitFor();
  await page.locator('[data-starter-kit="browser-research"]').getByText("已审查网页调研", { exact: true }).waitFor();
  await page.locator('[data-starter-kit="data-insight"]').getByText("本地数据洞察简报", { exact: true }).waitFor();
  await page.locator('[data-starter-kit="agent-build"]').getByText("聚焦 Agent 配置", { exact: true }).waitFor();
  await page.locator('[data-starter-kit="share-handoff"]').getByText("本地交接包", { exact: true }).waitFor();
  await page.locator('[data-starter-kit="memory-curation"]').getByText("长期记忆整理", { exact: true }).waitFor();
  await page.locator('[data-starter-kit="reuse-polish"]').getByText("可复用工作流打磨", { exact: true }).waitFor();
  await page.locator("#starter-kit-detail").getByRole("heading", { name: "安全边界", exact: true }).waitFor();
  await page.locator('[data-starter-kit="browser-research"]').getByRole("button", { name: "起草套件" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const starterDraft = await page.locator("#chat-input").inputValue();
  assert(starterDraft.includes("Launch the Astria reviewed web research starter kit"), "starter kit draft missing browser research prompt");
  assert(starterDraft.includes("Agent posture"), "starter kit draft missing agent posture");
  assert(starterDraft.includes("Reusable output"), "starter kit draft missing reusable output");
  await openManagePanel(page, "Starter Kits");
  await page.locator("#panel-starter.active").waitFor();
  await page.locator('[data-starter-kit="browser-research"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-browser.active").waitFor();
  await openManagePanel(page, "Starter Kits");
  await page.locator("#panel-starter.active").waitFor();
  await page.locator('[data-starter-kit="data-insight"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-data.active").waitFor();
  await openManagePanel(page, "Starter Kits");
  await page.locator("#panel-starter.active").waitFor();
  await page.locator('[data-starter-kit="share-handoff"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-share.active").waitFor();
  await openManagePanel(page, "Starter Kits");
  await page.locator("#panel-starter.active").waitFor();
  await page.locator('[data-starter-kit="memory-curation"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-memory.active").waitFor();
  await openManagePanel(page, "Share Pack");
  await page.locator("#panel-share.active").waitFor();
  await page.locator("#panel-share").getByRole("heading", { name: "交接包" }).waitFor();
  await page.getByLabel("交接包名称").fill("release research handoff");
  await page.getByLabel("交接包接收对象").fill("future reviewer");
  await page.getByLabel("交接包意图").fill("Reuse verified evidence and continue the release decision.");
  await page.locator(".share-pack-card").nth(4).waitFor();
  await page.locator('[data-share-pack="brief"]').getByText("交接总览简报", { exact: true }).waitFor();
  await page.locator('[data-share-pack="evidence"]').getByText("证据包清单", { exact: true }).waitFor();
  await page.locator('[data-share-pack="prompt"]').getByText("可复用 Prompt 起点", { exact: true }).waitFor();
  await page.locator('[data-share-pack="knowledge"]').getByText("记忆交接记录", { exact: true }).waitFor();
  await page.locator('[data-share-pack="review"]').getByText("审查者验收清单", { exact: true }).waitFor();
  await page.locator("#share-pack-detail").getByRole("heading", { name: "边界", exact: true }).waitFor();
  await page.locator('[data-share-pack="brief"]').getByRole("button", { name: "起草交接" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const shareDraft = await page.locator("#chat-input").inputValue();
  assert(shareDraft.includes("Build a local Astria share pack mission brief"), "share pack draft missing mission brief");
  assert(shareDraft.includes("release research handoff"), "share pack draft missing package name");
  assert(shareDraft.includes("future reviewer"), "share pack draft missing audience");
  await openManagePanel(page, "Share Pack");
  await page.locator("#panel-share.active").waitFor();
  await page.locator('[data-share-pack="prompt"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-reuse.active").waitFor();
  await openManagePanel(page, "Share Pack");
  await page.locator("#panel-share.active").waitFor();
  await page.locator('[data-share-pack="knowledge"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-memory.active").waitFor();
  await openManagePanel(page, "Workspace Snapshot");
  await page.locator("#panel-snapshot.active").waitFor();
  await page.locator("#panel-snapshot").getByRole("heading", { name: "工作区快照" }).waitFor();
  await page.locator(".workspace-snapshot-card").nth(6).waitFor();
  await page.locator('[data-workspace-snapshot="resume"]').getByText("会话续接快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="evidence"]').getByText("运行证据快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="memory-source"]').getByText("记忆与来源快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="result-archive"]').getByText("产物归档快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="playbook-reuse"]').getByText("手册与复用快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="delivery-schedule"]').getByText("交付与定时快照", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="privacy"]').getByText("脱敏与交接边界", { exact: true }).waitFor();
  await page.locator("#workspace-snapshot-detail").getByRole("heading", { name: "已包含上下文", exact: true }).waitFor();
  await page.locator("#workspace-snapshot-detail").getByRole("heading", { name: "隐私边界", exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="privacy"]').getByRole("button", { name: "快照简报" }).click();
  await page.locator("#workspace-snapshot-detail").getByText("默认本地优先。脱敏凭证、私有路径、用户数据和隐藏状态。", { exact: true }).waitFor();
  await page.locator('[data-workspace-snapshot="privacy"]').getByRole("button", { name: "起草快照" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const snapshotDraft = await page.locator("#chat-input").inputValue();
  assert(snapshotDraft.includes("Build an Astria redaction and handoff-boundary snapshot"), "workspace snapshot draft missing privacy prompt");
  assert(snapshotDraft.includes("Privacy/redaction boundary") || snapshotDraft.includes("privacy/redaction boundary"), "workspace snapshot draft missing privacy boundary shape");
  await openManagePanel(page, "Workspace Snapshot");
  await page.locator("#panel-snapshot.active").waitFor();
  await page.locator('[data-workspace-snapshot="result-archive"]').getByRole("button", { name: "打开路径" }).click();
  await page.locator("#panel-results.active").waitFor();
  await openManagePanel(page, "Browser Planner");
  await page.locator("#panel-browser.active").waitFor();
  await page.locator("#panel-browser").getByRole("heading", { name: "浏览器规划器" }).waitFor();
  await page.getByLabel("浏览器目标 URL").fill("https://example.com/research");
  await page.getByLabel("浏览器任务目标").fill("Verify the release notes and capture visual evidence.");
  await page.locator(".browser-mission-card").nth(4).waitFor();
  await page.locator('[data-browser-mission="inspect"]').getByText("检查", { exact: true }).waitFor();
  await page.locator('[data-browser-mission="screenshot"]').getByText("截图", { exact: true }).waitFor();
  await page.locator('[data-browser-mission="extract"]').getByText("抽取", { exact: true }).waitFor();
  await page.locator('[data-browser-mission="form-check"]').getByText("表单检查", { exact: true }).waitFor();
  await page.locator('[data-browser-mission="monitor"]').getByText("监控", { exact: true }).waitFor();
  await page.locator("#browser-mission-detail").getByRole("heading", { name: "风险", exact: true }).waitFor();
  await page.locator('[data-browser-mission="inspect"]').getByRole("button", { name: "起草任务" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const browserDraft = await page.locator("#chat-input").inputValue();
  assert(browserDraft.includes("Plan a reviewed browser inspection mission"), "browser planner draft missing inspection mission");
  assert(browserDraft.includes("https://example.com/research"), "browser planner draft missing target URL");
  assert(browserDraft.includes("Verify the release notes"), "browser planner draft missing goal");
  await openManagePanel(page, "Browser Planner");
  await page.locator("#panel-browser.active").waitFor();
  await page.locator('[data-browser-mission="monitor"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-schedules.active").waitFor();
  await openManagePanel(page, "Data Planner");
  await page.locator("#panel-data.active").waitFor();
  await page.locator("#panel-data").getByRole("heading", { name: "数据规划器" }).waitFor();
  await page.getByLabel("数据来源描述").fill("revenue_metrics.csv with weekly plan and actual columns");
  await page.getByLabel("数据分析问题").fill("Which segments are drifting from plan and need review?");
  await page.getByLabel("数据输出格式").fill("ranked findings with chart brief and reusable memory");
  await page.locator(".data-insight-card").nth(4).waitFor();
  await page.locator('[data-data-insight="profile"]').getByText("画像", { exact: true }).waitFor();
  await page.locator('[data-data-insight="trend"]').getByText("趋势", { exact: true }).waitFor();
  await page.locator('[data-data-insight="anomaly"]').getByText("异常", { exact: true }).waitFor();
  await page.locator('[data-data-insight="visual"]').getByText("图表简报", { exact: true }).waitFor();
  await page.locator('[data-data-insight="knowledge"]').getByText("知识", { exact: true }).waitFor();
  await page.locator("#data-insight-detail").getByRole("heading", { name: "护栏", exact: true }).waitFor();
  await page.locator('[data-data-insight="trend"]').getByRole("button", { name: "起草分析" }).click();
  await page.locator("#panel-chat.active").waitFor();
  const dataDraft = await page.locator("#chat-input").inputValue();
  assert(dataDraft.includes("Plan an Astria trend analysis mission"), "data planner draft missing trend mission");
  assert(dataDraft.includes("revenue_metrics.csv"), "data planner draft missing source descriptor");
  assert(dataDraft.includes("Which segments are drifting from plan"), "data planner draft missing analysis question");
  await openManagePanel(page, "Data Planner");
  await page.locator("#panel-data.active").waitFor();
  await page.locator('[data-data-insight="visual"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-reuse.active").waitFor();
  await openManagePanel(page, "Data Planner");
  await page.locator("#panel-data.active").waitFor();
  await page.locator('[data-data-insight="knowledge"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-memory.active").waitFor();
  await openHome(page);
  await openManagePanel(page, "Inbox");
  await page.locator("#panel-inbox").getByRole("heading", { name: /Inbox|收件箱/ }).waitFor();
  await page.locator("#inbox-provider-list").getByText("GitHub", { exact: true }).waitFor();
  await page.locator("#inbox-provider-list").getByText("/inbox/github").waitFor();
  await page.locator("#inbox-provider-list").getByText("issues").waitFor();
  await page.getByLabel(/Inbox external id|收件箱外部 ID/).fill("evt-smoke-core");
  await page.getByLabel(/Inbox sender|收件箱发送方/).fill("webui-smoke");
  await page.getByLabel(/Inbox text|收件箱内容/).fill("Review this inbound smoke task.");
  await page.getByRole("button", { name: /Receive webhook|接收 webhook/ }).click();
  await page.getByText(/Inbound task received\.|收件任务已接收。/).waitFor();
  await page.locator("#inbox-list").getByText("Review this inbound smoke task.").waitFor();
  await page.locator("#inbox-list").getByText("pending").waitFor();
  await page.locator("#inbox-list").getByRole("button", { name: /Reject|拒绝/ }).click();
  await page.getByText(/Inbox item rejected\.|收件项已拒绝。/).waitFor();
  await page.locator("#inbox-list .tag").getByText("rejected", { exact: true }).waitFor();
  await openManagePanel(page, "Schedules");
  await page.getByLabel(/Cron expression|Cron 表达式/).fill("* * * * *");
  await page.getByLabel(/Schedule prompt|定时任务 Prompt/).fill("webui smoke schedule");
  await page.getByRole("button", { name: /Create schedule|创建定时任务/ }).click();
  await page.locator("#schedules-list").getByText("webui smoke schedule").waitFor();
  await openManagePanel(page, "Proactive Delivery");
  await page.locator("#panel-delivery.active").waitFor();
  await page.locator("#panel-delivery").getByRole("heading", { name: "主动交付" }).waitFor();
  await page.locator(".delivery-lane").nth(3).waitFor();
  await page.locator('[data-delivery-lane="scheduled-work"]').getByText("webui smoke schedule").waitFor();
  await page.locator("#delivery-detail").getByRole("heading", { name: "证据", exact: true }).waitFor();
  await page.locator('[data-delivery-lane="scheduled-work"]').getByRole("button", { name: "起草交付" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert((await page.locator("#chat-input").inputValue()).includes("Plan proactive Astria delivery from schedules"), "delivery draft missing schedule prompt");
  await openManagePanel(page, "Proactive Delivery");
  await page.locator("#panel-delivery.active").waitFor();
  await page.locator('[data-delivery-lane="channel-readiness"]').getByRole("button", { name: "打开来源" }).click();
  await page.locator("#panel-inbox.active").waitFor();
  await openManagePanel(page, "Schedules");
  await page.getByRole("button", { name: /Pause|暂停/ }).click();
  await page.getByRole("button", { name: /Enable|启用/ }).waitFor();
  await page.getByRole("button", { name: /Delete|删除/ }).click();
  await page.getByText(/No schedules configured\.|还没有配置定时任务。/).waitFor();
  await openPanel(page, "chat");
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
  await approvalCard.getByText(/Approval required|需要人工确认/).waitFor();
  await approvalCard.getByText("smoke approval").waitFor();
  await page.getByRole("button", { name: /Allow|允许/ }).click();
  await approvalCard.getByText(/allowed|已允许/).waitFor();
  await openHome(page);
  await closeHomeDisclosures(page);
  assert((await page.locator("#home-count-pending").textContent()) === "0", "home pending approval count should be 0");
  await openPanel(page, "chat");
  const eventStatus = await page.evaluate(async (url) => {
    const response = await fetch(`${url}/events`);
    await response.body.cancel();
    return response.status;
  }, baseURL);
  assert(eventStatus === 200, `events status = ${eventStatus}`);
}

async function runPermissions(page) {
  await openSettingsPanel(page, "Permissions");
  await page.locator("#panel-permissions").getByRole("heading", { name: /Permissions|权限/ }).waitFor();
  await page.locator("#permissions-form").getByText("允许目录").waitFor();
  await page.locator("#permissions-form").getByText("网络白名单").waitFor();
  await page.getByLabel(/Allowed directories|允许目录/).fill("~\n.\n/tmp/smoke");
  await page.getByLabel(/Allowed commands|允许命令/).fill("go test\nstarclaw version");
  await page.getByLabel(/Denied commands|拒绝命令/).fill("shutdown\nreboot");
  await page.getByLabel(/Network allowlist|网络白名单/).fill("api.github.com\nsmoke.example.com");
  await page.getByLabel(/Sensitive patterns|敏感规则/).fill("*.secret\n.env.smoke");
  await page.locator("#permissions-pending-preview").getByText(/Pending changes|待保存改动/).waitFor();
  await page.locator("#permissions-pending-preview").getByText(/Broad local access is allowed\.|允许范围较宽的本地访问。/).waitFor();
  await page.locator("#permissions-pending-preview").getByText(/Allowed directories|允许目录/).waitFor();
  await page.getByRole("button", { name: /Save permissions|保存权限/ }).click();
  await page.getByText(/Permissions saved\.|权限已保存。/).waitFor();
  await page.locator("#permissions-list").getByText("/tmp/smoke").waitFor();
  await page.locator("#permissions-list").getByText("starclaw version").waitFor();
  await page.locator("#permissions-list").getByText("smoke.example.com").waitFor();
  assert((await page.getByLabel(/Allowed directories|允许目录/).inputValue()).includes("/tmp/smoke"), "permissions editor should retain saved allowed dirs");
  await page.getByRole("button", { name: /Clear rules|清空规则/ }).click();
  await page.getByText(/Permissions saved\.|权限已保存。/).waitFor();
  await page.locator("#permissions-overview").getByText(/Built-in defaults|内置默认/).waitFor();
  await page.locator("#permissions-pending-preview").getByText(/No denied commands are configured\.|尚未配置拒绝命令。/).waitFor();
  await page.locator("#permissions-pending-preview").getByText(/No sensitive file patterns are configured\.|尚未配置敏感文件规则。/).waitFor();
  assert(await page.getByLabel(/Allowed directories|允许目录/).inputValue() === "", "clear rules should empty allowed dirs");
  assert(await page.getByLabel(/Allowed commands|允许命令/).inputValue() === "", "clear rules should empty allowed commands");
}

async function runAgents(page) {
  await openManagePanel(page, "Agents");
  await page.locator("#panel-agents").getByRole("heading", { name: /Agents|智能体/ }).waitFor();
  await page.getByRole("button", { name: /New agent|新建 Agent/ }).click();
  await page.locator("#agent-name").fill("smoke-agent");
  await page.locator("#agent-prompt").fill("You are a smoke test agent.");
  await page.locator("#agent-memory").fill("Remember smoke.");
  await page.locator("#agent-model").fill("smoke-model");
  await page.locator("#agent-reasoning-effort").fill("low");
  const agentToolsAllow = page.locator("#agent-tools-allow");
  const agentToolsDeny = page.locator("#agent-tools-deny");
  const agentAutoApprove = page.locator("#agent-auto-approve");
  const agentHeartbeatEvery = page.locator("#agent-heartbeat-every");
  const agentHeartbeatActiveHours = page.locator("#agent-heartbeat-active-hours");
  const agentHeartbeatModel = page.locator("#agent-heartbeat-model");
  const agentCommandName = page.locator("#agent-command-name");
  const agentCommandBody = page.locator("#agent-command-body");
  const newCommandButton = page.locator("#agent-command-new-button");
  const saveCommandButton = page.locator("#agent-command-save-button");
  const cancelCommandButton = page.locator("#agent-command-cancel-button");
  const deleteCommandButton = page.locator("#agent-command-delete-button");
  const saveAgentButton = page.locator("#agent-form button[type=\"submit\"]");
  await agentToolsAllow.fill("file_read\ngrep");
  await agentToolsDeny.fill("bash");
  await agentAutoApprove.check();
  await page.locator("#agent-permission-preview").getByText("file_read, grep").waitFor();
  await page.locator("#agent-permission-preview").getByText("bash").waitFor();
  assert(await page.locator("#agent-permission-preview").evaluate((node) => node.textContent.includes("Enabled") || node.textContent.includes("启用")), "agent permission preview should show auto approve enabled");
  await page.locator("#agent-permission-preview").getByText(/Auto approve is enabled for this agent\.|此 Agent 已启用自动确认。/).waitFor();
  await agentToolsDeny.fill("bash\ngrep");
  await page.locator("#agent-permission-preview").getByText(/Allow\/deny conflict: grep|允许\/拒绝冲突：grep/).waitFor();
  await agentToolsDeny.fill("bash");
  await agentHeartbeatEvery.fill("15m");
  await agentHeartbeatActiveHours.fill("09:00-17:00");
  await agentHeartbeatModel.fill("smoke-heartbeat-model");
  await agentCommandName.fill("review");
  await agentCommandBody.fill("Review recent smoke changes.");
  await saveCommandButton.click();
  await page.locator("#agent-command-list").getByText("review").waitFor();
  await saveAgentButton.click();
  await page.getByText(/Agent saved\.|Agent 已保存。/).waitFor();
  await page.locator("#agents-list").getByText("smoke-agent").waitFor();
  const digestCard = page.locator("#agent-continuity-digest .agent-continuity-card").filter({ hasText: "smoke-agent" });
  await digestCard.getByText(/No recorded runs yet\.|没有运行记录/).waitFor();
  await digestCard.getByText(/Profile memory|配置记忆/, { exact: true }).waitFor();
  const createdDigestMetrics = await digestCard.locator(".agent-continuity-metrics").innerText();
  assert(createdDigestMetrics.includes("RUNS\n0") || createdDigestMetrics.includes("运行\n0"), `agent digest should show zero runs, got ${JSON.stringify(createdDigestMetrics)}`);
  assert(createdDigestMetrics.includes("COMMANDS\n1") || createdDigestMetrics.includes("命令\n1"), `agent digest should show one command, got ${JSON.stringify(createdDigestMetrics)}`);
  await digestCard.getByRole("button", { name: /Continue|继续/, exact: true }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-agent").inputValue() === "smoke-agent", "digest continue should select smoke-agent");
  const digestContinueDraft = await page.locator("#chat-input").inputValue();
  assert(digestContinueDraft.includes("Continue as smoke-agent"), "digest continue should draft continuity prompt");
  assert(digestContinueDraft.includes("Recent runs: 0"), "digest continue should include zero-run continuity context");
  await openManagePanel(page, "Agents");
  await digestCard.getByRole("button", { name: /Draft memory|起草记忆/, exact: true }).click();
  await page.locator("#panel-memory.active").waitFor();
  const memoryDraft = await page.locator("#memory-candidate").inputValue();
  assert(memoryDraft.includes("Agent smoke-agent"), "digest memory draft should include agent name");
  assert(memoryDraft.includes("0 recorded runs"), "digest memory draft should include run continuity");
  await openManagePanel(page, "Agents");
  const agentContinuityRunID = "run_agent_continuity_smoke";
  const agentContinuityRun = {
    id: agentContinuityRunID,
    status: "completed",
    agent: "smoke-agent",
    session_id: "sess_agent_continuity_smoke",
    started_at: new Date().toISOString(),
    prompt: "agent continuity smoke prompt",
    request: { text: "agent continuity smoke prompt", agent: "smoke-agent", request_id: agentContinuityRunID },
    response: { session_id: "sess_agent_continuity_smoke", messages: ["agent continuity smoke response"], usage: { prompt_tokens: 7, completion_tokens: 8 } },
  };
  await page.route("**/runs", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ runs: [agentContinuityRun] }),
    });
  });
  await page.route(`**/runs/${agentContinuityRunID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(agentContinuityRun),
    });
  });
  await page.locator("#refresh-button").click();
  await digestCard.getByText("agent continuity smoke prompt").waitFor();
  const refreshedDigestMetrics = await digestCard.locator(".agent-continuity-metrics").innerText();
  assert(refreshedDigestMetrics.includes("RUNS\n1") || refreshedDigestMetrics.includes("运行\n1"), `agent digest should show one run, got ${JSON.stringify(refreshedDigestMetrics)}`);
  await digestCard.getByRole("button", { name: /Continue|继续/, exact: true }).click();
  await page.locator("#panel-chat.active").waitFor();
  const digestRunContinueDraft = await page.locator("#chat-input").inputValue();
  assert(digestRunContinueDraft.includes("Recent runs: 1"), "digest continue should include latest run count");
  assert(digestRunContinueDraft.includes("agent continuity smoke prompt"), "digest continue should include latest run prompt");
  await openManagePanel(page, "Agents");
  await digestCard.getByRole("button", { name: /Open latest run|打开最近运行/, exact: true }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), agentContinuityRunID);
  await openManagePanel(page, "Agents");
  const rosterCard = page.locator("#agent-capability-roster .agent-roster-card").filter({ hasText: "smoke-agent" });
  await rosterCard.getByText("smoke-model", { exact: true }).waitFor();
  await rosterCard.getByText("low", { exact: true }).waitFor();
  await rosterCard.getByText(/Auto approve|自动确认/, { exact: true }).waitFor();
  await rosterCard.getByText(/Memory|有记忆/, { exact: true }).waitFor();
  await rosterCard.getByText("15m · 09:00-17:00").waitFor();
  await rosterCard.getByText(/Heartbeat scheduled|已安排心跳/, { exact: true }).waitFor();
  await rosterCard.getByText(/Approval bypass|绕过确认/, { exact: true }).waitFor();
  const createdRosterMetrics = await rosterCard.locator(".agent-roster-metrics").innerText();
  assert(createdRosterMetrics.includes("ALLOW\n2") || createdRosterMetrics.includes("允许\n2"), `agent roster should show two allowed tools, got ${JSON.stringify(createdRosterMetrics)}`);
  assert(createdRosterMetrics.includes("DENY\n1") || createdRosterMetrics.includes("拒绝\n1"), `agent roster should show one denied tool, got ${JSON.stringify(createdRosterMetrics)}`);
  assert(createdRosterMetrics.includes("COMMANDS\n1") || createdRosterMetrics.includes("命令\n1"), `agent roster should show one command, got ${JSON.stringify(createdRosterMetrics)}`);
  await rosterCard.getByRole("button", { name: "/review", exact: true }).waitFor();
  const reviewCommandDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await rosterCard.getByRole("button", { name: "/review", exact: true }).click();
  await reviewCommandDetailPromise;
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-agent").inputValue() === "smoke-agent", "roster command launch should select smoke-agent");
  assert((await page.locator("#chat-input").inputValue()).trim() === "Review recent smoke changes.", "roster command launch should draft command body");
  await openManagePanel(page, "Agents");
  await rosterCard.getByRole("button", { name: /Chat|对话/, exact: true }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-agent").inputValue() === "smoke-agent", "roster chat action should select smoke-agent");
  assert((await page.locator("#chat-input").inputValue()).includes("Continue as smoke-agent"), "roster chat action should draft a prompt");
  await openManagePanel(page, "Agents");
  await rosterCard.getByRole("button", { name: /Test|测试/, exact: true }).click();
  await page.locator("#panel-agents.active").waitFor();
  assert(await page.locator("#agent-test-agent").inputValue() === "smoke-agent", "roster test action should select smoke-agent");
  assert((await page.locator("#agent-test-prompt").inputValue()).includes("Test smoke-agent"), "roster test action should draft a test prompt");
  await rosterCard.getByRole("button", { name: /Council|议会/, exact: true }).click();
  await page.locator("#panel-council.active").waitFor();
  assert(await page.locator("#council-agent").inputValue() === "smoke-agent", "roster council action should select smoke-agent");
  assert((await page.locator("#council-goal").inputValue()).includes("Use smoke-agent as the lead agent"), "roster council action should draft a council goal");
  await openManagePanel(page, "Agents");
  const createdDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await rosterCard.getByRole("button", { name: /Edit profile|编辑配置/ }).click();
  await createdDetailPromise;
  assert(await agentToolsAllow.inputValue() === "file_read\ngrep", "agent allow rules should reload after create");
  assert(await agentToolsDeny.inputValue() === "bash", "agent deny rules should reload after create");
  assert(await agentAutoApprove.isChecked(), "agent auto approve should reload after create");
  assert(await agentHeartbeatEvery.inputValue() === "15m", "agent heartbeat interval should reload after create");
  assert(await agentHeartbeatActiveHours.inputValue() === "09:00-17:00", "agent heartbeat active hours should reload after create");
  assert(await agentHeartbeatModel.inputValue() === "smoke-heartbeat-model", "agent heartbeat model should reload after create");
  await page.locator("#agent-command-list [data-agent-command=\"review\"]").click();
  assert((await agentCommandBody.inputValue()).trim() === "Review recent smoke changes.", "agent command should reload after create");
  await agentCommandName.fill("audit");
  await agentCommandBody.fill("Audit smoke changes.");
  await saveCommandButton.click();
  await page.locator("#agent-command-list").getByText("audit").waitFor();
  assert(await page.locator("#agent-command-list").getByText("review").count() === 0, "renamed command should remove old name before save");
  await newCommandButton.click();
  assert(await agentCommandName.inputValue() === "", "new command should reset command name");
  assert(await agentCommandBody.inputValue() === "", "new command should reset command body");
  await agentCommandName.fill("scratch");
  await agentCommandBody.fill("Scratch command.");
  await cancelCommandButton.click();
  assert(await agentCommandName.inputValue() === "", "cancel command should reset command name");
  assert(await agentCommandBody.inputValue() === "", "cancel command should reset command body");
  await page.locator("#agent-prompt").fill("You are an edited smoke agent.");
  await agentToolsAllow.fill("version, file_read");
  await agentToolsDeny.fill("bash\nhttp");
  await agentAutoApprove.uncheck();
  await agentHeartbeatEvery.fill("30m");
  await agentHeartbeatActiveHours.fill("10:00-18:00");
  await agentHeartbeatModel.fill("smoke-heartbeat-edited");
  await agentCommandName.fill("deploy");
  await agentCommandBody.fill("Deploy smoke changes safely.");
  await saveCommandButton.click();
  await page.locator("#agent-command-list [data-agent-command=\"audit\"]").click();
  await deleteCommandButton.click();
  const allowBeforeSave = await agentToolsAllow.inputValue();
  assert(allowBeforeSave === "version, file_read", `agent allow input should update before save, got ${JSON.stringify(allowBeforeSave)}`);
  const updateResponsePromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "PUT"
  );
  await saveAgentButton.click();
  const updateResponse = await updateResponsePromise;
  assert(updateResponse.ok(), `agent update failed with ${updateResponse.status()}`);
  const updatedAgent = await updateResponse.json();
  const updatedConfig = agentConfig(updatedAgent);
  const updatedCommands = agentCommands(updatedAgent);
  assert(updatedConfig.allow.join("\n") === "version\nfile_read", `agent allow rules should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(updatedConfig.deny.join("\n") === "bash\nhttp", `agent deny rules should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(updatedConfig.autoApprove === false, `agent auto approve should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(updatedConfig.heartbeatEvery === "30m", `agent heartbeat interval should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(updatedConfig.heartbeatActiveHours === "10:00-18:00", `agent heartbeat active hours should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(updatedConfig.heartbeatModel === "smoke-heartbeat-edited", `agent heartbeat model should save after edit, got ${JSON.stringify(updatedConfig)}`);
  assert(!updatedCommands.audit, `deleted renamed command should not save after edit, got ${JSON.stringify(updatedCommands)}`);
  assert(!updatedCommands.review, `renamed command should not keep old name after edit, got ${JSON.stringify(updatedCommands)}`);
  assert(updatedCommands.deploy === "Deploy smoke changes safely.\n", `agent command should save after edit, got ${JSON.stringify(updatedCommands)}`);
  await page.locator("#agent-memory").fill("Unsaved smoke memory.");
  await page.getByText(/Unsaved changes|未保存改动/).waitFor();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "unsaved new-agent dialog should be a confirm");
    await dialog.dismiss();
  });
  await page.getByRole("button", { name: /New agent|新建 Agent/ }).click();
  assert(await page.locator("#agent-memory").inputValue() === "Unsaved smoke memory.", "dismissed dirty dialog should keep editor values");
  await page.locator("#agent-memory").fill("Remember smoke.");
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: /Export config|导出配置/ }).click();
  const download = await downloadPromise;
  const exportedPath = await download.path();
  const exportedAgent = JSON.parse(fs.readFileSync(exportedPath, "utf8"));
  assert(exportedAgent.name === "smoke-agent", "exported config should include agent name");
  assert(exportedAgent.commands.deploy.trim() === "Deploy smoke changes safely.", "exported config should include staged command body");
  exportedAgent.memory = "Imported smoke memory.";
  exportedAgent.tools_allow = ["version", "file_read", "grep"];
  exportedAgent.commands.imported = "Imported command.";
  const importPath = `${process.env.NODE_DIR}/imported-agent.json`;
  fs.writeFileSync(importPath, JSON.stringify(exportedAgent));
  await page.locator("#agent-import-file").setInputFiles(importPath);
  await page.getByText(/Agent config imported\. Save agent to apply\.|Agent 配置已导入。保存后生效。/).waitFor();
  assert(await page.locator("#agent-memory").inputValue() === "Imported smoke memory.", "import should update memory field");
  assert(await agentToolsAllow.inputValue() === "version\nfile_read\ngrep", "import should update allow rules");
  await page.locator("#agent-command-list").getByText("imported").waitFor();
  await page.getByText(/Unsaved changes|未保存改动/).waitFor();
  const importSavePromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "PUT"
  );
  await saveAgentButton.click();
  const importSaveResponse = await importSavePromise;
  assert(importSaveResponse.ok(), `agent import save failed with ${importSaveResponse.status()}`);
  await rosterCard.getByText(/Manual review|人工复查/, { exact: true }).waitFor();
  await rosterCard.getByText(/Approval gated|确认闸门/, { exact: true }).waitFor();
  await rosterCard.getByText("30m · 10:00-18:00").waitFor();
  const updatedRosterMetrics = await rosterCard.locator(".agent-roster-metrics").innerText();
  assert(updatedRosterMetrics.includes("ALLOW\n3") || updatedRosterMetrics.includes("允许\n3"), `agent roster should show imported allowed tool count, got ${JSON.stringify(updatedRosterMetrics)}`);
  assert(updatedRosterMetrics.includes("DENY\n2") || updatedRosterMetrics.includes("拒绝\n2"), `agent roster should show edited denied tool count, got ${JSON.stringify(updatedRosterMetrics)}`);
  assert(updatedRosterMetrics.includes("COMMANDS\n2") || updatedRosterMetrics.includes("命令\n2"), `agent roster should show deploy and imported commands, got ${JSON.stringify(updatedRosterMetrics)}`);
  await rosterCard.getByRole("button", { name: "/deploy", exact: true }).waitFor();
  await rosterCard.getByRole("button", { name: "/imported", exact: true }).waitFor();
  await page.getByRole("button", { name: /New agent|新建 Agent/ }).click();
  const updatedDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await page.locator("#agents-list [data-agent-detail=\"smoke-agent\"]").click();
  await updatedDetailPromise;
  const editedAllow = await agentToolsAllow.inputValue();
  const editedDeny = await agentToolsDeny.inputValue();
  assert(editedAllow === "version\nfile_read\ngrep", `agent allow rules should reload after import save, got ${JSON.stringify(editedAllow)}`);
  assert(editedDeny === "bash\nhttp", `agent deny rules should reload after edit, got ${JSON.stringify(editedDeny)}`);
  assert((await page.locator("#agent-memory").inputValue()).trim() === "Imported smoke memory.", "agent memory should reload after import save");
  assert(!(await agentAutoApprove.isChecked()), "agent auto approve should reload after edit");
  assert(await agentHeartbeatEvery.inputValue() === "30m", "agent heartbeat interval should reload after edit");
  assert(await agentHeartbeatActiveHours.inputValue() === "10:00-18:00", "agent heartbeat active hours should reload after edit");
  assert(await agentHeartbeatModel.inputValue() === "smoke-heartbeat-edited", "agent heartbeat model should reload after edit");
  assert(await page.locator("#agent-command-list").getByText("deploy").count() === 1, "agent command list should reload after edit");
  assert(await page.locator("#agent-command-list").getByText("imported").count() === 1, "imported agent command should reload after import save");
  assert(await page.locator("#agent-command-list").getByText("review").count() === 0, "deleted agent command should stay deleted after reload");
  assert(await page.locator("#agent-command-list").getByText("audit").count() === 0, "deleted renamed agent command should stay deleted after reload");
  await page.locator("#agent-command-list [data-agent-command=\"deploy\"]").click();
  assert((await agentCommandBody.inputValue()).trim() === "Deploy smoke changes safely.", "agent command body should reload after edit");
  await page.locator("#agent-test-run-button").click();
  assert(await page.locator("#panel-agents.active").count() === 1, "test run should stay on agents panel");
  assert(await page.locator("#agent-test-agent").inputValue() === "smoke-agent", "test run should select edited agent in runner");
  assert((await page.locator("#agent-test-prompt").inputValue()).includes("Test smoke-agent"), "test run should prefill agent test prompt");
  let capturedAgentTest = null;
  let capturedAgentTestRequestID = "";
  const agentTestMessageRoute = (url) => url.pathname === "/message";
  await page.route(agentTestMessageRoute, async (route) => {
    const body = route.request().postDataJSON();
    if (body.text === "agent test cancellation smoke") {
      await new Promise((resolve) => setTimeout(resolve, 500));
      await fulfillIfUnhandled(route, {
        status: 200,
        contentType: "text/event-stream",
        body: `event: text\ndata: ${JSON.stringify({ text: "late cancelled response" })}\n\n`,
      });
      return;
    }
    capturedAgentTest = body;
    capturedAgentTestRequestID = body.request_id;
    await fulfillIfUnhandled(route, {
      status: 200,
      contentType: "text/event-stream",
      body: [
        `event: text\ndata: ${JSON.stringify({ text: "agent test streamed response" })}`,
        `event: tool_call\ndata: ${JSON.stringify({ tool: "version", args: "{}", status: "running" })}`,
        `event: tool_result\ndata: ${JSON.stringify({ tool: "version", content: "agent test tool result", status: "completed", is_error: false })}`,
        `event: usage\ndata: ${JSON.stringify({ input_tokens: 5, output_tokens: 6 })}`,
        `event: done\ndata: ${JSON.stringify({
          session_id: "sess_agent_test_smoke",
          messages: ["agent test streamed response", "agent test smoke response"],
          usage: { prompt_tokens: 5, completion_tokens: 6 },
        })}`,
        "",
      ].join("\n\n"),
    });
  });
  await page.locator("#agent-test-prompt").fill("agent test direct smoke");
  await page.locator("#agent-test-form").getByRole("button", { name: /Run test|运行测试/ }).click();
  await page.locator("#agent-test-stop-button").waitFor();
  await page.locator("#panel-agents.active").waitFor();
  await page.locator("#agent-test-output").getByText("Agent 测试结果").waitFor();
  await page.locator("#agent-test-output").getByText("agent test direct smoke").waitFor();
  await page.locator("#agent-test-output").getByText(capturedAgentTestRequestID).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "观测运行" }).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "打开会话" }).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "复制摘要" }).click();
  await page.getByText(/Agent test summary copied\.|Agent 测试摘要已复制。/).waitFor();
  const copiedAgentTestSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedAgentTestSummary.includes("Agent: smoke-agent"), "agent test summary missing agent");
  assert(copiedAgentTestSummary.includes("Prompt: agent test direct smoke"), "agent test summary missing prompt");
  assert(copiedAgentTestSummary.includes(capturedAgentTestRequestID), "agent test summary missing request id");
  assert(capturedAgentTest.agent === "smoke-agent", `agent test payload should use smoke-agent, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.text === "agent test direct smoke", `agent test payload should include prompt, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.new_session === true, `agent test payload should create a new session, got ${JSON.stringify(capturedAgentTest)}`);
  await openManagePanel(page, "Agents");
  await page.locator("#agent-test-prompt").fill("agent test cancellation smoke");
  await page.locator("#agent-test-form").getByRole("button", { name: /Run test|运行测试/ }).click();
  await page.locator("#agent-test-stop-button").waitFor();
  await page.locator("#agent-test-stop-button").click();
  await page.locator("#agent-test-output").getByText("Agent 测试已取消").waitFor();
  await page.locator("#agent-test-form").getByRole("button", { name: /Run test|运行测试/ }).waitFor();
  await page.unroute(agentTestMessageRoute);
  await page.locator("#agents-list [data-agent-detail=\"smoke-agent\"]").click();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "agent delete dialog should be a confirm");
    await dialog.accept();
  });
  await page.locator("#agent-delete-button").click();
  await page.getByText(/Agent deleted\.|Agent 已删除。/).waitFor();
}

async function runRuns(page) {
  await openPanel(page, "chat");
  await page.locator("#chat-agent").selectOption("");
  await page.locator("#chat-new-session").check();
  await page.locator("#chat-input").fill("webui smoke session");
  await page.route("**/message", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        session_id: "sess_summary_smoke",
        messages: ["summary smoke response"],
        usage: { prompt_tokens: 3, completion_tokens: 4 },
      }),
    });
  });
  await page.keyboard.press(process.platform === "darwin" ? "Meta+Enter" : "Control+Enter");
  const chatRunSummary = page.locator("#chat-output .run-summary");
  await chatRunSummary.waitFor();
  assert(await page.locator("#chat-input").evaluate((element) => document.activeElement === element), "chat input should regain focus after run");
  await chatRunSummary.getByText("运行摘要").waitFor();
  await chatRunSummary.getByText("Agent").waitFor();
  await chatRunSummary.getByText("default").waitFor();
  await chatRunSummary.getByText("用量").waitFor();
  await chatRunSummary.getByText("prompt_tokens: 3").waitFor();
  await chatRunSummary.getByText("Run ID").waitFor();
  await chatRunSummary.getByRole("button", { name: "观测运行" }).waitFor();
  await chatRunSummary.getByRole("button", { name: "复制摘要" }).click();
  await page.getByText("Run summary copied.").waitFor();
  await chatRunSummary.getByRole("button", { name: /Copied|已复制/ }).waitFor();
  await chatRunSummary.getByRole("button", { name: "复制摘要" }).waitFor();
  const copiedSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedSummary.includes("Session: sess_summary_smoke"), "copied summary missing session");
  assert(copiedSummary.includes("Agent: default"), "copied summary missing agent");
  assert(copiedSummary.includes("Usage: prompt_tokens: 3, completion_tokens: 4"), "copied summary missing usage");
  await chatRunSummary.getByRole("button", { name: "起草后续" }).click();
  await page.locator("#panel-home.active").waitFor();
  const summaryFollowUp = await homeTaskInput(page).inputValue();
  assert(summaryFollowUp.includes("Continue from this completed Astria run"), "summary follow-up should draft next prompt");
  assert(summaryFollowUp.includes("webui smoke session"), "summary follow-up should include original prompt");
  await openPanel(page, "chat");
  await chatRunSummary.getByRole("button", { name: "打开会话" }).waitFor();
  await page.unroute("**/message");
  let runID = "run_history_smoke";
  const sessionID = await page.evaluate(async ({ url, runID }) => {
    const response = await fetch(`${url}/message`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text: "webui smoke session", new_session: true, request_id: runID })
    });
    const data = await response.json();
    return data.session_id;
  }, { url: baseURL, runID });
  assert(sessionID, "session id missing");
  runID = await page.evaluate(async ({ url, requestedRunID }) => {
    const response = await fetch(`${url}/runs`);
    const data = await response.json();
    const runs = Array.isArray(data.runs) ? data.runs : [];
    const exact = runs.find((run) => run.id === requestedRunID);
    return exact?.id || runs[0]?.id || "";
  }, { url: baseURL, requestedRunID: runID });
  assert(runID, "run id missing from runs list");
  await page.locator("#refresh-button").click();
  await openPanel(page, "runs");
  const selectedRunDetail = await page.evaluate(async (id) => {
    await window.selectRun(id);
    return document.querySelector("#run-detail")?.textContent || "";
  }, runID);
  assert(selectedRunDetail.includes(runID), `run detail missing ${runID}: ${selectedRunDetail}`);
  await openHome(page);
  await openHomeDisclosure(page, "#home-context-disclosure");
  await page.locator("#workspace-session-hub").evaluate((hub) => {
    const button = [...hub.querySelectorAll("button")].find((item) => /webui smoke session|sess_/.test(item.textContent));
    if (!button) throw new Error("workspace session button not found");
    button.click();
  });
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "workspace hub should resume latest session");
  await page.locator("#command-center-button").click();
  await page.locator("#command-center-input").fill("run_history_smoke");
  await page.locator("#command-center-list").getByRole("button", { name: /run_history_smoke/ }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), runID);
  await openPanel(page, "runs");
  await page.locator("#panel-runs").getByRole("heading", { name: "运行观测台" }).waitFor();
  assert(await page.locator("#mission-control-board").evaluate((node) => node.textContent.includes("已完成")), "mission control should show completed count");
  assert(await page.locator("#mission-control-board").evaluate((node) => node.textContent.includes("已恢复")), "mission control should show recovered count");
  await page.locator('#mission-control-filters [data-run-filter="completed"]').evaluate((el) => el.click());
  await page.locator('#mission-control-filters [data-run-filter="completed"]').evaluate((el) => {
    if (!el.classList.contains("active")) throw new Error("completed run filter should be active");
  });
  await page.locator('#mission-control-filters [data-run-filter="all"]').evaluate((el) => el.click());
  const selectedRunDetailAfterFilter = await page.evaluate(async (id) => {
    await window.selectRun(id);
    return document.querySelector("#run-detail")?.textContent || "";
  }, runID);
  assert(selectedRunDetailAfterFilter.includes(runID), `run detail missing after filter ${runID}: ${selectedRunDetailAfterFilter}`);
  const now = new Date().toISOString();
  await page.route(`**/runs/${runID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        id: runID,
        status: "completed",
        agent: "",
        channel: "http",
        prompt: "webui smoke session",
        session_id: sessionID,
        started_at: now,
        ended_at: now,
        usage: { input_tokens: 7, output_tokens: 8 },
        budget_status: {
          status: "ok",
          input_tokens: 7,
          output_tokens: 8,
          total_tokens: 15,
          detail: "within smoke budget",
        },
        routing: {
          complexity: "simple",
          route: "fast",
          model_tier: "small",
          reason: "smoke route",
        },
        fallback: {
          reason: "provider_error",
          route: "local",
          model_tier: "small",
          detail: "smoke fallback available",
        },
        control: [
          { action: "replay", status: "approval_required", reason: "review smoke replay", at: now },
          { action: "pause", status: "paused", reason: "inspect smoke pause", at: now },
          { action: "resume", status: "resumed", reason: "continue smoke run", at: now },
        ],
        steps: [
          {
            id: "replay-approval",
            title: "Replay approval",
            status: "waiting_approval",
            sequence: 1,
            updated_at: now,
            metadata: { source_run_id: runID },
          },
          {
            id: "runtime-pause",
            title: "Runtime pause",
            status: "completed",
            sequence: 2,
            updated_at: now,
            metadata: { runtime_status: "resumed" },
          },
        ],
        request: { text: "webui smoke session", new_session: true, request_id: runID },
        response: {
          session_id: sessionID,
          messages: ["summary smoke response"],
          usage: { input_tokens: 7, output_tokens: 8, total_tokens: 15 },
          budget_status: {
            status: "ok",
            input_tokens: 7,
            output_tokens: 8,
            total_tokens: 15,
            detail: "within smoke budget",
          },
          routing: {
            complexity: "simple",
            route: "fast",
            model_tier: "small",
            reason: "smoke route",
          },
          fallback: {
            reason: "provider_error",
            route: "local",
            model_tier: "small",
            detail: "smoke fallback available",
          },
        },
        events: [
          { type: "preamble", at: now, data: { preamble: "planning smoke run" } },
          { type: "tool_call", at: now, data: { tool: "grep", args: JSON.stringify({ pattern: "smoke" }), status: "running" } },
          { type: "tool_result", at: now, data: { tool: "grep", content: "smoke result", status: "completed", is_error: false } },
          { type: "usage", at: now, data: { input_tokens: 7, output_tokens: 8 } },
          { type: "budget_status", at: now, data: { status: "ok", input_tokens: 7, output_tokens: 8, total_tokens: 15 } },
          { type: "routing_selected", at: now, data: { complexity: "simple", route: "fast", model_tier: "small", reason: "smoke route" } },
          { type: "fallback_decision", at: now, data: { reason: "provider_error", route: "local", model_tier: "small", detail: "smoke fallback available" } },
          { type: "text", at: now, data: { text: "summary smoke response" } },
        ],
        structured_events: [
          { schema_version: "2026-06-08", id: `${runID}-000001`, run_id: runID, type: "run_started", phase: "start", at: now, data: { channel: "http" } },
          { schema_version: "2026-06-08", id: `${runID}-000002`, run_id: runID, type: "budget_status", phase: "budget", at: now, data: { status: "ok", total_tokens: 15 } },
          { schema_version: "2026-06-08", id: `${runID}-000003`, run_id: runID, type: "control_decision", phase: "control", at: now, data: { action: "replay", status: "approval_required" } },
        ],
      }),
    });
  });
  await page.route(`**/runs/${runID}/trace`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        trace: [
          {
            schema_version: "2026-06-08",
            trace_id: runID,
            span_id: `${runID}-000001`,
            run_id: runID,
            event_id: `${runID}-000001`,
            name: "run_started",
            phase: "start",
            timestamp: now,
            attributes: { channel: "http" },
          },
          {
            schema_version: "2026-06-08",
            trace_id: runID,
            span_id: `${runID}-000002`,
            run_id: runID,
            event_id: `${runID}-000002`,
            name: "budget_status",
            phase: "budget",
            timestamp: now,
            attributes: { status: "ok", total_tokens: 15 },
          },
          {
            schema_version: "2026-06-08",
            trace_id: runID,
            span_id: `${runID}-000003`,
            run_id: runID,
            event_id: `${runID}-000003`,
            name: "control_decision",
            phase: "control",
            timestamp: now,
            attributes: { action: "replay", status: "approval_required" },
          },
        ],
      }),
    });
  });
  await page.evaluate(async (id) => window.selectRun(id), runID);
  const mockedRunDetailText = await page.locator("#run-detail").evaluate((node) => node.textContent);
  assert(mockedRunDetailText.includes(runID), "run detail missing mocked run id");
  assert(mockedRunDetailText.includes("webui smoke session"), "run detail missing prompt");
  assert(mockedRunDetailText.includes(sessionID), "run detail missing session id");
  assert(mockedRunDetailText.includes("运行恢复"), "run detail missing runtime recovery");
  assert(mockedRunDetailText.includes("工作流阶段"), "run detail missing workflow steps");
  assert(mockedRunDetailText.includes("Replay approval"), "run detail missing replay approval step");
  assert(mockedRunDetailText.includes("Runtime pause"), "run detail missing runtime pause step");
  assert(mockedRunDetailText.includes("控制记录"), "run detail missing control history");
  assert(mockedRunDetailText.includes("review smoke replay"), "run detail missing replay control");
  assert(mockedRunDetailText.includes("continue smoke run"), "run detail missing resume control");
  assert(mockedRunDetailText.includes("Trace"), "run detail missing trace section");
  assert(await page.locator("#run-detail").getByText("budget_status").count() >= 1, "run detail missing budget trace/event");
  assert(await page.locator("#run-detail").getByText("control_decision").count() >= 1, "run detail missing control trace/event");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("观测时间线")), "run detail missing observation timeline");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("运行已完成")), "run detail missing completed milestone");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("Prompt 已锁定")), "run detail missing prompt milestone");
  await page.locator("#run-detail .run-milestone").getByRole("button", { name: "打开关联会话" }).waitFor();
  assert(await page.locator("#run-detail .run-tool-event").count() === 1, "run detail should group tool call/result into one tool card");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("grep")), "run detail missing grouped tool name");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("smoke result")), "run detail missing grouped tool result");
  await page.locator("#run-detail .run-tool-event").getByRole("button", { name: "复制结果" }).click();
  await page.getByText("Tool result copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "smoke result", "tool result copy should copy grouped tool result");
  assert(await page.locator("#run-detail").evaluate((node) => node.textContent.includes("前置说明") || node.textContent.includes("preamble")), "run detail missing preamble event");
  assert(await page.locator("#run-detail").getByText("input_tokens").count() >= 1, "run detail missing usage event");
  await page.locator("#run-detail").getByText("smoke route").waitFor();
  await page.locator("#run-detail").getByText("smoke fallback available").waitFor();
  await page.locator("#run-detail").getByRole("button", { name: "复制 Prompt" }).click();
  await page.getByText("Prompt copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "webui smoke session", "copy prompt should copy run prompt");
  await page.locator("#run-detail").getByRole("button", { name: "复制结果" }).click();
  await page.getByText("Result copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "summary smoke response", "copy result should copy formatted run response");
  await page.locator("#run-detail").getByRole("button", { name: "复制摘要" }).click();
  await page.getByText("Run summary copied.").waitFor();
  const copiedRunSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedRunSummary.includes(`Run: ${runID}`), "copied run detail summary missing run id");
  assert(copiedRunSummary.includes("Prompt: [REDACTED: use Copy prompt for local operator review]"), "copied run detail summary should redact prompt");
  await page.locator("#run-detail").getByRole("button", { name: "起草后续" }).click();
  await page.locator("#panel-home.active").waitFor();
  const detailFollowUp = await homeTaskInput(page).inputValue();
  assert(detailFollowUp.includes(`Run: ${runID}`), "run detail follow-up should include run id");
  assert(detailFollowUp.includes("Original prompt: webui smoke session"), "run detail follow-up should include original prompt");
  await openPanel(page, "runs");
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "观测运行" }).click();
  await page.locator("#run-detail").getByRole("button", { name: "重新运行" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-input").inputValue() === "webui smoke session", "rerun should prefill chat prompt");
  assert(await page.locator("#chat-new-session").isChecked(), "rerun should use a new session");
  assert(await page.locator("#chat-agent").inputValue() === "", "rerun should select default agent");
  await openPanel(page, "runs");
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "观测运行" }).click();
  await page.locator("#run-detail .run-milestone").getByRole("button", { name: "打开关联会话" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "timeline open linked session should select run session");
  await openPanel(page, "runs");
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "观测运行" }).click();
  await page.locator("#run-detail").getByRole("button", { name: "打开会话" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "open session should select run session");
  await page.unroute(`**/runs/${runID}`);
  await page.unroute(`**/runs/${runID}/trace`);
  const sessionRowSelector = `#sessions-list [data-session-id="${sessionID}"]`;
  await page.locator(sessionRowSelector).waitFor({ state: "attached" });
  await page.locator(sessionRowSelector).evaluate((row) => row.querySelector("[data-session-copy]")?.click());
  await page.getByText("Session ID copied.").waitFor();
  await page.waitForFunction((selector) => {
    const row = document.querySelector(selector);
    return row?.textContent?.includes("Copied") || row?.textContent?.includes("已复制");
  }, sessionRowSelector);
  const copiedSessionID = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedSessionID === sessionID, "copied session id should match row id");
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "prompt", "rename dialog should be a prompt");
    await dialog.accept("Smoke renamed session");
  });
  await page.locator(sessionRowSelector).evaluate((row) => row.querySelector("[data-session-rename]")?.click());
  await page.waitForFunction((selector) => document.querySelector(selector)?.textContent?.includes("Smoke renamed session"), sessionRowSelector);
  await page.locator("#session-search").evaluate((input) => {
    input.value = "Smoke renamed";
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.locator(sessionRowSelector).waitFor({ state: "attached" });
  await page.locator("#session-search-clear").evaluate((button) => button.click());
  assert(await page.locator("#session-search").inputValue() === "", "session search should clear");
  await page.locator(sessionRowSelector).waitFor({ state: "attached" });
  await page.locator(sessionRowSelector).evaluate((row) => row.querySelector("[data-session-favorite]")?.click());
  await page.waitForFunction((selector) => document.querySelector(selector)?.textContent?.includes("取消收藏"), sessionRowSelector);
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "delete dialog should be a confirm");
    await dialog.dismiss();
  });
  await page.locator(sessionRowSelector).evaluate((row) => row.querySelector("[data-session-delete]")?.click());
  await page.locator(sessionRowSelector).waitFor({ state: "attached" });

  const errorRunID = "run_error_smoke";
  const errorSessionID = await page.evaluate(async ({ url, errorRunID }) => {
    const response = await fetch(`${url}/message`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text: "webui smoke provider unavailable", new_session: true, request_id: errorRunID })
    });
    const data = await response.json();
    return data.session_id;
  }, { url: baseURL, errorRunID });
  assert(errorSessionID, "error run session id missing");
  await page.locator("#refresh-button").click();
  await openPanel(page, "runs");
  await page.locator("#panel-runs").getByRole("heading", { name: "运行观测台" }).waitFor();
  await page.locator(`[data-run-id="${errorRunID}"]`).waitFor();
  await page.locator(`[data-run-id="${errorRunID}"]`).getByRole("button", { name: "观测运行" }).click();
  const errorRunDetail = await page.evaluate(async ({ url, errorRunID }) => {
    const response = await fetch(`${url}/runs/${errorRunID}`);
    return response.json();
  }, { url: baseURL, errorRunID });
  assert(errorRunDetail.status === "error", `provider-unavailable run should be error, got ${JSON.stringify(errorRunDetail)}`);
  assert(errorRunDetail.error || errorRunDetail.response?.error, `provider-unavailable run should include an error, got ${JSON.stringify(errorRunDetail)}`);
  assert(await page.locator("#run-detail").getByText(errorRunID).count() >= 1, "error run detail missing run id");
  assert(await page.locator("#run-detail").getByText("webui smoke provider unavailable").count() >= 1, "error run detail missing prompt");
  await page.locator("#run-detail").getByText("失败", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(errorSessionID).count() >= 1, "error run detail missing session id");
  const errorResultText = await page.locator("#run-detail").evaluate((detail) => {
    const section = [...detail.querySelectorAll(".run-detail-section")].find((item) => item.querySelector("h3")?.textContent?.trim() === "结果");
    return section?.textContent || "";
  });
  assert(errorResultText.includes("error"), "error run result should include error field");
  await page.locator("#run-detail").getByRole("button", { name: "复制 Prompt" }).click();
  await page.getByText("Prompt copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "webui smoke provider unavailable", "copy prompt should copy error run prompt");
  await page.locator("#run-detail").getByRole("button", { name: "重新运行" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-input").inputValue() === "webui smoke provider unavailable", "error rerun should prefill chat prompt");
  assert(await page.locator("#chat-new-session").isChecked(), "error rerun should use a new session");
  assert(await page.locator("#chat-agent").inputValue() === "", "error rerun should select default agent");
}

async function runStreamingProvider(page) {
  const homePrompt = "webui home composer smoke";
  await openHome(page);
  await closeHomeDisclosures(page);
  await page.locator("#home-task-input").fill(homePrompt);
  await page.locator(".send-orbit-button").click();
  await page.locator("#panel-chat.active").waitFor();
  await page.locator("#chat-output").getByText(homePrompt).waitFor();
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();
  await page.locator("#new-chat-button").click();

  const prompt = "webui streaming provider smoke";
  await openPanel(page, "chat");
  await page.locator("#chat-agent").selectOption("");
  await page.locator("#chat-new-session").check();
  await page.locator("#chat-input").fill(prompt);
  await page.locator("#send-button").click();

  await page.locator("#stop-button:not([hidden])").waitFor();
  await page.locator("#live-run-status:not([hidden])").waitFor();
  await page.locator("#live-run-state", { hasText: /running|运行中/ }).waitFor();
  await page.waitForFunction(() => document.querySelector("[data-star-map]")?.dataset.starState === "running");
  await page.locator("#live-run-id", { hasText: /[0-9a-f-]{8,}|web-/ }).waitFor();
  await page.locator("#chat-output").getByText("Fake provider").waitFor();
  await page.locator("#live-run-event", { hasText: /Streaming text|Usage updated|Session started/ }).waitFor();
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();
  const summary = page.locator("#chat-output .run-summary").last();
  await summary.getByText("运行摘要").waitFor();
  await summary.getByText("input_tokens: 11").waitFor();
  await summary.getByText("output_tokens: 7").waitFor();
  await page.locator("#live-run-state", { hasText: /complete|已完成/ }).waitFor();
  await page.waitForFunction(() => document.querySelector("[data-star-map]")?.dataset.starState === "complete");
  await page.locator("#live-run-usage", { hasText: /input|in 11/ }).waitFor();
  const runID = await summary.getByRole("button", { name: "观测运行" }).getAttribute("data-run-summary-run");
  const sessionID = await summary.getByRole("button", { name: "打开会话" }).getAttribute("data-run-summary-session");
  assert(runID, "streaming run summary missing request id");
  assert(sessionID, "streaming run summary missing session id");
  await page.locator("#live-session-id", { hasText: sessionID }).waitFor();
  await page.waitForFunction(async ({ url, sessionID }) => {
    const response = await fetch(`${url}/sessions/${sessionID}`);
    if (!response.ok) return false;
    const data = await response.json();
    return JSON.stringify(data).includes("Fake provider streamed response for GUI smoke.");
  }, { url: baseURL, sessionID });

  await page.locator("#refresh-button").click();
  await openSession(page, sessionID);
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "streaming open session should select persisted session");
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();

  await openPanel(page, "runs");
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "观测运行" }).click();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), runID);
  await page.locator("#run-detail").getByText("完成", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(prompt).count() >= 1, "streaming run detail missing prompt");
  assert(await page.locator("#run-detail").getByText("Fake provider streamed response for GUI smoke.").count() >= 1, "streaming run detail missing response text");
  assert(await page.locator("#run-detail").getByText(sessionID).count() >= 1, "streaming run detail missing session id");
  assert(await page.locator("#run-detail").getByText("input_tokens").count() >= 1, "streaming run detail missing usage");
  await page.locator("#run-detail").getByRole("button", { name: "复制摘要" }).click();
  await page.getByText("Run summary copied.").waitFor();
  const copiedRunSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedRunSummary.includes(`Run: ${runID}`), "streaming copied run summary missing run id");
  assert(copiedRunSummary.includes("Usage: input_tokens: 11, output_tokens: 7, total_tokens: 18"), "streaming copied run summary missing usage");
}

async function runToolCallProvider(page) {
  const prompt = "webui tool call smoke";
  await openPanel(page, "chat");
  await page.locator("#chat-agent").selectOption("");
  await page.locator("#chat-new-session").check();
  await page.locator("#chat-input").fill(prompt);
  await page.locator("#send-button").click();

  const toolEvent = page.locator("#chat-output .tool-event").filter({ hasText: "version" }).first();
  await page.waitForFunction(() => ["tool", "artifact", "complete"].includes(document.querySelector("[data-star-map]")?.dataset.starState || ""));
  await toolEvent.getByText(/completed|完成/).waitFor();
  await toolEvent.locator("summary").click();
  await toolEvent.getByText("StarClaw").waitFor();
  await page.locator("#chat-output").getByText("Version tool call completed for GUI smoke.").waitFor();

  const summary = page.locator("#chat-output .run-summary").last();
  await summary.getByText("运行摘要").waitFor();
  await page.waitForFunction(() => document.querySelector("[data-star-map]")?.dataset.starState === "complete");
  const runID = await summary.getByRole("button", { name: "观测运行" }).getAttribute("data-run-summary-run");
  const sessionID = await summary.getByRole("button", { name: "打开会话" }).getAttribute("data-run-summary-session");
  assert(runID, "tool-call run summary missing request id");
  assert(sessionID, "tool-call run summary missing session id");

  await page.locator("#refresh-button").click();
  await openSession(page, sessionID);
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "tool-call open session should select persisted session");
  await page.locator("#chat-output").getByText("Version tool call completed for GUI smoke.").waitFor();

  await openPanel(page, "runs");
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "观测运行" }).click();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), runID);
  assert(await page.locator("#run-detail").getByText(runID).count() >= 1, "tool-call run detail missing run id");
  await page.locator("#run-detail").getByText("完成", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(prompt).count() >= 1, "tool-call run detail missing prompt");
  assert(await page.locator("#run-detail .run-tool-event").count() === 1, "tool-call run detail should group tool call/result into one tool card");
  const runToolEvent = page.locator("#run-detail .run-tool-event").first();
  await runToolEvent.getByText("version").waitFor();
  await runToolEvent.getByText("完成").waitFor();
  await runToolEvent.getByText("content_redacted").waitFor();
  await runToolEvent.getByRole("button", { name: "复制结果" }).click();
  await page.getByText("Tool result copied.").waitFor();
  assert((await page.evaluate(() => navigator.clipboard.readText())).includes("content_redacted"), "tool-call run detail tool copy should keep redacted result marker");
  assert(await page.locator("#run-detail").getByText("Version tool call completed for GUI smoke.").count() >= 1, "tool-call run detail missing final response");
}

async function setFakeProviderScenario(scenario) {
  const response = await fetch(`${process.env.WEBUI_FAKE_PROVIDER_URL || "http://127.0.0.1:17534"}/scenario`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ scenario }),
  });
  assert(response.ok, `failed to switch fake provider scenario to ${scenario}`);
}

async function saveOpenAIConfig(page, options = {}) {
  const endpoint = options.endpoint || (process.env.WEBUI_FAKE_PROVIDER_URL || "http://127.0.0.1:17534");
  const model = options.model || "fake-streaming-model";
  const apiKey = options.apiKey || "fake-key";
  await openPanel(page, "config");
  await page.locator("#panel-config.active").waitFor();
  await page.locator("#config-provider").selectOption("openai");
  await page.locator("#config-openai-endpoint").fill(endpoint);
  await page.locator("#config-openai-model").fill(model);
  await page.locator("#config-openai-api-key").fill(apiKey);
  await page.locator("#config-form").getByRole("button", { name: "保存 provider 设置" }).click();
  await page.getByText("provider 设置已保存。").waitFor();
}

async function checkConnectionMessage(page, scenario, expected) {
  await setFakeProviderScenario(scenario);
  const responsePromise = page.waitForResponse((response) => response.url().endsWith("/config/test") && response.request().method() === "POST");
  await page.locator("#config-test-button").click();
  const response = await responsePromise;
  const payload = await response.json();
  assert(payload.code === expected.code, `config test code for ${scenario} = ${JSON.stringify(payload)}, want ${expected.code}`);
  await page.locator("#connector-test-title", { hasText: expected.title }).waitFor();
  await page.locator("#connector-test-detail").getByText(expected.text).waitFor();
  const detail = await page.locator("#connector-test-detail").textContent();
  assert(!String(detail || "").includes("fake-key"), `connector detail leaked API key for ${scenario}: ${detail}`);
}

async function runConfigConnection(page) {
  await openHome(page, { calm: true });
  await homeTaskInput(page).fill("webui config guard smoke");
  await page.locator(".send-orbit-button").click();
  await page.locator("#panel-config.active").waitFor();
  await page.locator("[data-connector-status-title]", { hasText: "等待用户填写连接" }).waitFor();
  await page.locator("[data-connector-status-detail]").getByText("Base URL").waitFor();
  await page.locator("[data-connector-status-detail]").getByText("Model").waitFor();
  await page.locator("[data-connector-status-detail]").getByText("API key").waitFor();
  assert(await homeTaskInput(page).inputValue() === "webui config guard smoke", "blocked home task should remain in composer");

  await saveOpenAIConfig(page);
  await page.locator("[data-connector-status-title]", { hasText: "openai 已准备" }).waitFor();
  await checkConnectionMessage(page, "streaming", {
    code: "ok",
    title: "连接可用",
    text: "连接成功",
  });

  await openHome(page, { calm: true });
  await homeTaskInput(page).fill("webui config successful launch");
  await page.locator(".send-orbit-button").click();
  await page.locator("#panel-chat.active").waitFor();
  await page.locator("#chat-output").getByText("webui config successful launch").waitFor();
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();
  await page.locator("#chat-output .run-summary").last().getByText("运行摘要").waitFor();

  await openPanel(page, "config");
  await page.locator("#panel-config.active").waitFor();
  await checkConnectionMessage(page, "config_auth", {
    code: "auth_failed",
    title: "连接失败",
    text: "API key 无效或权限不足",
  });
  await checkConnectionMessage(page, "config_model", {
    code: "model_not_found",
    title: "连接失败",
    text: "模型不可用",
  });
  await checkConnectionMessage(page, "config_rate", {
    code: "rate_limited",
    title: "连接失败",
    text: "provider 返回限流",
  });
  await checkConnectionMessage(page, "config_invalid", {
    code: "invalid_response",
    title: "连接失败",
    text: "响应格式不兼容",
  });
  await setFakeProviderScenario("streaming");
}

async function runSelected(page) {
  switch (mode) {
    case "core":
      await runCore(page);
      return;
    case "permissions":
      await runPermissions(page);
      return;
    case "agents":
      await runAgents(page);
      return;
    case "runs":
      await runRuns(page);
      return;
    case "streaming":
      await runStreamingProvider(page);
      return;
    case "tool_call":
      await runToolCallProvider(page);
      return;
    case "config":
      await runConfigConnection(page);
      return;
    case "full":
      await runCore(page);
      await runPermissions(page);
      await runAgents(page);
      await runRuns(page);
      return;
    default:
      throw new Error(`unknown smoke mode: ${mode}`);
  }
}

async function runFullIsolated(context) {
  let lastPage = null;
  for (const runLayer of [runCore, runPermissions, runAgents, runRuns]) {
    const page = await context.newPage();
    await boot(page);
    await runLayer(page);
    if (lastPage) await lastPage.close();
    lastPage = page;
  }
  return lastPage;
}

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ viewport: { width: 1440, height: 960 }, acceptDownloads: true });
await context.grantPermissions(["clipboard-read", "clipboard-write"], { origin: baseURL });
let page = null;

try {
  if (mode === "full") {
    page = await runFullIsolated(context);
  } else {
    page = await context.newPage();
    await boot(page);
    await runSelected(page);
  }
  await page.screenshot({ path: screenshot, fullPage: true });
  assert(fs.existsSync(screenshot), "screenshot was not written");
} catch (error) {
  if (page && screenshot) {
    await page.screenshot({ path: screenshot, fullPage: true }).catch(() => {});
  }
  throw error;
} finally {
  await browser.close();
}
JS
}

check_routes() {
  echo "==> checking daemon routes"
  curl -fsS "$BASE_URL/status" >/dev/null
  curl -fsS "$BASE_URL/version" | grep -F '"update_command":"starclaw update --check"' >/dev/null || fail "version JSON missing update command"
  curl -fsS "$BASE_URL/version" | grep -F '"launch_command":"starclaw app"' >/dev/null || fail "version JSON missing launch command"
  curl -fsS "$BASE_URL/update/check" | grep -F '"status":"development"' >/dev/null || fail "update check JSON missing development status"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"checks"' >/dev/null || fail "diagnostics JSON missing checks"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"launch_command":"starclaw app"' >/dev/null || fail "diagnostics JSON missing launch command"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"web_url":"'"$BASE_URL"'/app/"' >/dev/null || fail "diagnostics JSON missing web URL"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"health_url":"'"$BASE_URL"'/health"' >/dev/null || fail "diagnostics JSON missing health URL"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"status_url":"'"$BASE_URL"'/status"' >/dev/null || fail "diagnostics JSON missing status URL"
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"diagnostics_url":"'"$BASE_URL"'/diagnostics"' >/dev/null || fail "diagnostics JSON missing diagnostics URL"
  curl -fsS "$BASE_URL/diagnostics" | grep -E '"config_path":"[^"]*config\.yaml"' >/dev/null || fail "diagnostics JSON missing config path"
  curl -fsS "$BASE_URL/permissions" | grep -F '"configured":true' >/dev/null || fail "permissions JSON missing configured policy"
  curl -fsSI "$BASE_URL/" | grep -F "Location: /app/" >/dev/null || fail "root redirect missing"
  curl -fsSI "$BASE_URL/app" | grep -F "Location: /app/" >/dev/null || fail "app redirect missing"
  curl -fsS "$BASE_URL/app/" | grep -F "Astria" >/dev/null || fail "app HTML missing Astria"
  curl -fsS "$BASE_URL/app/assets/app.js" | grep -F "connectEventStream" >/dev/null || fail "app JS missing event stream code"
  curl -fsS "$BASE_URL/app/assets/styles.css" | grep -F "approval-card" >/dev/null || fail "CSS missing approval styles"
}

require_cmd curl
require_cmd npx

echo "==> building StarClaw"
(cd "$ROOT_DIR" && go build -o "$BIN" ./main.go)

write_smoke_config
write_intake_fixture
write_node_package

echo "==> installing browser smoke dependency"
(cd "$NODE_DIR" && npm install --silent)
if [[ "${CI:-}" == "true" ]]; then
  (cd "$NODE_DIR" && npx playwright install chromium --with-deps >/dev/null)
else
  (cd "$NODE_DIR" && npx playwright install chromium >/dev/null)
fi

start_fake_provider_if_needed

echo "==> starting daemon"
env HOME="$SMOKE_HOME" STARCLAW_DAEMON_PORT="$DAEMON_PORT" "$BIN" daemon start >"$DAEMON_LOG" 2>&1 &
DAEMON_PID="$!"
wait_for_health
check_routes
write_browser_smoke

echo "==> running browser smoke ($SMOKE_MODE)"
env BASE_URL="$BASE_URL" WEBUI_FAKE_PROVIDER_URL="$FAKE_PROVIDER_URL" SCREENSHOT="$SCREENSHOT" HOME_SCREENSHOT="$HOME_SCREENSHOT" INTAKE_DOC="$INTAKE_DOC" NODE_DIR="$NODE_DIR" WEBUI_SMOKE_MODE="$SMOKE_MODE" SMOKE_HOME="$SMOKE_HOME" node "$NODE_SCRIPT"

echo "smoke_webui_${SMOKE_MODE}: ok"
echo "screenshot: $SCREENSHOT"
echo "home screenshot: $HOME_SCREENSHOT"
echo "daemon log: $DAEMON_LOG_ARTIFACT"
echo "metadata: $METADATA_ARTIFACT"
