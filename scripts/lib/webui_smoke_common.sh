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
BASE_URL="${WEBUI_SMOKE_BASE_URL:-http://127.0.0.1:7533}"
FAKE_PROVIDER_URL="${WEBUI_FAKE_PROVIDER_URL:-http://127.0.0.1:17534}"
ARTIFACT_DIR="${WEBUI_SMOKE_ARTIFACT_DIR:-$ROOT_DIR/output/playwright}"
SCREENSHOT_DIR="$ARTIFACT_DIR"
SCREENSHOT="$SCREENSHOT_DIR/daemon-webui-${SMOKE_MODE}-smoke.png"
HOME_SCREENSHOT="$SCREENSHOT_DIR/astria-home-${SMOKE_MODE}-smoke.png"
INTAKE_DOC="$SCREENSHOT_DIR/intake-smoke.docx"
DAEMON_LOG_ARTIFACT="$ARTIFACT_DIR/daemon-webui-${SMOKE_MODE}-smoke.log"
METADATA_ARTIFACT="$ARTIFACT_DIR/daemon-webui-${SMOKE_MODE}-smoke.metadata"
DAEMON_PID=""
FAKE_PROVIDER_PID=""

persist_artifacts() {
  mkdir -p "$ARTIFACT_DIR"
  if [[ -f "$DAEMON_LOG" ]]; then
    cp "$DAEMON_LOG" "$DAEMON_LOG_ARTIFACT"
  fi
  {
    printf 'mode=%s\n' "$SMOKE_MODE"
    printf 'base_url=%s\n' "$BASE_URL"
    printf 'screenshot=%s\n' "$SCREENSHOT"
    printf 'home_screenshot=%s\n' "$HOME_SCREENSHOT"
    printf 'daemon_log=%s\n' "$DAEMON_LOG_ARTIFACT"
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
const scenario = process.env.FAKE_PROVIDER_SCENARIO || "streaming";

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
  if (req.method !== "POST" || !req.url.endsWith("/chat/completions")) {
    writeJSON(res, 404, { error: { message: `unexpected route: ${req.method} ${req.url}` } });
    return;
  }

  const rawBody = await readBody(req);
  const request = JSON.parse(rawBody || "{}");
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
  if [[ "$SMOKE_MODE" != "streaming" && "$SMOKE_MODE" != "tool_call" ]]; then
    return
  fi
  write_fake_provider
  echo "==> starting fake OpenAI provider"
  FAKE_PROVIDER_PORT="${FAKE_PROVIDER_URL##*:}" FAKE_PROVIDER_SCENARIO="$SMOKE_MODE" node "$FAKE_PROVIDER_SCRIPT" >"$TMP_DIR/fake-provider.log" 2>&1 &
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

async function boot(page) {
  await page.goto(`${baseURL}/app/`, { waitUntil: "domcontentloaded" });
  await page.getByRole("heading", { name: "Hiya, welcome to Astria" }).waitFor();
  await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").waitFor();
  await page.getByRole("button", { name: "Open Command Center" }).click();
  await page.getByLabel("Command search").fill("memory");
  await page.locator("#command-center-list").getByRole("button", { name: /Memory Map/ }).click();
  await page.locator("#panel-memory.active").waitFor();
  await page.keyboard.press(process.platform === "darwin" ? "Meta+K" : "Control+K");
  await page.getByLabel("Command search").fill("review");
  await page.locator("#command-center-list").getByRole("button", { name: /代码评审/ }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#strategy-matrix").getByRole("button", { name: /Research Brief/ }).click();
  const strategyPrompt = await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").inputValue();
  assert(strategyPrompt.includes("Prepare a research brief"), "research strategy should prefill home prompt");
  await page.locator("#strategy-brief .strategy-brief-head").getByText("Research Brief", { exact: true }).waitFor();
  await page.locator("#focus-brief").getByText("Strategy", { exact: true }).waitFor();
  await page.locator("#strategy-brief").getByRole("button", { name: "打开运行" }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#prompt-suggestion-dock").getByRole("button", { name: /Research Brief next prompt|Default/ }).click();
  const suggestionPrompt = await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").inputValue();
  assert(suggestionPrompt.includes("Prepare a research brief"), "prompt suggestion should seed home prompt");
  await page.locator("#workflow-recipes").getByRole("button", { name: /代码评审/ }).click();
  const reviewPrompt = await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").inputValue();
  assert(reviewPrompt.includes("Review the current working tree"), "code review recipe should prefill home prompt");
  await page.locator("#workflow-stage-rail").getByRole("button", { name: /Draft.*代码评审/ }).waitFor();
  await page.locator("#workflow-stage-rail").getByRole("button", { name: /Running.*Daemon execution/ }).waitFor();
  await page.locator("#focus-brief").getByText("代码评审").waitFor();
  await page.locator("#focus-brief .board-kicker").getByText("draft", { exact: true }).waitFor();
  await page.locator("#workspace-health-strip").getByRole("button", { name: /Diagnostics/ }).waitFor();
  await page.locator("#workspace-health-strip").getByRole("button", { name: /Permissions/ }).waitFor();
  await page.locator("#workspace-health-strip").getByRole("button", { name: /MCP/ }).waitFor();
  await page.locator("#workspace-health-strip").getByRole("button", { name: /Memory/ }).click();
  await page.locator("#panel-memory.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#review-queue-list").getByRole("button", { name: /Permissions/ }).click();
  await page.locator("#panel-permissions.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#approval-center-grid").getByRole("button", { name: /Permissions/ }).click();
  await page.locator("#panel-permissions.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#knowledge-curation-grid").getByRole("button", { name: /Low context/ }).click();
  await page.locator("#panel-memory.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#tool-dock-inspector-grid").getByRole("button", { name: /No tool docks/ }).click();
  await page.locator("#panel-mcp.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#workflow-brief").getByText("一份按严重程度排序的评审报告").waitFor();
  await page.locator("#workflow-brief").getByText("当前 git diff").waitFor();
  await page.locator("#workflow-recipes").getByRole("button", { name: /文件理解/ }).click();
  await page.locator("#workflow-brief").getByText("把本地文件内容整理成可引用上下文").waitFor();
  await page.locator("#workflow-brief").getByRole("button", { name: "打开文件星舱" }).waitFor();
  await page.locator("#home-mode-route", { hasText: "打开文件星舱" }).waitFor();
  await page.locator("#workflow-brief").getByRole("button", { name: "打开文件星舱" }).click();
  await page.locator("#panel-intake.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#workspace-session-hub").getByRole("button", { name: /Session/ }).waitFor();
  await page.locator("#workspace-session-hub").getByRole("button", { name: /Runs/ }).waitFor();
  await page.locator("#workspace-session-hub").getByRole("button", { name: /Memory/ }).waitFor();
  await page.locator("#workspace-session-hub").getByRole("button", { name: /Files/ }).click();
  await page.locator("#panel-intake.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  if (homeScreenshot) {
    await page.screenshot({ path: homeScreenshot, fullPage: true });
    assert(fs.existsSync(homeScreenshot), "home screenshot was not written");
  }
  await page.getByRole("button", { name: "Chat", exact: true }).click();
  await page.locator("#panel-chat").waitFor();
  await page.getByPlaceholder("Message Astria").waitFor();
  await page.getByRole("button", { name: "Send" }).waitFor();
  assert(await page.locator(".sidebar").count() === 1, "sidebar missing");
  await page.locator("#diagnostics-chip").waitFor();
}

async function openManagePanel(page, name) {
  await page.getByRole("button", { name: "Manage" }).click();
  await page.locator("#panel-manage").getByRole("heading", { name: "Manage" }).waitFor();
  await page.locator("#panel-manage").getByRole("button", { name: new RegExp(`^${name}`) }).click();
}

async function openSettingsPanel(page, name) {
  await page.getByRole("button", { name: "Settings" }).click();
  await page.locator("#panel-settings").getByRole("heading", { name: "Settings" }).waitFor();
  await page.locator("#panel-settings").getByRole("button", { name: new RegExp(`^${name}`) }).click();
}

async function runCore(page) {
  await page.locator("#diagnostics-chip").click();
  await page.locator("#panel-diagnostics").getByRole("heading", { name: "Diagnostics" }).waitFor();
  await page.locator("#panel-diagnostics").getByText("Launch readiness").waitFor();
  await page.locator("#panel-diagnostics").getByText("starclaw app").waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/app/`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/health`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/status`).waitFor();
  await page.locator("#panel-diagnostics").getByText(`${baseURL}/diagnostics`).waitFor();
  await page.locator("#panel-diagnostics").getByText("Config", { exact: true }).waitFor();
  await page.locator("#panel-diagnostics").getByText("Agents", { exact: true }).waitFor();
  await page.locator("#panel-diagnostics").getByText("Sessions", { exact: true }).waitFor();
  await page.getByText("Provider", { exact: true }).waitFor();
  await page.getByText(/Ollama is configured/).waitFor();
  await page.getByRole("button", { name: "Fix provider setup" }).click();
  await page.locator("#panel-config").getByRole("heading", { name: "Config" }).waitFor();
  await page.getByLabel("Provider").selectOption("ollama");
  await page.getByLabel("Ollama endpoint").fill("http://127.0.0.1:1");
  await page.getByLabel("Ollama model").fill("smoke-gui-model");
  await page.getByRole("button", { name: "Save provider config" }).click();
  await page.getByText("Provider config saved.").waitFor();
  await openSettingsPanel(page, "Version");
  await page.locator("#panel-version").getByRole("heading", { name: "Version" }).waitFor();
  await page.locator("#version-list").getByText("Release readiness").waitFor();
  await page.locator("#version-list").getByText("Development build").waitFor();
  await page.locator("#version-list").getByText("Use a semver release build to enable update checks.").waitFor();
  await page.locator("#version-list").getByText("Version").waitFor();
  await page.locator("#version-list").getByText("Platform").waitFor();
  await page.locator("#version-list").getByText("Web UI", { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText("Launch", { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText("starclaw app", { exact: true }).first().waitFor();
  await page.locator("#version-list").getByText("starclaw update --check").waitFor();
  await page.locator("#version-list").getByText("Runtime context").waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/health`).waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/status`).waitFor();
  await page.locator("#version-list").getByText(`${baseURL}/diagnostics`).waitFor();
  await page.locator("#version-list").getByText(`${process.env.SMOKE_HOME}/.starclaw/config.yaml`).waitFor();
  await page.locator("#update-overview").getByText("Development build", { exact: true }).waitFor();
  assert(await page.getByRole("button", { name: "Check updates" }).isDisabled(), "development build should disable update check button");
  await page.locator("#update-check-state").getByText("Unavailable").waitFor();
  await page.getByRole("button", { name: "Copy support info" }).click();
  await page.getByText("Support info copied.").waitFor();
  const supportInfo = await page.evaluate(() => navigator.clipboard.readText());
  assert(supportInfo.includes("Astria support info"), "support info missing heading");
  assert(supportInfo.includes("Version: dev"), "support info missing version");
  assert(supportInfo.includes(`Web UI: ${baseURL}/app/`), "support info missing web URL");
  assert(supportInfo.includes(`Diagnostics URL: ${baseURL}/diagnostics`), "support info missing diagnostics URL");
  assert(supportInfo.includes(`Data dir: ${process.env.SMOKE_HOME}/.starclaw`), "support info missing data dir");
  assert(supportInfo.includes("Diagnostics status:"), "support info missing diagnostics status");
  assert(!supportInfo.toLowerCase().includes("api_key"), "support info should not include API key fields");
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#home-tool-orbit").getByRole("button", { name: /MCP/ }).click();
  await page.locator("#panel-mcp.active").waitFor();
  await page.getByRole("button", { name: "Add dock" }).click();
  await page.getByLabel("MCP server name").fill("smoke");
  await page.getByLabel("MCP command").fill("node");
  await page.getByLabel("MCP args").fill("smoke-mcp.js");
  await page.getByLabel("MCP env").fill("SMOKE_TOKEN=fake-secret");
  await page.locator("#mcp-form").getByRole("button", { name: "Save dock" }).click();
  await page.getByText("MCP dock saved.").waitFor();
  await page.locator("#mcp-list").getByText("smoke", { exact: true }).waitFor();
  await page.locator("#mcp-list").getByRole("button", { name: "Edit" }).click();
  await page.getByLabel("MCP args").fill("smoke-mcp.js, --edited");
  await page.locator("#mcp-form").getByRole("button", { name: "Save dock" }).click();
  await page.getByText("MCP dock saved.").waitFor();
  await page.locator("#mcp-list").getByRole("button", { name: "Disable" }).click();
  await page.getByText("MCP dock disabled.").waitFor();
  const mcpConfig = await page.evaluate(async () => {
    const response = await fetch("/config");
    return response.json();
  });
  const smokeDock = mcpConfig.config.mcp_servers.find((server) => server.name === "smoke");
  assert(smokeDock, "smoke MCP dock should be saved");
  assert(smokeDock.disabled === true, "smoke MCP dock should be disabled");
  assert(smokeDock.args.includes("--edited"), `smoke MCP args should include edit, got ${JSON.stringify(smokeDock.args)}`);
  assert(smokeDock.env_keys.includes("SMOKE_TOKEN"), `smoke MCP env key should be redacted, got ${JSON.stringify(smokeDock.env_keys)}`);
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#home-tool-orbit").getByRole("button", { name: /文件/ }).click();
  await page.locator("#panel-intake.active").waitFor();
  await page.getByLabel("File intake path").fill(intakeDoc);
  await page.getByLabel("File intake mode").selectOption("document_text");
  const intakeResponsePromise = page.waitForResponse((response) => response.url().endsWith("/intake/file") && response.request().method() === "POST");
  await page.locator("#intake-form").getByRole("button", { name: "Analyze file" }).click();
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
  await page.getByRole("button", { name: "Send to Chat" }).click();
  await page.locator("#panel-chat.active").waitFor();
  await page.getByPlaceholder("Message Astria").inputValue().then((value) => {
    assert(value.includes("Astria smoke intake document"), "chat prompt should include intake result");
  });
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#home-tool-orbit").getByRole("button", { name: /收件箱/ }).click();
  await page.locator("#panel-inbox.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator(".constellation-card").filter({ hasText: "记忆星图" }).click();
  await page.locator("#panel-memory.active").waitFor();
  await page.locator("#memory-taxonomy-bar").getByRole("button", { name: /Decisions/ }).waitFor();
  await page.getByLabel("Memory candidate").fill("- [decision] Keep Astria UI calm and native.");
  await page.locator("#memory-candidate-preview").getByText("Decisions").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.getByRole("button", { name: "MCP", exact: true }).click();
  await page.locator("#panel-mcp.active").waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.getByRole("button", { name: "多智能体" }).click();
  await page.locator("#panel-council.active").waitFor();
  await page.getByLabel("Council goal").fill("webui council smoke");
  await page.getByRole("button", { name: "Start council" }).click();
  await page.locator("#council-detail").getByRole("heading", { name: "webui council smoke" }).waitFor();
  await page.locator("#council-detail").getByText("planner", { exact: true }).waitFor();
  await page.locator("#council-detail").getByText("researcher", { exact: true }).waitFor();
  await page.locator("#council-detail").getByText("reviewer", { exact: true }).waitFor();
  await page.locator("#council-detail").getByText("Final synthesis").waitFor();
  await page.locator("#council-detail").getByRole("button", { name: "Copy synthesis" }).click();
  await page.getByText("Council synthesis copied.").waitFor();
  const councilSynthesis = await page.evaluate(() => navigator.clipboard.readText());
  assert(councilSynthesis.includes("Council synthesis for: webui council smoke"), "council synthesis copy missing goal");
  await page.locator("#council-detail").getByRole("button", { name: "Start run" }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.locator("#run-detail .run-meta-grid").getByText("council_handoff", { exact: true }).waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await openManagePanel(page, "Inbox");
  await page.locator("#panel-inbox").getByRole("heading", { name: "Inbox" }).waitFor();
  await page.locator("#inbox-provider-list").getByText("GitHub", { exact: true }).waitFor();
  await page.locator("#inbox-provider-list").getByText("/inbox/github").waitFor();
  await page.locator("#inbox-provider-list").getByText("issues").waitFor();
  await page.getByLabel("Inbox external id").fill("evt-smoke-core");
  await page.getByLabel("Inbox sender").fill("webui-smoke");
  await page.getByLabel("Inbox text").fill("Review this inbound smoke task.");
  await page.getByRole("button", { name: "Receive webhook" }).click();
  await page.getByText("Inbound task received.").waitFor();
  await page.locator("#inbox-list").getByText("Review this inbound smoke task.").waitFor();
  await page.locator("#inbox-list").getByText("pending").waitFor();
  await page.locator("#inbox-list").getByRole("button", { name: "Reject" }).click();
  await page.getByText("Inbox item rejected.").waitFor();
  await page.locator("#inbox-list .tag").getByText("rejected", { exact: true }).waitFor();
  await openManagePanel(page, "Schedules");
  await page.getByLabel("Cron expression").fill("* * * * *");
  await page.getByLabel("Schedule prompt").fill("webui smoke schedule");
  await page.getByRole("button", { name: "Create schedule" }).click();
  await page.getByText("webui smoke schedule").waitFor();
  await page.getByRole("button", { name: "Pause" }).click();
  await page.getByRole("button", { name: "Enable" }).waitFor();
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByText("No schedules configured.").waitFor();
  await page.getByRole("button", { name: "Chat" }).click();
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
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#home-count-pending").getByText("0", { exact: true }).waitFor();
  await page.getByRole("button", { name: "Chat" }).click();
  const eventStatus = await page.evaluate(async (url) => {
    const response = await fetch(`${url}/events`);
    await response.body.cancel();
    return response.status;
  }, baseURL);
  assert(eventStatus === 200, `events status = ${eventStatus}`);
}

async function runPermissions(page) {
  await openSettingsPanel(page, "Permissions");
  await page.locator("#panel-permissions").getByRole("heading", { name: "Permissions" }).waitFor();
  await page.locator("#permissions-form").getByText("Allowed directories").waitFor();
  await page.locator("#permissions-form").getByText("Network allowlist").waitFor();
  await page.getByLabel("Allowed directories").fill("~\n.\n/tmp/smoke");
  await page.getByLabel("Allowed commands").fill("go test\nstarclaw version");
  await page.getByLabel("Denied commands").fill("shutdown\nreboot");
  await page.getByLabel("Network allowlist").fill("api.github.com\nsmoke.example.com");
  await page.getByLabel("Sensitive patterns").fill("*.secret\n.env.smoke");
  await page.locator("#permissions-pending-preview").getByText("Pending changes").waitFor();
  await page.locator("#permissions-pending-preview").getByText("Broad local access is allowed.").waitFor();
  await page.locator("#permissions-pending-preview").getByText("Allowed directories").waitFor();
  await page.getByRole("button", { name: "Save permissions" }).click();
  await page.getByText("Permissions saved.").waitFor();
  await page.locator("#permissions-list").getByText("/tmp/smoke").waitFor();
  await page.locator("#permissions-list").getByText("starclaw version").waitFor();
  await page.locator("#permissions-list").getByText("smoke.example.com").waitFor();
  assert((await page.getByLabel("Allowed directories").inputValue()).includes("/tmp/smoke"), "permissions editor should retain saved allowed dirs");
  await page.getByRole("button", { name: "Clear rules" }).click();
  await page.getByText("Permissions saved.").waitFor();
  await page.locator("#permissions-overview").getByText("Built-in defaults").waitFor();
  await page.locator("#permissions-pending-preview").getByText("No denied commands are configured.").waitFor();
  await page.locator("#permissions-pending-preview").getByText("No sensitive file patterns are configured.").waitFor();
  assert(await page.getByLabel("Allowed directories").inputValue() === "", "clear rules should empty allowed dirs");
  assert(await page.getByLabel("Allowed commands").inputValue() === "", "clear rules should empty allowed commands");
}

async function runAgents(page) {
  await openManagePanel(page, "Agents");
  await page.locator("#panel-agents").getByRole("heading", { name: "Agents" }).waitFor();
  await page.getByRole("button", { name: "New agent" }).click();
  await page.getByLabel("Agent name").fill("smoke-agent");
  await page.getByLabel("Agent prompt").fill("You are a smoke test agent.");
  await page.getByLabel("Agent memory").fill("Remember smoke.");
  await page.getByLabel("Agent model").fill("smoke-model");
  await page.getByLabel("Agent reasoning effort").fill("low");
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
  await page.locator("#agent-permission-preview").getByText("Enabled", { exact: true }).waitFor();
  await page.locator("#agent-permission-preview").getByText("Auto approve is enabled for this agent.").waitFor();
  await agentToolsDeny.fill("bash\ngrep");
  await page.locator("#agent-permission-preview").getByText("Allow/deny conflict: grep").waitFor();
  await agentToolsDeny.fill("bash");
  await agentHeartbeatEvery.fill("15m");
  await agentHeartbeatActiveHours.fill("09:00-17:00");
  await agentHeartbeatModel.fill("smoke-heartbeat-model");
  await agentCommandName.fill("review");
  await agentCommandBody.fill("Review recent smoke changes.");
  await saveCommandButton.click();
  await page.locator("#agent-command-list").getByText("review").waitFor();
  await saveAgentButton.click();
  await page.getByText("Agent saved.").waitFor();
  await page.locator("#agents-list").getByText("smoke-agent").waitFor();
  const rosterCard = page.locator("#agent-capability-roster .agent-roster-card").filter({ hasText: "smoke-agent" });
  await rosterCard.getByText("smoke-model", { exact: true }).waitFor();
  await rosterCard.getByText("low", { exact: true }).waitFor();
  await rosterCard.getByText("Auto approve", { exact: true }).waitFor();
  await rosterCard.getByText("Memory", { exact: true }).waitFor();
  await rosterCard.getByText("15m · 09:00-17:00").waitFor();
  await rosterCard.getByText("Heartbeat scheduled", { exact: true }).waitFor();
  await rosterCard.getByText("Approval bypass", { exact: true }).waitFor();
  const createdRosterMetrics = await rosterCard.locator(".agent-roster-metrics").innerText();
  assert(createdRosterMetrics.includes("ALLOW\n2"), `agent roster should show two allowed tools, got ${JSON.stringify(createdRosterMetrics)}`);
  assert(createdRosterMetrics.includes("DENY\n1"), `agent roster should show one denied tool, got ${JSON.stringify(createdRosterMetrics)}`);
  assert(createdRosterMetrics.includes("COMMANDS\n1"), `agent roster should show one command, got ${JSON.stringify(createdRosterMetrics)}`);
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
  await rosterCard.getByRole("button", { name: "Chat", exact: true }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-agent").inputValue() === "smoke-agent", "roster chat action should select smoke-agent");
  assert((await page.locator("#chat-input").inputValue()).includes("Continue as smoke-agent"), "roster chat action should draft a prompt");
  await openManagePanel(page, "Agents");
  await rosterCard.getByRole("button", { name: "Test", exact: true }).click();
  await page.locator("#panel-agents.active").waitFor();
  assert(await page.locator("#agent-test-agent").inputValue() === "smoke-agent", "roster test action should select smoke-agent");
  assert((await page.locator("#agent-test-prompt").inputValue()).includes("Test smoke-agent"), "roster test action should draft a test prompt");
  await rosterCard.getByRole("button", { name: "Council", exact: true }).click();
  await page.locator("#panel-council.active").waitFor();
  assert(await page.locator("#council-agent").inputValue() === "smoke-agent", "roster council action should select smoke-agent");
  assert((await page.locator("#council-goal").inputValue()).includes("Use smoke-agent as the lead agent"), "roster council action should draft a council goal");
  await openManagePanel(page, "Agents");
  const createdDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await rosterCard.getByRole("button", { name: "Edit profile" }).click();
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
  await page.getByLabel("Agent prompt").fill("You are an edited smoke agent.");
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
  await page.getByLabel("Agent memory").fill("Unsaved smoke memory.");
  await page.getByText("Unsaved changes").waitFor();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "unsaved new-agent dialog should be a confirm");
    await dialog.dismiss();
  });
  await page.getByRole("button", { name: "New agent" }).click();
  assert(await page.getByLabel("Agent memory").inputValue() === "Unsaved smoke memory.", "dismissed dirty dialog should keep editor values");
  await page.getByLabel("Agent memory").fill("Remember smoke.");
  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export config" }).click();
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
  await page.getByText("Agent config imported. Save agent to apply.").waitFor();
  assert(await page.getByLabel("Agent memory").inputValue() === "Imported smoke memory.", "import should update memory field");
  assert(await agentToolsAllow.inputValue() === "version\nfile_read\ngrep", "import should update allow rules");
  await page.locator("#agent-command-list").getByText("imported").waitFor();
  await page.getByText("Unsaved changes").waitFor();
  const importSavePromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "PUT"
  );
  await saveAgentButton.click();
  const importSaveResponse = await importSavePromise;
  assert(importSaveResponse.ok(), `agent import save failed with ${importSaveResponse.status()}`);
  await rosterCard.getByText("Manual review", { exact: true }).waitFor();
  await rosterCard.getByText("Approval gated", { exact: true }).waitFor();
  await rosterCard.getByText("30m · 10:00-18:00").waitFor();
  const updatedRosterMetrics = await rosterCard.locator(".agent-roster-metrics").innerText();
  assert(updatedRosterMetrics.includes("ALLOW\n3"), `agent roster should show imported allowed tool count, got ${JSON.stringify(updatedRosterMetrics)}`);
  assert(updatedRosterMetrics.includes("DENY\n2"), `agent roster should show edited denied tool count, got ${JSON.stringify(updatedRosterMetrics)}`);
  assert(updatedRosterMetrics.includes("COMMANDS\n2"), `agent roster should show deploy and imported commands, got ${JSON.stringify(updatedRosterMetrics)}`);
  await rosterCard.getByRole("button", { name: "/deploy", exact: true }).waitFor();
  await rosterCard.getByRole("button", { name: "/imported", exact: true }).waitFor();
  await page.getByRole("button", { name: "New agent" }).click();
  const updatedDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await page.locator("#agents-list [data-agent-detail=\"smoke-agent\"]").click();
  await updatedDetailPromise;
  const editedAllow = await agentToolsAllow.inputValue();
  const editedDeny = await agentToolsDeny.inputValue();
  assert(editedAllow === "version\nfile_read\ngrep", `agent allow rules should reload after import save, got ${JSON.stringify(editedAllow)}`);
  assert(editedDeny === "bash\nhttp", `agent deny rules should reload after edit, got ${JSON.stringify(editedDeny)}`);
  assert((await page.getByLabel("Agent memory").inputValue()).trim() === "Imported smoke memory.", "agent memory should reload after import save");
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
  await page.locator("#agent-test-form").getByRole("button", { name: "Run test" }).click();
  await page.locator("#agent-test-stop-button").waitFor();
  await page.locator("#panel-agents.active").waitFor();
  await page.locator("#agent-test-output").getByText("Agent test result").waitFor();
  await page.locator("#agent-test-output").getByText("agent test direct smoke").waitFor();
  await page.locator("#agent-test-output").getByText(capturedAgentTestRequestID).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "Open run" }).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "Open session" }).waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "Copy summary" }).click();
  await page.getByText("Agent test summary copied.").waitFor();
  const copiedAgentTestSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedAgentTestSummary.includes("Agent: smoke-agent"), "agent test summary missing agent");
  assert(copiedAgentTestSummary.includes("Prompt: agent test direct smoke"), "agent test summary missing prompt");
  assert(copiedAgentTestSummary.includes(capturedAgentTestRequestID), "agent test summary missing request id");
  assert(capturedAgentTest.agent === "smoke-agent", `agent test payload should use smoke-agent, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.text === "agent test direct smoke", `agent test payload should include prompt, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.new_session === true, `agent test payload should create a new session, got ${JSON.stringify(capturedAgentTest)}`);
  await openManagePanel(page, "Agents");
  await page.locator("#agent-test-prompt").fill("agent test cancellation smoke");
  await page.locator("#agent-test-form").getByRole("button", { name: "Run test" }).click();
  await page.locator("#agent-test-stop-button").waitFor();
  await page.locator("#agent-test-stop-button").click();
  await page.locator("#agent-test-output").getByText("Agent test cancelled").waitFor();
  await page.locator("#agent-test-form").getByRole("button", { name: "Run test" }).waitFor();
  await page.unroute(agentTestMessageRoute);
  await page.locator("#agents-list [data-agent-detail=\"smoke-agent\"]").click();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "agent delete dialog should be a confirm");
    await dialog.accept();
  });
  await page.locator("#agent-delete-button").click();
  await page.getByText("Agent deleted.").waitFor();
}

async function runRuns(page) {
  await page.getByRole("button", { name: "Chat" }).click();
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
  await chatRunSummary.getByText("Run summary").waitFor();
  await chatRunSummary.getByText("Agent").waitFor();
  await chatRunSummary.getByText("default").waitFor();
  await chatRunSummary.getByText("Usage").waitFor();
  await chatRunSummary.getByText("prompt_tokens: 3").waitFor();
  await chatRunSummary.getByText("Request").waitFor();
  await chatRunSummary.getByRole("button", { name: "Open run" }).waitFor();
  await chatRunSummary.getByRole("button", { name: "Copy summary" }).click();
  await page.getByText("Run summary copied.").waitFor();
  await chatRunSummary.getByRole("button", { name: "Copied" }).waitFor();
  await chatRunSummary.getByRole("button", { name: "Copy summary" }).waitFor();
  const copiedSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedSummary.includes("Session: sess_summary_smoke"), "copied summary missing session");
  assert(copiedSummary.includes("Agent: default"), "copied summary missing agent");
  assert(copiedSummary.includes("Usage: prompt_tokens: 3, completion_tokens: 4"), "copied summary missing usage");
  await chatRunSummary.getByRole("button", { name: "Suggest follow-up" }).click();
  await page.locator("#panel-home.active").waitFor();
  const summaryFollowUp = await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").inputValue();
  assert(summaryFollowUp.includes("Continue from this completed Astria run"), "summary follow-up should draft next prompt");
  assert(summaryFollowUp.includes("webui smoke session"), "summary follow-up should include original prompt");
  await page.getByRole("button", { name: "Chat", exact: true }).click();
  await chatRunSummary.getByRole("button", { name: "Open session" }).waitFor();
  await page.unroute("**/message");
  const runID = "run_history_smoke";
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
  await page.getByRole("button", { name: "Refresh data" }).click();
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#workspace-session-hub").getByRole("button", { name: /webui smoke session|sess_/ }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "workspace hub should resume latest session");
  await page.getByRole("button", { name: "Open Command Center" }).click();
  await page.getByLabel("Command search").fill("run_history_smoke");
  await page.locator("#command-center-list").getByRole("button", { name: /run_history_smoke/ }).click();
  await page.locator("#panel-runs.active").waitFor();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), runID);
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator("#panel-runs").getByRole("heading", { name: "Runs" }).waitFor();
  await page.locator("#mission-control-board").getByText("Completed").waitFor();
  await page.locator("#mission-control-filters").getByRole("button", { name: /Completed/ }).click();
  await page.locator("#mission-control-filters").getByRole("button", { name: /Completed/ }).evaluate((el) => {
    if (!el.classList.contains("active")) throw new Error("completed run filter should be active");
  });
  await page.locator("#mission-control-filters").getByRole("button", { name: /All/ }).click();
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
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
        started_at: new Date().toISOString(),
        ended_at: new Date().toISOString(),
        usage: { input_tokens: 7, output_tokens: 8 },
        request: { text: "webui smoke session", new_session: true, request_id: runID },
        response: {
          session_id: sessionID,
          messages: ["summary smoke response"],
          usage: { input_tokens: 7, output_tokens: 8 },
        },
        events: [
          { type: "preamble", at: new Date().toISOString(), data: { preamble: "planning smoke run" } },
          { type: "tool_call", at: new Date().toISOString(), data: { tool: "grep", args: JSON.stringify({ pattern: "smoke" }), status: "running" } },
          { type: "tool_result", at: new Date().toISOString(), data: { tool: "grep", content: "smoke result", status: "completed", is_error: false } },
          { type: "usage", at: new Date().toISOString(), data: { input_tokens: 7, output_tokens: 8 } },
          { type: "text", at: new Date().toISOString(), data: { text: "summary smoke response" } },
        ],
      }),
    });
  });
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  await page.waitForFunction((id) => document.querySelector("#run-detail")?.textContent?.includes(id), runID);
  await page.locator("#run-detail").getByText("Status").waitFor();
  assert(await page.locator("#run-detail").getByText("webui smoke session").count() >= 1, "run detail missing prompt");
  assert(await page.locator("#run-detail").getByText(sessionID).count() >= 1, "run detail missing session id");
  await page.locator("#run-detail").getByText("Time Travel").waitFor();
  await page.locator("#run-detail .run-milestone").getByText("Run completed").waitFor();
  await page.locator("#run-detail .run-milestone").getByText("Prompt locked").waitFor();
  await page.locator("#run-detail .run-milestone").getByRole("button", { name: "Open linked session" }).waitFor();
  assert(await page.locator("#run-detail .run-tool-event").count() === 1, "run detail should group tool call/result into one tool card");
  await page.locator("#run-detail .run-tool-event").getByText("grep").waitFor();
  await page.locator("#run-detail .run-tool-event").getByText("smoke result").waitFor();
  await page.locator("#run-detail .run-tool-event").getByRole("button", { name: "Copy result" }).click();
  await page.getByText("Tool result copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "smoke result", "tool result copy should copy grouped tool result");
  await page.locator("#run-detail").getByText("planning smoke run").waitFor();
  assert(await page.locator("#run-detail").getByText("input_tokens").count() >= 1, "run detail missing usage event");
  await page.locator("#run-detail").getByRole("button", { name: "Copy prompt" }).click();
  await page.getByText("Prompt copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "webui smoke session", "copy prompt should copy run prompt");
  await page.locator("#run-detail").getByRole("button", { name: "Copy result" }).click();
  await page.getByText("Result copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "summary smoke response", "copy result should copy formatted run response");
  await page.locator("#run-detail").getByRole("button", { name: "Copy summary" }).click();
  await page.getByText("Run summary copied.").waitFor();
  const copiedRunSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedRunSummary.includes(`Run: ${runID}`), "copied run detail summary missing run id");
  assert(copiedRunSummary.includes("Prompt: webui smoke session"), "copied run detail summary missing prompt");
  await page.locator("#run-detail").getByRole("button", { name: "Suggest follow-up" }).click();
  await page.locator("#panel-home.active").waitFor();
  const detailFollowUp = await page.getByPlaceholder("请输入任务，交给 Astria 来帮您完成").inputValue();
  assert(detailFollowUp.includes(`Run: ${runID}`), "run detail follow-up should include run id");
  assert(detailFollowUp.includes("Original prompt: webui smoke session"), "run detail follow-up should include original prompt");
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  await page.locator("#run-detail").getByRole("button", { name: "Re-run" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-input").inputValue() === "webui smoke session", "rerun should prefill chat prompt");
  assert(await page.locator("#chat-new-session").isChecked(), "rerun should use a new session");
  assert(await page.locator("#chat-agent").inputValue() === "", "rerun should select default agent");
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  await page.locator("#run-detail .run-milestone").getByRole("button", { name: "Open linked session" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "timeline open linked session should select run session");
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  await page.locator("#run-detail").getByRole("button", { name: "Open session" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "open session should select run session");
  await page.unroute(`**/runs/${runID}`);
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).waitFor();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Copy ID" }).click();
  await page.getByText("Session ID copied.").waitFor();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Copied" }).waitFor();
  const copiedSessionID = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedSessionID === sessionID, "copied session id should match row id");
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "prompt", "rename dialog should be a prompt");
    await dialog.accept("Smoke renamed session");
  });
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Rename" }).click();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByText("Smoke renamed session", { exact: true }).waitFor();
  await page.locator("#session-search").fill("Smoke renamed");
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).waitFor();
  await page.locator("#session-search-clear").click();
  assert(await page.locator("#session-search").inputValue() === "", "session search should clear");
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).waitFor();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Favorite" }).click();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Unfavorite" }).waitFor();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "delete dialog should be a confirm");
    await dialog.dismiss();
  });
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).getByRole("button", { name: "Delete" }).click();
  await page.locator(`#sessions-list [data-session-id="${sessionID}"]`).waitFor();

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
  await page.getByRole("button", { name: "Refresh data" }).click();
  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator("#panel-runs").getByRole("heading", { name: "Runs" }).waitFor();
  await page.locator(`[data-run-id="${errorRunID}"]`).waitFor();
  await page.locator(`[data-run-id="${errorRunID}"]`).getByRole("button", { name: "Open run" }).click();
  const errorRunDetail = await page.evaluate(async ({ url, errorRunID }) => {
    const response = await fetch(`${url}/runs/${errorRunID}`);
    return response.json();
  }, { url: baseURL, errorRunID });
  assert(errorRunDetail.status === "error", `provider-unavailable run should be error, got ${JSON.stringify(errorRunDetail)}`);
  assert(errorRunDetail.error || errorRunDetail.response?.error, `provider-unavailable run should include an error, got ${JSON.stringify(errorRunDetail)}`);
  assert(await page.locator("#run-detail").getByText(errorRunID).count() >= 1, "error run detail missing run id");
  assert(await page.locator("#run-detail").getByText("webui smoke provider unavailable").count() >= 1, "error run detail missing prompt");
  await page.locator("#run-detail").getByText("error", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(errorSessionID).count() >= 1, "error run detail missing session id");
  const errorResultText = await page.locator("#run-detail").locator("pre").last().innerText();
  assert(errorResultText.includes(errorSessionID), "error run result should include session id");
  assert(errorResultText.includes("error"), "error run result should include error field");
  await page.locator("#run-detail").getByRole("button", { name: "Copy prompt" }).click();
  await page.getByText("Prompt copied.").waitFor();
  assert(await page.evaluate(() => navigator.clipboard.readText()) === "webui smoke provider unavailable", "copy prompt should copy error run prompt");
  await page.locator("#run-detail").getByRole("button", { name: "Re-run" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator("#chat-input").inputValue() === "webui smoke provider unavailable", "error rerun should prefill chat prompt");
  assert(await page.locator("#chat-new-session").isChecked(), "error rerun should use a new session");
  assert(await page.locator("#chat-agent").inputValue() === "", "error rerun should select default agent");
}

async function runStreamingProvider(page) {
  const homePrompt = "webui home composer smoke";
  await page.getByRole("button", { name: "Home" }).click();
  await page.locator("#panel-home.active").waitFor();
  await page.locator("#home-task-input").fill(homePrompt);
  await page.getByRole("button", { name: "Launch task" }).click();
  await page.locator("#panel-chat.active").waitFor();
  await page.locator("#chat-output").getByText(homePrompt).waitFor();
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();
  await page.getByRole("button", { name: "New mission" }).click();

  const prompt = "webui streaming provider smoke";
  await page.getByRole("button", { name: "Chat" }).click();
  await page.locator("#chat-agent").selectOption("");
  await page.locator("#chat-new-session").check();
  await page.locator("#chat-input").fill(prompt);
  await page.getByRole("button", { name: "Send" }).click();

  await page.locator("#stop-button:not([hidden])").waitFor();
  await page.locator("#chat-output").getByText("Fake provider").waitFor();
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();
  const summary = page.locator("#chat-output .run-summary").last();
  await summary.getByText("Run summary").waitFor();
  await summary.getByText("input_tokens: 11").waitFor();
  await summary.getByText("output_tokens: 7").waitFor();
  const runID = await summary.getByRole("button", { name: "Open run" }).getAttribute("data-run-summary-run");
  const sessionID = await summary.getByRole("button", { name: "Open session" }).getAttribute("data-run-summary-session");
  assert(runID, "streaming run summary missing request id");
  assert(sessionID, "streaming run summary missing session id");

  await page.getByRole("button", { name: "Refresh data" }).click();
  await summary.getByRole("button", { name: "Open session" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "streaming open session should select persisted session");
  await page.locator("#chat-output").getByText("Fake provider streamed response for GUI smoke.").waitFor();

  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  assert(await page.locator("#run-detail").getByText(runID).count() >= 1, "streaming run detail missing run id");
  await page.locator("#run-detail").getByText("completed", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(prompt).count() >= 1, "streaming run detail missing prompt");
  assert(await page.locator("#run-detail").getByText("Fake provider streamed response for GUI smoke.").count() >= 1, "streaming run detail missing response text");
  assert(await page.locator("#run-detail").getByText(sessionID).count() >= 1, "streaming run detail missing session id");
  assert(await page.locator("#run-detail").getByText("input_tokens").count() >= 1, "streaming run detail missing usage");
  await page.locator("#run-detail").getByRole("button", { name: "Copy summary" }).click();
  await page.getByText("Run summary copied.").waitFor();
  const copiedRunSummary = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedRunSummary.includes(`Run: ${runID}`), "streaming copied run summary missing run id");
  assert(copiedRunSummary.includes("Usage: input_tokens: 11, output_tokens: 7, total_tokens: 18"), "streaming copied run summary missing usage");
}

async function runToolCallProvider(page) {
  const prompt = "webui tool call smoke";
  await page.getByRole("button", { name: "Chat" }).click();
  await page.locator("#chat-agent").selectOption("");
  await page.locator("#chat-new-session").check();
  await page.locator("#chat-input").fill(prompt);
  await page.getByRole("button", { name: "Send" }).click();

  const toolEvent = page.locator("#chat-output .tool-event").filter({ hasText: "version" }).first();
  await toolEvent.getByText("completed").waitFor();
  await toolEvent.locator("summary").click();
  await toolEvent.getByText("StarClaw").waitFor();
  await page.locator("#chat-output").getByText("Version tool call completed for GUI smoke.").waitFor();

  const summary = page.locator("#chat-output .run-summary").last();
  await summary.getByText("Run summary").waitFor();
  const runID = await summary.getByRole("button", { name: "Open run" }).getAttribute("data-run-summary-run");
  const sessionID = await summary.getByRole("button", { name: "Open session" }).getAttribute("data-run-summary-session");
  assert(runID, "tool-call run summary missing request id");
  assert(sessionID, "tool-call run summary missing session id");

  await page.getByRole("button", { name: "Refresh data" }).click();
  await summary.getByRole("button", { name: "Open session" }).click();
  await page.locator("#panel-chat.active").waitFor();
  assert(await page.locator(`[data-session-id="${sessionID}"].active`).count() === 1, "tool-call open session should select persisted session");
  await page.locator("#chat-output").getByText("Version tool call completed for GUI smoke.").waitFor();

  await page.getByRole("button", { name: "Runs", exact: true }).click();
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  assert(await page.locator("#run-detail").getByText(runID).count() >= 1, "tool-call run detail missing run id");
  await page.locator("#run-detail").getByText("completed", { exact: true }).first().waitFor();
  assert(await page.locator("#run-detail").getByText(prompt).count() >= 1, "tool-call run detail missing prompt");
  assert(await page.locator("#run-detail .run-tool-event").count() === 1, "tool-call run detail should group tool call/result into one tool card");
  const runToolEvent = page.locator("#run-detail .run-tool-event").first();
  await runToolEvent.getByText("version").waitFor();
  await runToolEvent.getByText("completed").waitFor();
  await runToolEvent.getByText("StarClaw").waitFor();
  await runToolEvent.getByRole("button", { name: "Copy result" }).click();
  await page.getByText("Tool result copied.").waitFor();
  assert((await page.evaluate(() => navigator.clipboard.readText())).includes("StarClaw"), "tool-call run detail tool copy should include version output");
  assert(await page.locator("#run-detail").getByText("Version tool call completed for GUI smoke.").count() >= 1, "tool-call run detail missing final response");
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
  curl -fsS "$BASE_URL/diagnostics" | grep -F '"config_path":"'"$SMOKE_HOME"'/.starclaw/config.yaml"' >/dev/null || fail "diagnostics JSON missing config path"
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
env HOME="$SMOKE_HOME" "$BIN" daemon start >"$DAEMON_LOG" 2>&1 &
DAEMON_PID="$!"
wait_for_health
check_routes
write_browser_smoke

echo "==> running browser smoke ($SMOKE_MODE)"
env BASE_URL="$BASE_URL" SCREENSHOT="$SCREENSHOT" HOME_SCREENSHOT="$HOME_SCREENSHOT" INTAKE_DOC="$INTAKE_DOC" NODE_DIR="$NODE_DIR" WEBUI_SMOKE_MODE="$SMOKE_MODE" SMOKE_HOME="$SMOKE_HOME" node "$NODE_SCRIPT"

echo "smoke_webui_${SMOKE_MODE}: ok"
echo "screenshot: $SCREENSHOT"
echo "home screenshot: $HOME_SCREENSHOT"
echo "daemon log: $DAEMON_LOG_ARTIFACT"
echo "metadata: $METADATA_ARTIFACT"
