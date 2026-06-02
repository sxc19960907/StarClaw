const state = {
  panel: "chat",
  agents: [],
  skills: [],
  sessions: [],
  schedules: [],
  diagnostics: null,
  config: null,
  permissions: null,
  editingAgent: "",
  selectedAgentCommand: "",
  agentCommands: {},
  agentDirty: false,
  agentDirtyBaseline: "",
  activeRequestID: "",
  activeAbort: null,
  activeSessionID: "",
  toolEvents: new Map(),
  toolDetails: new Map(),
  approvals: new Map(),
  eventSource: null,
};

const views = {
  chat: ["Chat", "Work with StarClaw from the local daemon."],
  agents: ["Agents", "Inspect named agents available to the daemon."],
  skills: ["Skills", "Review installed skills exposed to StarClaw."],
  schedules: ["Schedules", "Create and manage cron-based local tasks."],
  diagnostics: ["Diagnostics", "Inspect daemon readiness and setup checks."],
  config: ["Config", "Repair provider setup for daemon runs."],
  permissions: ["Permissions", "Review local tool policy."],
};

const $ = (id) => document.getElementById(id);

function showToast(message) {
  const toast = $("toast");
  toast.textContent = message;
  toast.classList.add("visible");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.remove("visible"), 2600);
}

async function copyText(text, successMessage = "Copied.") {
  await navigator.clipboard.writeText(text);
  showToast(successMessage);
}

function markButtonCopied(button) {
  const label = button.textContent;
  button.textContent = "Copied";
  button.disabled = true;
  clearTimeout(button.copyFeedbackTimer);
  button.copyFeedbackTimer = setTimeout(() => {
    button.textContent = label;
    button.disabled = false;
  }, 1400);
}

function debounce(fn, delay = 200) {
  let timer = 0;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  const text = await response.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text };
    }
  }
  if (!response.ok) {
    const message = data?.error || data?.message || response.statusText;
    throw new Error(message);
  }
  return data;
}

function normalizeName(item) {
  return item.name || item.Name || "";
}

function normalizeDescription(item) {
  return item.description || item.Description || "";
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function renderEmpty(target, message) {
  target.innerHTML = `<div class="empty-state">${escapeHTML(message)}</div>`;
}

function renderError(target, message) {
  target.innerHTML = `<div class="error-state">${escapeHTML(message)}</div>`;
}

function statusLabel(status) {
  switch (status) {
    case "ready":
      return "Ready";
    case "warning":
      return "Warning";
    case "needs_setup":
      return "Needs setup";
    case "error":
      return "Error";
    default:
      return "Unknown";
  }
}

function scrollConversationToBottom() {
  const scroller = document.querySelector(".conversation-scroll");
  if (scroller) scroller.scrollTop = scroller.scrollHeight;
}

function updateActiveSessionLabel() {
  const label = $("active-session-label");
  if (!label) return;
  if (!state.activeSessionID) {
    label.textContent = "No session selected";
    return;
  }
  const session = state.sessions.find((item) => item.id === state.activeSessionID);
  label.textContent = session ? `Session: ${session.title || session.id}` : `Session: ${state.activeSessionID}`;
}

function setRunControls(isRunning) {
  if (isRunning) $("chat-state").textContent = "Running";
  $("chat-input").disabled = isRunning;
  $("chat-agent").disabled = isRunning;
  $("chat-new-session").disabled = isRunning;
  $("send-button").hidden = isRunning;
  $("stop-button").hidden = !isRunning;
}

function handleChatInputKeydown(event) {
  if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
    event.preventDefault();
    if (!state.activeRequestID) $("chat-form").requestSubmit();
    return;
  }
  if (event.key === "Escape" && state.activeRequestID) {
    event.preventDefault();
    cancelActiveRun();
  }
}

function renderEmptyThread() {
  state.toolEvents.clear();
  state.toolDetails.clear();
  state.approvals.clear();
  $("chat-output").innerHTML = `<div class="empty-thread">
    <div class="assistant-mark" aria-hidden="true">S</div>
    <div>
      <strong id="chat-heading">StarClaw is ready.</strong>
      <span>Start a local task or choose an agent from the composer.</span>
    </div>
  </div>`;
}

async function selectSession(sessionID) {
  state.activeSessionID = sessionID || "";
  $("chat-new-session").checked = false;
  document.querySelectorAll("[data-session-id]").forEach((item) => {
    item.classList.toggle("active", item.dataset.sessionId === state.activeSessionID);
  });
  if (state.activeSessionID) {
    const session = state.sessions.find((item) => item.id === state.activeSessionID);
    $("chat-state").textContent = "Session selected";
    updateActiveSessionLabel();
    switchPanel("chat");
    try {
      const detail = await api(`/sessions/${encodeURIComponent(state.activeSessionID)}`);
      renderSessionThread(detail);
    } catch (error) {
      $("chat-output").innerHTML = `<div class="empty-thread">
        <div class="assistant-mark" aria-hidden="true">S</div>
        <div>
          <strong>${escapeHTML(session?.title || state.activeSessionID)}</strong>
          <span>${escapeHTML(error.message)}. Next message will still try to continue this session.</span>
        </div>
      </div>`;
    }
  }
}

function renderSessionThread(session) {
  const messages = Array.isArray(session.messages) ? session.messages : [];
  state.toolEvents.clear();
  state.toolDetails.clear();
  state.approvals.clear();
  $("chat-output").innerHTML = "";
  if (!messages.length) {
    $("chat-output").innerHTML = `<div class="empty-thread">
      <div class="assistant-mark" aria-hidden="true">S</div>
      <div>
        <strong>${escapeHTML(session.title || session.id || "Selected session")}</strong>
        <span>This session has no saved messages yet.</span>
      </div>
    </div>`;
    return;
  }
  for (const message of messages) {
    appendMessage(message.role || "assistant", message.content || "");
  }
  scrollConversationToBottom();
}

function startNewChat() {
  state.activeSessionID = "";
  $("chat-new-session").checked = true;
  $("chat-state").textContent = "Ready";
  updateActiveSessionLabel();
  document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
  renderEmptyThread();
  switchPanel("chat");
}

function connectEventStream() {
  if (!("EventSource" in window) || state.eventSource) return;
  const source = new EventSource("/events");
  state.eventSource = source;
  source.addEventListener("approval_needed", (event) => {
    renderApprovalCard(parseEventData(event.data));
  });
  source.addEventListener("approval_resolved", (event) => {
    markApprovalResolved(parseEventData(event.data));
  });
  source.onerror = () => {
    $("daemon-pill").textContent = "Reconnecting";
    $("daemon-pill").className = "bad";
  };
  source.onopen = () => {
    if ($("daemon-pill").textContent === "Reconnecting") {
      $("daemon-pill").textContent = "Running";
      $("daemon-pill").className = "ok";
    }
  };
}

function formatUptime(seconds) {
  if (!Number.isFinite(seconds)) return "-";
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

function switchPanel(panel) {
  if (!views[panel]) return;
  state.panel = panel;
  document.querySelectorAll(".nav-item").forEach((button) => {
    button.classList.toggle("active", button.dataset.panel === panel);
  });
  document.querySelectorAll(".panel").forEach((section) => {
    section.classList.toggle("active", section.id === `panel-${panel}`);
  });
  $("view-title").textContent = views[panel][0];
}

async function loadStatus() {
  const pill = $("daemon-pill");
  try {
    const status = await api("/status");
    $("metric-version").textContent = status.version || "-";
    $("metric-uptime").textContent = formatUptime(status.uptime);
    $("metric-active").textContent = status.active_agents ?? 0;
    pill.textContent = "Running";
    pill.className = "ok";
  } catch (error) {
    $("metric-version").textContent = "-";
    $("metric-uptime").textContent = "-";
    $("metric-active").textContent = "-";
    pill.textContent = "Offline";
    pill.className = "bad";
  }
}

async function loadDiagnostics() {
  const list = $("diagnostics-list");
  const summary = $("diagnostics-summary");
  const overview = $("diagnostics-overview");
  const pill = $("diagnostics-pill");
  const chip = $("diagnostics-chip");
  try {
    const diagnostics = await api("/diagnostics");
    state.diagnostics = diagnostics;
    const status = diagnostics.status || "unknown";
    const label = statusLabel(status);
    pill.textContent = label;
    pill.className = status;
    chip.textContent = label;
    chip.className = `diagnostics-chip ${status}`;
    summary.textContent = diagnostics.summary || "Runtime readiness checks.";
    overview.innerHTML = `<strong>${escapeHTML(label)}</strong><span>${escapeHTML(diagnostics.summary || "")}</span>`;
    renderConfigDiagnosticsOverview(diagnostics);
    const checks = Array.isArray(diagnostics.checks) ? diagnostics.checks : [];
    if (!checks.length) {
      renderEmpty(list, "No diagnostics returned.");
      return;
    }
    list.innerHTML = checks.map((check) => `<article class="row-item diagnostic-item ${escapeHTML(check.status || "unknown")}">
      <div class="row-item-title">
        <span>${escapeHTML(check.label || check.id || "Check")}</span>
        <span class="tag diagnostic-tag ${escapeHTML(check.status || "unknown")}">${escapeHTML(statusLabel(check.status))}</span>
      </div>
      <p>${escapeHTML(check.detail || "")}</p>
      ${diagnosticActionHTML(check)}
    </article>`).join("");
  } catch (error) {
    state.diagnostics = null;
    pill.textContent = "Offline";
    pill.className = "error";
    chip.textContent = "Diagnostics unavailable";
    chip.className = "diagnostics-chip error";
    summary.textContent = "Diagnostics unavailable.";
    overview.innerHTML = `<strong>Error</strong><span>${escapeHTML(error.message)}</span>`;
    renderConfigDiagnosticsOverview({ status: "error", summary: error.message });
    renderError(list, error.message);
  }
}

function diagnosticActionHTML(check) {
  if (!check.action) return "";
  if (check.id === "provider") {
    return `<button class="diagnostic-action-button" type="button" data-panel="config">${escapeHTML(check.action)}</button>`;
  }
  if (check.id === "permissions") {
    return `<button class="diagnostic-action-button" type="button" data-panel="permissions">${escapeHTML(check.action)}</button>`;
  }
  return `<div class="diagnostic-action">${escapeHTML(check.action)}</div>`;
}

async function loadConfig() {
  try {
    const data = await api("/config");
    state.config = data.config || {};
    renderConfigForm();
  } catch (error) {
    state.config = null;
    $("config-pill").textContent = "Error";
    $("config-pill").className = "bad";
    $("config-save-state").textContent = error.message;
  }
}

function renderConfigForm() {
  const cfg = state.config || {};
  $("config-provider").value = cfg.provider || "anthropic";
  $("config-endpoint").value = cfg.endpoint || "";
  $("config-model-tier").value = cfg.model_tier || "";
  $("config-openai-endpoint").value = cfg.openai_endpoint || "";
  $("config-openai-model").value = cfg.openai_model || "";
  $("config-ollama-endpoint").value = cfg.ollama_endpoint || "";
  $("config-ollama-model").value = cfg.ollama_model || "";
  $("config-api-key").value = "";
  $("config-openai-api-key").value = "";
  $("config-api-key").placeholder = cfg.api_key_set ? "Saved; leave blank to keep" : "Required for Anthropic";
  $("config-openai-api-key").placeholder = cfg.openai_api_key_set ? "Saved; leave blank to keep" : "Required for OpenAI";
  $("config-pill").textContent = cfg.provider || "Provider";
  $("config-pill").className = "";
  $("config-save-state").textContent = "Loaded";
  updateProviderFields();
}

function updateProviderFields() {
  const provider = $("config-provider").value || "anthropic";
  document.querySelectorAll("[data-provider-fields]").forEach((group) => {
    group.hidden = group.dataset.providerFields !== provider;
  });
}

function buildConfigPatch() {
  const provider = $("config-provider").value || "anthropic";
  const patch = { provider };
  if (provider === "anthropic") {
    patch.endpoint = $("config-endpoint").value.trim();
    patch.model_tier = $("config-model-tier").value.trim();
    const key = $("config-api-key").value.trim();
    if (key) patch.api_key = key;
  } else if (provider === "openai") {
    patch.openai_endpoint = $("config-openai-endpoint").value.trim();
    patch.openai_model = $("config-openai-model").value.trim();
    const key = $("config-openai-api-key").value.trim();
    if (key) patch.openai_api_key = key;
  } else if (provider === "ollama") {
    patch.ollama_endpoint = $("config-ollama-endpoint").value.trim();
    patch.ollama_model = $("config-ollama-model").value.trim();
  }
  return patch;
}

async function submitConfig(event) {
  event.preventDefault();
  $("config-save-state").textContent = "Saving";
  try {
    const result = await api("/config", {
      method: "PATCH",
      body: JSON.stringify(buildConfigPatch()),
    });
    state.config = result.config || state.config;
    renderConfigForm();
    await loadDiagnostics();
    showToast("Provider config saved.");
  } catch (error) {
    $("config-save-state").textContent = "Error";
    showToast(error.message);
  }
}

function renderConfigDiagnosticsOverview(diagnostics) {
  const target = $("config-diagnostics-overview");
  if (!target) return;
  const status = diagnostics?.status || "unknown";
  target.innerHTML = `<strong>${escapeHTML(statusLabel(status))}</strong><span>${escapeHTML(diagnostics?.summary || "Runtime diagnostics unavailable.")}</span>`;
}

async function loadPermissions() {
  const list = $("permissions-list");
  try {
    const data = await api("/permissions");
    state.permissions = data.permissions || {};
    renderPermissions();
  } catch (error) {
    state.permissions = null;
    $("permissions-pill").textContent = "Error";
    $("permissions-pill").className = "bad";
    $("permissions-overview").innerHTML = `<strong>Error</strong><span>${escapeHTML(error.message)}</span>`;
    renderError(list, error.message);
  }
}

function renderPermissions() {
  const permissions = state.permissions || {};
  const configured = permissions.configured === true;
  $("permissions-pill").textContent = configured ? "Configured" : "Defaults";
  $("permissions-pill").className = configured ? "ready" : "warning";
  $("permissions-overview").innerHTML = `<strong>${configured ? "Configured" : "Built-in defaults"}</strong><span>${configured ? "Config permissions are present." : "No explicit permissions config is present."}</span>`;
  const categories = [
    ["Allowed directories", permissions.allowed_dirs],
    ["Allowed commands", permissions.allowed_commands],
    ["Denied commands", permissions.denied_commands],
    ["Network allowlist", permissions.network_allowlist],
    ["Sensitive patterns", permissions.sensitive_patterns],
  ];
  $("permissions-list").innerHTML = categories.map(([label, values]) => {
    const items = Array.isArray(values) ? values : [];
    return `<article class="row-item permission-item">
      <div class="row-item-title"><span>${escapeHTML(label)}</span><span class="tag">${items.length}</span></div>
      ${items.length ? `<div class="pill-list">${items.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}</div>` : `<p>No explicit entries.</p>`}
    </article>`;
  }).join("");
}

async function loadAgents() {
  const list = $("agents-list");
  try {
    const data = await api("/agents");
    state.agents = data.agents || [];
    $("agents-count").textContent = state.agents.length;
    updateAgentSelects();
    if (!state.agents.length) {
      renderEmpty(list, "No named agents found.");
      return;
    }
    list.innerHTML = state.agents.map((agent) => {
      const name = normalizeName(agent);
      const description = normalizeDescription(agent) || "No description.";
      return `<article class="row-item">
        <div class="row-item-title"><span>${escapeHTML(name)}</span><span class="tag">agent</span></div>
        <p>${escapeHTML(description)}</p>
        <div class="row-actions"><button data-agent-detail="${escapeHTML(name)}">Edit</button></div>
      </article>`;
    }).join("");
  } catch (error) {
    renderError(list, error.message);
  }
}

function updateAgentSelects() {
  const options = ['<option value="">Default agent</option>'].concat(
    state.agents.map((agent) => {
      const name = normalizeName(agent);
      return `<option value="${escapeHTML(name)}">${escapeHTML(name)}</option>`;
    })
  ).join("");
  $("chat-agent").innerHTML = options;
  $("schedule-agent").innerHTML = options;
}

async function inspectAgent(name) {
  if (!confirmDiscardAgentChanges()) return;
  try {
    const detail = await api(`/agents/${encodeURIComponent(name)}`);
    fillAgentForm(detail);
  } catch (error) {
    $("agent-form-state").textContent = error.message;
  }
}

function startNewAgent() {
  if (!confirmDiscardAgentChanges()) return;
  applyAgentPayload({
    name: "",
    prompt: "",
    memory: "",
    model: "",
    reasoning_effort: "",
    tools_allow: [],
    tools_deny: [],
    auto_approve: false,
    heartbeat_every: "",
    heartbeat_active_hours: "",
    heartbeat_model: "",
    commands: {},
  }, { dirty: false });
}

function fillAgentForm(agent) {
  applyAgentPayload(agentPayloadFromDetail(agent), { dirty: false });
}

function parseCSVList(value) {
  return value.split(/[,\n]/).map((item) => item.trim()).filter(Boolean);
}

function formatRuleList(values) {
  return values.join("\n");
}

function stableAgentSnapshot(payload = buildAgentPayload()) {
  const commands = {};
  Object.keys(payload.commands || {}).sort((a, b) => a.localeCompare(b)).forEach((name) => {
    commands[name] = payload.commands[name];
  });
  return JSON.stringify({ ...payload, commands });
}

function setAgentDirtyBaseline() {
  state.agentDirtyBaseline = stableAgentSnapshot();
  updateAgentDirtyState();
}

function updateAgentDirtyState() {
  const snapshot = stableAgentSnapshot();
  state.agentDirty = snapshot !== state.agentDirtyBaseline;
  const base = state.editingAgent ? `Editing ${state.editingAgent}` : "New agent";
  $("agent-form-state").textContent = state.agentDirty ? `${base} · Unsaved changes` : base;
  renderAgentPermissionPreview();
}

function confirmDiscardAgentChanges() {
  return !state.agentDirty || globalThis.confirm("Discard unsaved agent changes?");
}

function buildAgentPayload() {
  return {
    name: $("agent-name").value.trim(),
    prompt: $("agent-prompt").value,
    memory: $("agent-memory").value,
    model: $("agent-model").value.trim(),
    reasoning_effort: $("agent-reasoning-effort").value.trim(),
    tools_allow: parseCSVList($("agent-tools-allow").value),
    tools_deny: parseCSVList($("agent-tools-deny").value),
    auto_approve: $("agent-auto-approve").checked,
    heartbeat_every: $("agent-heartbeat-every").value.trim(),
    heartbeat_active_hours: $("agent-heartbeat-active-hours").value.trim(),
    heartbeat_model: $("agent-heartbeat-model").value.trim(),
    commands: { ...state.agentCommands },
  };
}

function applyAgentPayload(payload, { dirty = true } = {}) {
  state.editingAgent = payload.name || "";
  $("agent-name").disabled = Boolean(state.editingAgent);
  $("agent-name").value = payload.name || "";
  $("agent-prompt").value = payload.prompt || "";
  $("agent-memory").value = payload.memory || "";
  $("agent-model").value = payload.model || "";
  $("agent-reasoning-effort").value = payload.reasoning_effort || "";
  $("agent-tools-allow").value = formatRuleList(payload.tools_allow || []);
  $("agent-tools-deny").value = formatRuleList(payload.tools_deny || []);
  $("agent-auto-approve").checked = payload.auto_approve === true;
  $("agent-heartbeat-every").value = payload.heartbeat_every || "";
  $("agent-heartbeat-active-hours").value = payload.heartbeat_active_hours || "";
  $("agent-heartbeat-model").value = payload.heartbeat_model || "";
  state.agentCommands = { ...(payload.commands || {}) };
  clearAgentCommandEditor();
  renderAgentCommands();
  $("agent-delete-button").hidden = !state.editingAgent;
  $("agent-test-run-button").hidden = !state.editingAgent;
  $("selected-agent-description").textContent = state.editingAgent ? "Editing named agent." : "Create a named agent.";
  if (dirty) {
    state.agentDirtyBaseline = "";
    updateAgentDirtyState();
  } else {
    setAgentDirtyBaseline();
  }
}

function agentPayloadFromDetail(agent) {
  const cfg = agent.Config || agent.config || {};
  const modelCfg = cfg.Agent || cfg.agent || {};
  const toolsCfg = cfg.Tools || cfg.tools || {};
  const heartbeatCfg = cfg.Heartbeat || cfg.heartbeat || {};
  const autoApprove = cfg.AutoApprove ?? cfg.auto_approve;
  return {
    name: agent.Name || agent.name || "",
    prompt: agent.Prompt || agent.prompt || "",
    memory: agent.Memory || agent.memory || "",
    model: modelCfg.Model || modelCfg.model || "",
    reasoning_effort: modelCfg.ReasoningEffort || modelCfg.reasoning_effort || "",
    tools_allow: toolsCfg.Allow || toolsCfg.allow || [],
    tools_deny: toolsCfg.Deny || toolsCfg.deny || [],
    auto_approve: autoApprove === true,
    heartbeat_every: heartbeatCfg.Every || heartbeatCfg.every || "",
    heartbeat_active_hours: heartbeatCfg.ActiveHours || heartbeatCfg.active_hours || "",
    heartbeat_model: heartbeatCfg.Model || heartbeatCfg.model || "",
    commands: { ...(agent.Commands || agent.commands || {}) },
  };
}

function renderAgentPermissionPreview() {
  const payload = buildAgentPayload();
  const allow = payload.tools_allow.length ? payload.tools_allow.join(", ") : "None";
  const deny = payload.tools_deny.length ? payload.tools_deny.join(", ") : "None";
  $("agent-permission-preview").innerHTML = `<div class="agent-preview-row"><strong>Allow</strong><span>${escapeHTML(allow)}</span></div>
    <div class="agent-preview-row"><strong>Deny</strong><span>${escapeHTML(deny)}</span></div>
    <div class="agent-preview-row"><strong>Auto approve</strong><span>${payload.auto_approve ? "Enabled" : "Disabled"}</span></div>`;
}

function renderAgentCommands() {
  const list = $("agent-command-list");
  const names = Object.keys(state.agentCommands).sort((a, b) => a.localeCompare(b));
  if (!names.length) {
    renderEmpty(list, "No custom commands.");
    return;
  }
  list.innerHTML = names.map((name) => {
    const active = name === state.selectedAgentCommand ? " active" : "";
    return `<div class="row-item${active}">
      <div class="row-item-title"><span>${escapeHTML(name)}</span><span class="tag">command</span></div>
      <div class="row-actions"><button type="button" data-agent-command="${escapeHTML(name)}">Edit</button></div>
    </div>`;
  }).join("");
}

function clearAgentCommandEditor() {
  state.selectedAgentCommand = "";
  $("agent-command-name").disabled = false;
  $("agent-command-name").value = "";
  $("agent-command-body").value = "";
  $("agent-command-save-button").textContent = "Add command";
  $("agent-command-delete-button").hidden = true;
}

function selectAgentCommand(name) {
  state.selectedAgentCommand = name;
  $("agent-command-name").disabled = false;
  $("agent-command-name").value = name;
  $("agent-command-body").value = state.agentCommands[name] || "";
  $("agent-command-save-button").textContent = "Update command";
  $("agent-command-delete-button").hidden = false;
  renderAgentCommands();
}

function saveAgentCommand() {
  const name = $("agent-command-name").value.trim();
  const body = $("agent-command-body").value.trim();
  if (!name || !body) {
    showToast("Command name and body are required.");
    return;
  }
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/.test(name)) {
    showToast("Command name must use letters, numbers, dashes, or underscores.");
    return;
  }
  if (state.selectedAgentCommand && state.selectedAgentCommand !== name) {
    delete state.agentCommands[state.selectedAgentCommand];
  }
  state.agentCommands[name] = body;
  selectAgentCommand(name);
  updateAgentDirtyState();
  showToast("Command staged.");
}

function deleteAgentCommand() {
  const name = state.selectedAgentCommand;
  if (!name) return;
  delete state.agentCommands[name];
  clearAgentCommandEditor();
  renderAgentCommands();
  updateAgentDirtyState();
  showToast("Command removed.");
}

function testCurrentAgent() {
  if (!confirmDiscardAgentChanges()) return;
  const name = state.editingAgent;
  if (!name) {
    showToast("Save the agent before testing.");
    return;
  }
  $("chat-agent").value = name;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
  updateSelectedAgent();
  updateActiveSessionLabel();
  $("chat-input").value = `Test ${name}: introduce your role and list one useful next step.`;
  switchPanel("chat");
  $("chat-input").focus();
  showToast(`Ready to test ${name}.`);
}

async function submitAgent(event) {
  event.preventDefault();
  const payload = buildAgentPayload();
  if (!payload.name || !payload.prompt.trim()) {
    showToast("Agent name and prompt are required.");
    return;
  }
  const path = state.editingAgent ? `/agents/${encodeURIComponent(state.editingAgent)}` : "/agents";
  const method = state.editingAgent ? "PUT" : "POST";
  $("agent-form-state").textContent = "Saving";
  try {
    const saved = await api(path, { method, body: JSON.stringify(payload) });
    await loadAgents();
    fillAgentForm(saved);
    updateAgentSelects();
    showToast("Agent saved.");
  } catch (error) {
    $("agent-form-state").textContent = "Error";
    showToast(error.message);
  }
}

async function deleteCurrentAgent() {
  if (!confirmDiscardAgentChanges()) return;
  const name = state.editingAgent;
  if (!name || !globalThis.confirm(`Delete agent "${name}"?`)) return;
  try {
    await api(`/agents/${encodeURIComponent(name)}`, { method: "DELETE" });
    await loadAgents();
    applyAgentPayload({
      name: "",
      prompt: "",
      memory: "",
      model: "",
      reasoning_effort: "",
      tools_allow: [],
      tools_deny: [],
      auto_approve: false,
      heartbeat_every: "",
      heartbeat_active_hours: "",
      heartbeat_model: "",
      commands: {},
    }, { dirty: false });
    showToast("Agent deleted.");
  } catch (error) {
    showToast(error.message);
  }
}

function exportAgentConfig() {
  const payload = buildAgentPayload();
  const name = payload.name || "agent";
  const blob = new Blob([JSON.stringify(payload, null, 2) + "\n"], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `${name}-config.json`;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
  showToast("Agent config exported.");
}

function normalizeImportedAgentPayload(data) {
  return {
    name: data.name || data.Name || "",
    prompt: data.prompt || data.Prompt || "",
    memory: data.memory || data.Memory || "",
    model: data.model || "",
    reasoning_effort: data.reasoning_effort || "",
    tools_allow: Array.isArray(data.tools_allow) ? data.tools_allow : [],
    tools_deny: Array.isArray(data.tools_deny) ? data.tools_deny : [],
    auto_approve: data.auto_approve === true,
    heartbeat_every: data.heartbeat_every || "",
    heartbeat_active_hours: data.heartbeat_active_hours || "",
    heartbeat_model: data.heartbeat_model || "",
    commands: data.commands && typeof data.commands === "object" ? data.commands : {},
  };
}

async function importAgentConfig(file) {
  if (!file) return;
  try {
    const text = await file.text();
    const data = JSON.parse(text);
    applyAgentPayload(normalizeImportedAgentPayload(data), { dirty: true });
    showToast("Agent config imported. Save agent to apply.");
  } catch (error) {
    showToast(`Import failed: ${error.message}`);
  } finally {
    $("agent-import-file").value = "";
  }
}

function updateSelectedAgent() {
  const name = $("chat-agent").value;
  const selected = state.agents.find((agent) => normalizeName(agent) === name);
  $("selected-agent-description").textContent = selected
    ? (normalizeDescription(selected) || "No description.")
    : "Select an agent.";
}

async function loadSkills() {
  const list = $("skills-list");
  try {
    const data = await api("/skills");
    state.skills = data.skills || [];
    $("skills-count").textContent = state.skills.length;
    if (!state.skills.length) {
      renderEmpty(list, "No skills installed.");
      return;
    }
    list.innerHTML = state.skills.map((skill) => `<article class="row-item">
      <div class="row-item-title"><span>${escapeHTML(skill.name)}</span><span class="tag">${escapeHTML(skill.source || "skill")}</span></div>
      <p>${escapeHTML(skill.description || "No description.")}</p>
    </article>`).join("");
  } catch (error) {
    renderError(list, error.message);
  }
}

async function loadSessions(query = "") {
  const list = $("sessions-list");
  try {
    const data = query
      ? await api(`/sessions/search?q=${encodeURIComponent(query)}`)
      : await api("/sessions");
    state.sessions = data.sessions || data.results || [];
    if (!state.sessions.length) {
      renderEmpty(list, query ? "No matching sessions." : "No sessions saved.");
      return;
    }
    list.innerHTML = state.sessions.map((session) => `<article class="row-item session-item ${session.id === state.activeSessionID ? "active" : ""}" data-session-id="${escapeHTML(session.id)}">
      <div class="row-item-title">
        <span>${session.favorite ? "★ " : ""}${escapeHTML(session.title || session.id)}</span>
        <button class="icon-danger-button" type="button" title="Delete session" aria-label="Delete session" data-session-delete="${escapeHTML(session.id)}">Delete</button>
      </div>
      <span class="tag">${session.msg_count ?? 0} messages</span>
      <p>${escapeHTML(session.id)}</p>
      <div class="row-actions">
        <button type="button" data-session-copy="${escapeHTML(session.id)}">Copy ID</button>
        <button type="button" data-session-rename="${escapeHTML(session.id)}">Rename</button>
        <button type="button" data-session-favorite="${escapeHTML(session.id)}" data-favorite="${session.favorite ? "false" : "true"}">${session.favorite ? "Unfavorite" : "Favorite"}</button>
      </div>
    </article>`).join("");
    updateActiveSessionLabel();
  } catch (error) {
    renderError(list, error.message);
  }
}

async function loadSchedules() {
  const list = $("schedules-list");
  try {
    const data = await api("/schedules");
    state.schedules = data.schedules || [];
    $("schedules-count").textContent = state.schedules.length;
    if (!state.schedules.length) {
      renderEmpty(list, "No schedules configured.");
      return;
    }
    list.innerHTML = state.schedules.map((schedule) => `<article class="row-item">
      <div class="row-item-title">
        <span>${escapeHTML(schedule.prompt || schedule.id)}</span>
        <span class="tag">${schedule.enabled ? "enabled" : "paused"}</span>
      </div>
      <p>${escapeHTML(schedule.cron || "")} ${schedule.agent ? `with ${schedule.agent}` : "with default agent"}</p>
      <div class="row-actions">
        <button data-schedule-toggle="${escapeHTML(schedule.id)}" data-enabled="${schedule.enabled ? "false" : "true"}">${schedule.enabled ? "Pause" : "Enable"}</button>
        <button data-schedule-delete="${escapeHTML(schedule.id)}">Delete</button>
      </div>
    </article>`).join("");
  } catch (error) {
    renderError(list, error.message);
  }
}

async function submitChat(event) {
  event.preventDefault();
  if (state.activeRequestID) {
    showToast("A run is already in progress.");
    return;
  }
  const text = $("chat-input").value.trim();
  if (!text) {
    showToast("Enter a prompt first.");
    return;
  }
  const output = $("chat-output");
  const stateLabel = $("chat-state");
  state.toolEvents.clear();
  state.toolDetails.clear();
  output.innerHTML = "";
  appendMessage("user", text);
  const assistantMessage = appendMessage("assistant", "");
  const abort = new AbortController();
  const payload = {
    text,
    agent: $("chat-agent").value,
    new_session: $("chat-new-session").checked,
    request_id: globalThis.crypto?.randomUUID ? globalThis.crypto.randomUUID() : `web-${Date.now()}`,
  };
  if (state.activeSessionID && !payload.new_session) {
    payload.session_id = state.activeSessionID;
  }
  state.activeRequestID = payload.request_id;
  state.activeAbort = abort;
  setRunControls(true);
  $("chat-input").value = "";
  try {
    const result = await streamMessage(payload, assistantMessage, abort.signal);
    renderDoneResult(result, assistantMessage);
    renderRunSummary(result, payload);
    if (result?.session_id) {
      state.activeSessionID = result.session_id;
      $("chat-new-session").checked = false;
    }
    stateLabel.textContent = "Complete";
    await loadSessions();
    updateActiveSessionLabel();
  } catch (error) {
    if (!assistantMessage.querySelector(".message-body").textContent.trim()) {
      assistantMessage.remove();
    }
    if (error.name === "AbortError") {
      appendMessage("system", "Run cancelled.");
      stateLabel.textContent = "Cancelled";
    } else {
      appendMessage("error", error.message);
      stateLabel.textContent = "Error";
    }
  } finally {
    state.activeRequestID = "";
    state.activeAbort = null;
    setRunControls(false);
    $("chat-input").focus();
    scrollConversationToBottom();
  }
}

function appendMessage(role, text) {
  const message = document.createElement("div");
  message.className = `message message-${role}`;
  message.innerHTML = `<span class="message-role">${escapeHTML(messageRoleLabel(role))}</span><div class="message-body">${escapeHTML(text)}</div>`;
  $("chat-output").appendChild(message);
  return message;
}

function messageRoleLabel(role) {
  switch (role) {
    case "user":
      return "You";
    case "assistant":
      return "StarClaw";
    case "system":
      return "System";
    case "error":
      return "Error";
    default:
      return role;
  }
}

function appendToolEvent(data, eventType) {
  const tool = data.tool || "tool";
  let event = state.toolEvents.get(tool);
  if (!event || eventType === "tool_call") {
    event = document.createElement("details");
    event.className = "tool-event";
    event.open = data.status === "error";
    event.dataset.tool = tool;
    $("chat-output").appendChild(event);
    state.toolEvents.set(tool, event);
  }
  const previous = state.toolDetails.get(tool) || {};
  const detail = {
    ...previous,
    ...data,
    args: data.args ?? previous.args,
    content: data.content ?? previous.content,
    status: data.status || previous.status || (eventType === "tool_call" ? "running" : "completed"),
  };
  state.toolDetails.set(tool, detail);
  const status = detail.status;
  event.classList.toggle("bad", status === "error" || data.is_error === true);
  const args = detail.args ? formatToolPayload(detail.args) : "";
  const content = detail.content ? formatToolPayload(detail.content) : "";
  const errorCategory = detail.error_category ? `<div class="tool-meta">category: ${escapeHTML(detail.error_category)}</div>` : "";
  event.innerHTML = `<summary>
    <span>${escapeHTML(tool)}</span>
    <strong>${escapeHTML(status)}</strong>
  </summary>
  <div class="tool-detail">
    ${args ? `<div class="tool-label">args</div><pre>${escapeHTML(args)}</pre>` : ""}
    ${content ? `<div class="tool-label">result</div><pre>${escapeHTML(content)}</pre>` : ""}
    ${errorCategory}
  </div>`;
}

function renderApprovalCard(data) {
  const requestID = data.request_id || "";
  if (!requestID || state.approvals.has(requestID)) return;
  switchPanel("chat");
  const card = document.createElement("div");
  card.className = "approval-card pending";
  card.dataset.approvalId = requestID;
  card.innerHTML = approvalCardHTML(data, "pending");
  $("chat-output").appendChild(card);
  state.approvals.set(requestID, { data, element: card });
  $("chat-state").textContent = "Approval required";
  scrollConversationToBottom();
}

function approvalCardHTML(data, status) {
  const args = data.args ? formatToolPayload(data.args) : "";
  const reason = data.reason || "Approval required";
  const statusLabel = status === "pending" ? "pending" : status;
  const disabled = status === "pending" ? "" : "disabled";
  return `<div class="approval-header">
    <span>Approval required</span>
    <strong>${escapeHTML(statusLabel)}</strong>
  </div>
  <div class="approval-body">
    <div><span>Tool</span><strong>${escapeHTML(data.tool || "tool")}</strong></div>
    <div><span>Reason</span><strong>${escapeHTML(reason)}</strong></div>
    ${data.agent ? `<div><span>Agent</span><strong>${escapeHTML(data.agent)}</strong></div>` : ""}
    ${args ? `<pre>${escapeHTML(args)}</pre>` : ""}
  </div>
  <div class="approval-actions">
    <button class="primary-button" data-approval-decision="allow" data-approval-id="${escapeHTML(data.request_id || "")}" ${disabled}>Allow</button>
    <button class="danger-button" data-approval-decision="deny" data-approval-id="${escapeHTML(data.request_id || "")}" ${disabled}>Deny</button>
  </div>`;
}

function markApprovalResolved(data) {
  const requestID = data.request_id || "";
  const item = state.approvals.get(requestID);
  if (!item) return;
  const status = data.decision === "allow" ? "allowed" : "denied";
  item.element.classList.remove("pending", "allowed", "denied");
  item.element.classList.add(status);
  item.element.innerHTML = approvalCardHTML(item.data, status);
  $("chat-state").textContent = status === "allowed" ? "Approval allowed" : "Approval denied";
}

async function resolveApproval(requestID, decision) {
  const item = state.approvals.get(requestID);
  if (!item) return;
  item.element.querySelectorAll("button").forEach((button) => {
    button.disabled = true;
  });
  try {
    await api("/approval", {
      method: "POST",
      body: JSON.stringify({ request_id: requestID, decision }),
    });
    markApprovalResolved({ request_id: requestID, decision });
    showToast(decision === "allow" ? "Tool approved." : "Tool denied.");
  } catch (error) {
    item.element.querySelectorAll("button").forEach((button) => {
      button.disabled = false;
    });
    showToast(error.message);
  }
}

function formatToolPayload(value) {
  if (typeof value !== "string") return JSON.stringify(value, null, 2);
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}

function appendAssistantText(message, text) {
  const body = message.querySelector(".message-body");
  body.textContent += text;
}

function renderDoneResult(result, assistantMessage) {
  if (!result) return;
  const body = assistantMessage.querySelector(".message-body");
  if (!body.textContent.trim() && Array.isArray(result.messages)) {
    body.textContent = result.messages.join("\n");
  }
  if (result.error) {
    appendMessage("error", result.error);
  }
}

function renderRunSummary(result, payload) {
  if (!result || result.error) return;
  const usage = result.usage || {};
  const sessionID = result.session_id || "";
  const agent = payload.agent || "default";
  const usageItems = Object.entries(usage)
    .filter(([, value]) => Number.isFinite(value))
    .map(([key, value]) => `${key}: ${value}`);
  const usageText = usageItems.length ? usageItems.join(", ") : "-";
  const requestID = payload.request_id || "-";
  const summaryText = [
    `Session: ${sessionID || "-"}`,
    `Agent: ${agent}`,
    `Usage: ${usageText}`,
    `Request: ${requestID}`,
  ].join("\n");
  const card = document.createElement("div");
  card.className = "run-summary";
  const openAction = sessionID
    ? `
        <button type="button" data-run-summary-session="${escapeHTML(sessionID)}">Open session</button>
      `
    : "";
  card.innerHTML = `<div class="run-summary-title">Run summary</div>
    <div class="run-summary-grid">
      <span>Session</span><strong>${escapeHTML(sessionID || "-")}</strong>
      <span>Agent</span><strong>${escapeHTML(agent)}</strong>
      <span>Usage</span><strong>${escapeHTML(usageText)}</strong>
      <span>Request</span><strong>${escapeHTML(requestID)}</strong>
    </div>
    <div class="run-summary-actions">
      <button type="button" data-run-summary-copy="${escapeHTML(summaryText)}">Copy summary</button>
      ${openAction}
    </div>`;
  $("chat-output").appendChild(card);
}

async function streamMessage(payload, assistantMessage, signal) {
  const response = await fetch("/message", {
    method: "POST",
    headers: {
      "Accept": "text/event-stream",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    signal,
  });
  if (!response.ok || !response.body) {
    return api("/message", { method: "POST", body: JSON.stringify(payload) });
  }
  if (!response.headers.get("Content-Type")?.includes("text/event-stream")) {
    return response.json();
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let doneResult = null;

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop() || "";
    for (const rawEvent of events) {
      doneResult = handleSSEEvent(rawEvent, assistantMessage, doneResult);
    }
  }
  if (buffer.trim()) {
    doneResult = handleSSEEvent(buffer, assistantMessage, doneResult);
  }
  return doneResult;
}

function handleSSEEvent(rawEvent, assistantMessage, doneResult) {
  const parsed = parseSSE(rawEvent);
  if (!parsed) return doneResult;
  const data = parseEventData(parsed.data);
  switch (parsed.event) {
    case "text":
      appendAssistantText(assistantMessage, data.text || "");
      break;
    case "preamble":
      appendAssistantText(assistantMessage, data.preamble || "");
      break;
    case "tool_call":
    case "tool_result":
      appendToolEvent(data, parsed.event);
      break;
    case "done":
      doneResult = data;
      break;
    case "error":
      throw new Error(data.error || "stream failed");
  }
  scrollConversationToBottom();
  return doneResult;
}

async function cancelActiveRun() {
  const requestID = state.activeRequestID;
  if (!requestID) return;
  try {
    await fetch("/cancel", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ request_id: requestID }),
    });
  } catch {
    // Closing the stream also cancels the request context on the daemon.
  } finally {
    state.activeAbort?.abort();
  }
}

function parseSSE(rawEvent) {
  let event = "message";
  const dataLines = [];
  for (const line of rawEvent.split("\n")) {
    if (line.startsWith("event:")) {
      event = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice(5).trimStart());
    }
  }
  if (!dataLines.length) return null;
  return { event, data: dataLines.join("\n") };
}

function parseEventData(data) {
  try {
    return JSON.parse(data);
  } catch {
    return { text: data };
  }
}

async function submitSchedule(event) {
  event.preventDefault();
  const cron = $("schedule-cron").value.trim();
  const prompt = $("schedule-prompt").value.trim();
  if (!cron || !prompt) {
    showToast("Cron and prompt are required.");
    return;
  }
  try {
    await api("/schedules", {
      method: "POST",
      body: JSON.stringify({ cron, prompt, agent: $("schedule-agent").value }),
    });
    $("schedule-prompt").value = "";
    await loadSchedules();
    showToast("Schedule created.");
  } catch (error) {
    showToast(error.message);
  }
}

async function toggleSchedule(id, enabled) {
  try {
    await api(`/schedules/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify({ enabled }),
    });
    await loadSchedules();
  } catch (error) {
    showToast(error.message);
  }
}

async function deleteSchedule(id) {
  try {
    await api(`/schedules/${encodeURIComponent(id)}`, { method: "DELETE" });
    await loadSchedules();
  } catch (error) {
    showToast(error.message);
  }
}

async function deleteSession(id) {
  const session = state.sessions.find((item) => item.id === id);
  const label = session?.title || id;
  if (!globalThis.confirm(`Delete session "${label}"?`)) return;
  try {
    await api(`/sessions/${encodeURIComponent(id)}`, { method: "DELETE" });
    if (state.activeSessionID === id) {
      startNewChat();
    }
    await loadSessions($("session-search").value.trim());
    showToast("Session deleted.");
  } catch (error) {
    showToast(error.message);
  }
}

async function renameSession(id) {
  const session = state.sessions.find((item) => item.id === id);
  const nextTitle = globalThis.prompt("Rename session", session?.title || id);
  if (nextTitle === null) return;
  const title = nextTitle.trim();
  if (!title) {
    showToast("Session title cannot be empty.");
    return;
  }
  try {
    await api(`/sessions/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify({ title }),
    });
    await loadSessions($("session-search").value.trim());
    showToast("Session renamed.");
  } catch (error) {
    showToast(error.message);
  }
}

async function toggleSessionFavorite(id, favorite) {
  try {
    await api(`/sessions/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify({ favorite }),
    });
    await loadSessions($("session-search").value.trim());
  } catch (error) {
    showToast(error.message);
  }
}

async function copySessionID(button) {
  await copyText(button.dataset.sessionCopy, "Session ID copied.");
  markButtonCopied(button);
}

async function refreshAll() {
  await Promise.allSettled([
    loadStatus(),
    loadDiagnostics(),
    loadConfig(),
    loadPermissions(),
    loadAgents(),
    loadSkills(),
    loadSessions(),
    loadSchedules(),
  ]);
}

document.addEventListener("click", (event) => {
  const approvalButton = event.target.closest("[data-approval-decision]");
  if (approvalButton) {
    resolveApproval(approvalButton.dataset.approvalId, approvalButton.dataset.approvalDecision);
    return;
  }

  const sessionDelete = event.target.closest("[data-session-delete]");
  if (sessionDelete) {
    event.stopPropagation();
    deleteSession(sessionDelete.dataset.sessionDelete);
    return;
  }

  const sessionRename = event.target.closest("[data-session-rename]");
  if (sessionRename) {
    event.stopPropagation();
    renameSession(sessionRename.dataset.sessionRename);
    return;
  }

  const sessionFavorite = event.target.closest("[data-session-favorite]");
  if (sessionFavorite) {
    event.stopPropagation();
    toggleSessionFavorite(sessionFavorite.dataset.sessionFavorite, sessionFavorite.dataset.favorite === "true");
    return;
  }

  const sessionCopy = event.target.closest("[data-session-copy]");
  if (sessionCopy) {
    event.stopPropagation();
    copySessionID(sessionCopy).catch((error) => showToast(error.message));
    return;
  }

  const nav = event.target.closest("[data-panel]");
  if (nav) switchPanel(nav.dataset.panel);

  const sessionItem = event.target.closest("[data-session-id]");
  if (sessionItem) selectSession(sessionItem.dataset.sessionId);

  const runSummarySession = event.target.closest("[data-run-summary-session]");
  if (runSummarySession) {
    selectSession(runSummarySession.dataset.runSummarySession);
    return;
  }

  const runSummaryCopy = event.target.closest("[data-run-summary-copy]");
  if (runSummaryCopy) {
    copyText(runSummaryCopy.dataset.runSummaryCopy, "Run summary copied.")
      .then(() => markButtonCopied(runSummaryCopy))
      .catch((error) => showToast(error.message));
    return;
  }

  const agentButton = event.target.closest("[data-agent-detail]");
  if (agentButton) inspectAgent(agentButton.dataset.agentDetail);

  const agentCommand = event.target.closest("[data-agent-command]");
  if (agentCommand) selectAgentCommand(agentCommand.dataset.agentCommand);

  const toggle = event.target.closest("[data-schedule-toggle]");
  if (toggle) toggleSchedule(toggle.dataset.scheduleToggle, toggle.dataset.enabled === "true");

  const remove = event.target.closest("[data-schedule-delete]");
  if (remove) deleteSchedule(remove.dataset.scheduleDelete);
});

$("refresh-button").addEventListener("click", refreshAll);
$("new-chat-button").addEventListener("click", startNewChat);
$("chat-new-session").addEventListener("change", () => {
  if ($("chat-new-session").checked) {
    state.activeSessionID = "";
    document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
    updateActiveSessionLabel();
  }
});
$("chat-form").addEventListener("submit", submitChat);
$("chat-input").addEventListener("keydown", handleChatInputKeydown);
$("stop-button").addEventListener("click", cancelActiveRun);
$("schedule-form").addEventListener("submit", submitSchedule);
$("config-form").addEventListener("submit", submitConfig);
$("config-provider").addEventListener("change", updateProviderFields);
$("agent-form").addEventListener("submit", submitAgent);
$("new-agent-button").addEventListener("click", startNewAgent);
$("agent-delete-button").addEventListener("click", deleteCurrentAgent);
$("agent-test-run-button").addEventListener("click", testCurrentAgent);
$("agent-export-button").addEventListener("click", exportAgentConfig);
$("agent-import-button").addEventListener("click", () => $("agent-import-file").click());
$("agent-import-file").addEventListener("change", (event) => importAgentConfig(event.target.files?.[0]));
$("agent-form").addEventListener("input", updateAgentDirtyState);
$("agent-command-save-button").addEventListener("click", saveAgentCommand);
$("agent-command-new-button").addEventListener("click", clearAgentCommandEditor);
$("agent-command-cancel-button").addEventListener("click", clearAgentCommandEditor);
$("agent-command-delete-button").addEventListener("click", deleteAgentCommand);
$("chat-agent").addEventListener("change", updateSelectedAgent);
$("session-search-form").addEventListener("submit", (event) => {
  event.preventDefault();
  loadSessions($("session-search").value.trim());
});
const debouncedSessionSearch = debounce(() => loadSessions($("session-search").value.trim()));
$("session-search").addEventListener("input", debouncedSessionSearch);
$("session-search-clear").addEventListener("click", () => {
  $("session-search").value = "";
  loadSessions();
  $("session-search").focus();
});

connectEventStream();
refreshAll();
