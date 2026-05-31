const state = {
  panel: "chat",
  agents: [],
  skills: [],
  sessions: [],
  schedules: [],
  activeRequestID: "",
  activeAbort: null,
  activeSessionID: "",
  toolEvents: new Map(),
  toolDetails: new Map(),
};

const views = {
  chat: ["Chat", "Work with StarClaw from the local daemon."],
  agents: ["Agents", "Inspect named agents available to the daemon."],
  skills: ["Skills", "Review installed skills exposed to StarClaw."],
  schedules: ["Schedules", "Create and manage cron-based local tasks."],
};

const $ = (id) => document.getElementById(id);

function showToast(message) {
  const toast = $("toast");
  toast.textContent = message;
  toast.classList.add("visible");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.remove("visible"), 2600);
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

function scrollConversationToBottom() {
  const scroller = document.querySelector(".conversation-scroll");
  if (scroller) scroller.scrollTop = scroller.scrollHeight;
}

function setRunControls(isRunning) {
  if (isRunning) $("chat-state").textContent = "Running";
  $("chat-input").disabled = isRunning;
  $("chat-agent").disabled = isRunning;
  $("chat-new-session").disabled = isRunning;
  $("send-button").hidden = isRunning;
  $("stop-button").hidden = !isRunning;
}

function renderEmptyThread() {
  state.toolEvents.clear();
  state.toolDetails.clear();
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
  document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
  renderEmptyThread();
  switchPanel("chat");
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
        <div class="row-actions"><button data-agent-detail="${escapeHTML(name)}">Inspect</button></div>
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
  try {
    const detail = await api(`/agents/${encodeURIComponent(name)}`);
    $("agent-detail").textContent = JSON.stringify(detail, null, 2);
  } catch (error) {
    $("agent-detail").textContent = error.message;
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
        <span>${escapeHTML(session.title || session.id)}</span>
        <button class="icon-danger-button" type="button" title="Delete session" aria-label="Delete session" data-session-delete="${escapeHTML(session.id)}">Delete</button>
      </div>
      <span class="tag">${session.msg_count ?? 0} messages</span>
      <p>${escapeHTML(session.id)}</p>
    </article>`).join("");
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
    if (result?.session_id) {
      state.activeSessionID = result.session_id;
      $("chat-new-session").checked = false;
    }
    stateLabel.textContent = "Complete";
    await loadSessions();
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
    scrollConversationToBottom();
  }
}

function appendMessage(role, text) {
  const message = document.createElement("div");
  message.className = `message message-${role}`;
  message.innerHTML = `<span class="message-role">${escapeHTML(role)}</span><div class="message-body">${escapeHTML(text)}</div>`;
  $("chat-output").appendChild(message);
  return message;
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

async function refreshAll() {
  await Promise.allSettled([
    loadStatus(),
    loadAgents(),
    loadSkills(),
    loadSessions(),
    loadSchedules(),
  ]);
}

document.addEventListener("click", (event) => {
  const sessionDelete = event.target.closest("[data-session-delete]");
  if (sessionDelete) {
    event.stopPropagation();
    deleteSession(sessionDelete.dataset.sessionDelete);
    return;
  }

  const nav = event.target.closest("[data-panel]");
  if (nav) switchPanel(nav.dataset.panel);

  const sessionItem = event.target.closest("[data-session-id]");
  if (sessionItem) selectSession(sessionItem.dataset.sessionId);

  const agentButton = event.target.closest("[data-agent-detail]");
  if (agentButton) inspectAgent(agentButton.dataset.agentDetail);

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
  }
});
$("chat-form").addEventListener("submit", submitChat);
$("stop-button").addEventListener("click", cancelActiveRun);
$("schedule-form").addEventListener("submit", submitSchedule);
$("chat-agent").addEventListener("change", updateSelectedAgent);
$("session-search-form").addEventListener("submit", (event) => {
  event.preventDefault();
  loadSessions($("session-search").value.trim());
});

refreshAll();
