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
  currentRunTrace: [],
  currentRunTraceError: "",
  missionRunDetail: null,
  missionRunTrace: [],
  missionRunTraceError: "",
  missionRunHydrating: "",
  currentCouncilRun: null,
  selectedComparisonLane: "",
  selectedRunQuality: "",
  selectedPromptVariant: "",
  selectedBudgetGuard: "",
  selectedReuseAsset: "",
  selectedResultArchive: "",
  selectedPlaybook: "",
  selectedStarterKit: "",
  selectedSharePack: "",
  selectedWorkspaceSnapshot: "",
  sharePackName: "",
  sharePackAudience: "",
  sharePackIntent: "",
  selectedBrowserMission: "",
  browserTargetURL: "",
  browserMissionGoal: "",
  selectedDataInsight: "",
  dataSourceDescriptor: "",
  dataAnalysisQuestion: "",
  dataOutputFormat: "",
  selectedDeliveryLane: "",
  selectedSourceRow: "",
  selectedReconcileRisk: "",
  selectedCitationGrounding: "",
  citationClaimScope: "",
  citationSourcePosture: "",
  citationEvidenceLevel: "",
  promptLabGoal: "",
  diagnostics: null,
  config: null,
  permissions: null,
  memory: null,
  memoryStatus: null,
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
  liveRun: {
    visible: false,
    state: "idle",
    runID: "",
    sessionID: "",
    usage: null,
    latest: "就绪",
  },
  starMap: {
    state: "calm",
    active: "target",
    timer: null,
    signature: "",
    decayTimer: null,
  },
  railCollapsed: false,
  activeSessionID: "",
  toolEvents: new Map(),
  toolDetails: new Map(),
  approvals: new Map(),
  missionEvents: [],
  eventSource: null,
  eventStream: {
    lastEventID: "",
    status: "idle",
    reconnects: 0,
    reconnecting: false,
    lastRecoveredAt: "",
    refreshingRuns: false,
  },
  homeMode: "general",
  workflowStrategy: "direct",
  workflowStage: "draft",
  workflowStageLabel: "通用任务",
  memoryCategory: "all",
  editingMCPServer: "",
  runFilter: "all",
};

const views = {
  home: ["任务台", "从 Astria 启动本地 Agent 任务。"],
  chat: ["对话", "通过本地 daemon 与 Astria 协作。"],
  manage: ["更多功能", "管理 Agent 资源、本地自动化和上下文星图。"],
  settings: ["系统", "检查 daemon 设置、权限和构建状态。"],
  agents: ["智能体", "检查 daemon 可用的命名 Agent。"],
  skills: ["技能", "查看 Astria 已加载技能。"],
  mcp: ["MCP 星港", "检查 MCP 服务器和工具 dock 就绪状态。"],
  memory: ["记忆星图", "复查来源会话并起草记忆候选。"],
  sources: ["来源登记", "检查知识来源的鲜度和可靠性。"],
  reconcile: ["知识校验", "解决过期、冲突、薄弱或敏感知识。"],
  citation: ["引用校准", "规划来源覆盖、引用和证据缺口。"],
  council: ["智能体议会", "协调规划者、研究者和审阅者角色。"],
  quality: ["运行质量", "按证据、预算姿态、风险和下一步评估运行。"],
  compare: ["路径比较台", "比较运行、Agent、记忆和评议证据。"],
  promptlab: ["Prompt 实验室", "跨 Agent、证据和交付路径测试 Prompt 变体。"],
  budget: ["预算守卫", "规划 token 上限、模型 fallback、复杂度路由和停止规则。"],
  reuse: ["复用星库", "从可复用 Prompt、Agent、来源和产物启动。"],
  results: ["产物星库", "复查已保存报告、证据简报和可复用产物。"],
  playbooks: ["实践手册", "从审核过的本地最佳实践路径启动。"],
  starter: ["启动套件", "从预制 Astria 工作流套件启动。"],
  share: ["交接包", "把本地工作打包成可审核、可复制的交接材料。"],
  snapshot: ["工作区快照", "规划本地续接、证据、来源和隐私快照包。"],
  browser: ["浏览器规划器", "规划可审核的浏览器检查和证据任务。"],
  data: ["数据规划器", "规划可审核的数据分析和知识沉淀任务。"],
  delivery: ["主动交付", "监控定时任务和出站渠道就绪状态。"],
  inbox: ["收件箱", "执行前审核进入渠道的任务。"],
  intake: ["文件星舱", "在运行前检查本地文档和归档。"],
  schedules: ["定时任务", "创建和管理 cron 本地任务。"],
  runs: ["运行", "观测最近的 daemon 执行。"],
  diagnostics: ["诊断", "检查 daemon 就绪状态和设置检查。"],
  config: ["连接器", "修复 daemon 运行所需的 provider 设置。"],
  permissions: ["权限", "复查本地工具策略。"],
  version: ["版本", "检查构建和更新状态。"],
};

const homeActions = {
  publish: {
    title: "发布资源",
    status: "就绪",
    description: "整理当前项目中可以交付或发布的资源，并生成打包清单。",
    prompt: "Prepare a publishable resource from the current project and list the files you would package.",
    notice: "已为发布资源任务预填提示，可以直接启动。",
  },
  browser: {
    title: "浏览器检查",
    status: "就绪",
    description: "用于网页检查、截图、表单验证和变更摘要；从 Chat 中按需请求浏览器操作。",
    prompt: "Use browser automation to inspect the relevant page and summarize what changed.",
    notice: "已为浏览器检查预填提示，Astria 会在需要操作网页时说明动作。",
  },
  data: {
    title: "数据信号",
    status: "就绪",
    description: "从当前工作区的数据、日志或导出文件里找出关键结论。",
    prompt: "Analyze the local data or logs in this workspace and return the key signal.",
    notice: "数据分析会从当前工作区上下文开始。",
  },
  writing: {
    title: "写作润色",
    status: "就绪",
    description: "起草、润色或压缩文字交付物，适合 PRD、说明、汇报和发布稿。",
    prompt: "Draft a concise, polished write-up for this task.",
  },
  research: {
    title: "调研轨道",
    status: "就绪",
    description: "进行带证据链的调研，并输出可追踪来源和结论。",
    prompt: "Run deep research for this task and produce an evidence-backed brief.",
  },
  council: {
    title: "Agent 议会",
    status: "就绪",
    description: "多智能体规划和评审模式。启动一个议会运行，分别生成规划、调研和评审意见。",
    prompt: "Split this task across multiple named agents and propose a coordination plan.",
    panel: "council",
    notice: "Agent Council 会先生成可审核的角色贡献，不会自动执行代码改动。",
  },
  desktop: {
    title: "桌面控制",
    status: "受控",
    description: "需要操作本机 UI 时使用；Astria 会先说明动作并等待授权。",
    prompt: "Use desktop control only if needed and explain the intended action first.",
    notice: "桌面控制需要明确授权，Astria 会先说明动作。",
  },
  files: {
    title: "文件星舱",
    status: "就绪",
    description: "读取本地文档、检查归档内容，并把结果送入普通任务流。",
    prompt: "Inspect local files and recommend the safest next edit.",
    panel: "intake",
    notice: "已打开文件星舱。",
  },
  mcp: {
    title: "MCP 星港",
    status: "就绪",
    description: "查看配置的 MCP 服务器、连接测试和可用工具。",
    prompt: "Review MCP docking options and suggest the first server to connect.",
    panel: "mcp",
    notice: "已打开 MCP 星港。",
  },
  memory: {
    title: "记忆星图",
    status: "就绪",
    description: "查看记忆文件、会话来源，并审核写入 MEMORY.md 的候选内容。",
    prompt: "Create a memory map for this project: people, decisions, recurring tasks, and useful files.",
    panel: "memory",
    notice: "已打开记忆星图。",
  },
};

const workflowRecipes = {
  "code-review": {
    title: "代码评审",
    status: "审查",
    description: "检查当前改动的风险、回归点和测试缺口。",
    prompt: "Review the current working tree like a senior engineer. Lead with concrete findings, include file/line references where possible, and call out missing tests or risky behavior.",
    outcome: "一份按严重程度排序的评审报告，包含文件位置、行为风险和测试缺口。",
    context: ["当前 git diff", "相关测试输出", "高风险改动路径"],
    checklist: ["先列 findings", "标明残余风险", "建议最小验证命令"],
  },
  "feature-plan": {
    title: "功能规划",
    status: "规划",
    description: "把一个产品想法拆成 PRD、设计和可验证实施步骤。",
    prompt: "Turn this feature idea into a concise PRD, technical design, implementation plan, and validation checklist. Keep the scope shippable and aligned with the current codebase.",
    outcome: "一份可落地的 PRD、设计边界、实施顺序和验收清单。",
    context: ["目标用户", "现有代码路径", "非目标范围"],
    checklist: ["定义验收标准", "识别复用点", "拆成可提交切片"],
  },
  "file-intake": {
    title: "文件理解",
    status: "文件",
    description: "先进入 File Intake 读取文档或归档，再把结果送入任务。",
    displayPrompt: "先进入文件星舱读取相关本地文档或归档，再总结重要内容并提出下一步动作。",
    prompt: "Use File Intake to inspect the relevant local document or archive, then summarize the important content and propose the next action.",
    panel: "intake",
    outcome: "把本地文件内容整理成可引用上下文，再决定是否进入 Chat 或 run。",
    context: ["文件路径", "读取模式", "提取出的关键段落"],
    checklist: ["选择 intake mode", "审查结果摘要", "发送到下一步任务"],
  },
  "research-brief": {
    title: "调研简报",
    status: "调研",
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
    status: "收件",
    description: "审核外部渠道任务，决定拒绝、重试或转成运行。",
    prompt: "Triage the pending Inbox items. Identify which should become runs, which need more context, and which should be rejected.",
    panel: "inbox",
    outcome: "把外部任务分成可运行、需补充、应拒绝三类，并保留处理轨迹。",
    context: ["待处理 inbox 项", "来源渠道", "缺失上下文"],
    checklist: ["审查来源", "决定处理动作", "转成可追踪 run"],
  },
  "memory-update": {
    title: "记忆更新",
    status: "记忆",
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
    title: "快速运行",
    status: "快速",
    description: "最短路径进入本地执行，适合范围清楚、风险较低的任务。",
    prompt: "Execute this task directly in the current workspace. Keep the scope tight, report the changed files, and run the relevant validation.",
    panel: "runs",
    stageLabel: "快速本地执行",
    outcome: "Astria 直接推进任务，并在运行记录中保留结果。",
    checks: ["确认范围", "执行最小改动", "验证并汇报"],
  },
  research: {
    title: "调研简报",
    status: "深潜",
    description: "先做证据链、方案取舍和上下文归纳，再进入执行。",
    prompt: "Prepare a research brief before implementation. Separate facts, assumptions, options, tradeoffs, and recommended next steps.",
    panel: "runs",
    stageLabel: "先调研再执行",
    outcome: "先形成可审查的研究简报，减少盲目执行。",
    checks: ["列出证据", "标注假设", "给出建议路径"],
  },
  council: {
    title: "Agent 议会",
    status: "协作",
    description: "把复杂任务拆给规划、调研和评审角色，再合并成执行方案。",
    prompt: "Coordinate this task through multiple named agents. Ask planner, researcher, and reviewer roles for input, then synthesize a concrete plan.",
    panel: "council",
    stageLabel: "议会策略",
    outcome: "多智能体先分工评估，再收敛到一个可执行方案。",
    checks: ["拆分角色", "合并观点", "保留评审意见"],
  },
  guarded: {
    title: "人工确认",
    status: "闸门",
    description: "高风险命令、文件写入或外部动作先进入人工确认路径。",
    prompt: "Plan this task with explicit approval gates. Identify risky commands, file writes, network calls, and rollback points before acting.",
    panel: "permissions",
    stageLabel: "受控确认路径",
    outcome: "先标记风险动作和回滚点，再推进需要授权的步骤。",
    checks: ["识别风险", "设置审批点", "准备回滚"],
  },
  memory: {
    title: "记忆捕获",
    status: "回忆",
    description: "先从最近工作中提炼项目事实、偏好和风险，再继续任务。",
    prompt: "Draft a memory capture before continuing. Extract decisions, preferences, commands, risks, and project facts without writing durable memory until reviewed.",
    panel: "memory",
    stageLabel: "记忆捕获策略",
    outcome: "把上下文沉淀成可审核记忆，降低后续重复解释。",
    checks: ["提炼事实", "检查冲突", "审核后写入"],
  },
  tooling: {
    title: "MCP 工具接入",
    status: "工具",
    description: "先检查 MCP dock、外部工具和连接状态，再启动工具密集任务。",
    prompt: "Review the required tools for this task. Check MCP docks, missing environment keys, safety boundaries, and a minimal connection test plan.",
    panel: "mcp",
    stageLabel: "工具就绪检查",
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

async function copyText(text, successMessage = "已复制。") {
  await navigator.clipboard.writeText(text);
  showToast(successMessage);
}

function markButtonCopied(button) {
  const label = button.textContent;
  button.textContent = "已复制";
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

const panelDisplayNames = {
  agents: "智能体",
  browser: "浏览器规划器",
  budget: "预算守卫",
  chat: "对话",
  citation: "引用校准",
  compare: "路径比较台",
  council: "智能体议会",
  data: "数据规划器",
  delivery: "主动交付",
  diagnostics: "诊断",
  inbox: "收件箱",
  intake: "文件星舱",
  memory: "记忆星图",
  mcp: "MCP 星港",
  promptlab: "Prompt 实验室",
  results: "产物星库",
  reuse: "复用星库",
  runs: "运行",
  schedules: "定时任务",
  share: "交接包",
  snapshot: "工作区快照",
  sources: "来源登记",
};

function panelName(panel) {
  return panelDisplayNames[panel] || views[panel]?.[0] || panel || "-";
}

const uiTermLabels = {
  Agent: "Agent",
  Agents: "Agent",
  Anomaly: "异常",
  Browser: "浏览器",
  Budget: "预算",
  Channels: "渠道",
  "Chart brief": "图表简报",
  Citation: "引用",
  Command: "命令",
  Completion: "完成度",
  Context: "上下文",
  Council: "议会",
  Data: "数据",
  Delivery: "交付",
  Direct: "直接执行",
  Evidence: "证据",
  Fallback: "降级",
  Handoff: "交接",
  "Handoff pack": "交接包",
  Knowledge: "知识",
  Latest: "最近运行",
  Memory: "记忆",
  Monitor: "监控",
  Outcome: "结果",
  Prompt: "Prompt",
  Profile: "画像",
  Readiness: "就绪度",
  Research: "研究",
  Results: "产物",
  Retry: "重试",
  Reuse: "复用",
  "Run report": "运行报告",
  Runs: "运行",
  Schedules: "定时任务",
  Share: "交接",
  "Stop rules": "停止规则",
  "Token cap": "Token 上限",
  Trend: "趋势",
  Usage: "用量",
  ready: "就绪",
  review: "待复查",
  seed: "待建立",
  draft: "草稿",
  targeted: "已定向",
  guarded: "受保护",
  saveable: "可保存",
  schedulable: "可定时",
  manual: "手动",
  synthesized: "已综合",
  recovered: "已恢复",
  active: "活跃",
  completed: "完成",
  failed: "失败",
  error: "失败",
  running: "运行中",
  queued: "排队中",
  pending: "等待中",
  unknown: "未知",
  default: "默认",
  no_local_memory: "尚无本地记忆",
  degraded: "降级可用",
};

function uiTerm(value) {
  const text = String(value ?? "").trim();
  if (!text) return "";
  return uiTermLabels[text] || text;
}

function uiCount(count, singular, plural, zero) {
  const n = Number(count || 0);
  if (n === 0 && zero) return zero;
  return `${n} ${n === 1 ? singular : plural}`;
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
    type: "工作流",
    title: recipe.title,
    detail: recipe.description || recipe.outcome || "",
    run: () => selectWorkflowRecipe(id),
  }));
  const panelItems = [
    ["home", "任务台", "回到 Astria 本地任务入口。"],
    ["chat", "对话", "打开当前对话与执行表面。"],
    ["runs", "运行", "查看最近运行、状态筛选和事件轨道。"],
    ["intake", "文件星舱", "检查本地文档和归档。"],
    ["memory", "记忆星图", "复查长期上下文候选。"],
    ["mcp", "MCP 星港", "管理已配置的工具 dock。"],
    ["council", "Agent 议会", "协调规划、调研和评审角色。"],
    ["inbox", "收件箱", "分拣外部渠道任务。"],
    ["schedules", "定时任务", "管理周期性本地任务。"],
  ].map(([panel, title, detail]) => ({
    id: `panel:${panel}`,
    type: "面板",
    title,
    detail,
    run: () => switchPanel(panel),
  }));
  const actionItems = [
    ["research", "深度调研", "准备带证据链的简报。"],
    ["mcp", "规划 MCP 接入", "起草第一个工具 dock 连接。"],
    ["memory", "起草记忆星图", "准备待审记忆候选。"],
    ["council", "启动 Agent 议会", "把工作拆给命名角色。"],
  ].map(([action, title, detail]) => ({
    id: `action:${action}`,
    type: "动作",
    title,
    detail,
    run: () => runHomeAction(action),
  }));
  const recentSessionItems = state.sessions.slice(0, 3).map((session) => ({
    id: `session:${session.id}`,
    type: "最近",
    title: session.title || session.id,
    detail: `${session.msg_count ?? 0} 条消息 · 继续会话`,
    run: () => selectSession(session.id),
  }));
  const recentRunItems = state.runs.slice(0, 3).map((run) => ({
    id: `run:${run.id}`,
    type: "最近",
    title: run.prompt && run.id ? `${run.prompt} · ${run.id}` : run.prompt || run.id,
    detail: `${run.status || "unknown"} · ${run.agent || "默认"} · 打开运行`,
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
    renderEmpty(list, "没有匹配的快速指令。");
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
      return "就绪";
    case "warning":
      return "警告";
    case "needs_setup":
      return "需要设置";
    case "error":
      return "错误";
    default:
      return "未知";
  }
}

function liveStateLabel(value) {
  switch (value) {
    case "idle":
      return "空闲";
    case "running":
      return "运行中";
    case "complete":
    case "completed":
      return "已完成";
    case "cancelled":
      return "已取消";
    case "error":
      return "错误";
    default:
      return value || "空闲";
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
    label.textContent = "未选择会话";
    return;
  }
  const session = state.sessions.find((item) => item.id === state.activeSessionID);
  label.textContent = session ? `Session: ${session.title || session.id}` : `Session: ${state.activeSessionID}`;
}

function resetLiveRunStatus() {
  state.liveRun = {
    visible: false,
    state: "idle",
    runID: "",
    sessionID: "",
    usage: null,
    latest: "就绪",
  };
  renderLiveRunStatus();
  setStarMapActivity("calm", { label: "静默星图", hold: 0 });
}

function startLiveRunStatus(payload) {
  state.liveRun = {
    visible: true,
    state: "running",
    runID: payload?.request_id || "",
    sessionID: payload?.session_id || "",
    usage: null,
    latest: "正在连接事件流",
  };
  renderLiveRunStatus();
  setStarMapActivity("running", { label: "目标已锁定", active: "target" });
}

function updateLiveRunStatus(patch) {
  state.liveRun = { ...state.liveRun, ...patch, visible: true };
  renderLiveRunStatus();
}

function setStarMapActivity(kind, options = {}) {
  const map = document.querySelector("[data-star-map]");
  const status = document.querySelector("[data-star-map-status]");
  if (!map) return;
  if (state.starMap.timer) {
    clearTimeout(state.starMap.timer);
    state.starMap.timer = null;
  }
  const normalized = {
    idle: "calm",
    cancelled: "error",
    complete: "complete",
    completed: "complete",
    approval_resolved: "complete",
  }[kind] || kind || "calm";
  const active = options.active || starMapActiveForState(normalized);
  state.starMap.state = normalized;
  state.starMap.active = active;
  map.dataset.starState = normalized;
  map.dataset.starActive = active;
  if (status) status.textContent = options.label || starMapLabel(normalized);
  document.querySelectorAll("[data-star-step]").forEach((step) => {
    step.classList.toggle("active", step.dataset.starStep === active);
  });
  const hold = Number.isFinite(options.hold) ? options.hold : starMapHoldForState(normalized);
  if (normalized !== "calm" && hold > 0) {
    state.starMap.timer = setTimeout(() => {
      setStarMapActivity("calm", { label: "静默星图", hold: 0 });
    }, hold);
  }
}

const STAR_DECAY_POINTS = {
  target: { x: 50, y: 50, tone: "cyan" },
  context: { x: 16, y: 30, tone: "gold" },
  tool: { x: 84, y: 36, tone: "blue" },
  artifact: { x: 78, y: 76, tone: "green" },
  gate: { x: 18, y: 78, tone: "rose" },
};

function missionGraphSignature(graph) {
  const runID = graph.run?.id || state.activeRunID || state.liveRun.runID || "idle";
  const latestEvent = graph.latest?.event?.type || graph.latest?.tool?.title || graph.latest?.trace?.title || "";
  return [runID, graph.mapState, graph.active, graph.metrics.traceCount, graph.metrics.toolCount, latestEvent].join("|");
}

function shouldSpawnStarDecay(previous, graph) {
  if (!previous) return false;
  if (graph.mapState === "calm" && !graph.run) return false;
  if (previous.runID && previous.runID !== graph.run?.id) return true;
  if (["complete", "error"].includes(graph.mapState) && previous.mapState !== graph.mapState) return true;
  if (previous.active !== graph.active && ["tool", "artifact", "gate"].includes(previous.active)) return true;
  return previous.mapState === "tool" && graph.mapState !== "tool";
}

function renderStarDecay(previous, graph) {
  const field = document.querySelector("[data-star-decay-field]");
  if (!field || !shouldSpawnStarDecay(previous, graph)) return;
  const staleNode = previous.active || "context";
  const point = STAR_DECAY_POINTS[staleNode] || STAR_DECAY_POINTS.context;
  const label = previous.label || previous.mapState || "stale";
  const burst = document.createElement("span");
  burst.className = `star-decay-burst ${point.tone}`;
  burst.style.left = `${point.x}%`;
  burst.style.top = `${point.y}%`;
  burst.innerHTML = `<i></i><b></b><em>${escapeHTML(compactText(label, 12))}</em>`;
  field.appendChild(burst);
  while (field.children.length > 4) {
    field.firstElementChild?.remove();
  }
  window.setTimeout(() => burst.remove(), 3400);
}

function starMapActiveForState(kind) {
  switch (kind) {
    case "context":
    case "running":
      return "context";
    case "tool":
      return "tool";
    case "artifact":
    case "complete":
      return "artifact";
    case "approval":
      return "gate";
    case "error":
      return "target";
    default:
      return "target";
  }
}

function starMapLabel(kind) {
  switch (kind) {
    case "running":
      return "运行点亮";
    case "context":
      return "上下文唤醒";
    case "tool":
      return "工具链路";
    case "artifact":
      return "产物形成";
    case "approval":
      return "审核星门";
    case "complete":
      return "轨道完成";
    case "error":
      return "路径断裂";
    default:
      return "静默星图";
  }
}

function starMapHoldForState(kind) {
  if (kind === "complete" || kind === "error") return 3200;
  if (kind === "approval") return 0;
  return 1800;
}

function activeMissionRun() {
  if (state.currentRunDetail?.id) return state.currentRunDetail;
  if (state.missionRunDetail?.id) return state.missionRunDetail;
  if (state.activeRunID) {
    return state.runs.find((run) => run.id === state.activeRunID) || state.runs[0] || null;
  }
  return state.runs[0] || null;
}

function contextReadinessModel() {
  const mcpServers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled);
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const memoryReady = state.memoryStatus?.ready || memoryFacts.length > 0 || memoryEntries.length > 0;
  const sessionReady = Boolean(state.activeSessionID || state.sessions.length);
  const intakeReady = Boolean(state.intakeResult);
  const signals = [
    {
      id: "agents",
      label: "Agent",
      value: state.agents.length,
      ready: state.agents.length > 0,
      detail: state.agents.length ? `${state.agents.length} 个配置` : "使用默认 Agent",
      panel: "agents",
    },
    {
      id: "skills",
      label: "Skill",
      value: state.skills.length,
      ready: state.skills.length > 0,
      detail: state.skills.length ? `${state.skills.length} 个已加载` : "暂无技能",
      panel: "skills",
    },
    {
      id: "mcp",
      label: "MCP",
      value: enabledMCP.length,
      ready: enabledMCP.length > 0,
      detail: mcpServers.length ? `${enabledMCP.length}/${mcpServers.length} 启用` : "未配置 dock",
      panel: "mcp",
    },
    {
      id: "memory",
      label: "记忆",
      value: memoryFacts.length || memoryEntries.length || Number(state.memoryStatus?.local_facts || 0),
      ready: memoryReady,
      warning: memoryWarnings.length > 0 || state.memoryStatus?.provider === "degraded",
      detail: memoryWarnings.length
        ? `${memoryWarnings.length} 条警告`
        : state.memoryStatus?.reason
          ? uiTerm(state.memoryStatus.reason)
          : memoryReady
            ? `${memoryFacts.length || memoryEntries.length || state.memoryStatus?.local_facts || 0} 条可用`
            : "未注入",
      panel: "memory",
    },
    {
      id: "session",
      label: "会话",
      value: state.sessions.length,
      ready: sessionReady,
      detail: state.activeSessionID ? "当前会话" : state.sessions.length ? `${state.sessions.length} 条历史` : "新会话",
      panel: "chat",
    },
    {
      id: "intake",
      label: "输入",
      value: intakeReady ? 1 : 0,
      ready: intakeReady,
      detail: intakeReady ? "文件已解析" : "按需接入",
      panel: "intake",
    },
  ];
  const ready = signals.filter((item) => item.ready).length;
  const warnings = signals.filter((item) => item.warning).length;
  const count = signals.reduce((total, item) => total + Number(item.value || 0), 0);
  const percent = Math.max(18, Math.min(100, Math.round((ready / signals.length) * 100) - warnings * 7));
  const primary = signals.find((item) => item.warning) || [...signals].reverse().find((item) => item.ready) || signals[0];
  return {
    signals,
    ready,
    total: signals.length,
    warnings,
    count,
    percent,
    primary,
    label: warnings ? "需复查" : ready ? `${ready}/${signals.length} 就绪` : "待接入",
  };
}

function renderContextReadinessBoard() {
  const target = $("context-readiness-board");
  if (!target) return;
  const model = contextReadinessModel();
  const primary = model.primary || {};
  const signalCards = model.signals.map((signal) => {
    const tone = signal.warning ? "warning" : signal.ready ? "ready" : "idle";
    return `<button type="button" class="context-signal-card ${escapeHTML(tone)}" data-panel="${escapeHTML(signal.panel)}">
      <span>${escapeHTML(signal.label)}</span>
      <strong>${escapeHTML(String(signal.value ?? 0))}</strong>
      <small>${escapeHTML(signal.detail || "")}</small>
    </button>`;
  }).join("");
  target.innerHTML = `<section class="context-readiness-summary ${escapeHTML(model.warnings ? "warning" : model.ready ? "ready" : "idle")}">
    <div>
      <span>Context bundle</span>
      <strong>${escapeHTML(model.label)}</strong>
      <small>${escapeHTML(primary.detail || "选择上下文后再启动任务。")}</small>
    </div>
    <b>${escapeHTML(String(model.percent))}%</b>
  </section>
  <div class="context-signal-grid">${signalCards}</div>`;
}

function missionRunTelemetry(run) {
  const runID = run?.id || state.activeRunID || state.liveRun.runID || "";
  const related = state.missionEvents.filter((event) => !event.runID || !runID || event.runID === runID);
  return {
    usage: [...related].reverse().find((event) => event.type === "usage")?.data || null,
    budget: [...related].reverse().find((event) => event.type === "budget_status")?.data || null,
    runtime: [...related].reverse().find((event) => event.type === "run_status")?.data || null,
    tool: [...related].reverse().find((event) => event.type === "tool_status") || null,
    events: related,
  };
}

function missionTraceToolEvents(trace) {
  return trace
    .filter((item) => /tool|mcp/i.test(item.name || item.phase || ""))
    .map((item) => ({
      type: "trace_tool",
      at: item.timestamp,
      data: item.attributes || {},
      trace: item,
    }));
}

function missionToolName(event) {
  const data = event?.data || event?.trace?.attributes || {};
  return data.tool || data.name || data.server || data.mcp_server || event?.trace?.name || "";
}

function missionLatestApproval(steps, control) {
  const controlItem = [...control].reverse().find((item) => item.status === "approval_required" || item.action === "approval" || item.action === "replay");
  if (controlItem) {
    return {
      title: controlItem.action || "approval",
      detail: `${uiTerm(controlItem.status || "pending")} ${controlItem.reason || ""}`.trim(),
    };
  }
  const step = [...steps].reverse().find((item) => item.status === "waiting_approval" || item.metadata?.approval_id);
  if (step) {
    return {
      title: step.title || step.id || "approval",
      detail: `${uiTerm(step.status || "pending")} · ${formatTimestamp(step.updated_at)}`,
    };
  }
  const pending = state.approvals.values().next().value;
  if (pending) {
    return {
      title: pending.tool || pending.id || "approval",
      detail: pending.reason || "等待本地确认",
    };
  }
  return null;
}

function missionTraceSummary(trace) {
  const latest = [...trace].reverse().find((item) => item.name || item.phase);
  if (!latest) return null;
  return {
    title: latest.name || latest.event_id || "trace_event",
    detail: `${latest.phase || "event"} · ${formatTimestamp(latest.timestamp)}`,
  };
}

function missionGraphFromState() {
  const run = activeMissionRun();
  const runStatus = runHealthGroup(run);
  const events = Array.isArray(run?.events) ? run.events : [];
  const trace = Array.isArray(state.currentRunTrace) && run?.id === state.activeRunID
    ? state.currentRunTrace
    : Array.isArray(state.missionRunTrace) && run?.id === state.missionRunDetail?.id
      ? state.missionRunTrace
      : [];
  const steps = Array.isArray(run?.steps) ? run.steps : [];
  const control = Array.isArray(run?.control) ? run.control : [];
  const telemetry = missionRunTelemetry(run);
  const usage = run?.usage || run?.response?.usage || telemetry.usage || state.liveRun.usage || {};
  const context = contextReadinessModel();
  const toolEvents = [
    ...events.filter((event) => ["tool_call", "tool_result", "tool_status", "tool"].includes(event.type)),
    ...state.missionEvents.filter((event) => event.type === "tool_status"),
    ...missionTraceToolEvents(trace),
  ];
  const lastTool = missionToolName([...toolEvents].reverse().find((event) => missionToolName(event))) || "";
  const latestTool = [...toolEvents].reverse().find((event) => missionToolName(event));
  const approvalCount = state.approvals.size + control.filter((item) => item.status === "approval_required").length + steps.filter((step) => step.status === "waiting_approval").length;
  const artifactReady = runStatus === "completed" || Boolean(run?.response?.messages?.length) || Boolean(run?.response && !run?.response?.error);
  const artifactCount = resultArtifactRuns().length;
  const active = missionActiveNode(run, { toolEvents, approvalCount, artifactReady, runStatus });
  const mapState = missionMapState(run, { active, runStatus, approvalCount });
  return {
    run,
    status: run?.status || state.liveRun.state || "idle",
    runStatus,
    active,
    mapState,
    nodes: {
      target: {
        label: run ? "当前目标" : "目标",
        meta: run ? compactText(runPrompt(run) || run.id || "run", 24) : "等待任务",
      },
      context: {
        label: "上下文",
        meta: context.label,
      },
      tool: {
        label: "工具",
        meta: lastTool || context.signals.find((item) => item.id === "mcp")?.detail || `${state.skills.length} 技能`,
      },
      artifact: {
        label: "产物",
        meta: artifactReady ? `${Math.max(1, artifactCount)} 个可复查` : `${formatUsage(usage) || "等待结果"}`,
      },
      gate: {
        label: "审核",
        meta: approvalCount ? `${approvalCount} 待确认` : "clear",
      },
    },
    metrics: {
      contextCount: context.count,
      contextReady: context.ready,
      contextTotal: context.total,
      contextWarnings: context.warnings,
      contextPercent: context.percent,
      enabledMCP: context.signals.find((item) => item.id === "mcp")?.value || 0,
      memoryCount: context.signals.find((item) => item.id === "memory")?.value || 0,
      memoryWarnings: context.signals.find((item) => item.id === "memory")?.warning ? 1 : 0,
      traceCount: trace.length || run?.trace_events || 0,
      toolCount: toolEvents.length,
      approvalCount,
      artifactCount,
      usage,
      budget: telemetry.budget,
      runtime: telemetry.runtime,
    },
    context,
    latest: {
      trace: missionTraceSummary(trace),
      tool: latestTool ? {
        title: missionToolName(latestTool),
        detail: missionEventDetail(latestTool),
      } : null,
      approval: missionLatestApproval(steps, control),
      event: [...events, ...telemetry.events].reverse().find((event) => event.type) || null,
    },
    events: missionObservationEvents(run, { events: [...events, ...telemetry.events], trace, steps, control, toolEvents }),
  };
}

function missionActiveNode(run, detail) {
  if (detail.approvalCount) return "gate";
  if (!run) return state.starMap.active || "target";
  if (detail.runStatus === "failed") return "target";
  if (detail.artifactReady) return "artifact";
  if (detail.toolEvents.length) return "tool";
  if (run.session_id || detail.runStatus === "running") return "context";
  return "target";
}

function missionMapState(run, detail) {
  if (detail.approvalCount) return "approval";
  if (!run) return state.starMap.state || "calm";
  if (detail.runStatus === "failed") return "error";
  if (detail.runStatus === "completed") return "complete";
  if (detail.active === "tool") return "tool";
  if (detail.active === "artifact") return "artifact";
  if (detail.active === "context") return "context";
  if (detail.runStatus === "running") return "running";
  return "calm";
}

function missionObservationEvents(run, detail) {
  const rows = [];
  if (run) {
    rows.push({
      id: "OBS-1",
      kind: "RUN",
      tone: detail.runStatus === "failed" ? "error" : detail.runStatus === "completed" ? "result" : "run",
      time: formatTimestamp(run.started_at) || "now",
      title: `运行${uiTerm(run.status || "recorded")}`,
      detail: `${run.agent || "default"} · ${runPrompt(run) || run.id || "本地运行"}`,
    });
  } else {
    rows.push({ id: "OBS-1", kind: "READY", tone: "idle", time: "standby", title: "目标星待命", detail: "描述任务后，Astria 会创建一条本地 run 轨道。" });
  }
  const latestEvent = [...detail.events].reverse().find((event) => event.type);
  const latestTrace = [...detail.trace].reverse().find((event) => event.name);
  if (latestEvent) {
    rows.push({
      id: "OBS-2",
      kind: missionObservationKind(latestEvent.type),
      tone: missionObservationTone(latestEvent.type),
      time: formatTimestamp(latestEvent.at) || "live",
      title: runEventLabel(latestEvent.type),
      detail: missionEventDetail(latestEvent),
    });
  } else if (latestTrace) {
    rows.push({
      id: "OBS-2",
      kind: "TRACE",
      tone: "trace",
      time: formatTimestamp(latestTrace.timestamp) || "trace",
      title: latestTrace.name || "Trace",
      detail: latestTrace.phase || latestTrace.event_id || "结构化事件已记录。",
    });
  } else {
    rows.push({ id: "OBS-2", kind: "READ", tone: "idle", time: "queued", title: "上下文星待命", detail: "Agent、记忆、MCP、文件和会话会在运行中接入。" });
  }
  const latestStep = [...detail.steps].reverse().find((step) => step.status);
  const latestControl = [...detail.control].reverse().find((item) => item.action || item.status);
  if (latestStep) {
    rows.push({
      id: "OBS-3",
      kind: "STEP",
      tone: latestStep.status === "waiting_approval" ? "approval" : "run",
      time: formatTimestamp(latestStep.updated_at) || "step",
      title: latestStep.title || latestStep.id || "工作流阶段",
      detail: `${uiTerm(latestStep.status)} · ${formatTimestamp(latestStep.updated_at)}`,
    });
  } else if (latestControl) {
    rows.push({
      id: "OBS-3",
      kind: latestControl.status === "approval_required" ? "APPROVAL" : "CONTROL",
      tone: latestControl.status === "approval_required" ? "approval" : "tool",
      time: formatTimestamp(latestControl.updated_at || latestControl.at) || "gate",
      title: `控制：${latestControl.action || "decision"}`,
      detail: `${uiTerm(latestControl.status)} ${latestControl.reason || ""}`.trim(),
    });
  } else {
    rows.push({ id: "OBS-3", kind: "TOOL", tone: "idle", time: "waiting", title: "活动叠层待命", detail: "工具调用、审批、用量和产物会在这里形成观测序列。" });
  }
  return rows.slice(0, 3);
}

function missionObservationKind(type) {
  if (["tool_call", "tool_result", "tool_status", "tool"].includes(type)) return "TOOL";
  if (["approval_needed", "approval_resolved"].includes(type)) return "APPROVAL";
  if (type === "usage" || type === "budget_status") return "BUDGET";
  if (type === "run_status") return "RUN";
  if (type === "memory.saved" || type === "context" || type === "source") return "READ";
  return "EVENT";
}

function missionObservationTone(type) {
  if (["approval_needed", "budget_status"].includes(type)) return "approval";
  if (["tool_call", "tool_result", "tool_status", "tool"].includes(type)) return "tool";
  if (["usage", "run_status"].includes(type)) return "run";
  if (["approval_resolved", "memory.saved"].includes(type)) return "result";
  return "trace";
}

function missionEventDetail(event) {
  const data = event?.data || {};
  if (data.tool) {
    return `${data.tool} · ${uiTerm(data.status || (data.is_error ? "error" : "observed"))}`;
  }
  if (event?.type === "budget_status") return data.status || data.detail || formatToolPayload(safeRenderPayload(data));
  if (event?.type === "run_status") return data.status || data.detail || data.reason || "运行状态已更新";
  if (data.status) return uiTerm(data.status);
  if (data.input_tokens || data.output_tokens) return formatLiveUsage(data);
  if (data.preview) return String(data.preview);
  return formatTimestamp(event?.at) || "事件已记录。";
}

function compactText(value, max = 32) {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  if (text.length <= max) return text;
  return `${text.slice(0, Math.max(0, max - 1))}…`;
}

function renderMissionGraph() {
  const map = document.querySelector("[data-star-map]");
  if (!map) return;
  const graph = missionGraphFromState();
  const previous = state.starMap.signature
    ? {
        signature: state.starMap.signature,
        runID: state.starMap.runID,
        mapState: state.starMap.state,
        active: state.starMap.active,
        label: state.starMap.label,
      }
    : null;
  const nextSignature = missionGraphSignature(graph);
  if (nextSignature !== state.starMap.signature) {
    renderStarDecay(previous, graph);
  }
  map.dataset.starState = graph.mapState;
  map.dataset.starActive = graph.active;
  state.starMap.state = graph.mapState;
  state.starMap.active = graph.active;
  state.starMap.runID = graph.run?.id || state.activeRunID || state.liveRun.runID || "";
  state.starMap.label = missionGraphStatusLabel(graph);
  state.starMap.signature = nextSignature;
  const status = document.querySelector("[data-star-map-status]");
  if (status) status.textContent = missionGraphStatusLabel(graph);
  for (const [key, node] of Object.entries(graph.nodes)) {
    const target = document.querySelector(`[data-star-node="${key}"]`);
    if (!target) continue;
    const label = target.querySelector("span");
    const meta = target.querySelector("small");
    if (label) label.textContent = node.label;
    if (meta) meta.textContent = node.meta;
  }
  document.querySelectorAll("[data-star-step]").forEach((step) => {
    step.classList.toggle("active", step.dataset.starStep === graph.active);
  });
  renderMissionMapOverlay(graph);
  renderObservationLog(graph);
  renderMissionInspector(graph);
}

function setMissionOverlayText(selector, value) {
  const target = document.querySelector(selector);
  if (target) target.textContent = value;
}

function missionOrbitPercent(graph) {
  if (graph.metrics.approvalCount) return 72;
  if (graph.runStatus === "completed") return 100;
  if (graph.runStatus === "failed") return 34;
  if (graph.active === "artifact") return 84;
  if (graph.active === "tool") return 62;
  if (graph.active === "context") return Math.max(38, Number(graph.metrics.contextPercent || 0));
  if (graph.run) return 22;
  return Math.min(18, Number(graph.metrics.contextPercent || 0));
}

function renderMissionMapOverlay(graph) {
  const title = graph.run ? runPrompt(graph.run) || graph.run.id || "当前本地运行" : "等待新的本地任务";
  const detail = graph.run
    ? `${uiTerm(graph.status)} · ${graph.nodes.context.meta} · ${graph.nodes.tool.meta}`
    : "星图负责关系判断；观测日志负责过程证据；Inspector 负责当前选中星体。";
  const percent = Math.max(0, Math.min(100, Math.round(missionOrbitPercent(graph))));
  setMissionOverlayText("[data-mission-object-title]", compactText(title, 34));
  setMissionOverlayText("[data-mission-object-detail]", compactText(detail, 72));
  setMissionOverlayText("[data-mission-object-target]", compactText(graph.nodes.target.meta, 16));
  setMissionOverlayText("[data-mission-object-context]", compactText(graph.nodes.context.meta, 16));
  setMissionOverlayText("[data-mission-object-tool]", compactText(graph.nodes.tool.meta, 16));
  setMissionOverlayText("[data-mission-object-artifact]", compactText(graph.nodes.artifact.meta, 16));
  setMissionOverlayText("[data-mission-orbit-percent]", `${percent}%`);
  setMissionOverlayText("[data-mission-orbit-label]", missionGraphStatusLabel(graph));
  const meter = document.querySelector("[data-mission-orbit-meter]");
  if (meter) meter.style.width = `${percent}%`;
  renderMissionMotionOverlay(graph, percent);
}

function missionGraphStatusLabel(graph) {
  if (graph.metrics.approvalCount) return `${graph.metrics.approvalCount} 个审核门`;
  if (graph.run) return `${uiTerm(graph.status)} · ${graph.metrics.traceCount} trace`;
  if (graph.metrics.contextReady) return `上下文 ${graph.metrics.contextReady}/${graph.metrics.contextTotal}`;
  return "静默星图";
}

function missionMotionModel(graph, orbitPercent) {
  const activeHealth = Math.max(8, Math.min(96, orbitPercent || 0));
  const reviewHealth = graph.metrics.approvalCount ? 58 : graph.runStatus === "failed" ? 34 : 12;
  const decayHealth = graph.run ? Math.max(12, 42 - Math.min(30, graph.metrics.traceCount * 3)) : 22;
  const hasToolMotion = graph.active === "tool" || graph.metrics.toolCount > 0;
  const hasArtifactMotion = graph.active === "artifact" || graph.runStatus === "completed" || graph.metrics.artifactCount > 0;
  const hasRetryMotion = graph.runStatus === "failed" || graph.metrics.approvalCount > 0;
  const hasMemoryMotion = graph.metrics.memoryCount > 0 || graph.events.some((event) => event.kind === "READ");
  const focusTrail = ["target"];
  if (graph.metrics.contextReady || graph.active === "context" || hasToolMotion || hasArtifactMotion) focusTrail.push("context");
  if (hasToolMotion || hasArtifactMotion) focusTrail.push("tool");
  if (hasArtifactMotion) focusTrail.push("artifact");
  if (hasMemoryMotion) focusTrail.push("memory");
  return {
    activeHealth,
    reviewHealth,
    decayHealth,
    focusTrail: focusTrail.join(" / "),
    hasToolMotion,
    hasArtifactMotion,
    hasRetryMotion,
    hasMemoryMotion,
    hasSpawnMotion: graph.runStatus === "running" || graph.metrics.traceCount > 1,
    mode: graph.metrics.approvalCount ? "review" : graph.runStatus === "failed" ? "retry" : hasArtifactMotion ? "artifact" : hasToolMotion ? "tool" : graph.active,
  };
}

function setMotionToggle(selector, enabled) {
  const target = document.querySelector(selector);
  if (target) target.classList.toggle("active", Boolean(enabled));
}

function setHealthValue(barSelector, labelSelector, value) {
  const safeValue = Math.max(0, Math.min(100, Math.round(Number(value) || 0)));
  const bar = document.querySelector(barSelector);
  const label = document.querySelector(labelSelector);
  if (bar) bar.style.setProperty("--health", `${safeValue}%`);
  if (label) label.textContent = String(safeValue);
}

function renderMissionMotionOverlay(graph, orbitPercent) {
  const map = document.querySelector("[data-star-map]");
  if (!map) return;
  const motion = missionMotionModel(graph, orbitPercent);
  map.dataset.motionMode = motion.mode;
  map.dataset.motionMemory = motion.hasMemoryMotion ? "active" : "idle";
  map.dataset.motionSpawn = motion.hasSpawnMotion ? "active" : "idle";
  const focusTrail = document.querySelector("[data-focus-trail]");
  if (focusTrail) focusTrail.textContent = motion.focusTrail;
  setMotionToggle('[data-causal-chip="context"]', motion.hasToolMotion || graph.active === "context");
  setMotionToggle('[data-causal-chip="edit"]', motion.hasArtifactMotion);
  setMotionToggle('[data-causal-chip="retry"]', motion.hasRetryMotion);
  setHealthValue("[data-health-active]", "[data-health-active-label]", motion.activeHealth);
  setHealthValue("[data-health-review]", "[data-health-review-label]", motion.reviewHealth);
  setHealthValue("[data-health-decay]", "[data-health-decay-label]", motion.decayHealth);
}

function renderObservationLog(graph = missionGraphFromState()) {
  const target = $("home-observation-log");
  if (!target) return;
  target.innerHTML = graph.events.map((item, index) => `<div class="observation-row ${escapeHTML(item.tone || "idle")} ${index === graph.events.length - 1 ? "latest" : ""}">
    <span class="observation-time">${escapeHTML(item.time || item.id)}</span>
    <span class="observation-kind">${escapeHTML(item.kind || "EVENT")}</span>
    <div class="observation-copy">
      <strong>${escapeHTML(compactText(item.title, 28))}</strong>
      <small>${escapeHTML(compactText(item.detail || "", 96))}</small>
    </div>
  </div>`).join("");
}

function renderMissionInspector(graph = missionGraphFromState()) {
  const target = $("home-mission-inspector");
  if (!target) return;
  const run = graph.run;
  const title = run ? compactText(runPrompt(run) || run.id || "当前运行", 46) : "等待任务星体";
  const tag = run?.id || "local";
  const runAction = run?.id ? `<button type="button" data-run-open="${escapeHTML(run.id)}">观测运行</button>` : `<button type="button" data-panel="chat">打开对话</button>`;
  const contextSignals = graph.context.signals.map((item) => `<button type="button" class="${item.ready ? "ready" : ""} ${item.warning ? "warning" : ""}" data-panel="${escapeHTML(item.panel)}">
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(item.detail)}</strong>
    </button>`).join("");
  const latestResult = resultArtifactRuns()[0];
  const artifactPreview = latestResult ? resultArtifactSummary(latestResult) : "";
  const telemetryLine = [
    graph.metrics.budget?.status ? `预算：${uiTerm(graph.metrics.budget.status)}` : "",
    graph.metrics.runtime?.status ? `运行：${uiTerm(graph.metrics.runtime.status)}` : "",
    formatUsage(graph.metrics.usage) !== "-" ? `用量：${formatUsage(graph.metrics.usage)}` : "",
  ].filter(Boolean).join(" · ");
  const latestRows = missionInspectorLatestRows(graph);
  const latestHTML = latestRows.map((item) => `<button type="button" class="${escapeHTML(item.tone || "")}" ${item.action}>
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(compactText(item.title, 42))}</strong>
      <small>${escapeHTML(compactText(item.detail || "", 78))}</small>
    </button>`).join("");
  const quality = missionInspectorQualityModel(graph);
  const motion = missionMotionModel(graph, Math.round(missionOrbitPercent(graph)));
  const budgetStatus = graph.metrics.budget?.status ? uiTerm(graph.metrics.budget.status) : graph.run ? "跟随运行" : "待规划";
  const runtimeStatus = graph.metrics.runtime?.status ? uiTerm(graph.metrics.runtime.status) : uiTerm(graph.status);
  const usageLabel = formatUsage(graph.metrics.usage) !== "-" ? formatUsage(graph.metrics.usage) : "未捕获";
  const resultAction = artifactPreview ? `<button type="button" data-panel="results">打开产物</button>` : "";
  const contextAction = graph.metrics.approvalCount
    ? `<button type="button" data-panel="permissions">审核门</button>`
    : `<button type="button" data-panel="memory">上下文</button>`;
  target.innerHTML = `<div class="mission-inspector-head">
      <div>
        <span class="board-kicker">Inspector</span>
        <strong>${escapeHTML(title)}</strong>
      </div>
      <span>${escapeHTML(compactText(tag, 18))}</span>
    </div>
    <div class="mission-inspector-body">
      <div class="mission-inspector-card ${escapeHTML(graph.runStatus)}">
        <strong>${escapeHTML(run ? `状态：${uiTerm(graph.status)}` : "暂无活动运行")}</strong>
        <small>${escapeHTML(run ? `${run.agent || "default"} · ${run.channel || "local"} · ${formatTimestamp(run.started_at)}` : "发起任务后会显示真实 run、trace、上下文和审核状态。")}</small>
      </div>
      <section class="mission-inspector-section">
        <div class="mission-inspector-section-head">
          <span>Context bundle</span>
          <strong>${escapeHTML(`${graph.metrics.contextReady}/${graph.metrics.contextTotal}`)}</strong>
        </div>
        <div class="mission-inspector-meter" aria-label="上下文完整度"><i style="width:${escapeHTML(String(contextReadinessPercent(graph)))}%"></i></div>
        <div class="mission-context-signals" aria-label="上下文就绪信号">
          ${contextSignals}
        </div>
      </section>
      <section class="mission-inspector-section compact">
        <div class="mission-inspector-section-head">
          <span>Budget / quality</span>
          <strong>${escapeHTML(String(quality.score))}</strong>
        </div>
        <div class="mission-inspector-status-grid">
          <button type="button" data-panel="budget"><span>预算</span><strong>${escapeHTML(budgetStatus)}</strong></button>
          <button type="button" data-panel="quality"><span>质量</span><strong>${escapeHTML(quality.label)}</strong></button>
          <button type="button" data-panel="runs"><span>运行</span><strong>${escapeHTML(runtimeStatus)}</strong></button>
          <button type="button" data-panel="runs"><span>用量</span><strong>${escapeHTML(usageLabel)}</strong></button>
        </div>
        <small class="mission-inspector-caption">${escapeHTML(quality.detail)}</small>
      </section>
      <div class="mission-inspector-grid">
        <span>Trace</span><b>${escapeHTML(String(graph.metrics.traceCount))}</b>
        <span>工具</span><b>${escapeHTML(String(graph.metrics.toolCount))}</b>
        <span>产物</span><b>${escapeHTML(String(graph.metrics.artifactCount))}</b>
        <span>审核</span><b>${escapeHTML(String(graph.metrics.approvalCount))}</b>
      </div>
      <section class="mission-inspector-section compact">
        <div class="mission-inspector-section-head">
          <span>Live signals</span>
          <strong>${escapeHTML(compactText(graph.active, 12))}</strong>
        </div>
        <div class="mission-inspector-feed" aria-label="最新运行信号">
          ${latestHTML}
        </div>
      </section>
      <section class="mission-inspector-section compact">
        <div class="mission-inspector-section-head">
          <span>Motion grammar</span>
          <strong>${escapeHTML(motion.mode)}</strong>
        </div>
        <div class="mission-motion-grid">
          <span class="${motion.hasToolMotion ? "active" : ""}">tool</span>
          <span class="${motion.hasArtifactMotion ? "active" : ""}">edit</span>
          <span class="${motion.hasSpawnMotion ? "active" : ""}">spawn</span>
          <span class="${motion.hasMemoryMotion ? "active" : ""}">memory</span>
          <span class="${motion.decayHealth < 34 ? "active" : ""}">fade</span>
        </div>
        <small class="mission-inspector-caption">${escapeHTML(`路径：${motion.focusTrail} · active ${Math.round(motion.activeHealth)} / review ${Math.round(motion.reviewHealth)} / remote ${Math.round(motion.decayHealth)}`)}</small>
      </section>
      ${telemetryLine ? `<div class="mission-inspector-note">${escapeHTML(telemetryLine)}</div>` : ""}
      ${artifactPreview ? `<div class="mission-inspector-note result">${escapeHTML(compactText(artifactPreview, 110))}</div>` : ""}
      <div class="mission-inspector-actions">
        ${runAction}
        ${resultAction}
        ${contextAction}
      </div>
    </div>`;
}

function missionInspectorQualityModel(graph) {
  let score = 42;
  if (graph.run) score += 10;
  if (graph.metrics.contextReady) score += Math.min(18, graph.metrics.contextReady * 4);
  if (graph.metrics.traceCount) score += Math.min(12, graph.metrics.traceCount * 2);
  if (graph.metrics.toolCount) score += Math.min(10, graph.metrics.toolCount * 2);
  if (graph.metrics.artifactCount) score += 8;
  if (graph.metrics.approvalCount) score -= 6;
  if (graph.runStatus === "failed") score -= 18;
  if (graph.runStatus === "completed") score += 8;
  score = Math.max(12, Math.min(96, score));
  const label = score >= 78 ? "可交付" : score >= 58 ? "可复查" : graph.metrics.approvalCount ? "待审核" : "待运行";
  const detail = graph.run
    ? `${graph.metrics.traceCount} trace，${graph.metrics.toolCount} 个工具信号，${graph.metrics.contextReady}/${graph.metrics.contextTotal} 上下文就绪。`
    : "质量评分会随上下文、trace、工具事件、产物和审核状态更新。";
  return { score, label, detail };
}

function missionInspectorLatestRows(graph) {
  const rows = [];
  const runID = graph.run?.id || "";
  if (graph.latest.trace) {
    rows.push({
      label: "Trace",
      title: graph.latest.trace.title,
      detail: graph.latest.trace.detail,
      action: runID ? `data-run-open="${escapeHTML(runID)}"` : `data-panel="runs"`,
      tone: "trace",
    });
  } else {
    rows.push({
      label: "Trace",
      title: "尚未记录",
      detail: "运行开始后显示结构化 trace phase。",
      action: `data-panel="runs"`,
      tone: "",
    });
  }
  if (graph.latest.tool) {
    rows.push({
      label: "工具",
      title: graph.latest.tool.title,
      detail: graph.latest.tool.detail,
      action: `data-panel="mcp"`,
      tone: "tool",
    });
  } else {
    rows.push({
      label: "工具",
      title: graph.nodes.tool.meta,
      detail: "工具事件会从 run events、trace 和 daemon SSE 汇入。",
      action: `data-panel="mcp"`,
      tone: "",
    });
  }
  if (graph.latest.approval) {
    rows.push({
      label: "审核",
      title: graph.latest.approval.title,
      detail: graph.latest.approval.detail,
      action: `data-panel="permissions"`,
      tone: "warning",
    });
  } else if (graph.latest.event) {
    rows.push({
      label: "事件",
      title: runEventLabel(graph.latest.event.type),
      detail: missionEventDetail(graph.latest.event),
      action: runID ? `data-run-open="${escapeHTML(runID)}"` : `data-panel="runs"`,
      tone: "event",
    });
  } else {
    rows.push({
      label: "事件",
      title: "等待活动",
      detail: "审批、预算、用量和运行状态会在这里出现。",
      action: `data-panel="runs"`,
      tone: "",
    });
  }
  return rows;
}

function contextReadinessPercent(graph) {
  const score = Number(graph.metrics.contextPercent || 0);
  return Math.max(graph.run ? 36 : 18, Math.min(100, score));
}

function formatLiveUsage(usage) {
  if (!usage) return "-";
  const input = usage.input_tokens ?? usage.InputTokens;
  const output = usage.output_tokens ?? usage.OutputTokens;
  const total = usage.total_tokens ?? (Number(input || 0) + Number(output || 0));
  const parts = [];
  if (Number.isFinite(input)) parts.push(`in ${input}`);
  if (Number.isFinite(output)) parts.push(`out ${output}`);
  if (Number.isFinite(total) && parts.length) parts.push(`total ${total}`);
  return parts.length ? parts.join(" · ") : "-";
}

function renderLiveRunStatus() {
  const box = $("live-run-status");
  if (!box) return;
  box.hidden = !state.liveRun.visible;
  box.dataset.state = state.liveRun.state || "idle";
  setText("live-run-state", liveStateLabel(state.liveRun.state));
  setText("live-run-id", state.liveRun.runID || "-");
  setText("live-session-id", state.liveRun.sessionID || "-");
  setText("live-run-usage", formatLiveUsage(state.liveRun.usage));
  setText("live-run-event", state.liveRun.latest || "-");
}

function setRunControls(isRunning) {
  if (isRunning) $("chat-state").textContent = "运行中";
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
  const title = needsSetup ? "Astria 需要设置。" : "Astria 已就绪。";
  const subtitle = needsSetup
    ? (diagnostics.summary || "打开诊断或连接器完成设置。")
    : "从输入框发起本地任务，或选择一个 Agent 继续。";
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
    $("chat-state").textContent = "已选择会话";
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
          <span>${escapeHTML(error.message)}。下一条消息仍会尝试继续此会话。</span>
        </div>
      </div>`;
    }
  }
}

window.selectSession = selectSession;

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
        <strong>${escapeHTML(session.title || session.id || "已选择会话")}</strong>
        <span>这个会话还没有保存的消息。</span>
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
  $("chat-state").textContent = "就绪";
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

function displayPromptFor(item) {
  return item?.displayPrompt || item?.prompt || "";
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
    seedMissionPrompt(displayPromptFor(action));
  }
  if (action.notice) showToast(action.notice);
}

function selectWorkflowRecipe(id) {
  const recipe = workflowRecipes[id];
  if (!recipe) return;
  state.homeMode = `recipe:${id}`;
  state.workflowStage = "draft";
  state.workflowStageLabel = recipe.title || "Workflow draft";
  $("home-task-input").value = displayPromptFor(recipe);
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
  $("home-task-input").value = displayPromptFor(strategy);
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
    title: "通用任务",
    status: "就绪",
    description: "直接描述目标，Astria 会从当前工作区和默认智能体开始。",
  };
  setText("home-mode-kicker", mode.status || "就绪");
  setText("home-mode-title", mode.title || "通用任务");
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
  if ((state.homeMode.startsWith("recipe:") || state.homeMode.startsWith("strategy:")) && state.workflowStage === "draft") {
    return {
      stage: "draft",
      label: state.workflowStageLabel || "通用任务",
    };
  }
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const latestRun = state.runs[0];
  if (memoryFacts.length || memoryWarnings.length) {
    return {
      stage: "memory",
      label: memoryWarnings.length ? `${memoryWarnings.length} 条记忆警告` : `${memoryFacts.length} 条记忆事实`,
    };
  }
  if (latestRun) {
    const group = runHealthGroup(latestRun);
    if (group === "running") {
      return { stage: "running", label: latestRun.prompt || latestRun.id || "运行中任务" };
    }
    return { stage: "review", label: latestRun.prompt || latestRun.id || "复查最近运行" };
  }
  return {
    stage: state.workflowStage || "draft",
    label: state.workflowStageLabel || "通用任务",
  };
}

function renderWorkflowStageRail() {
  const rail = $("workflow-stage-rail");
  if (!rail) return;
  const current = currentWorkflowStage();
  const order = ["draft", "running", "review", "memory"];
  const activeIndex = order.indexOf(current.stage);
  const stages = [
    ["draft", "草稿", "工作流和 Prompt"],
    ["running", "运行", "Daemon 执行"],
    ["review", "复查", "运行观测"],
    ["memory", "记忆", "长期上下文"],
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
  const title = recipe?.title || stage.label || strategy.title || "通用任务";
  const context = latestRun
    ? `${latestRun.status || "unknown"} 运行 · ${latestRun.agent || "default"}`
    : latestSession
      ? `${latestSession.msg_count ?? 0} 条消息会话`
      : `${strategy.title || "策略"} · 暂无最近工作`;
  const next = stage.stage === "memory"
    ? "复查记忆候选"
    : stage.stage === "review"
      ? "打开运行观测"
      : stage.stage === "running"
        ? "监控进行中的运行"
        : "起草或发起下一次任务";
  const runAction = latestRun?.id ? `<button type="button" data-run-open="${escapeHTML(latestRun.id)}">观测运行</button>` : "";
  const sessionAction = latestSession?.id ? `<button type="button" data-session-resume="${escapeHTML(latestSession.id)}">恢复会话</button>` : "";
  target.innerHTML = `<div class="focus-brief-head">
      <div>
        <span class="board-kicker">${escapeHTML(stage.stage)}</span>
        <strong>${escapeHTML(title)}</strong>
      </div>
      <small>${escapeHTML(next)}</small>
    </div>
    <div class="focus-brief-grid">
      <span>上下文</span><strong>${escapeHTML(context)}</strong>
      <span>策略</span><strong>${escapeHTML(strategy.title || "Quick Run")}</strong>
      <span>会话</span><strong>${escapeHTML(latestSession?.title || latestSession?.id || "暂无会话")}</strong>
      <span>运行</span><strong>${escapeHTML(latestRun?.prompt || latestRun?.id || "暂无运行")}</strong>
    </div>
    <div class="focus-brief-actions">
      ${runAction}
      ${sessionAction}
      <button type="button" data-panel="${stage.stage === "memory" ? "memory" : "runs"}">${escapeHTML(stage.stage === "memory" ? "打开记忆" : "打开运行观测")}</button>
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
      label: "诊断",
      value: statusLabel(diagnosticsStatus),
      detail: state.diagnostics?.summary || "运行就绪",
    },
    {
      panel: "permissions",
      tone: permissionsConfigured ? "ready" : "warning",
      label: "权限",
      value: permissionsConfigured ? "已配置" : "默认",
      detail: permissionsConfigured ? "显式工具策略" : "内置护栏",
    },
    {
      panel: "mcp",
      tone: enabledMCP ? "ready" : mcpServers.length ? "warning" : "",
      label: "MCP",
      value: enabledMCP ? `${enabledMCP} 个启用` : mcpServers.length ? "已停用" : "无 dock",
      detail: "工具连接",
    },
    {
      panel: "memory",
      tone: memoryWarnings.length ? "attention" : memoryFacts.length ? "ready" : "",
      label: "记忆",
      value: memoryWarnings.length ? `${memoryWarnings.length} 条警告` : memoryFacts.length ? `${memoryFacts.length} 条事实` : "预览",
      detail: memoryWarnings.length ? "复查分类" : "长期上下文",
    },
  ];
  strip.innerHTML = items.map((item) => `<button type="button" class="workspace-health-item ${escapeHTML(item.tone)}" data-panel="${escapeHTML(item.panel)}">
    <span>${escapeHTML(item.label)}</span>
    <strong>${escapeHTML(item.value)}</strong>
    <small>${escapeHTML(item.detail)}</small>
  </button>`).join("");
}

function systemStatusModel() {
  const diagnosticsStatus = String(state.diagnostics?.status || "unknown").toLowerCase();
  const runtimeReady = ["ready", "ok", "healthy"].includes(diagnosticsStatus);
  const cfg = state.config || {};
  const provider = cfg.provider || "未配置";
  const providerReady = configReadiness(cfg);
  const permissions = state.permissions || {};
  const permissionRuleCount = [
    permissions.allowed_dirs,
    permissions.allowed_commands,
    permissions.denied_commands,
    permissions.network_allowlist,
    permissions.sensitive_patterns,
  ].reduce((total, values) => total + (Array.isArray(values) ? values.length : 0), 0);
  const mcpServers = Array.isArray(cfg.mcp_servers) ? cfg.mcp_servers : [];
  const enabledMCP = mcpServers.filter((server) => !server.disabled).length;
  const version = state.version || {};
  const cards = [
    {
      panel: "diagnostics",
      tone: runtimeReady ? "ready" : diagnosticsStatus === "warning" ? "warning" : diagnosticsStatus === "unknown" ? "idle" : "attention",
      label: "运行时",
      value: statusLabel(state.diagnostics?.status || "unknown"),
      detail: localizedSystemMessage(state.diagnostics?.summary) || "等待诊断结果",
      ready: runtimeReady,
    },
    {
      panel: "config",
      tone: providerReady.ready ? "ready" : cfg.provider ? "warning" : "attention",
      label: "Provider",
      value: provider,
      detail: providerReady.ready ? "连接器就绪" : providerReady.missing.length ? `缺少 ${providerReady.missing.join("、")}` : "尚未加载配置",
      ready: providerReady.ready,
    },
    {
      panel: "permissions",
      tone: permissions.configured === true ? "ready" : state.permissions ? "warning" : "idle",
      label: "权限",
      value: permissions.configured === true ? "已配置" : "默认",
      detail: permissions.configured === true ? (permissionRuleCount ? `${permissionRuleCount} 条显式规则` : "显式策略已加载") : "使用内置护栏",
      ready: Boolean(state.permissions),
    },
    {
      panel: "version",
      tone: version.version ? "ready" : "idle",
      label: "版本",
      value: version.version || "待加载",
      detail: version.update_supported ? "支持更新检查" : (localizedSystemMessage(version.message) || "本地构建状态"),
      ready: Boolean(version.version),
    },
    {
      panel: "mcp",
      tone: enabledMCP ? "ready" : mcpServers.length ? "warning" : "idle",
      label: "MCP",
      value: enabledMCP ? `${enabledMCP}/${mcpServers.length}` : mcpServers.length ? "已停用" : "无 dock",
      detail: enabledMCP ? "工具 dock 可用" : mcpServers.length ? "检查停用项" : "按需添加工具连接",
      ready: Boolean(enabledMCP),
    },
  ];
  const readyCount = cards.filter((card) => card.ready).length;
  const attentionCount = cards.filter((card) => card.tone === "attention" || card.tone === "warning").length;
  const score = Math.max(12, Math.min(100, Math.round((readyCount / cards.length) * 100) - attentionCount * 5));
  return {
    cards,
    readyCount,
    attentionCount,
    score,
    label: attentionCount ? "需要复查" : readyCount ? `${readyCount}/${cards.length} 就绪` : "待加载",
    primary: cards.find((card) => card.tone === "attention" || card.tone === "warning") || cards.find((card) => card.ready) || cards[0],
  };
}

function localizedSystemMessage(value) {
  const text = String(value || "").trim();
  const labels = {
    "StarClaw is ready to run agents.": "StarClaw 已准备好运行 Agent。",
    "Development build - update checks require a release version.": "开发构建，更新检查需要发布版本。",
  };
  return labels[text] || text;
}

function renderSystemStatusBoard() {
  const target = $("system-status-board");
  if (!target) return;
  const model = systemStatusModel();
  const primary = model.primary || {};
  target.innerHTML = `<section class="system-status-summary ${escapeHTML(model.attentionCount ? "warning" : model.readyCount ? "ready" : "idle")}">
    <div>
      <span>系统控制</span>
      <strong>${escapeHTML(model.label)}</strong>
      <small>${escapeHTML(primary.detail || "等待系统状态。")}</small>
    </div>
    <b>${escapeHTML(String(model.score))}%</b>
  </section>
  <div class="system-signal-grid">${model.cards.map((card) => `<button type="button" class="system-signal-card ${escapeHTML(card.tone || "idle")}" data-panel="${escapeHTML(card.panel)}">
    <span>${escapeHTML(card.label)}</span>
    <strong>${escapeHTML(card.value)}</strong>
    <small>${escapeHTML(card.detail)}</small>
  </button>`).join("")}</div>`;
}

function reviewQueueItems() {
  const items = [];
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const activeRuns = state.runs.filter((run) => runHealthGroup(run) === "running");
  if (failedRuns.length) {
    const run = failedRuns[0];
    items.push({
      tone: "attention",
      label: "运行",
      title: `${failedRuns.length} 条运行需要复查`,
      detail: run.prompt || run.id || "打开运行面板检查失败状态。",
      runID: run.id || "",
    });
  } else if (activeRuns.length) {
    const run = activeRuns[0];
    items.push({
      tone: "active",
      label: "运行",
      title: `${activeRuns.length} 条任务轨道活跃`,
      detail: run.prompt || run.id || "观察正在执行的 daemon 运行。",
      runID: run.id || "",
    });
  }

  const inboxCounts = inboxStatusCounts();
  if (inboxCounts.failed) {
    items.push({
      tone: "attention",
      label: "收件箱",
      title: `${inboxCounts.failed} 个进入项失败`,
      detail: "在阻塞队列前重试或拒绝失败的渠道任务。",
      panel: "inbox",
    });
  } else if (inboxCounts.pending) {
    items.push({
      tone: "warning",
      label: "收件箱",
      title: `${inboxCounts.pending} 个进入项待审`,
      detail: "外部任务转成 Astria 运行前需要先复查。",
      panel: "inbox",
    });
  }

  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  if (memoryWarnings.length) {
    items.push({
      tone: "attention",
      label: "记忆",
      title: `${memoryWarnings.length} 条记忆警告`,
      detail: String(memoryWarnings[0] || "添加长期上下文前先复查分类警告。"),
      panel: "memory",
    });
  }

  const diagnosticsStatus = state.diagnostics?.status || "unknown";
  if (!["ready", "unknown"].includes(diagnosticsStatus)) {
    items.push({
      tone: diagnosticsStatus === "warning" ? "warning" : "attention",
      label: "诊断",
      title: `诊断${statusLabel(diagnosticsStatus)}`,
      detail: state.diagnostics?.summary || "检查启动就绪状态。",
      panel: "diagnostics",
    });
  }

  if (state.permissions && state.permissions.configured !== true) {
    items.push({
      tone: "warning",
      label: "权限",
      title: "权限仍使用默认护栏",
      detail: "为这个工作区设置显式工具边界。",
      panel: "permissions",
    });
  } else if (state.permissions) {
    const hints = permissionsRiskHints(state.permissions);
    if (hints.length) {
      items.push({
        tone: "warning",
        label: "权限",
        title: `${hints.length} 条策略提示`,
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
      title: "MCP dock 已停用",
      detail: "工具密集任务前先启用或测试 dock。",
      panel: "mcp",
    });
  } else if (!mcpServers.length && state.config) {
    items.push({
      tone: "",
      label: "MCP",
      title: "尚未配置 MCP dock",
      detail: "当工作区需要外部工具时再添加 dock。",
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
      <span>清空</span>
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
      label: "确认",
      title: `${state.approvals.size} 个确认请求待处理`,
      detail: "继续前请在对话中复查工具或命令确认卡。",
      panel: "chat",
    });
  }
  if (state.permissions && state.permissions.configured !== true) {
    items.push({
      tone: "warning",
      label: "权限",
      title: "默认护栏",
      detail: "高风险工作流前先建立显式权限。",
      panel: "permissions",
    });
  } else if (state.permissions) {
    const hints = permissionsRiskHints(state.permissions);
    if (hints.length) {
      items.push({
        tone: "warning",
        label: "权限",
        title: `${hints.length} 条策略待审`,
        detail: hints[0],
        panel: "permissions",
      });
    }
  }
  if (!["ready", "unknown"].includes(diagnosticsStatus)) {
    items.push({
      tone: diagnosticsStatus === "warning" ? "warning" : "attention",
      label: "诊断",
      title: `运行时${statusLabel(diagnosticsStatus)}`,
      detail: state.diagnostics?.summary || "风险动作前先解决启动就绪问题。",
      panel: "diagnostics",
    });
  }
  if (failedRuns.length) {
    const run = failedRuns[0];
    items.push({
      tone: "attention",
      label: "恢复",
      title: `${failedRuns.length} 条运行需要恢复`,
      detail: run.prompt || run.id || "打开运行面板复查失败状态。",
      runID: run.id || "",
    });
  }
  if (inboxCounts.failed || inboxCounts.pending) {
    items.push({
      tone: inboxCounts.failed ? "attention" : "warning",
      label: "收件箱",
      title: inboxCounts.failed ? `${inboxCounts.failed} 个进入项失败` : `${inboxCounts.pending} 个进入项待审`,
      detail: "外部渠道任务运行前需要确认、重试或拒绝。",
      panel: "inbox",
    });
  }
  if (mcpServers.length && !enabledMCP) {
    items.push({
      tone: "warning",
      label: "MCP",
      title: "工具 dock 已停用",
      detail: "工具密集运行前先启用或测试 dock。",
      panel: "mcp",
    });
  } else if (!mcpServers.length && state.config) {
    items.push({
      tone: "",
      label: "MCP",
      title: "尚无工具 dock",
      detail: "需要外部工具时，再添加受审批边界保护的 dock。",
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
      <span>无阻塞</span>
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

function rememberMissionEvent(type, event) {
  const data = parseEventData(event.data);
  state.missionEvents = [
    ...state.missionEvents,
    {
      id: event.lastEventId || `${Date.now()}-${state.missionEvents.length}`,
      type,
      at: new Date().toISOString(),
      runID: data.run_id || data.id || data.request_id || "",
      data,
    },
  ].slice(-60);
  return data;
}

function handleMissionTelemetryEvent(type, event) {
  const data = rememberMissionEvent(type, event);
  if (type === "tool_status") {
    setStarMapActivity("tool", { label: data.tool || "工具事件", active: "tool" });
  } else if (type === "usage" || type === "budget_status") {
    setStarMapActivity("running", { label: type === "usage" ? "用量更新" : "预算状态", active: "context" });
  } else if (type === "run_status") {
    setStarMapActivity(data.status === "waiting_approval" ? "approval" : "running", { label: uiTerm(data.status || "运行状态") });
  }
  renderMissionGraph();
}

function connectEventStream() {
  if (!("EventSource" in window) || state.eventSource) return;
  state.eventStream.status = "connecting";
  const source = new EventSource("/events");
  state.eventSource = source;
  const trackEventStreamID = (event) => {
    if (event.lastEventId) {
      state.eventStream.lastEventID = event.lastEventId;
    }
  };
  source.addEventListener("approval_needed", (event) => {
    trackEventStreamID(event);
    const data = rememberMissionEvent("approval_needed", event);
    renderApprovalCard(data);
    renderMissionGraph();
  });
  source.addEventListener("approval_resolved", (event) => {
    trackEventStreamID(event);
    const data = rememberMissionEvent("approval_resolved", event);
    markApprovalResolved(data);
    renderMissionGraph();
  });
  source.addEventListener("run_started", (event) => {
    trackEventStreamID(event);
    handleRunLifecycleEvent("run_started", rememberMissionEvent("run_started", event));
  });
  source.addEventListener("run_completed", (event) => {
    trackEventStreamID(event);
    handleRunLifecycleEvent("run_completed", rememberMissionEvent("run_completed", event));
  });
  source.addEventListener("run_error", (event) => {
    trackEventStreamID(event);
    handleRunLifecycleEvent("run_error", rememberMissionEvent("run_error", event));
  });
  for (const type of ["tool_status", "usage", "budget_status", "run_status"]) {
    source.addEventListener(type, (event) => {
      trackEventStreamID(event);
      handleMissionTelemetryEvent(type, event);
    });
  }
  for (const type of ["cloud_delegate_start", "cloud_delegate_progress", "cloud_delegate_complete", "preamble", "stream_delta"]) {
    source.addEventListener(type, (event) => {
      trackEventStreamID(event);
      rememberMissionEvent(type, event);
      renderMissionGraph();
    });
  }
  source.onerror = () => {
    state.eventStream.status = "reconnecting";
    state.eventStream.reconnecting = true;
    $("daemon-pill").textContent = "Reconnecting";
    $("daemon-pill").className = "bad";
  };
  source.onopen = () => {
    if (state.eventStream.reconnecting) {
      state.eventStream.reconnects += 1;
      state.eventStream.status = "recovered";
      state.eventStream.reconnecting = false;
      state.eventStream.lastRecoveredAt = new Date().toISOString();
      refreshRunsAfterEventStreamRecovery();
    } else {
      state.eventStream.status = "running";
    }
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

function primaryNavPanel(panel) {
  if (["chat", "agents", "schedules", "inbox"].includes(panel)) return "home";
  if (["quality", "compare", "promptlab", "budget", "council"].includes(panel)) return "runs";
  if (["reuse", "playbooks", "starter", "share", "snapshot", "delivery"].includes(panel)) return "results";
  if (["skills", "browser", "data", "mcp", "intake", "memory", "sources", "reconcile", "citation"].includes(panel)) return "memory";
  if (["manage", "settings", "diagnostics", "config", "permissions", "version"].includes(panel)) return "settings";
  return panel;
}

function switchPanel(panel) {
  if (!views[panel]) return;
  hideToast();
  state.panel = panel;
  const activeNavPanel = primaryNavPanel(panel);
  document.querySelectorAll(".nav-item").forEach((button) => {
    button.classList.toggle("active", button.dataset.panel === activeNavPanel);
  });
  document.querySelectorAll(".panel").forEach((section) => {
    section.classList.toggle("active", section.id === `panel-${panel}`);
  });
  $("view-title").textContent = views[panel][0];
}

window.switchPanel = switchPanel;

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
  setText("home-brief-run-count", state.runs.length);
  setText("home-brief-pending-count", pending);
  renderMissionGraph();
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
  renderRunQualityScorecard();
  renderSourceRegistry();
  renderKnowledgeReconciliation();
  renderCitationGroundingPlanner();
  renderPromptExperimentLab();
  renderBudgetGuardPlanner();
  renderReuseGallery();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderStarterKitLauncher();
  renderSharePackBuilder();
  renderWorkspaceSnapshotPlanner();
  renderBrowserMissionPlanner();
  renderDataInsightPlanner();
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
    setText("home-brief-current-title", "等待任务");
    setText("home-brief-current-detail", "发起任务后显示最近运行、状态和 Agent。");
    const brief = $("home-brief-current");
    if (brief) {
      delete brief.dataset.runOpen;
      brief.dataset.panel = "runs";
      brief.className = "home-brief-current";
    }
    return;
  }
  const status = runHealthGroup(latest);
  delete target.dataset.panel;
  target.dataset.runOpen = latest.id || "";
  target.className = `board-run ${status}`;
  target.innerHTML = `<strong>${escapeHTML(latest.prompt || latest.id || "Untitled run")}</strong>
    <small>${escapeHTML(latest.status || "unknown")} · ${escapeHTML(latest.agent || "default")} · ${escapeHTML(formatTimestamp(latest.started_at))}</small>`;
  const brief = $("home-brief-current");
  if (brief) {
    delete brief.dataset.panel;
    brief.dataset.runOpen = latest.id || "";
    brief.className = `home-brief-current ${status}`;
  }
  setText("home-brief-current-title", latest.prompt || latest.id || "Untitled run");
  setText("home-brief-current-detail", `${uiTerm(latest.status || "unknown")} · ${latest.agent || "default"} · ${formatTimestamp(latest.started_at)}`);
}

function renderHomeDockedTools() {
  setText("home-skill-count", state.skills.length);
  setText("home-agent-count", state.agents.length);
  setText("home-brief-agent-count", state.agents.length);
  setText("home-schedule-count", state.schedules.length);
  const mcpCount = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers.length : 0;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  setText("home-mcp-count", mcpCount);
  setText("home-brief-mcp-count", mcpCount);
  setText("home-memory-count", memoryCount);
  setText("home-council-count", state.councilRuns.length);
  setText("home-inbox-count", (inboxStatusCounts().pending || 0));
  setText("home-intake-count", state.intakeResult ? "就绪" : "本地");
  renderMissionGraph();
  renderWorkspaceHub();
  renderKnowledgeCuration();
  renderToolDockInspector();
  renderPromptSuggestionDock();
  renderFocusBrief();
  renderWorkspaceHealthStrip();
  renderApprovalCenter();
  renderReviewQueue();
  renderComparisonWorkbench();
  renderRunQualityScorecard();
  renderSourceRegistry();
  renderKnowledgeReconciliation();
  renderCitationGroundingPlanner();
  renderPromptExperimentLab();
  renderBudgetGuardPlanner();
  renderReuseGallery();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderStarterKitLauncher();
  renderSharePackBuilder();
  renderWorkspaceSnapshotPlanner();
  renderBrowserMissionPlanner();
  renderDataInsightPlanner();
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
    ? `${memoryWarnings.length} 条警告`
    : memoryFacts.length
      ? `${memoryFacts.length} 条事实`
      : `${memoryEntries.length} 个来源`;
  const runHealth = failed
    ? `${failed} 条需处理`
    : running
      ? `${running} 条活跃`
      : completed
        ? `${completed} 条完成`
        : "还没有运行";
  const intakeLabel = state.intakeResult
    ? (state.intakeResult.mode || "结果就绪")
    : "本地路径";
  hub.innerHTML = [
    {
      panel: "chat",
      sessionID: latestSession?.id || "",
      kicker: "会话",
      title: latestSession?.title || latestSession?.id || "等待会话",
      detail: latestSession ? `${latestSession.msg_count ?? 0} 条消息 · 打开对话` : "开始一次对话后，这里会显示最近会话。",
      tone: latestSession ? "ready" : "",
    },
    {
      panel: "runs",
      runID: state.runs[0]?.id || "",
      kicker: "运行",
      title: runHealth,
      detail: `共 ${state.runs.length} 条 · 打开运行观测`,
      tone: failed ? "attention" : running ? "active" : completed ? "ready" : "",
    },
    {
      panel: "memory",
      kicker: "记忆",
      title: memoryLabel,
      detail: memoryWarnings.length ? "添加更多记忆前先复查分类警告。" : "打开记忆星图管理长期上下文。",
      tone: memoryWarnings.length ? "attention" : memoryFacts.length || memoryEntries.length ? "ready" : "",
    },
    {
      panel: "intake",
      kicker: "文件",
      title: intakeLabel,
      detail: state.intakeResult ? "复查最近提取的本地上下文。" : "打开文件星舱检查文档或归档。",
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
  if (!memoryWarnings.length && !memoryFacts.length && !memoryEntries.length) {
    items.push({
      tone: "clear",
      label: "Low context",
      title: "暂无知识候选",
      detail: "Open Memory Map to review sources before promoting context.",
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
      label: "会话",
      title: latestSession.title || latestSession.id || "最近会话",
      reason: "从最新对话上下文续接。",
      prompt: `Resume the recent session and identify the next useful task.\n\nSession: ${latestSession.id || "unknown"}\nMessages: ${latestSession.msg_count ?? 0}`,
    });
  }
  if (pendingInbox.length) {
    items.push({
      tone: "active",
      label: "收件箱",
      title: `${pendingInbox.length} 条进入项`,
      reason: "把外部请求转成已审查工作。",
      prompt: "Triage pending inbox items. Group them into run now, needs context, and reject, then propose the next reviewed action.",
    });
  }
  if (diagnosticsStatus && !["ok", "ready", "healthy"].includes(String(diagnosticsStatus).toLowerCase())) {
    items.push({
      tone: "attention",
      label: "就绪度",
      title: "诊断需要复查",
      reason: state.diagnostics?.summary || "运行时就绪状态仍不清晰。",
      prompt: "Review daemon diagnostics and list the smallest setup or configuration fixes needed before the next mission.",
    });
  }
  if (!enabledMCP && state.config) {
    items.push({
      tone: "",
      label: "工具",
      title: "规划第一条 MCP dock",
      reason: "重工具工作流需要已配置的 dock。",
      prompt: "Suggest the first MCP dock for this workspace. Include the command or URL, required env keys, safety considerations, and a connection test plan.",
    });
  }
  if (state.intakeResult) {
    items.push({
      tone: "ready",
      label: "文件",
      title: state.intakeResult.path || "Intake result",
      reason: "把已提取的本地上下文用于下一次任务。",
      prompt: `Summarize this file intake result and turn it into a concrete next action.\n\nPath: ${state.intakeResult.path || ""}\nMode: ${state.intakeResult.mode || ""}\n\n${String(state.intakeResult.content || "").slice(0, 1200)}`,
    });
  }
  if (recipe) {
    items.push({
      tone: "active",
      label: "工作流",
      title: recipe.title || "Selected workflow",
      reason: recipe.outcome || recipe.description || "继续当前选中的工作流。",
      prompt: displayPromptFor(recipe) || "继续当前 Astria 工作流，并定义下一步检查。",
    });
  } else {
    items.push({
      tone: "clear",
      label: "默认",
      title: `${strategy.title || "快速运行"}下一步`,
      reason: strategy.outcome || strategy.description || "从当前策略继续。",
      prompt: displayPromptFor(strategy) || "继续当前 Astria 任务，复查工作区状态，识别下一步有价值动作，并说明需要的验证。",
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
      source: "运行",
      panel: "runs",
      title: latestRun?.prompt || "最近运行证据",
      metric: state.runs.length ? `${completedRuns.length}/${state.runs.length} 完成` : "待建立",
      evidence: [
        latestRun ? `最近：${uiTerm(latestRun.status || "unknown")} · ${latestRun.agent || "默认 Agent"}` : "尚未捕获最近运行",
        failedRuns.length ? `${failedRuns.length} 条失败运行需要复查` : "当前列表没有失败运行",
        latestRun ? `开始于 ${formatTimestamp(latestRun.started_at)}` : "完成第一次执行后打开运行观测台",
      ],
      tradeoff: "适合让下一步沿着真实执行证据推进，而不是重新规划。",
      recommendation: failedRuns.length ? "先复查失败原因，再发起相似运行。" : state.runs.length ? "以最近成功运行作为最短续接路径。" : "先创建一条基准运行，再依据执行证据决策。",
      prompt: `Compare recent Astria runs and decide the next execution path.\n\nLatest run: ${latestRun?.prompt || "none"}\nStatus: ${latestRun?.status || "unknown"}\nCompleted runs: ${completedRuns.length}\nFailed runs: ${failedRuns.length}`,
    },
    {
      id: "agent-profiles",
      source: "Agent",
      panel: "agents",
      title: latestAgent?.name || "Agent 配置候选",
      metric: state.agents.length ? `${state.agents.length} 个配置` : "待建立",
      evidence: [
        latestAgent ? `候选主 Agent：${latestAgent.name}` : "尚未选择主 Agent",
        `${commandCount} 条已保存命令`,
        latestAgent?.model ? `模型：${latestAgent.model}` : "模型沿用默认配置",
      ],
      tradeoff: "适合角色、模型、工具权限或保存命令会影响成败的任务。",
      recommendation: commandCount ? "从最接近的已保存命令所属 Agent 开始。" : "先选择一个聚焦 Agent，再增加工作流状态。",
      prompt: `Compare Astria agent profiles for the next task.\n\nProfiles: ${state.agents.map((agent) => agent.name).join(", ") || "none"}\nSaved commands: ${commandCount}`,
    },
    {
      id: "memory-context",
      source: "记忆",
      panel: "memory",
      title: "长期上下文",
      metric: memoryEntries.length ? `${memoryEntries.length} 条记忆` : "待建立",
      evidence: [
        `${memoryCategories.size || 0} 个记忆分类`,
        memoryEntries[0]?.text ? `最近：${String(memoryEntries[0].text).slice(0, 90)}` : "尚未选择长期记忆",
        "适合承载偏好、决策、风险和架构约束",
      ],
      tradeoff: "适合正确性依赖既有项目决策，而不是只看最近上下文的任务。",
      recommendation: memoryEntries.length ? "使用记忆上下文，避免重复已经确认的决策。" : "先沉淀一个决策或偏好，再依赖记忆推进。",
      prompt: `Compare current options against Astria memory.\n\nMemory categories: ${Array.from(memoryCategories).join(", ") || "uncategorized"}\nMemory count: ${memoryEntries.length}`,
    },
    {
      id: "council-synthesis",
      source: "议会",
      panel: "council",
      title: latestCouncil?.goal || "议会综合",
      metric: roles.length ? `${roles.length} 个角色` : "待建立",
      evidence: [
        latestCouncil ? `目标：${latestCouncil.goal}` : "尚未选择议会结果",
        roles.length ? `角色：${roles.map((role) => role.role).join(", ")}` : "尚未捕获角色记录",
        latestCouncil?.synthesis ? "综合结果可进入交接" : "综合结果待生成",
      ],
      tradeoff: "适合下一步需要平衡规划、研究和审查视角的任务。",
      recommendation: latestCouncil?.synthesis ? "当取舍比速度更重要时，使用议会综合。" : "先运行议会，再把结果视为已审查。",
      prompt: `Compare the council synthesis against other Astria options.\n\nGoal: ${latestCouncil?.goal || "none"}\nRoles: ${roles.map((role) => role.role).join(", ") || "none"}\nSynthesis:\n${latestCouncil?.synthesis || ""}`,
    },
  ];
}

function renderComparisonWorkbench() {
  const lanes = comparisonCandidates();
  setText("nav-compare-count", lanes.length);
  setText("manage-compare-count", `${lanes.length} 条路径`);
  setText("compare-summary", `${lanes.length} 条比较路径，来自当前工作区的运行、Agent、记忆和评议证据。`);
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
      <button type="button" data-compare-select="${escapeHTML(lane.id)}">决策简报</button>
      <button type="button" data-compare-draft="${escapeHTML(lane.id)}">起草比较</button>
      <button type="button" data-panel="${escapeHTML(lane.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderComparisonDetail(lanes.find((lane) => lane.id === state.selectedComparisonLane) || lanes[0]);
}

function renderComparisonDetail(lane) {
  const target = $("comparison-detail");
  if (!target) return;
  if (!lane) {
    target.innerHTML = `<div class="empty-state">选择一条比较路径。</div>`;
    return;
  }
  const title = String(lane.title || "");
  const displayTitle = title.length > 140 ? `${title.slice(0, 137)}...` : title;
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(displayTitle)}</h3>
      <div class="run-meta-grid">
        <span>来源</span><strong>${escapeHTML(lane.source)}</strong>
        <span>就绪度</span><strong>${escapeHTML(lane.metric)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(lane.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>证据</h3>
      <div class="comparison-evidence detail">
        ${lane.evidence.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>取舍</h3>
      <p>${escapeHTML(lane.tradeoff)}</p>
      <h3>建议</h3>
      <p>${escapeHTML(lane.recommendation)}</p>
      <div class="run-detail-actions">
        <button type="button" data-compare-draft="${escapeHTML(lane.id)}">起草比较</button>
        <button type="button" data-panel="${escapeHTML(lane.panel)}">打开来源</button>
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
  showToast("比较任务已写入对话。");
}

function runQualityScore(run, emphasis = "overall") {
  if (!run) return 32;
  const health = runHealthGroup(run);
  const hasResult = Boolean(formatRunResponse(run.response).trim());
  const hasPrompt = Boolean(runPrompt(run));
  const hasUsage = Boolean(run.usage || run.response?.usage);
  let score = 48;
  if (health === "completed") score += 24;
  if (health === "running") score += 8;
  if (health === "failed") score -= 16;
  if (hasResult) score += 12;
  if (hasPrompt) score += 8;
  if (hasUsage) score += 4;
  if (run.error) score -= 12;
  if (emphasis === "evidence" && hasResult) score += 6;
  if (emphasis === "budget" && hasUsage) score += 8;
  if (emphasis === "reuse" && health === "completed" && hasResult) score += 8;
  return Math.max(5, Math.min(98, score));
}

function runQualityGrade(score) {
  if (score >= 85) return "A";
  if (score >= 72) return "B";
  if (score >= 58) return "C";
  if (score >= 42) return "D";
  return "复查";
}

function runQualityCards() {
  const latestRun = state.runs[0];
  const completedRuns = state.runs.filter((run) => runHealthGroup(run) === "completed");
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed");
  const latestCompleted = completedRuns[0] || latestRun;
  const latestFailed = failedRuns[0] || latestRun;
  const sourceCount = sourceRegistryRows().length;
  const resultCount = resultArchiveEntries().length;
  const budgetCount = budgetGuardCards().length;
  const citationCount = citationGroundingCards().length;
  const shareCount = sharePackCards().length;
  const deliveryCount = deliveryLanes().length;
  const latestScore = runQualityScore(latestRun);
  const completedScore = runQualityScore(latestCompleted, "reuse");
  const failedScore = failedRuns.length ? Math.max(18, 52 - failedRuns.length * 7) : 76;
  const evidenceScore = Math.min(96, 42 + sourceCount * 6 + citationCount * 5 + (latestRun ? 12 : 0));
  const budgetScore = Math.min(94, 44 + budgetCount * 5 + (latestRun?.usage || latestRun?.response?.usage ? 12 : 0));
  const reuseScore = Math.min(96, 40 + resultCount * 7 + shareCount * 5 + (completedRuns.length ? 12 : 0));
  const deliveryScore = Math.min(94, 46 + deliveryCount * 6 + (state.schedules.length ? 8 : 0));
  return [
    {
      id: "latest-run",
      type: "最近运行",
      title: latestRun?.prompt || "最近运行评分",
      panel: "runs",
      runID: latestRun?.id || "",
      score: latestScore,
      signal: latestRun ? `${uiTerm(latestRun.status || "unknown")} · ${latestRun.agent || "默认 Agent"}` : "尚未捕获运行",
      risk: latestRun ? (runHealthGroup(latestRun) === "failed" ? "最近运行失败；应先检查错误，避免盲目重试。" : "最近运行仍需证据与复用审查，才能沉淀为长期产物。") : "尚无执行证据。",
      gate: "继续前复查 Prompt、结果、用量、时间线和后续动作。",
      route: "打开运行观测台检查执行细节。",
      prompt: `Evaluate the latest Astria run quality.\n\nRun: ${latestRun?.prompt || "none"}\nStatus: ${latestRun?.status || "unknown"}\nAgent: ${latestRun?.agent || "default"}\nScore estimate: ${latestScore}\n\nReturn completion quality, evidence strength, budget posture, risk, and the next action.`,
    },
    {
      id: "completed-output",
      type: "完成度",
      title: "已完成产物就绪度",
      panel: "results",
      runID: latestCompleted?.id || "",
      score: completedScore,
      signal: `${completedRuns.length} 条完成运行；${resultCount} 个产物条目`,
      risk: completedRuns.length ? "完成不等于已引用、可复用或已接受。" : "尚无可复用的完成运行。",
      gate: "结果需要证据、鲜度、验收检查和可复用的下一条路线。",
      route: "打开产物星库归档或续接。",
      prompt: `Evaluate completed Astria output readiness.\n\nCompleted runs: ${completedRuns.length}\nResult archive entries: ${resultCount}\nLatest completed: ${latestCompleted?.prompt || "none"}\nScore estimate: ${completedScore}\n\nDecide whether the output is ready to archive, reuse, share, or needs more validation.`,
    },
    {
      id: "failure-retry",
      type: "重试",
      title: "失败与重试风险",
      panel: failedRuns.length ? "budget" : "runs",
      runID: latestFailed?.id || "",
      score: failedScore,
      signal: failedRuns.length ? `${failedRuns.length} 条失败运行` : "当前列表没有失败运行",
      risk: failedRuns.length ? "若计划不变，重复失败会消耗上下文和预算。" : "重试风险较低，但仍需要停止规则。",
      gate: "重试前需要根因、变更后的 Prompt/工具路线、降级阶梯和停止条件。",
      route: failedRuns.length ? "打开预算守卫制定降级和停止规则。" : "打开运行观测台检查基线历史。",
      prompt: `Evaluate Astria failure and retry risk.\n\nFailed runs: ${failedRuns.length}\nLatest failed: ${latestFailed?.prompt || "none"}\nScore estimate: ${failedScore}\n\nReturn likely failure class, retry risk, changed plan required before retry, fallback route, and stop rule.`,
    },
    {
      id: "evidence-quality",
      type: "证据",
      title: "证据质量评分",
      panel: "citation",
      runID: latestRun?.id || "",
      score: evidenceScore,
      signal: `${sourceCount} 个来源；${citationCount} 条引用检查`,
      risk: "运行看似成功时，未支撑的结论可能仍被隐藏。",
      gate: "结论需要来源覆盖、引用鲜度、未支撑清单和安全表述。",
      route: "打开引用校准检查来源覆盖和证据缺口。",
      prompt: `Evaluate Astria evidence quality for recent work.\n\nSources: ${sourceCount}\nCitation checks: ${citationCount}\nLatest run: ${latestRun?.prompt || "none"}\nScore estimate: ${evidenceScore}\n\nReturn claim coverage, weak evidence, missing citations, freshness risks, and safe wording recommendations.`,
    },
    {
      id: "budget-posture",
      type: "预算",
      title: "预算与停止规则姿态",
      panel: "budget",
      runID: latestRun?.id || "",
      score: budgetScore,
      signal: `${budgetCount} 条预算守卫；用量${latestRun?.usage || latestRun?.response?.usage ? "已捕获" : "未捕获"}`,
      risk: "长任务重跑前需要上限、上下文裁剪、降级路线和明确停止规则。",
      gate: "预算形态、模型路线、降级阶梯和停止条件必须清楚。",
      route: "打开预算守卫规划更便宜或更安全的路线。",
      prompt: `Evaluate Astria budget posture for recent work.\n\nBudget guards: ${budgetCount}\nLatest run usage captured: ${Boolean(latestRun?.usage || latestRun?.response?.usage)}\nScore estimate: ${budgetScore}\n\nReturn token/time risk, context trimming plan, model route, fallback ladder, and stop conditions.`,
    },
    {
      id: "reuse-readiness",
      type: "复用",
      title: "可复用产物就绪度",
      panel: "share",
      runID: latestCompleted?.id || "",
      score: reuseScore,
      signal: `${resultCount} 个产物；${shareCount} 个交接包`,
      risk: "可复用资产需要边界，否则未来会话会继承过期或私密假设。",
      gate: "可复用产物需要摘要、证据、边界、验收检查和下一步。",
      route: "打开交接包打包已审查内容。",
      prompt: `Evaluate Astria reusable output readiness.\n\nResults: ${resultCount}\nShare packs: ${shareCount}\nCompleted runs: ${completedRuns.length}\nScore estimate: ${reuseScore}\n\nReturn what can be reused, what needs redaction, what evidence is missing, and the next starter prompt.`,
    },
    {
      id: "delivery-readiness",
      type: "交付",
      title: "交付就绪度评分",
      panel: "delivery",
      runID: latestRun?.id || "",
      score: deliveryScore,
      signal: `${deliveryCount} 条交付链路；${state.schedules.length} 条定时任务`,
      risk: "交付需要审批边界、目标、产物、验证和回滚路线。",
      gate: "没有明确审批，不外发、不定时、不改变远端状态。",
      route: "打开主动交付复查出站就绪度。",
      prompt: `Evaluate Astria delivery readiness.\n\nDelivery lanes: ${deliveryCount}\nSchedules: ${state.schedules.length}\nLatest run: ${latestRun?.prompt || "none"}\nScore estimate: ${deliveryScore}\n\nReturn destination readiness, approval gate, artifact quality, verification, rollback, and whether delivery should stay local.`,
    },
  ];
}

function renderRunQualityScorecard() {
  const cards = runQualityCards();
  setText("nav-quality-count", cards.length);
  setText("manage-quality-count", `${cards.length} 张卡片`);
  setText("quality-summary", `${cards.length} 张运行质量卡，覆盖最近运行、完成度、重试、证据、预算、复用和交付就绪度。`);
  const list = $("run-quality-grid");
  if (!list) return;
  if (!state.selectedRunQuality || !cards.some((card) => card.id === state.selectedRunQuality)) {
    state.selectedRunQuality = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="run-quality-card ${card.id === state.selectedRunQuality ? "active" : ""}" data-lane="Q" data-run-quality="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(runQualityGrade(card.score))}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="run-quality-score"><span>${escapeHTML(String(card.score))}</span><small>quality score</small></div>
    <div class="run-quality-gridline">
      <span>Signal</span><strong>${escapeHTML(card.signal)}</strong>
      <span>Risk</span><strong>${escapeHTML(card.risk)}</strong>
      <span>Gate</span><strong>${escapeHTML(card.gate)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-quality-select="${escapeHTML(card.id)}">质量简报</button>
      <button type="button" data-quality-draft="${escapeHTML(card.id)}">起草复查</button>
      ${card.runID ? `<button type="button" data-run-open="${escapeHTML(card.runID)}">观测运行</button>` : ""}
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
    </div>
  </article>`).join("");
  renderRunQualityDetail(cards.find((card) => card.id === state.selectedRunQuality) || cards[0]);
}

function renderRunQualityDetail(card) {
  const target = $("run-quality-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一张运行质量卡。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>评分</span><strong>${escapeHTML(`${card.score} (${runQualityGrade(card.score)})`)}</strong>
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>信号</h3>
      <p>${escapeHTML(card.signal)}</p>
      <h3>风险</h3>
      <p>${escapeHTML(card.risk)}</p>
      <h3>复查门槛</h3>
      <p>${escapeHTML(card.gate)}</p>
      <h3>建议路径</h3>
      <p>${escapeHTML(card.route)}</p>
      <div class="run-detail-actions">
        <button type="button" data-quality-draft="${escapeHTML(card.id)}">起草复查</button>
        ${card.runID ? `<button type="button" data-run-open="${escapeHTML(card.runID)}">观测运行</button>` : ""}
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function runQualityByID(id) {
  return runQualityCards().find((card) => card.id === id) || null;
}

function draftRunQualityToChat(id) {
  const card = runQualityByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a Run Quality review with score rationale, evidence, budget posture, risk, review gate, recommended route, and next action.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("运行质量复查已写入对话。");
}

function promptLabGoal() {
  const field = $("promptlab-goal");
  return (field?.value || state.promptLabGoal || state.workflowStageLabel || state.runs[0]?.prompt || "").trim();
}

function promptLabVariants() {
  const goal = promptLabGoal() || "Define the next Astria task and validation plan.";
  const agent = state.agents[0]?.name || "默认";
  const latestRun = state.runs[0];
  const latestCouncil = state.councilRuns[0];
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const compare = comparisonCandidates();
  const delivery = deliveryLanes();
  return [
    {
      id: "direct",
      label: "直接",
      title: "直接执行",
      panel: "chat",
      agent,
      context: "当前目标",
      source: "对话",
      risk: "最快路径；若目标需要证据或审查，可靠性会偏弱。",
      evaluation: "成功标准是一条具体实施步骤、明确验证方式，并且不发生范围漂移。",
      prompt: `Execute this Astria goal directly.\n\nGoal: ${goal}\n\nReturn the smallest useful implementation step, explain the validation command, and call out any risk before editing.`,
    },
    {
      id: "evidence",
      label: "证据",
      title: "证据优先实验",
      panel: "compare",
      agent,
      context: `${state.runs.length} 条运行，${memoryCount} 条记忆`,
      source: "路径比较台",
      risk: "速度较慢；适合过期上下文或隐藏回归风险较高的任务。",
      evaluation: "成功标准是在建议动作前引用运行、记忆或比较证据。",
      prompt: `Run an evidence-first Astria prompt experiment.\n\nGoal: ${goal}\nLatest run: ${latestRun?.prompt || "none"}\nComparison lanes: ${compare.map((lane) => lane.source).join(", ")}\nMemory items: ${memoryCount}\n\nUse the evidence to choose one next action and explain why alternatives are weaker.`,
    },
    {
      id: "council",
      label: "议会",
      title: "议会评审变体",
      panel: "council",
      agent,
      context: latestCouncil?.goal || "规划者 / 研究者 / 审阅者角色",
      source: "智能体议会",
      risk: "增加审查开销；适合取舍或协作成本较高的任务。",
      evaluation: "成功标准是规划、研究和审阅视角综合成一份可交接动作。",
      prompt: `Prepare a council-reviewed prompt variant.\n\nGoal: ${goal}\nExisting council goal: ${latestCouncil?.goal || "none"}\n\nSplit the goal into planner, researcher, and reviewer concerns, then synthesize the next concrete Astria action.`,
    },
    {
      id: "delivery",
      label: "交付",
      title: "交付就绪变体",
      panel: "delivery",
      agent,
      context: `${state.schedules.length} 条定时任务，${state.inboxProviders.length} 个渠道`,
      source: "主动交付",
      risk: "任何外部渠道交付前都需要审批边界。",
      evaluation: "成功标准是产出目标、审批门、产物和验证方式。",
      prompt: `Create a delivery-ready Astria prompt variant.\n\nGoal: ${goal}\nDelivery lanes: ${delivery.map((lane) => lane.source).join(", ")}\n\nReturn the message shape, destination assumption, approval gate, and validation checklist.`,
    },
  ];
}

function renderPromptExperimentLab() {
  const variants = promptLabVariants();
  setText("nav-promptlab-count", variants.length);
  setText("manage-promptlab-count", `${variants.length} 个变体`);
  setText("promptlab-summary", `${variants.length} 个 Prompt 变体，覆盖直接执行、证据优先、议会评审和交付路径。`);
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
      <span>上下文：${escapeHTML(variant.context)}</span>
      <span>来源：${escapeHTML(variant.source)}</span>
      <span>风险：${escapeHTML(variant.risk)}</span>
    </div>
    <div class="row-actions">
      <button type="button" data-prompt-variant-select="${escapeHTML(variant.id)}">变体简报</button>
      <button type="button" data-prompt-variant-draft="${escapeHTML(variant.id)}">起草变体</button>
      <button type="button" data-panel="${escapeHTML(variant.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderPromptVariantDetail(variants.find((variant) => variant.id === state.selectedPromptVariant) || variants[0]);
}

function renderPromptVariantDetail(variant) {
  const target = $("promptlab-detail");
  if (!target) return;
  if (!variant) {
    target.innerHTML = `<div class="empty-state">选择一个 Prompt 变体。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(variant.title)}</h3>
      <div class="run-meta-grid">
        <span>Agent</span><strong>${escapeHTML(variant.agent)}</strong>
        <span>上下文</span><strong>${escapeHTML(variant.context)}</strong>
        <span>来源</span><strong>${escapeHTML(variant.source)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>评估方式</h3>
      <p>${escapeHTML(variant.evaluation)}</p>
      <h3>风险</h3>
      <p>${escapeHTML(variant.risk)}</p>
      <div class="run-detail-actions">
        <button type="button" data-prompt-variant-draft="${escapeHTML(variant.id)}">起草变体</button>
        <button type="button" data-panel="${escapeHTML(variant.panel)}">打开来源</button>
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
  showToast("Prompt 变体已写入对话。");
}

function budgetGuardCards() {
  const goal = promptLabGoal() || "Define the next Astria task before launch.";
  const variants = promptLabVariants();
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const sourceCount = sourceRegistryRows().length;
  const resultCount = resultArchiveEntries().length;
  const snapshotCount = workspaceSnapshotCards().length;
  const failedRuns = state.runs.filter((run) => runHealthGroup(run) === "failed").length;
  const model = state.config?.model_tier || state.config?.openai_model || state.config?.ollama_model || "待填写模型";
  const provider = state.config?.provider || "provider";
  return [
    {
      id: "hard-cap",
      type: "Token 上限",
      title: "硬预算上限",
      panel: "chat",
      budget: "发起前设定 token / 时间最大包络。",
      trigger: "任务较长或目标模糊，存在开放式探索风险。",
      guardrail: "达到上限时停止，返回发现、阻塞项和最便宜的下一步。",
      fallback: "把范围裁剪到一个可验证交付物，扩展前先询问。",
      boundary: "这是规划守卫；此 UI 卡片不声称后端计费强制执行。",
      prompt: `Plan this Astria task with a hard budget cap.\n\nGoal: ${goal}\nProvider/model context: ${provider} / ${model}\nRecent runs: ${state.runs.length}\n\nDefine a token/time cap, what must fit inside it, what to stop doing first, what evidence is enough, and the next cheapest action if the cap is hit. Do not assume backend billing enforcement; write this as an operator-reviewed launch constraint.`,
    },
    {
      id: "model-route",
      type: "模型路线",
      title: "按复杂度选择模型路线",
      panel: "promptlab",
      budget: `${variants.length} 个 Prompt 变体可用于路由。`,
      trigger: "任务复杂度可能不足以使用最强模型或最长推理路线。",
      guardrail: "先分类为简单、证据密集、需要议会或交付敏感，再选择路线。",
      fallback: "先用更便宜的直接路线；只有证据或测试失败时才升级。",
      boundary: "除非显式修改运行时配置，模型路由只是 Prompt 计划。",
      prompt: `Plan an Astria complexity-based model route.\n\nGoal: ${goal}\nPrompt variants: ${variants.map((variant) => variant.label).join(", ")}\nProvider/model context: ${provider} / ${model}\n\nClassify the task as simple, evidence-heavy, council-worthy, or delivery-sensitive. Choose the cheapest safe route first, define escalation triggers, and state when fallback to a stronger model or reviewer path is justified.`,
    },
    {
      id: "context-trim",
      type: "上下文",
      title: "上下文裁剪",
      panel: "snapshot",
      budget: `${memoryCount} 条记忆，${sourceCount} 个来源，${snapshotCount} 个快照包`,
      trigger: "过大的本地上下文可能淹没任务或抬高成本。",
      guardrail: "只使用能直接证明需求、风险或验证状态的上下文。",
      fallback: "打开工作区快照，选择更小的续接/证据包。",
      boundary: "绝不裁掉明确用户需求、安全约束或验证失败记录。",
      prompt: `Plan an Astria context trimming pass.\n\nGoal: ${goal}\nMemory entries: ${memoryCount}\nSources: ${sourceCount}\nSnapshot packs: ${snapshotCount}\n\nChoose the smallest context set needed to proceed. Keep explicit requirements, validation state, risks, and relevant evidence. Exclude stale, duplicate, private, or speculative context unless it changes the decision.`,
    },
    {
      id: "fallback",
      type: "降级",
      title: "自动降级阶梯",
      panel: "diagnostics",
      budget: `${failedRuns} 条失败运行；诊断 ${uiTerm(state.diagnostics?.status || "unknown")}`,
      trigger: "运行时就绪、provider 设置或历史失败可能让首选路线不可靠。",
      guardrail: "发起前定义降级顺序：缩小重试、切换路线、请求审批或停止。",
      fallback: "先打开诊断，再在改变 provider 或模型设置前缩小范围。",
      boundary: "没有操作者明确动作，不改变远端 provider 或账号状态。",
      prompt: `Plan an Astria fallback ladder.\n\nGoal: ${goal}\nDiagnostics: ${state.diagnostics?.status || "unknown"}\nFailed runs: ${failedRuns}\nProvider/model context: ${provider} / ${model}\n\nDefine the fallback order if the first route fails: smaller retry, evidence-only pass, different agent/variant, diagnostics repair, or stop and ask. Include what evidence proves each fallback is needed.`,
    },
    {
      id: "stop-rules",
      type: "停止规则",
      title: "长运行停止规则",
      panel: "runs",
      budget: `${state.runs.length} 条运行；${state.approvals.size} 个待审批`,
      trigger: "任务可能进入重复调试、宽泛调研或无限工具使用。",
      guardrail: "遇到同类重复失败、缺少需求证据或审批边界时停止。",
      fallback: "总结当前证据，并创建更窄的后续任务。",
      boundary: "越过破坏性、外发、购买或账号变更边界后，不继续使用工具。",
      prompt: `Plan Astria long-run stop rules.\n\nGoal: ${goal}\nRuns: ${state.runs.length}\nPending approvals: ${state.approvals.size}\n\nDefine stop conditions for repeated failures, missing evidence, destructive boundaries, external delivery, and uncertainty. Include what summary to return when stopping and how to create the next narrower mission.`,
    },
    {
      id: "schedule-limit",
      type: "定时",
      title: "定时工作预算",
      panel: "schedules",
      budget: `${state.schedules.length} 条定时任务；${deliveryLanes().length} 条交付链路`,
      trigger: "若频率过宽，周期任务会悄悄消耗时间、token 或注意力。",
      guardrail: "每条定时任务都需要频率、最大投入、输出形态、审批门和停用条件。",
      fallback: "启用周期任务前，先手动运行一次并复查输出。",
      boundary: "定时任务不应暗示外发或无人值守的远端状态变更。",
      prompt: `Plan an Astria scheduled-work budget.\n\nGoal: ${goal}\nSchedules: ${state.schedules.length}\nDelivery lanes: ${deliveryLanes().length}\n\nDefine cadence, max effort, output shape, approval gate, disable condition, and manual dry-run requirements before any recurring task is enabled.`,
    },
    {
      id: "evidence-cost",
      type: "证据",
      title: "证据成本取舍",
      panel: "citation",
      budget: `${sourceCount} 个来源；${resultCount} 个产物条目`,
      trigger: "调研可能过度收集来源，或让关键结论证据不足。",
      guardrail: "把结论映射到最低充分证据，达到置信阈值后停止。",
      fallback: "只把未支撑或高影响结论升级到深度调研。",
      boundary: "不要为低影响结论投入超过所需置信度的证明成本。",
      prompt: `Plan an Astria evidence-cost tradeoff.\n\nGoal: ${goal}\nSources: ${sourceCount}\nResult entries: ${resultCount}\n\nIdentify claims, required confidence, minimum sufficient evidence, when to stop collecting, and which high-impact gaps deserve deeper research. Keep unsupported claims visibly downgraded.`,
    },
  ];
}

function renderBudgetGuardPlanner() {
  const cards = budgetGuardCards();
  setText("nav-budget-count", cards.length);
  setText("manage-budget-count", `${cards.length} 条守卫`);
  setText("budget-summary", `${cards.length} 条预算守卫，覆盖 token 上限、模型路由、上下文裁剪、fallback、停止规则、定时任务和证据成本。`);
  const list = $("budget-guard-grid");
  if (!list) return;
  if (!state.selectedBudgetGuard || !cards.some((card) => card.id === state.selectedBudgetGuard)) {
    state.selectedBudgetGuard = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="budget-guard-card ${card.id === state.selectedBudgetGuard ? "active" : ""}" data-lane="B" data-budget-guard="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.panel)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="budget-guard-gridline">
      <span>预算</span><strong>${escapeHTML(card.budget)}</strong>
      <span>触发</span><strong>${escapeHTML(card.trigger)}</strong>
      <span>护栏</span><strong>${escapeHTML(card.guardrail)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-budget-select="${escapeHTML(card.id)}">预算简报</button>
      <button type="button" data-budget-draft="${escapeHTML(card.id)}">起草守卫</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
    </div>
  </article>`).join("");
  renderBudgetGuardDetail(cards.find((card) => card.id === state.selectedBudgetGuard) || cards[0]);
}

function renderBudgetGuardDetail(card) {
  const target = $("budget-guard-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一条预算守卫。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
        <span>预算</span><strong>${escapeHTML(card.budget)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>触发条件</h3>
      <p>${escapeHTML(card.trigger)}</p>
      <h3>执行护栏</h3>
      <p>${escapeHTML(card.guardrail)}</p>
      <h3>Fallback</h3>
      <p>${escapeHTML(card.fallback)}</p>
      <h3>复查边界</h3>
      <p>${escapeHTML(card.boundary)}</p>
      <div class="run-detail-actions">
        <button type="button" data-budget-draft="${escapeHTML(card.id)}">起草守卫</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function budgetGuardByID(id) {
  return budgetGuardCards().find((card) => card.id === id) || null;
}

function draftBudgetGuardToChat(id) {
  const card = budgetGuardByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a Budget Guard launch brief with budget shape, trigger, guardrail, fallback route, review boundary, stop condition, and validation plan.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("预算守卫已写入对话。");
}

function resultArtifactRuns() {
  const byID = new Map();
  const add = (run) => {
    if (!run?.id) return;
    const health = runHealthGroup(run);
    const hasResponse = Boolean(formatRunResponse(run.response).trim());
    if (health !== "completed" && !hasResponse) return;
    byID.set(run.id, { ...(byID.get(run.id) || {}), ...run });
  };
  add(state.missionRunDetail);
  add(state.currentRunDetail);
  state.runs.forEach(add);
  return Array.from(byID.values()).sort((a, b) => String(b.started_at || "").localeCompare(String(a.started_at || "")));
}

function resultArtifactSummary(run) {
  const text = formatRunResponse(run?.response).trim();
  if (text) return compactText(text, 180);
  return compactText(run?.prompt || run?.id || "本地运行产物", 180);
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
  const completedRun = resultArtifactRuns()[0] || latestRun;
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
      readiness: `${summary.allow.length} 个工具，${summary.deny.length} 个拦截`,
      evidence: summary.model,
      reuse: summary.description,
      action: "使用此配置发起任务，并延续它的运行约束。",
      prompt: `Reuse this Astria agent profile for the next mission.\n\nAgent: ${summary.name}\nModel: ${summary.model}\nReasoning: ${summary.reasoning}\nTools allowed: ${summary.allow.join(", ") || "default"}\nAuto approve: ${summary.autoApprove}\n\nDescribe the mission, choose whether this profile fits, and call out any safety constraint before acting.`,
    });
  } else {
    assets.push({
      id: "agent-default",
      kind: "Agent",
      title: "默认 Agent 起点",
      panel: "agents",
      readiness: "默认",
      evidence: "尚无命名配置",
      reuse: "在命名 Agent 配置建立前，使用默认 daemon Agent。",
      action: "先起草一个聚焦角色，再创建可复用配置。",
      prompt: "Create a reusable Astria agent profile for this workspace. Include role, model expectations, tool boundaries, memory needs, and one saved command.",
    });
  }

  commandAssets.slice(0, 2).forEach((command) => {
    const agentName = normalizeName(command.agent);
    assets.push({
      id: `command-${agentName}-${command.name}`,
      kind: "命令",
      title: `/${command.name}`,
      panel: "agents",
      readiness: agentName,
      evidence: "已保存命令",
      reuse: String(command.body || "").slice(0, 150) || "已保存命令内容",
      action: "把这条命令连同 Agent 配置写入对话。",
      prompt: `Reuse this saved Astria command.\n\nAgent: ${agentName}\nCommand: /${command.name}\n\n${String(command.body || "")}`,
      agent: agentName,
    });
  });

  sources.slice(0, 2).forEach((source) => {
    assets.push({
      id: `source-${source.id}`,
      kind: "知识",
      title: source.title,
      panel: source.panel,
      readiness: source.reliability,
      evidence: `${source.evidence} 条证据`,
      reuse: source.action,
      action: "发起下一次任务前，用此来源校准上下文。",
      prompt: `Reuse this Astria knowledge source as mission context.\n\nSource: ${source.title}\nType: ${source.type}\nFreshness: ${source.freshness}\nReliability: ${source.reliability}\nEvidence: ${source.evidence}\n\nDraft the next task using only what this source can reliably support.`,
    });
  });

  if (completedRun) {
    assets.push({
      id: `run-${completedRun.id || "latest"}`,
      kind: "结果",
      title: completedRun.prompt || completedRun.id || "Latest run outcome",
      panel: "runs",
      runID: completedRun.id || "",
      readiness: uiTerm(completedRun.status || "unknown"),
      evidence: completedRun.agent || "默认 Agent",
      reuse: resultArtifactSummary(completedRun),
      action: "把这条运行转成下一次任务起点。",
      prompt: `Reuse this Astria run outcome as the next starting point.\n\nRun: ${completedRun.id || "unknown"}\nStatus: ${completedRun.status || "unknown"}\nAgent: ${completedRun.agent || "default"}\nPrompt: ${completedRun.prompt || ""}\n\nSummarize what can be reused, what remains uncertain, and the next concrete action with validation.`,
    });
  }

  if (latestCouncil) {
    assets.push({
      id: `council-${latestCouncil.id || "latest"}`,
      kind: "议会",
      title: latestCouncil.goal || "议会综合",
      panel: "council",
      readiness: latestCouncil.synthesis ? "已综合" : "待复查",
      evidence: `${(latestCouncil.roles || []).length} 个角色`,
      reuse: latestCouncil.synthesis || "Planner, researcher, and reviewer context can seed the next handoff.",
      action: "把已审查综合结果复用为任务简报。",
      prompt: `Reuse this Astria council result as the next mission brief.\n\nGoal: ${latestCouncil.goal || "none"}\nRoles: ${(latestCouncil.roles || []).map((role) => role.role).join(", ") || "none"}\nSynthesis:\n${latestCouncil.synthesis || ""}\n\nTurn the synthesis into one executable next step and validation plan.`,
    });
  } else {
    assets.push({
      id: "council-starter",
      kind: "议会",
      title: "议会评审起点",
      panel: "council",
      readiness: "待建立",
      evidence: "规划者 / 研究者 / 审阅者",
      reuse: "当可复用决策需要更强证据时，使用多角色评审。",
      action: "起草一份可进入议会的任务简报。",
      prompt: "Create a reusable council mission brief. Split the work into planner, researcher, and reviewer concerns, then define the synthesis criteria.",
    });
  }

  return assets.slice(0, 8);
}

function renderReuseGallery() {
  const assets = reuseGalleryAssets();
  setText("nav-reuse-count", assets.length);
  setText("manage-reuse-count", `${assets.length} 个星体`);
  setText("reuse-summary", `${assets.length} 个可复用星体，来自 Prompt、Agent、知识来源、运行结果和多 Agent 评议。`);
  const list = $("reuse-gallery-assets");
  if (!list) return;
  if (!state.selectedReuseAsset || !assets.some((asset) => asset.id === state.selectedReuseAsset)) {
    state.selectedReuseAsset = assets[0]?.id || "";
  }
  list.innerHTML = assets.map((asset) => `<article class="reuse-asset ${asset.id === state.selectedReuseAsset ? "active" : ""}" data-reuse-asset="${escapeHTML(asset.id)}">
    <div class="row-item-title"><span>${escapeHTML(asset.kind)}</span><span class="tag">${escapeHTML(asset.readiness)}</span></div>
    <strong>${escapeHTML(asset.title)}</strong>
    <div class="reuse-grid">
      <span>证据</span><strong>${escapeHTML(asset.evidence)}</strong>
      <span>复用价值</span><strong>${escapeHTML(asset.reuse)}</strong>
      <span>下一步</span><strong>${escapeHTML(asset.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-reuse-select="${escapeHTML(asset.id)}">星体简报</button>
      <button type="button" data-reuse-draft="${escapeHTML(asset.id)}">起草任务</button>
      ${asset.runID ? `<button type="button" data-run-open="${escapeHTML(asset.runID)}">观测运行</button>` : `<button type="button" data-panel="${escapeHTML(asset.panel)}">打开来源</button>`}
    </div>
  </article>`).join("");
  renderReuseAssetDetail(assets.find((asset) => asset.id === state.selectedReuseAsset) || assets[0]);
}

function renderReuseAssetDetail(asset) {
  const target = $("reuse-gallery-detail");
  if (!target) return;
  if (!asset) {
    target.innerHTML = `<div class="empty-state">选择一个可复用星体。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(asset.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(asset.kind)}</strong>
        <span>状态</span><strong>${escapeHTML(asset.readiness)}</strong>
        <span>路线</span><strong>${escapeHTML(asset.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>复用价值</h3>
      <p>${escapeHTML(asset.reuse)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(asset.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-reuse-draft="${escapeHTML(asset.id)}">起草任务</button>
        ${asset.runID ? `<button type="button" data-run-open="${escapeHTML(asset.runID)}">观测运行</button>` : `<button type="button" data-panel="${escapeHTML(asset.panel)}">打开来源</button>`}
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
  showToast("复用起点已写入对话。");
}

function resultArchiveEntries() {
  const entries = [];
  const latestRun = state.runs[0];
  const completedRun = resultArtifactRuns()[0] || latestRun;
  const latestCouncil = state.councilRuns[0];
  const shareCards = sharePackCards();
  const dataCards = dataInsightCards();
  const citationCards = citationGroundingCards();
  const reuseAssets = reuseGalleryAssets();

  if (completedRun) {
    entries.push({
      id: `run-report-${completedRun.id || "latest"}`,
      type: "运行报告",
      title: completedRun.prompt || completedRun.id || "最近 Astria 运行",
      source: completedRun.agent || "默认 Agent",
      panel: "runs",
      evidence: completedRun.id || "local run record",
      review: uiTerm(completedRun.status || "unknown"),
      freshness: completedRun.created_at || completedRun.updated_at || "最近本地运行",
      reuse: resultArtifactSummary(completedRun),
      action: "复查真实运行结果、未解决风险和下一条已验证动作。",
      prompt: `Open an Astria result archive follow-up for a completed run.\n\nResult: ${completedRun.prompt || completedRun.id || "Latest run"}\nRun id: ${completedRun.id || "unknown"}\nStatus: ${completedRun.status || "unknown"}\nAgent: ${completedRun.agent || "default"}\n\nExtract what was produced, what evidence supports it, what remains unresolved, and the best next mission starter.`,
    });
  } else {
    entries.push({
      id: "run-report-seed",
      type: "运行报告",
      title: "第一份完成产物",
      source: "运行",
      panel: "runs",
      evidence: "尚无完成运行",
      review: "待建立",
      freshness: "等待运行",
      reuse: "完成运行会以本地报告形式出现在这里。",
      action: "先发起任务，再把结果归档为可复用产物。",
      prompt: "Plan the first Astria result worth saving. Define the target output, evidence needed, review gate, and how the result should be reused later.",
    });
  }

  if (shareCards[0]) {
    const card = shareCards[0];
    entries.push({
      id: `share-result-${card.id}`,
      type: "交接包",
      title: card.title,
      source: "交接包",
      panel: "share",
      evidence: card.evidence,
      review: card.readiness,
      freshness: "本地交接上下文",
      reuse: card.action,
      action: "把交接包转成后续任务或审查清单。",
      prompt: `${card.prompt}\n\nArchive review: identify the durable result, evidence included, boundaries, freshness notes, and the next reusable launch path.`,
    });
  }

  if (dataCards[1] || dataCards[0]) {
    const card = dataCards[1] || dataCards[0];
    entries.push({
      id: `data-result-${card.id}`,
      type: "洞察简报",
      title: card.title,
      source: "数据规划器",
      panel: "data",
      evidence: card.evidence,
      review: card.readiness,
      freshness: "取决于来源提取日期",
      reuse: "把发现保存为可审查记忆或可复用分析模式。",
      action: card.action,
      prompt: `${card.prompt}\n\nArchive review: produce a saved insight brief with observed findings, source limits, freshness date, reusable memory candidates, and follow-up analysis.`,
    });
  }

  if (citationCards[0]) {
    const card = citationCards[0];
    entries.push({
      id: `citation-result-${card.id}`,
      type: "引用简报",
      title: card.title,
      source: "引用校准",
      panel: "citation",
      evidence: card.evidence,
      review: card.readiness,
      freshness: "需要来源日期",
      reuse: "把结论图谱和证据缺口带入下一次回答或交接。",
      action: "先解决未支撑结论，再把结果视为最终产物。",
      prompt: `${card.prompt}\n\nArchive review: save the claim map, accepted citations, missing evidence, source freshness risks, and safe wording for reuse.`,
    });
  }

  if (reuseAssets[0]) {
    const asset = reuseAssets[0];
    entries.push({
      id: `reuse-result-${asset.id}`,
      type: "可复用产物",
      title: asset.title,
      source: "复用星库",
      panel: "reuse",
      evidence: asset.evidence,
      review: asset.readiness,
      freshness: "模式复查",
      reuse: asset.reuse,
      action: asset.action,
      prompt: `${asset.prompt}\n\nArchive review: decide whether this result should become a reusable starter, what context it requires, and how to validate it next time.`,
    });
  }

  if (latestCouncil) {
    entries.push({
      id: `council-result-${latestCouncil.id || "latest"}`,
      type: "议会综合",
      title: latestCouncil.goal || "议会综合",
      source: "智能体议会",
      panel: "council",
      evidence: `${(latestCouncil.roles || []).length} 份角色简报`,
      review: latestCouncil.synthesis ? "已综合" : "待复查",
      freshness: "当前议会运行",
      reuse: latestCouncil.synthesis || "把角色拆分作为已审查决策起点。",
      action: "把综合结果转成一个可执行下一步。",
      prompt: `Archive this Astria council result.\n\nGoal: ${latestCouncil.goal || "none"}\nRoles: ${(latestCouncil.roles || []).map((role) => role.role).join(", ") || "none"}\nSynthesis:\n${latestCouncil.synthesis || ""}\n\nReturn a saved result brief with decision, evidence, dissent or uncertainty, and the next action.`,
    });
  }

  return entries.slice(0, 8);
}

function resultArchiveStats(entries) {
  const sourceTypes = new Set(entries.map((entry) => entry.source).filter(Boolean));
  const reusable = entries.filter((entry) => String(entry.reuse || "").trim()).length;
  const needsReview = entries.filter((entry) => {
    const review = String(entry.review || "").toLowerCase();
    return review.includes("待") || review.includes("unknown") || review.includes("failed") || review.includes("error");
  }).length;
  const runBacked = entries.filter((entry) => entry.panel === "runs" || String(entry.evidence || "").includes("run")).length;
  return {
    total: entries.length,
    reusable,
    needsReview,
    sourceTypes: sourceTypes.size,
    runBacked,
  };
}

function renderArtifactReadinessBoard(entries) {
  const target = $("artifact-readiness-board");
  if (!target) return;
  const stats = resultArchiveStats(entries);
  const cards = [
    ["archive", "归档产物", stats.total, "可在本地复查的结果条目"],
    ["reuse", "可复用", stats.reusable, "具备下一步路线或 Prompt 起点"],
    ["review", "待复查", stats.needsReview, "缺少证据、鲜度或明确验收"],
    ["source", "来源覆盖", stats.sourceTypes, `${stats.runBacked} 条连接真实运行`],
  ];
  target.innerHTML = cards.map(([key, label, value, hint]) => `<button type="button" class="artifact-readiness-card ${escapeHTML(key)}" data-result-board="${escapeHTML(key)}">
    <span>${escapeHTML(label)}</span>
    <strong>${escapeHTML(String(value))}</strong>
    <small>${escapeHTML(hint)}</small>
  </button>`).join("");
}

function renderResultLibrary() {
  const entries = resultArchiveEntries();
  setText("nav-results-count", entries.length);
  setText("manage-results-count", `${entries.length} 个产物`);
  setText("results-summary", `${entries.length} 个归档产物，来自运行、交接包、洞察简报、引用校准、复用输出和多 Agent 综合。`);
  renderArtifactReadinessBoard(entries);
  const list = $("result-library-grid");
  if (!list) return;
  if (!state.selectedResultArchive || !entries.some((entry) => entry.id === state.selectedResultArchive)) {
    state.selectedResultArchive = entries[0]?.id || "";
  }
  list.innerHTML = entries.map((entry) => `<article class="result-library-card ${entry.id === state.selectedResultArchive ? "active" : ""}" data-result-archive="${escapeHTML(entry.id)}">
    <div class="row-item-title"><span>${escapeHTML(entry.type)}</span><span class="tag">${escapeHTML(entry.review)}</span></div>
    <strong>${escapeHTML(entry.title)}</strong>
    <div class="result-library-gridline">
      <span>来源</span><strong>${escapeHTML(entry.source)}</strong>
      <span>证据</span><strong>${escapeHTML(entry.evidence)}</strong>
      <span>复用路线</span><strong>${escapeHTML(entry.reuse)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-result-select="${escapeHTML(entry.id)}">产物简报</button>
      <button type="button" data-result-draft="${escapeHTML(entry.id)}">起草后续</button>
      <button type="button" data-panel="${escapeHTML(entry.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderResultLibraryDetail(entries.find((entry) => entry.id === state.selectedResultArchive) || entries[0]);
}

function renderResultLibraryDetail(entry) {
  const target = $("result-library-detail");
  if (!target) return;
  if (!entry) {
    target.innerHTML = `<div class="empty-state">选择一个归档产物。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="artifact-review-card">
      <div class="artifact-review-head">
        <div>
          <span>${escapeHTML(entry.type)}</span>
          <strong>${escapeHTML(entry.title)}</strong>
          <small>${escapeHTML(entry.source)} · ${escapeHTML(entry.freshness)}</small>
        </div>
        <b>${escapeHTML(entry.review)}</b>
      </div>
      <div class="artifact-review-grid">
        <span>证据</span><strong>${escapeHTML(entry.evidence)}</strong>
        <span>复用</span><strong>${escapeHTML(entry.reuse)}</strong>
        <span>下一步</span><strong>${escapeHTML(entry.action)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>${escapeHTML(entry.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(entry.type)}</strong>
        <span>审查</span><strong>${escapeHTML(entry.review)}</strong>
        <span>路线</span><strong>${escapeHTML(entry.panel)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>鲜度</h3>
      <p>${escapeHTML(entry.freshness)}</p>
      <h3>复用路线</h3>
      <p>${escapeHTML(entry.reuse)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(entry.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-result-draft="${escapeHTML(entry.id)}">起草后续</button>
        <button type="button" data-panel="${escapeHTML(entry.panel)}">打开来源</button>
      </div>
    </section>
  </div>`;
}

function resultArchiveByID(id) {
  return resultArchiveEntries().find((entry) => entry.id === id) || null;
}

function draftResultArchiveToChat(id) {
  const entry = resultArchiveByID(id);
  if (!entry) return;
  $("chat-input").value = `${entry.prompt}\n\nReturn a result archive follow-up with saved outcome, source evidence, freshness, reuse path, open risks, and next Astria route.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("产物后续已写入对话。");
}

function playbookLibraryCards() {
  const starterCount = starterKits().length;
  const resultCount = resultArchiveEntries().length;
  const reuseCount = reuseGalleryAssets().length;
  const sourceCount = sourceRegistryRows().length;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const councilCount = state.councilRuns.length;
  const deliveryCount = deliveryLanes().length;
  const browserTarget = browserMissionContext().url;
  const dataSource = dataInsightContext().source;
  return [
    {
      id: "reviewed-research",
      type: "研究",
      title: "已审查证据调研",
      route: "browser",
      trigger: "网页或产品结论在决策前需要当前、可引用的证据。",
      evidenceGate: "可见来源、带日期摘要、引用图谱和未支撑结论清单。",
      safety: "只读浏览；表单、登录态、下载、购买、发布或账号变更前先询问。",
      reusableOutput: "可进入产物星库、引用校准和交接包的引用简报。",
      next: `从浏览器规划器开始，目标为 ${browserTarget}。`,
      prompt: `Run the Astria reviewed evidence research playbook.\n\nTrigger: verify a web or product claim with current evidence.\nCurrent target: ${browserTarget}\nAvailable sources: ${sourceCount}\n\nSteps:\n1. Define the exact claim and required freshness.\n2. Inspect sources read-only.\n3. Capture citations, dates, and gaps.\n4. Produce safe wording and next route.\n\nEvidence gate: visible source, dated summary, citation map, unsupported-claim list.\nSafety boundary: no forms, account changes, downloads, purchases, posts, or destructive actions without approval.\nReusable output: cited brief for Result Library, Citation Planner, and Share Pack.`,
    },
    {
      id: "data-insight",
      type: "数据",
      title: "可审查数据洞察",
      route: "data",
      trigger: "本地文件、表格、指标或导出需要形成可决策发现。",
      evidenceGate: "来源描述、schema 或字段限制、观察发现、不确定性和鲜度日期。",
      safety: "不推断隐藏字段或因果解释；区分观察和假设。",
      reusableOutput: "洞察简报、记忆候选和可复用分析 Prompt。",
      next: `从数据规划器开始，来源为 ${dataSource}。`,
      prompt: `Run the Astria reviewable data insight playbook.\n\nTrigger: turn local data into a decision-ready finding.\nCurrent source: ${dataSource}\nMemory entries: ${memoryCount}\n\nSteps:\n1. Profile the source and freshness.\n2. State what can and cannot be answered.\n3. Produce findings with evidence and caveats.\n4. Save durable facts as reviewed memory candidates.\n\nEvidence gate: source descriptor, schema or field limits, observed findings, uncertainty, freshness date.\nSafety boundary: do not invent hidden fields or causal explanations.\nReusable output: insight brief, memory candidates, reusable analysis prompt.`,
    },
    {
      id: "handoff-pack",
      type: "交接",
      title: "本地交接包",
      route: "share",
      trigger: "有价值工作需要被未来会话、审查者或队友继续。",
      evidenceGate: "摘要、包含产物、来源鲜度、边界、验收检查和下一步。",
      safety: "保持本地交接；脱敏密钥、私有路径和只有当前会话知道的假设。",
      reusableOutput: "可复制交接包章节和后续 Prompt。",
      next: `使用 ${resultCount} 个产物归档和 ${reuseCount} 个可复用星体。`,
      prompt: `Run the Astria local handoff package playbook.\n\nTrigger: package useful work for a future session, reviewer, or teammate.\nResult archive entries: ${resultCount}\nReusable assets: ${reuseCount}\n\nSteps:\n1. Summarize the durable result.\n2. List evidence and freshness.\n3. State boundaries, risks, and redactions.\n4. Write a next-action checklist.\n\nEvidence gate: summary, artifacts, source freshness, boundaries, acceptance checks, next actions.\nSafety boundary: local-only handoff; redact secrets and private paths.\nReusable output: copyable Share Pack section and follow-up prompt.`,
    },
    {
      id: "citation-grounding",
      type: "引用",
      title: "结论溯源复查",
      route: "citation",
      trigger: "回答、简报或产物包含需要可靠支撑的结论。",
      evidenceGate: "原子结论图谱、来源路线、引用摘要、鲜度检查和缺口升级。",
      safety: "证据薄弱、过期、私密或冲突时，阻止过度自信表述。",
      reusableOutput: "可复用的引用简报和安全措辞。",
      next: `在引用校准中使用 ${sourceCount} 条来源路线。`,
      prompt: `Run the Astria claim grounding review playbook.\n\nTrigger: claims need reliable support before reuse or delivery.\nSource lanes: ${sourceCount}\n\nSteps:\n1. Split the answer into atomic claims.\n2. Match claims to source lanes.\n3. Capture citations, freshness, and gaps.\n4. Rewrite unsafe claims with uncertainty.\n\nEvidence gate: atomic claim map, source lane, quote or cited summary, freshness check, gap escalation.\nSafety boundary: block confident wording when evidence is weak, stale, private, or conflicting.\nReusable output: citation brief and safe wording.`,
    },
    {
      id: "agent-profile",
      type: "Agent",
      title: "聚焦 Agent 配置",
      route: "agents",
      trigger: "重复工作流需要命名 Agent，并明确角色、记忆和工具边界。",
      evidenceGate: "角色、模型姿态、允许工具、拒绝工具、记忆需求、命令和测试 Prompt。",
      safety: "保持权限收窄，避免宽泛自动批准默认值。",
      reusableOutput: "命名 Agent 配置和已保存命令起点。",
      next: `复查 ${state.agents.length} 个当前 Agent 配置。`,
      prompt: `Run the Astria focused agent profile playbook.\n\nTrigger: repeated workflow needs a named agent.\nCurrent agents: ${state.agents.length}\nStarter kits: ${starterCount}\n\nSteps:\n1. Define one repeatable job.\n2. Specify role, model posture, memory needs, and tool boundaries.\n3. Write one saved command and a test prompt.\n4. Validate permissions before reuse.\n\nEvidence gate: role, model posture, allowed tools, denied tools, memory needs, command, test prompt.\nSafety boundary: narrow permissions, no broad auto-approval.\nReusable output: named agent profile plus saved command starter.`,
    },
    {
      id: "memory-curation",
      type: "记忆",
      title: "长期记忆整理",
      route: "memory",
      trigger: "结果或会话产生了值得记住的事实、偏好、决策、风险或命令。",
      evidenceGate: "来源、分类、长期价值理由、鲜度说明、重复/冲突检查和拒绝标准。",
      safety: "只保存长期事实；拒绝模糊、过期、敏感或无支撑记录。",
      reusableOutput: "已审查记忆候选和分类改进记录。",
      next: `复查 ${memoryCount} 条记忆和 ${resultCount} 个归档产物。`,
      prompt: `Run the Astria durable memory curation playbook.\n\nTrigger: decide what from recent work should become durable memory.\nMemory entries: ${memoryCount}\nResult archive entries: ${resultCount}\n\nSteps:\n1. Extract candidate facts from results and sessions.\n2. Categorize each candidate.\n3. Check source, freshness, duplicate, and conflict risk.\n4. Approve only durable, useful memory.\n\nEvidence gate: source, category, durability reason, freshness note, duplicate/conflict check, rejection criteria.\nSafety boundary: reject vague, stale, sensitive, or unsupported notes.\nReusable output: reviewed memory candidates and taxonomy notes.`,
    },
    {
      id: "delivery-review",
      type: "交付",
      title: "审批优先交付",
      route: "delivery",
      trigger: "结果可能需要出站交付、定时任务或外部渠道跟进。",
      evidenceGate: "目标、产物、审批边界、定时或渠道、验证和回滚步骤。",
      safety: "没有明确审批，不外发、不定时、不改变远端状态。",
      reusableOutput: "交付链路清单和已审查出站 Prompt。",
      next: `复查 ${deliveryCount} 条交付链路和 ${state.schedules.length} 条定时任务。`,
      prompt: `Run the Astria approval-first delivery playbook.\n\nTrigger: result may need outbound delivery or scheduling.\nDelivery lanes: ${deliveryCount}\nSchedules: ${state.schedules.length}\n\nSteps:\n1. Identify destination and artifact.\n2. Define approval boundary and verification.\n3. Choose schedule or channel only after review.\n4. Include rollback and confirmation steps.\n\nEvidence gate: destination, artifact, approval boundary, schedule or channel, verification, rollback step.\nSafety boundary: no external send, schedule, or remote state change without explicit approval.\nReusable output: delivery checklist and reviewed outbound prompt.`,
    },
    {
      id: "council-decision",
      type: "议会",
      title: "多角色决策复查",
      route: "council",
      trigger: "决策足够重要，需要规划者、研究者和审阅者视角。",
      evidenceGate: "角色简报、分歧或风险记录、综合结论、验收标准和下一条可执行步骤。",
      safety: "综合解决冲突和缺口前，不把角色输出当成最终结论。",
      reusableOutput: "可进入产物星库或交接包的议会综合。",
      next: `参考 ${councilCount} 条议会运行。`,
      prompt: `Run the Astria multi-role decision review playbook.\n\nTrigger: decision needs planner, researcher, and reviewer perspectives.\nCouncil runs: ${councilCount}\n\nSteps:\n1. Split the decision into planning, research, and review concerns.\n2. Require each role to state evidence and uncertainty.\n3. Synthesize agreement, disagreement, and gaps.\n4. Produce one executable next step.\n\nEvidence gate: role briefs, disagreement or risk notes, synthesis, acceptance criteria, next executable step.\nSafety boundary: role output is not final until conflicts and gaps are resolved.\nReusable output: council synthesis for Result Library or Share Pack.`,
    },
  ];
}

function renderPlaybookLibrary() {
  const cards = playbookLibraryCards();
  setText("nav-playbooks-count", cards.length);
  setText("manage-playbooks-count", `${cards.length} 本手册`);
  setText("playbooks-summary", `${cards.length} 本已审核本地实践手册，覆盖研究、数据、交接、引用、Agent、记忆、交付和议会评审。`);
  const list = $("playbook-library-grid");
  if (!list) return;
  if (!state.selectedPlaybook || !cards.some((card) => card.id === state.selectedPlaybook)) {
    state.selectedPlaybook = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="playbook-card ${card.id === state.selectedPlaybook ? "active" : ""}" data-playbook="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.route)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="playbook-gridline">
      <span>触发</span><strong>${escapeHTML(card.trigger)}</strong>
      <span>证据门槛</span><strong>${escapeHTML(card.evidenceGate)}</strong>
      <span>可复用产物</span><strong>${escapeHTML(card.reusableOutput)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-playbook-select="${escapeHTML(card.id)}">手册简报</button>
      <button type="button" data-playbook-draft="${escapeHTML(card.id)}">起草手册</button>
      <button type="button" data-panel="${escapeHTML(card.route)}">打开路径</button>
    </div>
  </article>`).join("");
  renderPlaybookDetail(cards.find((card) => card.id === state.selectedPlaybook) || cards[0]);
}

function renderPlaybookDetail(card) {
  const target = $("playbook-library-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一本实践手册。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.route))}</strong>
        <span>下一步</span><strong>${escapeHTML(card.next)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>触发条件</h3>
      <p>${escapeHTML(card.trigger)}</p>
      <h3>证据门槛</h3>
      <p>${escapeHTML(card.evidenceGate)}</p>
      <h3>安全边界</h3>
      <p>${escapeHTML(card.safety)}</p>
      <h3>可复用产物</h3>
      <p>${escapeHTML(card.reusableOutput)}</p>
      <div class="run-detail-actions">
        <button type="button" data-playbook-draft="${escapeHTML(card.id)}">起草手册</button>
        <button type="button" data-panel="${escapeHTML(card.route)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function playbookByID(id) {
  return playbookLibraryCards().find((card) => card.id === id) || null;
}

function draftPlaybookToChat(id) {
  const card = playbookByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a playbook launch brief with trigger, steps, evidence gate, safety boundary, reusable output, and next Astria route.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("实践手册已写入对话。");
}

function starterKits() {
  const agentCount = state.agents.length;
  const sourceCount = sourceRegistryRows().length;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const reuseCount = reuseGalleryAssets().length;
  return [
    {
      id: "browser-research",
      type: "浏览器",
      title: "已审查网页调研",
      route: "browser",
      agent: agentCount ? "researcher" : "默认",
      evidence: "浏览器检查 + 引用",
      reuse: "交接包证据章节",
      safety: "只读导航；表单、账号动作或下载前先询问。",
      objective: "检查网页、捕获证据，并生成带引用的决策简报。",
      prompt: "Launch the Astria reviewed web research starter kit.\n\nObjective: inspect a target page, capture cited evidence, and summarize the decision impact.\nAgent posture: careful researcher with read-only browser behavior.\nSource/evidence plan: target URL, visible claims, relevant links, screenshot or selector evidence if needed.\nReview gate: ask before forms, account changes, downloads, purchases, posts, or destructive actions.\nReusable output: evidence notes suitable for Share Pack and Reuse Gallery.",
    },
    {
      id: "data-insight",
      type: "数据",
      title: "本地数据洞察简报",
      route: "data",
      agent: "analyst",
      evidence: `${sourceCount} 个已登记来源`,
      reuse: "记忆事实 + 图表简报",
      safety: "不推断缺失字段；标记来源限制和不确定性。",
      objective: "把本地表格、导出或指标集转成可审查发现。",
      prompt: "Launch the Astria local data insight starter kit.\n\nObjective: profile a source, answer one analysis question, and produce ranked findings.\nAgent posture: analyst who separates observed evidence from hypotheses.\nSource/evidence plan: source descriptor, key fields, freshness, missing data, and comparison dimensions.\nReview gate: list source limits before conclusions and ask for missing fields instead of inventing them.\nReusable output: memory candidates, chart brief, and prompt pattern for future data reviews.",
    },
    {
      id: "agent-build",
      type: "Agent",
      title: "聚焦 Agent 配置",
      route: "agents",
      agent: "architect",
      evidence: `${agentCount} 个当前 Agent`,
      reuse: "Agent 命令 + Prompt 星体",
      safety: "工具权限保持明确，避免宽泛自动批准默认值。",
      objective: "为可重复任务设计命名 Agent 配置和命令集。",
      prompt: "Launch the Astria focused agent profile starter kit.\n\nObjective: define a named agent profile for one repeatable workflow.\nAgent posture: systems designer who keeps permissions narrow.\nSource/evidence plan: task type, required memory, allowed tools, denied tools, model posture, and test prompt.\nReview gate: explain why each permission is needed and avoid broad auto-approval.\nReusable output: agent profile, saved command, and launch prompt.",
    },
    {
      id: "share-handoff",
      type: "交接",
      title: "本地交接包",
      route: "share",
      agent: "reviewer",
      evidence: `${state.runs.length} 条运行 + ${reuseCount} 个星体`,
      reuse: "可复制交接包",
      safety: "仅本地交接；脱敏密钥和私有路径。",
      objective: "把有用结果打包成供未来会话或队友继续的已审查交接。",
      prompt: "Launch the Astria local handoff package starter kit.\n\nObjective: package the useful result of current work into a local, copyable handoff.\nAgent posture: reviewer who checks evidence, privacy, and next-action clarity.\nSource/evidence plan: latest run, reusable prompts, memory, sources, and unresolved risks.\nReview gate: redact secrets/private data and require approval before publishing or sending externally.\nReusable output: Share Pack sections with evidence, boundaries, acceptance checklist, and next steps.",
    },
    {
      id: "memory-curation",
      type: "记忆",
      title: "长期记忆整理",
      route: "memory",
      agent: "curator",
      evidence: `${memoryCount} 条记忆`,
      reuse: "已审查记忆候选",
      safety: "只保存长期事实；包含来源和鲜度说明。",
      objective: "从近期工作中提取长期事实、风险、偏好和决策。",
      prompt: "Launch the Astria durable memory curation starter kit.\n\nObjective: identify what should become durable memory from recent Astria work.\nAgent posture: curator who rejects vague or stale notes.\nSource/evidence plan: recent runs, sources, decisions, user preferences, and known risks.\nReview gate: include source, freshness, and why each item should be saved or rejected.\nReusable output: memory candidates and a short taxonomy update if needed.",
    },
    {
      id: "reuse-polish",
      type: "复用",
      title: "可复用工作流打磨",
      route: "reuse",
      agent: "operator",
      evidence: `${reuseCount} 个可复用星体`,
      reuse: "可启动 Prompt 星体",
      safety: "优先沉淀一个实用可复用模式，而不是宽泛抽象。",
      objective: "把有用工作流转成清晰可复用 Prompt 和发起路线。",
      prompt: "Launch the Astria reusable workflow polish starter kit.\n\nObjective: convert one successful workflow into a starter-ready reusable asset.\nAgent posture: operator who favors clear launch steps over abstraction.\nSource/evidence plan: prompt shape, agent fit, source requirements, expected output, and validation command.\nReview gate: prove the workflow is reusable and state where it should not be used.\nReusable output: Reuse Gallery starter, validation checklist, and suggested follow-up route.",
    },
  ];
}

function renderStarterKitLauncher() {
  const kits = starterKits();
  setText("nav-starter-count", kits.length);
  setText("manage-starter-count", `${kits.length} 个套件`);
  setText("starter-summary", `${kits.length} 个 Astria 启动套件，覆盖浏览器、数据、Agent、交接、记忆和复用工作流。`);
  const list = $("starter-kit-grid");
  if (!list) return;
  if (!state.selectedStarterKit || !kits.some((kit) => kit.id === state.selectedStarterKit)) {
    state.selectedStarterKit = kits[0]?.id || "";
  }
  list.innerHTML = kits.map((kit) => `<article class="starter-kit-card ${kit.id === state.selectedStarterKit ? "active" : ""}" data-starter-kit="${escapeHTML(kit.id)}">
    <div class="row-item-title"><span>${escapeHTML(kit.type)}</span><span class="tag">${escapeHTML(kit.agent)}</span></div>
    <strong>${escapeHTML(kit.title)}</strong>
    <div class="starter-kit-gridline">
      <span>路径</span><strong>${escapeHTML(panelName(kit.route))}</strong>
      <span>证据</span><strong>${escapeHTML(kit.evidence)}</strong>
      <span>可复用产物</span><strong>${escapeHTML(kit.reuse)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-starter-select="${escapeHTML(kit.id)}">套件简报</button>
      <button type="button" data-starter-draft="${escapeHTML(kit.id)}">起草套件</button>
      <button type="button" data-panel="${escapeHTML(kit.route)}">打开路径</button>
    </div>
  </article>`).join("");
  renderStarterKitDetail(kits.find((kit) => kit.id === state.selectedStarterKit) || kits[0]);
}

function renderStarterKitDetail(kit) {
  const target = $("starter-kit-detail");
  if (!target) return;
  if (!kit) {
    target.innerHTML = `<div class="empty-state">选择一个启动套件。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(kit.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(kit.type)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(kit.route))}</strong>
        <span>Agent 姿态</span><strong>${escapeHTML(kit.agent)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>目标</h3>
      <p>${escapeHTML(kit.objective)}</p>
      <h3>证据</h3>
      <p>${escapeHTML(kit.evidence)}</p>
      <h3>安全边界</h3>
      <p>${escapeHTML(kit.safety)}</p>
      <h3>可复用产物</h3>
      <p>${escapeHTML(kit.reuse)}</p>
      <div class="run-detail-actions">
        <button type="button" data-starter-draft="${escapeHTML(kit.id)}">起草套件</button>
        <button type="button" data-panel="${escapeHTML(kit.route)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function starterKitByID(id) {
  return starterKits().find((kit) => kit.id === id) || null;
}

function draftStarterKitToChat(id) {
  const kit = starterKitByID(id);
  if (!kit) return;
  $("chat-input").value = `${kit.prompt}\n\nReturn a starter-kit launch brief with objective, agent posture, evidence plan, review gate, reusable output, and the next Astria panel to open.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("启动套件已写入对话。");
}

function sharePackContext() {
  const name = ($("share-pack-name")?.value || state.sharePackName || "").trim();
  const audience = ($("share-pack-audience")?.value || state.sharePackAudience || "").trim();
  const intent = ($("share-pack-intent")?.value || state.sharePackIntent || "").trim();
  const latestRun = state.runs[0];
  const defaultName = latestRun?.prompt ? `交接：${String(latestRun.prompt).slice(0, 72)}` : "Astria 本地交接包";
  return {
    name: name || defaultName,
    audience: audience || "未来 Astria 会话或本地审查者",
    intent: intent || "帮助接收者复用有效上下文、验证证据并安全继续。",
    hasName: Boolean(name),
  };
}

function sharePackCards() {
  const ctx = sharePackContext();
  const latestRun = state.runs[0];
  const latestAgent = state.agents[0]?.name || "default";
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const reuseCount = reuseGalleryAssets().length;
  const sourceCount = sourceRegistryRows().length;
  const dataCount = dataInsightCards().length;
  const artifacts = [
    `${state.runs.length} runs`,
    `${state.agents.length} agents`,
    `${memoryCount} memory entries`,
    `${reuseCount} reuse assets`,
    `${sourceCount} sources`,
  ].join(", ");
  return [
    {
      id: "brief",
      type: "任务简报",
      title: "交接总览简报",
      panel: "runs",
      evidence: latestRun?.id || "尚无最近运行",
      readiness: latestRun ? "就绪" : "待建立",
      boundary: "仅生成本地可复制简报；不暗示云共享或外部权限。",
      action: "起草总览、范围和下一项决策。",
      prompt: `Build a local Astria share pack mission brief.\n\nPackage: ${ctx.name}\nAudience: ${ctx.audience}\nIntent: ${ctx.intent}\nIncluded artifacts: ${artifacts}\nLatest run: ${latestRun?.prompt || "none"}\n\nCreate a concise handoff with objective, what is known, what remains uncertain, who should review it, and the safest next action. Keep it local and copyable; do not claim cloud sharing, account access, or remote permissions.`,
    },
    {
      id: "evidence",
      type: "证据",
      title: "证据包清单",
      panel: "compare",
      evidence: `${sourceCount} 个来源 + ${state.runs.length} 条运行`,
      readiness: sourceCount || state.runs.length ? "待复查" : "缺少证据",
      boundary: "包含来源鲜度和缺失证据；排除密钥和私有数据。",
      action: "起草证据表和验证清单。",
      prompt: `Build a local Astria evidence bundle checklist.\n\nPackage: ${ctx.name}\nAudience: ${ctx.audience}\nIntent: ${ctx.intent}\nIncluded artifacts: ${artifacts}\n\nList the evidence that should be included, where each item came from, freshness or reliability concerns, missing proof, and verification steps. Redact secrets and private data before anything is copied outside the local workspace.`,
    },
    {
      id: "prompt",
      type: "Prompt",
      title: "可复用 Prompt 起点",
      panel: "reuse",
      evidence: `${reuseCount} 个可复用星体`,
      readiness: reuseCount ? "可复用" : "草稿",
      boundary: "打包 Prompt 模式，而不是隐藏状态或凭证。",
      action: "起草可供下一条运行复用的起点 Prompt。",
      prompt: `Build a reusable Astria prompt starter for a share pack.\n\nPackage: ${ctx.name}\nAudience: ${ctx.audience}\nIntent: ${ctx.intent}\nLead agent: ${latestAgent}\nReuse assets: ${reuseCount}\n\nExtract the reusable prompt pattern, required context, expected output, review guardrails, and validation commands. Do not include secrets, private paths unless necessary, or assumptions that only this session knows.`,
    },
    {
      id: "knowledge",
      type: "知识",
      title: "记忆交接记录",
      panel: "memory",
      evidence: `${memoryCount} 条记忆`,
      readiness: memoryCount ? "待整理" : "待建立",
      boundary: "只保存带来源和鲜度说明的长期事实。",
      action: "起草记忆候选和过期说明。",
      prompt: `Build Astria memory handoff notes for a local share pack.\n\nPackage: ${ctx.name}\nAudience: ${ctx.audience}\nIntent: ${ctx.intent}\nMemory entries: ${memoryCount}\nIncluded artifacts: ${artifacts}\n\nIdentify durable facts, decisions, preferences, risks, and stale items. Write memory candidates with evidence, freshness, and why each should or should not be saved.`,
    },
    {
      id: "review",
      type: "审核",
      title: "审查者验收清单",
      panel: dataCount ? "data" : "council",
      evidence: dataCount ? `${dataCount} 个数据透镜` : `${state.councilRuns.length} 条议会运行`,
      readiness: "闸门",
      boundary: "发布、定时或外发交接包前必须人工复查。",
      action: "起草验收标准和拒绝触发条件。",
      prompt: `Build a reviewer acceptance checklist for a local Astria share pack.\n\nPackage: ${ctx.name}\nAudience: ${ctx.audience}\nIntent: ${ctx.intent}\nIncluded artifacts: ${artifacts}\n\nDefine acceptance criteria, rejection triggers, required evidence, privacy checks, and follow-up routes. Require explicit approval before publishing, scheduling, or sending this pack outside the local workspace.`,
    },
  ];
}

function renderSharePackBuilder() {
  const cards = sharePackCards();
  setText("nav-share-count", cards.length);
  setText("manage-share-count", `${cards.length} 个交接包`);
  setText("share-summary", `${cards.length} 张本地交接卡，覆盖任务简报、证据、Prompt、记忆和审核门槛。`);
  const list = $("share-pack-cards");
  if (!list) return;
  if (!state.selectedSharePack || !cards.some((card) => card.id === state.selectedSharePack)) {
    state.selectedSharePack = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="share-pack-card ${card.id === state.selectedSharePack ? "active" : ""}" data-share-pack="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.readiness)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="share-pack-grid">
      <span>证据</span><strong>${escapeHTML(card.evidence)}</strong>
      <span>边界</span><strong>${escapeHTML(card.boundary)}</strong>
      <span>下一步</span><strong>${escapeHTML(card.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-share-select="${escapeHTML(card.id)}">交接简报</button>
      <button type="button" data-share-draft="${escapeHTML(card.id)}">起草交接</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderSharePackDetail(cards.find((card) => card.id === state.selectedSharePack) || cards[0]);
}

function renderSharePackDetail(card) {
  const target = $("share-pack-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一张交接卡。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>就绪度</span><strong>${escapeHTML(card.readiness)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>边界</h3>
      <p>${escapeHTML(card.boundary)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(card.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-share-draft="${escapeHTML(card.id)}">起草交接</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
      </div>
    </section>
  </div>`;
}

function sharePackByID(id) {
  return sharePackCards().find((card) => card.id === id) || null;
}

function draftSharePackToChat(id) {
  const card = sharePackByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a local share pack section with package summary, included artifacts, evidence, boundaries, review steps, and reusable next actions.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("交接包已写入对话。");
}

function workspaceSnapshotCards() {
  const latestRun = state.runs[0];
  const latestSession = state.sessions[0];
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const sourceCount = sourceRegistryRows().length;
  const resultCount = resultArchiveEntries().length;
  const playbookCount = playbookLibraryCards().length;
  const reuseCount = reuseGalleryAssets().length;
  const shareCount = sharePackCards().length;
  const agentCount = state.agents.length;
  const scheduleCount = state.schedules.length;
  const riskCount = knowledgeReconciliationItems().length;
  const runLabel = latestRun?.prompt || latestRun?.id || "尚无最近运行";
  const sessionLabel = latestSession?.title || latestSession?.id || "尚无最近会话";
  const localInventory = `${state.sessions.length} 个会话，${state.runs.length} 条运行，${memoryCount} 条记忆，${sourceCount} 个来源，${resultCount} 个产物`;
  return [
    {
      id: "resume",
      type: "续接",
      title: "会话续接快照",
      panel: latestSession ? "chat" : "runs",
      included: `${sessionLabel}; ${runLabel}`,
      missing: latestSession ? "续接前确认未解决下一步和当前分支。" : "先创建或选择会话，再把它视为可续接。",
      reviewGate: "验证当前目标、最新用户请求、活跃文件和未完成检查。",
      privacy: "除非接收者确实需要，否则本地路径和会话 ID 保持内部可见。",
      route: "打开对话，从选中的本地上下文继续。",
      prompt: `Build an Astria session resume snapshot.\n\nLocal inventory: ${localInventory}\nLatest session: ${sessionLabel}\nLatest run: ${runLabel}\n\nReturn a resume pack with current objective, relevant context, completed work, open risks, files to inspect first, validation state, and the next safe action. Mark missing context instead of guessing.`,
    },
    {
      id: "evidence",
      type: "证据",
      title: "运行证据快照",
      panel: "runs",
      included: `${state.runs.length} 条运行；${sourceCount} 个已登记来源；${resultCount} 份产物简报`,
      missing: state.runs.length ? "识别哪些输出是最终版、草稿或阻塞态。" : "运行历史为空；证据快照应从清单开始。",
      reviewGate: "区分可观察工具输出、模型综合和未支撑假设。",
      privacy: "脱敏包含密钥、私有路径或账号数据的命令输出。",
      route: "打开运行观测台检查执行详情和可复制摘要。",
      prompt: `Build an Astria run evidence snapshot.\n\nRuns: ${state.runs.length}\nSources: ${sourceCount}\nResult archive entries: ${resultCount}\nLatest run: ${runLabel}\n\nReturn evidence grouped by run, source, result, confidence, freshness, and unresolved gaps. Flag anything that needs citation grounding or reviewer approval.`,
    },
    {
      id: "memory-source",
      type: "知识",
      title: "记忆与来源快照",
      panel: sourceCount ? "sources" : "memory",
      included: `${memoryCount} 条记忆；${sourceCount} 条来源路线；${riskCount} 个知识校验风险`,
      missing: riskCount ? "复用前解决过期、冲突、薄弱或敏感知识。" : "为长期记忆候选补充来源和鲜度说明。",
      reviewGate: "每条长期事实都需要来源、鲜度、分类和拒绝标准。",
      privacy: "没有明确需要时，不快照敏感记录、密钥或私有事实。",
      route: "打开来源登记或记忆星图整理长期上下文。",
      prompt: `Build an Astria memory and source snapshot.\n\nMemory entries: ${memoryCount}\nSource lanes: ${sourceCount}\nReconciliation risks: ${riskCount}\n\nReturn durable facts, source coverage, stale or conflicting items, privacy exclusions, and memory candidates that are safe to reuse.`,
    },
    {
      id: "result-archive",
      type: "产物",
      title: "产物归档快照",
      panel: "results",
      included: `${resultCount} 个归档产物；${shareCount} 张交接卡；${reuseCount} 个可复用星体`,
      missing: resultCount ? "确认哪些归档产物拥有证据和验收检查。" : "尚无归档产物；从最近完成运行开始建立。",
      reviewGate: "每个保存结果都需要产出、来源证据、鲜度、复用路线和开放风险。",
      privacy: "快照产物时不包含隐藏推理链、凭证或私有工作区数据。",
      route: "打开产物星库检查已保存报告和后续 Prompt。",
      prompt: `Build an Astria result archive snapshot.\n\nArchived results: ${resultCount}\nShare pack cards: ${shareCount}\nReusable assets: ${reuseCount}\n\nReturn a local result package with outcome summaries, evidence links, freshness, reusable prompt paths, acceptance checks, and unresolved risks.`,
    },
    {
      id: "playbook-reuse",
      type: "复用",
      title: "手册与复用快照",
      panel: "playbooks",
      included: `${playbookCount} 本手册；${reuseCount} 个可复用星体；${agentCount} 个 Agent`,
      missing: playbookCount ? "确认手册仍匹配当前工具和安全边界。" : "复用前先把成功工作流沉淀为已审查手册。",
      reviewGate: "可复用工作流需要触发条件、步骤、证据门、安全边界和验证。",
      privacy: "保存可复用模式，而不是敏感项目状态或凭证。",
      route: "打开实践手册，发起已审查的本地最佳实践路线。",
      prompt: `Build an Astria playbook and reuse snapshot.\n\nPlaybooks: ${playbookCount}\nReusable assets: ${reuseCount}\nAgents: ${agentCount}\n\nReturn repeatable workflows, agent/profile dependencies, prompts to reuse, safety boundaries, validation commands, and stale-pattern risks.`,
    },
    {
      id: "delivery-schedule",
      type: "交付",
      title: "交付与定时快照",
      panel: scheduleCount ? "delivery" : "schedules",
      included: `${scheduleCount} 条定时任务；${deliveryLanes().length} 条交付链路；${state.inboxItems.length} 条收件箱事项`,
      missing: scheduleCount ? "确认每个定时输出的目标、审批门和回滚路径。" : "尚无定时任务；将交付快照保持为已审查计划。",
      reviewGate: "出站、定时或渠道工作都需要明确审批和验证。",
      privacy: "此快照不暗示外发、发布、定时或远端状态变更。",
      route: "打开主动交付或定时任务，复查频率和审批边界。",
      prompt: `Build an Astria delivery and schedule snapshot.\n\nSchedules: ${scheduleCount}\nDelivery lanes: ${deliveryLanes().length}\nInbox items: ${state.inboxItems.length}\n\nReturn destination candidates, schedule cadence, approval gates, verification steps, rollback paths, and what must stay local.`,
    },
    {
      id: "privacy",
      type: "隐私",
      title: "脱敏与交接边界",
      panel: riskCount ? "reconcile" : "share",
      included: `${riskCount} 个知识风险；${sourceCount} 个来源；${shareCount} 张交接卡`,
      missing: "复制任何内容前，复查本地路径、API key、账号数据、私有文件和未支撑假设。",
      reviewGate: "密钥、私有数据和薄弱结论移除前，任何内容都不离开本地工作区。",
      privacy: "默认本地优先。脱敏凭证、私有路径、用户数据和隐藏状态。",
      route: "打开知识校验或交接包，完成边界检查。",
      prompt: `Build an Astria redaction and handoff-boundary snapshot.\n\nKnowledge risks: ${riskCount}\nSources: ${sourceCount}\nShare pack cards: ${shareCount}\n\nReturn what can be copied, what must be redacted, what requires approval, weak or unsupported claims, and the local-only boundary for this handoff.`,
    },
  ];
}

function renderWorkspaceSnapshotPlanner() {
  const cards = workspaceSnapshotCards();
  setText("nav-snapshot-count", cards.length);
  setText("manage-snapshot-count", `${cards.length} 个快照`);
  setText("snapshot-summary", `${cards.length} 个本地快照包，覆盖续接、证据、记忆、产物、手册、交付和隐私复查。`);
  const list = $("workspace-snapshot-grid");
  if (!list) return;
  if (!state.selectedWorkspaceSnapshot || !cards.some((card) => card.id === state.selectedWorkspaceSnapshot)) {
    state.selectedWorkspaceSnapshot = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="workspace-snapshot-card ${card.id === state.selectedWorkspaceSnapshot ? "active" : ""}" data-lane="S" data-workspace-snapshot="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.panel)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="workspace-snapshot-gridline">
      <span>已包含</span><strong>${escapeHTML(card.included)}</strong>
      <span>缺口</span><strong>${escapeHTML(card.missing)}</strong>
      <span>复查门槛</span><strong>${escapeHTML(card.reviewGate)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-snapshot-select="${escapeHTML(card.id)}">快照简报</button>
      <button type="button" data-snapshot-draft="${escapeHTML(card.id)}">起草快照</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
    </div>
  </article>`).join("");
  renderWorkspaceSnapshotDetail(cards.find((card) => card.id === state.selectedWorkspaceSnapshot) || cards[0]);
}

function renderWorkspaceSnapshotDetail(card) {
  const target = $("workspace-snapshot-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一个工作区快照。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
        <span>下一步</span><strong>${escapeHTML(card.route)}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>已包含上下文</h3>
      <p>${escapeHTML(card.included)}</p>
      <h3>缺失部分</h3>
      <p>${escapeHTML(card.missing)}</p>
      <h3>复查门槛</h3>
      <p>${escapeHTML(card.reviewGate)}</p>
      <h3>隐私边界</h3>
      <p>${escapeHTML(card.privacy)}</p>
      <div class="run-detail-actions">
        <button type="button" data-snapshot-draft="${escapeHTML(card.id)}">起草快照</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function workspaceSnapshotByID(id) {
  return workspaceSnapshotCards().find((card) => card.id === id) || null;
}

function draftWorkspaceSnapshotToChat(id) {
  const card = workspaceSnapshotByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a local Workspace Snapshot pack with included context, missing pieces, review gate, privacy/redaction boundary, route, and next action. Do not claim that a file was exported unless explicitly asked to create one.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("工作区快照已写入对话。");
}

function browserMissionContext() {
  const url = ($("browser-target-url")?.value || state.browserTargetURL || "").trim();
  const goal = ($("browser-mission-goal")?.value || state.browserMissionGoal || "").trim();
  return {
    url: url || "目标页面",
    goal: goal || "检查页面，并捕获下一步 Astria 决策需要的证据。",
    hasURL: Boolean(url),
  };
}

function browserMissionCards() {
  const ctx = browserMissionContext();
  const intakeLabel = state.intakeResult ? `文件上下文已就绪：${state.intakeResult.path || state.intakeResult.mode || "文件星舱"}` : "尚未附加文件上下文";
  const inboxPending = state.inboxItems.filter((item) => String(item.status || "pending").toLowerCase() === "pending").length;
  const readyDiagnostics = ["ok", "ready", "healthy"].includes(String(state.diagnostics?.status || "").toLowerCase());
  return [
    {
      id: "inspect",
      type: "检查",
      title: "页面检查",
      panel: "chat",
      evidence: ctx.hasURL ? ctx.url : "需要 URL",
      readiness: readyDiagnostics ? "就绪" : "待复查",
      risk: "只做只读导航和页面总结；不要点击会改变账号状态的控件。",
      action: "起草一条带来源引用的检查运行。",
      prompt: `Plan a reviewed browser inspection mission.\n\nTarget: ${ctx.url}\nGoal: ${ctx.goal}\n\nUse browser navigation only as needed. Summarize visible page structure, key claims, relevant links, and evidence to cite. Do not submit forms, change account settings, purchase, delete, or post anything without explicit approval.`,
    },
    {
      id: "screenshot",
      type: "截图",
      title: "视觉证据捕获",
      panel: "diagnostics",
      evidence: "浏览器 + 截图",
      readiness: readyDiagnostics ? "就绪" : "检查运行时",
      risk: "捕获证据时避免暴露密钥或私有账号数据。",
      action: "起草截图清单和证据摘要。",
      prompt: `Plan a browser screenshot evidence mission.\n\nTarget: ${ctx.url}\nGoal: ${ctx.goal}\n\nOpen the target, capture the necessary visual evidence, describe what the screenshot proves, and call out any private or sensitive information that should be cropped or avoided. Ask before interacting with authenticated or destructive UI.`,
    },
    {
      id: "extract",
      type: "抽取",
      title: "结构化页面抽取",
      panel: state.intakeResult ? "intake" : "chat",
      evidence: intakeLabel,
      readiness: ctx.hasURL ? "已定向" : "需要目标",
      risk: "只抽取公开或操作者批准的内容；标注不确定性和缺失字段。",
      action: "读取前先起草抽取 schema。",
      prompt: `Plan a structured browser extraction mission.\n\nTarget: ${ctx.url}\nGoal: ${ctx.goal}\nLocal context: ${intakeLabel}\n\nDefine the fields to extract, inspect the page, return structured findings with citations or selectors where possible, and identify anything that needs manual verification.`,
    },
    {
      id: "form-check",
      type: "表单检查",
      title: "表单与流程复查",
      panel: "permissions",
      evidence: "需要审批",
      readiness: "受保护",
      risk: "没有明确审批，不提交表单、付款、账号变更或消息。",
      action: "起草安全的表单 dry-run 复查。",
      prompt: `Plan a safe browser form-check mission.\n\nTarget: ${ctx.url}\nGoal: ${ctx.goal}\n\nInspect form fields, validation states, required data, and risks. You may type only harmless placeholder data if needed for local validation, but do not submit or trigger remote state changes without explicit approval.`,
    },
    {
      id: "monitor",
      type: "监控",
      title: "变化监控简报",
      panel: inboxPending ? "inbox" : "schedules",
      evidence: inboxPending ? `${inboxPending} 条待处理进入项` : `${state.schedules.length} 条定时任务`,
      readiness: state.schedules.length ? "可定时" : "手动",
      risk: "监控进入定时前，需要先定义频率、阈值和通知路线。",
      action: "基于当前目标起草监控计划。",
      prompt: `Plan a browser change-monitoring mission.\n\nTarget: ${ctx.url}\nGoal: ${ctx.goal}\nSchedules: ${state.schedules.length}\nPending inbox items: ${inboxPending}\n\nDefine what should be monitored, the cadence, change threshold, evidence to capture, and how Astria should report changes before any schedule is created.`,
    },
  ];
}

function renderBrowserMissionPlanner() {
  const cards = browserMissionCards();
  setText("nav-browser-count", cards.length);
  setText("manage-browser-count", `${cards.length} 个计划`);
  setText("browser-summary", `${cards.length} 个浏览器任务计划，覆盖检查、截图、抽取、表单检查和监控。`);
  const list = $("browser-mission-cards");
  if (!list) return;
  if (!state.selectedBrowserMission || !cards.some((card) => card.id === state.selectedBrowserMission)) {
    state.selectedBrowserMission = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="browser-mission-card ${card.id === state.selectedBrowserMission ? "active" : ""}" data-browser-mission="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.readiness)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="browser-mission-grid">
      <span>证据</span><strong>${escapeHTML(card.evidence)}</strong>
      <span>风险</span><strong>${escapeHTML(card.risk)}</strong>
      <span>下一步</span><strong>${escapeHTML(card.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-browser-select="${escapeHTML(card.id)}">任务简报</button>
      <button type="button" data-browser-draft="${escapeHTML(card.id)}">起草任务</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderBrowserMissionDetail(cards.find((card) => card.id === state.selectedBrowserMission) || cards[0]);
}

function renderBrowserMissionDetail(card) {
  const target = $("browser-mission-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一个浏览器任务。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>就绪度</span><strong>${escapeHTML(card.readiness)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>风险</h3>
      <p>${escapeHTML(card.risk)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(card.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-browser-draft="${escapeHTML(card.id)}">起草任务</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
      </div>
    </section>
  </div>`;
}

function browserMissionByID(id) {
  return browserMissionCards().find((card) => card.id === id) || null;
}

function draftBrowserMissionToChat(id) {
  const card = browserMissionByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a browser mission starter with objective, target, evidence plan, safety boundary, and validation.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("浏览器任务已写入对话。");
}

function dataInsightContext() {
  const source = ($("data-source-descriptor")?.value || state.dataSourceDescriptor || "").trim();
  const question = ($("data-analysis-question")?.value || state.dataAnalysisQuestion || "").trim();
  const output = ($("data-output-format")?.value || state.dataOutputFormat || "").trim();
  const intakeLabel = state.intakeResult ? `当前文件星舱：${state.intakeResult.path || state.intakeResult.mode || "就绪"}` : "尚未附加文件星舱结果";
  return {
    source: source || intakeLabel,
    question: question || "识别这份数据能支持的决策，以及仍需复查的不确定性。",
    output: output || "带证据、限制和可复用下一步的排序发现",
    hasSource: Boolean(source) || Boolean(state.intakeResult),
    intakeLabel,
  };
}

function dataInsightCards() {
  const ctx = dataInsightContext();
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const sourceCount = sourceRegistryRows().length;
  const reuseCount = reuseGalleryAssets().length;
  return [
    {
      id: "profile",
      type: "画像",
      title: "来源画像与 schema 检查",
      panel: ctx.hasSource ? "chat" : "intake",
      evidence: ctx.source,
      readiness: ctx.hasSource ? "就绪" : "需要来源",
      guardrail: "不要推断缺失列或隐藏行；结论前先列出来源限制。",
      action: "起草 schema、质量和覆盖度复查。",
      prompt: `Plan a reviewed Astria data profiling mission.\n\nSource: ${ctx.source}\nQuestion: ${ctx.question}\nExpected output: ${ctx.output}\n\nInspect available columns, sample shape, freshness, missing values, duplicate risks, and source limits. Do not invent unavailable data. Return a compact profile, quality risks, and what analysis is safe to run next.`,
    },
    {
      id: "trend",
      type: "趋势",
      title: "趋势与分群解读",
      panel: "compare",
      evidence: `${sourceCount} 个已登记来源`,
      readiness: ctx.hasSource ? "已定向" : "待复查",
      guardrail: "区分观察到的变化和解释；相关性只能标为假设。",
      action: "按时间、分群或来源起草趋势比较。",
      prompt: `Plan an Astria trend analysis mission.\n\nSource: ${ctx.source}\nQuestion: ${ctx.question}\nExpected output: ${ctx.output}\nRegistered sources: ${sourceCount}\n\nIdentify time fields or comparable segments, compute or request only reviewable summaries, compare alternative explanations, and return findings with caveats instead of unsupported causal claims.`,
    },
    {
      id: "anomaly",
      type: "异常",
      title: "离群点与风险复查",
      panel: "runs",
      evidence: `${state.runs.length} 条最近运行`,
      readiness: ctx.hasSource ? "受保护" : "需要来源",
      guardrail: "在核对来源质量和上下文前，异常只能作为候选。",
      action: "起草带验证步骤的离群点复查。",
      prompt: `Plan an Astria anomaly review mission.\n\nSource: ${ctx.source}\nQuestion: ${ctx.question}\nExpected output: ${ctx.output}\nRecent runs: ${state.runs.length}\n\nDefine anomaly criteria, inspect source quality first, list candidate outliers, explain why each matters, and propose validation before any decision is made.`,
    },
    {
      id: "visual",
      type: "图表简报",
      title: "可视摘要计划",
      panel: "reuse",
      evidence: "图表就绪简报",
      readiness: "草稿",
      guardrail: "选择匹配字段的可视化；避免用装饰性图表遮蔽不确定性。",
      action: "起草图表简报和叙事结构。",
      prompt: `Plan an Astria visual data summary mission.\n\nSource: ${ctx.source}\nQuestion: ${ctx.question}\nExpected output: ${ctx.output}\n\nRecommend the smallest useful chart set, define axes and grouping, state what each visual should prove, and include the text summary that should accompany the charts. If fields are missing, ask for them instead of fabricating visuals.`,
    },
    {
      id: "knowledge",
      type: "知识",
      title: "可复用洞察捕获",
      panel: "memory",
      evidence: `${memoryCount} 条记忆`,
      readiness: "可保存",
      guardrail: "只保存长期且有来源支撑的发现；区分一次性观察和可复用事实。",
      action: "从分析中起草记忆与复用候选。",
      prompt: `Plan an Astria reusable data insight capture mission.\n\nSource: ${ctx.source}\nQuestion: ${ctx.question}\nExpected output: ${ctx.output}\nMemory entries: ${memoryCount}\nReuse assets: ${reuseCount}\n\nExtract only durable findings that are backed by the source, write memory candidates with evidence and expiry/freshness notes, and propose which prompts or analysis patterns should become reusable starters.`,
    },
  ];
}

function renderDataInsightPlanner() {
  const cards = dataInsightCards();
  setText("nav-data-count", cards.length);
  setText("manage-data-count", `${cards.length} 个透镜`);
  setText("data-summary", `${cards.length} 个数据洞察透镜，覆盖画像、趋势、异常、可视摘要和可复用知识。`);
  const list = $("data-insight-cards");
  if (!list) return;
  if (!state.selectedDataInsight || !cards.some((card) => card.id === state.selectedDataInsight)) {
    state.selectedDataInsight = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="data-insight-card ${card.id === state.selectedDataInsight ? "active" : ""}" data-data-insight="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.readiness)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="data-insight-grid">
      <span>证据</span><strong>${escapeHTML(card.evidence)}</strong>
      <span>护栏</span><strong>${escapeHTML(card.guardrail)}</strong>
      <span>下一步</span><strong>${escapeHTML(card.action)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-data-select="${escapeHTML(card.id)}">洞察简报</button>
      <button type="button" data-data-draft="${escapeHTML(card.id)}">起草分析</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderDataInsightDetail(cards.find((card) => card.id === state.selectedDataInsight) || cards[0]);
}

function renderDataInsightDetail(card) {
  const target = $("data-insight-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一个数据洞察任务。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>就绪度</span><strong>${escapeHTML(card.readiness)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>护栏</h3>
      <p>${escapeHTML(card.guardrail)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(card.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-data-draft="${escapeHTML(card.id)}">起草分析</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
      </div>
    </section>
  </div>`;
}

function dataInsightByID(id) {
  return dataInsightCards().find((card) => card.id === id) || null;
}

function draftDataInsightToChat(id) {
  const card = dataInsightByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a data insight mission starter with source assumptions, analysis steps, evidence requirements, review guardrails, reusable findings, and validation.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("数据洞察任务已写入对话。");
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
      source: "定时任务",
      panel: "schedules",
      title: "定时工作",
      metric: `${enabledSchedules.length}/${state.schedules.length} 启用`,
      evidence: [
        enabledSchedules[0] ? `下一条启用 Prompt：${enabledSchedules[0].prompt || "未命名定时任务"}` : "尚未配置启用中的定时任务",
        state.schedules.length ? `${state.schedules.length} 条已配置定时任务` : "期待主动工作前，先创建 cron 计划",
        enabledSchedules[0]?.cron ? `Cron：${enabledSchedules[0].cron}` : "Cron 频率未设置",
      ],
      risk: enabledSchedules.length ? "定时 Prompt 仍需要明确交付目标或复查预期。" : "没有定时任务时，不会发生主动工作。",
      action: enabledSchedules.length ? "复查启用频率和交付目标。" : "起草第一条定时交付计划。",
      prompt: `Plan proactive Astria delivery from schedules.\n\nActive schedules: ${enabledSchedules.length}\nConfigured schedules: ${state.schedules.length}\nFirst prompt: ${enabledSchedules[0]?.prompt || "none"}\n\nDefine cadence, expected output, destination channel, and validation.`,
    },
    {
      id: "delivery-runs",
      source: "运行",
      panel: "runs",
      title: "最近出站运行",
      metric: `${scheduledRuns.length} 条定时运行`,
      evidence: [
        scheduledRuns[0] ? `最近定时运行：${uiTerm(scheduledRuns[0].status || "unknown")}` : "尚未捕获定时运行历史",
        failedRuns.length ? `${failedRuns.length} 条失败运行需要恢复` : "当前列表没有失败运行",
        state.runs[0] ? `最近运行：${state.runs[0].prompt || state.runs[0].id}` : "运行历史为空",
      ],
      risk: failedRuns.length ? "失败未分诊前，交付可信度偏低。" : "外部交付前仍需复查最近运行。",
      action: failedRuns.length ? "为失败的出站工作起草恢复记录。" : "从最近运行起草出站摘要。",
      prompt: `Review proactive delivery run history.\n\nScheduled runs: ${scheduledRuns.length}\nFailed runs: ${failedRuns.length}\nLatest run: ${state.runs[0]?.prompt || "none"}\n\nDecide what is safe to deliver and what needs retry or review.`,
    },
    {
      id: "channel-readiness",
      source: "渠道",
      panel: "inbox",
      title: "渠道就绪度",
      metric: `${providers.length} 个 provider`,
      evidence: [
        providers.length ? `Provider：${providers.map((provider) => provider.name || provider.id).filter(Boolean).join(", ")}` : "尚未列出渠道 provider",
        pendingInbox.length ? `${pendingInbox.length} 条进入项等待处理` : "没有待处理进入项",
        "出站交付应遵循已审查的进入渠道策略",
      ],
      risk: providers.length ? "渠道状态可见，但出站交付仍需要明确审批。" : "没有可见渠道时，主动输出保持本地。",
      action: providers.length ? "起草面向具体渠道的交付文案。" : "起草渠道设置要求。",
      prompt: `Prepare proactive Astria channel delivery.\n\nProviders: ${providers.map((provider) => provider.name || provider.id).filter(Boolean).join(", ") || "none"}\nPending inbox items: ${pendingInbox.length}\n\nWrite the delivery target, approval gate, message shape, and rollback path.`,
    },
    {
      id: "delivery-recovery",
      source: "就绪度",
      panel: readyDiagnostics ? "delivery" : "diagnostics",
      title: "恢复与护栏",
      metric: readyDiagnostics ? "就绪" : "待复查",
      evidence: [
        `诊断：${uiTerm(state.diagnostics?.status || "unknown")}`,
        state.diagnostics?.summary || "诊断摘要不可用",
        state.permissions?.configured === true ? "权限已配置" : "使用默认权限",
      ],
      risk: readyDiagnostics ? "就绪状态仍需在外部发布前设置审批边界。" : "运行时未就绪可能阻碍可靠的定时交付。",
      action: readyDiagnostics ? "起草审批清单。" : "打开诊断并修复阻塞项。",
      prompt: `Create a proactive delivery recovery checklist.\n\nDiagnostics: ${state.diagnostics?.status || "unknown"}\nSummary: ${state.diagnostics?.summary || ""}\nPermissions configured: ${state.permissions?.configured === true}\n\nList blockers, approval gates, retry rules, and verification.`,
    },
  ];
}

function renderProactiveDeliveryBoard() {
  const lanes = deliveryLanes();
  setText("nav-delivery-count", lanes.length);
  setText("manage-delivery-count", `${lanes.length} 条交付链路`);
  setText("delivery-summary", `${lanes.length} 条主动交付链路，覆盖定时任务、运行、渠道和恢复就绪度。`);
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
      <button type="button" data-delivery-select="${escapeHTML(lane.id)}">交付简报</button>
      <button type="button" data-delivery-draft="${escapeHTML(lane.id)}">起草交付</button>
      <button type="button" data-panel="${escapeHTML(lane.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderDeliveryDetail(lanes.find((lane) => lane.id === state.selectedDeliveryLane) || lanes[0]);
}

function renderDeliveryDetail(lane) {
  const target = $("delivery-detail");
  if (!target) return;
  if (!lane) {
    target.innerHTML = `<div class="empty-state">选择一条交付链路。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(lane.title)}</h3>
      <div class="run-meta-grid">
        <span>来源</span><strong>${escapeHTML(lane.source)}</strong>
        <span>就绪度</span><strong>${escapeHTML(lane.metric)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(lane.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>证据</h3>
      <div class="delivery-evidence detail">
        ${lane.evidence.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>风险</h3>
      <p>${escapeHTML(lane.risk)}</p>
      <h3>下一步</h3>
      <p>${escapeHTML(lane.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-delivery-draft="${escapeHTML(lane.id)}">起草交付</button>
        <button type="button" data-panel="${escapeHTML(lane.panel)}">打开来源</button>
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
  showToast("交付简报已写入对话。");
}

function renderManageCount() {
  const mcpCount = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers.length : 0;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const sourceCount = sourceRegistryRows().length;
  const reconcileCount = knowledgeReconciliationItems().length;
  const citationCount = citationGroundingCards().length;
  const compareCount = comparisonCandidates().length;
  const qualityCount = runQualityCards().length;
  const promptVariantCount = promptLabVariants().length;
  const budgetCount = budgetGuardCards().length;
  const reuseCount = reuseGalleryAssets().length;
  const resultsCount = resultArchiveEntries().length;
  const playbooksCount = playbookLibraryCards().length;
  const starterCount = starterKits().length;
  const shareCount = sharePackCards().length;
  const snapshotCount = workspaceSnapshotCards().length;
  const browserCount = browserMissionCards().length;
  const dataCount = dataInsightCards().length;
  const deliveryCount = deliveryLanes().length;
  setText("manage-intake-count", state.intakeResult ? "Result ready" : "Local paths");
  setText("manage-sources-count", `${sourceCount} source${sourceCount === 1 ? "" : "s"}`);
  setText("nav-sources-count", sourceCount);
  setText("manage-reconcile-count", `${reconcileCount} risk${reconcileCount === 1 ? "" : "s"}`);
  setText("nav-reconcile-count", reconcileCount);
  setText("manage-citation-count", `${citationCount} check${citationCount === 1 ? "" : "s"}`);
  setText("nav-citation-count", citationCount);
  setText("manage-compare-count", `${compareCount} lane${compareCount === 1 ? "" : "s"}`);
  setText("nav-compare-count", compareCount);
  setText("manage-quality-count", `${qualityCount} card${qualityCount === 1 ? "" : "s"}`);
  setText("nav-quality-count", qualityCount);
  setText("manage-promptlab-count", `${promptVariantCount} variant${promptVariantCount === 1 ? "" : "s"}`);
  setText("nav-promptlab-count", promptVariantCount);
  setText("manage-budget-count", `${budgetCount} guard${budgetCount === 1 ? "" : "s"}`);
  setText("nav-budget-count", budgetCount);
  setText("manage-reuse-count", `${reuseCount} asset${reuseCount === 1 ? "" : "s"}`);
  setText("nav-reuse-count", reuseCount);
  setText("manage-results-count", `${resultsCount} result${resultsCount === 1 ? "" : "s"}`);
  setText("nav-results-count", resultsCount);
  setText("manage-playbooks-count", `${playbooksCount} playbook${playbooksCount === 1 ? "" : "s"}`);
  setText("nav-playbooks-count", playbooksCount);
  setText("manage-starter-count", `${starterCount} kit${starterCount === 1 ? "" : "s"}`);
  setText("nav-starter-count", starterCount);
  setText("manage-share-count", `${shareCount} pack${shareCount === 1 ? "" : "s"}`);
  setText("nav-share-count", shareCount);
  setText("manage-snapshot-count", `${snapshotCount} pack${snapshotCount === 1 ? "" : "s"}`);
  setText("nav-snapshot-count", snapshotCount);
  setText("manage-browser-count", `${browserCount} plan${browserCount === 1 ? "" : "s"}`);
  setText("nav-browser-count", browserCount);
  setText("manage-data-count", `${dataCount} lens${dataCount === 1 ? "" : "es"}`);
  setText("nav-data-count", dataCount);
  setText("manage-delivery-count", `${deliveryCount} lane${deliveryCount === 1 ? "" : "s"}`);
  setText("nav-delivery-count", deliveryCount);
  setText("manage-count", state.agents.length + state.skills.length + state.schedules.length + mcpCount + memoryCount + sourceCount + reconcileCount + citationCount + state.councilRuns.length + state.inboxItems.length + compareCount + qualityCount + promptVariantCount + budgetCount + reuseCount + resultsCount + playbooksCount + starterCount + shareCount + snapshotCount + browserCount + dataCount + deliveryCount + 1);
}

function renderMCPStarport() {
  const servers = Array.isArray(state.config?.mcp_servers) ? state.config.mcp_servers : [];
  setText("nav-mcp-count", servers.length);
  setText("manage-mcp-count", `${servers.length} 个 dock`);
  renderManageCount();
  setText("mcp-summary", servers.length ? `${servers.length} 个 MCP 服务器已配置。` : "尚未配置 MCP 服务器。");
  const enabled = servers.filter((server) => !server.disabled).length;
  const overview = $("mcp-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(enabled ? `${enabled} 个启用` : "无 dock")}</strong><span>${escapeHTML(servers.length ? "可在 Astria 中编辑、停用或测试已配置 MCP dock。" : "从 Astria 添加 stdio dock，然后测试连接。")}</span>`;
  }
  renderMCPForm();
  const list = $("mcp-list");
  if (!list) return;
  if (!servers.length) {
    renderEmptyAction(list, "尚未配置 MCP 服务器。添加一个 stdio dock，或让 Astria 建议第一条连接。", [
      { label: "添加 dock", action: "mcp-new", primary: true },
      { label: "询问 Astria", homeAction: "mcp" },
    ]);
    return;
  }
  list.innerHTML = servers.map((server) => {
    const transport = server.type || "stdio";
    const endpoint = transport === "http" ? (server.url || "缺少 URL") : [server.command || "缺少命令"].concat(server.args || []).join(" ");
    const envKeys = Array.isArray(server.env_keys) ? server.env_keys : [];
    return `<article class="row-item mcp-server-card ${server.disabled ? "disabled" : "enabled"}">
      <div class="row-item-title">
        <span>${escapeHTML(server.name || "未命名服务器")}</span>
        <span class="tag">${escapeHTML(server.disabled ? "已停用" : transport)}</span>
      </div>
      <p>${escapeHTML(endpoint)}</p>
      <div class="pill-list">
        <span>${server.keep_alive ? "保持连接" : "按需连接"}</span>
        <span>${server.context ? "有上下文" : "无上下文"}</span>
        <span>${envKeys.length} 个 env key</span>
      </div>
      ${envKeys.length ? `<p class="secret-note">Env 值已隐藏：${envKeys.map(escapeHTML).join(", ")}</p>` : ""}
      <div class="row-actions">
        <button type="button" data-mcp-edit="${escapeHTML(server.name || "")}">编辑</button>
        <button type="button" data-mcp-toggle="${escapeHTML(server.name || "")}">${server.disabled ? "启用" : "停用"}</button>
        <button type="button" data-mcp-test="${escapeHTML(server.name || "")}" ${server.disabled ? "disabled" : ""}>测试连接</button>
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
  setText("mcp-save-state", server ? `正在编辑 ${server.name}` : "就绪");
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
  $("mcp-save-state").textContent = "保存中";
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
    showToast("MCP dock 已保存。");
  } catch (error) {
    $("mcp-save-state").textContent = "错误";
    showToast(error.message);
  }
}

async function toggleMCPServer(name) {
  const server = getMCPServer(name);
  if (!server) return;
  const replacement = mcpViewToPatch(server);
  replacement.disabled = !server.disabled;
  $("mcp-save-state").textContent = "保存中";
  try {
    const result = await api("/config", {
      method: "PATCH",
      body: JSON.stringify({ mcp_servers: buildMCPPatchServers(replacement) }),
    });
    state.config = result.config || state.config;
    renderMCPStarport();
    renderHomeDockedTools();
    showToast(replacement.disabled ? "MCP dock 已停用。" : "MCP dock 已启用。");
  } catch (error) {
    $("mcp-save-state").textContent = "错误";
    showToast(error.message);
  }
}

async function testMCPServer(name) {
  if (!name) return;
  const target = $(`mcp-test-${name}`);
  if (target) target.innerHTML = `<div class="inline-state">正在测试连接...</div>`;
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
    <strong>${escapeHTML(status === "ok" ? `发现 ${result.tool_count || tools.length} 个工具` : status)}</strong>
    <p>${escapeHTML(result?.error || (status === "ok" ? "连接测试成功。" : "连接测试完成。"))}</p>
    ${preview ? `<div class="pill-list">${preview}</div>` : ""}
  </div>`;
}

function renderFileIntake() {
  const result = state.intakeResult;
  setText("intake-summary", result ? `${result.mode === "archive_inspect" ? "已检查归档" : "已提取文档文本"}：${result.path || "本地路径"}。` : "先检查本地文档和归档，再把结果送入运行。");
  const overview = $("intake-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(result ? (result.is_error ? "需要处理" : "就绪") : "本地")}</strong><span>${escapeHTML(result ? (result.is_error ? "修复路径或模式后再分析。" : "结果可送入普通对话或运行流程。") : "先只读检查，再决定是否提取或总结。")}</span>`;
  }
  const target = $("intake-result");
  if (!target) return;
  $("intake-chat-button").disabled = !result || result.is_error;
  $("intake-extract-button").disabled = !result || result.mode !== "archive_inspect" || result.is_error;
  renderHomeDockedTools();
  renderContextReadinessBoard();
  renderManageCount();
  if (!result) {
    renderEmptyAction(target, "选择一个本地路径，用 document_text 或 archive_inspect 检查。", [
      { label: "打开对话", panel: "chat" },
    ]);
    return;
  }
  const status = result.is_error ? "error" : "ok";
  const preview = String(result.content || "").slice(0, 12000);
  target.innerHTML = `<article class="intake-result-card ${escapeHTML(status)}">
    <div class="row-item-title">
      <span>${escapeHTML(result.path || "本地文件")}</span>
      <span class="tag">${escapeHTML(result.mode || "intake")}</span>
    </div>
    <pre>${escapeHTML(preview || "没有返回内容。")}</pre>
  </article>`;
}

async function submitFileIntake(event) {
  event?.preventDefault?.();
  const path = $("intake-path").value.trim();
  if (!path) {
    showToast("需要填写文件路径。");
    return;
  }
  $("intake-state").textContent = "分析中";
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
    $("intake-state").textContent = result.is_error ? "错误" : "就绪";
    renderFileIntake();
    renderSourceRegistry();
    showToast(result.is_error ? "文件星舱返回错误。" : "文件星舱已就绪。");
  } catch (error) {
    $("intake-state").textContent = "错误";
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
  setText("manage-memory-count", `${count} 个来源`);
  renderManageCount();
  setText("memory-summary", count ? `${memoryFacts.length} 条分类事实 · ${memoryWarnings.length} 条警告` : "还没有记忆候选。");
  renderContextReadinessBoard();
  renderWorkspaceHealthStrip();
  renderKnowledgeCuration();
  renderPromptSuggestionDock();
  renderApprovalCenter();
  renderReviewQueue();
  const overview = $("memory-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(memoryFacts.length ? `${memoryFacts.length} 条事实` : memoryEntries.length ? `${memoryEntries.length} 个记忆文件` : count ? "来源就绪" : "预览")}</strong><span>${escapeHTML(memoryWarnings.length ? `添加更多记忆前需要复查 ${memoryWarnings.length} 条分类警告。` : state.memory?.memory_dir || (count ? "写入 MEMORY.md 前，先从最近工作起草可审核记忆。" : "收藏会话或完成运行后，可生成更强的记忆候选。"))}</span>`;
  }
  renderMemoryTaxonomyBar(state.memory?.categories || {});
  renderMemoryWarnings(memoryWarnings);
  const cards = [];
  const selectedCategory = state.memoryCategory || "all";
  const filteredFacts = selectedCategory === "all" ? memoryFacts : memoryFacts.filter((fact) => fact.category === selectedCategory);
  for (const fact of filteredFacts) {
    cards.push(`<article class="row-item memory-fact-card ${escapeHTML(fact.category || "uncategorized")}">
      <div class="row-item-title"><span>${escapeHTML(fact.text)}</span><span class="tag">${escapeHTML(fact.category || "未分类")}</span></div>
      <p>${escapeHTML(fact.entry || "MEMORY.md")} · line ${escapeHTML(fact.line || "-")}${fact.subject ? ` · ${escapeHTML(fact.subject)}` : ""}</p>
    </article>`);
  }
  for (const entry of memoryEntries) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card ${entry.primary ? "primary" : ""}">
      <div class="row-item-title"><span>${escapeHTML(entry.name)}</span><span class="tag">${entry.primary ? "当前记忆" : "记忆文件"}</span></div>
      <p>${escapeHTML(formatBytes(entry.size || 0))} · ${escapeHTML(formatTimestamp(entry.modified))}</p>
      <div class="row-actions">
        <button type="button" class="danger-button" data-memory-delete="${escapeHTML(entry.name)}">删除</button>
      </div>
    </article>`);
  }
  for (const session of favoriteSessions.slice(0, 4)) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card">
      <div class="row-item-title"><span>${escapeHTML(session.title || session.id)}</span><span class="tag">收藏会话</span></div>
      <p>${escapeHTML(session.id)}</p>
      <div class="row-actions">
        <button type="button" data-session-id="${escapeHTML(session.id)}">打开会话</button>
        <button type="button" data-memory-draft="session:${escapeHTML(session.id)}">起草记忆</button>
      </div>
    </article>`);
  }
  for (const run of recentRuns) {
    if (selectedCategory !== "all") continue;
    cards.push(`<article class="row-item memory-source-card">
      <div class="row-item-title"><span>${escapeHTML(run.prompt || run.id)}</span><span class="tag">最近运行</span></div>
      <p>${escapeHTML(run.status || "unknown")} · ${escapeHTML(formatTimestamp(run.started_at))}</p>
      <div class="row-actions">
        <button type="button" data-run-open="${escapeHTML(run.id)}">观测运行</button>
        <button type="button" data-memory-draft="run:${escapeHTML(run.id)}">起草记忆</button>
      </div>
    </article>`);
  }
  if (!cards.length) {
    renderEmptyAction(list, "还没有记忆来源。完成一次运行、收藏一个会话，或让 Astria 起草记忆星图。", [
      { label: "起草记忆星图", homeAction: "memory", primary: true },
      { label: "打开对话", panel: "chat" },
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
      type: "记忆",
      title: "已审查记忆",
      panel: "memory",
      evidence: memoryEntries.length + memoryFacts.length,
      freshness: memoryEntries[0]?.modified ? formatTimestamp(memoryEntries[0].modified) : "没有记忆文件",
      reliability: memoryFacts.length ? "已分类事实" : memoryEntries.length ? "文件支撑" : "需要种子",
      action: memoryFacts.length ? "审计分类和过期事实。" : "起草第一条已审查记忆来源。",
      prompt: `Audit Astria memory sources.\n\nMemory files: ${memoryEntries.length}\nClassified facts: ${memoryFacts.length}\nWarnings: ${(state.memory?.warnings || []).length}\n\nIdentify stale, duplicate, or missing durable facts and propose a maintenance action.`,
    },
    {
      id: "sessions",
      type: "会话",
      title: "收藏会话",
      panel: "memory",
      evidence: favoriteSessions.length,
      freshness: favoriteSessions[0]?.updated_at ? formatTimestamp(favoriteSessions[0].updated_at) : favoriteSessions[0]?.id || "没有收藏会话",
      reliability: favoriteSessions.length ? "操作者选择" : "需要收藏",
      action: favoriteSessions.length ? "把有用收藏转成记忆。" : "作为来源信任前，先收藏会话。",
      prompt: `Review favorite sessions as Astria knowledge sources.\n\nFavorites: ${favoriteSessions.map((session) => session.title || session.id).join(", ") || "none"}\n\nChoose what should become durable memory and what should remain ephemeral.`,
    },
    {
      id: "runs",
      type: "运行",
      title: "执行证据",
      panel: "runs",
      evidence: state.runs.length,
      freshness: latestRun?.started_at ? formatTimestamp(latestRun.started_at) : "没有运行",
      reliability: latestRun ? `最近状态 ${latestRun.status || "unknown"}` : "需要执行",
      action: latestRun ? "把稳定运行结果提升为记忆。" : "引用执行证据前，先跑一次基线任务。",
      prompt: `Review recent runs as knowledge sources.\n\nLatest run: ${latestRun?.prompt || "none"}\nStatus: ${latestRun?.status || "unknown"}\nRun count: ${state.runs.length}\n\nIdentify which outcomes are reliable enough to cite in future prompts.`,
    },
    {
      id: "intake",
      type: "文件星舱",
      title: intake?.path || "本地文件证据",
      panel: "intake",
      evidence: intake && !intake.is_error ? 1 : 0,
      freshness: intake?.mode || "没有文件星舱结果",
      reliability: intake?.is_error ? "错误" : intake ? "只读样本" : "需要文件",
      action: intake ? "使用前先总结来源限制。" : "检查一个文件，为来源知识播种。",
      prompt: `Review file intake as an Astria source.\n\nPath: ${intake?.path || "none"}\nMode: ${intake?.mode || "none"}\nError: ${Boolean(intake?.is_error)}\n\nState what can be trusted, what is incomplete, and what should be re-read.`,
    },
    {
      id: "council",
      type: "议会",
      title: latestCouncil?.goal || "议会综合结论",
      panel: "council",
      evidence: latestCouncil ? 1 + (latestCouncil.roles || []).length : 0,
      freshness: latestCouncil?.created_at ? formatTimestamp(latestCouncil.created_at) : "没有议会运行",
      reliability: latestCouncil?.synthesis ? "多角色综合" : "需要审查",
      action: latestCouncil ? "检查综合结论是否应成为记忆。" : "引用审查判断前先运行议会。",
      prompt: `Review council output as a knowledge source.\n\nGoal: ${latestCouncil?.goal || "none"}\nRoles: ${(latestCouncil?.roles || []).map((role) => role.role).join(", ") || "none"}\n\nDecide which conclusions are durable and which need another review.`,
    },
  ];
}

function renderSourceRegistry() {
  const rows = sourceRegistryRows();
  setText("nav-sources-count", rows.length);
  setText("manage-sources-count", `${rows.length} 个来源`);
  setText("sources-summary", `${rows.length} 条来源轨道正在追踪鲜度、可靠性和维护动作。`);
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
      <button type="button" data-source-select="${escapeHTML(row.id)}">来源简报</button>
      <button type="button" data-source-draft="${escapeHTML(row.id)}">起草维护</button>
      <button type="button" data-panel="${escapeHTML(row.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderSourceRegistryDetail(rows.find((row) => row.id === state.selectedSourceRow) || rows[0]);
}

function renderSourceRegistryDetail(row) {
  const target = $("source-registry-detail");
  if (!target) return;
  if (!row) {
    target.innerHTML = `<div class="empty-state">选择一条来源记录。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(row.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(row.type)}</strong>
        <span>证据</span><strong>${escapeHTML(String(row.evidence))}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(row.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>可靠性</h3>
      <p>${escapeHTML(row.reliability)}</p>
      <h3>维护动作</h3>
      <p>${escapeHTML(row.action)}</p>
      <div class="run-detail-actions">
        <button type="button" data-source-draft="${escapeHTML(row.id)}">起草维护</button>
        <button type="button" data-panel="${escapeHTML(row.panel)}">打开来源</button>
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

function knowledgeReconciliationItems() {
  const sources = sourceRegistryRows();
  const citationCards = citationGroundingCards();
  const resultEntries = resultArchiveEntries();
  const memoryWarnings = Array.isArray(state.memory?.warnings) ? state.memory.warnings : [];
  const memoryFacts = Array.isArray(state.memory?.facts) ? state.memory.facts : [];
  const memoryEntries = Array.isArray(state.memory?.entries) ? state.memory.entries : [];
  const uncategorizedFacts = memoryFacts.filter((fact) => String(fact.category || "").toLowerCase() === "uncategorized").length;
  const staleSource = sources.find((source) => /no |needs|stale|error/i.test(`${source.freshness} ${source.reliability} ${source.action}`)) || sources[0];
  const weakCitation = citationCards.find((card) => /gap|needs|unsupported|stale|missing/i.test(`${card.readiness} ${card.gap}`)) || citationCards[0];
  const latestResult = resultEntries[0];
  return [
    {
      id: "source-conflict",
      type: "Conflict",
      title: "Source conflict review",
      route: "sources",
      severity: sources.length > 1 ? "review" : "seed",
      evidence: `${sources.length} source lanes`,
      risk: "Different source lanes may support different versions of a claim.",
      resolution: "Compare source freshness, reliability, operator selection, and direct evidence before reusing the claim.",
      boundary: "Do not merge conflicting claims into memory until one source wins or the uncertainty is stated.",
      prompt: `Run an Astria source conflict reconciliation.\n\nSource lanes: ${sources.map((source) => `${source.title} (${source.reliability}, ${source.freshness})`).join("; ") || "none"}\n\nIdentify conflicting claims, stale sources, reliability differences, and the safest wording. Return a resolution table with winning source, rejected source, uncertainty, and next route.`,
    },
    {
      id: "stale-memory",
      type: "Freshness",
      title: "Stale memory review",
      route: "memory",
      severity: memoryEntries.length ? "audit" : "seed",
      evidence: `${memoryEntries.length} memory files, ${memoryFacts.length} facts`,
      risk: "Durable memory can become stale when source dates, product versions, or user preferences change.",
      resolution: "Add freshness notes, reject expired facts, and route unstable claims to browser or source review.",
      boundary: "Treat time-sensitive memory as untrusted until a dated source confirms it.",
      prompt: `Run an Astria stale memory reconciliation.\n\nMemory files: ${memoryEntries.length}\nFacts: ${memoryFacts.length}\nWarnings: ${memoryWarnings.length}\n\nFind stale, time-sensitive, duplicate, or unsupported memory. For each item, decide keep, update, re-source, or reject, and include a freshness note.`,
    },
    {
      id: "weak-citation",
      type: "Citation",
      title: "Weak citation escalation",
      route: "citation",
      severity: weakCitation?.readiness || "review",
      evidence: weakCitation?.evidence || "citation gap",
      risk: weakCitation?.gap || "A claim has weak or missing citation support.",
      resolution: "Map atomic claims to sources, capture direct evidence, and block confident wording for gaps.",
      boundary: "Do not present assumptions as sourced facts.",
      prompt: `Run an Astria weak citation escalation.\n\nCard: ${weakCitation?.title || "none"}\nGap: ${weakCitation?.gap || "missing citation"}\nEvidence: ${weakCitation?.evidence || "none"}\n\nProduce a claim map, required source list, accepted citations, unsupported claims, and safe uncertainty wording.`,
    },
    {
      id: "duplicate-memory",
      type: "Duplicate",
      title: "Duplicate or uncategorized memory",
      route: "memory",
      severity: memoryWarnings.length || uncategorizedFacts ? "cleanup" : "monitor",
      evidence: `${memoryWarnings.length} warnings, ${uncategorizedFacts} uncategorized facts`,
      risk: "Duplicate or uncategorized memory makes future retrieval ambiguous.",
      resolution: "Normalize subjects, merge duplicates, categorize useful facts, and reject vague notes.",
      boundary: "Do not add more memory until obvious duplicates or uncategorized facts are reviewed.",
      prompt: `Run an Astria memory duplicate reconciliation.\n\nWarnings: ${memoryWarnings.join("; ") || "none"}\nUncategorized facts: ${uncategorizedFacts}\n\nFind duplicate subjects, uncategorized facts, vague notes, and merge candidates. Return approve/update/reject actions with categories and source notes.`,
    },
    {
      id: "missing-coverage",
      type: "Coverage",
      title: "Missing source coverage",
      route: "sources",
      severity: sources.some((source) => !source.evidence) ? "gap" : "review",
      evidence: `${sources.filter((source) => Number(source.evidence || 0) > 0).length}/${sources.length} source lanes with evidence`,
      risk: "A result may rely on source lanes that have no evidence count or no operator-selected source.",
      resolution: "Require at least one reliable lane per important claim or escalate to evidence capture.",
      boundary: "Do not reuse results as durable knowledge when the source lane is empty.",
      prompt: `Run an Astria source coverage reconciliation.\n\nSource lanes: ${sources.map((source) => `${source.id}: ${source.evidence} evidence`).join("; ") || "none"}\n\nIdentify claims or result types without enough source coverage, then route each gap to source registry, browser capture, data planner, memory review, or share-pack caveat.`,
    },
    {
      id: "privacy-boundary",
      type: "Privacy",
      title: "Privacy and approval boundary",
      route: "share",
      severity: "guarded",
      evidence: `${state.inboxItems.length} inbox items, ${state.schedules.length} schedules`,
      risk: "Handoffs, delivery, browser actions, and file evidence can expose private paths, secrets, or remote state changes.",
      resolution: "Redact private data, require approval before outbound delivery, and state what must remain local.",
      boundary: "No external send, schedule, form submit, account action, or private excerpt reuse without approval.",
      prompt: `Run an Astria privacy and approval reconciliation.\n\nInbox items: ${state.inboxItems.length}\nSchedules: ${state.schedules.length}\nLatest result: ${latestResult?.title || "none"}\n\nFind private paths, secrets, external delivery risks, browser/form actions, and approval blockers. Return redactions, local-only boundaries, and required approvals.`,
    },
    {
      id: "result-freshness",
      type: "Result",
      title: "Result freshness review",
      route: "results",
      severity: latestResult ? latestResult.review : "seed",
      evidence: latestResult?.evidence || "no archived result",
      risk: "Saved results can outlive source freshness or hide unresolved assumptions.",
      resolution: "Attach source dates, unresolved risks, reuse limits, and a recheck route before follow-up.",
      boundary: "Do not promote archived results into playbooks or memory without freshness and evidence review.",
      prompt: `Run an Astria result freshness reconciliation.\n\nResult: ${latestResult?.title || "none"}\nEvidence: ${latestResult?.evidence || "none"}\nFreshness: ${latestResult?.freshness || "unknown"}\n\nReview whether the result is still reusable, what source dates are missing, which assumptions remain open, and where to recheck before reuse.`,
    },
  ];
}

function renderKnowledgeReconciliation() {
  const items = knowledgeReconciliationItems();
  setText("nav-reconcile-count", items.length);
  setText("manage-reconcile-count", `${items.length} risk${items.length === 1 ? "" : "s"}`);
  setText("reconcile-summary", `${items.length} reconciliation item${items.length === 1 ? "" : "s"} across conflicts, stale memory, weak citations, duplicate facts, source coverage, privacy, and result freshness.`);
  const list = $("knowledge-reconcile-list");
  if (!list) return;
  if (!state.selectedReconcileRisk || !items.some((item) => item.id === state.selectedReconcileRisk)) {
    state.selectedReconcileRisk = items[0]?.id || "";
  }
  list.innerHTML = items.map((item) => `<article class="reconcile-card ${item.id === state.selectedReconcileRisk ? "active" : ""}" data-reconcile-risk="${escapeHTML(item.id)}">
    <div class="row-item-title"><span>${escapeHTML(item.type)}</span><span class="tag">${escapeHTML(item.severity)}</span></div>
    <strong>${escapeHTML(item.title)}</strong>
    <div class="reconcile-grid">
      <span>Evidence</span><strong>${escapeHTML(item.evidence)}</strong>
      <span>Risk</span><strong>${escapeHTML(item.risk)}</strong>
      <span>Resolution</span><strong>${escapeHTML(item.resolution)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-reconcile-select="${escapeHTML(item.id)}">解决简报</button>
      <button type="button" data-reconcile-draft="${escapeHTML(item.id)}">起草解决</button>
      <button type="button" data-panel="${escapeHTML(item.route)}">打开路径</button>
    </div>
  </article>`).join("");
  renderKnowledgeReconciliationDetail(items.find((item) => item.id === state.selectedReconcileRisk) || items[0]);
}

function renderKnowledgeReconciliationDetail(item) {
  const target = $("knowledge-reconcile-detail");
  if (!target) return;
  if (!item) {
    target.innerHTML = `<div class="empty-state">选择一个知识校验项。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(item.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(item.type)}</strong>
        <span>严重度</span><strong>${escapeHTML(item.severity)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(item.route))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>风险</h3>
      <p>${escapeHTML(item.risk)}</p>
      <h3>解决动作</h3>
      <p>${escapeHTML(item.resolution)}</p>
      <h3>可信边界</h3>
      <p>${escapeHTML(item.boundary)}</p>
      <div class="run-detail-actions">
        <button type="button" data-reconcile-draft="${escapeHTML(item.id)}">起草解决</button>
        <button type="button" data-panel="${escapeHTML(item.route)}">打开路径</button>
      </div>
    </section>
  </div>`;
}

function reconcileRiskByID(id) {
  return knowledgeReconciliationItems().find((item) => item.id === id) || null;
}

function draftReconciliationToChat(id) {
  const item = reconcileRiskByID(id);
  if (!item) return;
  $("chat-input").value = `${item.prompt}\n\nReturn a reconciliation brief with risk, evidence, resolution action, confidence boundary, source route, and what must not be reused yet.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Reconciliation drafted to chat.");
}

function citationGroundingContext() {
  const claim = ($("citation-claim-scope")?.value || state.citationClaimScope || "").trim();
  const posture = ($("citation-source-posture")?.value || state.citationSourcePosture || "").trim();
  const evidence = ($("citation-evidence-level")?.value || state.citationEvidenceLevel || "").trim();
  return {
    claim: claim || "The claim, answer, data finding, browser result, or handoff section that needs grounding.",
    posture: posture || "Use the strongest available local sources first, then request fresh browser evidence if needed.",
    evidence: evidence || "Cited summary with direct quotes or gap report for unsupported claims.",
    hasClaim: Boolean(claim),
  };
}

function citationGroundingCards() {
  const ctx = citationGroundingContext();
  const sourceCount = sourceRegistryRows().length;
  const memoryCount = Array.isArray(state.memory?.entries) ? state.memory.entries.length : 0;
  const latestRun = state.runs[0];
  return [
    {
      id: "coverage",
      type: "Coverage",
      title: "Source coverage check",
      panel: "sources",
      evidence: `${sourceCount} source lanes`,
      readiness: sourceCount ? "review" : "seed",
      rule: "List which claims have sources and which still need evidence.",
      gap: "Missing source lane, stale memory, or unsupported conclusion.",
      prompt: `Plan an Astria source coverage check.\n\nClaim scope: ${ctx.claim}\nSource posture: ${ctx.posture}\nEvidence level: ${ctx.evidence}\nSource lanes: ${sourceCount}\n\nMap the claim into citeable parts, match each part to current Astria sources, identify unsupported pieces, and return a coverage table with source, reliability, freshness, and next evidence action. Keep the result local and do not invent citations.`,
    },
    {
      id: "claim-map",
      type: "Claim map",
      title: "Claim-to-citation map",
      panel: "memory",
      evidence: `${memoryCount} memory entries`,
      readiness: memoryCount ? "grounded" : "needs memory",
      rule: "Separate durable memory from fresh evidence requirements.",
      gap: "Claim depends on memory without source or freshness note.",
      prompt: `Plan an Astria claim-to-citation map.\n\nClaim scope: ${ctx.claim}\nSource posture: ${ctx.posture}\nEvidence level: ${ctx.evidence}\nMemory entries: ${memoryCount}\nLatest run: ${latestRun?.prompt || "none"}\n\nBreak the answer into atomic claims, assign each claim to memory, run evidence, browser evidence, data evidence, or unresolved gap, and state what cannot be cited safely.`,
    },
    {
      id: "quote-capture",
      type: "Quote",
      title: "Quote and evidence capture",
      panel: "browser",
      evidence: "browser or source excerpt",
      readiness: "capture",
      rule: "Use short quotes or precise summaries; avoid private or irrelevant content.",
      gap: "No direct excerpt, screenshot, selector, or source location for key claim.",
      prompt: `Plan an Astria quote and evidence capture mission.\n\nClaim scope: ${ctx.claim}\nSource posture: ${ctx.posture}\nEvidence level: ${ctx.evidence}\n\nIdentify which claims need direct quotes, screenshots, selectors, or source locations. Capture only what is necessary, cite the source location, and flag private or sensitive content that should not be copied.`,
    },
    {
      id: "freshness",
      type: "Freshness",
      title: "Freshness and version risk",
      panel: "data",
      evidence: "timestamp + source version",
      readiness: "time-sensitive",
      rule: "Treat unstable facts as stale until verified with a dated source.",
      gap: "No timestamp, version, release date, or data extraction date.",
      prompt: `Plan an Astria freshness and version risk review.\n\nClaim scope: ${ctx.claim}\nSource posture: ${ctx.posture}\nEvidence level: ${ctx.evidence}\n\nIdentify time-sensitive claims, required source dates, data extraction timestamps, version labels, and stale-risk warnings. Return what must be rechecked before the answer is trusted.`,
    },
    {
      id: "gap-escalation",
      type: "Gap",
      title: "Evidence gap escalation",
      panel: "share",
      evidence: "gap report",
      readiness: "escalate",
      rule: "Block confident wording when evidence is incomplete.",
      gap: "Unverified claim, weak source, conflict, or missing approval.",
      prompt: `Plan an Astria evidence gap escalation.\n\nClaim scope: ${ctx.claim}\nSource posture: ${ctx.posture}\nEvidence level: ${ctx.evidence}\n\nList unsupported claims, weak citations, conflicting sources, privacy or approval blockers, and the next route to resolve each gap. Produce safe wording that distinguishes facts, assumptions, and open questions.`,
    },
  ];
}

function renderCitationGroundingPlanner() {
  const cards = citationGroundingCards();
  setText("nav-citation-count", cards.length);
  setText("manage-citation-count", `${cards.length} check${cards.length === 1 ? "" : "s"}`);
  setText("citation-summary", `${cards.length} grounding check${cards.length === 1 ? "" : "s"} for source coverage, claim maps, quote capture, freshness, and evidence gaps.`);
  const list = $("citation-grounding-cards");
  if (!list) return;
  if (!state.selectedCitationGrounding || !cards.some((card) => card.id === state.selectedCitationGrounding)) {
    state.selectedCitationGrounding = cards[0]?.id || "";
  }
  list.innerHTML = cards.map((card) => `<article class="citation-grounding-card ${card.id === state.selectedCitationGrounding ? "active" : ""}" data-citation-grounding="${escapeHTML(card.id)}">
    <div class="row-item-title"><span>${escapeHTML(card.type)}</span><span class="tag">${escapeHTML(card.readiness)}</span></div>
    <strong>${escapeHTML(card.title)}</strong>
    <div class="citation-grounding-grid">
      <span>Evidence</span><strong>${escapeHTML(card.evidence)}</strong>
      <span>Citation rule</span><strong>${escapeHTML(card.rule)}</strong>
      <span>Gap trigger</span><strong>${escapeHTML(card.gap)}</strong>
    </div>
    <div class="row-actions">
      <button type="button" data-citation-select="${escapeHTML(card.id)}">校准简报</button>
      <button type="button" data-citation-draft="${escapeHTML(card.id)}">起草校准</button>
      <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
    </div>
  </article>`).join("");
  renderCitationGroundingDetail(cards.find((card) => card.id === state.selectedCitationGrounding) || cards[0]);
}

function renderCitationGroundingDetail(card) {
  const target = $("citation-grounding-detail");
  if (!target) return;
  if (!card) {
    target.innerHTML = `<div class="empty-state">选择一张引用校准卡。</div>`;
    return;
  }
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(card.title)}</h3>
      <div class="run-meta-grid">
        <span>类型</span><strong>${escapeHTML(card.type)}</strong>
        <span>就绪度</span><strong>${escapeHTML(card.readiness)}</strong>
        <span>路径</span><strong>${escapeHTML(panelName(card.panel))}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>Citation rule</h3>
      <p>${escapeHTML(card.rule)}</p>
      <h3>Gap trigger</h3>
      <p>${escapeHTML(card.gap)}</p>
      <div class="run-detail-actions">
        <button type="button" data-citation-draft="${escapeHTML(card.id)}">起草校准</button>
        <button type="button" data-panel="${escapeHTML(card.panel)}">打开来源</button>
      </div>
    </section>
  </div>`;
}

function citationGroundingByID(id) {
  return citationGroundingCards().find((card) => card.id === id) || null;
}

function draftCitationGroundingToChat(id) {
  const card = citationGroundingByID(id);
  if (!card) return;
  $("chat-input").value = `${card.prompt}\n\nReturn a citation grounding brief with claim map, required sources, citation rules, freshness checks, evidence gaps, uncertainty wording, and next Astria route.`;
  $("chat-new-session").checked = true;
  state.activeSessionID = "";
  updateActiveSessionLabel();
  switchPanel("chat");
  $("chat-input").focus();
  showToast("Citation grounding drafted to chat.");
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
    setText("settings-version-state", data.version || "构建");
    setClass("settings-version-state", data.update_supported ? "ready" : "warning");
    renderVersion();
    renderSystemStatusBoard();
  } catch (error) {
    state.version = null;
    setText("settings-version-state", "错误");
    setClass("settings-version-state", "bad");
    $("version-summary").textContent = "版本元数据不可用。";
    $("update-check-state").textContent = "错误";
    renderError(list, error.message);
    renderSystemStatusBoard();
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
    summary.textContent = diagnostics.summary || "运行时就绪检查。";
    overview.innerHTML = `<strong>${escapeHTML(label)}</strong><span>${escapeHTML(diagnostics.summary || "")}</span>`;
    renderConfigDiagnosticsOverview(diagnostics);
    renderWorkspaceHealthStrip();
    renderSystemStatusBoard();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
    if ($("chat-output").querySelector(".empty-thread")) renderEmptyThread();
    const checks = Array.isArray(diagnostics.checks) ? diagnostics.checks : [];
    const launchRows = diagnosticsLaunchRows(diagnostics);
    const launchCard = `<article class="row-item diagnostic-launch-card">
      <div class="row-item-title"><span>启动就绪</span><span class="tag">${escapeHTML(label)}</span></div>
      <div class="run-meta-grid">
        ${launchRows.map(([rowLabel, value]) => `<span>${escapeHTML(rowLabel)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
      </div>
    </article>`;
    if (!checks.length) {
      list.innerHTML = `${launchCard}<article class="row-item empty-state"><p>尚未返回诊断结果。</p></article>`;
      return;
    }
    const checkCards = checks.map((check) => `<article class="row-item diagnostic-item ${escapeHTML(check.status || "unknown")}">
      <div class="row-item-title">
        <span>${escapeHTML(check.label || check.id || "检查项")}</span>
        <span class="tag diagnostic-tag ${escapeHTML(check.status || "unknown")}">${escapeHTML(statusLabel(check.status))}</span>
      </div>
      <p>${escapeHTML(check.detail || "")}</p>
      ${diagnosticActionHTML(check)}
    </article>`).join("");
    list.innerHTML = `${launchCard}${checkCards}`;
  } catch (error) {
    state.diagnostics = null;
    setText("settings-state", "离线");
    setClass("settings-state", "error");
    setText("nav-diagnostics-state", "离线");
    setClass("nav-diagnostics-state", "error");
    setText("settings-diagnostics-state", "离线");
    setClass("settings-diagnostics-state", "error");
    chip.textContent = "诊断不可用";
    chip.className = "diagnostics-chip error";
    summary.textContent = "诊断不可用。";
    overview.innerHTML = `<strong>错误</strong><span>${escapeHTML(error.message)}</span>`;
    renderConfigDiagnosticsOverview({ status: "error", summary: error.message });
    renderError(list, error.message);
    renderWorkspaceHealthStrip();
    renderSystemStatusBoard();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  }
}

function diagnosticsLaunchRows(diagnostics) {
  const rows = [
    ["启动命令", diagnostics.launch_command || "starclaw app"],
    ["Web UI", diagnostics.web_url || "-"],
    ["健康检查", diagnostics.health_url || "-"],
    ["Status API", diagnostics.status_url || "-"],
    ["诊断", diagnostics.diagnostics_url || "-"],
    ["数据目录", diagnostics.starclaw_dir || "-"],
    ["配置", diagnostics.config_path || "-"],
    ["Agent 目录", diagnostics.agents_dir || "-"],
    ["会话目录", diagnostics.sessions_dir || "-"],
  ];
  if (diagnostics.executable_path) rows.push(["可执行文件", diagnostics.executable_path]);
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
    renderContextReadinessBoard();
    renderSystemStatusBoard();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
    renderMissionGraph();
  } catch (error) {
    state.config = null;
    setText("settings-config-state", "Error");
    setClass("settings-config-state", "bad");
    $("config-save-state").textContent = error.message;
    renderMCPStarport();
    renderToolDockInspector();
    renderContextReadinessBoard();
    renderSystemStatusBoard();
    renderPromptSuggestionDock();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
    renderMissionGraph();
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
  $("config-api-key").placeholder = cfg.api_key_set ? "已保存；留空保持不变" : "Anthropic 必填";
  $("config-openai-api-key").placeholder = cfg.openai_api_key_set ? "已保存；留空保持不变" : "OpenAI 必填";
  setText("settings-config-state", cfg.provider || "未配置");
  setClass("settings-config-state");
  $("config-save-state").textContent = "已加载";
  updateProviderFields();
  renderConnectorSetupCard(cfg);
}

function updateProviderFields() {
  const provider = $("config-provider").value || "anthropic";
  document.querySelectorAll("[data-provider-fields]").forEach((group) => {
    group.hidden = group.dataset.providerFields !== provider;
  });
  renderConnectorSetupCard(state.config || {});
}

function configReadiness(cfg = {}, values = {}) {
  const provider = values.provider || cfg.provider || "anthropic";
  const missing = [];
  let keySet = false;
  if (provider === "openai") {
    if (!String(values.openaiEndpoint ?? cfg.openai_endpoint ?? "").trim()) missing.push("Base URL");
    if (!String(values.openaiModel ?? cfg.openai_model ?? "").trim()) missing.push("Model");
    keySet = Boolean(cfg.openai_api_key_set || String(values.openaiAPIKey || "").trim());
    if (!keySet) missing.push("API key");
  } else if (provider === "ollama") {
    if (!String(values.ollamaEndpoint ?? cfg.ollama_endpoint ?? "").trim()) missing.push("Base URL");
    if (!String(values.ollamaModel ?? cfg.ollama_model ?? "").trim()) missing.push("Model");
    keySet = true;
  } else {
    if (!String(values.endpoint ?? cfg.endpoint ?? "").trim()) missing.push("Base URL");
    if (!String(values.modelTier ?? cfg.model_tier ?? "").trim()) missing.push("Model");
    keySet = Boolean(cfg.api_key_set || String(values.apiKey || "").trim());
    if (!keySet) missing.push("API key");
  }
  return {
    provider,
    missing,
    ready: missing.length === 0,
    keySet,
  };
}

function connectorReadiness(cfg = {}) {
  return configReadiness(cfg, {
    provider: $("config-provider")?.value || cfg.provider || "anthropic",
    endpoint: $("config-endpoint")?.value,
    modelTier: $("config-model-tier")?.value,
    apiKey: $("config-api-key")?.value,
    openaiEndpoint: $("config-openai-endpoint")?.value,
    openaiModel: $("config-openai-model")?.value,
    openaiAPIKey: $("config-openai-api-key")?.value,
    ollamaEndpoint: $("config-ollama-endpoint")?.value,
    ollamaModel: $("config-ollama-model")?.value,
  });
}

function renderConnectorSetupCard(cfg = {}) {
  const card = document.querySelector("[data-connector-setup-card]");
  if (!card) return;
  const title = card.querySelector("[data-connector-status-title]");
  const detail = card.querySelector("[data-connector-status-detail]");
  const readiness = connectorReadiness(cfg);
  card.dataset.ready = readiness.ready ? "true" : "false";
  if (title) title.textContent = readiness.ready ? `${readiness.provider} 已准备` : "等待用户填写连接";
  if (detail) {
    detail.textContent = readiness.ready
      ? "URL、模型和凭据状态完整。API key 已保存时不会在界面回显。"
      : `还需要：${readiness.missing.join("、") || "选择 provider"}`;
  }
}

function connectorTestMessage(result = {}) {
  const messages = {
    auth_failed: "API key 无效或权限不足。请检查 key 是否属于当前 Base URL，并确认已开通该模型。",
    invalid_response: "响应格式不兼容。请确认 Base URL 指向兼容的 Chat Completions 或 Messages API。",
    missing_fields: result.detail || "连接配置还不完整。请先补齐 Base URL、模型和 API key。",
    model_not_found: "模型不可用。请检查模型名拼写，或确认当前 key 有权限访问该模型。",
    network_unreachable: "无法连接到 Base URL。请检查地址、端口、代理或本地服务是否已启动。",
    ok: result.detail || "连接成功，模型返回了有效响应。",
    provider_error: result.detail || "provider 返回错误。请检查 Base URL、模型和凭据。",
    provider_unavailable: "provider 暂时不可用或返回服务端错误。请稍后重试。",
    rate_limited: "provider 返回限流。请稍后重试，或检查额度与频率限制。",
    timeout: "连接超时。请检查网络、代理或本地服务状态，然后重试。",
    unsupported_provider: result.detail || "当前 provider 暂不支持连接测试。",
  };
  return messages[result.code] || result.detail || "保存 provider 设置后，手动检查 Base URL、模型和凭据是否可用。";
}

function renderConnectorTestState(result) {
  const card = $("connector-test-card");
  if (!card) return;
  const title = $("connector-test-title");
  const detail = $("connector-test-detail");
  const status = result?.status || "idle";
  card.dataset.status = status;
  if (title) {
    title.textContent = status === "ready"
      ? "连接可用"
      : status === "needs_setup"
        ? "配置未完成"
        : status === "checking"
          ? "检查中"
          : status === "error"
            ? "连接失败"
            : "尚未检查";
  }
  if (detail) {
    if (status === "checking") {
      detail.textContent = "正在向已保存的 provider 发送一次最小测试请求。";
    } else if (result?.detail || result?.code) {
      const suffix = result.duration_ms ? ` · ${result.duration_ms}ms` : "";
      detail.textContent = `${connectorTestMessage(result)}${suffix}`;
    } else {
      detail.textContent = "保存 provider 设置后，手动检查 Base URL、模型和凭据是否可用。";
    }
  }
}

async function testProviderConnection() {
  const button = $("config-test-button");
  if (button) button.disabled = true;
  renderConnectorTestState({ status: "checking" });
  try {
    const result = await api("/config/test", { method: "POST", body: "{}" });
    renderConnectorTestState(result);
    await loadDiagnostics();
    showToast(result.status === "ready" ? "LLM 连接检查通过。" : "LLM 连接检查未通过。");
  } catch (error) {
    renderConnectorTestState({ status: "error", detail: error.message });
    showToast(error.message);
  } finally {
    if (button) button.disabled = false;
  }
}

function firstMissingProviderField(provider, missing = []) {
  const needs = new Set(missing);
  if (provider === "openai") {
    if (needs.has("Base URL")) return "config-openai-endpoint";
    if (needs.has("Model")) return "config-openai-model";
    if (needs.has("API key")) return "config-openai-api-key";
  } else if (provider === "ollama") {
    if (needs.has("Base URL")) return "config-ollama-endpoint";
    if (needs.has("Model")) return "config-ollama-model";
  } else {
    if (needs.has("Base URL")) return "config-endpoint";
    if (needs.has("Model")) return "config-model-tier";
    if (needs.has("API key")) return "config-api-key";
  }
  return "config-provider";
}

function focusProviderSetup(readiness) {
  const field = $(firstMissingProviderField(readiness.provider, readiness.missing));
  const setupCard = document.querySelector("[data-connector-setup-card]");
  requestAnimationFrame(() => {
    setupCard?.scrollIntoView({ block: "nearest", behavior: "smooth" });
    field?.focus();
  });
}

function ensureProviderReadyForLaunch(surface = "任务") {
  const readiness = configReadiness(state.config || {});
  if (readiness.ready) return true;
  switchPanel("config");
  renderConnectorSetupCard(state.config || {});
  focusProviderSetup(readiness);
  showToast(`${surface}启动前请先完成 LLM 连接：缺少 ${readiness.missing.join("、") || "provider 设置"}。`);
  return false;
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
  $("config-save-state").textContent = "保存中";
  try {
    const result = await api("/config", {
      method: "PATCH",
      body: JSON.stringify(buildConfigPatch()),
    });
    state.config = result.config || state.config;
    renderConfigForm();
    renderConnectorTestState({ status: "idle", detail: "provider 设置已保存，可以手动检查连接。" });
    await loadDiagnostics();
    showToast("provider 设置已保存。");
  } catch (error) {
    $("config-save-state").textContent = "错误";
    showToast(error.message);
  }
}

function renderConfigDiagnosticsOverview(diagnostics) {
  const target = $("config-diagnostics-overview");
  if (!target) return;
  const status = diagnostics?.status || "unknown";
  target.innerHTML = `<strong>${escapeHTML(statusLabel(status))}</strong><span>${escapeHTML(diagnostics?.summary || "运行时诊断不可用。")}</span>`;
}

async function loadPermissions() {
  const list = $("permissions-list");
  try {
    const data = await api("/permissions");
    state.permissions = data.permissions || {};
    fillPermissionsForm();
    renderPermissions();
    renderWorkspaceHealthStrip();
    renderSystemStatusBoard();
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
    renderSystemStatusBoard();
    renderApprovalCenter();
    renderReviewQueue();
    renderProactiveDeliveryBoard();
  }
}

async function loadMemory() {
  try {
    const [memory, memoryStatus] = await Promise.all([
      api("/memory"),
      api("/memory/status").catch(() => null),
    ]);
    state.memory = memory;
    state.memoryStatus = memoryStatus;
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderKnowledgeReconciliation();
    renderAgentContinuityDigest();
    renderHomeDockedTools();
    renderComparisonWorkbench();
    renderRunQualityScorecard();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderResultLibrary();
    renderPlaybookLibrary();
    renderSharePackBuilder();
    renderWorkspaceSnapshotPlanner();
    renderMissionGraph();
  } catch (error) {
    state.memory = { entries: [], content: "", memory_dir: "" };
    state.memoryStatus = null;
    setText("memory-save-state", "Error");
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderKnowledgeReconciliation();
    renderAgentContinuityDigest();
    renderHomeDockedTools();
    renderComparisonWorkbench();
    renderRunQualityScorecard();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderResultLibrary();
    renderPlaybookLibrary();
    renderSharePackBuilder();
    renderWorkspaceSnapshotPlanner();
    renderMissionGraph();
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
    renderKnowledgeReconciliation();
    renderComparisonWorkbench();
    renderRunQualityScorecard();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderResultLibrary();
    renderPlaybookLibrary();
    renderSharePackBuilder();
    renderWorkspaceSnapshotPlanner();
    showToast("Memory approved.");
  } catch (error) {
    $("memory-save-state").textContent = "Error";
    showToast(error.message);
  }
}

async function deleteMemoryEntry(name) {
  if (!name) return;
  if (!globalThis.confirm(`删除记忆条目 "${name}"？`)) return;
  try {
    state.memory = await api(`/memory/${encodeURIComponent(name)}`, { method: "DELETE" });
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderKnowledgeReconciliation();
    renderComparisonWorkbench();
    renderRunQualityScorecard();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderResultLibrary();
    renderPlaybookLibrary();
    renderSharePackBuilder();
    renderWorkspaceSnapshotPlanner();
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
    target.textContent = "还没有候选。";
    return;
  }
  const categories = facts.map((fact) => memoryCategoryLabel(fact.category)).join(", ");
  const existingTexts = new Set((state.memory?.facts || []).map((fact) => normalizeCandidateText(fact.text)));
  const duplicate = facts.some((fact) => existingTexts.has(normalizeCandidateText(fact.text)));
  target.innerHTML = `<strong>${escapeHTML(categories)}</strong><span>${duplicate ? "确认前可能存在重复。" : "已准备好进入记忆审核。"}</span>`;
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
    renderKnowledgeReconciliation();
    renderComparisonWorkbench();
    renderReuseGallery();
    renderResultLibrary();
  } catch (error) {
    state.councilRuns = [];
    state.currentCouncilRun = null;
    setText("council-state", "错误");
    renderError(list, error.message);
    renderComparisonWorkbench();
    renderSourceRegistry();
    renderKnowledgeReconciliation();
    renderReuseGallery();
    renderResultLibrary();
  }
}

function renderCouncilRuns() {
  const list = $("council-list");
  if (!list) return;
  const count = state.councilRuns.length;
  setText("nav-council-count", count);
  setText("manage-council-count", `${count} 条评审`);
  renderManageCount();
  setText("council-summary", count ? `${count} 条议会运行包含角色贡献。` : "还没有议会运行。");
  if (!count) {
    renderEmptyAction(list, "还没有议会运行。先从一个规划或评审目标开始。", [
      { label: "填入议会目标", homeAction: "council", primary: true },
      { label: "打开对话", panel: "chat" },
    ]);
    renderCouncilDetail(state.currentCouncilRun);
    return;
  }
  list.innerHTML = state.councilRuns.map((run) => `<article class="row-item council-run-card ${state.currentCouncilRun?.id === run.id ? "active" : ""}" data-council-id="${escapeHTML(run.id)}">
    <div class="row-item-title"><span>${escapeHTML(run.goal || run.id)}</span><span class="tag">${escapeHTML(run.status || "unknown")}</span></div>
    <p>${escapeHTML((run.roles || []).map((role) => role.role).join(" · ") || "暂无角色贡献")} · ${escapeHTML(formatTimestamp(run.created_at))}</p>
    <div class="row-actions">
      <button type="button" data-council-open="${escapeHTML(run.id)}">打开议会</button>
      <button type="button" data-council-copy="${escapeHTML(run.id)}">复制综合结论</button>
    </div>
  </article>`).join("");
  if (!state.currentCouncilRun) renderCouncilDetail(state.councilRuns[0]);
}

function renderCouncilDetail(run) {
  const target = $("council-detail");
  if (!target) return;
  state.currentCouncilRun = run || null;
  if (!run) {
    target.innerHTML = `<div class="empty-state">启动或选择一条议会运行。</div>`;
    return;
  }
  const roles = Array.isArray(run.roles) ? run.roles : [];
  const stages = councilStages(run, roles);
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-detail-section">
      <h3>${escapeHTML(run.goal || "议会运行")}</h3>
      <div class="run-meta-grid">
        <span>状态</span><strong>${escapeHTML(run.status || "unknown")}</strong>
        <span>创建时间</span><strong>${escapeHTML(formatTimestamp(run.created_at))}</strong>
        <span>角色</span><strong>${roles.length}</strong>
      </div>
    </section>
    <section class="run-detail-section">
      <h3>议会阶段</h3>
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
      <h3>角色贡献</h3>
      <div class="council-role-list">
        ${roles.map((role, index) => `<article class="council-role-card">
          <div class="row-item-title"><span>${escapeHTML(role.role || "角色")}</span><span class="tag">${escapeHTML(role.status || "unknown")}</span></div>
          <strong>${escapeHTML(role.summary || "")}</strong>
          <p>${escapeHTML(role.notes || "")}</p>
          <div class="row-actions">
            <button type="button" data-council-role-copy="${escapeHTML(run.id)}" data-council-role-index="${index}">复制笔记</button>
            <button type="button" data-council-role-draft="${escapeHTML(run.id)}" data-council-role-index="${index}">起草到对话</button>
          </div>
        </article>`).join("")}
      </div>
    </section>
    <section class="run-detail-section">
      <h3>最终综合结论</h3>
      <pre>${escapeHTML(run.synthesis || "")}</pre>
      <div class="run-detail-actions">
        <button type="button" data-council-copy="${escapeHTML(run.id)}">复制综合结论</button>
        <button type="button" data-council-send="${escapeHTML(run.id)}">发送到对话</button>
        <button type="button" class="primary-button" data-council-run="${escapeHTML(run.id)}">启动运行</button>
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
      summary: role.summary || "角色输出待生成。",
      preview: role.notes || "还没有记录笔记。",
      actions: [
        { label: "复制笔记", attr: `data-council-role-copy="${escapeHTML(run.id)}" data-council-role-index="${roleIndex}"` },
        { label: "起草到对话", attr: `data-council-role-draft="${escapeHTML(run.id)}" data-council-role-index="${roleIndex}"` },
      ],
    };
  });
  return [
    ...roleStages,
    {
      kind: "synthesis",
      step: "4",
      title: "综合结论",
      status: run.synthesis ? "就绪" : "待生成",
      summary: "把角色输出合并成一个实施方向。",
      preview: run.synthesis || "还没有综合结论。",
      actions: [
        { label: "复制综合结论", attr: `data-council-copy="${escapeHTML(run.id)}"` },
        { label: "发送到对话", attr: `data-council-send="${escapeHTML(run.id)}"` },
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
    renderEmpty(list, "尚未返回收件渠道。");
    return;
  }
  list.innerHTML = state.inboxProviders.map((provider) => `<article class="provider-route-card ${provider.kind || ""}">
    <div class="row-item-title">
      <span>${escapeHTML(provider.name || provider.kind || "渠道")}</span>
      <span class="tag">${escapeHTML(provider.configured ? "就绪" : "待设置")}</span>
    </div>
    <code>${escapeHTML(provider.endpoint || "")}</code>
    <p>${escapeHTML(provider.description || "")}</p>
    <div class="pill-list">
      ${(provider.supported_events || []).map((event) => `<span>${escapeHTML(event)}</span>`).join("")}
      <span>${provider.secret_configured ? "secret 已设置" : "无 secret"}</span>
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
  setText("manage-inbox-count", `${pending} 个待审`);
  setText("home-inbox-count", pending);
  renderManageCount();
  renderPromptSuggestionDock();
  renderApprovalCenter();
  renderReviewQueue();
  setText("inbox-summary", state.inboxItems.length ? `${pending} 待审 · ${failed} 失败 · ${completed} 完成` : "还没有进入的渠道任务。");
  const overview = $("inbox-overview");
  if (overview) {
    overview.innerHTML = `<strong>${escapeHTML(pending ? `${pending} 个等待` : "已设闸门")}</strong><span>${escapeHTML(pending ? "渠道任务转成 Astria 运行前需要先复查。" : "进入项在确认前不会自动执行。")}</span>`;
  }
  if (!state.inboxItems.length) {
    renderEmptyAction(list, "还没有进入的渠道任务。可用本地 webhook 收件模拟一条。", [
      { label: "打开收件", panel: "inbox", primary: true },
      { label: "打开对话", panel: "chat" },
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
      <span>${escapeHTML(item.agent || "默认 Agent")}</span>
      ${item.run_id ? `<span>run ${escapeHTML(item.run_id)}</span>` : "<span>需要确认</span>"}
    </div>
    <div class="row-actions">${inboxActionsHTML(item)}</div>
  </article>`).join("");
}

function inboxActionsHTML(item) {
  const id = escapeHTML(item.id || "");
  switch (item.status) {
    case "pending":
      return `<button type="button" class="primary-button" data-inbox-approve="${id}">确认</button><button type="button" data-inbox-reject="${id}">拒绝</button>`;
    case "failed":
      return `<button type="button" class="primary-button" data-inbox-retry="${id}">重试</button><button type="button" data-inbox-reject="${id}">拒绝</button>`;
    case "completed":
      return item.run_id ? `<button type="button" data-inbox-run="${escapeHTML(item.run_id)}">观测运行</button>` : "";
    case "running":
      return `<button type="button" disabled>运行中</button>`;
    case "rejected":
      return `<button type="button" disabled>已拒绝</button>`;
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
    $("inbox-state").textContent = data.duplicate ? "重复" : "已接收";
    if (!data.duplicate) {
      $("inbox-external-id").value = "";
      $("inbox-sender").value = "";
      $("inbox-text").value = "";
    }
    await loadInbox();
    showToast(data.duplicate ? "重复 webhook 已忽略。" : "收件任务已接收。");
  } catch (error) {
    $("inbox-state").textContent = "错误";
    showToast(error.message);
  }
}

async function approveInboxItem(id) {
  await runInboxAction(id, "approve", "确认中", "收件项已确认。");
}

async function rejectInboxItem(id) {
  await runInboxAction(id, "reject", "拒绝中", "收件项已拒绝。");
}

async function retryInboxItem(id) {
  await runInboxAction(id, "retry", "重试中", "收件项已重试。");
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
    $("inbox-state").textContent = "就绪";
    await Promise.allSettled([loadInbox(), loadRuns()]);
    showToast(doneText);
  } catch (error) {
    $("inbox-state").textContent = "错误";
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
  $("permissions-save-state").textContent = "已加载";
  renderPermissionsPendingPreview();
}

function renderPermissions() {
  const permissions = state.permissions || {};
  const configured = permissions.configured === true;
  setText("settings-permissions-state", configured ? "已配置" : "默认");
  setClass("settings-permissions-state", configured ? "ready" : "warning");
  renderSystemStatusBoard();
  $("permissions-overview").innerHTML = `<strong>${configured ? "已配置" : "内置默认"}</strong><span>${configured ? "已发现显式权限配置。" : "当前没有显式权限配置。"}</span>`;
  const categories = [
    ["允许目录", permissions.allowed_dirs],
    ["允许命令", permissions.allowed_commands],
    ["拒绝命令", permissions.denied_commands],
    ["网络白名单", permissions.network_allowlist],
    ["敏感规则", permissions.sensitive_patterns],
  ];
  $("permissions-list").innerHTML = categories.map(([label, values]) => {
    const items = Array.isArray(values) ? values : [];
    return `<article class="row-item permission-item">
      <div class="row-item-title"><span>${escapeHTML(label)}</span><span class="tag">${items.length}</span></div>
      ${items.length ? `<div class="pill-list">${items.map((item) => `<span>${escapeHTML(item)}</span>`).join("")}</div>` : `<p>没有显式条目。</p>`}
    </article>`;
  }).join("");
  renderPermissionsPendingPreview();
}

function renderVersion() {
  const info = state.version || {};
  const check = state.updateCheck;
  const supported = info.update_supported === true;
  $("version-summary").textContent = info.message || "构建和更新状态。";
  $("update-check-button").disabled = !supported;
  if (!check) {
    $("update-check-state").textContent = supported ? "就绪" : "不可用";
    $("update-overview").innerHTML = `<strong>${escapeHTML(supported ? "就绪" : "开发构建")}</strong><span>${escapeHTML(info.message || "")}</span>`;
  } else {
    const label = updateStatusLabel(check.status);
    $("update-check-state").textContent = label;
    $("update-overview").innerHTML = `<strong>${escapeHTML(label)}</strong><span>${escapeHTML(check.message || "")}</span>`;
  }
  const updateRows = [
    ["版本", info.version || "-"],
    ["平台", info.platform || "-"],
    ["Web UI", info.web_url || "-"],
    ["启动命令", info.launch_command || "starclaw app"],
    ["更新检查", supported ? "支持" : "需要发布构建"],
    ["命令", info.update_command || "starclaw update --check"],
  ];
  if (check?.latest_version) updateRows.push(["最新版本", check.latest_version]);
  if (check?.release_url) updateRows.push(["Release URL", check.release_url]);
  const runtimeRows = [
    ["Web UI", info.web_url || "-"],
    ["健康检查", info.health_url || "-"],
    ["Status API", info.status_url || "-"],
    ["诊断", info.diagnostics_url || "-"],
    ["数据目录", info.starclaw_dir || "-"],
    ["配置", info.config_path || "-"],
  ];
  const readinessRows = [
    ["构建", supported ? "发布构建" : "开发构建"],
    ["更新", supported ? "可检查更新" : "需要发布构建"],
    ["启动命令", info.launch_command || "starclaw app"],
    ["Web UI", info.web_url || "-"],
  ];
  $("version-list").innerHTML = `<article class="row-item version-readiness-card">
    <div class="row-item-title"><span>发布就绪</span><span class="tag">${escapeHTML(supported ? "就绪" : "开发")}</span></div>
    <div class="run-meta-grid">
      ${readinessRows.map(([label, value]) => `<span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong>`).join("")}
    </div>
    ${supported ? `<p>发布元数据可用于更新检查。</p>` : `<p>使用 semver 发布构建后可启用更新检查。</p>`}
  </article>
  <article class="row-item version-card">
    <div class="row-item-title"><span>运行时上下文</span><span class="tag">本地</span></div>
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
      return "有可用更新";
    case "current":
      return "已是最新";
    case "development":
      return "开发构建";
    default:
      return "未知";
  }
}

function supportInfoText() {
  const info = state.version || {};
  const diagnostics = state.diagnostics || {};
  const rows = [
    ["Astria 支持信息", ""],
    ["Version", info.version || "-"],
    ["Platform", info.platform || "-"],
    ["Build status", info.status || "-"],
    ["Update supported", info.update_supported === true ? "是" : "否"],
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
  await copyText(supportInfoText(), "支持信息已复制。");
}

async function checkForUpdates() {
  if (state.version && state.version.update_supported !== true) {
    showToast("更新检查需要发布构建。");
    return;
  }
  $("update-check-button").disabled = true;
  $("update-check-state").textContent = "检查中";
  try {
    state.updateCheck = await api("/update/check");
    renderVersion();
    showToast(state.updateCheck.message || "更新检查完成。");
  } catch (error) {
    $("update-check-state").textContent = "错误";
    $("update-overview").innerHTML = `<strong>错误</strong><span>${escapeHTML(error.message)}</span>`;
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
    hints.push("允许范围较宽的本地访问。");
  }
  if (!deniedCommands.length) {
    hints.push("尚未配置拒绝命令。");
  }
  if (!sensitivePatterns.length) {
    hints.push("尚未配置敏感文件规则。");
  }
  if (networkAllowlist.some((host) => ["*", "*.*"].includes(host.trim()))) {
    hints.push("网络白名单包含较宽泛通配符。");
  }
  return hints;
}

function renderPermissionsPendingPreview() {
  const target = $("permissions-pending-preview");
  if (!target) return;
  const permissions = buildPermissionsPayload();
  const categories = [
    ["允许目录", permissions.allowed_dirs],
    ["允许命令", permissions.allowed_commands],
    ["拒绝命令", permissions.denied_commands],
    ["网络白名单", permissions.network_allowlist],
    ["敏感规则", permissions.sensitive_patterns],
  ];
  const hints = permissionsRiskHints(permissions);
  target.innerHTML = `<article class="row-item permission-preview">
    <div class="row-item-title"><span>待保存改动</span><span class="tag">${hints.length ? "需复查" : "就绪"}</span></div>
    <div class="permission-preview-grid">
      ${categories.map(([label, values]) => `<div class="permission-preview-count"><strong>${Array.isArray(values) ? values.length : 0}</strong><span>${escapeHTML(label)}</span></div>`).join("")}
    </div>
    ${hints.length ? `<div class="pill-list permission-risk-list">${hints.map((hint) => `<span>${escapeHTML(hint)}</span>`).join("")}</div>` : `<p>待保存改动中未发现明显权限风险。</p>`}
  </article>`;
}

async function submitPermissions(event) {
  event.preventDefault();
  $("permissions-save-state").textContent = "保存中";
  try {
    await api("/config", {
      method: "PATCH",
      body: JSON.stringify({ permissions: buildPermissionsPayload() }),
    });
    await Promise.allSettled([loadPermissions(), loadDiagnostics()]);
    $("permissions-save-state").textContent = "已保存";
    showToast("权限已保存。");
  } catch (error) {
    $("permissions-save-state").textContent = "错误";
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
    setText("manage-agents-count", `${state.agents.length} 个配置`);
    setText("nav-agents-count", state.agents.length);
    renderManageCount();
    renderHomeDockedTools();
    updateAgentSelects();
    renderAgentContinuityDigest();
    renderAgentCapabilityRoster();
    renderContextReadinessBoard();
    renderComparisonWorkbench();
    renderPromptExperimentLab();
    renderReuseGallery();
    renderMissionGraph();
    if (!state.agents.length) {
      renderEmpty(list, "还没有命名 Agent。");
      return;
    }
    list.innerHTML = state.agents.map((agent) => {
      const name = normalizeName(agent);
      const description = normalizeDescription(agent) || "没有描述。";
      return `<article class="row-item">
        <div class="row-item-title"><span>${escapeHTML(name)}</span><span class="tag">agent</span></div>
        <p>${escapeHTML(description)}</p>
        <div class="row-actions"><button data-agent-detail="${escapeHTML(name)}">编辑</button></div>
      </article>`;
    }).join("");
  } catch (error) {
    renderError(list, error.message);
    if ($("agent-continuity-digest")) renderError($("agent-continuity-digest"), error.message);
    if (roster) renderError(roster, error.message);
    renderComparisonWorkbench();
    renderReuseGallery();
    renderMissionGraph();
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
    description: normalizeDescription(agent) || "没有描述。",
    model: agent.Model || agent.model || modelCfg.Model || modelCfg.model || "默认",
    reasoning: agent.ReasoningEffort || agent.reasoning_effort || modelCfg.ReasoningEffort || modelCfg.reasoning_effort || "默认",
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
  if (!runs.length) return "还没有记录运行。先从一次聚焦测试或首个任务开始。";
  if (!summary.hasMemory) return "配置记忆为空。复杂任务前先捕获长期角色上下文。";
  if (runMissionGroup(latest) === "attention") return "继续使用此 Agent 前，需要先复查最近运行。";
  if (!summary.commandCount) return "还没有自定义命令。可以为此 Agent 添加可重复 Prompt。";
  return "可基于最近工作、既有记忆和命令继续。";
}

function renderAgentContinuityDigest() {
  const target = $("agent-continuity-digest");
  if (!target) return;
  if (!state.agents.length) {
    renderEmpty(target, "还没有 Agent 连续性可汇总。");
    return;
  }
  target.innerHTML = state.agents.map((agent) => {
    const summary = agentCapabilitySummary(agent);
    const runs = runsForAgent(summary.name);
    const latest = runs[0] || null;
    const latestStatus = latest ? (latest.status || "unknown") : "无";
    const latestPrompt = latest ? (runPrompt(latest) || latest.id || "最近运行") : "没有运行记录";
    const latestAction = latest ? `<button type="button" data-agent-open-run="${escapeHTML(latest.id)}">打开最近运行</button>` : "";
    const memoryLabel = summary.hasMemory ? "配置记忆" : "记忆缺口";
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
        <div><span>运行</span><strong>${runs.length}</strong></div>
        <div><span>记忆</span><strong>${escapeHTML(memoryLabel)}</strong></div>
        <div><span>命令</span><strong>${summary.commandCount}</strong></div>
      </div>
      <p>${escapeHTML(hint)}</p>
      <small>${escapeHTML(latestPrompt)}</small>
      <div class="row-actions">
        <button type="button" data-agent-continue="${escapeHTML(summary.name)}">继续</button>
        <button type="button" data-agent-memory-draft="${escapeHTML(summary.name)}">起草记忆</button>
        ${latestAction}
      </div>
    </article>`;
  }).join("");
}

function renderAgentCapabilityRoster() {
  const roster = $("agent-capability-roster");
  if (!roster) return;
  if (!state.agents.length) {
    renderEmpty(roster, "还没有 Agent 能力可映射。");
    return;
  }
  roster.innerHTML = state.agents.map((agent) => {
    const summary = agentCapabilitySummary(agent);
    const heartbeat = summary.heartbeatEvery
      ? `${summary.heartbeatEvery}${summary.heartbeatHours ? ` · ${summary.heartbeatHours}` : ""}`
      : "关闭";
    const posture = summary.autoApprove ? "自动确认" : "人工复查";
    const memory = summary.hasMemory ? "有记忆" : "无记忆";
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
        <div><span>模型</span><strong>${escapeHTML(summary.model)}</strong></div>
        <div><span>推理</span><strong>${escapeHTML(summary.reasoning)}</strong></div>
        <div><span>允许</span><strong>${summary.allow.length}</strong></div>
        <div><span>拒绝</span><strong>${summary.deny.length}</strong></div>
        <div><span>心跳</span><strong>${escapeHTML(heartbeat)}</strong></div>
        <div><span>命令</span><strong>${summary.commandCount}</strong></div>
      </div>
      <div class="pill-list agent-roster-tags">
        <span>${escapeHTML(memory)}</span>
        <span>${summary.autoApprove ? "绕过确认" : "确认闸门"}</span>
        <span>${summary.heartbeatEvery ? "已安排心跳" : "无心跳"}</span>
      </div>
      ${commandLaunchers}
      <div class="row-actions">
        <button type="button" data-agent-launch-chat="${escapeHTML(summary.name)}">对话</button>
        <button type="button" data-agent-launch-test="${escapeHTML(summary.name)}">测试</button>
        <button type="button" data-agent-launch-council="${escapeHTML(summary.name)}">议会</button>
        <button type="button" data-agent-detail="${escapeHTML(summary.name)}">编辑配置</button>
      </div>
    </article>`;
  }).join("");
}

function updateAgentSelects() {
  const options = ['<option value="">默认 Agent</option>'].concat(
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
  const base = state.editingAgent ? `正在编辑 ${state.editingAgent}` : "新 Agent";
  $("agent-form-state").textContent = state.agentDirty ? `${base} · 未保存改动` : base;
  renderAgentPermissionPreview();
}

function confirmDiscardAgentChanges() {
  return !state.agentDirty || globalThis.confirm("丢弃未保存的 Agent 改动？");
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
  $("selected-agent-description").textContent = state.editingAgent ? "正在编辑命名 Agent。" : "创建命名 Agent。";
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
  const allow = payload.tools_allow.length ? payload.tools_allow.join(", ") : "无";
  const deny = payload.tools_deny.length ? payload.tools_deny.join(", ") : "无";
  const allowSet = new Set(payload.tools_allow);
  const conflicts = payload.tools_deny.filter((tool) => allowSet.has(tool));
  const warnings = [];
  if (payload.auto_approve) warnings.push("此 Agent 已启用自动确认。");
  if (conflicts.length) warnings.push(`允许/拒绝冲突：${conflicts.join(", ")}`);
  $("agent-permission-preview").innerHTML = `<div class="agent-preview-row"><strong>允许</strong><span>${escapeHTML(allow)}</span></div>
    <div class="agent-preview-row"><strong>拒绝</strong><span>${escapeHTML(deny)}</span></div>
    <div class="agent-preview-row"><strong>自动确认</strong><span>${payload.auto_approve ? "启用" : "停用"}</span></div>
    ${warnings.map((warning) => `<div class="agent-preview-row warning"><strong>复查</strong><span>${escapeHTML(warning)}</span></div>`).join("")}`;
}

function renderAgentCommands() {
  const list = $("agent-command-list");
  const names = Object.keys(state.agentCommands).sort((a, b) => a.localeCompare(b));
  if (!names.length) {
    renderEmpty(list, "还没有自定义命令。");
    return;
  }
  list.innerHTML = names.map((name) => {
    const active = name === state.selectedAgentCommand ? " active" : "";
    return `<div class="row-item${active}">
      <div class="row-item-title"><span>${escapeHTML(name)}</span><span class="tag">command</span></div>
      <div class="row-actions"><button type="button" data-agent-command="${escapeHTML(name)}">编辑</button></div>
    </div>`;
  }).join("");
}

function clearAgentCommandEditor() {
  state.selectedAgentCommand = "";
  $("agent-command-name").disabled = false;
  $("agent-command-name").value = "";
  $("agent-command-body").value = "";
  $("agent-command-save-button").textContent = "添加命令";
  $("agent-command-delete-button").hidden = true;
}

function selectAgentCommand(name) {
  state.selectedAgentCommand = name;
  $("agent-command-name").disabled = false;
  $("agent-command-name").value = name;
  $("agent-command-body").value = state.agentCommands[name] || "";
  $("agent-command-save-button").textContent = "更新命令";
  $("agent-command-delete-button").hidden = false;
  renderAgentCommands();
}

function saveAgentCommand() {
  const name = $("agent-command-name").value.trim();
  const body = $("agent-command-body").value.trim();
  if (!name || !body) {
    showToast("需要填写命令名称和内容。");
    return;
  }
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$/.test(name)) {
    showToast("命令名称只能使用字母、数字、短横线或下划线。");
    return;
  }
  if (state.selectedAgentCommand && state.selectedAgentCommand !== name) {
    delete state.agentCommands[state.selectedAgentCommand];
  }
  state.agentCommands[name] = body;
  selectAgentCommand(name);
  updateAgentDirtyState();
  showToast("命令已暂存。");
}

function deleteAgentCommand() {
  const name = state.selectedAgentCommand;
  if (!name) return;
  delete state.agentCommands[name];
  clearAgentCommandEditor();
  renderAgentCommands();
  updateAgentDirtyState();
  showToast("命令已移除。");
}

function testCurrentAgent() {
  if (!confirmDiscardAgentChanges()) return;
  const name = state.editingAgent;
  if (!name) {
    showToast("测试前请先保存 Agent。");
    return;
  }
  $("agent-test-agent").value = name;
  $("agent-test-prompt").value = `Test ${name}: introduce your role and list one useful next step.`;
  $("agent-test-prompt").focus();
  $("agent-test-state").textContent = `准备测试 ${name}`;
  showToast(`已准备测试 ${name}。`);
}

function prepareAgentChat(name) {
  if (!name) return;
  startNewChat();
  $("chat-agent").value = name;
  $("chat-input").value = `Continue as ${name}. Review the current Astria workspace context, identify the next useful action, and call out any risks before changing files.`;
  $("chat-input").focus();
  showToast(`已为 ${name} 起草对话。`);
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
  showToast(`已为 ${name} 起草连续性 Prompt。`);
}

function prepareAgentTest(name) {
  if (!name) return;
  if (!confirmDiscardAgentChanges()) return;
  $("agent-test-agent").value = name;
  $("agent-test-prompt").value = `Test ${name}: introduce your operating role, summarize your configured strengths, and propose one concrete next step.`;
  $("agent-test-state").textContent = `准备测试 ${name}`;
  switchPanel("agents");
  $("agent-test-prompt").focus();
  showToast(`已为 ${name} 起草测试。`);
}

function prepareAgentCouncil(name) {
  if (!name) return;
  $("council-agent").value = name;
  $("council-goal").value = `Use ${name} as the lead agent. Split the current Astria task into planner, researcher, and reviewer perspectives, then synthesize a concrete next action.`;
  $("council-state").textContent = `已选择 ${name}`;
  switchPanel("council");
  $("council-goal").focus();
  showToast(`已为 ${name} 起草议会目标。`);
}

async function launchAgentCommand(agentName, commandName) {
  if (!agentName || !commandName) return;
  try {
    const detail = await api(`/agents/${encodeURIComponent(agentName)}`);
    const commands = detail.Commands || detail.commands || {};
    const body = commands[commandName] || "";
    if (!body.trim()) {
      showToast(`命令 /${commandName} 为空。`);
      return;
    }
    startNewChat();
    $("chat-agent").value = agentName;
    $("chat-input").value = body.trim();
    $("chat-input").focus();
    showToast(`已为 ${agentName} 起草 /${commandName}。`);
  } catch (error) {
    showToast(error.message);
  }
}

async function submitAgent(event) {
  event.preventDefault();
  const payload = buildAgentPayload();
  if (!payload.name || !payload.prompt.trim()) {
    showToast("需要填写 Agent 名称和 Prompt。");
    return;
  }
  const path = state.editingAgent ? `/agents/${encodeURIComponent(state.editingAgent)}` : "/agents";
  const method = state.editingAgent ? "PUT" : "POST";
  $("agent-form-state").textContent = "保存中";
  try {
    const saved = await api(path, { method, body: JSON.stringify(payload) });
    await loadAgents();
    fillAgentForm(saved);
    updateAgentSelects();
    showToast("Agent 已保存。");
  } catch (error) {
    $("agent-form-state").textContent = "错误";
    showToast(error.message);
  }
}

async function deleteCurrentAgent() {
  if (!confirmDiscardAgentChanges()) return;
  const name = state.editingAgent;
  if (!name || !globalThis.confirm(`删除 Agent "${name}"？`)) return;
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
    showToast("Agent 已删除。");
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
  showToast("Agent 配置已导出。");
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
    showToast("Agent 配置已导入。保存后生效。");
  } catch (error) {
    showToast(`导入失败：${error.message}`);
  } finally {
    $("agent-import-file").value = "";
  }
}

function updateSelectedAgent() {
  const name = $("chat-agent").value;
  const selected = state.agents.find((agent) => normalizeName(agent) === name);
  $("selected-agent-description").textContent = selected
    ? (normalizeDescription(selected) || "没有描述。")
    : "选择一个 Agent。";
}

function setAgentTestRunning(isRunning) {
  $("agent-test-agent").disabled = isRunning;
  $("agent-test-prompt").disabled = isRunning;
  $("agent-test-submit-button").hidden = isRunning;
  $("agent-test-stop-button").hidden = !isRunning;
  if (isRunning) $("agent-test-state").textContent = "运行中";
}

async function submitAgentTest(event) {
  event?.preventDefault();
  if (state.activeAgentTestRequestID) {
    showToast("已有 Agent 测试正在运行。");
    return;
  }
  const text = $("agent-test-prompt").value.trim();
  if (!text) {
    showToast("请先输入测试 Prompt。");
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
    $("agent-test-state").textContent = "完成";
  } catch (error) {
    if (error.name === "AbortError") {
      $("agent-test-state").textContent = "已取消";
      renderAgentTestCancelled(payload);
    } else {
      $("agent-test-state").textContent = "错误";
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
    <div class="run-summary-title">正在流式测试 Agent</div>
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
    <div class="run-summary-title">Agent 测试已取消</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "默认")}</strong>
      <span>Run ID</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
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
    : "没有返回消息。";
  const errorHTML = result?.error ? `<div class="error-state">${escapeHTML(result.error)}</div>` : "";
  const openRunAction = payload.request_id
    ? `<button type="button" data-run-summary-run="${escapeHTML(payload.request_id)}">观测运行</button>`
    : "";
  const openSessionAction = sessionID
    ? `<button type="button" data-run-summary-session="${escapeHTML(sessionID)}">打开会话</button>`
    : "";
  const summaryText = agentTestSummaryText(result, payload);
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-result">
    <div class="run-summary-title">Agent 测试结果</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "默认")}</strong>
      <span>Prompt</span><strong>${escapeHTML(payload.text || "-")}</strong>
      <span>会话</span><strong>${escapeHTML(sessionID || "-")}</strong>
      <span>用量</span><strong>${escapeHTML(usageText)}</strong>
      <span>Run ID</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
    </div>
    ${errorHTML}
    <pre>${escapeHTML(messages)}</pre>
    <div class="run-summary-actions">
      ${openRunAction}
      ${openSessionAction}
      <button type="button" data-agent-test-copy-summary="${escapeHTML(summaryText)}">复制摘要</button>
    </div>
  </div>`;
}

function renderAgentTestError(error, payload) {
  const summaryText = agentTestSummaryText({ error: error.message || String(error) }, payload);
  $("agent-test-output").innerHTML = `<div class="run-summary agent-test-result">
    <div class="run-summary-title">Agent 测试错误</div>
    <div class="run-summary-grid">
      <span>Agent</span><strong>${escapeHTML(payload.agent || "默认")}</strong>
      <span>Prompt</span><strong>${escapeHTML(payload.text || "-")}</strong>
      <span>Run ID</span><strong>${escapeHTML(payload.request_id || "-")}</strong>
    </div>
    <div class="error-state">${escapeHTML(error.message || String(error))}</div>
    <div class="run-summary-actions">
      <button type="button" data-agent-test-copy-summary="${escapeHTML(summaryText)}">复制摘要</button>
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
    "Agent 测试",
    `Agent: ${payload.agent || "默认"}`,
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
    setText("manage-skills-count", `${state.skills.length} 个已安装`);
    setText("nav-skills-count", state.skills.length);
    renderManageCount();
    renderHomeDockedTools();
    renderMissionGraph();
    renderContextReadinessBoard();
    if (!state.skills.length) {
      renderEmpty(list, "还没有安装技能。");
      return;
    }
    list.innerHTML = state.skills.map((skill) => `<article class="row-item">
      <div class="row-item-title"><span>${escapeHTML(skill.name)}</span><span class="tag">${escapeHTML(skill.source || "skill")}</span></div>
      <p>${escapeHTML(skill.description || "没有描述。")}</p>
    </article>`).join("");
  } catch (error) {
    renderError(list, error.message);
    renderMissionGraph();
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
      renderEmpty(list, query ? "没有匹配会话。" : "还没有保存会话。");
      renderMemoryMapPreview();
      renderSourceRegistry();
      renderWorkspaceHub();
      renderKnowledgeCuration();
      renderPromptSuggestionDock();
      renderFocusBrief();
      renderComparisonWorkbench();
      renderRunQualityScorecard();
      renderReuseGallery();
      renderResultLibrary();
      renderWorkspaceSnapshotPlanner();
      renderMissionGraph();
      return;
    }
    list.innerHTML = state.sessions.map((session) => `<article class="row-item session-item ${session.id === state.activeSessionID ? "active" : ""}" data-session-id="${escapeHTML(session.id)}">
      <div class="row-item-title">
        <span>${session.favorite ? "★ " : ""}${escapeHTML(session.title || session.id)}</span>
        <button class="icon-danger-button" type="button" title="删除会话" aria-label="删除会话" data-session-delete="${escapeHTML(session.id)}">删除</button>
      </div>
      <span class="tag">${session.msg_count ?? 0} 条消息</span>
      <p>${escapeHTML(session.id)}</p>
      <div class="row-actions">
        <button type="button" data-session-copy="${escapeHTML(session.id)}">复制 ID</button>
        <button type="button" data-session-rename="${escapeHTML(session.id)}">重命名</button>
        <button type="button" data-session-favorite="${escapeHTML(session.id)}" data-favorite="${session.favorite ? "false" : "true"}">${session.favorite ? "取消收藏" : "收藏"}</button>
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
    renderRunQualityScorecard();
    renderReuseGallery();
    renderResultLibrary();
    renderWorkspaceSnapshotPlanner();
    renderMissionGraph();
  } catch (error) {
    renderError(list, error.message);
    renderMemoryMapPreview();
    renderSourceRegistry();
    renderWorkspaceHub();
    renderKnowledgeCuration();
    renderPromptSuggestionDock();
    renderFocusBrief();
    renderComparisonWorkbench();
    renderRunQualityScorecard();
    renderReuseGallery();
    renderResultLibrary();
    renderWorkspaceSnapshotPlanner();
    renderMissionGraph();
  }
}

async function loadSchedules() {
  const list = $("schedules-list");
  try {
    const data = await api("/schedules");
    state.schedules = data.schedules || [];
    setText("manage-schedules-count", `${state.schedules.length} 个已配置`);
    setText("nav-schedules-count", state.schedules.length);
    renderManageCount();
    renderHomeDockedTools();
    renderProactiveDeliveryBoard();
    renderWorkspaceSnapshotPlanner();
    if (!state.schedules.length) {
      renderEmpty(list, "还没有配置定时任务。");
      return;
    }
    list.innerHTML = state.schedules.map((schedule) => `<article class="row-item">
      <div class="row-item-title">
        <span>${escapeHTML(schedule.prompt || schedule.id)}</span>
        <span class="tag">${schedule.enabled ? "已启用" : "已暂停"}</span>
      </div>
      <p>${escapeHTML(schedule.cron || "")} ${schedule.agent ? `使用 ${schedule.agent}` : "使用默认 Agent"}</p>
      <div class="row-actions">
        <button data-schedule-toggle="${escapeHTML(schedule.id)}" data-enabled="${schedule.enabled ? "false" : "true"}">${schedule.enabled ? "暂停" : "启用"}</button>
        <button data-schedule-delete="${escapeHTML(schedule.id)}">删除</button>
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
    renderRunDependentViews();
    if (!state.runs.length) {
      state.activeRunID = "";
      renderRunDetail(null);
    }
    if (state.activeRunID && !state.runs.some((run) => run.id === state.activeRunID)) {
      state.activeRunID = "";
      renderRunDetail(null);
    }
    hydrateHomeMissionRun();
  } catch (error) {
    state.runs = [];
    state.missionRunDetail = null;
    state.missionRunTrace = [];
    state.missionRunTraceError = "";
    $("runs-count").textContent = "0";
    renderRunDependentViews({ skipRunsList: true });
    renderError(list, error.message);
  }
}

function renderRunDependentViews(options = {}) {
  renderHomeActivity();
  renderMissionGraph();
  renderMemoryMapPreview();
  renderSourceRegistry();
  renderKnowledgeReconciliation();
  renderAgentContinuityDigest();
  if (!options.skipRunsList) {
    renderRunsList();
  }
  renderComparisonWorkbench();
  renderRunQualityScorecard();
  renderPromptExperimentLab();
  renderReuseGallery();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderStarterKitLauncher();
  renderSharePackBuilder();
  renderWorkspaceSnapshotPlanner();
  renderProactiveDeliveryBoard();
}

async function hydrateHomeMissionRun(options = {}) {
  const candidate = state.activeRunID
    ? state.runs.find((run) => run.id === state.activeRunID) || state.runs[0]
    : state.runs[0];
  const runID = candidate?.id || "";
  if (!runID || state.currentRunDetail?.id === runID || state.missionRunHydrating === runID) return;
  if (!options.force && state.missionRunDetail?.id === runID && Array.isArray(state.missionRunTrace)) {
    renderMissionGraph();
    return;
  }
  state.missionRunHydrating = runID;
  try {
    const encodedRunID = encodeURIComponent(runID);
    const [run, traceResult] = await Promise.all([
      api(`/runs/${encodedRunID}`),
      api(`/runs/${encodedRunID}/trace`).catch((error) => ({ trace: [], error: error.message })),
    ]);
    if ((run?.id || runID) !== runID) return;
    state.missionRunDetail = run || candidate;
    state.missionRunTrace = Array.isArray(traceResult.trace) ? traceResult.trace : [];
    state.missionRunTraceError = traceResult.error || "";
    renderMissionGraph();
    renderArtifactDependentViews();
  } catch {
    state.missionRunDetail = candidate || null;
    state.missionRunTrace = [];
    state.missionRunTraceError = "运行观测详情暂不可用。";
    renderMissionGraph();
    renderArtifactDependentViews();
  } finally {
    if (state.missionRunHydrating === runID) state.missionRunHydrating = "";
  }
}

function renderArtifactDependentViews() {
  renderRunQualityScorecard();
  renderReuseGallery();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderStarterKitLauncher();
  renderSharePackBuilder();
  renderWorkspaceSnapshotPlanner();
}

function handleRunLifecycleEvent(eventType, eventPayload) {
  const run = lifecycleRunSummary(eventType, eventPayload);
  if (!run) return;
  if (eventType === "run_started") {
    setStarMapActivity("running", { label: "event stream" });
  } else if (eventType === "run_completed") {
    setStarMapActivity("complete");
  } else if (eventType === "run_error") {
    setStarMapActivity("error");
  }
  upsertRecoveredRun(run);
  hydrateHomeMissionRun({ force: true });
  if (state.activeRunID && state.activeRunID === run.id) {
    selectRun(run.id);
  }
}

function lifecycleRunSummary(eventType, eventPayload) {
  const safe = safeLifecyclePayload(eventPayload);
  const id = safe.run_id || safe.id || "";
  if (!id) return null;
  const status = lifecycleRunStatus(eventType, safe.status);
  const run = {
    id,
    status,
    recovered: true,
    recovery_source: "event_stream",
  };
  for (const [sourceKey, targetKey] of [
    ["agent", "agent"],
    ["channel", "channel"],
    ["source", "source"],
    ["session_id", "session_id"],
    ["started_at", "started_at"],
    ["ended_at", "ended_at"],
    ["usage", "usage"],
  ]) {
    if (safe[sourceKey] !== undefined && safe[sourceKey] !== "") {
      run[targetKey] = safe[sourceKey];
    }
  }
  if (safe.error) {
    run.error = safe.error;
  }
  return run;
}

function lifecycleRunStatus(eventType, status) {
  if (status) return String(status);
  if (eventType === "run_started") return "running";
  if (eventType === "run_completed") return "completed";
  if (eventType === "run_error") return "error";
  return "unknown";
}

function safeLifecyclePayload(eventPayload) {
  if (!eventPayload || typeof eventPayload !== "object") return {};
  const allowedKeys = new Set(["run_id", "id", "status", "agent", "channel", "source", "session_id", "started_at", "ended_at", "usage", "error"]);
  return Object.fromEntries(Object.entries(eventPayload).filter(([key]) => allowedKeys.has(key)));
}

function upsertRecoveredRun(run) {
  const index = state.runs.findIndex((item) => item.id === run.id);
  if (index >= 0) {
    state.runs[index] = { ...state.runs[index], ...run };
  } else {
    state.runs = [run, ...state.runs];
  }
  $("runs-count").textContent = state.runs.length;
  renderRunDependentViews();
}

async function refreshRunsAfterEventStreamRecovery() {
  if (state.eventStream.refreshingRuns) return;
  state.eventStream.refreshingRuns = true;
  try {
    await loadRuns();
  } finally {
    state.eventStream.refreshingRuns = false;
  }
}

function renderRunsList() {
  const list = $("runs-list");
  renderMissionControl();
  if (!state.runs.length) {
    list.innerHTML = `<section class="run-empty-console">
      <div>
        <span>RUN OBSERVER</span>
        <strong>观测台待命</strong>
        <p>发起第一条任务后，这里会按时间记录运行、工具事件、审批、预算和产物线索。</p>
      </div>
      <div class="run-empty-grid">
        <button type="button" data-panel="home"><span>01</span><strong>回到任务台</strong><small>描述目标并创建第一条本地 run。</small></button>
        <button type="button" data-panel="diagnostics"><span>02</span><strong>检查运行环境</strong><small>确认 provider、权限和本地 daemon 状态。</small></button>
        <button type="button" data-panel="results"><span>03</span><strong>查看产物库</strong><small>运行完成后稳定输出会进入归档视图。</small></button>
      </div>
    </section>`;
    return;
  }
  const runs = filteredRuns();
  if (!runs.length) {
    renderEmpty(list, "没有匹配当前观测筛选的运行。");
    return;
  }
  list.innerHTML = runs.map((run) => {
    const active = run.id === state.activeRunID ? " active" : "";
    const agent = run.agent || "default";
    const session = run.session_id || "no session";
    const badges = runRuntimeBadges(run);
    return `<article class="row-item run-row${active}" data-run-id="${escapeHTML(run.id)}">
      <div class="row-item-title">
        <span>${escapeHTML(run.prompt || run.id)}</span>
        <span class="tag run-status ${escapeHTML(run.status || "unknown")}">${escapeHTML(run.status || "unknown")}</span>
      </div>
      <p>${escapeHTML(agent)} · ${escapeHTML(session)} · ${escapeHTML(formatTimestamp(run.started_at))}</p>
      ${badges.length ? `<div class="run-runtime-badges">${badges.map((badge) => `<span class="runtime-badge ${escapeHTML(badge.tone)}">${escapeHTML(badge.label)}</span>`).join("")}</div>` : ""}
      <div class="row-actions">
        <button type="button" data-run-open="${escapeHTML(run.id)}">观测运行</button>
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
    case "recovered":
      return state.runs.filter((run) => isRecoveredRun(run));
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
    recovered: state.runs.filter((run) => isRecoveredRun(run)).length,
    council: state.runs.filter((run) => run.channel === "council_handoff" || String(run.source || "").startsWith("council:")).length,
  };
  board.innerHTML = [
    ["active", "活跃轨道", counts.active, "正在运行或排队的工作"],
    ["recovered", "已恢复", counts.recovered, "从持久运行状态恢复"],
    ["attention", "待处理", counts.attention, "失败、取消或未知状态"],
    ["completed", "已完成", counts.completed, "形成产物的任务"],
    ["total", "总计", counts.total, "全部记录运行"],
  ].map(([key, label, value, hint]) => `<button type="button" class="mission-control-card ${escapeHTML(key)}" data-run-filter="${escapeHTML(key === "total" ? "all" : key)}">
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(String(value))}</strong>
      <small>${escapeHTML(hint)}</small>
    </button>`).join("");
  filters.innerHTML = [
    ["all", "全部", counts.total],
    ["active", "活跃", counts.active],
    ["recovered", "恢复", counts.recovered],
    ["attention", "待处理", counts.attention],
    ["completed", "完成", counts.completed],
    ["council", "议会", counts.council],
  ].map(([key, label, count]) => `<button type="button" class="${state.runFilter === key ? "active" : ""}" data-run-filter="${escapeHTML(key)}">${escapeHTML(label)} <span>${escapeHTML(String(count))}</span></button>`).join("");
}

async function selectRun(runID) {
  if (!runID) return;
  state.activeRunID = runID;
  state.currentRunTrace = [];
  state.currentRunTraceError = "";
  renderRunsList();
  switchPanel("runs");
  try {
    const encodedRunID = encodeURIComponent(runID);
    const [run, traceResult] = await Promise.all([
      api(`/runs/${encodedRunID}`),
      api(`/runs/${encodedRunID}/trace`).catch((error) => ({ trace: [], error: error.message })),
    ]);
    state.activeRunID = run.id || runID;
    state.currentRunTrace = Array.isArray(traceResult.trace) ? traceResult.trace : [];
    state.currentRunTraceError = traceResult.error || "";
    renderRunsList();
    renderRunDetail(run);
  } catch (error) {
    $("run-detail-summary").textContent = "运行详情不可用。";
    renderError($("run-detail"), error.message);
  }
}

function renderRunDetail(run) {
  const target = $("run-detail");
  state.currentRunDetail = run || null;
  renderMissionGraph();
  if (!run) {
    $("run-detail-summary").textContent = "选择一条运行，检查请求、结果和事件。";
    target.innerHTML = `<section class="run-detail-empty">
      <span>DETAIL CHANNEL</span>
      <strong>等待选择运行</strong>
      <p>选择左侧运行后，这里会展开 prompt 摘要、trace、工具事件、用量和复用动作。</p>
      <div>
        <small>Trace</small>
        <small>工具</small>
        <small>产物</small>
      </div>
    </section>`;
    return;
  }
  const usage = run.usage || run.response?.usage || {};
  const usageText = formatUsage(usage);
  const sessionID = runSessionID(run);
  const prompt = runPrompt(run);
  const observer = runObserverModel(run);
  $("run-detail-summary").textContent = `${run.status || "unknown"} · ${formatTimestamp(run.started_at)}`;
  target.innerHTML = `<div class="run-detail-stack">
    <section class="run-observer-card ${escapeHTML(observer.tone)}">
      <div class="run-observer-head">
        <div>
          <span>Run orbit</span>
          <strong>${escapeHTML(compactText(prompt || run.id || "当前运行", 54))}</strong>
          <small>${escapeHTML(observer.summary)}</small>
        </div>
        <b>${escapeHTML(observer.score)}%</b>
      </div>
      <div class="run-observer-meter" aria-label="运行观测完整度"><i style="width:${escapeHTML(String(observer.score))}%"></i></div>
      <div class="run-observer-grid">
        <button type="button" data-run-detail-rerun ${prompt ? "" : "disabled"}><span>状态</span><strong>${escapeHTML(uiTerm(run.status || "unknown"))}</strong></button>
        <button type="button" data-panel="agents"><span>Agent</span><strong>${escapeHTML(run.agent || "default")}</strong></button>
        <button type="button" data-panel="runs"><span>Trace</span><strong>${escapeHTML(String(observer.traceCount))}</strong></button>
        <button type="button" data-panel="mcp"><span>工具</span><strong>${escapeHTML(String(observer.toolCount))}</strong></button>
        <button type="button" data-panel="budget"><span>用量</span><strong>${escapeHTML(usageText)}</strong></button>
        <button type="button" data-panel="permissions"><span>审核</span><strong>${escapeHTML(observer.approvalLabel)}</strong></button>
      </div>
    </section>
    <section class="run-detail-section">
      <div class="run-meta-grid">
        <span>ID</span><strong>${escapeHTML(run.id || "-")}</strong>
        <span>Status</span><strong>${escapeHTML(run.status || "-")}</strong>
        <span>Agent</span><strong>${escapeHTML(run.agent || "default")}</strong>
        <span>Channel</span><strong>${escapeHTML(run.channel || "-")}</strong>
        <span>Session</span><strong>${escapeHTML(sessionID || "-")}</strong>
        <span>Started</span><strong>${escapeHTML(formatTimestamp(run.started_at))}</strong>
        <span>Ended</span><strong>${escapeHTML(formatTimestamp(run.ended_at))}</strong>
        <span>用量</span><strong>${escapeHTML(usageText)}</strong>
      </div>
      <div class="run-detail-actions">
        <button type="button" data-run-detail-copy-summary>复制摘要</button>
        <button type="button" data-run-detail-copy-prompt>复制 Prompt</button>
        <button type="button" data-run-detail-copy-result>复制结果</button>
        ${prompt ? `<button type="button" data-run-detail-follow-up>起草后续</button>` : ""}
        ${sessionID ? `<button type="button" data-run-detail-open-session="${escapeHTML(sessionID)}">打开会话</button>` : ""}
        ${prompt ? `<button type="button" data-run-detail-rerun>重新运行</button>` : ""}
      </div>
      ${run.error ? `<div class="error-state">${escapeHTML(run.error)}</div>` : ""}
    </section>
    <section class="run-detail-section">
      <h3>运行恢复</h3>
      ${renderRuntimeRecovery(run)}
    </section>
    <section class="run-detail-section">
      <h3>工作流阶段</h3>
      ${renderWorkflowSteps(run.steps || [])}
    </section>
    <section class="run-detail-section">
      <h3>控制记录</h3>
      ${renderControlHistory(run.control || [])}
    </section>
    <section class="run-detail-section">
      <h3>Trace</h3>
      ${renderRunTrace(state.currentRunTrace, state.currentRunTraceError)}
    </section>
    <section class="run-detail-section">
      <h3>Prompt</h3>
      <pre>${escapeHTML(prompt)}</pre>
    </section>
    <section class="run-detail-section">
      <h3>结果</h3>
      <pre>${escapeHTML(formatRunResponse(run.response))}</pre>
    </section>
    <section class="run-detail-section">
      <h3>观测时间线</h3>
      ${renderRunTimeline(run)}
    </section>
  </div>`;
}

function runObserverModel(run) {
  const status = runHealthGroup(run);
  const events = Array.isArray(run?.events) ? run.events : [];
  const control = Array.isArray(run?.control) ? run.control : [];
  const steps = Array.isArray(run?.steps) ? run.steps : [];
  const traceCount = state.currentRunTrace.length || Number(run?.trace_events || run?.structured_events?.length || 0);
  const toolCount = events.filter((event) => ["tool_call", "tool_result", "tool_status", "tool"].includes(event.type)).length;
  const approvalCount = control.filter((item) => item.status === "approval_required").length + steps.filter((step) => step.status === "waiting_approval").length;
  const usage = run?.usage || run?.response?.usage || {};
  let score = 24;
  if (run?.id) score += 12;
  if (traceCount) score += Math.min(18, traceCount * 3);
  if (toolCount) score += Math.min(14, toolCount * 3);
  if (formatUsage(usage) !== "-") score += 12;
  if (run?.response && !run?.response?.error) score += 12;
  if (approvalCount) score -= 6;
  if (status === "completed") score += 10;
  if (status === "failed") score -= 14;
  score = Math.max(8, Math.min(100, score));
  const summary = [
    `${run?.agent || "default"} · ${run?.channel || "local"}`,
    `${traceCount} trace`,
    `${toolCount} tool`,
    approvalCount ? `${approvalCount} 审核门` : "审核 clear",
  ].join(" · ");
  return {
    score,
    tone: status,
    traceCount,
    toolCount,
    approvalLabel: approvalCount ? `${approvalCount} 待确认` : "clear",
    summary,
  };
}

function renderRunTimeline(run) {
  const entries = buildRunTimelineEntries(run);
  if (!entries.length) return `<div class="empty-state">这条运行还没有时间线数据。</div>`;
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
    title: `运行${run.status ? `：${uiTerm(run.status)}` : "已记录"}`,
    detail: `${run.agent || "默认 Agent"} · ${run.channel || "本地"} · ${run.id || "run"}`,
  });
  if (prompt) {
    entries.push({
      kind: "milestone",
      tone: "prompt",
      at: run.started_at,
      title: "Prompt 已锁定",
      detail: "Prompt 可在明确的 Prompt 区域查看。",
    });
  }
  if (sessionID) {
    entries.push({
      kind: "milestone",
      tone: "session",
      at: run.started_at,
      title: "会话已关联",
      detail: sessionID,
      sessionID,
    });
  }
  entries.push(...groupRunTimelineEvents(run.events || [], run.structured_events || []));
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
      title: run.error || run.response?.error ? "运行需要复查" : "运行已完成",
      detail: run.error || run.response?.error || uiTerm(run.status || "completed"),
    });
  }
  return entries;
}

function renderRunEvents(events) {
  if (!events.length) return `<div class="empty-state">这条运行尚未捕获事件。</div>`;
  const entries = groupRunTimelineEvents(events);
  return `<div class="run-timeline">${entries.map(renderRunTimelineEntry).join("")}</div>`;
}

function structuredToolResultRedactions(structuredEvents) {
  const redactions = new Map();
  for (const event of structuredEvents || []) {
    if (event?.type !== "tool_result") continue;
    const data = event.data || {};
    const tool = data.tool || "tool";
    if (data.content_redacted === true) {
      redactions.set(tool, { content_redacted: true });
    }
  }
  return redactions;
}

function groupRunTimelineEvents(events, structuredEvents = []) {
  const entries = [];
  const openTools = new Map();
  const redactedToolResults = structuredToolResultRedactions(structuredEvents);
  for (const event of events) {
    const data = event.data || {};
    const tool = data.tool || "tool";
    if (event.type === "tool_call") {
      const entry = {
        kind: "tool",
        at: event.at,
        tool,
        status: data.status || "running",
        args: data.args ? safeRenderPayload({ args: data.args }) : null,
        result: null,
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
        args: null,
        result: null,
        isError: false,
        errorCategory: "",
      };
      if (!openTools.has(tool)) entries.push(entry);
      entry.status = data.status || (data.is_error ? "error" : "completed");
      entry.result = redactedToolResults.get(tool) || (Object.prototype.hasOwnProperty.call(data, "content")
        ? safeRenderPayload(data.content)
        : null);
      entry.isError = data.is_error === true;
      entry.errorCategory = data.error_category || "";
      openTools.delete(tool);
      continue;
    }
    entries.push({ kind: event.type || "event", at: event.at, data: safeRenderPayload(data) });
  }
  return entries;
}

function renderRunTimelineEntry(entry) {
  if (entry.kind === "milestone") {
    const action = entry.sessionID
      ? `<button type="button" data-run-detail-open-session="${escapeHTML(entry.sessionID)}">打开关联会话</button>`
      : "";
    return `<article class="run-event run-milestone ${escapeHTML(entry.tone || "")}">
      <div class="run-event-header">
        <strong>${escapeHTML(entry.title || "里程碑")}</strong>
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
      ? `<button type="button" data-run-tool-copy-result="${escapeHTML(resultText)}">复制结果</button>`
      : "";
    return `<article class="run-event run-tool-event ${entry.isError ? "bad" : ""}">
      <div class="run-event-header">
        <strong>${escapeHTML(entry.tool)}</strong>
        <span>${escapeHTML(uiTerm(status))} · ${escapeHTML(formatTimestamp(entry.at))}</span>
      </div>
      ${resultAction ? `<div class="run-event-actions">${resultAction}</div>` : ""}
      <div class="run-tool-grid">
        ${entry.args ? `<div><span>Args</span><pre>${escapeHTML(formatToolPayload(entry.args))}</pre></div>` : ""}
        ${resultText ? `<div><span>Result</span><pre>${escapeHTML(resultText)}</pre></div>` : ""}
        ${entry.errorCategory ? `<div class="tool-meta">分类：${escapeHTML(entry.errorCategory)}</div>` : ""}
      </div>
    </article>`;
  }
  const label = runEventLabel(entry.kind);
  return `<article class="run-event">
    <div class="run-event-header">
      <strong>${escapeHTML(label)}</strong>
      <span>${escapeHTML(formatTimestamp(entry.at))}</span>
    </div>
    <pre>${escapeHTML(formatToolPayload(safeRenderPayload(entry.data || {})))}</pre>
  </article>`;
}

function runEventLabel(type) {
  switch (type) {
    case "text":
      return "文本";
    case "preamble":
      return "前置说明";
    case "usage":
      return "用量";
    case "approval_needed":
      return "需要审批";
    case "approval_resolved":
      return "审批已处理";
    default:
      return type || "Event";
  }
}

function isRecoveredRun(run) {
  if (!run) return false;
  const steps = run.steps || [];
  const controls = run.control || [];
  const hasRecoveryStep = steps.some((step) => step.id === "runtime-recovery" || step.metadata?.recovered === true || step.metadata?.recovery_status);
  const hasRecoveryControl = controls.some((item) => item.action === "recover" || item.status === "recovered");
  return hasRecoveryStep || hasRecoveryControl || run.recovered === true;
}

function replayState(run) {
  const controls = run?.control || [];
  const steps = run?.steps || [];
  if (controls.some((item) => item.action === "replay" && item.status === "approval_required") || steps.some((step) => step.status === "waiting_approval")) {
    return "approval";
  }
  if (controls.some((item) => item.action === "replay" && item.status === "approved") || steps.some((step) => step.metadata?.replay_run_id)) {
    return "approved";
  }
  return "";
}

function pauseState(run) {
  const controls = run?.control || [];
  for (let index = controls.length - 1; index >= 0; index--) {
    const item = controls[index];
    if (item.action === "pause" || item.action === "resume" || item.action === "cancel") {
      return item.status || item.action;
    }
  }
  const step = (run?.steps || []).find((item) => item.id === "runtime-pause");
  return step?.metadata?.runtime_status || "";
}

function runRuntimeBadges(run) {
  const badges = [];
  if (isRecoveredRun(run)) badges.push({ label: "已恢复", tone: "recovered" });
  const replay = replayState(run);
  if (replay === "approval") badges.push({ label: "重放待审批", tone: "attention" });
  if (replay === "approved") badges.push({ label: "重放已批准", tone: "ok" });
  const pause = pauseState(run);
  if (pause) badges.push({ label: uiTerm(pause), tone: pause === "paused" ? "attention" : "neutral" });
  const traceCount = Number(run?.trace_events || run?.structured_events?.length || 0);
  if (traceCount > 0) badges.push({ label: `${traceCount} 条 Trace`, tone: "trace" });
  return badges;
}

function renderRuntimeRecovery(run) {
  const traceCount = state.currentRunTrace.length || Number(run?.trace_events || run?.structured_events?.length || 0);
  const items = [
    ["重启状态", isRecoveredRun(run) ? "已从持久存储恢复" : "当前 daemon 状态"],
    ["重放", replayStateLabel(replayState(run))],
    ["暂停 / 恢复", uiTerm(pauseState(run) || "无暂停边界")],
    ["工作流步骤", String((run?.steps || []).length)],
    ["控制决策", String((run?.control || []).length)],
    ["Trace 事件", String(traceCount)],
  ];
  return `<div class="runtime-recovery-grid">${items.map(([label, value]) => `
    <div>
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(value)}</strong>
    </div>`).join("")}</div>`;
}

function replayStateLabel(value) {
  switch (value) {
    case "approval":
      return "等待重放审批";
    case "approved":
      return "重放已批准或已发起";
    default:
      return "无重放边界";
  }
}

function renderWorkflowSteps(steps) {
  if (!steps.length) return `<div class="empty-state">尚未记录工作流步骤。</div>`;
  return `<div class="runtime-table">${steps.map((step) => `
    <article>
      <div>
        <strong>${escapeHTML(step.title || step.id || "工作流步骤")}</strong>
        <span>${escapeHTML(uiTerm(step.status || "unknown"))} · ${escapeHTML(formatTimestamp(step.updated_at))}</span>
      </div>
      ${step.metadata ? `<pre>${escapeHTML(formatToolPayload(safeRenderPayload(step.metadata)))}</pre>` : ""}
    </article>`).join("")}</div>`;
}

function renderControlHistory(control) {
  if (!control.length) return `<div class="empty-state">尚未记录控制决策。</div>`;
  return `<div class="runtime-table">${control.map((item) => `
    <article>
      <div>
        <strong>${escapeHTML(item.action || "控制")}</strong>
        <span>${escapeHTML(uiTerm(item.status || "unknown"))} · ${escapeHTML(formatTimestamp(item.at))}</span>
      </div>
      ${item.reason ? `<p>${escapeHTML(item.reason)}</p>` : ""}
    </article>`).join("")}</div>`;
}

function renderRunTrace(trace, error) {
  if (error) return `<div class="error-state">Trace 不可用：${escapeHTML(error)}</div>`;
  if (!trace.length) return `<div class="empty-state">尚未记录结构化 Trace 事件。</div>`;
  return `<div class="runtime-table trace-table">${trace.map((item) => `
    <article>
      <div>
        <strong>${escapeHTML(item.name || item.event_id || "trace_event")}</strong>
        <span>${escapeHTML(item.phase || "event")} · ${escapeHTML(formatTimestamp(item.timestamp))}</span>
      </div>
      <div class="trace-meta">
        <span>${escapeHTML(item.event_id || "-")}</span>
        <span>${escapeHTML(item.span_id || "-")}</span>
      </div>
      ${item.attributes ? `<pre>${escapeHTML(formatToolPayload(safeRenderPayload(item.attributes)))}</pre>` : ""}
    </article>`).join("")}</div>`;
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
    `Prompt: ${runPrompt(run) ? "[REDACTED: use Copy prompt for local operator review]" : "-"}`,
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
  if (!ensureProviderReadyForLaunch("对话")) return;
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
  startLiveRunStatus(payload);
  setRunControls(true);
  $("chat-input").value = "";
  try {
    const result = await streamMessage(payload, chatStreamRenderer(assistantMessage), abort.signal);
    renderDoneResult(result, assistantMessage);
    updateLiveRunStatus({
      state: "complete",
      sessionID: result?.session_id || state.liveRun.sessionID,
      usage: result?.usage || state.liveRun.usage,
      latest: "Run complete",
    });
    setStarMapActivity("complete");
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
      updateLiveRunStatus({ state: "cancelled", latest: "Cancel requested" });
      setStarMapActivity("error", { label: "route cancelled" });
    } else {
      appendMessage("error", error.message);
      stateLabel.textContent = "Error";
      updateLiveRunStatus({ state: "error", latest: error.message || "Stream error" });
      setStarMapActivity("error");
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
  const failed = data.is_error === true || data.status === "error";
  if (failed) {
    setStarMapActivity("error", { label: "tool route broken" });
  } else if (eventType === "tool_call") {
    setStarMapActivity("tool", { label: "tool link" });
  } else if (eventType === "tool_result" || eventType === "tool") {
    setStarMapActivity("artifact", { label: "artifact forming" });
  }
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
    content: data.content ?? data.preview ?? previous.content,
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
  setStarMapActivity("approval");
  renderHomeActivity();
  $("chat-state").textContent = "需要确认";
  scrollConversationToBottom();
}

function approvalCardHTML(data, status) {
  const args = data.args ? formatToolPayload(data.args) : "";
  const reason = data.reason || "需要确认";
  const statusLabel = status === "pending" ? "待确认" : status === "allowed" ? "已允许" : status === "denied" ? "已拒绝" : status;
  const disabled = status === "pending" ? "" : "disabled";
  return `<div class="approval-header">
    <span>需要人工确认</span>
    <strong>${escapeHTML(statusLabel)}</strong>
  </div>
  <div class="approval-body">
    <div><span>工具</span><strong>${escapeHTML(data.tool || "tool")}</strong></div>
    <div><span>原因</span><strong>${escapeHTML(reason)}</strong></div>
    ${data.agent ? `<div><span>Agent</span><strong>${escapeHTML(data.agent)}</strong></div>` : ""}
    ${args ? `<pre>${escapeHTML(args)}</pre>` : ""}
  </div>
  <div class="approval-actions">
    <button class="primary-button" data-approval-decision="allow" data-approval-id="${escapeHTML(data.request_id || "")}" ${disabled}>允许</button>
    <button class="danger-button" data-approval-decision="deny" data-approval-id="${escapeHTML(data.request_id || "")}" ${disabled}>拒绝</button>
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
  setStarMapActivity(status === "allowed" ? "complete" : "error", {
    label: status === "allowed" ? "gate cleared" : "gate denied",
  });
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

function safeRenderPayload(value) {
  if (Array.isArray(value)) return value.map((item) => safeRenderPayload(item));
  if (!value || typeof value !== "object") return secretLikeValue(value) ? "[REDACTED]" : value;
  return Object.fromEntries(Object.entries(value).map(([key, item]) => {
    if (unsafePayloadKey(key)) return [`${key}_redacted`, true];
    if (secretLikeValue(key)) return [key, "[REDACTED]"];
    return [key, safeRenderPayload(item)];
  }));
}

function unsafePayloadKey(key) {
  return ["args", "content", "text", "delta", "preamble", "prompt", "request", "response"].includes(String(key || ""));
}

function secretLikeValue(value) {
  return /api_key|token|secret|password|bearer\s+/i.test(String(value || ""));
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
    ? `<button type="button" data-run-summary-session="${escapeHTML(sessionID)}">打开会话</button>`
    : "";
  const openRunAction = requestID && requestID !== "-"
    ? `<button type="button" data-run-summary-run="${escapeHTML(requestID)}">观测运行</button>`
    : "";
  card.innerHTML = `<div class="run-summary-title">运行摘要</div>
    <div class="run-summary-grid">
      <span>Session</span><strong>${escapeHTML(sessionID || "-")}</strong>
      <span>Agent</span><strong>${escapeHTML(agent)}</strong>
      <span>用量</span><strong>${escapeHTML(usageText)}</strong>
      <span>Run ID</span><strong>${escapeHTML(requestID)}</strong>
    </div>
    <div class="run-summary-actions">
      <button type="button" data-run-summary-copy="${escapeHTML(summaryText)}">复制摘要</button>
      <button type="button" data-run-follow-up="${escapeHTML(runFollowUpPrompt(run))}">起草后续</button>
      ${openRunAction}
      ${openSessionAction}
    </div>`;
  $("chat-output").appendChild(card);
}

function chatStreamRenderer(assistantMessage) {
  return {
    appendText(text) {
      appendAssistantText(assistantMessage, text);
      updateLiveRunStatus({ state: "running", latest: "Streaming text" });
      setStarMapActivity("running", { label: "text stream" });
    },
    appendEvent(eventType, data) {
      if (eventType === "session_started") {
        updateLiveRunStatus({ sessionID: data.session_id || data.id || state.liveRun.sessionID, latest: "Session started" });
        setStarMapActivity("context", { label: "session linked" });
      } else if (eventType === "usage") {
        updateLiveRunStatus({ usage: data, latest: "Usage updated" });
        setStarMapActivity("context", { label: "usage observed" });
      } else if (eventType === "tool_call" || eventType === "tool_result" || eventType === "tool") {
        const label = data.tool || data.name || eventType;
        updateLiveRunStatus({ latest: `${eventType}: ${label}` });
      } else {
        updateLiveRunStatus({ latest: eventType || "Stream event" });
      }
      if (eventType === "tool_call" || eventType === "tool_result" || eventType === "tool") {
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
  const streamState = newStreamEventState();

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop() || "";
    for (const rawEvent of events) {
      doneResult = handleSSEEvent(rawEvent, renderer, doneResult, streamState);
    }
  }
  if (buffer.trim()) {
    doneResult = handleSSEEvent(buffer, renderer, doneResult, streamState);
  }
  return doneResult;
}

function newStreamEventState() {
  return {
    lastAssistantText: null,
    lastAssistantTextType: "",
  };
}

function shouldAppendStreamText(streamState, eventType, text) {
  if (!text) return false;
  const previous = streamState.lastAssistantText;
  const previousType = streamState.lastAssistantTextType;
  streamState.lastAssistantText = text;
  streamState.lastAssistantTextType = eventType;
  if (eventType === "delta" && previousType === "text" && previous === text) return false;
  if (eventType === "assistant_text" && previousType === "preamble" && previous === text) return false;
  return true;
}

function streamTextForEvent(eventType, data) {
  switch (eventType) {
    case "text":
      return data.text || "";
    case "delta":
      return data.text || data.delta || "";
    case "preamble":
      return data.preamble || "";
    case "assistant_text":
      return data.text || data.preamble || "";
    default:
      return "";
  }
}

function handleSSEEvent(rawEvent, renderer, doneResult, streamState = newStreamEventState()) {
  const parsed = parseSSE(rawEvent);
  if (!parsed) return doneResult;
  const data = parseEventData(parsed.data);
  switch (parsed.event) {
    case "text":
    case "delta":
    case "preamble":
    case "assistant_text": {
      const text = streamTextForEvent(parsed.event, data);
      if (shouldAppendStreamText(streamState, parsed.event, text)) {
        renderer.appendText?.(text);
      }
      break;
    }
    case "session_started":
    case "usage":
    case "tool":
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
  updateLiveRunStatus({ state: "cancelled", latest: "Cancelling run" });
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
  if (!ensureProviderReadyForLaunch("任务")) return;
  document.querySelectorAll(".home-disclosure[open]").forEach((disclosure) => {
    disclosure.open = false;
  });
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

function applyRailCollapsed(collapsed) {
  state.railCollapsed = Boolean(collapsed);
  const shell = document.querySelector(".shell");
  shell?.classList.toggle("rail-collapsed", state.railCollapsed);
  const button = $("rail-toggle-button");
  if (button) {
    button.setAttribute("aria-expanded", String(!state.railCollapsed));
    button.setAttribute("aria-label", state.railCollapsed ? "展开侧边栏" : "收起侧边栏");
    button.textContent = state.railCollapsed ? "›" : "‹";
  }
}

function toggleRailCollapsed() {
  applyRailCollapsed(!state.railCollapsed);
  try {
    localStorage.setItem("astriaRailCollapsed", state.railCollapsed ? "1" : "0");
  } catch {
    // Local preference only; ignore restricted storage.
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
  if (!globalThis.confirm(`删除会话 "${label}"？`)) return;
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

  const qualitySelect = event.target.closest("[data-quality-select]");
  if (qualitySelect) {
    state.selectedRunQuality = qualitySelect.dataset.qualitySelect || "";
    renderRunQualityScorecard();
    return;
  }

  const qualityDraft = event.target.closest("[data-quality-draft]");
  if (qualityDraft) {
    draftRunQualityToChat(qualityDraft.dataset.qualityDraft);
    return;
  }

  const runQuality = event.target.closest("[data-run-quality]");
  if (runQuality && !event.target.closest("button")) {
    state.selectedRunQuality = runQuality.dataset.runQuality || "";
    renderRunQualityScorecard();
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

  const budgetSelect = event.target.closest("[data-budget-select]");
  if (budgetSelect) {
    state.selectedBudgetGuard = budgetSelect.dataset.budgetSelect || "";
    renderBudgetGuardPlanner();
    return;
  }

  const budgetDraft = event.target.closest("[data-budget-draft]");
  if (budgetDraft) {
    draftBudgetGuardToChat(budgetDraft.dataset.budgetDraft);
    return;
  }

  const budgetGuard = event.target.closest("[data-budget-guard]");
  if (budgetGuard && !event.target.closest("button")) {
    state.selectedBudgetGuard = budgetGuard.dataset.budgetGuard || "";
    renderBudgetGuardPlanner();
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

  const resultSelect = event.target.closest("[data-result-select]");
  if (resultSelect) {
    state.selectedResultArchive = resultSelect.dataset.resultSelect || "";
    renderResultLibrary();
    return;
  }

  const resultDraft = event.target.closest("[data-result-draft]");
  if (resultDraft) {
    draftResultArchiveToChat(resultDraft.dataset.resultDraft);
    return;
  }

  const resultArchive = event.target.closest("[data-result-archive]");
  if (resultArchive && !event.target.closest("button")) {
    state.selectedResultArchive = resultArchive.dataset.resultArchive || "";
    renderResultLibrary();
    return;
  }

  const playbookSelect = event.target.closest("[data-playbook-select]");
  if (playbookSelect) {
    state.selectedPlaybook = playbookSelect.dataset.playbookSelect || "";
    renderPlaybookLibrary();
    return;
  }

  const playbookDraft = event.target.closest("[data-playbook-draft]");
  if (playbookDraft) {
    draftPlaybookToChat(playbookDraft.dataset.playbookDraft);
    return;
  }

  const playbook = event.target.closest("[data-playbook]");
  if (playbook && !event.target.closest("button")) {
    state.selectedPlaybook = playbook.dataset.playbook || "";
    renderPlaybookLibrary();
    return;
  }

  const starterSelect = event.target.closest("[data-starter-select]");
  if (starterSelect) {
    state.selectedStarterKit = starterSelect.dataset.starterSelect || "";
    renderStarterKitLauncher();
    return;
  }

  const starterDraft = event.target.closest("[data-starter-draft]");
  if (starterDraft) {
    draftStarterKitToChat(starterDraft.dataset.starterDraft);
    return;
  }

  const starterKit = event.target.closest("[data-starter-kit]");
  if (starterKit && !event.target.closest("button")) {
    state.selectedStarterKit = starterKit.dataset.starterKit || "";
    renderStarterKitLauncher();
    return;
  }

  const shareSelect = event.target.closest("[data-share-select]");
  if (shareSelect) {
    state.selectedSharePack = shareSelect.dataset.shareSelect || "";
    renderSharePackBuilder();
    return;
  }

  const shareDraft = event.target.closest("[data-share-draft]");
  if (shareDraft) {
    draftSharePackToChat(shareDraft.dataset.shareDraft);
    return;
  }

  const sharePack = event.target.closest("[data-share-pack]");
  if (sharePack && !event.target.closest("button")) {
    state.selectedSharePack = sharePack.dataset.sharePack || "";
    renderSharePackBuilder();
    return;
  }

  const snapshotSelect = event.target.closest("[data-snapshot-select]");
  if (snapshotSelect) {
    state.selectedWorkspaceSnapshot = snapshotSelect.dataset.snapshotSelect || "";
    renderWorkspaceSnapshotPlanner();
    return;
  }

  const snapshotDraft = event.target.closest("[data-snapshot-draft]");
  if (snapshotDraft) {
    draftWorkspaceSnapshotToChat(snapshotDraft.dataset.snapshotDraft);
    return;
  }

  const workspaceSnapshot = event.target.closest("[data-workspace-snapshot]");
  if (workspaceSnapshot && !event.target.closest("button")) {
    state.selectedWorkspaceSnapshot = workspaceSnapshot.dataset.workspaceSnapshot || "";
    renderWorkspaceSnapshotPlanner();
    return;
  }

  const browserSelect = event.target.closest("[data-browser-select]");
  if (browserSelect) {
    state.selectedBrowserMission = browserSelect.dataset.browserSelect || "";
    renderBrowserMissionPlanner();
    return;
  }

  const browserDraft = event.target.closest("[data-browser-draft]");
  if (browserDraft) {
    draftBrowserMissionToChat(browserDraft.dataset.browserDraft);
    return;
  }

  const browserMission = event.target.closest("[data-browser-mission]");
  if (browserMission && !event.target.closest("button")) {
    state.selectedBrowserMission = browserMission.dataset.browserMission || "";
    renderBrowserMissionPlanner();
    return;
  }

  const dataSelect = event.target.closest("[data-data-select]");
  if (dataSelect) {
    state.selectedDataInsight = dataSelect.dataset.dataSelect || "";
    renderDataInsightPlanner();
    return;
  }

  const dataDraft = event.target.closest("[data-data-draft]");
  if (dataDraft) {
    draftDataInsightToChat(dataDraft.dataset.dataDraft);
    return;
  }

  const dataInsight = event.target.closest("[data-data-insight]");
  if (dataInsight && !event.target.closest("button")) {
    state.selectedDataInsight = dataInsight.dataset.dataInsight || "";
    renderDataInsightPlanner();
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

  const reconcileSelect = event.target.closest("[data-reconcile-select]");
  if (reconcileSelect) {
    state.selectedReconcileRisk = reconcileSelect.dataset.reconcileSelect || "";
    renderKnowledgeReconciliation();
    return;
  }

  const reconcileDraft = event.target.closest("[data-reconcile-draft]");
  if (reconcileDraft) {
    draftReconciliationToChat(reconcileDraft.dataset.reconcileDraft);
    return;
  }

  const reconcileRisk = event.target.closest("[data-reconcile-risk]");
  if (reconcileRisk && !event.target.closest("button")) {
    state.selectedReconcileRisk = reconcileRisk.dataset.reconcileRisk || "";
    renderKnowledgeReconciliation();
    return;
  }

  const citationSelect = event.target.closest("[data-citation-select]");
  if (citationSelect) {
    state.selectedCitationGrounding = citationSelect.dataset.citationSelect || "";
    renderCitationGroundingPlanner();
    return;
  }

  const citationDraft = event.target.closest("[data-citation-draft]");
  if (citationDraft) {
    draftCitationGroundingToChat(citationDraft.dataset.citationDraft);
    return;
  }

  const citationGrounding = event.target.closest("[data-citation-grounding]");
  if (citationGrounding && !event.target.closest("button")) {
    state.selectedCitationGrounding = citationGrounding.dataset.citationGrounding || "";
    renderCitationGroundingPlanner();
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
    showToast("后续 Prompt 已起草。");
    return;
  }

  const agentTestCopySummary = event.target.closest("[data-agent-test-copy-summary]");
  if (agentTestCopySummary) {
    copyText(agentTestCopySummary.dataset.agentTestCopySummary, "Agent 测试摘要已复制。")
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
$("rail-toggle-button").addEventListener("click", toggleRailCollapsed);
$("command-center-input").addEventListener("input", renderCommandCenterList);
document.querySelectorAll(".home-disclosure").forEach((disclosure) => {
  disclosure.addEventListener("beforetoggle", (event) => {
    if (event.newState !== "open") return;
    document.querySelectorAll(".home-disclosure").forEach((other) => {
      if (other !== disclosure) other.open = false;
    });
  });
  disclosure.addEventListener("toggle", () => {
    if (!disclosure.open) return;
    document.querySelectorAll(".home-disclosure").forEach((other) => {
      if (other !== disclosure) other.open = false;
    });
  });
});
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
$("config-provider").addEventListener("change", () => {
  updateProviderFields();
  renderConnectorTestState({ status: "idle", detail: "有未保存的连接改动；保存后再检查连接。" });
});
$("config-test-button").addEventListener("click", testProviderConnection);
[
  "config-endpoint",
  "config-model-tier",
  "config-api-key",
  "config-openai-endpoint",
  "config-openai-model",
  "config-openai-api-key",
  "config-ollama-endpoint",
  "config-ollama-model",
].forEach((id) => {
  $(id)?.addEventListener("input", () => {
    renderConnectorSetupCard(state.config || {});
    renderConnectorTestState({ status: "idle", detail: "有未保存的连接改动；保存后再检查连接。" });
  });
});
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
  renderBudgetGuardPlanner();
  renderReuseGallery();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderStarterKitLauncher();
  renderSharePackBuilder();
  renderWorkspaceSnapshotPlanner();
});
$("share-pack-name").addEventListener("input", (event) => {
  state.sharePackName = event.target.value;
  renderSharePackBuilder();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("share-pack-audience").addEventListener("input", (event) => {
  state.sharePackAudience = event.target.value;
  renderSharePackBuilder();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("share-pack-intent").addEventListener("input", (event) => {
  state.sharePackIntent = event.target.value;
  renderSharePackBuilder();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("browser-target-url").addEventListener("input", (event) => {
  state.browserTargetURL = event.target.value;
  renderBrowserMissionPlanner();
});
$("browser-mission-goal").addEventListener("input", (event) => {
  state.browserMissionGoal = event.target.value;
  renderBrowserMissionPlanner();
});
$("data-source-descriptor").addEventListener("input", (event) => {
  state.dataSourceDescriptor = event.target.value;
  renderDataInsightPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("data-analysis-question").addEventListener("input", (event) => {
  state.dataAnalysisQuestion = event.target.value;
  renderDataInsightPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("data-output-format").addEventListener("input", (event) => {
  state.dataOutputFormat = event.target.value;
  renderDataInsightPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderWorkspaceSnapshotPlanner();
});
$("citation-claim-scope").addEventListener("input", (event) => {
  state.citationClaimScope = event.target.value;
  renderCitationGroundingPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderBudgetGuardPlanner();
  renderWorkspaceSnapshotPlanner();
});
$("citation-source-posture").addEventListener("input", (event) => {
  state.citationSourcePosture = event.target.value;
  renderCitationGroundingPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderBudgetGuardPlanner();
  renderWorkspaceSnapshotPlanner();
});
$("citation-evidence-level").addEventListener("input", (event) => {
  state.citationEvidenceLevel = event.target.value;
  renderCitationGroundingPlanner();
  renderResultLibrary();
  renderPlaybookLibrary();
  renderBudgetGuardPlanner();
  renderWorkspaceSnapshotPlanner();
});

try {
  applyRailCollapsed(localStorage.getItem("astriaRailCollapsed") === "1");
} catch {
  applyRailCollapsed(false);
}
renderHomeMode();
renderStrategyMatrix();
renderFileIntake();
renderSourceRegistry();
renderKnowledgeReconciliation();
renderCitationGroundingPlanner();
renderWorkspaceHub();
renderKnowledgeCuration();
renderToolDockInspector();
renderPromptSuggestionDock();
renderFocusBrief();
renderApprovalCenter();
renderWorkspaceHealthStrip();
renderComparisonWorkbench();
renderRunQualityScorecard();
renderPromptExperimentLab();
renderBudgetGuardPlanner();
renderReuseGallery();
renderResultLibrary();
renderPlaybookLibrary();
renderStarterKitLauncher();
renderSharePackBuilder();
renderWorkspaceSnapshotPlanner();
renderBrowserMissionPlanner();
renderDataInsightPlanner();
renderProactiveDeliveryBoard();
connectEventStream();
refreshAll();
