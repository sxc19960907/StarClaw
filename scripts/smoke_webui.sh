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
curl -fsS "$BASE_URL/diagnostics" | grep -F '"checks"' >/dev/null || fail "diagnostics JSON missing checks"
curl -fsS "$BASE_URL/permissions" | grep -F '"configured":true' >/dev/null || fail "permissions JSON missing configured policy"
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

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 960 } });

try {
  await page.goto(`${baseURL}/app/`, { waitUntil: "domcontentloaded" });
  await page.getByRole("heading", { name: "Chat" }).waitFor();
  await page.getByPlaceholder("Message StarClaw").waitFor();
  await page.getByRole("button", { name: "Send" }).waitFor();
  assert(await page.locator(".sidebar").count() === 1, "sidebar missing");
  await page.locator("#diagnostics-chip").waitFor();
  await page.getByRole("button", { name: /Diagnostics/ }).click();
  await page.locator("#panel-diagnostics").getByRole("heading", { name: "Diagnostics" }).waitFor();
  await page.getByText("Provider", { exact: true }).waitFor();
  await page.getByText(/Ollama is configured/).waitFor();
  await page.getByRole("button", { name: "Fix provider setup" }).click();
  await page.locator("#panel-config").getByRole("heading", { name: "Config" }).waitFor();
  await page.getByLabel("Provider").selectOption("ollama");
  await page.getByLabel("Ollama endpoint").fill("http://127.0.0.1:1");
  await page.getByLabel("Ollama model").fill("smoke-gui-model");
  await page.getByRole("button", { name: "Save provider config" }).click();
  await page.getByText("Provider config saved.").waitFor();
  await page.getByRole("button", { name: /Permissions/ }).click();
  await page.locator("#panel-permissions").getByRole("heading", { name: "Permissions" }).waitFor();
  await page.getByText("Allowed directories").waitFor();
  await page.getByText("Network allowlist").waitFor();

  await page.getByRole("button", { name: /Agents/ }).click();
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
  const saveCommandButton = page.locator("#agent-command-save-button");
  const clearCommandButton = page.locator("#agent-command-clear-button");
  const deleteCommandButton = page.locator("#agent-command-delete-button");
  const saveAgentButton = page.locator("#agent-form button[type=\"submit\"]");
  await agentToolsAllow.fill("file_read\ngrep");
  await agentToolsDeny.fill("bash");
  await agentAutoApprove.check();
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
  const createdDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await page.locator("[data-agent-detail=\"smoke-agent\"]").click();
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
  await clearCommandButton.click();
  assert(await agentCommandName.inputValue() === "", "clear command should reset command name");
  assert(await agentCommandBody.inputValue() === "", "clear command should reset command body");
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
  await page.getByRole("button", { name: "New agent" }).click();
  const updatedDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await page.locator("[data-agent-detail=\"smoke-agent\"]").click();
  await updatedDetailPromise;
  const editedAllow = await agentToolsAllow.inputValue();
  const editedDeny = await agentToolsDeny.inputValue();
  assert(editedAllow === "version\nfile_read", `agent allow rules should reload after edit, got ${JSON.stringify(editedAllow)}`);
  assert(editedDeny === "bash\nhttp", `agent deny rules should reload after edit, got ${JSON.stringify(editedDeny)}`);
  assert(!(await agentAutoApprove.isChecked()), "agent auto approve should reload after edit");
  assert(await agentHeartbeatEvery.inputValue() === "30m", "agent heartbeat interval should reload after edit");
  assert(await agentHeartbeatActiveHours.inputValue() === "10:00-18:00", "agent heartbeat active hours should reload after edit");
  assert(await agentHeartbeatModel.inputValue() === "smoke-heartbeat-edited", "agent heartbeat model should reload after edit");
  assert(await page.locator("#agent-command-list").getByText("deploy").count() === 1, "agent command list should reload after edit");
  assert(await page.locator("#agent-command-list").getByText("review").count() === 0, "deleted agent command should stay deleted after reload");
  assert(await page.locator("#agent-command-list").getByText("audit").count() === 0, "deleted renamed agent command should stay deleted after reload");
  await page.locator("#agent-command-list [data-agent-command=\"deploy\"]").click();
  assert((await agentCommandBody.inputValue()).trim() === "Deploy smoke changes safely.", "agent command body should reload after edit");
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "agent delete dialog should be a confirm");
    await dialog.accept();
  });
  await page.locator("#agent-delete-button").click();
  await page.getByText("Agent deleted.").waitFor();

  await page.getByRole("button", { name: /Schedules/ }).click();
  await page.getByLabel("Cron expression").fill("* * * * *");
  await page.getByLabel("Schedule prompt").fill("webui smoke schedule");
  await page.getByRole("button", { name: "Create schedule" }).click();
  await page.getByText("webui smoke schedule").waitFor();
  await page.getByRole("button", { name: "Pause" }).click();
  await page.getByRole("button", { name: "Enable" }).waitFor();
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByText("No schedules configured.").waitFor();

  const sessionID = await page.evaluate(async (url) => {
    const response = await fetch(`${url}/message`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text: "webui smoke session", new_session: true })
    });
    const data = await response.json();
    return data.session_id;
  }, baseURL);
  assert(sessionID, "session id missing");
  await page.getByRole("button", { name: /Chat/ }).click();
  await page.getByRole("button", { name: "Refresh data" }).click();
  await page.locator(`[data-session-id="${sessionID}"]`).waitFor();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "prompt", "rename dialog should be a prompt");
    await dialog.accept("Smoke renamed session");
  });
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Rename" }).click();
  await page.getByText("Smoke renamed session").waitFor();
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Favorite" }).click();
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Unfavorite" }).waitFor();
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "confirm", "delete dialog should be a confirm");
    await dialog.dismiss();
  });
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Delete" }).click();
  await page.locator(`[data-session-id="${sessionID}"]`).waitFor();

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
