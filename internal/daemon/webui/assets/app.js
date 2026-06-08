const state = {
  panel: "home",
  agents: [],
  skills: [],
  sessions: [],
  schedules: [],
  runs: [],
  councilRuns: [],
  inboxItems: [],
  inboxProviders: [],
  intakeResult: null,
  currentRunDetail: null,
  currentCouncilRun: null,
  selectedComparisonLane: "",
  selectedPromptVariant: "",
  selectedReuseAsset: "",
  selectedDeliveryLane: "",
  selectedSourceRow: "",
  promptLabGoal: "",
  diagnostics: null,
  config: null,
  permissions: null,
  memory: null,
  version: null,
  updateCheck: null,
  editingAgent: "",
  selectedAgentCommand: "",
  agentCommands: {},
  agentDirty: false,
  agentDirtyBaseline: "",
  activeRequestID: "",
  activeRunID: "",
  activeAgentTestRequestID: "",
  activeAgentTestAbort: null,
  activeAbort: null,
  activeSessionID: "",
  toolEvents: new Map(),
  toolDetails: new Map(),
  approvals: new Map(),
  eventSource: null,
  homeMode: "general",
  workflowStrategy: "direct",
  workflowStage: "draft",
  workflowStageLabel: "General mission",
  memoryCategory: "all",
  editingMCPServer: "",
  runFilter: "all",
};

const views = {
  home: ["首页", "Launch local missions from Astria."],
  chat: ["消息", "Work with Astria from the local daemon."],
  manage: ["管理", "Configure agent resources and local automation."],
  settings: ["设置", "Inspect daemon setup, permissions, and build state."],
  agents: ["智能体", "Inspect named agents available to the daemon."],
  skills: ["技能", "Review installed skills exposed to Astria."],
  mcp: ["MCP 星港", "Inspect configured MCP servers and docking readiness."],
  memory: ["记忆星图", "Review source sessions and draft memory candidates."],
  sources: ["来源登记", "Inspect freshness and reliability for knowledge sources."],
  council: ["智能体议会", "Coordinate planner, researcher, and reviewer roles."],
  compare: ["比较工作台", "Compare runs, agents, memory, and council evidence."],
  promptlab: ["Prompt Lab", "Test prompt variants across agents and context sources."],
  reuse: ["复用星库", "Start from reusable prompts, agents, sources, and outcomes."],
  delivery: ["主动投递", "Monitor scheduled work and outbound channel readiness."],
  inbox: ["收件箱", "Review inbound channel tasks before running them."],
  intake: ["文件星舱", "Inspect local documents and archives before a run."],
  schedules: ["定时任务", "Create and manage cron-based local tasks."],
  runs: ["运行", "Inspect recent daemon executions."],
  diagnostics: ["诊断", "Inspect daemon readiness and setup checks."],
  config: ["连接器", "Repair provider setup for daemon runs."],
  permissions: ["权限", "Review local tool policy."],
  version: ["版本", "Inspect build and update status."],
};

const homeActions = {
  publish: {
    title: "Publish resource",
    status: "Ready",
    description: "整理当前项目中可以交付或发布的资源，并生成打包清单。",
    prompt: "Prepare a publishable resource from the current project and list the files you would package.",
    notice: "已为发布资源任务预填提示，可以直接启动。",
  },
  browser: {
    title: "Browser probe",
    status: "Ready",
    description: "用于网页检查、截图、表单验证和变更摘要；从 Chat 中按需请求浏览器操作。",
    prompt: "Use browser automation to inspect the relevant page and summarize what changed.",
    notice: "已为浏览器检查预填提示，Astria 会在需要操作网页时说明动作。",
  },
  data: {
    title: "Data signal",
    status: "Ready",
    description: "从当前工作区的数据、日志或导出文件里找出关键结论。",
    prompt: "Analyze the local data or logs in this workspace and return the key signal.",
    notice: "数据分析会从当前工作区上下文开始。",
  },
  writing: {
    title: "Writing pass",
    status: "Ready",
    description: "起草、润色或压缩文字交付物，适合 PRD、说明、汇报和发布稿。",
    prompt: "Draft a concise, polished write-up for this task.",
  },
  research: {
    title: "Research orbit",
    status: "Ready",
    description: "进行带证据链的调研，并输出可追踪来源和结论。",
    prompt: "Run deep research for this task and produce an evidence-backed brief.",
  },
  council: {
    title: "Agent Council",
    status: "Ready",
    description: "多智能体规划和评审模式。启动一个议会运行，分别生成规划、调研和评审意见。",
    prompt: "Split this task across multiple named agents and propose a coordination plan.",
    panel: "council",
    notice: "Agent Council 会先生成可审核的角色贡献，不会自动执行代码改动。",
  },
  desktop: {
    title: "Desktop control",
    status: "Guarded",
    description: "需要操作本机 UI 时使用；Astria 会先说明动作并等待授权。",
    prompt: "Use desktop control only if needed and explain the intended action first.",
    notice: "桌面控制需要明确授权，Astria 会先说明动作。",
  },
  files: {
    title: "File Intake",
    status: "Ready",
    description: "读取本地文档、检查归档内容，并把结果送入普通任务流。",
    prompt: "Inspect local files and recommend the safest next edit.",
    panel: "intake",
    notice: "已打开 File Intake。",
  },
  mcp: {
    title: "MCP Starport",
    status: "Ready",
    description: "查看配置的 MCP 服务器、连接测试和可用工具。",
    prompt: "Review MCP docking options and suggest the first server to connect.",
    panel: "mcp",
    notice: "已打开 MCP Starport。",
  },
  memory: {
    title: "Memory Map",
    status: "Ready",
    description: "查看记忆文件、会话来源，并审核写入 MEMORY.md 的候选内容。",
    prompt: "Create a memory map for this project: people, decisions, recurring tasks, and useful files.",
    panel: "memory",
    notice: "已打开 Memory Map。",
  },
};

const workflowRecipes = {
  "code-review": {
    title: "代码评审",
    status: "Review",
    description: "检查当前改动的风险、回归点和测试缺口。",
    prompt: "Review the current working tree like a senior engineer. Lead with concrete findings, include file/line references where possible, and call out missing tests or risky behavior.",
    outcome: "一份按严重程度排序的评审报告，包含文件位置、行为风险和测试缺口。",
    context: ["当前 git diff", "相关测试输出", "高风险改动路径"],
    checklist: ["先列 findings", "标明残余风险", "建议最小验证命令"],
  },
  "feature-plan": {
    title: "功能规划",
    status: "Plan",
    description: "把一个产品想法拆成 PRD、设计和可验证实施步骤。",
    prompt: "Turn this feature idea into a concise PRD, technical design, implementation plan, and validation checklist. Keep the scope shippable and aligned with the current codebase.",
    outcome: "一份可落地的 PRD、设计边界、实施顺序和验收清单。",
    context: ["目标用户", "现有代码路径", "非目标范围"],
    checklist: ["定义验收标准", "识别复用点", "拆成可提交切片"],
  },
  "file-intake": {
    title: "文件理解",
    status: "Files",
    description: "先进入 File Intake 读取文档或归档，再把结果送入任务。",
    prompt: "Use File Intake to inspect the relevant local document or archive, then summarize the important content and propose the next action.",
    panel: "intake",
    outcome: "把本地文件内容整理成可引用上下文，再决定是否进入 Chat 或 run。",
    context: ["文件路径", "读取模式", "提取出的关键段落"],
    checklist: ["选择 intake mode", "审查结果摘要", "发送到下一步任务"],
  },
  "research-brief": {
    title: "调研简报",
    status: "Research",
    description: "生成带证据链的调研结论和行动建议。",
    prompt: "Prepare a research brief for this topic. Separate facts, assumptions, tradeoffs, and recommended next steps. Include sources if external research is needed.",
    outcome: "一份区分事实、假设、取舍和建议的证据链简报。",
    context: ["研究问题", "可信来源", "时间敏感点"],
    checklist: ["确认是否需要联网", "记录来源", "输出建议路径"],
  },
  "mcp-setup": {
    title: "工具接入",
    status: "MCP",
    description: "规划新的 MCP dock，检查配置，并测试连接。",
    prompt: "Help set up an MCP server for this workflow. Identify the server command or URL, required env keys, safety considerations, and a test plan.",
    panel: "mcp",
    outcome: "一个可审查的 MCP 接入方案，包含命令、环境变量、安全边界和测试。",
    context: ["server command 或 URL", "env keys", "工具权限范围"],
    checklist: ["补齐配置", "运行连接测试", "记录失败处理"],
  },
  "inbox-triage": {
    title: "任务分拣",
    status: "Inbox",
    description: "审核外部渠道任务，决定拒绝、重试或转成运行。",
    prompt: "Triage the pending Inbox items. Identify which should become runs, which need more context, and which should be rejected.",
    panel: "inbox",
    outcome: "把外部任务分成可运行、需补充、应拒绝三类，并保留处理轨迹。",
    context: ["待处理 inbox 项", "来源渠道", "缺失上下文"],
    checklist: ["审查来源", "决定处理动作", "转成可追踪 run"],
  },
  "memory-update": {
    title: "记忆更新",
    status: "Memory",
    description: "从最近工作中提炼决策、偏好、命令和风险。",
    prompt: "Draft a memory update from recent work. Categorize decisions, preferences, commands, architecture notes, people, and risks. Do not write memory without review.",
    panel: "memory",
    outcome: "一组经过分类的记忆候选，等待审核后再写入项目记忆。",
    context: ["最近会话", "决策和偏好", "风险或命令"],
    checklist: ["分类候选", "检查重复和冲突", "审核后再写入"],
  },
};

const workflowStrategies = {
  direct: {
    title: "Quick Run",
    status: "Fast",
    description: "最短路径进入本地执行，适合范围清楚、风险较低的任务。",
    prompt: "Execute this task directly in the current workspace. Keep the scope tight, report the changed files, and run the relevant validation.",
    panel: "runs",
    stageLabel: "Quick local execution",
    outcome: "Astria 直接推进任务，并在运行记录中保留结果。",
    checks: ["确认范围", "执行最小改动", "验证并汇报"],
  },
  research: {
    title: "Research Brief",
    status: "Deep",
    description: "先做证据链、方案取舍和上下文归纳，再进入执行。",
    prompt: "Prepare a research brief before implementation. Separate facts, assumptions, options, tradeoffs, and recommended next steps.",
    panel: "runs",
    stageLabel: "Research before execution",
    outcome: "先形成可审查的研究简报，减少盲目执行。",
    checks: ["列出证据", "标注假设", "给出建议路径"],
  },
  council: {
    title: "Agent Council",
    status: "Swarm",
    description: "把复杂任务拆给规划、调研和评审角色，再合并成执行方案。",
    prompt: "Coordinate this task through multiple named agents. Ask planner, researcher, and reviewer roles for input, then synthesize a concrete plan.",
    panel: "council",
    stageLabel: "Council strategy",
    outcome: "多智能体先分工评估，再收敛到一个可执行方案。",
    checks: ["拆分角色", "合并观点", "保留评审意见"],
  },
  guarded: {
    title: "Human Approval",
    status: "Gate",
    description: "高风险命令、文件写入或外部动作先进入人工确认路径。",
    prompt: "Plan this task with explicit approval gates. Identify risky commands, file writes, network calls, and rollback points before acting.",
    panel: "permissions",
    stageLabel: "Guarded approval path",
    outcome: "先标记风险动作和回滚点，再推进需要授权的步骤。",
    checks: ["识别风险", "设置审批点", "准备回滚"],
  },
  memory: {
    title: "Memory Capture",
    status: "Recall",
    description: "先从最近工作中提炼项目事实、偏好和风险，再继续任务。",
    prompt: "Draft a memory capture before continuing. Extract decisions, preferences, commands, risks, and project facts without writing durable memory until reviewed.",
    panel: "memory",
    stageLabel: "Memory capture strategy",
    outcome: "把上下文沉淀成可审核记忆，降低后续重复解释。",
    checks: ["提炼事实", "检查冲突", "审核后写入"],
  },
  tooling: {
    title: "MCP Tooling",
    status: "Tools",
    description: "先检查 MCP dock、外部工具和连接状态，再启动工具密集任务。",
    prompt: "Review the required tools for this task. Check MCP docks, missing environment keys, safety boundaries, and a minimal connection test plan.",
    panel: "mcp",
    stageLabel: "Tooling readiness",
    outcome: "先确认工具 dock 和权限边界，再进入执行。",
    checks: ["检查 dock", "确认 env", "测试连接"],
  },
};

const $ = (id) => document.getElementById(id);

function setText(id, value) {
  const target = $(id);
  if (target) target.textContent = value;
}

function setClass(id, value = "") {
  const target = $(id);
  if (target) target.className = value;
}

function showToast(message) {
  const toast = $("toast");
  toast.textContent = message;
  toast.classList.add("visible");
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.remove("visible"), 2600);
}

function hideToast() {
  clearTimeout(showToast.timer);
  $("toast").classList.remove("visible");
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

function renderEmptyAction(target, message, actions = []) {
  const buttons = actions.map((action) => {
    const attrs = action.panel ? ` data-panel="${escapeHTML(action.panel)}"` : action.homeAction ? ` data-home-action="${escapeHTML(action.homeAction)}"` : action.action ? ` data-action="${escapeHTML(action.action)}"` : "";
    const className = action.primary ? ' class="primary-button"' : "";
    return `<button type="button"${className}${attrs}>${escapeHTML(action.label)}</button>`;
  }).join("");
  target.innerHTML = `<div class="empty-state empty-action"><span>${escapeHTML(message)}</span>${buttons ? `<div class="inline-actions">${buttons}</div>` : ""}</div>`;
}

function renderError(target, message) {
  target.innerHTML = `<div class="error-state">${escapeHTML(message)}</div>`;
}

function commandCenterItems() {
  const recipeItems = Object.entries(workflowRecipes).map(([id, recipe]) => ({
    id: `recipe:${id}`,
    type: "Workflow",
    title: recipe.title,
    detail: recipe.description || recipe.outcome || "",
    run: () => selectWorkflowRecipe(id),
  }));
  const panelItems = [
    ["home", "Home", "Return to Astria launch workspace."],
    ["chat", "Chat", "Open the current conversation surface."],
    ["runs", "Mission Control", "Review recent runs and status filters."],
    ["intake", "File Intake", "Inspect local documents and archives."],
    ["memory", "Memory Map", "Review durable context candidates."],
    ["mcp", "MCP Starport", "Manage configured tool docks."],
    ["council", "Agent Council", "Coordinate planner, researcher, and reviewer roles."],
    ["inbox", "Inbox", "Triage external channel work."],
    ["schedules", "Schedules", "Manage recurring local tasks."],
  ].map(([panel, title, detail]) => ({
    id: `panel:${panel}`,
    type: "Panel",
    title,
    detail,
    run: () => switchPanel(panel),
  }));
  const actionItems = [
    ["research", "Deep research", "Prepare an evidence-backed brief."],
    ["mcp", "Plan MCP setup", "Draft a first tool dock connection."],
    ["memory", "Draft memory map", "Prepare reviewed memory candidates."],
    ["council", "Start Agent Council", "Split work across named roles."],
  ].map(([action, title, detail]) => ({
    id: `action:${action}`,
    type: "Action",
    title,
    detail,
    run: () => runHomeAction(action),
  }));
  const recentSessionItems = state.sessions.slice(0, 3).map((session) => ({
    id: `session:${session.id}`,
    type: "Recent",
    title: session.title || session.id,
    detail: `${session.msg_count ?? 0} messages · resume session`,
    run: () => selectSession(session.id),
  }));
  const recentRunItems = state.runs.slice(0, 3).map((run) => ({
    id: `run:${run.id}`,
    type: "Recent",
    title: run.prompt && run.id ? `${run.prompt} · ${run.id}` : run.prompt || run.id,
    detail: `${run.status || "unknown"} · ${run.agent || "default"} · open run`,
    run: () => selectRun(run.id),
  }));
  return [...recentSessionItems, ...recentRunItems, ...recipeItems, ...panelItems, ...actionItems];
}

function openCommandCenter() {
  const center = $("command-center");
  if (!center) return;
  center.hidden = false;
  $("command-center-input").value = "";
  renderCommandCenterList();
  $("command-center-input").focus();
}

function closeCommandCenter() {
  const center = $("command-center");
  if (!center) return;
  center.hidden = true;
}

function renderCommandCenterList() {
  const list = $("command-center-list");
  if (!list) return;
  const query = ($("command-center-input")?.value || "").trim().toLowerCase();
  const items = commandCenterItems().filter((item) => {
    const haystack = `${item.id} ${item.type} ${item.title} ${item.detail}`.toLowerCase();
    return !query || haystack.includes(query);
  }).slice(0, 16);
  if (!items.length) {
    renderEmpty(list, "No matching commands.");
    return;
  }
  list.innerHTML = items.map((item) => `<button type="button" data-command-id="${escapeHTML(item.id)}">
    <span>${escapeHTML(item.type)}</span>
    <strong>${escapeHTML(item.title)}</strong>
    <small>${escapeHTML(item.detail)}</small>
  </button>`).join("");
}

function runCommandCenterItem(id) {
  const item = commandCenterItems().find((candidate) => candidate.id === id);
  if (!item) return;
  closeCommandCenter();
  item.run();
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
  if ($("home-task-input")) $("home-task-input").disabled = isRunning;
  $("chat-agent").disabled = isRunning;
  if ($("home-agent")) $("home-agent").disabled = isRunning;
  $("chat-new-session").disabled = isRunning;
  $("send-button").hidden = isRunning;
  $("stop-button").hidden = !isRunning;
  document.querySelectorAll(".send-orbit-button").forEach((button) => {
    button.disabled = isRunning;
  });
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
  const diagnostics = state.diagnostics || {};
  const needsSetup = diagnostics.status && diagnostics.status !== "ready";
  const title = needsSetup ? "Astria needs setup." : "Astria is ready.";
  const subtitle = needsSetup
    ? (diagnostics.summary || "Open Diagnostics or Config to finish setup.")
    : "Start a local task or choose an agent from the composer.";
  $("chat-output").innerHTML = `<div class="empty-thread">
    <div class="assistant-mark" aria-hidden="true">S</div>
    <div>
      <strong id="chat-heading">${escapeHTML(title)}</strong>
      <span>${escapeHTML(subtitle)}</span>
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

function seedMissionPrompt(prompt, panel = "home") {
  const text = prompt || "";
  if (panel === "chat") {
    $("chat-input").value = text;
    switchPanel("chat");
    $("chat-input").focus();
    return;
  }
  $("home-task-input").value = text;
  switchPanel("home");
  $("home-task-input").focus();
}

function runHomeAction(name) {
  const action = homeActions[name];
  if (!action) return;
  state.homeMode = name;
  renderHomeMode();
  renderWorkflowBrief("");
  if (name === "council") {
    $("council-goal").value = action.prompt || "";
    switchPanel("council");
    $("council-goal").focus();
  } else if (action.panel) {
    switchPanel(action.panel);
  } else {
    seedMissionPrompt(action.prompt || "");
  }
  if (action.notice) showToast(action.notice);
}

function selectWorkflowRecipe(id) {
  const recipe = workflowRecipes[id];
  if (!recipe) return;
  state.homeMode = `recipe:${id}`;
  state.workflowStage = "draft";
  state.workflowStageLabel = recipe.title || "Workflow draft";
  $("home-task-input").value = recipe.prompt || "";
  renderHomeMode();
  renderWorkflowBrief(id);
  renderWorkflowStageRail();
  renderFocusBrief();
  renderPromptSuggestionDock();
  switchPanel("home");
  $("home-task-input").focus();
  showToast(`${recipe.title} workflow ready.`);
}

function selectWorkflowStrategy(id) {
  const strategy = workflowStrategies[id];
  if (!strategy) return;
  state.workflowStrategy = id;
  state.homeMode = `strategy:${id}`;
  state.workflowStage = "draft";
  state.workflowStageLabel = strategy.stageLabel || strategy.title || "Strategy draft";
  $("home-task-input").value = strategy.prompt || "";
  renderHomeMode();
  renderWorkflowBrief("");
  renderStrategyMatrix();
  renderWorkflowStageRail();
  renderFocusBrief();
  renderPromptSuggestionDock();
  switchPanel("home");
  $("home-task-input").focus();
  showToast(`${strategy.title} strategy ready.`);
}

function renderWorkflowBrief(id) {
  const brief = $("workflow-brief");
  if (!brief) return;
  const recipe = workflowRecipes[id];
  if (!recipe) {
    brief.innerHTML = `<div>
      <span class="board-kicker">Workflow brief</span>
      <strong>选择一个工作流，Astria 会生成可执行简报。</strong>
    </div>
    <p>简报会把目标、上下文材料、关联面板和下一步检查点放在同一个工作包里。</p>`;
    return;
  }
  const context = Array.isArray(recipe.context) ? recipe.context : [];
  const checklist = Array.isArray(recipe.checklist) ? recipe.checklist : [];
  const routeLabel = recipe.panel === "mcp" ? "打开星港" : recipe.panel === "memory" ? "打开星图" : recipe.panel === "council" ? "打开议会" : recipe.panel === "intake" ? "打开文件星舱" : recipe.panel === "inbox" ? "打开收件箱" : "打开关联面板";
  brief.innerHTML = `<div class="workflow-brief-head">
      <div>
        <span class="board-kicker">${escapeHTML(recipe.status || "Workflow")}</span>
        <strong>${escapeHTML(recipe.title || "Workflow brief")}</strong>
      </div>
      ${recipe.panel ? `<button type="button" data-panel="${escapeHTML(recipe.panel)}">${escapeHTML(routeLabel)}</button>` : ""}
    </div>
    <p>${escapeHTML(recipe.outcome || recipe.description || "")}</p>
    <div class="workflow-brief-grid">
      <div>
        <span>Context orbit</span>
        <ul>${context.map((item) => `<li>${escapeHTML(item)}</li>`).join("")}</ul>
      </div>
      <div>
        <span>Next checks</span>
        <ul>${checklist.map((item) => `<li>${escapeHTML(item)}</li>`).join("")}</ul>
      </div>
    </div>`;
}

function renderStrategyMatrix() {
  const matrix = $("strategy-matrix");
  if (!matrix) return;
  matrix.innerHTML = Object.entries(workflowStrategies).map(([id, strategy]) => {
    const active = state.workflowStrategy === id;
    return `<button type="button" class="strategy-card ${active ? "active" : ""}" data-strategy="${escapeHTML(id)}">
      <span>${escapeHTML(strategy.status || "Strategy")}</span>
      <strong>${escapeHTML(strategy.title || id)}</strong>
      <small>${escapeHTML(strategy.description || "")}</small>
    </button>`;
  }).join("");
  const brief = $("strategy-brief");
  if (!brief) return;
  const strategy = workflowStrategies[state.workflowStrategy] || workflowStrategies.direct;
  const checks = Array.isArray(strategy.checks) ? strategy.checks : [];
  const routeLabel = strategy.panel === "mcp" ? "打开星港" : strategy.panel === "memory" ? "打开星图" : strategy.panel === "council" ? "打开议会" : strategy.panel === "permissions" ? "打开权限" : strategy.panel === "runs" ? "打开运行" : "打开面板";
  brief.innerHTML = `<div class="strategy-brief-head">
      <div>
        <span class="board-kicker">${escapeHTML(strategy.status || "Strategy")}</span>
        <strong>${escapeHTML(strategy.title || "Strategy")}</strong>
      </div>
      ${strategy.panel ? `<button type="button" data-panel="${escapeHTML(strategy.panel)}">${escapeHTML(routeLabel)}</button>` : ""}
    </div>
    <p>${escapeHTML(strategy.outcome || strategy.description || "")}</p>
    <div class="strategy-checks">
      ${checks.map((check) => `<span>${escapeHTML(check)}</span>`).join("")}
    </div>`;
}

function renderHomeMode() {
  const action = state.homeMode.startsWith("recipe:")
    ? workflowRecipes[state.homeMode.slice("recipe:".length)]
    : state.homeMode.startsWith("strategy:")
      ? workflowStrategies[state.homeMode.slice("strategy:".length)]
    : homeActions[state.homeMode];
  const mode = action || {
    title: "General mission",
    status: "Ready",
    description: "直接描述目标，Astria 会从当前工作区和默认智能体开始。",
  };
  setText("home-mode-kicker", mode.status || "Ready");
  setText("home-mode-title", mode.title || "General mission");
  setText("home-mode-description", mode.description || "");
  const route = $("home-mode-route");
  if (!route) return;
  if (mode.panel) {
    route.hidden = false;
    route.dataset.panel = mode.panel;
    route.textContent = mode.panel === "mcp" ? "打开星港" : mode.panel === "memory" ? "打开星图" : mode.panel === "council" ? "打开议会" : mode.panel === "intake" ? "打开文件星舱" : mode.panel === "inbox" ? "打开收件箱" : mode.panel === "permissions" ? "打开权限" : mode.panel === "runs" ? "打开运行" : "打开面板";
  } else {
    route.hidden = true;
    delete route.dataset.panel;
  }
}

function currentWorkflowStage() {
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const latestRun = state.runs[0];
  if (memoryFacts.length || memoryWarnings.length) {
    return {
      stage: "memory",
      label: memoryWarnings.length ? `${memoryWarnings.length} memory warning${memoryWarnings.length === 1 ? "" : "s"}` : `${memoryFacts.length} memory fact${memoryFacts.length === 1 ? "" : "s"}`,
    };
  }
  if (latestRun) {
    const group = runHealthGroup(latestRun);
    if (group === "running") {
      return { stage: "running", label: latestRun.prompt || latestRun.id || "Running mission" };
    }
    return { stage: "review", label: latestRun.prompt || latestRun.id || "Review latest run" };
  }
  return {
    stage: state.workflowStage || "draft",
    label: state.workflowStageLabel || "General mission",
  };
}

function renderWorkflowStageRail() {
  const rail = $("workflow-stage-rail");
  if (!rail) return;
  const current = currentWorkflowStage();
  const order = ["draft", "running", "review", "memory"];
  const activeIndex = order.indexOf(current.stage);
  const stages = [
    ["draft", "Draft", "Recipe and prompt"],
    ["running", "Running", "Daemon execution"],
    ["review", "Review", "Mission Control"],
    ["memory", "Memory", "Durable context"],
  ];
  rail.innerHTML = stages.map(([key, label, hint], index) => {
    const active = current.stage === key;
    const done = activeIndex > index;
    return `<button type="button" class="workflow-stage ${active ? "active" : ""} ${done ? "done" : ""}" data-panel="${key === "running" || key === "review" ? "runs" : key === "memory" ? "memory" : "home"}">
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(active ? current.label : hint)}</strong>
    </button>`;
  }).join("");
}

function renderFocusBrief() {
  const target = $("focus-brief");
  if (!target) return;
  const stage = currentWorkflowStage();
  const recipe = state.homeMode.startsWith("recipe:")
    ? workflowRecipes[state.homeMode.slice("recipe:".length)]
    : null;
  const strategy = workflowStrategies[state.workflowStrategy] || workflowStrategies.direct;
  const latestSession = state.sessions[0];
  const latestRun = state.runs[0];
  const title = recipe?.title || stage.label || strategy.title || "General mission";
  const context = latestRun
    ? `${latestRun.status || "unknown"} run · ${latestRun.agent || "default"}`
    : latestSession
      ? `${latestSession.msg_count ?? 0} message session`
      : `${strategy.title || "Strategy"} · No recent work yet`;
  const next = stage.stage === "memory"
    ? "Review memory candidates"
    : stage.stage === "review"
      ? "Open Mission Control"
      : stage.stage === "running"
        ? "Monitor active run"
        : "Draft or launch the next mission";
  const runAction = latestRun?.id ? `<button type="button" data-run-open="${escapeHTML(latestRun.id)}">Open run</button>` : "";
  const sessionAction = latestSession?.id ? `<button type="button" data-session-resume="${escapeHTML(latestSession.id)}">Resume session</button>` : "";
  target.innerHTML = `<div class="focus-brief-head">
      <div>
        <span class="board-kicker">${escapeHTML(stage.stage)}</span>
        <strong>${escapeHTML(title)}</strong>
      </div>
      <small>${escapeHTML(next)}</small>
    </div>
    <div class="focus-brief-grid">
      <span>Context</span><strong>${escapeHTML(context)}</strong>
      <span>Strategy</span><strong>${escapeHTML(strategy.title || "Quick Run")}</strong>
      <span>Session</span><strong>${escapeHTML(latestSession?.title || latestSession?.id || "No session")}</strong>
      <span>Run</span><strong>${escapeHTML(latestRun?.prompt || latestRun?.id || "No run")}</strong>
    </div>
    <div class="focus-brief-actions">
      ${runAction}
      ${sessionAction}
      <button type="button" data-panel="${stage.stage === "memory" ? "memory" : "runs"}">${escapeHTML(stage.stage === "memory" ? "Open Memory" : "Open Mission Control")}</button>
    </div>`;
}

function renderWorkspaceHealthStrip() {
  const strip = $("workspace-health-strip");
  if (!strip) return;
  const diagnosticsStatus = state.diagnostics?.status || "unknown";
  const permissionsConfigured = state.permissions?.configured === true;
  const mcpServers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled).length;
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const items = [
    {
      panel: "diagnostics",
      tone: diagnosticsStatus === "ready" ? "ready" : diagnosticsStatus === "warning" ? "warning" : diagnosticsStatus === "unknown" ? "" : "attention",
      label: "Diagnostics",
      value: statusLabel(diagnosticsStatus),
      detail: state.diagnostics?.summary || "Runtime readiness",
    },
    {
      panel: "permissions",
      tone: permissionsConfigured ? "ready" : "warning",
      label: "Permissions",
      value: permissionsConfigured ? "Configured" : "Defaults",
      detail: permissionsConfigured ? "Explicit tool policy" : "Built-in guardrails",
    },
    {
      panel: "mcp",
      tone: enabledMCP ? "ready" : mcpServers.length ? "warning" : "",
      label: "MCP",
      value: enabledMCP ? `${enabledMCP} enabled` : mcpServers.length ? "Disabled" : "No docks",
      detail: "Tool connections",
    },
    {
      panel: "memory",
      tone: memoryWarnings.length ? "attention" : memoryFacts.length ? "ready" : "",
      label: "Memory",
      value: memoryWarnings.length ? `${memoryWarnings.length} warnings` : memoryFacts.length ? `${memoryFacts.length} facts` : "Preview",
      detail: memoryWarnings.length ? "Review taxonomy" : "Durable context",
    },
  ];
  strip.innerHTML = items.map((item) => `<button type="button" class="workspace-health-item ${escapeHTML(item.tone)}" data-panel="${escapeHTML(item.panel)}">
    <span>${escapeHTML(item.label)}</span>
    <strong>${escapeHTML(item.value)}</strong>
    <small>${escapeHTML(item.detail)}</small>
  </button>`).join("");
}

function reviewQueueItems() {
  const items = [];
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const activeRuns = state.runs.filter((run) => runHealthGroup(run) === "running");
  if (failedRuns.length) {
    const run = failedRuns[0];
    items.push({
      tone: "attention",
      label: "Run",
      title: `${failedRuns.length} run${failedRuns.length === 1 ? "" : "s"} need review`,
      detail: run.prompt || run.id || "Open Mission Control to inspect the failed run.",
      runID: run.id || "",
    });
  } else if (activeRuns.length) {
    const run = activeRuns[0];
    items.push({
      tone: "active",
      label: "Run",
      title: `${activeRuns.length} active mission${activeRuns.length === 1 ? "" : "s"}`,
      detail: run.prompt || run.id || "Monitor the active daemon execution.",
      runID: run.id || "",
    });
  }

  const inboxCounts = inboxStatusCounts();
  if (inboxCounts.failed) {
    items.push({
      tone: "attention",
      label: "Inbox",
      title: `${inboxCounts.failed} inbound item${inboxCounts.failed === 1 ? "" : "s"} failed`,
      detail: "Retry or reject failed channel work before it blocks the queue.",
      panel: "inbox",
    });
  } else if (inboxCounts.pending) {
    items.push({
      tone: "warning",
      label: "Inbox",
      title: `${inboxCounts.pending} inbound item${inboxCounts.pending === 1 ? "" : "s"} waiting`,
      detail: "Review external tasks before they become Astria runs.",
      panel: "inbox",
    });
  }

  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  if (memoryWarnings.length) {
    items.push({
      tone: "attention",
      label: "Memory",
      title: `${memoryWarnings.length} memory warning${memoryWarnings.length === 1 ? "" : "s"}`,
      detail: String(memoryWarnings[0] || "Review taxonomy warnings before adding durable context."),
      panel: "memory",
    });
  }

  const diagnosticsStatus = state.diagnostics?.status || "unknown";
  if (!["ready", "unknown"].includes(diagnosticsStatus)) {
    items.push({
      tone: diagnosticsStatus === "warning" ? "warning" : "attention",
      label: "Diagnostics",
      title: `Diagnostics ${statusLabel(diagnosticsStatus)}`,
      detail: state.diagnostics?.summary || "Inspect launch readiness checks.",
      panel: "diagnostics",
    });
  }

  if (state.permissions && state.permissions.configured !== true) {
    items.push({
      tone: "warning",
      label: "Permissions",
      title: "Permissions using defaults",
      detail: "Set explicit tool guardrails for this workspace.",
      panel: "permissions",
    });
  } else if (state.permissions) {
    const hints = permissionsRiskHints(state.permissions);
    if (hints.length) {
      items.push({
        tone: "warning",
        label: "Permissions",
        title: `${hints.length} policy hint${hints.length === 1 ? "" : "s"}`,
        detail: hints[0],
        panel: "permissions",
      });
    }
  }

  const mcpServers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled).length;
  if (mcpServers.length && !enabledMCP) {
    items.push({
      tone: "warning",
      label: "MCP",
      title: "MCP docks disabled",
      detail: "Enable or test a dock before tool-heavy workflows.",
      panel: "mcp",
    });
  } else if (!mcpServers.length && state.config) {
    items.push({
      tone: "",
      label: "MCP",
      title: "No MCP docks configured",
      detail: "Add a dock when this workspace needs external tools.",
      panel: "mcp",
    });
  }

  return items.slice(0, 6);
}

function renderReviewQueue() {
  const target = $("review-queue-list");
  if (!target) return;
  const items = reviewQueueItems();
  if (!items.length) {
    target.innerHTML = `<button type="button" class="review-queue-item clear" data-panel="runs">
      <span>Clear</span>
      <strong>队列已清空</strong>
      <small>没有失败运行、待审收件箱、记忆警告或配置风险。</small>
    </button>`;
    return;
  }
  target.innerHTML = items.map((item) => {
    const actionAttr = item.runID
      ? `data-run-open="${escapeHTML(item.runID)}"`
      : `data-panel="${escapeHTML(item.panel || "home")}"`;
    return `<button type="button" class="review-queue-item ${escapeHTML(item.tone)}" ${actionAttr}>
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(item.title)}</strong>
      <small>${escapeHTML(item.detail)}</small>
    </button>`;
  }).join("");
}

function approvalCenterItems() {
  const items = [];
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const inboxCounts = inboxStatusCounts();
  const diagnosticsStatus = state.diagnostics?.status || "unknown";
  const mcpServers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled).length;
  if (state.approvals.size) {
    items.push({
      tone: "attention",
      label: "Approvals",
      title: `${state.approvals.size} pending confirmation${state.approvals.size === 1 ? "" : "s"}`,
      detail: "Review tool or command approval cards in Chat before continuing.",
      panel: "chat",
    });
  }
  if (state.permissions && state.permissions.configured !== true) {
    items.push({
      tone: "warning",
      label: "Permissions",
      title: "Default guardrails",
      detail: "Create explicit permissions before high-risk workflows.",
      panel: "permissions",
    });
  } else if (state.permissions) {
    const hints = permissionsRiskHints(state.permissions);
    if (hints.length) {
      items.push({
        tone: "warning",
        label: "Permissions",
        title: `${hints.length} policy review${hints.length === 1 ? "" : "s"}`,
        detail: hints[0],
        panel: "permissions",
      });
    }
  }
  if (!["ready", "unknown"].includes(diagnosticsStatus)) {
    items.push({
      tone: diagnosticsStatus === "warning" ? "warning" : "attention",
      label: "Diagnostics",
      title: `Runtime ${statusLabel(diagnosticsStatus)}`,
      detail: state.diagnostics?.summary || "Resolve launch readiness before risky work.",
      panel: "diagnostics",
    });
  }
  if (failedRuns.length) {
    const run = failedRuns[0];
    items.push({
      tone: "attention",
      label: "Recovery",
      title: `${failedRuns.length} run${failedRuns.length === 1 ? "" : "s"} need recovery`,
      detail: run.prompt || run.id || "Open Mission Control to review failure state.",
      runID: run.id || "",
    });
  }
  if (inboxCounts.failed || inboxCounts.pending) {
    items.push({
      tone: inboxCounts.failed ? "attention" : "warning",
      label: "Inbox",
      title: inboxCounts.failed ? `${inboxCounts.failed} failed inbound` : `${inboxCounts.pending} pending inbound`,
      detail: "Approve, retry, or reject external channel work before it runs.",
      panel: "inbox",
    });
  }
  if (mcpServers.length && !enabledMCP) {
    items.push({
      tone: "warning",
      label: "MCP",
      title: "Tool docks disabled",
      detail: "Enable or test docks before tool-heavy runs.",
      panel: "mcp",
    });
  } else if (!mcpServers.length && state.config) {
    items.push({
      tone: "",
      label: "MCP",
      title: "No tool docks",
      detail: "Add a dock when approval-gated work needs external tools.",
      panel: "mcp",
    });
  }
  return items.slice(0, 6);
}

function renderApprovalCenter() {
  const target = $("approval-center-grid");
  if (!target) return;
  const items = approvalCenterItems();
  if (!items.length) {
    target.innerHTML = `<button type="button" class="approval-center-item clear" data-panel="permissions">
      <span>Clear</span>
      <strong>没有待确认风险</strong>
      <small>审批、权限、诊断、收件箱和失败运行当前没有阻塞项。</small>
    </button>`;
    return;
  }
  target.innerHTML = items.map((item) => {
    const actionAttr = item.runID
      ? `data-run-open="${escapeHTML(item.runID)}"`
      : `data-panel="${escapeHTML(item.panel || "home")}"`;
    return `<button type="button" class="approval-center-item ${escapeHTML(item.tone)}" ${actionAttr}>
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(item.title)}</strong>
      <small>${escapeHTML(item.detail)}</small>
    </button>`;
  }).join("");
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
  hideToast();
  state.panel = panel;
  document.querySelectorAll(".nav-item").forEach((button) => {
    button.classList.toggle("active", button.dataset.panel === panel);
  });
  document.querySelectorAll(".panel").forEach((section) => {
    section.classList.toggle("active", section.id === `panel-${panel}`);
  });
  $("view-title").textContent = views[panel][0];
}

function runStatusValue(run) {
  return String(run?.status || "").toLowerCase();
}

function runHealthGroup(run) {
  const status = runStatusValue(run);
  if (["running", "queued", "pending"].includes(status)) return "running";
  if (["completed", "complete", "success", "succeeded"].includes(status)) return "completed";
  if (["failed", "failure", "error", "cancelled", "canceled"].includes(status)) return "failed";
  return "unknown";
}

function renderHomeActivity() {
  const running = state.runs.filter((run) => runHealthGroup(run) === "running").length;
  const completed = state.runs.filter((run) => runHealthGroup(run) === "completed").length;
  const failed = state.runs.filter((run) => runHealthGroup(run) === "failed").length;
  const pending = state.approvals.size;
  setText("home-count-pending", pending);
  setText("home-count-running", running);
  setText("home-count-completed", completed);
  setText("home-count-failed", failed);
  setText("home-orbit-count", state.runs.length);
  renderHomeLatestRun();
  renderWorkspaceHub();
  renderKnowledgeCuration();
  renderToolDockInspector();
  renderPromptSuggestionDock();
  renderWorkflowStageRail();
  renderFocusBrief();
  renderApprovalCenter();
  renderReviewQueue();
  renderComparisonWorkbench();
  renderSourceRegistry();
  renderPromptExperimentLab();
  renderReuseGallery();
  renderProactiveDeliveryBoard();
}

function renderHomeLatestRun() {
  const target = $("home-latest-run");
  if (!target) return;
  const latest = state.runs[0];
  if (!latest) {
    target.className = "board-run";
    target.removeAttribute("data-run-open");
    target.dataset.panel = "runs";
    target.innerHTML = `<strong>暂无运行记录</strong><small>开始一个任务后，这里会显示最近一次运行。</small>`;
    return;
  }
  const status = runHealthGroup(latest);
  delete target.dataset.panel;
  target.dataset.runOpen = latest.id || "";
  target.className = `board-run ${status}`;
  target.innerHTML = `<strong>${escapeHTML(latest.prompt || latest.id || "Untitled run")}</strong>
    <small>${escapeHTML(latest.status || "unknown")} · ${escapeHTML(latest.agent || "default")} · ${escapeHTML(formatTimestamp(latest.started_at))}</small>`;
}

function renderHomeDockedTools() {
  setText("home-skill-count", state.skills.length);
  setText("home-agent-count", state.agents.length);
  setText("home-schedule-count", state.schedules.length);
  const mcpCount = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers.length : 0;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  setText("home-mcp-count", mcpCount);
  setText("home-memory-count", memoryCount);
  setText("home-council-count", state.councilRuns.length);
  setText("home-inbox-count", (inboxStatusCounts().pending || 0));
  setText("home-intake-count", state.intakeResult ? "ready" : "local");
  renderWorkspaceHub();
  renderKnowledgeCuration();
  renderToolDockInspector();
  renderPromptSuggestionDock();
  renderFocusBrief();
  renderWorkspaceHealthStrip();
  renderApprovalCenter();
  renderReviewQueue();
  renderComparisonWorkbench();
  renderSourceRegistry();
  renderPromptExperimentLab();
  renderReuseGallery();
  renderProactiveDeliveryBoard();
}

function renderWorkspaceHub() {
  const hub = $("workspace-session-hub");
  if (!hub) return;
  const latestSession = state.sessions[0];
  const running = state.runs.filter((run) => runHealthGroup(run) === "running").length;
  const failed = state.runs.filter((run) => runHealthGroup(run) === "failed").length;
  const completed = state.runs.filter((run) => runHealthGroup(run) === "completed").length;
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const memoryLabel = memoryWarnings.length
    ? `${memoryWarnings.length} warning${memoryWarnings.length === 1 ? "" : "s"}`
    : memoryFacts.length
      ? `${memoryFacts.length} facts`
      : `${memoryEntries.length} sources`;
  const runHealth = failed
    ? `${failed} need attention`
    : running
      ? `${running} active`
      : completed
        ? `${completed} completed`
        : "No runs yet";
  const intakeLabel = state.intakeResult
    ? (state.intakeResult.mode || "Result ready")
    : "Local paths";
  hub.innerHTML = [
    {
      panel: "chat",
      sessionID: latestSession?.id || "",
      kicker: "Session",
      title: latestSession?.title || latestSession?.id || "等待会话",
      detail: latestSession ? `${latestSession.msg_count ?? 0} messages · open Chat` : "开始一次对话后，这里会显示最近会话。",
      tone: latestSession ? "ready" : "",
    },
    {
      panel: "runs",
      runID: state.runs[0]?.id || "",
      kicker: "Runs",
      title: runHealth,
      detail: `${state.runs.length} total · open Mission Control`,
      tone: failed ? "attention" : running ? "active" : completed ? "ready" : "",
    },
    {
      panel: "memory",
      kicker: "Memory",
      title: memoryLabel,
      detail: memoryWarnings.length ? "Review taxonomy warnings before adding more memory." : "Open Memory Map for durable context.",
      tone: memoryWarnings.length ? "attention" : memoryFacts.length || memoryEntries.length ? "ready" : "",
    },
    {
      panel: "intake",
      kicker: "Files",
      title: intakeLabel,
      detail: state.intakeResult ? "Review the latest extracted local context." : "Open File Intake to inspect a document or archive.",
      tone: state.intakeResult ? "ready" : "",
    },
  ].map((item) => {
    const actionAttr = item.sessionID
      ? `data-session-resume="${escapeHTML(item.sessionID)}"`
      : item.runID
        ? `data-run-open="${escapeHTML(item.runID)}"`
        : `data-panel="${escapeHTML(item.panel)}"`;
    return `<button type="button" class="workspace-hub-card ${escapeHTML(item.tone)}" ${actionAttr}>
      <span>${escapeHTML(item.kicker)}</span>
      <strong>${escapeHTML(item.title)}</strong>
      <small>${escapeHTML(item.detail)}</small>
    </button>`;
  }).join("");
}

function knowledgeCurationItems() {
  const items = [];
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const favoriteSession = state.sessions.find((session) => session.favorite) || state.sessions[0];
  const completedRun = state.runs.find((run) => runHealthGroup(run) === "completed") || state.runs[0];
  if (memoryWarnings.length) {
    items.push({
      tone: "attention",
      label: "Warnings",
      title: `${memoryWarnings.length} taxonomy warning${memoryWarnings.length === 1 ? "" : "s"}`,
      detail: String(memoryWarnings[0] || "Review memory taxonomy before adding more context."),
      panel: "memory",
    });
  }
  if (memoryFacts.length) {
    const fact = memoryFacts[0];
    items.push({
      tone: "ready",
      label: "Facts",
      title: `${memoryFacts.length} classified fact${memoryFacts.length === 1 ? "" : "s"}`,
      detail: fact?.text || "Review classified project knowledge.",
      panel: "memory",
    });
  }
  if (memoryEntries.length) {
    const primary = memoryEntries.find((entry) => entry.primary) || memoryEntries[0];
    items.push({
      tone: "ready",
      label: "Sources",
      title: `${memoryEntries.length} memory source${memoryEntries.length === 1 ? "" : "s"}`,
      detail: primary?.name ? `${primary.name} · ${formatBytes(primary.size || 0)}` : "Open Memory Map to inspect sources.",
      panel: "memory",
    });
  }
  if (favoriteSession) {
    items.push({
      tone: favoriteSession.favorite ? "ready" : "",
      label: favoriteSession.favorite ? "Favorite session" : "Recent session",
      title: favoriteSession.title || favoriteSession.id || "Session",
      detail: `${favoriteSession.msg_count ?? 0} messages ready for curation.`,
      sessionID: favoriteSession.id || "",
    });
  }
  if (completedRun) {
    items.push({
      tone: runHealthGroup(completedRun) === "failed" ? "attention" : "active",
      label: runHealthGroup(completedRun) === "completed" ? "Completed run" : "Recent run",
      title: completedRun.prompt || completedRun.id || "Run",
      detail: `${completedRun.status || "unknown"} · ${completedRun.agent || "default"}`,
      runID: completedRun.id || "",
    });
  }
  if (!items.length) {
    items.push({
      tone: "clear",
      label: "Low context",
      title: "暂无知识候选",
      detail: "Run a task, favorite a session, or open Memory Capture to create reviewable context.",
      panel: "memory",
    });
  }
  return items.slice(0, 6);
}

function renderKnowledgeCuration() {
  const target = $("knowledge-curation-grid");
  if (!target) return;
  const items = knowledgeCurationItems();
  target.innerHTML = items.map((item) => {
    const actionAttr = item.sessionID
      ? `data-session-resume="${escapeHTML(item.sessionID)}"`
      : item.runID
        ? `data-run-open="${escapeHTML(item.runID)}"`
        : `data-panel="${escapeHTML(item.panel || "memory")}"`;
    return `<button type="button" class="knowledge-curation-item ${escapeHTML(item.tone || "")}" ${actionAttr}>
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(item.title)}</strong>
      <small>${escapeHTML(item.detail)}</small>
    </button>`;
  }).join("");
}

function toolDockInspectorItems() {
  const servers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  if (!servers.length) {
    return [{
      tone: "clear",
      label: "MCP",
      title: "No tool docks",
      detail: "Open MCP Starport to add stdio or HTTP tool connections.",
    }];
  }
  const enabled = servers.filter((server) => !server.disabled).length;
  const disabled = servers.length - enabled;
  const transports = servers.reduce((counts, server) => {
    const key = server.type || "stdio";
    counts[key] = (counts[key] || 0) + 1;
    return counts;
  }, {});
  const envCount = servers.reduce((total, server) => total + (Array.isArray(server.env_keys) ? server.env_keys.length : 0), 0);
  const keepAlive = servers.filter((server) => server.keep_alive).length;
  const contextual = servers.filter((server) => server.context || server.context_text).length;
  const transportDetail = Object.entries(transports)
    .map(([name, count]) => `${count} ${name}`)
    .join(" · ");
  const items = [{
    tone: disabled ? "attention" : "ready",
    label: "Docks",
    title: `${enabled}/${servers.length} enabled`,
    detail: `${transportDetail || "stdio"} · ${envCount} env key${envCount === 1 ? "" : "s"}`,
  }, {
    tone: keepAlive || contextual ? "active" : "",
    label: "Readiness",
    title: `${keepAlive} keep-alive · ${contextual} context`,
    detail: disabled ? `${disabled} disabled dock${disabled === 1 ? "" : "s"} need review.` : "Tool docks are available for agent workflows.",
  }];
  servers.slice(0, 4).forEach((server) => {
    const transport = server.type || "stdio";
    const envKeys = Array.isArray(server.env_keys) ? server.env_keys : [];
    const endpoint = transport === "http"
      ? (server.url || "missing url")
      : [server.command || "missing command"].concat(server.args || []).join(" ");
    const flags = [
      server.keep_alive ? "keep alive" : "on demand",
      server.context || server.context_text ? "context" : "no context",
      `${envKeys.length} env`,
    ].join(" · ");
    items.push({
      tone: server.disabled ? "attention" : "ready",
      label: server.disabled ? "Disabled" : transport,
      title: server.name || "Unnamed dock",
      detail: `${flags} · ${endpoint}`,
    });
  });
  return items.slice(0, 6);
}

function renderToolDockInspector() {
  const target = $("tool-dock-inspector-grid");
  if (!target) return;
  target.innerHTML = toolDockInspectorItems().map((item) => `<button type="button" class="tool-dock-item ${escapeHTML(item.tone || "")}" data-panel="mcp">
    <span>${escapeHTML(item.label)}</span>
    <strong>${escapeHTML(item.title)}</strong>
    <small>${escapeHTML(item.detail)}</small>
  </button>`).join("");
}

function promptSuggestionItems() {
  const items = [];
  const pendingApprovals = state.approvals.size;
  const failedRun = state.runs.find((run) => runHealthGroup(run) === "failed");
  const runningRun = state.runs.find((run) => runHealthGroup(run) === "running");
  const latestRun = state.runs[0];
  const latestSession = state.sessions[0];
  const pendingInbox = state.inboxItems.filter((item) => String(item.status || "pending").toLowerCase() === "pending");
  const diagnosticsStatus = state.diagnostics?.status || "";
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const mcpServers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled).length;
  const recipe = state.homeMode.startsWith("recipe:")
    ? workflowRecipes[state.homeMode.slice("recipe:".length)]
    : null;
  const strategy = workflowStrategies[state.workflowStrategy] || workflowStrategies.direct;

  if (pendingApprovals) {
    items.push({
      tone: "attention",
      label: "Approval",
      title: `${pendingApprovals} request${pendingApprovals === 1 ? "" : "s"} waiting`,
      reason: "Resolve human gates before launching more work.",
      prompt: "Review the pending approval requests. For each one, explain the risk, the requested action, whether to approve or deny, and the safest follow-up.",
    });
  }
  if (failedRun) {
    items.push({
      tone: "attention",
      label: "Recovery",
      title: failedRun.prompt || failedRun.id || "Failed run",
      reason: "Turn the failed run into a concrete repair plan.",
      prompt: `Analyze the failed run and propose the smallest safe recovery plan.\n\nRun: ${failedRun.id || "unknown"}\nPrompt: ${failedRun.prompt || ""}\nStatus: ${failedRun.status || "failed"}`,
    });
  } else if (runningRun) {
    items.push({
      tone: "active",
      label: "Monitor",
      title: runningRun.prompt || runningRun.id || "Active run",
      reason: "Check whether the current mission needs intervention.",
      prompt: `Review the active run and summarize current progress, risks, and the next operator decision.\n\nRun: ${runningRun.id || "unknown"}\nPrompt: ${runningRun.prompt || ""}`,
    });
  } else if (latestRun) {
    items.push({
      tone: "ready",
      label: "Follow-up",
      title: latestRun.prompt || latestRun.id || "Latest run",
      reason: "Use the last result as the next working context.",
      prompt: `Continue from the latest run. Summarize what was achieved, what remains uncertain, and the next concrete action.\n\nRun: ${latestRun.id || "unknown"}\nPrompt: ${latestRun.prompt || ""}\nStatus: ${latestRun.status || "unknown"}`,
    });
  }
  if (memoryWarnings.length) {
    items.push({
      tone: "attention",
      label: "Memory",
      title: `${memoryWarnings.length} memory warning${memoryWarnings.length === 1 ? "" : "s"}`,
      reason: "Clean durable context before depending on it.",
      prompt: "Review memory taxonomy warnings and produce a safe memory cleanup plan. Do not write durable memory until the changes are reviewed.",
    });
  } else if (latestSession) {
    items.push({
      tone: "ready",
      label: "Session",
      title: latestSession.title || latestSession.id || "Recent session",
      reason: "Resume from the freshest conversation context.",
      prompt: `Resume the recent session and identify the next useful task.\n\nSession: ${latestSession.id || "unknown"}\nMessages: ${latestSession.msg_count ?? 0}`,
    });
  }
  if (pendingInbox.length) {
    items.push({
      tone: "active",
      label: "Inbox",
      title: `${pendingInbox.length} inbound item${pendingInbox.length === 1 ? "" : "s"}`,
      reason: "Convert external asks into reviewed work.",
      prompt: "Triage pending inbox items. Group them into run now, needs context, and reject, then propose the next reviewed action.",
    });
  }
  if (diagnosticsStatus && !["ok", "ready", "healthy"].includes(String(diagnosticsStatus).toLowerCase())) {
    items.push({
      tone: "attention",
      label: "Readiness",
      title: "Diagnostics need review",
      reason: state.diagnostics?.summary || "Runtime readiness is not fully clear.",
      prompt: "Review daemon diagnostics and list the smallest setup or configuration fixes needed before the next mission.",
    });
  }
  if (!enabledMCP && state.config) {
    items.push({
      tone: "",
      label: "Tools",
      title: "Plan first MCP dock",
      reason: "Tool-heavy workflows need a configured dock.",
      prompt: "Suggest the first MCP dock for this workspace. Include the command or URL, required env keys, safety considerations, and a connection test plan.",
    });
  }
  if (state.intakeResult) {
    items.push({
      tone: "ready",
      label: "Files",
      title: state.intakeResult.path || "Intake result",
      reason: "Use extracted local context in the next task.",
      prompt: `Summarize this file intake result and turn it into a concrete next action.\n\nPath: ${state.intakeResult.path || ""}\nMode: ${state.intakeResult.mode || ""}\n\n${String(state.intakeResult.content || "").slice(0, 1200)}`,
    });
  }
  if (recipe) {
    items.push({
      tone: "active",
      label: "Workflow",
      title: recipe.title || "Selected workflow",
      reason: recipe.outcome || recipe.description || "Continue the selected workflow.",
      prompt: recipe.prompt || "Continue the selected Astria workflow and define the next check.",
    });
  } else {
    items.push({
      tone: "clear",
      label: "Default",
      title: `${strategy.title || "Quick Run"} next prompt`,
      reason: strategy.outcome || strategy.description || "Start from the current strategy.",
      prompt: strategy.prompt || "Continue the current Astria mission. Review the workspace state, identify the next useful action, and explain the validation needed.",
    });
  }
  return items.slice(0, 6);
}

function renderPromptSuggestionDock() {
  const target = $("prompt-suggestion-dock");
  if (!target) return;
  target.innerHTML = promptSuggestionItems().map((item) => `<button type="button" class="prompt-suggestion-item ${escapeHTML(item.tone || "")}" data-home-prompt="${escapeHTML(item.prompt || "")}">
    <span>${escapeHTML(item.label)}</span>
    <strong>${escapeHTML(item.title)}</strong>
    <small>${escapeHTML(item.reason)}</small>
  </button>`).join("");
}

function comparisonCandidates() {
  const latestRun = state.runs[0];
  const completedRuns = state.runs.filter((run) => runHealthGroup(run) === "completed");
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const latestAgent = state.agents[0];
  const latestCouncil = state.councilRuns[0];
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryCategories = new Set(memoryEntries.map((entry) => normalizeMemoryCategory(entry.category || entry.type)));
  const commandCount = state.agents.reduce((total, agent) => total + Object.keys(agent.commands || {}).length, 0);
  const roles = Array.isArray(latestCouncil?.roles) ? latestCouncil.roles : [];
  return [
    {
      id: "recent-runs",
      source: "Runs",
      panel: "runs",
      title: latestRun?.prompt || "Recent run evidence",
      metric: state.runs.length ? `${completedRuns.length}/${state.runs.length} complete` : "seed",
      evidence: [
        latestRun ? `Latest: ${latestRun.status || "unknown"} with ${latestRun.agent || "default"}` : "No latest run captured",
        failedRuns.length ? `${failedRuns.length} failed run${failedRuns.length === 1 ? "" : "s"} need review` : "No failed runs in the current list",
        latestRun ? `Started ${formatTimestamp(latestRun.started_at)}` : "Open Runs after the first execution",
      ],
      tradeoff: "Best when the next decision should follow observed execution rather than a fresh plan.",
      recommendation: failedRuns.length ? "Review failures before launching a similar run." : state.runs.length ? "Use the latest successful run as the shortest path to continue." : "Create a baseline run before choosing by execution evidence.",
      prompt: `Compare recent Astria runs and decide the next execution path.\n\nLatest run: ${latestRun?.prompt || "none"}\nStatus: ${latestRun?.status || "unknown"}\nCompleted runs: ${completedRuns.length}\nFailed runs: ${failedRuns.length}`,
    },
    {
      id: "agent-profiles",
      source: "Agents",
      panel: "agents",
      title: latestAgent?.name || "Agent profile options",
      metric: state.agents.length ? `${state.agents.length} profile${state.agents.length === 1 ? "" : "s"}` : "seed",
      evidence: [
        latestAgent ? `Lead candidate: ${latestAgent.name}` : "No lead agent selected",
        `${commandCount} saved command${commandCount === 1 ? "" : "s"}`,
        latestAgent?.model ? `Model: ${latestAgent.model}` : "Model inherits default configuration",
      ],
      tradeoff: "Best when the decision depends on role, model, tools, or saved command fit.",
      recommendation: commandCount ? "Start from the agent with the closest saved command." : "Pick a focused agent before adding more workflow state.",
      prompt: `Compare Astria agent profiles for the next task.\n\nProfiles: ${state.agents.map((agent) => agent.name).join(", ") || "none"}\nSaved commands: ${commandCount}`,
    },
    {
      id: "memory-context",
      source: "Memory",
      panel: "memory",
      title: "Durable context",
      metric: memoryEntries.length ? `${memoryEntries.length} item${memoryEntries.length === 1 ? "" : "s"}` : "seed",
      evidence: [
        `${memoryCategories.size || 0} memorized categor${memoryCategories.size === 1 ? "y" : "ies"}`,
        memoryEntries[0]?.text ? `Latest: ${String(memoryEntries[0].text).slice(0, 90)}` : "No durable memory selected",
        "Useful for preferences, decisions, risks, and architecture constraints",
      ],
      tradeoff: "Best when correctness depends on remembered project decisions instead of raw recency.",
      recommendation: memoryEntries.length ? "Use memory context to avoid repeating settled decisions." : "Capture a decision or preference before relying on memory.",
      prompt: `Compare current options against Astria memory.\n\nMemory categories: ${Array.from(memoryCategories).join(", ") || "uncategorized"}\nMemory count: ${memoryEntries.length}`,
    },
    {
      id: "council-synthesis",
      source: "Council",
      panel: "council",
      title: latestCouncil?.goal || "Council synthesis",
      metric: roles.length ? `${roles.length} role${roles.length === 1 ? "" : "s"}` : "seed",
      evidence: [
        latestCouncil ? `Goal: ${latestCouncil.goal}` : "No council selected",
        roles.length ? `Roles: ${roles.map((role) => role.role).join(", ")}` : "No role notes captured",
        latestCouncil?.synthesis ? "Synthesis is ready for handoff" : "Synthesis pending",
      ],
      tradeoff: "Best when the next step needs planner, researcher, and reviewer balance.",
      recommendation: latestCouncil?.synthesis ? "Use the council synthesis when tradeoffs matter more than speed." : "Run council before treating this as reviewed.",
      prompt: `Compare the council synthesis against other Astria options.\n\nGoal: ${latestCouncil?.goal || "none"}\nRoles: ${roles.map((role) => role.role).join(", ") || "none"}\nSynthesis:\n${latestCouncil?.synthesis || ""}`,
    },
  ];
}

function renderComparisonWorkbench() {
  const lanes = comparisonCandidates();
  setText("nav-compare-count", lanes.length);
  setText("manage-compare-count", `${lanes.length} lane${lanes.length === 1 ? "" : "s"}`);
  setText("compare-summary", `${lanes.length} comparison lane${lanes.length === 1 ? "" : "s"} from current workspace evidence.`);
  const list = $("comparison-lanes");
  if (!list) return;
  if (!state.selectedComparisonLane || !lanes.some((lane) => lane.id === state.selectedComparisonLane)) {
    state.selectedComparisonLane = lanes[0]?.id || "";
  }
  list.innerHTML = lanes.map((lane) => `<article class="comparison-lane ${lane.id === state.selectedComparisonLane ? "active" : ""}" data-compare-lane="${escapeHTML(lane.id)}">
    <div class="row-item-title"><span>${escapeHTML(lane.source)}</span><span class="tag">${escapeHTML(lane.metric)}</span></div>
    <strong>${escapeHTML(lane.title)}</strong>
    <p>${escapeHTML(lane.recommendation)}</p>
    <div class="comparison-evidence">
      ${lane.evidence.slice(0, 3).map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
    </div>
    <div class="row-actions">
      <button type="button" data-compare-select="${escapeHTML(lane.id)}">Decision brief</button>
      <button type="button" data-compare-draft="${escapeHTML(lane.id)}">Draft compare</button>
      <button type="button" data-panel="${escapeHTML(lane.panel)}">Open source</button>
    </div>
  </article>`).join("");
  renderComparisonDetail(lanes.find((lane) => lane.id === state.selectedComparisonLane) || lanes[0]);
}

function renderComparisonDetail(lane) {
  const target = $("comparison-detail");
  if (!target) return;
  if (!lane) {
    target.innerHTML = `<div class="empty-state">Select a comparison lane.</div>`;
    return;
  }
  const title = String(lane.title || "");
  const displayTitle = title.length > 140 ? `${title.slice(0, 137)}...` : title;
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(displayTitle)}</h3>
      <div class="run-meta-grid">
        <span>Source</span><strong>${escapeHTML(lane.source)}</strong>
        <span>Readiness</span><strong>${escapeHTML(lane.metric)}</strong>
        <span>Route</span><strong>${escapeHTML(lane.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Evidence</h3>
      <div class="comparison-evidence detail">
        ${lane.evidence.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Tradeoff</h3>
      <p>${escapeHTML(lane.tradeoff)}</p>
      <h3>Recommendation</h3>
      <p>${escapeHTML(lane.recommendation)}</p>
      <div class="run-detail-actions">
        <button type="button" data-compare-draft="${escapeHTML(lane.id)}">Draft compare</button>
        <button type="button" data-panel="${escapeHTML(lane.panel)}">Open source</button>
      </div>
    </section>
  </div>`;
}

function comparisonLaneByID(id) {
  return comparisonCandidates().find((lane) => lane.id === id) || null;
}

function draftComparisonToChat(id) {
  const lane = comparisonLaneByID(id);
  if (!lane) return;
  $("chat-input").value = `${lane.prompt}\n\nExplain why this path is better or worse than the other Astria options, then recommend one next action with validation.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Comparison drafted to chat.");
}

function promptLabGoal() {
  const field = $("promptlab-goal");
  return (field?.value || state.promptLabGoal || state.workflowStageLabel || state.runs[0]?.prompt || "").trim();
}

function promptLabVariants() {
  const goal = promptLabGoal() || "Define the next Astria task and validation plan.";
  const agent = state.agents[0]?.name || "default";
  const latestRun = state.runs[0];
  const latestCouncil = state.councilRuns[0];
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const compare = comparisonCandidates();
  const delivery = deliveryLanes();
  return [
    {
      id: "direct",
      label: "Direct",
      title: "Direct execution",
      panel: "chat",
      agent,
      context: "Current goal",
      source: "Chat",
      risk: "Fast path; weaker if the goal needs evidence or review first.",
      evaluation: "Success means one concrete implementation step, explicit validation, and no broad scope drift.",
      prompt: `Execute this Astria goal directly.\n\nGoal: ${goal}\n\nReturn the smallest useful implementation step, explain the validation command, and call out any risk before editing.`,
    },
    {
      id: "evidence",
      label: "Evidence",
      title: "Evidence-first experiment",
      panel: "compare",
      agent,
      context: `${state.runs.length} runs, ${memoryCount} memory items`,
      source: "Comparison Workbench",
      risk: "Slower path; best when stale context or hidden regressions are likely.",
      evaluation: "Success means the answer cites run, memory, or comparison evidence before recommending action.",
      prompt: `Run an evidence-first Astria prompt experiment.\n\nGoal: ${goal}\nLatest run: ${latestRun?.prompt || "none"}\nComparison lanes: ${compare.map((lane) => lane.source).join(", ")}\nMemory items: ${memoryCount}\n\nUse the evidence to choose one next action and explain why alternatives are weaker.`,
    },
    {
      id: "council",
      label: "Council",
      title: "Council-reviewed variant",
      panel: "council",
      agent,
      context: latestCouncil?.goal || "Planner/researcher/reviewer roles",
      source: "Agent Council",
      risk: "Adds review overhead; best when tradeoffs or coordination matter.",
      evaluation: "Success means planner, researcher, and reviewer perspectives produce a single handoff.",
      prompt: `Prepare a council-reviewed prompt variant.\n\nGoal: ${goal}\nExisting council goal: ${latestCouncil?.goal || "none"}\n\nSplit the goal into planner, researcher, and reviewer concerns, then synthesize the next concrete Astria action.`,
    },
    {
      id: "delivery",
      label: "Delivery",
      title: "Delivery-ready variant",
      panel: "delivery",
      agent,
      context: `${state.schedules.length} schedules, ${state.inboxProviders.length} channels`,
      source: "Proactive Delivery",
      risk: "Requires approval boundary before any external channel delivery.",
      evaluation: "Success means the prompt produces destination, approval gate, artifact, and verification.",
      prompt: `Create a delivery-ready Astria prompt variant.\n\nGoal: ${goal}\nDelivery lanes: ${delivery.map((lane) => lane.source).join(", ")}\n\nReturn the message shape, destination assumption, approval gate, and validation checklist.`,
    },
  ];
}

function renderPromptExperimentLab() {
  const variants = promptLabVariants();
  setText("nav-promptlab-count", variants.length);
  setText("manage-promptlab-count", `${variants.length} variant${variants.length === 1 ? "" : "s"}`);
  setText("promptlab-summary", `${variants.length} prompt variant${variants.length === 1 ? "" : "s"} across direct, evidence, council, and delivery paths.`);
  const list = $("promptlab-variants");
  if (!list) return;
  if (!state.selectedPromptVariant || !variants.some((variant) => variant.id === state.selectedPromptVariant)) {
    state.selectedPromptVariant = variants[0]?.id || "";
  }
  list.innerHTML = variants.map((variant) => `<article class="prompt-variant ${variant.id === state.selectedPromptVariant ? "active" : ""}" data-prompt-variant="${escapeHTML(variant.id)}">
    <div class="row-item-title"><span>${escapeHTML(variant.label)}</span><span class="tag">${escapeHTML(variant.agent)}</span></div>
    <strong>${escapeHTML(variant.title)}</strong>
    <p>${escapeHTML(variant.evaluation)}</p>
    <div class="prompt-variant-meta">
      <span>Context: ${escapeHTML(variant.context)}</span>
      <span>Source: ${escapeHTML(variant.source)}</span>
      <span>Risk: ${escapeHTML(variant.risk)}</span>
    </div>
    <div class="row-actions">
      <button type="button" data-prompt-variant-select="${escapeHTML(variant.id)}">Variant brief</button>
      <button type="button" data-prompt-variant-draft="${escapeHTML(variant.id)}">Draft variant</button>
      <button type="button" data-panel="${escapeHTML(variant.panel)}">Open source</button>
    </div>
  </article>`).join("");
  renderPromptVariantDetail(variants.find((variant) => variant.id === state.selectedPromptVariant) || variants[0]);
}

function renderPromptVariantDetail(variant) {
  const target = $("promptlab-detail");
  if (!target) return;
  if (!variant) {
    target.innerHTML = `<div class="empty-state">Select a prompt variant.</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(variant.title)}</h3>
      <div class="run-meta-grid">
        <span>Agent</span><strong>${escapeHTML(variant.agent)}</strong>
        <span>Context</span><strong>${escapeHTML(variant.context)}</strong>
        <span>Source</span><strong>${escapeHTML(variant.source)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Evaluation</h3>
      <p>${escapeHTML(variant.evaluation)}</p>
      <h3>Risk</h3>
      <p>${escapeHTML(variant.risk)}</p>
      <div class="run-detail-actions">
        <button type="button" data-prompt-variant-draft="${escapeHTML(variant.id)}">Draft variant</button>
        <button type="button" data-panel="${escapeHTML(variant.panel)}">Open source</button>
      </div>
    </section>
  </div>`;
}

function promptVariantByID(id) {
  return promptLabVariants().find((variant) => variant.id === id) || null;
}

function draftPromptVariantToChat(id) {
  const variant = promptVariantByID(id);
  if (!variant) return;
  $("chat-input").value = variant.prompt;
  $("chat-agent").value = variant.agent === "default" ? "" : variant.agent;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Prompt variant drafted to chat.");
}

function reuseGalleryAssets() {
  const assets = [];
  const variants = promptLabVariants();
  const sources = sourceRegistryRows();
  const commandAssets = state.agents.flatMap((agent) => {
    const commands = agent.Commands || agent.commands || {};
    return Object.entries(commands).slice(0, 2).map(([name, body]) => ({
      agent,
      name,
      body,
    }));
  });
  const latestRun = state.runs[0];
  const completedRun = state.runs.find((run) => runHealthGroup(run) === "completed") || latestRun;
  const latestCouncil = state.councilRuns[0];

  variants.slice(0, 2).forEach((variant) => {
    assets.push({
      id: `prompt-${variant.id}`,
      kind: "Prompt",
      title: variant.title,
      panel: "promptlab",
      readiness: variant.context,
      evidence: variant.source,
      reuse: variant.evaluation,
      action: "Start a new mission from this prompt shape.",
      prompt: `Reuse this Astria prompt asset as the starting point for a new mission.\n\nAsset: ${variant.title}\nSource: ${variant.source}\nAgent: ${variant.agent}\nEvaluation: ${variant.evaluation}\n\nPrompt:\n${variant.prompt}`,
    });
  });

  if (state.agents[0]) {
    const agent = state.agents[0];
    const summary = agentCapabilitySummary(agent);
    assets.push({
      id: `agent-${summary.name}`,
      kind: "Agent",
      title: summary.name,
      panel: "agents",
      readiness: `${summary.allow.length} tools, ${summary.deny.length} blocked`,
      evidence: summary.model,
      reuse: summary.description,
      action: "Launch with this profile and carry its operating constraints forward.",
      prompt: `Reuse this Astria agent profile for the next mission.\n\nAgent: ${summary.name}\nModel: ${summary.model}\nReasoning: ${summary.reasoning}\nTools allowed: ${summary.allow.join(", ") || "default"}\nAuto approve: ${summary.autoApprove}\n\nDescribe the mission, choose whether this profile fits, and call out any safety constraint before acting.`,
    });
  } else {
    assets.push({
      id: "agent-default",
      kind: "Agent",
      title: "Default agent starter",
      panel: "agents",
      readiness: "default",
      evidence: "no named profile",
      reuse: "Use the default daemon agent until a named profile exists.",
      action: "Draft a focused role before creating a reusable profile.",
      prompt: "Create a reusable Astria agent profile for this workspace. Include role, model expectations, tool boundaries, memory needs, and one saved command.",
    });
  }

  commandAssets.slice(0, 2).forEach((command) => {
    const agentName = normalizeName(command.agent);
    assets.push({
      id: `command-${agentName}-${command.name}`,
      kind: "Command",
      title: `/${command.name}`,
      panel: "agents",
      readiness: agentName,
      evidence: "saved command",
      reuse: String(command.body || "").slice(0, 150) || "Saved command body",
      action: "Draft this saved command into Chat with its agent profile.",
      prompt: `Reuse this saved Astria command.\n\nAgent: ${agentName}\nCommand: /${command.name}\n\n${String(command.body || "")}`,
      agent: agentName,
    });
  });

  sources.slice(0, 2).forEach((source) => {
    assets.push({
      id: `source-${source.id}`,
      kind: "Knowledge",
      title: source.title,
      panel: source.panel,
      readiness: source.reliability,
      evidence: `${source.evidence} evidence`,
      reuse: source.action,
      action: "Ground the next mission in this source before launching.",
      prompt: `Reuse this Astria knowledge source as mission context.\n\nSource: ${source.title}\nType: ${source.type}\nFreshness: ${source.freshness}\nReliability: ${source.reliability}\nEvidence: ${source.evidence}\n\nDraft the next task using only what this source can reliably support.`,
    });
  });

  if (completedRun) {
    assets.push({
      id: `run-${completedRun.id || "latest"}`,
      kind: "Outcome",
      title: completedRun.prompt || completedRun.id || "Latest run outcome",
      panel: "runs",
      readiness: completedRun.status || "unknown",
      evidence: completedRun.agent || "default",
      reuse: "Continue from a concrete execution result instead of restarting from scratch.",
      action: "Turn this run into the next mission starter.",
      prompt: `Reuse this Astria run outcome as the next starting point.\n\nRun: ${completedRun.id || "unknown"}\nStatus: ${completedRun.status || "unknown"}\nAgent: ${completedRun.agent || "default"}\nPrompt: ${completedRun.prompt || ""}\n\nSummarize what can be reused, what remains uncertain, and the next concrete action with validation.`,
    });
  }

  if (latestCouncil) {
    assets.push({
      id: `council-${latestCouncil.id || "latest"}`,
      kind: "Council",
      title: latestCouncil.goal || "Council synthesis",
      panel: "council",
      readiness: latestCouncil.synthesis ? "synthesized" : "review",
      evidence: `${(latestCouncil.roles || []).length} roles`,
      reuse: latestCouncil.synthesis || "Planner, researcher, and reviewer context can seed the next handoff.",
      action: "Reuse the reviewed synthesis as a mission brief.",
      prompt: `Reuse this Astria council result as the next mission brief.\n\nGoal: ${latestCouncil.goal || "none"}\nRoles: ${(latestCouncil.roles || []).map((role) => role.role).join(", ") || "none"}\nSynthesis:\n${latestCouncil.synthesis || ""}\n\nTurn the synthesis into one executable next step and validation plan.`,
    });
  } else {
    assets.push({
      id: "council-starter",
      kind: "Council",
      title: "Council review starter",
      panel: "council",
      readiness: "seed",
      evidence: "planner/researcher/reviewer",
      reuse: "Use multi-role review when a reusable decision needs stronger evidence.",
      action: "Draft a council-ready mission brief.",
      prompt: "Create a reusable council mission brief. Split the work into planner, researcher, and reviewer concerns, then define the synthesis criteria.",
    });
  }

  return assets.slice(0, 8);
}

function renderReuseGallery() {
  const assets = reuseGalleryAssets();
  setText("nav-reuse-count", assets.length);
  setText("manage-reuse-count", `${assets.length} asset${assets.length === 1 ? "" : "s"}`);
  setText("reuse-summary", `${assets.length} reusable asset${assets.length === 1 ? "" : "s"} from prompts, agents, knowledge, outcomes, and council review.`);
  const list = $("reuse-gallery-assets");
  if (!list) return;
  if (!state.selectedReuseAsset || !assets.some((asset) => asset.id === state.selectedReuseAsset)) {
    state.selectedReuseAsset = assets[0]?.id || "";
  }
  list.innerHTML = assets.map((asset) => `<article class="reuse-asset ${asset.id === state.selectedReuseAsset ? "active" : ""}" data-reuse-asset="${escapeHTML(asset.id)}">
    <div class="row-item-title"><span>${escapeHTML(asset.kind)}</span><span class="tag">${escapeHTML(asset.readiness)}</span></div>
    <strong>${escapeHTML(asset.title)}</strong>
    <div class="reuse-grid">
      <span>Evidence</span><strong>${escapeHTML(asset.evidence)}</strong>
      <span>Reuse value</span><strong>${escapeHTML(asset.reuse)}</strong>
      <span>Next action</span><strong>${escapeHTML(asset.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-reuse-select="${escapeHTML(asset.id)}">Asset brief</button>
      <button type="button" data-reuse-draft="${escapeHTML(asset.id)}">Draft starter</button>
      <button type="button" data-panel="${escapeHTML(asset.panel)}">Open source</button>
    </div>
  </article>`).join("");
  renderReuseAssetDetail(assets.find((asset) => asset.id === state.selectedReuseAsset) || assets[0]);
}

function renderReuseAssetDetail(asset) {
  const target = $("reuse-gallery-detail");
  if (!target) return;
  if (!asset) {
    target.innerHTML = `<div class="empty-state">Select a reusable asset.</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(asset.title)}</h3>
      <div class="run-meta-grid">
        <span>Kind</span><strong>${escapeHTML(asset.kind)}</strong>
        <span>Readiness</span><strong>${escapeHTML(asset.readiness)}</strong>
        <span>Route</span><strong>${escapeHTML(asset.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Reuse value</h3>
      <p>${escapeHTML(asset.reuse)}</p>
      <h3>Next action</h3>
      <p>${escapeHTML(asset.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-reuse-draft="${escapeHTML(asset.id)}">Draft starter</button>
        <button type="button" data-panel="${escapeHTML(asset.panel)}">Open source</button>
      </div>
    </section>
  </div>`;
}

function reuseAssetByID(id) {
  return reuseGalleryAssets().find((asset) => asset.id === id) || null;
}

function draftReuseAssetToChat(id) {
  const asset = reuseAssetByID(id);
  if (!asset) return;
  $("chat-input").value = `${asset.prompt}\n\nReturn a reusable mission starter with objective, context to carry forward, launch path, and validation.`;
  if (asset.agent) $("chat-agent").value = asset.agent;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Reusable starter drafted to chat.");
}

function deliveryLanes() {
  const enabledSchedules = state.schedules.filter((schedule) => schedule.enabled !== false);
  const scheduledRuns = state.runs.filter((run) => String(run.channel || "").includes("schedule") || String(run.source || "").includes("schedule"));
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const providers = Array.isArray(state.inboxProviders) ? state.inboxProviders : [];
  const pendingInbox = state.inboxItems.filter((item) => item.status === "pending");
  const diagnosticsStatus = String(state.diagnostics?.status || "unknown").toLowerCase();
  const readyDiagnostics = ["ok", "ready", "healthy"].includes(diagnosticsStatus);
  return [
    {
      id: "scheduled-work",
      source: "Schedules",
      panel: "schedules",
      title: "Scheduled work",
      metric: `${enabledSchedules.length}/${state.schedules.length} active`,
      evidence: [
        enabledSchedules[0] ? `Next active prompt: ${enabledSchedules[0].prompt || "Untitled schedule"}` : "No active schedule configured",
        state.schedules.length ? `${state.schedules.length} configured schedule${state.schedules.length === 1 ? "" : "s"}` : "Create a cron plan before expecting proactive work",
        enabledSchedules[0]?.cron ? `Cron: ${enabledSchedules[0].cron}` : "Cron cadence not set",
      ],
      risk: enabledSchedules.length ? "Scheduled prompts still need clear delivery or review expectations." : "No proactive work will run until a schedule exists.",
      action: enabledSchedules.length ? "Review active cadence and delivery target." : "Draft the first scheduled delivery plan.",
      prompt: `Plan proactive Astria delivery from schedules.\n\nActive schedules: ${enabledSchedules.length}\nConfigured schedules: ${state.schedules.length}\nFirst prompt: ${enabledSchedules[0]?.prompt || "none"}\n\nDefine cadence, expected output, destination channel, and validation.`,
    },
    {
      id: "delivery-runs",
      source: "Runs",
      panel: "runs",
      title: "Recent outbound runs",
      metric: `${scheduledRuns.length} scheduled`,
      evidence: [
        scheduledRuns[0] ? `Latest scheduled run: ${scheduledRuns[0].status || "unknown"}` : "No scheduled run history captured",
        failedRuns.length ? `${failedRuns.length} failed run${failedRuns.length === 1 ? "" : "s"} need recovery` : "No failed runs in the current list",
        state.runs[0] ? `Latest run: ${state.runs[0].prompt || state.runs[0].id}` : "Run history is empty",
      ],
      risk: failedRuns.length ? "Delivery confidence is low until failures are triaged." : "Recent runs should still be reviewed before external delivery.",
      action: failedRuns.length ? "Draft recovery notes for failed outbound work." : "Draft an outbound summary from recent runs.",
      prompt: `Review proactive delivery run history.\n\nScheduled runs: ${scheduledRuns.length}\nFailed runs: ${failedRuns.length}\nLatest run: ${state.runs[0]?.prompt || "none"}\n\nDecide what is safe to deliver and what needs retry or review.`,
    },
    {
      id: "channel-readiness",
      source: "Channels",
      panel: "inbox",
      title: "Channel readiness",
      metric: `${providers.length} provider${providers.length === 1 ? "" : "s"}`,
      evidence: [
        providers.length ? `Providers: ${providers.map((provider) => provider.name || provider.id).filter(Boolean).join(", ")}` : "No channel provider listed",
        pendingInbox.length ? `${pendingInbox.length} inbound item${pendingInbox.length === 1 ? "" : "s"} waiting` : "No pending inbound items",
        "Outbound delivery should mirror reviewed inbound channel policy",
      ],
      risk: providers.length ? "Channel state is visible, but outbound delivery still needs explicit approval." : "No visible channel means proactive output stays local.",
      action: providers.length ? "Draft channel-specific delivery copy." : "Draft channel setup requirements.",
      prompt: `Prepare proactive Astria channel delivery.\n\nProviders: ${providers.map((provider) => provider.name || provider.id).filter(Boolean).join(", ") || "none"}\nPending inbox items: ${pendingInbox.length}\n\nWrite the delivery target, approval gate, message shape, and rollback path.`,
    },
    {
      id: "delivery-recovery",
      source: "Readiness",
      panel: readyDiagnostics ? "delivery" : "diagnostics",
      title: "Recovery and guardrails",
      metric: readyDiagnostics ? "ready" : "review",
      evidence: [
        `Diagnostics: ${state.diagnostics?.status || "unknown"}`,
        state.diagnostics?.summary || "Diagnostics summary unavailable",
        state.permissions?.configured === true ? "Permissions configured" : "Using default permissions",
      ],
      risk: readyDiagnostics ? "Ready state still needs an approval boundary before external posting." : "Runtime readiness may block reliable scheduled delivery.",
      action: readyDiagnostics ? "Draft the approval checklist." : "Open diagnostics and repair blockers.",
      prompt: `Create a proactive delivery recovery checklist.\n\nDiagnostics: ${state.diagnostics?.status || "unknown"}\nSummary: ${state.diagnostics?.summary || ""}\nPermissions configured: ${state.permissions?.configured === true}\n\nList blockers, approval gates, retry rules, and verification.`,
    },
  ];
}

function renderProactiveDeliveryBoard() {
  const lanes = deliveryLanes();
  setText("nav-delivery-count", lanes.length);
  setText("manage-delivery-count", `${lanes.length} lane${lanes.length === 1 ? "" : "s"}`);
  setText("delivery-summary", `${lanes.length} proactive delivery lane${lanes.length === 1 ? "" : "s"} across schedules, runs, channels, and readiness.`);
  const list = $("delivery-lanes");
  if (!list) return;
  if (!state.selectedDeliveryLane || !lanes.some((lane) => lane.id === state.selectedDeliveryLane)) {
    state.selectedDeliveryLane = lanes[0]?.id || "";
  }
  list.innerHTML = lanes.map((lane) => `<article class="delivery-lane ${lane.id === state.selectedDeliveryLane ? "active" : ""}" data-delivery-lane="${escapeHTML(lane.id)}">
    <div class="row-item-title"><span>${escapeHTML(lane.source)}</span><span class="tag">${escapeHTML(lane.metric)}</span></div>
    <strong>${escapeHTML(lane.title)}</strong>
    <p>${escapeHTML(lane.action)}</p>
    <div class="delivery-evidence">
      ${lane.evidence.slice(0, 3).map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
    </div>
    <div class="row-actions">
      <button type="button" data-delivery-select="${escapeHTML(lane.id)}">Delivery brief</button>
      <button type="button" data-delivery-draft="${escapeHTML(lane.id)}">Draft delivery</button>
      <button type="button" data-panel="${escapeHTML(lane.panel)}">Open source</button>
    </div>
  </article>`).join("");
  renderDeliveryDetail(lanes.find((lane) => lane.id === state.selectedDeliveryLane) || lanes[0]);
}

function renderDeliveryDetail(lane) {
  const target = $("delivery-detail");
  if (!target) return;
  if (!lane) {
    target.innerHTML = `<div class="empty-state">Select a delivery lane.</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(lane.title)}</h3>
      <div class="run-meta-grid">
        <span>Source</span><strong>${escapeHTML(lane.source)}</strong>
        <span>Readiness</span><strong>${escapeHTML(lane.metric)}</strong>
        <span>Route</span><strong>${escapeHTML(lane.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Evidence</h3>
      <div class="delivery-evidence detail">
        ${lane.evidence.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Risk</h3>
      <p>${escapeHTML(lane.risk)}</p>
      <h3>Next action</h3>
      <p>${escapeHTML(lane.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-delivery-draft="${escapeHTML(lane.id)}">Draft delivery</button>
        <button type="button" data-panel="${escapeHTML(lane.panel)}">Open source</button>
      </div>
    </section>
  </div>`;
}

function deliveryLaneByID(id) {
  return deliveryLanes().find((lane) => lane.id === id) || null;
}

function draftDeliveryToChat(id) {
  const lane = deliveryLaneByID(id);
  if (!lane) return;
  $("chat-input").value = `${lane.prompt}\n\nReturn a concise delivery brief with destination, approval gate, expected artifact, and verification.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Delivery brief drafted to chat.");
}

function renderManageCount() {
  const mcpCount = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers.length : 0;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const sourceCount = sourceRegistryRows().length;
  const compareCount = comparisonCandidates().length;
  const promptVariantCount = promptLabVariants().length;
  const reuseCount = reuseGalleryAssets().length;
  const deliveryCount = deliveryLanes().length;
  setText("manage-intake-count", state.intakeResult ? "Result ready" : "Local paths");
  setText("manage-sources-count", `${sourceCount} source${sourceCount === 1 ? "" : "s"}`);
  setText("nav-sources-count", sourceCount);
  setText("manage-compare-count", `${compareCount} lane${compareCount === 1 ? "" : "s"}`);
  setText("nav-compare-count", compareCount);
  setText("manage-promptlab-count", `${promptVariantCount} variant${promptVariantCount === 1 ? "" : "s"}`);
  setText("nav-promptlab-count", promptVariantCount);
  setText("manage-reuse-count", `${reuseCount} asset${reuseCount === 1 ? "" : "s"}`);
  setText("nav-reuse-count", reuseCount);
  setText("manage-delivery-count", `${deliveryCount} lane${deliveryCount === 1 ? "" : "s"}`);
  setText("nav-delivery-count", deliveryCount);
  setText("manage-count", state.agents.length + state.skills.length + state.schedules.length + mcpCount + memoryCount + sourceCount + state.councilRuns.length + state.inboxItems.length + compareCount + promptVariantCount + reuseCount + deliveryCount + 1);
}

function renderMCPStarport() {
  const servers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  setText("nav-mcp-count", servers.length);
  setText("manage-mcp-count", `${servers.length} ${servers.length === 1 ? "dock" : "docks"}`);
  renderManageCount();
  setText("mcp-summary", servers.length ? `${servers.length} configured MCP server${servers.length === 1 ? "" : "s"}.` : "No MCP servers configured.");
  const enabled = servers.filter((server) => !server.disabled).length;
  const overview = $("mcp-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(enabled ? `${enabled} enabled` : "No docks")}</strong><span>${escapeHTML(servers.length ? "Edit, disable, or test configured MCP docks from Astria." : "Add a stdio dock from Astria, then test the connection.")}</span>`;
  }
  renderMCPForm();
  const list = $("mcp-list");
  if (!list) return;
  if (!servers.length) {
    renderEmptyAction(list, "No MCP servers configured. Add a stdio dock or ask Astria to suggest a first connection.", [
      { label: "Add dock", action: "mcp-new", primary: true },
      { label: "Ask Astria", homeAction: "mcp" },
    ]);
    return;
  }
  list.innerHTML = servers.map((server) => {
    const transport = server.type || "stdio";
    const endpoint = transport === "http" ? (server.url || "missing url") : [server.command || "missing command"].concat(server.args || []).join(" ");
    const envKeys = Array.isArray(server.env_keys) ? server.env_keys : [];
    return `<article class="row-item mcp-server-card ${server.disabled ? "disabled" : "enabled"}">
      <div class="row-item-title">
        <span>${escapeHTML(server.name || "Unnamed server")}</span>
        <span class="tag">${escapeHTML(server.disabled ? "disabled" : transport)}</span>
      </div>
      <p>${escapeHTML(endpoint)}</p>
      <div class="pill-list">
        <span>${server.keep_alive ? "keep alive" : "on demand"}</span>
        <span>${server.context ? "context" : "no context"}</span>
        <span>${envKeys.length} env keys</span>
      </div>
      ${envKeys.length ? `<p class="secret-note">Env values hidden: ${envKeys.map(escapeHTML).join(", ")}</p>` : ""}
      <div class="row-actions">
        <button type="button" data-mcp-edit="${escapeHTML(server.name || "")}">Edit</button>
        <button type="button" data-mcp-toggle="${escapeHTML(server.name || "")}">${server.disabled ? "Enable" : "Disable"}</button>
        <button type="button" data-mcp-test="${escapeHTML(server.name || "")}" ${server.disabled ? "disabled" : ""}>Test connection</button>
      </div>
      <div class="mcp-test-result" id="mcp-test-${escapeHTML(server.name || "")}"></div>
    </article>`;
  }).join("");
}

function renderMCPForm() {
  const form = $("mcp-form");
  if (!form) return;
  const server = getMCPServer(state.editingMCPServer) || null;
  $("mcp-name").value = server?.name || "";
  $("mcp-name").disabled = Boolean(server?.name);
  $("mcp-type").value = server?.type || "stdio";
  $("mcp-command").value = server?.command || "";
  $("mcp-args").value = Array.isArray(server?.args) ? server.args.join("\n") : "";
  $("mcp-url").value = server?.url || "";
  $("mcp-context").value = server?.context_text || "";
  $("mcp-env").value = Array.isArray(server?.env_keys) ? server.env_keys.map((key) => `${key}=`).join("\n") : "";
  $("mcp-keep-alive").checked = Boolean(server?.keep_alive);
  $("mcp-disabled").checked = Boolean(server?.disabled);
  setText("mcp-save-state", server ? `Editing ${server.name}` : "Ready");
  updateMCPTransportFields();
}

function getMCPServer(name) {
  const servers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  return servers.find((server) => server.name === name);
}

function beginMCPCreate() {
  state.editingMCPServer = "";
  renderMCPForm();
  $("mcp-name")?.focus();
}

function editMCPServer(name) {
  state.editingMCPServer = name || "";
  renderMCPForm();
  $("mcp-command")?.focus();
}

function updateMCPTransportFields() {
  const type = $("mcp-type")?.value || "stdio";
  for (const field of document.querySelectorAll(".mcp-stdio-field")) {
    field.hidden = type !== "stdio";
  }
  for (const field of document.querySelectorAll(".mcp-http-field")) {
    field.hidden = type !== "http";
  }
}

function buildMCPPatchServers(replacement) {
  const servers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const next = [];
  let replaced = false;
  for (const server of servers) {
    if (server.name === replacement.name) {
      next.push(replacement);
      replaced = true;
    } else {
      next.push(mcpViewToPatch(server));
    }
  }
  if (!replaced) next.push(replacement);
  return next;
}

function mcpViewToPatch(server) {
  const env = {};
  for (const key of Array.isArray(server.env_keys) ? server.env_keys : []) {
    env[key] = "";
  }
  return {
    name: server.name || "",
    type: server.type || "stdio",
    command: server.command || "",
    args: Array.isArray(server.args) ? server.args : [],
    url: server.url || "",
    env,
    disabled: Boolean(server.disabled),
    keep_alive: Boolean(server.keep_alive),
    context: server.context_text || "",
  };
}

function buildMCPFormPatch() {
  const name = $("mcp-name").value.trim();
  const type = $("mcp-type").value || "stdio";
  const patch = {
    name,
    type,
    command: $("mcp-command").value.trim(),
    args: parseCSVList($("mcp-args").value),
    url: $("mcp-url").value.trim(),
    env: parseMCPEnv($("mcp-env").value),
    context: $("mcp-context").value.trim(),
    keep_alive: $("mcp-keep-alive").checked,
    disabled: $("mcp-disabled").checked,
  };
  return patch;
}

function parseMCPEnv(value) {
  const env = {};
  for (const rawLine of String(value || "").split("\n")) {
    const line = rawLine.trim();
    if (!line) continue;
    const idx = line.indexOf("=");
    if (idx < 0) {
      env[line] = "";
    } else {
      const key = line.slice(0, idx).trim();
      if (!key) continue;
      env[key] = line.slice(idx + 1);
    }
  }
  return env;
}

async function submitMCPServer(event) {
  event.preventDefault();
  $("mcp-save-state").textContent = "Saving";
  try {
    const replacement = buildMCPFormPatch();
    const result = await api("/config", {
      method: "PATCH",
      body: JSON.stringify({ mcp_servers: buildMCPPatchServers(replacement) }),
    });
    state.config = result.config || state.config;
    state.editingMCPServer = replacement.name;
    renderMCPStarport();
    renderHomeDockedTools();
    showToast("MCP dock saved.");
  } catch (error) {
    $("mcp-save-state").textContent = "Error";
    showToast(error.message);
  }
}

async function toggleMCPServer(name) {
  const server = getMCPServer(name);
  if (!server) return;
  const replacement = mcpViewToPatch(server);
  replacement.disabled = !server.disabled;
  $("mcp-save-state").textContent = "Saving";
  try {
    const result = await api("/config", {
      method: "PATCH",
      body: JSON.stringify({ mcp_servers: buildMCPPatchServers(replacement) }),
    });
    state.config = result.config || state.config;
    renderMCPStarport();
    renderHomeDockedTools();
    showToast(replacement.disabled ? "MCP dock disabled." : "MCP dock enabled.");
  } catch (error) {
    $("mcp-save-state").textContent = "Error";
    showToast(error.message);
  }
}

async function testMCPServer(name) {
  if (!name) return;
  const target = $(`mcp-test-${name}`);
  if (target) target.innerHTML = `<div class="inline-state">Testing connection...</div>`;
  try {
    const result = await api("/mcp/test", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
    renderMCPTestResult(name, result);
  } catch (error) {
    renderMCPTestResult(name, { status: "error", error: error.message, tools: [] });
  }
}

function renderMCPTestResult(name, result) {
  const target = $(`mcp-test-${name}`);
  if (!target) return;
  const status = result?.status || "unknown";
  const tools = Array.isArray(result?.tools) ? result.tools : [];
  const preview = tools.slice(0, 6).map((tool) => `<span>${escapeHTML(tool.name || "tool")}</span>`).join("");
  target.innerHTML = `<div class="mcp-test-card ${escapeHTML(status)}">
    <strong>${escapeHTML(status === "ok" ? `${result.tool_count || tools.length} tools discovered` : status)}</strong>
    <p>${escapeHTML(result?.error || (status === "ok" ? "Connection test succeeded." : "Connection test finished."))}</p>
    ${preview ? `<div class="pill-list">${preview}</div>` : ""}
  </div>`;
}

function renderFileIntake() {
  const result = state.intakeResult;
  setText("intake-summary", result ? `${result.mode === "archive_inspect" ? "Archive inspected" : "Document text extracted"} from ${result.path || "local path"}.` : "Inspect local documents and archives before sending them into a run.");
  const overview = $("intake-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(result ? (result.is_error ? "Needs attention" : "Ready") : "Local")}</strong><span>${escapeHTML(result ? (result.is_error ? "Fix the path or mode, then analyze again." : "Result can be sent into a normal chat/run workflow.") : "Read-only intake runs before extraction or summarization.")}</span>`;
  }
  const target = $("intake-result");
  if (!target) return;
  $("intake-chat-button").disabled = !result || result.is_error;
  $("intake-extract-button").disabled = !result || result.mode !== "archive_inspect" || result.is_error;
  renderHomeDockedTools();
  renderManageCount();
  if (!result) {
    renderEmptyAction(target, "Choose a local path to inspect with document_text or archive_inspect.", [
      { label: "Open chat", panel: "chat" },
    ]);
    return;
  }
  const status = result.is_error ? "error" : "ok";
  const preview = String(result.content || "").slice(0, 12000);
  target.innerHTML = `<article class="intake-result-card ${escapeHTML(status)}">
    <div class="row-item-title">
      <span>${escapeHTML(result.path || "Local file")}</span>
      <span class="tag">${escapeHTML(result.mode || "intake")}</span>
    </div>
    <pre>${escapeHTML(preview || "No content returned.")}</pre>
  </article>`;
}

async function submitFileIntake(event) {
  event?.preventDefault?.();
  const path = $("intake-path").value.trim();
  if (!path) {
    showToast("File path is required.");
    return;
  }
  $("intake-state").textContent = "Analyzing";
  try {
    const result = await api("/intake/file", {
      method: "POST",
      body: JSON.stringify({
        path,
        mode: $("intake-mode").value || "auto",
        max_chars: Number($("intake-max-chars").value || 0),
        max_entries: Number($("intake-max-entries").value || 0),
      }),
    });
    state.intakeResult = result;
    $("intake-state").textContent = result.is_error ? "Error" : "Ready";
    renderFileIntake();
    renderSourceRegistry();
    showToast(result.is_error ? "File intake returned an error." : "File intake ready.");
  } catch (error) {
    $("intake-state").textContent = "Error";
    showToast(error.message);
  }
}

function sendIntakeToChat() {
  const result = state.intakeResult;
  if (!result || result.is_error) return;
  $("chat-input").value = `Summarize this local file intake result and identify useful next actions.\n\nPath: ${result.path}\nMode: ${result.mode}\n\n${String(result.content || "").slice(0, 8000)}`;
  $("chat-new-session").checked = true;
  switchPanel("chat");
  $("chat-input").focus();
  showToast("File intake copied into Chat.");
}

function draftArchiveExtractRun() {
  const result = state.intakeResult;
  if (!result || result.mode !== "archive_inspect" || result.is_error) return;
  $("chat-input").value = `Inspect this archive result and, if extraction is appropriate, call archive_extract with approval. Ask before choosing a destination if it is not obvious.\n\nArchive path: ${result.path}\n\nArchive inspection:\n${String(result.content || "").slice(0, 8000)}`;
  $("chat-new-session").checked = true;
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Archive extraction prompt drafted.");
}

function renderMemoryMapPreview() {
  const list = $("memory-list");
  if (!list) return;
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const favoriteSessions = state.sessions.filter((session) => session.favorite);
  const recentRuns = state.runs.slice(0, 3);
  const count = memoryEntries.length + favoriteSessions.length + recentRuns.length + memoryFacts.length;
  setText("nav-memory-count", count);
  setText("manage-memory-count", `${count} ${count === 1 ? "source" : "sources"}`);
  renderManageCount();
  setText("memory-summary", count ? `${memoryFacts.length} classified fact${memoryFacts.length === 1 ? "" : "s"} · ${memoryWarnings.length} warning${memoryWarnings.length === 1 ? "" : "s"}` : "No memory candidates yet.");
  renderWorkspaceHealthStrip();
  renderKnowledgeCuration();
  renderPromptSuggestionDock();
  renderApprovalCenter();
  renderReviewQueue();
  const overview = $("memory-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(memoryFacts.length ? `${memoryFacts.length} facts` : memoryEntries.length ? `${memoryEntries.length} memory files` : count ? "Sources ready" : "Preview")}</strong><span>${escapeHTML(memoryWarnings.length ? `${memoryWarnings.length} taxonomy warning${memoryWarnings.length === 1 ? "" : "s"} need review before adding more memory.` : state.memory?.memory_dir || (count ? "Draft reviewable memory from recent work before writing MEMORY.md." : "Favorite sessions or complete runs to create stronger memory candidates."))}</span>`;
  }
  renderMemoryTaxonomyBar(state.memory?.categories || {});
  renderMemoryWarnings(memoryWarnings);
  const cards = [];
  const selectedCategory = state.memoryCategory || "all";
  const filteredFacts = selectedCategory === "all" ? memoryFacts : memoryFacts.filter((fact) => fact.category === selectedCategory);
  for (const fact of filteredFacts) {
    cards.push(`<article class="row-item memory-fact-card ${escapeHTML(fact.category || "uncategorized")}">
      <div class="row-item-title"><span>${escapeHTML(fact.text)}</span><span class="tag">${escapeHTML(fact.category || "uncategorized")}</span></div>
      <p>${escapeHTML(fact.entry || "MEMORY.md")} · line ${escapeHTML(fact.line || "-")}${fact.subject ? ` · ${escapeHTML(fact.subject)}` : ""}</p>
    </article>`);
  }
  for (const entry of memoryEntries) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card ${entry.primary ? "primary" : ""}">
      <div class="row-item-title"><span>${escapeHTML(entry.name)}</span><span class="tag">${entry.primary ? "active memory" : "memory file"}</span></div>
      <p>${escapeHTML(formatBytes(entry.size || 0))} · ${escapeHTML(formatTimestamp(entry.modified))}</p>
      <div class="row-actions">
        <button type="button" class="danger-button" data-memory-delete="${escapeHTML(entry.name)}">Delete</button>
      </div>
    </article>`);
  }
  for (const session of favoriteSessions.slice(0, 4)) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card">
      <div class="row-item-title"><span>${escapeHTML(session.title || session.id)}</span><span class="tag">favorite session</span></div>
      <p>${escapeHTML(session.id)}</p>
      <div class="row-actions">
        <button type="button" data-session-id="${escapeHTML(session.id)}">Open session</button>
        <button type="button" data-memory-draft="session:${escapeHTML(session.id)}">Draft memory</button>
      </div>
    </article>`);
  }
  for (const run of recentRuns) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card">
      <div class="row-item-title"><span>${escapeHTML(run.prompt || run.id)}</span><span class="tag">recent run</span></div>
      <p>${escapeHTML(run.status || "unknown")} · ${escapeHTML(formatTimestamp(run.started_at))}</p>
      <div class="row-actions">
        <button type="button" data-run-open="${escapeHTML(run.id)}">Open run</button>
        <button type="button" data-memory-draft="run:${escapeHTML(run.id)}">Draft memory</button>
      </div>
    </article>`);
  }
  if (!cards.length) {
    renderEmptyAction(list, "No memory sources yet. Complete a run, favorite a session, or ask Astria to draft a memory map.", [
      { label: "Draft memory map", homeAction: "memory", primary: true },
      { label: "Open chat", panel: "chat" },
    ]);
    return;
  }
  list.innerHTML = cards.join("");
}

function sourceRegistryRows() {
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const favoriteSessions = state.sessions.filter((session) => session.favorite);
  const latestRun = state.runs[0];
  const latestCouncil = state.councilRuns[0];
  const intake = state.intakeResult;
  return [
    {
      id: "memory",
      type: "Memory",
      title: "Reviewed memory",
      panel: "memory",
      evidence: memoryEntries.length + memoryFacts.length,
      freshness: memoryEntries[0]?.modified ? formatTimestamp(memoryEntries[0].modified) : "No memory file",
      reliability: memoryFacts.length ? "classified facts" : memoryEntries.length ? "file-backed" : "needs seed",
      action: memoryFacts.length ? "Audit categories and stale facts." : "Draft a first reviewed memory source.",
      prompt: `Audit Astria memory sources.\n\nMemory files: ${memoryEntries.length}\nClassified facts: ${memoryFacts.length}\nWarnings: ${(state.memory?.warnings || []).length}\n\nIdentify stale, duplicate, or missing durable facts and propose a maintenance action.`,
    },
    {
      id: "sessions",
      type: "Sessions",
      title: "Favorite sessions",
      panel: "memory",
      evidence: favoriteSessions.length,
      freshness: favoriteSessions[0]?.updated_at ? formatTimestamp(favoriteSessions[0].updated_at) : favoriteSessions[0]?.id || "No favorite session",
      reliability: favoriteSessions.length ? "operator selected" : "needs favorite",
      action: favoriteSessions.length ? "Convert useful favorites into memory." : "Favorite a session before trusting it as a source.",
      prompt: `Review favorite sessions as Astria knowledge sources.\n\nFavorites: ${favoriteSessions.map((session) => session.title || session.id).join(", ") || "none"}\n\nChoose what should become durable memory and what should remain ephemeral.`,
    },
    {
      id: "runs",
      type: "Runs",
      title: "Execution evidence",
      panel: "runs",
      evidence: state.runs.length,
      freshness: latestRun?.started_at ? formatTimestamp(latestRun.started_at) : "No runs",
      reliability: latestRun ? `${latestRun.status || "unknown"} latest` : "needs execution",
      action: latestRun ? "Promote stable run outcomes into memory." : "Run a baseline task before citing execution evidence.",
      prompt: `Review recent runs as knowledge sources.\n\nLatest run: ${latestRun?.prompt || "none"}\nStatus: ${latestRun?.status || "unknown"}\nRun count: ${state.runs.length}\n\nIdentify which outcomes are reliable enough to cite in future prompts.`,
    },
    {
      id: "intake",
      type: "File Intake",
      title: intake?.path || "Local file evidence",
      panel: "intake",
      evidence: intake && !intake.is_error ? 1 : 0,
      freshness: intake?.mode || "No intake result",
      reliability: intake?.is_error ? "error" : intake ? "read-only sample" : "needs file",
      action: intake ? "Summarize source limits before using it." : "Inspect a file to seed source-grounded knowledge.",
      prompt: `Review file intake as an Astria source.\n\nPath: ${intake?.path || "none"}\nMode: ${intake?.mode || "none"}\nError: ${Boolean(intake?.is_error)}\n\nState what can be trusted, what is incomplete, and what should be re-read.`,
    },
    {
      id: "council",
      type: "Council",
      title: latestCouncil?.goal || "Council synthesis",
      panel: "council",
      evidence: latestCouncil ? 1 + (latestCouncil.roles || []).length : 0,
      freshness: latestCouncil?.created_at ? formatTimestamp(latestCouncil.created_at) : "No council run",
      reliability: latestCouncil?.synthesis ? "multi-role synthesis" : "needs review",
      action: latestCouncil ? "Check whether synthesis should become memory." : "Run council before citing reviewed judgment.",
      prompt: `Review council output as a knowledge source.\n\nGoal: ${latestCouncil?.goal || "none"}\nRoles: ${(latestCouncil?.roles || []).map((role) => role.role).join(", ") || "none"}\n\nDecide which conclusions are durable and which need another review.`,
    },
  ];
}

function renderSourceRegistry() {
  const rows = sourceRegistryRows();
  setText("nav-sources-count", rows.length);
  setText("manage-sources-count", `${rows.length} source${rows.length === 1 ? "" : "s"}`);
  setText("sources-summary", `${rows.length} source lane${rows.length === 1 ? "" : "s"} tracking freshness, reliability, and maintenance.`);
  const list = $("source-registry-list");
  if (!list) return;
  if (!state.selectedSourceRow || !rows.some((row) => row.id === state.selectedSourceRow)) {
    state.selectedSourceRow = rows[0]?.id || "";
  }
  list.innerHTML = rows.map((row) => `<article class="source-row ${row.id === state.selectedSourceRow ? "active" : ""}" data-source-row="${escapeHTML(row.id)}">
    <div class="row-item-title"><span>${escapeHTML(row.type)}</span><span class="tag">${escapeHTML(String(row.evidence))}</span></div>
    <strong>${escapeHTML(row.title)}</strong>
    <div class="source-grid">
      <span>Freshness</span><strong>${escapeHTML(row.freshness)}</strong>
      <span>Reliability</span><strong>${escapeHTML(row.reliability)}</strong>
      <span>Action</span><strong>${escapeHTML(row.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-source-select="${escapeHTML(row.id)}">Source brief</button>
      <button type="button" data-source-draft="${escapeHTML(row.id)}">Draft maintenance</button>
      <button type="button" data-panel="${escapeHTML(row.panel)}">Open source</button>
    </div>
  </article>`).join("");
  renderSourceRegistryDetail(rows.find((row) => row.id === state.selectedSourceRow) || rows[0]);
}

function renderSourceRegistryDetail(row) {
  const target = $("source-registry-detail");
  if (!target) return;
  if (!row) {
    target.innerHTML = `<div class="empty-state">Select a source row.</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(row.title)}</h3>
      <div class="run-meta-grid">
        <span>Type</span><strong>${escapeHTML(row.type)}</strong>
        <span>Evidence</span><strong>${escapeHTML(String(row.evidence))}</strong>
        <span>Route</span><strong>${escapeHTML(row.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Reliability</h3>
      <p>${escapeHTML(row.reliability)}</p>
      <h3>Maintenance</h3>
      <p>${escapeHTML(row.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-source-draft="${escapeHTML(row.id)}">Draft maintenance</button>
        <button type="button" data-panel="${escapeHTML(row.panel)}">Open source</button>
      </div>
    </section>
  </div>`;
}

function sourceRowByID(id) {
  return sourceRegistryRows().find((row) => row.id === id) || null;
}

function draftSourceMaintenanceToChat(id) {
  const row = sourceRowByID(id);
  if (!row) return;
  $("chat-input").value = `${row.prompt}\n\nReturn a concise source maintenance brief with freshness, reliability, citations to inspect, and next action.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Source maintenance drafted to chat.");
}

function renderMemoryTaxonomyBar(categories) {
  const bar = $("memory-taxonomy-bar");
  if (!bar) return;
  const order = ["all", "preferences", "decisions", "commands", "architecture", "people", "risks", "uncategorized"];
  const total = Object.values(categories).reduce((sum, value) => sum + Number(value || 0), 0);
  bar.innerHTML = order.map((category) => {
    const count = category === "all" ? total : (categories[category] || 0);
    const active = state.memoryCategory === category || (!state.memoryCategory && category === "all");
    return `<button type="button" class="${active ? "active" : ""}" data-memory-category="${escapeHTML(category)}">${escapeHTML(memoryCategoryLabel(category))}<span>${count}</span></button>`;
  }).join("");
}

function renderMemoryWarnings(warnings) {
  const list = $("memory-warning-list");
  if (!list) return;
  if (!warnings.length) {
    list.innerHTML = "";
    return;
  }
  list.innerHTML = warnings.map((warning) => `<article class="memory-warning-card ${escapeHTML(warning.type || "")}">
    <strong>${escapeHTML(warning.type || "warning")}</strong>
    <span>${escapeHTML(warning.message || "")}</span>
    ${warning.lines?.length ? `<small>lines ${warning.lines.map(escapeHTML).join(", ")}</small>` : ""}
  </article>`).join("");
}

function memoryCategoryLabel(category) {
  const labels = {
    all: "All",
    preferences: "Prefs",
    decisions: "Decisions",
    commands: "Commands",
    architecture: "Architecture",
    people: "People",
    risks: "Risks",
    uncategorized: "Other",
  };
  return labels[category] || category;
}

function formatBytes(size) {
  if (!Number.isFinite(size)) return "-";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
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

async function loadVersion() {
  const list = $("version-list");
  try {
    const data = await api("/version");
    state.version = data;
    state.updateCheck = null;
    setText("settings-version-state", data.version || "Build");
    setClass("settings-version-state", data.update_supported ? "ready" : "warning");
    renderVersion();
  } catch (error) {
    state.version = null;
    setText("settings-version-state", "Error");
    setClass("settings-version-state", "bad");
    $("version-summary").textContent = "Version metadata unavailable.";
    $("update-check-state").textContent = "Error";
    renderError(list, error.message);
  }
}

async function loadDiagnostics() {
  const list = $("diagnostics-list");
  const summary = $("diagnostics-summary");
  const overview = $("diagnostics-overview");
  const chip = $("diagnostics-chip");
  try {
    const diagnostics = await api("/diagnostics");
    state.diagnostics = diagnostics;
    const status = diagnostics.status || "unknown";
    const label = statusLabel(status);
    setText("settings-state", label);
    setClass("settings-state", status);
    setText("nav-diagnostics-state", label);
    setClass("nav-diagnostics-state", status);
    setText("settings-diagnostics-state", label);
    setClass("settings-diagnostics-state", status);
    chip.textContent = label;
    chip.className = `diagnostics-chip ${status}`;
    summary.textContent = diagnostics.summary || "Runtime readiness checks.";
    overview.innerHTML = `<strong>${escapeHTML(label)}</strong><span>${escapeHTML(diagnostics.summary || "")}</span>`;
    renderConfigDiagnosticsOverview(diagnostics);
    renderWorkspaceHealthStrip();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
    if ($("chat-output").querySelector(".empty-thread")) renderEmptyThread();
    const checks = Array.isArray(diagnostics.checks) ? diagnostics.checks : [];
    const launchRows = diagnosticsLaunchRows(diagnostics);
    const launchCard = `<article class="row-item diagnostic-launch-card">
      <div class="row-item-title"><span>Launch readiness</span><span class="tag">${escapeHTML(label)}</span></div>
      <div class="run-meta-grid">
        ${launchRows.map(([rowLabel, value]) => `<span>${escapeHTML(rowLabel)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
      </div>
    </article>`;
    if (!checks.length) {
      list.innerHTML = `${launchCard}<article class="row-item empty-state"><p>No diagnostics returned.</p></article>`;
      return;
    }
    const checkCards = checks.map((check) => `<article class="row-item diagnostic-item ${escapeHTML(check.status || "unknown")}">
      <div class="row-item-title">
        <span>${escapeHTML(check.label || check.id || "Check")}</span>
        <span class="tag diagnostic-tag ${escapeHTML(check.status || "unknown")}">${escapeHTML(statusLabel(check.status))}</span>
      </div>
      <p>${escapeHTML(check.detail || "")}</p>
      ${diagnosticActionHTML(check)}
    </article>`).join("");
    list.innerHTML = `${launchCard}${checkCards}`;
  } catch (error) {
    state.diagnostics = null;
    setText("settings-state", "Offline");
    setClass("settings-state", "error");
    setText("nav-diagnostics-state", "Offline");
    setClass("nav-diagnostics-state", "error");
    setText("settings-diagnostics-state", "Offline");
    setClass("settings-diagnostics-state", "error");
    chip.textContent = "Diagnostics unavailable";
    chip.className = "diagnostics-chip error";
    summary.textContent = "Diagnostics unavailable.";
    overview.innerHTML = `<strong>Error</strong><span>${escapeHTML(error.message)}</span>`;
    renderConfigDiagnosticsOverview({ status: "error", summary: error.message });
    renderError(list, error.message);
    renderWorkspaceHealthStrip();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  }
}

function diagnosticsLaunchRows(diagnostics) {
  const rows = [
    ["Launch", diagnostics.launch_command || "starclaw app"],
    ["Web UI", diagnostics.web_url || "-"],
    ["Health", diagnostics.health_url || "-"],
    ["Status API", diagnostics.status_url || "-"],
    ["Diagnostics", diagnostics.diagnostics_url || "-"],
    ["Data", diagnostics.starclaw_dir || "-"],
    ["Config", diagnostics.config_path || "-"],
    ["Agents", diagnostics.agents_dir || "-"],
    ["Sessions", diagnostics.sessions_dir || "-"],
  ];
  if (diagnostics.executable_path) rows.push(["Executable", diagnostics.executable_path]);
  return rows;
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
    renderMCPStarport();
    renderToolDockInspector();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  } catch (error) {
    state.config = null;
    setText("settings-config-state", "Error");
    setClass("settings-config-state", "bad");
    $("config-save-state").textContent = error.message;
    renderMCPStarport();
    renderToolDockInspector();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
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
  setText("settings-config-state", cfg.provider || "Provider");
  setClass("settings-config-state");
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
    fillPermissionsForm();
    renderPermissions();
    renderWorkspaceHealthStrip();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  } catch (error) {
    state.permissions = null;
    setText("settings-permissions-state", "Error");
    setClass("settings-permissions-state", "bad");
    $("permissions-save-state").textContent = "Error";
    $("permissions-overview").innerHTML = `<strong>Error</strong><span>${escapeHTML(error.message)}</span>`;
    renderError(list, error.message);
    renderWorkspaceHealthStrip();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  }
}

async function loadMemory() {
  try {
    state.memory = await api("/memory");
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderAgentContinuityDigest();
    renderHomeDockedTools();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
  } catch (error) {
    state.memory = { entries: [], content: "", memory_dir: "" };
    setText("memory-save-state", "Error");
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderAgentContinuityDigest();
    renderHomeDockedTools();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    showToast(error.message);
  }
}

async function submitMemoryCandidate(event) {
  event.preventDefault();
  const content = $("memory-candidate").value.trim();
  if (!content) {
    showToast("Add a memory candidate first.");
    return;
  }
  $("memory-save-state").textContent = "Saving";
  try {
    state.memory = await api("/memory", {
      method: "POST",
      body: JSON.stringify({ content }),
    });
    $("memory-candidate").value = "";
    $("memory-save-state").textContent = "Saved";
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    showToast("Memory approved.");
  } catch (error) {
    $("memory-save-state").textContent = "Error";
    showToast(error.message);
  }
}

async function deleteMemoryEntry(name) {
  if (!name) return;
  if (!globalThis.confirm(`Delete memory entry "${name}"?`)) return;
  try {
    state.memory = await api(`/memory/${encodeURIComponent(name)}`, { method: "DELETE" });
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    showToast("Memory entry deleted.");
  } catch (error) {
    showToast(error.message);
  }
}

function draftMemoryCandidate(source) {
  const [kind, id] = String(source || "").split(":", 2);
  const prefix = kind === "run" ? `run ${id}` : kind === "session" ? `session ${id}` : "recent work";
  $("memory-candidate").value = `- From ${prefix}: `;
  renderMemoryCandidatePreview();
  switchPanel("memory");
  $("memory-candidate").focus();
}

function draftAgentMemoryCandidate(name) {
  if (!name) return;
  const summary = agentCapabilitySummary(state.agents.find((agent) => normalizeName(agent) === name) || { name });
  const runs = runsForAgent(name);
  const latest = runs[0] || null;
  $("memory-candidate").value = [
    `- Agent ${name}: ${summary.description}`,
    `- Agent ${name} continuity: ${runs.length} recorded run${runs.length === 1 ? "" : "s"}; latest status ${latest?.status || "none"}.`,
    `- Agent ${name} next memory review: ${agentContinuityHint(summary, runs)}`,
  ].join("\n");
  renderMemoryCandidatePreview();
  switchPanel("memory");
  $("memory-candidate").focus();
  showToast(`Memory draft prepared for ${name}.`);
}

function renderMemoryCandidatePreview() {
  const target = $("memory-candidate-preview");
  if (!target) return;
  const text = $("memory-candidate")?.value || "";
  const facts = parseCandidateFacts(text);
  if (!facts.length) {
    target.textContent = "No candidate yet.";
    return;
  }
  const categories = facts.map((fact) => memoryCategoryLabel(fact.category)).join(", ");
  const existingTexts = new Set((state.memory?.facts || []).map((fact) => normalizeCandidateText(fact.text)));
  const duplicate = facts.some((fact) => existingTexts.has(normalizeCandidateText(fact.text)));
  target.innerHTML = `<strong>${escapeHTML(categories)}</strong><span>${duplicate ? "Possible duplicate before approval." : "Ready for reviewed memory approval."}</span>`;
}

function parseCandidateFacts(text) {
  return String(text || "").split("\n").map((line) => {
    const trimmed = line.trim();
    if (!trimmed.startsWith("-")) return null;
    const bracket = trimmed.match(/^[-*]\s*\[([A-Za-z _-]+)\]\s*(.+)$/);
    if (bracket) return { category: normalizeCandidateCategory(bracket[1]), text: bracket[2].trim() };
    const colon = trimmed.match(/^[-*]\s*([A-Za-z _-]+):\s*(.+)$/);
    if (colon) {
      const category = normalizeCandidateCategory(colon[1]);
      if (category !== "uncategorized") return { category, text: colon[2].trim() };
    }
    return { category: "uncategorized", text: trimmed.replace(/^[-*]\s*/, "") };
  }).filter(Boolean);
}

function normalizeCandidateCategory(value) {
  const key = String(value || "").toLowerCase().trim().replaceAll(" ", "_").replaceAll("-", "_");
  const map = {
    preference: "preferences",
    preferences: "preferences",
    decision: "decisions",
    decisions: "decisions",
    command: "commands",
    commands: "commands",
    architecture: "architecture",
    arch: "architecture",
    person: "people",
    people: "people",
    risk: "risks",
    risks: "risks",
  };
  return map[key] || "uncategorized";
}

function normalizeCandidateText(text) {
  return String(text || "").toLowerCase().trim().replace(/\s+/g, " ");
}

async function loadCouncilRuns() {
  const list = $("council-list");
  try {
    const data = await api("/council");
    state.councilRuns = Array.isArray(data.runs) ? data.runs : [];
    renderCouncilRuns();
    renderSourceRegistry();
    renderComparisonWorkbench();
    renderReuseGallery();
  } catch (error) {
    state.councilRuns = [];
    state.currentCouncilRun = null;
    setText("council-state", "Error");
    renderError(list, error.message);
    renderComparisonWorkbench();
    renderSourceRegistry();
    renderReuseGallery();
  }
}

function renderCouncilRuns() {
  const list = $("council-list");
  if (!list) return;
  const count = state.councilRuns.length;
  setText("nav-council-count", count);
  setText("manage-council-count", `${count} ${count === 1 ? "review" : "reviews"}`);
  renderManageCount();
  setText("council-summary", count ? `${count} council run${count === 1 ? "" : "s"} with role contributions.` : "No council runs yet.");
  if (!count) {
    renderEmptyAction(list, "No council runs yet. Start with a planning or review goal.", [
      { label: "Seed council goal", homeAction: "council", primary: true },
      { label: "Open chat", panel: "chat" },
    ]);
    renderCouncilDetail(state.currentCouncilRun);
    return;
  }
  list.innerHTML = state.councilRuns.map((run) => `<article class="row-item council-run-card ${state.currentCouncilRun?.id === run.id ? "active" : ""}" data-council-id="${escapeHTML(run.id)}">
    <div class="row-item-title"><span>${escapeHTML(run.goal || run.id)}</span><span class="tag">${escapeHTML(run.status || "unknown")}</span></div>
    <p>${escapeHTML((run.roles || []).map((role) => role.role).join(" · ") || "No role contributions")} · ${escapeHTML(formatTimestamp(run.created_at))}</p>
    <div class="row-actions">
      <button type="button" data-council-open="${escapeHTML(run.id)}">Open council</button>
      <button type="button" data-council-copy="${escapeHTML(run.id)}">Copy synthesis</button>
    </div>
  </article>`).join("");
  if (!state.currentCouncilRun) renderCouncilDetail(state.councilRuns[0]);
}

function renderCouncilDetail(run) {
  const target = $("council-detail");
  if (!target) return;
  state.currentCouncilRun = run || null;
  if (!run) {
    target.innerHTML = `<div class="empty-state">Start or select a council run.</div>`;
    return;
  }
  const roles = Array.isArray(run.roles) ? run.roles : [];
  const stages = councilStages(run, roles);
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(run.goal || "Council run")}</h3>
      <div class="run-meta-grid">
        <span>Status</span><strong>${escapeHTML(run.status || "unknown")}</strong>
        <span>Created</span><strong>${escapeHTML(formatTimestamp(run.created_at))}</strong>
        <span>Roles</span><strong>${roles.length}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Council stages</h3>
      <div class="council-stage-rail">
        ${stages.map((stage) => `<article class="council-stage-card ${stage.kind === "role" ? "role-stage" : "system-stage"}">
          <div class="council-stage-orbit"><span>${escapeHTML(stage.step)}</span></div>
          <div class="council-stage-body">
            <div class="row-item-title"><span>${escapeHTML(stage.title)}</span><span class="tag">${escapeHTML(stage.status)}</span></div>
            <strong>${escapeHTML(stage.summary)}</strong>
            <p>${escapeHTML(stage.preview)}</p>
            <div class="row-actions">
              ${stage.actions.map((action) => `<button type="button" class="${action.primary ? "primary-button" : ""}" ${action.attr}>${escapeHTML(action.label)}</button>`).join("")}
            </div>
          </div>
        </article>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Role contributions</h3>
      <div class="council-role-list">
        ${roles.map((role, index) => `<article class="council-role-card">
          <div class="row-item-title"><span>${escapeHTML(role.role || "role")}</span><span class="tag">${escapeHTML(role.status || "unknown")}</span></div>
          <strong>${escapeHTML(role.summary || "")}</strong>
          <p>${escapeHTML(role.notes || "")}</p>
          <div class="row-actions">
            <button type="button" data-council-role-copy="${escapeHTML(run.id)}" data-council-role-index="${index}">Copy notes</button>
            <button type="button" data-council-role-draft="${escapeHTML(run.id)}" data-council-role-index="${index}">Draft to chat</button>
          </div>
        </article>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Final synthesis</h3>
      <pre>${escapeHTML(run.synthesis || "")}</pre>
      <div class="run-detail-actions">
        <button type="button" data-council-copy="${escapeHTML(run.id)}">Copy synthesis</button>
        <button type="button" data-council-send="${escapeHTML(run.id)}">Send to chat</button>
        <button type="button" class="primary-button" data-council-run="${escapeHTML(run.id)}">Start run</button>
      </div>
    </section>
  </div>`;
}

function councilStages(run, roles) {
  const roleStages = ["planner", "researcher", "reviewer"].map((name, index) => {
    const matchedIndex = roles.findIndex((item) => String(item.role || "").toLowerCase() === name);
    const roleIndex = matchedIndex >= 0 ? matchedIndex : index;
    const role = roles[roleIndex] || {};
    return {
      kind: "role",
      step: String(index + 1),
      title: role.role || name,
      status: role.status || run.status || "pending",
      summary: role.summary || "Role output pending.",
      preview: role.notes || "No notes captured yet.",
      actions: [
        { label: "Copy notes", attr: `data-council-role-copy="${escapeHTML(run.id)}" data-council-role-index="${roleIndex}"` },
        { label: "Draft to chat", attr: `data-council-role-draft="${escapeHTML(run.id)}" data-council-role-index="${roleIndex}"` },
      ],
    };
  });
  return [
    ...roleStages,
    {
      kind: "synthesis",
      step: "4",
      title: "Synthesis",
      status: run.synthesis ? "ready" : "pending",
      summary: "Merge role outputs into one implementation direction.",
      preview: run.synthesis || "No synthesis captured yet.",
      actions: [
        { label: "Copy synthesis", attr: `data-council-copy="${escapeHTML(run.id)}"` },
        { label: "Send to chat", attr: `data-council-send="${escapeHTML(run.id)}"` },
      ],
    },
    {
      kind: "handoff",
      step: "5",
      title: "Handoff",
      status: run.status === "completed" ? "ready" : run.status || "pending",
      summary: "Start the next concrete Astria run from the council result.",
      preview: run.agent ? `Lead agent: ${run.agent}` : "Use the selected agent for the handoff run.",
      actions: [
        { label: "Start run", attr: `data-council-run="${escapeHTML(run.id)}"`, primary: true },
      ],
    },
  ];
}

async function submitCouncilRun(event) {
  event.preventDefault();
  const goal = $("council-goal").value.trim();
  if (!goal) {
    showToast("Add a council goal first.");
    return;
  }
  $("council-state").textContent = "Running";
  try {
    const run = await api("/council", {
      method: "POST",
      body: JSON.stringify({ goal, agent: $("council-agent").value }),
    });
    $("council-goal").value = "";
    $("council-state").textContent = "Completed";
    state.currentCouncilRun = run;
    await loadCouncilRuns();
    renderCouncilDetail(run);
    showToast("Council synthesis ready.");
  } catch (error) {
    $("council-state").textContent = "Error";
    showToast(error.message);
  }
}

function councilRunByID(id) {
  if (state.currentCouncilRun?.id === id) return state.currentCouncilRun;
  return state.councilRuns.find((run) => run.id === id) || null;
}

function councilSynthesisText(run) {
  return run?.synthesis || "";
}

function councilRoleByIndex(id, index) {
  const run = councilRunByID(id);
  const roles = Array.isArray(run?.roles) ? run.roles : [];
  const roleIndex = Number.parseInt(index, 10);
  if (!Number.isInteger(roleIndex) || roleIndex < 0 || roleIndex >= roles.length) return null;
  return { run, role: roles[roleIndex] };
}

function copyCouncilSynthesis(id, button) {
  copyText(councilSynthesisText(councilRunByID(id)), "Council synthesis copied.")
    .then(() => {
      if (button) markButtonCopied(button);
    })
    .catch((error) => showToast(error.message));
}

function copyCouncilRoleNotes(id, index, button) {
  const found = councilRoleByIndex(id, index);
  if (!found) return;
  const text = councilRoleText(found.run, found.role);
  copyText(text, "Council role notes copied.")
    .then(() => {
      if (button) markButtonCopied(button);
    })
    .catch((error) => showToast(error.message));
}

function councilRoleText(run, role) {
  return [
    `Council goal: ${run?.goal || "Untitled council run"}`,
    `Role: ${role?.role || "role"}`,
    `Status: ${role?.status || "unknown"}`,
    "",
    role?.summary || "",
    role?.notes || "",
  ].filter((line, index) => index < 3 || line !== "").join("\n");
}

function draftCouncilRoleToChat(id, index) {
  const found = councilRoleByIndex(id, index);
  if (!found) return;
  $("chat-input").value = `${councilRoleText(found.run, found.role)}\n\nTurn this role perspective into a concrete next action for Astria.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Council role drafted to chat.");
}

function sendCouncilToChat(id) {
  const run = councilRunByID(id);
  if (!run) return;
  $("chat-input").value = councilSynthesisText(run);
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
}

async function runCouncilSynthesis(id) {
  if (!id) return;
  $("council-state").textContent = "Starting run";
  try {
    const result = await api(`/council/${encodeURIComponent(id)}/run`, {
      method: "POST",
      body: JSON.stringify({ agent: $("council-agent").value }),
    });
    $("council-state").textContent = "Run started";
    await loadRuns();
    if (result.run_id) {
      selectRun(result.run_id);
    }
    showToast("Council handoff run completed.");
  } catch (error) {
    $("council-state").textContent = "Error";
    showToast(error.message);
  }
}

async function loadInbox() {
  const list = $("inbox-list");
  try {
    const data = await api("/inbox");
    state.inboxItems = Array.isArray(data.items) ? data.items : [];
    renderInbox();
    renderProactiveDeliveryBoard();
  } catch (error) {
    state.inboxItems = [];
    setText("inbox-state", "Error");
    renderError(list, error.message);
    renderProactiveDeliveryBoard();
  }
}

async function loadInboxProviders() {
  const list = $("inbox-provider-list");
  try {
    const data = await api("/inbox/providers");
    state.inboxProviders = Array.isArray(data.providers) ? data.providers : [];
    renderInboxProviders();
    renderProactiveDeliveryBoard();
  } catch (error) {
    state.inboxProviders = [];
    renderError(list, error.message);
    renderProactiveDeliveryBoard();
  }
}

function renderInboxProviders() {
  const list = $("inbox-provider-list");
  if (!list) return;
  if (!state.inboxProviders.length) {
    renderEmpty(list, "No inbox providers reported.");
    return;
  }
  list.innerHTML = state.inboxProviders.map((provider) => `<article class="provider-route-card ${provider.kind || ""}">
    <div class="row-item-title">
      <span>${escapeHTML(provider.name || provider.kind || "Provider")}</span>
      <span class="tag">${escapeHTML(provider.configured ? "ready" : "setup")}</span>
    </div>
    <code>${escapeHTML(provider.endpoint || "")}</code>
    <p>${escapeHTML(provider.description || "")}</p>
    <div class="pill-list">
      ${(provider.supported_events || []).map((event) => `<span>${escapeHTML(event)}</span>`).join("")}
      <span>${provider.secret_configured ? "secret set" : "no secret"}</span>
    </div>
  </article>`).join("");
}

function inboxStatusCounts() {
  return state.inboxItems.reduce((counts, item) => {
    const status = item.status || "pending";
    counts[status] = (counts[status] || 0) + 1;
    return counts;
  }, {});
}

function renderInbox() {
  const list = $("inbox-list");
  if (!list) return;
  const counts = inboxStatusCounts();
  const pending = counts.pending || 0;
  const failed = counts.failed || 0;
  const completed = counts.completed || 0;
  setText("nav-inbox-count", pending);
  setText("manage-inbox-count", `${pending} pending`);
  setText("home-inbox-count", pending);
  renderManageCount();
  renderPromptSuggestionDock();
  renderApprovalCenter();
  renderReviewQueue();
  setText("inbox-summary", state.inboxItems.length ? `${pending} pending · ${failed} failed · ${completed} completed` : "No inbound tasks yet.");
  const overview = $("inbox-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(pending ? `${pending} waiting` : "Guarded")}</strong><span>${escapeHTML(pending ? "Review inbound channel work before it can become an Astria run." : "Inbound items never execute until approved.")}</span>`;
  }
  if (!state.inboxItems.length) {
    renderEmptyAction(list, "No inbound tasks yet. Use the local webhook intake to simulate one.", [
      { label: "Open intake", panel: "inbox", primary: true },
      { label: "Open chat", panel: "chat" },
    ]);
    return;
  }
  list.innerHTML = state.inboxItems.map((item) => `<article class="row-item inbox-card ${escapeHTML(item.status || "pending")}">
    <div class="row-item-title">
      <span>${escapeHTML(item.text || item.external_id || item.id)}</span>
      <span class="tag">${escapeHTML(item.status || "pending")}</span>
    </div>
    <p>${escapeHTML(item.sender || "unknown sender")} · ${escapeHTML(item.provider || "webhook")} · ${escapeHTML(item.external_id || item.id)} · ${escapeHTML(formatTimestamp(item.created_at))}</p>
    ${item.error ? `<p class="error-copy">${escapeHTML(item.error)}</p>` : ""}
    <div class="pill-list">
      <span>${escapeHTML(item.agent || "default agent")}</span>
      ${item.run_id ? `<span>run ${escapeHTML(item.run_id)}</span>` : "<span>approval required</span>"}
    </div>
    <div class="row-actions">${inboxActionsHTML(item)}</div>
  </article>`).join("");
}

function inboxActionsHTML(item) {
  const id = escapeHTML(item.id || "");
  switch (item.status) {
    case "pending":
      return `<button type="button" class="primary-button" data-inbox-approve="${id}">Approve</button><button type="button" data-inbox-reject="${id}">Reject</button>`;
    case "failed":
      return `<button type="button" class="primary-button" data-inbox-retry="${id}">Retry</button><button type="button" data-inbox-reject="${id}">Reject</button>`;
    case "completed":
      return item.run_id ? `<button type="button" data-inbox-run="${escapeHTML(item.run_id)}">Open run</button>` : "";
    case "running":
      return `<button type="button" disabled>Running</button>`;
    case "rejected":
      return `<button type="button" disabled>Rejected</button>`;
    default:
      return "";
  }
}

async function submitInboxWebhook(event) {
  event.preventDefault();
  const externalID = $("inbox-external-id").value.trim();
  const text = $("inbox-text").value.trim();
  if (!externalID || !text) {
    showToast("External ID and text are required.");
    return;
  }
  $("inbox-state").textContent = "Receiving";
  try {
    const data = await api("/inbox/webhook", {
      method: "POST",
      body: JSON.stringify({
        provider: "webhook",
        external_id: externalID,
        sender: $("inbox-sender").value.trim(),
        text,
        agent: $("inbox-agent").value,
      }),
    });
    $("inbox-state").textContent = data.duplicate ? "Duplicate" : "Received";
    if (!data.duplicate) {
      $("inbox-external-id").value = "";
      $("inbox-sender").value = "";
      $("inbox-text").value = "";
    }
    await loadInbox();
    showToast(data.duplicate ? "Duplicate webhook ignored." : "Inbound task received.");
  } catch (error) {
    $("inbox-state").textContent = "Error";
    showToast(error.message);
  }
}

async function approveInboxItem(id) {
  await runInboxAction(id, "approve", "Approving", "Inbox item approved.");
}

async function rejectInboxItem(id) {
  await runInboxAction(id, "reject", "Rejecting", "Inbox item rejected.");
}

async function retryInboxItem(id) {
  await runInboxAction(id, "retry", "Retrying", "Inbox item retried.");
}

async function runInboxAction(id, action, busyText, doneText) {
  if (!id) return;
  $("inbox-state").textContent = busyText;
  try {
    const options = { method: "POST" };
    if (action === "approve" || action === "retry") {
      options.body = JSON.stringify({ agent: $("inbox-agent").value });
    }
    await api(`/inbox/${encodeURIComponent(id)}/${action}`, options);
    $("inbox-state").textContent = "Ready";
    await Promise.allSettled([loadInbox(), loadRuns()]);
    showToast(doneText);
  } catch (error) {
    $("inbox-state").textContent = "Error";
    await loadInbox();
    showToast(error.message);
  }
}

function openInboxRun(runID) {
  if (!runID) return;
  selectRun(runID);
}

function fillPermissionsForm() {
  const permissions = state.permissions || {};
  $("permissions-allowed-dirs").value = formatRuleList(permissions.allowed_dirs || []);
  $("permissions-allowed-commands").value = formatRuleList(permissions.allowed_commands || []);
  $("permissions-denied-commands").value = formatRuleList(permissions.denied_commands || []);
  $("permissions-network-allowlist").value = formatRuleList(permissions.network_allowlist || []);
  $("permissions-sensitive-patterns").value = formatRuleList(permissions.sensitive_patterns || []);
  $("permissions-save-state").textContent = "Loaded";
  renderPermissionsPendingPreview();
}

function renderPermissions() {
  const permissions = state.permissions || {};
  const configured = permissions.configured === true;
  setText("settings-permissions-state", configured ? "Configured" : "Defaults");
  setClass("settings-permissions-state", configured ? "ready" : "warning");
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
  renderPermissionsPendingPreview();
}

function renderVersion() {
  const info = state.version || {};
  const check = state.updateCheck;
  const supported = info.update_supported === true;
  $("version-summary").textContent = info.message || "Build and update status.";
  $("update-check-button").disabled = !supported;
  if (!check) {
    $("update-check-state").textContent = supported ? "Ready" : "Unavailable";
    $("update-overview").innerHTML = `<strong>${escapeHTML(supported ? "Ready" : "Development build")}</strong><span>${escapeHTML(info.message || "")}</span>`;
  } else {
    const label = updateStatusLabel(check.status);
    $("update-check-state").textContent = label;
    $("update-overview").innerHTML = `<strong>${escapeHTML(label)}</strong><span>${escapeHTML(check.message || "")}</span>`;
  }
  const updateRows = [
    ["Version", info.version || "-"],
    ["Platform", info.platform || "-"],
    ["Web UI", info.web_url || "-"],
    ["Launch", info.launch_command || "starclaw app"],
    ["Update checks", supported ? "Supported" : "Release build required"],
    ["Command", info.update_command || "starclaw update --check"],
  ];
  if (check?.latest_version) updateRows.push(["Latest", check.latest_version]);
  if (check?.release_url) updateRows.push(["Release URL", check.release_url]);
  const runtimeRows = [
    ["Web UI", info.web_url || "-"],
    ["Health", info.health_url || "-"],
    ["Status API", info.status_url || "-"],
    ["Diagnostics", info.diagnostics_url || "-"],
    ["Data", info.starclaw_dir || "-"],
    ["Config", info.config_path || "-"],
  ];
  const readinessRows = [
    ["Build", supported ? "Release build" : "Development build"],
    ["Updates", supported ? "Update checks available" : "Release build required"],
    ["Launch", info.launch_command || "starclaw app"],
    ["Web UI", info.web_url || "-"],
  ];
  $("version-list").innerHTML = `<article class="row-item version-readiness-card">
    <div class="row-item-title"><span>Release readiness</span><span class="tag">${escapeHTML(supported ? "Ready" : "Development")}</span></div>
    <div class="run-meta-grid">
      ${readinessRows.map(([label, value]) => `<span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
    </div>
    ${supported ? `<p>Release metadata is available for update checks.</p>` : `<p>Use a semver release build to enable update checks.</p>`}
  </article>
  <article class="row-item version-card">
    <div class="row-item-title"><span>Runtime context</span><span class="tag">local</span></div>
    <div class="run-meta-grid">
      ${runtimeRows.map(([label, value]) => `<span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
    </div>
  </article>
  <article class="row-item version-card">
    <div class="run-meta-grid">
      ${updateRows.map(([label, value]) => `<span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
    </div>
  </article>`;
}

function updateStatusLabel(status) {
  switch (status) {
    case "available":
      return "Update available";
    case "current":
      return "Up to date";
    case "development":
      return "Development build";
    default:
      return "Unknown";
  }
}

function supportInfoText() {
  const info = state.version || {};
  const diagnostics = state.diagnostics || {};
  const rows = [
    ["Astria support info", ""],
    ["Version", info.version || "-"],
    ["Platform", info.platform || "-"],
    ["Build status", info.status || "-"],
    ["Update supported", info.update_supported === true ? "yes" : "no"],
    ["Update command", info.update_command || "starclaw update --check"],
    ["Launch command", info.launch_command || "starclaw app"],
    ["Web UI", info.web_url || "-"],
    ["Health URL", info.health_url || "-"],
    ["Status URL", info.status_url || "-"],
    ["Diagnostics URL", info.diagnostics_url || "-"],
    ["Data dir", info.starclaw_dir || "-"],
    ["Config path", info.config_path || "-"],
    ["Diagnostics status", diagnostics.status || "-"],
    ["Diagnostics summary", diagnostics.summary || "-"],
  ];
  return rows.map(([label, value]) => (value ? `${label}: ${value}` : label)).join("\n");
}

async function copySupportInfo() {
  await copyText(supportInfoText(), "Support info copied.");
}

async function checkForUpdates() {
  if (state.version && state.version.update_supported !== true) {
    showToast("Update checks require a release build.");
    return;
  }
  $("update-check-button").disabled = true;
  $("update-check-state").textContent = "Checking";
  try {
    state.updateCheck = await api("/update/check");
    renderVersion();
    showToast(state.updateCheck.message || "Update check complete.");
  } catch (error) {
    $("update-check-state").textContent = "Error";
    $("update-overview").innerHTML = `<strong>Error</strong><span>${escapeHTML(error.message)}</span>`;
    showToast(error.message);
  } finally {
    $("update-check-button").disabled = state.version?.update_supported !== true;
  }
}

function buildPermissionsPayload() {
  return {
    allowed_dirs: parseCSVList($("permissions-allowed-dirs").value),
    allowed_commands: parseCSVList($("permissions-allowed-commands").value),
    denied_commands: parseCSVList($("permissions-denied-commands").value),
    network_allowlist: parseCSVList($("permissions-network-allowlist").value),
    sensitive_patterns: parseCSVList($("permissions-sensitive-patterns").value),
  };
}

function permissionsRiskHints(permissions) {
  const hints = [];
  const allowedDirs = permissions.allowed_dirs || [];
  const deniedCommands = permissions.denied_commands || [];
  const networkAllowlist = permissions.network_allowlist || [];
  const sensitivePatterns = permissions.sensitive_patterns || [];
  if (allowedDirs.some((dir) => ["/", "~", "."].includes(dir.trim()))) {
    hints.push("Broad local access is allowed.");
  }
  if (!deniedCommands.length) {
    hints.push("No denied commands are configured.");
  }
  if (!sensitivePatterns.length) {
    hints.push("No sensitive file patterns are configured.");
  }
  if (networkAllowlist.some((host) => ["*", "*.*"].includes(host.trim()))) {
    hints.push("Network allowlist includes a broad wildcard.");
  }
  return hints;
}

function renderPermissionsPendingPreview() {
  const target = $("permissions-pending-preview");
  if (!target) return;
  const permissions = buildPermissionsPayload();
  const categories = [
    ["Allowed directories", permissions.allowed_dirs],
    ["Allowed commands", permissions.allowed_commands],
    ["Denied commands", permissions.denied_commands],
    ["Network allowlist", permissions.network_allowlist],
    ["Sensitive patterns", permissions.sensitive_patterns],
  ];
  const hints = permissionsRiskHints(permissions);
  target.innerHTML = `<article class="row-item permission-preview">
    <div class="row-item-title"><span>Pending changes</span><span class="tag">${hints.length ? "Review" : "Ready"}</span></div>
    <div class="permission-preview-grid">
      ${categories.map(([label, values]) => `<div class="permission-preview-count"><strong>${Array.isArray(values) ? values.length : 0}</strong><span>${escapeHTML(label)}</span></div>`).join("")}
    </div>
    ${hints.length ? `<div class="pill-list permission-risk-list">${hints.map((hint) => `<span>${escapeHTML(hint)}</span>`).join("")}</div>` : `<p>No obvious permission risks in pending changes.</p>`}
  </article>`;
}

async function submitPermissions(event) {
  event.preventDefault();
  $("permissions-save-state").textContent = "Saving";
  try {
    await api("/config", {
      method: "PATCH",
      body: JSON.stringify({ permissions: buildPermissionsPayload() }),
    });
    await Promise.allSettled([loadPermissions(), loadDiagnostics()]);
    $("permissions-save-state").textContent = "Saved";
    showToast("Permissions saved.");
  } catch (error) {
    $("permissions-save-state").textContent = "Error";
    showToast(error.message);
  }
}

async function clearPermissions() {
  $("permissions-allowed-dirs").value = "";
  $("permissions-allowed-commands").value = "";
  $("permissions-denied-commands").value = "";
  $("permissions-network-allowlist").value = "";
  $("permissions-sensitive-patterns").value = "";
  renderPermissionsPendingPreview();
  await submitPermissions(new Event("submit"));
}

async function loadAgents() {
  const list = $("agents-list");
  const roster = $("agent-capability-roster");
  try {
    const data = await api("/agents");
    state.agents = data.agents || [];
    setText("manage-agents-count", `${state.agents.length} ${state.agents.length === 1 ? "profile" : "profiles"}`);
    setText("nav-agents-count", state.agents.length);
    renderManageCount();
    renderHomeDockedTools();
    updateAgentSelects();
    renderAgentContinuityDigest();
    renderAgentCapabilityRoster();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
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
    if ($("agent-continuity-digest")) renderError($("agent-continuity-digest"), error.message);
    if (roster) renderError(roster, error.message);
    renderComparisonWorkbench();
    renderReuseGallery();
  }
}

function agentInfoList(value) {
  return Array.isArray(value) ? value.filter(Boolean) : [];
}

function agentCapabilitySummary(agent) {
  const cfg = agent.Config || agent.config || {};
  const modelCfg = cfg.Agent || cfg.agent || {};
  const toolsCfg = cfg.Tools || cfg.tools || {};
  const heartbeatCfg = cfg.Heartbeat || cfg.heartbeat || {};
  const commands = agent.Commands || agent.commands || {};
  return {
    name: normalizeName(agent),
    description: normalizeDescription(agent) || "No description.",
    model: agent.Model || agent.model || modelCfg.Model || modelCfg.model || "default",
    reasoning: agent.ReasoningEffort || agent.reasoning_effort || modelCfg.ReasoningEffort || modelCfg.reasoning_effort || "default",
    allow: agentInfoList(agent.ToolsAllow || agent.tools_allow || toolsCfg.Allow || toolsCfg.allow),
    deny: agentInfoList(agent.ToolsDeny || agent.tools_deny || toolsCfg.Deny || toolsCfg.deny),
    autoApprove: (agent.AutoApprove ?? agent.auto_approve ?? cfg.AutoApprove ?? cfg.auto_approve) === true,
    heartbeatEvery: agent.HeartbeatEvery || agent.heartbeat_every || heartbeatCfg.Every || heartbeatCfg.every || "",
    heartbeatHours: agent.HeartbeatHours || agent.heartbeat_hours || heartbeatCfg.ActiveHours || heartbeatCfg.active_hours || "",
    heartbeatModel: agent.HeartbeatModel || agent.heartbeat_model || heartbeatCfg.Model || heartbeatCfg.model || "",
    commandCount: Number(agent.CommandCount ?? agent.command_count ?? Object.keys(commands).length ?? 0),
    commandNames: agentInfoList(agent.CommandNames || agent.command_names || Object.keys(commands)),
    hasMemory: (agent.HasMemory ?? agent.has_memory ?? Boolean(agent.Memory || agent.memory)) === true,
  };
}

function runsForAgent(name) {
  return state.runs.filter((run) => String(run?.agent || "") === name);
}

function latestAgentRun(name) {
  return runsForAgent(name)[0] || null;
}

function agentContinuityHint(summary, runs) {
  const latest = runs[0];
  if (!runs.length) return "No recorded runs yet. Start with a focused test or first mission.";
  if (!summary.hasMemory) return "Profile memory is empty. Capture durable role context before complex work.";
  if (runMissionGroup(latest) === "attention") return "Latest run needs review before this agent continues.";
  if (!summary.commandCount) return "No custom commands yet. Add repeatable prompts for this agent.";
  return "Ready to continue from recent work with existing memory and commands.";
}

function renderAgentContinuityDigest() {
  const target = $("agent-continuity-digest");
  if (!target) return;
  if (!state.agents.length) {
    renderEmpty(target, "No agent continuity to summarize yet.");
    return;
  }
  target.innerHTML = state.agents.map((agent) => {
    const summary = agentCapabilitySummary(agent);
    const runs = runsForAgent(summary.name);
    const latest = runs[0] || null;
    const latestStatus = latest ? (latest.status || "unknown") : "none";
    const latestPrompt = latest ? (runPrompt(latest) || latest.id || "Latest run") : "No runs recorded";
    const latestAction = latest ? `<button type="button" data-agent-open-run="${escapeHTML(latest.id)}">Open latest run</button>` : "";
    const memoryLabel = summary.hasMemory ? "Profile memory" : "Memory gap";
    const hint = agentContinuityHint(summary, runs);
    return `<article class="agent-continuity-card ${escapeHTML(runMissionGroup(latest || {}))}">
      <div class="agent-continuity-head">
        <div>
          <strong>${escapeHTML(summary.name)}</strong>
          <span>${escapeHTML(summary.description)}</span>
        </div>
        <span class="tag">${escapeHTML(latestStatus)}</span>
      </div>
      <div class="agent-continuity-metrics">
        <div><span>Runs</span><strong>${runs.length}</strong></div>
        <div><span>Memory</span><strong>${escapeHTML(memoryLabel)}</strong></div>
        <div><span>Commands</span><strong>${summary.commandCount}</strong></div>
      </div>
      <p>${escapeHTML(hint)}</p>
      <small>${escapeHTML(latestPrompt)}</small>
      <div class="row-actions">
        <button type="button" data-agent-continue="${escapeHTML(summary.name)}">Continue</button>
        <button type="button" data-agent-memory-draft="${escapeHTML(summary.name)}">Draft memory</button>
        ${latestAction}
      </div>
    </article>`;
  }).join("");
}

function renderAgentCapabilityRoster() {
  const roster = $("agent-capability-roster");
  if (!roster) return;
  if (!state.agents.length) {
    renderEmpty(roster, "No agent capabilities to map yet.");
    return;
  }
  roster.innerHTML = state.agents.map((agent) => {
    const summary = agentCapabilitySummary(agent);
    const heartbeat = summary.heartbeatEvery
      ? `${summary.heartbeatEvery}${summary.heartbeatHours ? ` · ${summary.heartbeatHours}` : ""}`
      : "off";
    const posture = summary.autoApprove ? "Auto approve" : "Manual review";
    const memory = summary.hasMemory ? "Memory" : "No memory";
    const commandLaunchers = summary.commandNames.length
      ? `<div class="agent-command-launchers" aria-label="${escapeHTML(summary.name)} commands">
        ${summary.commandNames.map((command) => `<button type="button" data-agent-command-launch-agent="${escapeHTML(summary.name)}" data-agent-command-launch="${escapeHTML(command)}">/${escapeHTML(command)}</button>`).join("")}
      </div>`
      : "";
    return `<article class="agent-roster-card">
      <div class="agent-roster-head">
        <span class="agent-orbit-dot" aria-hidden="true"></span>
        <div>
          <strong>${escapeHTML(summary.name)}</strong>
          <small>${escapeHTML(summary.description)}</small>
        </div>
        <span class="tag">${escapeHTML(posture)}</span>
      </div>
      <div class="agent-roster-metrics">
        <div><span>Model</span><strong>${escapeHTML(summary.model)}</strong></div>
        <div><span>Reasoning</span><strong>${escapeHTML(summary.reasoning)}</strong></div>
        <div><span>Allow</span><strong>${summary.allow.length}</strong></div>
        <div><span>Deny</span><strong>${summary.deny.length}</strong></div>
        <div><span>Heartbeat</span><strong>${escapeHTML(heartbeat)}</strong></div>
        <div><span>Commands</span><strong>${summary.commandCount}</strong></div>
      </div>
      <div class="pill-list agent-roster-tags">
        <span>${escapeHTML(memory)}</span>
        <span>${summary.autoApprove ? "Approval bypass" : "Approval gated"}</span>
        <span>${summary.heartbeatEvery ? "Heartbeat scheduled" : "No heartbeat"}</span>
      </div>
      ${commandLaunchers}
      <div class="row-actions">
        <button type="button" data-agent-launch-chat="${escapeHTML(summary.name)}">Chat</button>
        <button type="button" data-agent-launch-test="${escapeHTML(summary.name)}">Test</button>
        <button type="button" data-agent-launch-council="${escapeHTML(summary.name)}">Council</button>
        <button type="button" data-agent-detail="${escapeHTML(summary.name)}">Edit profile</button>
      </div>
    </article>`;
  }).join("");
}

function updateAgentSelects() {
  const options = ['<option value="">Default agent</option>'].concat(
    state.agents.map((agent) => {
      const name = normalizeName(agent);
      return `<option value="${escapeHTML(name)}">${escapeHTML(name)}</option>`;
    })
  ).join("");
  $("chat-agent").innerHTML = options;
  $("home-agent").innerHTML = options;
  $("schedule-agent").innerHTML = options;
  $("agent-test-agent").innerHTML = options;
  $("council-agent").innerHTML = options;
  $("inbox-agent").innerHTML = options;
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
  const allowSet = new Set(payload.tools_allow);
  const conflicts = payload.tools_deny.filter((tool) => allowSet.has(tool));
  const warnings = [];
  if (payload.auto_approve) warnings.push("Auto approve is enabled for this agent.");
  if (conflicts.length) warnings.push(`Allow/deny conflict: ${conflicts.join(", ")}`);
  $("agent-permission-preview").innerHTML = `<div class="agent-preview-row"><strong>Allow</strong><span>${escapeHTML(allow)}</span></div>
    <div class="agent-preview-row"><strong>Deny</strong><span>${escapeHTML(deny)}</span></div>
    <div class="agent-preview-row"><strong>Auto approve</strong><span>${payload.auto_approve ? "Enabled" : "Disabled"}</span></div>
    ${warnings.map((warning) => `<div class="agent-preview-row warning"><strong>Review</strong><span>${escapeHTML(warning)}</span></div>`).join("")}`;
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
  $("agent-test-agent").value = name;
  $("agent-test-prompt").value = `Test ${name}: introduce your role and list one useful next step.`;
  $("agent-test-prompt").focus();
  $("agent-test-state").textContent = `Ready to test ${name}`;
  showToast(`Ready to test ${name}.`);
}

function prepareAgentChat(name) {
  if (!name) return;
  startNewChat();
  $("chat-agent").value = name;
  $("chat-input").value = `Continue as ${name}. Review the current Astria workspace context, identify the next useful action, and call out any risks before changing files.`;
  $("chat-input").focus();
  showToast(`Chat drafted for ${name}.`);
}

function continueAgentFromDigest(name) {
  if (!name) return;
  const summary = agentCapabilitySummary(state.agents.find((agent) => normalizeName(agent) === name) || { name });
  const runs = runsForAgent(name);
  const latest = runs[0] || null;
  startNewChat();
  $("chat-agent").value = name;
  $("chat-input").value = [
    `Continue as ${name} with your current Astria continuity context.`,
    "",
    `Recent runs: ${runs.length}`,
    `Latest run: ${latest ? `${latest.status || "unknown"} · ${runPrompt(latest) || latest.id}` : "none"}`,
    `Profile memory: ${summary.hasMemory ? "present" : "missing"}`,
    `Custom commands: ${summary.commandNames.length ? summary.commandNames.map((command) => `/${command}`).join(", ") : "none"}`,
    "",
    "Summarize what context should carry forward, identify the next concrete action, and name any risk that needs review before acting.",
  ].join("\n");
  $("chat-input").focus();
  showToast(`Continuity prompt drafted for ${name}.`);
}

function prepareAgentTest(name) {
  if (!name) return;
  if (!confirmDiscardAgentChanges()) return;
  $("agent-test-agent").value = name;
  $("agent-test-prompt").value = `Test ${name}: introduce your operating role, summarize your configured strengths, and propose one concrete next step.`;
  $("agent-test-state").textContent = `Ready to test ${name}`;
  switchPanel("agents");
  $("agent-test-prompt").focus();
  showToast(`Test drafted for ${name}.`);
}

function prepareAgentCouncil(name) {
  if (!name) return;
  $("council-agent").value = name;
  $("council-goal").value = `Use ${name} as the lead agent. Split the current Astria task into planner, researcher, and reviewer perspectives, then synthesize a concrete next action.`;
  $("council-state").textContent = `Ready with ${name}`;
  switchPanel("council");
  $("council-goal").focus();
  showToast(`Council drafted for ${name}.`);
}

async function launchAgentCommand(agentName, commandName) {
  if (!agentName || !commandName) return;
  try {
    const detail = await api(`/agents/${encodeURIComponent(agentName)}`);
    const commands = detail.Commands || detail.commands || {};
    const body = commands[commandName] || "";
    if (!body.trim()) {
      showToast(`Command /${commandName} is empty.`);
      return;
    }
    startNewChat();
    $("chat-agent").value = agentName;
    $("chat-input").value = body.trim();
    $("chat-input").focus();
    showToast(`/${commandName} drafted for ${agentName}.`);
  } catch (error) {
    showToast(error.message);
  }
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

function setAgentTestRunning(isRunning) {
  $("agent-test-agent").disabled = isRunning;
  $("agent-test-prompt").disabled = isRunning;
  $("agent-test-submit-button").hidden = isRunning;
  $("agent-test-stop-button").hidden = !isRunning;
  if (isRunning) $("agent-test-state").textContent = "Running";
}

async function submitAgentTest(event) {
  event?.preventDefault();
  if (state.activeAgentTestRequestID) {
    showToast("An agent test is already running.");
    return;
  }
  const text = $("agent-test-prompt").value.trim();
  if (!text) {
    showToast("Enter a test prompt first.");
    return;
  }
  const agent = $("agent-test-agent").value;
  const requestID = globalThis.crypto?.randomUUID ? globalThis.crypto.randomUUID() : `agent-test-${Date.now()}`;
  const payload = {
    text,
    agent,
    new_session: true,
    request_id: requestID,
  };
  const abort = new AbortController();
  state.activeAgentTestRequestID = requestID;
  state.activeAgentTestAbort = abort;
  setAgentTestRunning(true);
  const renderer = beginAgentTestStream();
  try {
    const result = await streamMessage(payload, renderer, abort.signal);
    renderAgentTestResult(result, payload);
    await Promise.allSettled([loadRuns(), loadSessions()]);
    $("agent-test-state").textContent = "Complete";
  } catch (error) {
    if (error.name === "AbortError") {
      $("agent-test-state").textContent = "Cancelled";
      renderAgentTestCancelled(payload);
    } else {
      $("agent-test-state").textContent = "Error";
      renderAgentTestError(error, payload);
    }
  } finally {
    state.activeAgentTestRequestID = "";
    state.activeAgentTestAbort = null;
    setAgentTestRunning(false);
  }
}

function beginAgentTestStream() {
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-stream">
    <div class="run-summary-title">Streaming agent test</div>
    <pre data-agent-test-stream-text></pre>
    <div class="run-timeline" data-agent-test-stream-events></div>
  </div>`;
  const textTarget = $("agent-test-output").querySelector("[data-agent-test-stream-text]");
  const eventsTarget = $("agent-test-output").querySelector("[data-agent-test-stream-events]");
  const events = [];
  return {
    appendText(text) {
      textTarget.textContent += text;
    },
    appendEvent(eventType, data) {
      events.push({ type: eventType || "event", at: new Date().toISOString(), data: data || {} });
      eventsTarget.innerHTML = groupRunTimelineEvents(events).map(renderRunTimelineEntry).join("");
    },
  };
}

function renderAgentTestCancelled(payload) {
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-result">
    <div class="run-summary-title">Agent test cancelled</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "default")}</strong>
      <span>Request</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
    </div>
  </div>`;
}

async function cancelAgentTestRun() {
  const requestID = state.activeAgentTestRequestID;
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
    state.activeAgentTestAbort?.abort();
  }
}

function renderAgentTestResult(result, payload) {
  const usage = result?.usage || {};
  const usageItems = Object.entries(usage).map(([key, value]) => `${key}: ${value}`);
  const usageText = usageItems.length ? usageItems.join(", ") : "-";
  const sessionID = result?.session_id || "";
  const messages = Array.isArray(result?.messages) && result.messages.length
    ? result.messages.join("\n")
    : "No messages returned.";
  const errorHTML = result?.error ? `<div class="error-state">${escapeHTML(result.error)}</div>` : "";
  const openRunAction = payload.request_id
    ? `<button type="button" data-run-summary-run="${escapeHTML(payload.request_id)}">Open run</button>`
    : "";
  const openSessionAction = sessionID
    ? `<button type="button" data-run-summary-session="${escapeHTML(sessionID)}">Open session</button>`
    : "";
  const summaryText = agentTestSummaryText(result, payload);
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-result">
    <div class="run-summary-title">Agent test result</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "default")}</strong>
      <span>Prompt</span><strong>${escapeHTML(payload.text || "-")}</strong>
      <span>Session</span><strong>${escapeHTML(sessionID || "-")}</strong>
      <span>Usage</span><strong>${escapeHTML(usageText)}</strong>
      <span>Request</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
    </div>
    ${errorHTML}
    <pre>${escapeHTML(messages)}</pre>
    <div class="run-summary-actions">
      ${openRunAction}
      ${openSessionAction}
      <button type="button" data-agent-test-copy-summary="${escapeHTML(summaryText)}">Copy summary</button>
    </div>
  </div>`;
}

function renderAgentTestError(error, payload) {
  const summaryText = agentTestSummaryText({ error: error.message || String(error) }, payload);
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-result">
    <div class="run-summary-title">Agent test error</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "default")}</strong>
      <span>Prompt</span><strong>${escapeHTML(payload.text || "-")}</strong>
      <span>Request</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
    </div>
    <div class="error-state">${escapeHTML(error.message || String(error))}</div>
    <div class="run-summary-actions">
      <button type="button" data-agent-test-copy-summary="${escapeHTML(summaryText)}">Copy summary</button>
    </div>
  </div>`;
}

function agentTestSummaryText(result, payload) {
  const usage = result?.usage || {};
  const usageItems = Object.entries(usage).map(([key, value]) => `${key}: ${value}`);
  const messages = Array.isArray(result?.messages) && result.messages.length
    ? result.messages.join("\n")
    : "";
  return [
    "Agent test",
    `Agent: ${payload.agent || "default"}`,
    `Prompt: ${payload.text || ""}`,
    `Request: ${payload.request_id || ""}`,
    `Session: ${result?.session_id || ""}`,
    `Usage: ${usageItems.length ? usageItems.join(", ") : "-"}`,
    result?.error ? `Error: ${result.error}` : "",
    messages ? `Messages:\n${messages}` : "",
  ].filter(Boolean).join("\n");
}

async function loadSkills() {
  const list = $("skills-list");
  try {
    const data = await api("/skills");
    state.skills = data.skills || [];
    setText("manage-skills-count", `${state.skills.length} installed`);
    setText("nav-skills-count", state.skills.length);
    renderManageCount();
    renderHomeDockedTools();
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
      renderMemoryMapPreview();
      renderSourceRegistry();
      renderWorkspaceHub();
      renderKnowledgeCuration();
      renderPromptSuggestionDock();
      renderFocusBrief();
      renderComparisonWorkbench();
      renderReuseGallery();
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
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderWorkspaceHub();
    renderKnowledgeCuration();
    renderPromptSuggestionDock();
    renderFocusBrief();
    renderComparisonWorkbench();
    renderReuseGallery();
  } catch (error) {
    renderError(list, error.message);
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderWorkspaceHub();
    renderKnowledgeCuration();
    renderPromptSuggestionDock();
    renderFocusBrief();
    renderComparisonWorkbench();
    renderReuseGallery();
  }
}

async function loadSchedules() {
  const list = $("schedules-list");
  try {
    const data = await api("/schedules");
    state.schedules = data.schedules || [];
    setText("manage-schedules-count", `${state.schedules.length} configured`);
    setText("nav-schedules-count", state.schedules.length);
    renderManageCount();
    renderHomeDockedTools();
    renderProactiveDeliveryBoard();
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

async function loadRuns() {
  const list = $("runs-list");
  try {
    const data = await api("/runs");
    state.runs = data.runs || [];
    $("runs-count").textContent = state.runs.length;
    renderHomeActivity();
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderAgentContinuityDigest();
    renderRunsList();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderProactiveDeliveryBoard();
    if (state.activeRunID && !state.runs.some((run) => run.id === state.activeRunID)) {
      state.activeRunID = "";
      renderRunDetail(null);
    }
  } catch (error) {
    state.runs = [];
    $("runs-count").textContent = "0";
    renderHomeActivity();
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderAgentContinuityDigest();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderProactiveDeliveryBoard();
    renderError(list, error.message);
  }
}

function renderRunsList() {
  const list = $("runs-list");
  renderMissionControl();
  if (!state.runs.length) {
    renderEmpty(list, "No runs recorded yet.");
    return;
  }
  const runs = filteredRuns();
  if (!runs.length) {
    renderEmpty(list, "No runs match this Mission Control filter.");
    return;
  }
  list.innerHTML = runs.map((run) => {
    const active = run.id === state.activeRunID ? " active" : "";
    const agent = run.agent || "default";
    const session = run.session_id || "no session";
    return `<article class="row-item run-row${active}" data-run-id="${escapeHTML(run.id)}">
      <div class="row-item-title">
        <span>${escapeHTML(run.prompt || run.id)}</span>
        <span class="tag run-status ${escapeHTML(run.status || "unknown")}">${escapeHTML(run.status || "unknown")}</span>
      </div>
      <p>${escapeHTML(agent)} · ${escapeHTML(session)} · ${escapeHTML(formatTimestamp(run.started_at))}</p>
      <div class="row-actions">
        <button type="button" data-run-open="${escapeHTML(run.id)}">Open run</button>
      </div>
    </article>`;
  }).join("");
}

function runMissionGroup(run) {
  const status = String(run?.status || "").toLowerCase();
  if (status === "running" || status === "queued" || status === "pending") return "active";
  if (status === "error" || status === "failed" || status === "cancelled" || status === "canceled") return "attention";
  if (status === "completed" || status === "done" || status === "success") return "completed";
  return "attention";
}

function filteredRuns() {
  switch (state.runFilter) {
    case "active":
    case "attention":
    case "completed":
      return state.runs.filter((run) => runMissionGroup(run) === state.runFilter);
    case "council":
      return state.runs.filter((run) => run.channel === "council_handoff" || String(run.source || "").startsWith("council:"));
    default:
      return state.runs;
  }
}

function renderMissionControl() {
  const board = $("mission-control-board");
  const filters = $("mission-control-filters");
  if (!board || !filters) return;
  const counts = {
    total: state.runs.length,
    active: state.runs.filter((run) => runMissionGroup(run) === "active").length,
    attention: state.runs.filter((run) => runMissionGroup(run) === "attention").length,
    completed: state.runs.filter((run) => runMissionGroup(run) === "completed").length,
    council: state.runs.filter((run) => run.channel === "council_handoff" || String(run.source || "").startsWith("council:")).length,
  };
  board.innerHTML = [
    ["active", "Active", counts.active, "Running or queued work"],
    ["attention", "Needs attention", counts.attention, "Failed, cancelled, or unknown"],
    ["completed", "Completed", counts.completed, "Finished missions"],
    ["total", "Total", counts.total, "All recorded runs"],
  ].map(([key, label, value, hint]) => `<button type="button" class="mission-control-card ${escapeHTML(key)}" data-run-filter="${escapeHTML(key === "total" ? "all" : key)}">
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(String(value))}</strong>
      <small>${escapeHTML(hint)}</small>
    </button>`).join("");
  filters.innerHTML = [
    ["all", "All", counts.total],
    ["active", "Active", counts.active],
    ["attention", "Attention", counts.attention],
    ["completed", "Completed", counts.completed],
    ["council", "Council", counts.council],
  ].map(([key, label, count]) => `<button type="button" class="${state.runFilter === key ? "active" : ""}" data-run-filter="${escapeHTML(key)}">${escapeHTML(label)} <span>${escapeHTML(String(count))}</span></button>`).join("");
}

async function selectRun(runID) {
  if (!runID) return;
  state.activeRunID = runID;
  renderRunsList();
  switchPanel("runs");
  try {
    const run = await api(`/runs/${encodeURIComponent(runID)}`);
    state.activeRunID = run.id || runID;
    renderRunsList();
    renderRunDetail(run);
  } catch (error) {
    $("run-detail-summary").textContent = "Run detail unavailable.";
    renderError($("run-detail"), error.message);
  }
}

function renderRunDetail(run) {
  const target = $("run-detail");
  state.currentRunDetail = run || null;
  if (!run) {
    $("run-detail-summary").textContent = "Select a run to inspect request, result, and events.";
    renderEmpty(target, "No run selected.");
    return;
  }
  const usage = run.usage || run.response?.usage || {};
  const usageText = formatUsage(usage);
  const sessionID = runSessionID(run);
  const prompt = runPrompt(run);
  $("run-detail-summary").textContent = `${run.status || "unknown"} · ${formatTimestamp(run.started_at)}`;
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <div class="run-meta-grid">
        <span>ID</span><strong>${escapeHTML(run.id || "-")}</strong>
        <span>Status</span><strong>${escapeHTML(run.status || "-")}</strong>
        <span>Agent</span><strong>${escapeHTML(run.agent || "default")}</strong>
        <span>Channel</span><strong>${escapeHTML(run.channel || "-")}</strong>
        <span>Session</span><strong>${escapeHTML(sessionID || "-")}</strong>
        <span>Started</span><strong>${escapeHTML(formatTimestamp(run.started_at))}</strong>
        <span>Ended</span><strong>${escapeHTML(formatTimestamp(run.ended_at))}</strong>
        <span>Usage</span><strong>${escapeHTML(usageText)}</strong>
      </div>
      <div class="run-detail-actions">
        <button type="button" data-run-detail-copy-summary>Copy summary</button>
        <button type="button" data-run-detail-copy-prompt>Copy prompt</button>
        <button type="button" data-run-detail-copy-result>Copy result</button>
        ${prompt ? `<button type="button" data-run-detail-follow-up>Suggest follow-up</button>` : ""}
        ${sessionID ? `<button type="button" data-run-detail-open-session="${escapeHTML(sessionID)}">Open session</button>` : ""}
        ${prompt ? `<button type="button" data-run-detail-rerun>Re-run</button>` : ""}
      </div>
      ${run.error ? `<div class="error-state">${escapeHTML(run.error)}</div>` : ""}
    </section>
    <section class="run-detail-section">
      <h3>Prompt</h3>
      <pre>${escapeHTML(prompt)}</pre>
    </section>
    <section class="run-detail-section">
      <h3>Result</h3>
      <pre>${escapeHTML(formatRunResponse(run.response))}</pre>
    </section>
    <section class="run-detail-section">
      <h3>Time Travel</h3>
      ${renderRunTimeline(run)}
    </section>
  </div>`;
}

function renderRunTimeline(run) {
  const entries = buildRunTimelineEntries(run);
  if (!entries.length) return `<div class="empty-state">No timeline data captured for this run.</div>`;
  return `<div class="run-timeline">${entries.map(renderRunTimelineEntry).join("")}</div>`;
}

function buildRunTimelineEntries(run) {
  if (!run) return [];
  const entries = [];
  const prompt = runPrompt(run);
  const sessionID = runSessionID(run);
  const usage = run.usage || run.response?.usage || {};
  entries.push({
    kind: "milestone",
    tone: runHealthGroup(run),
    at: run.started_at,
    title: `Run ${run.status || "recorded"}`,
    detail: `${run.agent || "default"} · ${run.channel || "local"} · ${run.id || "run"}`,
  });
  if (prompt) {
    entries.push({
      kind: "milestone",
      tone: "prompt",
      at: run.started_at,
      title: "Prompt locked",
      detail: prompt,
    });
  }
  if (sessionID) {
    entries.push({
      kind: "milestone",
      tone: "session",
      at: run.started_at,
      title: "Session linked",
      detail: sessionID,
      sessionID,
    });
  }
  entries.push(...groupRunTimelineEvents(run.events || []));
  if (Object.keys(usage).length && !(run.events || []).some((event) => event.type === "usage")) {
    entries.push({
      kind: "usage",
      at: run.ended_at || run.started_at,
      data: usage,
    });
  }
  if (run.ended_at || run.error || run.response) {
    entries.push({
      kind: "milestone",
      tone: run.error || run.response?.error ? "failed" : "completed",
      at: run.ended_at || run.started_at,
      title: run.error || run.response?.error ? "Run needs review" : "Run finished",
      detail: run.error || run.response?.error || formatRunResponse(run.response) || run.status || "Completed",
    });
  }
  return entries;
}

function renderRunEvents(events) {
  if (!events.length) return `<div class="empty-state">No events captured for this run.</div>`;
  const entries = groupRunTimelineEvents(events);
  return `<div class="run-timeline">${entries.map(renderRunTimelineEntry).join("")}</div>`;
}

function groupRunTimelineEvents(events) {
  const entries = [];
  const openTools = new Map();
  for (const event of events) {
    const data = event.data || {};
    const tool = data.tool || "tool";
    if (event.type === "tool_call") {
      const entry = {
        kind: "tool",
        at: event.at,
        tool,
        status: data.status || "running",
        args: data.args || "",
        result: "",
        isError: false,
        errorCategory: "",
      };
      entries.push(entry);
      openTools.set(tool, entry);
      continue;
    }
    if (event.type === "tool_result") {
      const entry = openTools.get(tool) || {
        kind: "tool",
        at: event.at,
        tool,
        status: "",
        args: "",
        result: "",
        isError: false,
        errorCategory: "",
      };
      if (!openTools.has(tool)) entries.push(entry);
      entry.status = data.status || (data.is_error ? "error" : "completed");
      entry.result = data.content || "";
      entry.isError = data.is_error === true;
      entry.errorCategory = data.error_category || "";
      openTools.delete(tool);
      continue;
    }
    entries.push({ kind: event.type || "event", at: event.at, data });
  }
  return entries;
}

function renderRunTimelineEntry(entry) {
  if (entry.kind === "milestone") {
    const action = entry.sessionID
      ? `<button type="button" data-run-detail-open-session="${escapeHTML(entry.sessionID)}">Open linked session</button>`
      : "";
    return `<article class="run-event run-milestone ${escapeHTML(entry.tone || "")}">
      <div class="run-event-header">
        <strong>${escapeHTML(entry.title || "Milestone")}</strong>
        <span>${escapeHTML(formatTimestamp(entry.at))}</span>
      </div>
      ${action ? `<div class="run-event-actions">${action}</div>` : ""}
      <p>${escapeHTML(entry.detail || "")}</p>
    </article>`;
  }
  if (entry.kind === "tool") {
    const status = entry.status || (entry.isError ? "error" : "completed");
    const resultText = entry.result ? formatToolPayload(entry.result) : "";
    const resultAction = resultText
      ? `<button type="button" data-run-tool-copy-result="${escapeHTML(resultText)}">Copy result</button>`
      : "";
    return `<article class="run-event run-tool-event ${entry.isError ? "bad" : ""}">
      <div class="run-event-header">
        <strong>${escapeHTML(entry.tool)}</strong>
        <span>${escapeHTML(status)} · ${escapeHTML(formatTimestamp(entry.at))}</span>
      </div>
      ${resultAction ? `<div class="run-event-actions">${resultAction}</div>` : ""}
      <div class="run-tool-grid">
        ${entry.args ? `<div><span>Args</span><pre>${escapeHTML(formatToolPayload(entry.args))}</pre></div>` : ""}
        ${resultText ? `<div><span>Result</span><pre>${escapeHTML(resultText)}</pre></div>` : ""}
        ${entry.errorCategory ? `<div class="tool-meta">category: ${escapeHTML(entry.errorCategory)}</div>` : ""}
      </div>
    </article>`;
  }
  const label = runEventLabel(entry.kind);
  return `<article class="run-event">
    <div class="run-event-header">
      <strong>${escapeHTML(label)}</strong>
      <span>${escapeHTML(formatTimestamp(entry.at))}</span>
    </div>
    <pre>${escapeHTML(formatToolPayload(entry.data || {}))}</pre>
  </article>`;
}

function runEventLabel(type) {
  switch (type) {
    case "text":
      return "Text";
    case "preamble":
      return "Preamble";
    case "usage":
      return "Usage";
    case "approval_needed":
      return "Approval needed";
    case "approval_resolved":
      return "Approval resolved";
    default:
      return type || "Event";
  }
}

function runSessionID(run) {
  return run?.session_id || run?.response?.session_id || "";
}

function runPrompt(run) {
  return run?.prompt || run?.request?.text || "";
}

function formatUsage(usage) {
  return usage && Object.keys(usage).length
    ? Object.entries(usage).map(([key, value]) => `${key}: ${value}`).join(", ")
    : "-";
}

function runSummaryText(run) {
  const usage = run?.usage || run?.response?.usage || {};
  return [
    `Run: ${run?.id || "-"}`,
    `Status: ${run?.status || "-"}`,
    `Agent: ${run?.agent || "default"}`,
    `Session: ${runSessionID(run) || "-"}`,
    `Usage: ${formatUsage(usage)}`,
    `Prompt: ${runPrompt(run) || "-"}`,
  ].join("\n");
}

function runFollowUpPrompt(run) {
  const prompt = runPrompt(run);
  const result = runResultText(run);
  const usage = run?.usage || run?.response?.usage || {};
  return [
    "Continue from this completed Astria run. Summarize what was achieved, identify unresolved risks, and propose the next concrete action with validation.",
    "",
    `Run: ${run?.id || "-"}`,
    `Status: ${run?.status || "-"}`,
    `Agent: ${run?.agent || "default"}`,
    `Session: ${runSessionID(run) || "-"}`,
    `Usage: ${formatUsage(usage)}`,
    `Original prompt: ${prompt || "-"}`,
    "",
    "Result preview:",
    String(result || "").slice(0, 1600),
  ].join("\n");
}

function runResultText(run) {
  return formatRunResponse(run?.response);
}

function rerunCurrentRun() {
  const run = state.currentRunDetail;
  if (!run) return;
  const prompt = runPrompt(run);
  if (!prompt) {
    showToast("Run prompt is empty.");
    return;
  }
  state.activeSessionID = "";
  $("chat-new-session").checked = true;
  $("chat-input").value = prompt;
  const agent = run.agent || "";
  if (agent && agent !== "default" && [...$("chat-agent").options].some((option) => option.value === agent)) {
    $("chat-agent").value = agent;
  } else {
    $("chat-agent").value = "";
  }
  updateActiveSessionLabel();
  document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
  renderEmptyThread();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Run copied to chat.");
}

function formatRunResponse(response) {
  if (!response) return "No response recorded.";
  if (Array.isArray(response.messages) && response.messages.length) {
    return response.messages.join("\n");
  }
  return JSON.stringify(response, null, 2);
}

function formatTimestamp(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString();
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
  switchPanel("chat");
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
    const result = await streamMessage(payload, chatStreamRenderer(assistantMessage), abort.signal);
    renderDoneResult(result, assistantMessage);
    renderRunSummary(result, payload);
    if (result?.session_id) {
      state.activeSessionID = result.session_id;
      $("chat-new-session").checked = false;
    }
    stateLabel.textContent = "Complete";
    await Promise.allSettled([loadSessions(), loadRuns()]);
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
      return "Astria";
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
  renderHomeActivity();
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
  state.approvals.delete(requestID);
  $("chat-state").textContent = status === "allowed" ? "Approval allowed" : "Approval denied";
  renderHomeActivity();
  renderApprovalCenter();
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
  const run = {
    id: requestID,
    status: "completed",
    agent,
    session_id: sessionID,
    prompt: payload.text || "",
    response: result,
    usage,
  };
  const summaryText = [
    `Session: ${sessionID || "-"}`,
    `Agent: ${agent}`,
    `Usage: ${usageText}`,
    `Request: ${requestID}`,
  ].join("\n");
  const card = document.createElement("div");
  card.className = "run-summary";
  const openSessionAction = sessionID
    ? `<button type="button" data-run-summary-session="${escapeHTML(sessionID)}">Open session</button>`
    : "";
  const openRunAction = requestID && requestID !== "-"
    ? `<button type="button" data-run-summary-run="${escapeHTML(requestID)}">Open run</button>`
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
      <button type="button" data-run-follow-up="${escapeHTML(runFollowUpPrompt(run))}">Suggest follow-up</button>
      ${openRunAction}
      ${openSessionAction}
    </div>`;
  $("chat-output").appendChild(card);
}

function chatStreamRenderer(assistantMessage) {
  return {
    appendText(text) {
      appendAssistantText(assistantMessage, text);
    },
    appendEvent(eventType, data) {
      if (eventType === "tool_call" || eventType === "tool_result") {
        appendToolEvent(data, eventType);
      }
    },
    scroll() {
      scrollConversationToBottom();
    },
  };
}

async function streamMessage(payload, renderer, signal) {
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
      doneResult = handleSSEEvent(rawEvent, renderer, doneResult);
    }
  }
  if (buffer.trim()) {
    doneResult = handleSSEEvent(buffer, renderer, doneResult);
  }
  return doneResult;
}

function handleSSEEvent(rawEvent, renderer, doneResult) {
  const parsed = parseSSE(rawEvent);
  if (!parsed) return doneResult;
  const data = parseEventData(parsed.data);
  switch (parsed.event) {
    case "text":
      renderer.appendText?.(data.text || "");
      break;
    case "preamble":
      renderer.appendText?.(data.preamble || "");
      break;
    case "usage":
    case "tool_call":
    case "tool_result":
      renderer.appendEvent?.(parsed.event, data);
      break;
    case "done":
      doneResult = data;
      break;
    case "error":
      throw new Error(data.error || "stream failed");
  }
  renderer.scroll?.();
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

function submitHomeTask(event) {
  event.preventDefault();
  const text = $("home-task-input").value.trim();
  if (!text) {
    showToast("Enter a mission first.");
    return;
  }
  state.workflowStage = "running";
  state.workflowStageLabel = text;
  renderWorkflowStageRail();
  $("chat-input").value = text;
  $("chat-agent").value = $("home-agent").value;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  document.querySelectorAll("[data-session-id]").forEach((item) => item.classList.remove("active"));
  updateActiveSessionLabel();
  $("home-task-input").value = "";
  $("chat-form").requestSubmit();
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
    loadVersion(),
    loadDiagnostics(),
    loadConfig(),
    loadPermissions(),
    loadMemory(),
    loadCouncilRuns(),
    loadInbox(),
    loadInboxProviders(),
    loadAgents(),
    loadSkills(),
    loadSessions(),
    loadSchedules(),
    loadRuns(),
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

  const mcpTest = event.target.closest("[data-mcp-test]");
  if (mcpTest) {
    testMCPServer(mcpTest.dataset.mcpTest);
    return;
  }

  const mcpEdit = event.target.closest("[data-mcp-edit]");
  if (mcpEdit) {
    editMCPServer(mcpEdit.dataset.mcpEdit);
    return;
  }

  const mcpToggle = event.target.closest("[data-mcp-toggle]");
  if (mcpToggle) {
    toggleMCPServer(mcpToggle.dataset.mcpToggle);
    return;
  }

  const compareSelect = event.target.closest("[data-compare-select]");
  if (compareSelect) {
    state.selectedComparisonLane = compareSelect.dataset.compareSelect || "";
    renderComparisonWorkbench();
    return;
  }

  const compareDraft = event.target.closest("[data-compare-draft]");
  if (compareDraft) {
    draftComparisonToChat(compareDraft.dataset.compareDraft);
    return;
  }

  const compareLane = event.target.closest("[data-compare-lane]");
  if (compareLane && !event.target.closest("button")) {
    state.selectedComparisonLane = compareLane.dataset.compareLane || "";
    renderComparisonWorkbench();
    return;
  }

  const promptVariantSelect = event.target.closest("[data-prompt-variant-select]");
  if (promptVariantSelect) {
    state.selectedPromptVariant = promptVariantSelect.dataset.promptVariantSelect || "";
    renderPromptExperimentLab();
    return;
  }

  const promptVariantDraft = event.target.closest("[data-prompt-variant-draft]");
  if (promptVariantDraft) {
    draftPromptVariantToChat(promptVariantDraft.dataset.promptVariantDraft);
    return;
  }

  const promptVariant = event.target.closest("[data-prompt-variant]");
  if (promptVariant && !event.target.closest("button")) {
    state.selectedPromptVariant = promptVariant.dataset.promptVariant || "";
    renderPromptExperimentLab();
    return;
  }

  const reuseSelect = event.target.closest("[data-reuse-select]");
  if (reuseSelect) {
    state.selectedReuseAsset = reuseSelect.dataset.reuseSelect || "";
    renderReuseGallery();
    return;
  }

  const reuseDraft = event.target.closest("[data-reuse-draft]");
  if (reuseDraft) {
    draftReuseAssetToChat(reuseDraft.dataset.reuseDraft);
    return;
  }

  const reuseAsset = event.target.closest("[data-reuse-asset]");
  if (reuseAsset && !event.target.closest("button")) {
    state.selectedReuseAsset = reuseAsset.dataset.reuseAsset || "";
    renderReuseGallery();
    return;
  }

  const deliverySelect = event.target.closest("[data-delivery-select]");
  if (deliverySelect) {
    state.selectedDeliveryLane = deliverySelect.dataset.deliverySelect || "";
    renderProactiveDeliveryBoard();
    return;
  }

  const deliveryDraft = event.target.closest("[data-delivery-draft]");
  if (deliveryDraft) {
    draftDeliveryToChat(deliveryDraft.dataset.deliveryDraft);
    return;
  }

  const deliveryLane = event.target.closest("[data-delivery-lane]");
  if (deliveryLane && !event.target.closest("button")) {
    state.selectedDeliveryLane = deliveryLane.dataset.deliveryLane || "";
    renderProactiveDeliveryBoard();
    return;
  }

  const sourceSelect = event.target.closest("[data-source-select]");
  if (sourceSelect) {
    state.selectedSourceRow = sourceSelect.dataset.sourceSelect || "";
    renderSourceRegistry();
    return;
  }

  const sourceDraft = event.target.closest("[data-source-draft]");
  if (sourceDraft) {
    draftSourceMaintenanceToChat(sourceDraft.dataset.sourceDraft);
    return;
  }

  const sourceRow = event.target.closest("[data-source-row]");
  if (sourceRow && !event.target.closest("button")) {
    state.selectedSourceRow = sourceRow.dataset.sourceRow || "";
    renderSourceRegistry();
    return;
  }

  const action = event.target.closest("[data-action]");
  if (action?.dataset.action === "mcp-new") {
    beginMCPCreate();
    return;
  }

  const memoryDelete = event.target.closest("[data-memory-delete]");
  if (memoryDelete) {
    deleteMemoryEntry(memoryDelete.dataset.memoryDelete);
    return;
  }

  const memoryDraft = event.target.closest("[data-memory-draft]");
  if (memoryDraft) {
    draftMemoryCandidate(memoryDraft.dataset.memoryDraft);
    return;
  }

  const memoryCategory = event.target.closest("[data-memory-category]");
  if (memoryCategory) {
    state.memoryCategory = memoryCategory.dataset.memoryCategory || "all";
    renderMemoryMapPreview();
    renderSourceRegistry();
    return;
  }

  const councilOpen = event.target.closest("[data-council-open]");
  if (councilOpen) {
    renderCouncilDetail(councilRunByID(councilOpen.dataset.councilOpen));
    return;
  }

  const councilCopy = event.target.closest("[data-council-copy]");
  if (councilCopy) {
    copyCouncilSynthesis(councilCopy.dataset.councilCopy, councilCopy);
    return;
  }

  const councilRoleCopy = event.target.closest("[data-council-role-copy]");
  if (councilRoleCopy) {
    copyCouncilRoleNotes(councilRoleCopy.dataset.councilRoleCopy, councilRoleCopy.dataset.councilRoleIndex, councilRoleCopy);
    return;
  }

  const councilRoleDraft = event.target.closest("[data-council-role-draft]");
  if (councilRoleDraft) {
    draftCouncilRoleToChat(councilRoleDraft.dataset.councilRoleDraft, councilRoleDraft.dataset.councilRoleIndex);
    return;
  }

  const councilSend = event.target.closest("[data-council-send]");
  if (councilSend) {
    sendCouncilToChat(councilSend.dataset.councilSend);
    return;
  }

  const councilRun = event.target.closest("[data-council-run]");
  if (councilRun) {
    runCouncilSynthesis(councilRun.dataset.councilRun);
    return;
  }

  const inboxApprove = event.target.closest("[data-inbox-approve]");
  if (inboxApprove) {
    approveInboxItem(inboxApprove.dataset.inboxApprove);
    return;
  }

  const inboxReject = event.target.closest("[data-inbox-reject]");
  if (inboxReject) {
    rejectInboxItem(inboxReject.dataset.inboxReject);
    return;
  }

  const inboxRetry = event.target.closest("[data-inbox-retry]");
  if (inboxRetry) {
    retryInboxItem(inboxRetry.dataset.inboxRetry);
    return;
  }

  const inboxRun = event.target.closest("[data-inbox-run]");
  if (inboxRun) {
    openInboxRun(inboxRun.dataset.inboxRun);
    return;
  }

  const commandItem = event.target.closest("[data-command-id]");
  if (commandItem) {
    runCommandCenterItem(commandItem.dataset.commandId);
    return;
  }

  const commandClose = event.target.closest("[data-command-close]");
  if (commandClose) {
    closeCommandCenter();
    return;
  }

  const sessionResume = event.target.closest("[data-session-resume]");
  if (sessionResume) {
    selectSession(sessionResume.dataset.sessionResume);
    return;
  }

  const nav = event.target.closest("[data-panel]");
  if (nav) {
    switchPanel(nav.dataset.panel);
    return;
  }

  const homeAction = event.target.closest("[data-home-action]");
  if (homeAction) {
    runHomeAction(homeAction.dataset.homeAction);
    return;
  }

  const strategy = event.target.closest("[data-strategy]");
  if (strategy) {
    selectWorkflowStrategy(strategy.dataset.strategy);
    return;
  }

  const recipe = event.target.closest("[data-recipe]");
  if (recipe) {
    selectWorkflowRecipe(recipe.dataset.recipe);
    return;
  }

  const promptButton = event.target.closest("[data-home-prompt]");
  if (promptButton) {
    seedMissionPrompt(promptButton.dataset.homePrompt || "");
    return;
  }

  const sessionItem = event.target.closest("[data-session-id]");
  if (sessionItem) selectSession(sessionItem.dataset.sessionId);

  const runSummarySession = event.target.closest("[data-run-summary-session]");
  if (runSummarySession) {
    selectSession(runSummarySession.dataset.runSummarySession);
    return;
  }

  const runSummaryRun = event.target.closest("[data-run-summary-run]");
  if (runSummaryRun) {
    selectRun(runSummaryRun.dataset.runSummaryRun);
    return;
  }

  const runFilter = event.target.closest("[data-run-filter]");
  if (runFilter) {
    state.runFilter = runFilter.dataset.runFilter || "all";
    renderRunsList();
    return;
  }

  const runSummaryCopy = event.target.closest("[data-run-summary-copy]");
  if (runSummaryCopy) {
    copyText(runSummaryCopy.dataset.runSummaryCopy, "Run summary copied.")
      .then(() => markButtonCopied(runSummaryCopy))
      .catch((error) => showToast(error.message));
    return;
  }

  const runFollowUp = event.target.closest("[data-run-follow-up]");
  if (runFollowUp) {
    seedMissionPrompt(runFollowUp.dataset.runFollowUp || "");
    showToast("Follow-up prompt drafted.");
    return;
  }

  const agentTestCopySummary = event.target.closest("[data-agent-test-copy-summary]");
  if (agentTestCopySummary) {
    copyText(agentTestCopySummary.dataset.agentTestCopySummary, "Agent test summary copied.")
      .then(() => markButtonCopied(agentTestCopySummary))
      .catch((error) => showToast(error.message));
    return;
  }

  const runDetailCopySummary = event.target.closest("[data-run-detail-copy-summary]");
  if (runDetailCopySummary) {
    copyText(runSummaryText(state.currentRunDetail), "Run summary copied.")
      .then(() => markButtonCopied(runDetailCopySummary))
      .catch((error) => showToast(error.message));
    return;
  }

  const runDetailCopyPrompt = event.target.closest("[data-run-detail-copy-prompt]");
  if (runDetailCopyPrompt) {
    copyText(runPrompt(state.currentRunDetail), "Prompt copied.")
      .then(() => markButtonCopied(runDetailCopyPrompt))
      .catch((error) => showToast(error.message));
    return;
  }

  const runDetailCopyResult = event.target.closest("[data-run-detail-copy-result]");
  if (runDetailCopyResult) {
    copyText(runResultText(state.currentRunDetail), "Result copied.")
      .then(() => markButtonCopied(runDetailCopyResult))
      .catch((error) => showToast(error.message));
    return;
  }

  const runDetailFollowUp = event.target.closest("[data-run-detail-follow-up]");
  if (runDetailFollowUp) {
    seedMissionPrompt(runFollowUpPrompt(state.currentRunDetail));
    showToast("Follow-up prompt drafted.");
    return;
  }

  const runToolCopyResult = event.target.closest("[data-run-tool-copy-result]");
  if (runToolCopyResult) {
    copyText(runToolCopyResult.dataset.runToolCopyResult || "", "Tool result copied.")
      .then(() => markButtonCopied(runToolCopyResult))
      .catch((error) => showToast(error.message));
    return;
  }

  const runDetailOpenSession = event.target.closest("[data-run-detail-open-session]");
  if (runDetailOpenSession) {
    selectSession(runDetailOpenSession.dataset.runDetailOpenSession);
    return;
  }

  const runDetailRerun = event.target.closest("[data-run-detail-rerun]");
  if (runDetailRerun) {
    rerunCurrentRun();
    return;
  }

  const agentLaunchChat = event.target.closest("[data-agent-launch-chat]");
  if (agentLaunchChat) {
    prepareAgentChat(agentLaunchChat.dataset.agentLaunchChat);
    return;
  }

  const agentContinue = event.target.closest("[data-agent-continue]");
  if (agentContinue) {
    continueAgentFromDigest(agentContinue.dataset.agentContinue);
    return;
  }

  const agentMemoryDraft = event.target.closest("[data-agent-memory-draft]");
  if (agentMemoryDraft) {
    draftAgentMemoryCandidate(agentMemoryDraft.dataset.agentMemoryDraft);
    return;
  }

  const agentOpenRun = event.target.closest("[data-agent-open-run]");
  if (agentOpenRun) {
    selectRun(agentOpenRun.dataset.agentOpenRun);
    return;
  }

  const agentLaunchTest = event.target.closest("[data-agent-launch-test]");
  if (agentLaunchTest) {
    prepareAgentTest(agentLaunchTest.dataset.agentLaunchTest);
    return;
  }

  const agentLaunchCouncil = event.target.closest("[data-agent-launch-council]");
  if (agentLaunchCouncil) {
    prepareAgentCouncil(agentLaunchCouncil.dataset.agentLaunchCouncil);
    return;
  }

  const agentCommandLaunch = event.target.closest("[data-agent-command-launch]");
  if (agentCommandLaunch) {
    launchAgentCommand(agentCommandLaunch.dataset.agentCommandLaunchAgent, agentCommandLaunch.dataset.agentCommandLaunch);
    return;
  }

  const agentButton = event.target.closest("[data-agent-detail]");
  if (agentButton) inspectAgent(agentButton.dataset.agentDetail);

  const agentCommand = event.target.closest("[data-agent-command]");
  if (agentCommand) selectAgentCommand(agentCommand.dataset.agentCommand);

  const runOpen = event.target.closest("[data-run-open]");
  if (runOpen) {
    selectRun(runOpen.dataset.runOpen);
    return;
  }

  const runRow = event.target.closest("[data-run-id]");
  if (runRow) {
    selectRun(runRow.dataset.runId);
    return;
  }

  const toggle = event.target.closest("[data-schedule-toggle]");
  if (toggle) toggleSchedule(toggle.dataset.scheduleToggle, toggle.dataset.enabled === "true");

  const remove = event.target.closest("[data-schedule-delete]");
  if (remove) deleteSchedule(remove.dataset.scheduleDelete);
});

$("refresh-button").addEventListener("click", refreshAll);
$("new-chat-button").addEventListener("click", startNewChat);
$("command-center-button").addEventListener("click", openCommandCenter);
$("command-center-input").addEventListener("input", renderCommandCenterList);
document.addEventListener("keydown", (event) => {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
    event.preventDefault();
    openCommandCenter();
    return;
  }
  if (event.key === "Escape" && !$("command-center")?.hidden) {
    event.preventDefault();
    closeCommandCenter();
  }
});
$("home-task-form").addEventListener("submit", submitHomeTask);
$("home-agent").addEventListener("change", () => {
  $("chat-agent").value = $("home-agent").value;
  updateSelectedAgent();
});
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
$("mcp-form").addEventListener("submit", submitMCPServer);
$("mcp-type").addEventListener("change", updateMCPTransportFields);
$("mcp-new-button").addEventListener("click", beginMCPCreate);
$("mcp-clear-button").addEventListener("click", beginMCPCreate);
$("intake-form").addEventListener("submit", submitFileIntake);
$("intake-analyze-button").addEventListener("click", submitFileIntake);
$("intake-chat-button").addEventListener("click", sendIntakeToChat);
$("intake-extract-button").addEventListener("click", draftArchiveExtractRun);
$("permissions-form").addEventListener("submit", submitPermissions);
$("permissions-form").addEventListener("input", renderPermissionsPendingPreview);
$("permissions-clear-button").addEventListener("click", clearPermissions);
$("memory-review-form").addEventListener("submit", submitMemoryCandidate);
$("memory-candidate").addEventListener("input", renderMemoryCandidatePreview);
$("council-form").addEventListener("submit", submitCouncilRun);
$("inbox-webhook-form").addEventListener("submit", submitInboxWebhook);
$("update-check-button").addEventListener("click", checkForUpdates);
$("copy-support-info-button").addEventListener("click", copySupportInfo);
$("agent-form").addEventListener("submit", submitAgent);
$("agent-test-submit-button").addEventListener("click", submitAgentTest);
$("agent-test-prompt").addEventListener("keydown", (event) => {
  if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    submitAgentTest(event);
  }
});
$("agent-test-stop-button").addEventListener("click", cancelAgentTestRun);
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
$("promptlab-goal").addEventListener("input", (event) => {
  state.promptLabGoal = event.target.value;
  renderPromptExperimentLab();
  renderReuseGallery();
});

renderHomeMode();
renderStrategyMatrix();
renderFileIntake();
renderSourceRegistry();
renderWorkspaceHub();
renderKnowledgeCuration();
renderToolDockInspector();
renderPromptSuggestionDock();
renderFocusBrief();
renderApprovalCenter();
renderWorkspaceHealthStrip();
renderComparisonWorkbench();
renderPromptExperimentLab();
renderReuseGallery();
renderProactiveDeliveryBoard();
connectEventStream();
refreshAll();
