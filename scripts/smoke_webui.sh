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
const context = await browser.newContext({ viewport: { width: 1440, height: 960 }, acceptDownloads: true });
await context.grantPermissions(["clipboard-read", "clipboard-write"], { origin: baseURL });
const page = await context.newPage();

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
  await page.locator("#permissions-form").getByText("Allowed directories").waitFor();
  await page.locator("#permissions-form").getByText("Network allowlist").waitFor();
  await page.getByLabel("Allowed directories").fill("~\n.\n/tmp/smoke");
  await page.getByLabel("Allowed commands").fill("go test\nstarclaw version");
  await page.getByLabel("Denied commands").fill("shutdown\nreboot");
  await page.getByLabel("Network allowlist").fill("api.github.com\nsmoke.example.com");
  await page.getByLabel("Sensitive patterns").fill("*.secret\n.env.smoke");
  await page.getByRole("button", { name: "Save permissions" }).click();
  await page.getByText("Permissions saved.").waitFor();
  await page.locator("#permissions-list").getByText("/tmp/smoke").waitFor();
  await page.locator("#permissions-list").getByText("starclaw version").waitFor();
  await page.locator("#permissions-list").getByText("smoke.example.com").waitFor();
  assert((await page.getByLabel("Allowed directories").inputValue()).includes("/tmp/smoke"), "permissions editor should retain saved allowed dirs");
  await page.getByRole("button", { name: "Clear rules" }).click();
  await page.getByText("Permissions saved.").waitFor();
  await page.locator("#permissions-overview").getByText("Built-in defaults").waitFor();
  assert(await page.getByLabel("Allowed directories").inputValue() === "", "clear rules should empty allowed dirs");
  assert(await page.getByLabel("Allowed commands").inputValue() === "", "clear rules should empty allowed commands");

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
  await page.locator("#agent-permission-preview").getByText("Enabled").waitFor();
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
  await page.getByRole("button", { name: "New agent" }).click();
  const updatedDetailPromise = page.waitForResponse((response) =>
    response.url().endsWith("/agents/smoke-agent") && response.request().method() === "GET"
  );
  await page.locator("[data-agent-detail=\"smoke-agent\"]").click();
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
  await page.route("**/message", async (route) => {
    capturedAgentTest = route.request().postDataJSON();
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        session_id: "sess_agent_test_smoke",
        messages: ["agent test smoke response"],
        usage: { prompt_tokens: 5, completion_tokens: 6 },
      }),
    });
  });
  await page.locator("#agent-test-prompt").fill("agent test direct smoke");
  await page.locator("#agent-test-form").getByRole("button", { name: "Run test" }).click();
  await page.locator("#agent-test-output").getByText("Agent test result").waitFor();
  await page.locator("#agent-test-output").getByText("agent test smoke response").waitFor();
  await page.locator("#agent-test-output").getByText("sess_agent_test_smoke").waitFor();
  await page.locator("#agent-test-output").getByText("prompt_tokens: 5").waitFor();
  await page.locator("#agent-test-output").getByRole("button", { name: "Open run" }).waitFor();
  assert(capturedAgentTest.agent === "smoke-agent", `agent test payload should use smoke-agent, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.text === "agent test direct smoke", `agent test payload should include prompt, got ${JSON.stringify(capturedAgentTest)}`);
  assert(capturedAgentTest.new_session === true, `agent test payload should create a new session, got ${JSON.stringify(capturedAgentTest)}`);
  await page.unroute("**/message");
  await page.locator("[data-agent-detail=\"smoke-agent\"]").click();
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

  await page.getByRole("button", { name: /Chat/ }).click();
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
  await page.getByRole("button", { name: /Runs/ }).click();
  await page.locator("#panel-runs").getByRole("heading", { name: "Runs" }).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).waitFor();
  await page.locator(`[data-run-id="${runID}"]`).getByRole("button", { name: "Open run" }).click();
  await page.locator("#run-detail").getByText(runID).waitFor();
  await page.locator("#run-detail").getByText("Status").waitFor();
  await page.locator("#run-detail").getByText("webui smoke session").waitFor();
  assert(await page.locator("#run-detail").getByText(sessionID).count() >= 1, "run detail missing session id");
  await page.locator(`[data-session-id="${sessionID}"]`).waitFor();
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Copy ID" }).click();
  await page.getByText("Session ID copied.").waitFor();
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Copied" }).waitFor();
  const copiedSessionID = await page.evaluate(() => navigator.clipboard.readText());
  assert(copiedSessionID === sessionID, "copied session id should match row id");
  page.once("dialog", async (dialog) => {
    assert(dialog.type() === "prompt", "rename dialog should be a prompt");
    await dialog.accept("Smoke renamed session");
  });
  await page.locator(`[data-session-id="${sessionID}"]`).getByRole("button", { name: "Rename" }).click();
  await page.getByText("Smoke renamed session").waitFor();
  await page.locator("#session-search").fill("Smoke renamed");
  await page.locator(`[data-session-id="${sessionID}"]`).waitFor();
  await page.locator("#session-search-clear").click();
  assert(await page.locator("#session-search").inputValue() === "", "session search should clear");
  await page.locator(`[data-session-id="${sessionID}"]`).waitFor();
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
env BASE_URL="$BASE_URL" SCREENSHOT="$SCREENSHOT" NODE_DIR="$NODE_DIR" node "$NODE_SCRIPT"

echo "smoke_webui: ok"
echo "screenshot: $SCREENSHOT"
