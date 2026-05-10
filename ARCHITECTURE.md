# StarClaw 架构文档

## 概述

StarClaw 是一个 AI 驱动的个人 CLI 工具，约 47,000 行 Go 代码（含测试），23 个 internal 包，36 个内置工具，3 种模型 provider（Anthropic + OpenAI + Ollama）。

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLI / TUI                                │
│  chat · interactive · daemon · schedule · sessions · completion  │
│  mcp serve · mcp list · setup · version · update                 │
├──────────────────────────────────────────────────────────────────┤
│                       Agent Loop                                 │
│  · Loop Detection (9 patterns) · Retry + Backoff · Spill        │
│  · Read Tracker · Context Compaction · Thinking Mode             │
│  · Time/Semantic Compaction · Tool Budget · RunStatus            │
│  · PhaseTracker · UsageTracker · StateCache · Watchdog           │
│  · SkillDiscovery · Normalize · Streaming (OnPreamble+Delta)     │
├─────────────────────┬────────────────────┬───────────────────────┤
│    LLM Client       │    Tool System     │    Session Manager     │
│  Anthropic+OpenAI   │  34 built-in tools │  Tags+Favorites+Export │
│  +Ollama (3 models) │  + MCP Server      │  JSON file-based       │
├─────────────────────┴────────────────────┴───────────────────────┤
│                   Daemon (Background Service)                     │
│  HTTP Server(23eps) · Scheduler · Heartbeat · SessionCache       │
│  EventBus · ApprovalBroker · Attachment · Safeguard · Rules      │
│  ReadTrackerCache · SessionCWD · Marketplace · File Watcher      │
├─────────────────────┬────────────────────┬───────────────────────┤
│   Infrastructure    │     Security       │      Skills            │
│  Config · Hooks     │  Permissions (4L)  │  14 bundled skills     │
│  Prompt Builder     │  Audit Logging     │  Instructions Loader   │
│  Watcher · Update   │  Path Validation   │  MCP Client+Server     │
│  cwdctx·uploads·runstatus  │  Safeguard │  Skill Discovery       │
└─────────────────────┴────────────────────┴───────────────────────┘
```

## 核心模块

### Agent Loop (`internal/agent/`)

Agent 执行引擎，~6,000 行。核心职责：

- **Run()** — 主循环：发送 prompt → 收到工具调用 → 执行 → 发回结果 → 重复
- **Loop Detection** — 9 种模式检测 + normalize（URL/query 规范化）
- **Retry + Backoff** — 瞬时错误自动重试，指数退避
- **Context Compaction** — LLM 摘要 + 时间压缩 + 语义合并
- **Tool Budget** — 单工具结果字符预算，防止上下文膨胀
- **Thinking Mode** — adaptive/enabled extended thinking + streaming
- **RunStatus** — context bloat 检测 + 阶段跟踪 (PhaseTracker)
- **UsageTracker** — token 用量和成本统计
- **StateCache** — 跨轮次状态缓存
- **Watchdog** — 防卡死看门狗定时器
- **Skill Discovery** — 自动发现 ~/.starclaw/skills/ 中的 skill
- **ResultShape** — LLM 响应形状分析（text/tool_calls/error）
- **WarmSet** — 预热工具集管理
- **CacheMetric** — 缓存命中率 + P50/P95/P99 延迟统计

### Tool System (`internal/tools/`)

36 个内置工具，~6,000 行：

| 类别 | 工具 |
|---|---|
| 文件 | file_read, file_write, file_edit, filepreview |
| 搜索 | glob, directory_list, grep (rg + VCS skip + glob + mtime) |
| 执行 | bash (output cap), http |
| 推理 | think |
| 系统 | system_info, wait, version |
| 会话 | session_search, memory_append |
| 调度 | schedule_create, schedule_list, schedule_update, schedule_remove |
| 桌面 | clipboard, notify, screenshot, applescript, process |
| macOS | accessibility, computer (mouse/keyboard), browser |
| 发布 | publish_to_web |
| 技能 | use_skill, skill (动态加载) |
| 安全 | readonly (只读模式) |
| 记忆 | memory (搜索/列出/删除) |
| 图像 | imaging (describe/resize/convert/OCR) |
| 协议 | MCP server (stdio) |

### LLM Client (`internal/client/`)

统一的多模型接口 `LLMClient`：

- **AnthropicClient** — Claude API（默认）
- **OpenAIClient** — GPT-4o / 兼容 API
- **OllamaClient** — 本地 Ollama 模型（llama3.1 等）
- 配置 `provider: anthropic|openai|ollama` 切换

### Daemon (`internal/daemon/`)

后台服务，~6,500 行：

- **Server** — Go 1.22+ 路由模式，JSON + SSE 双协议，23 端点
- **Scheduler** — 内置 cron 解析器，并发上限 5，同分钟去重
- **Runner** — RunAgent：创建/恢复会话 → 运行 agent loop → 保存
- **SessionCache** — 懒加载 session.Manager 池，route-level 锁
- **ApprovalBroker** — 远程审批：等待 → 超时/批准/拒绝
- **EventBus** — SSE 事件广播
- **Attachment** — 文件附件上传、列表
- **ReadTrackerCache** — 跨重启读追踪缓存
- **Rules** — Agent 规则文件管理
- **Safeguard** — 危险命令拦截（rm -rf /, fork bomb 等）
- **Marketplace** — Skill 市场列表
- **SessionCWD** — 按会话跟踪工作目录
- **ProjectInit** — 项目初始化（创建 .starclaw/ 骨架）
- **Checkpoint** — Agent 检查点保存/恢复

### TUI (`internal/tui/`)

Bubbletea 终端 UI，~2,000 行：Frog 像素动画（12 帧）、Glamour Markdown 渲染、工具调用紧凑/详细双模式、两栏启动头部。

## 支撑模块

| 模块 | 行数 | 功能 |
|---|---|---|
| config | 1,200+ | YAML 配置 + 3 provider 支持 + thinking/agent/tools 设置 |
| context | 1,700+ | 上下文窗口管理 + 压缩 + 语义合并 |
| mcp | 1,100 | MCP 客户端 + stdio server |
| watcher | 924 | fsnotify 递归文件监听 + debounce + glob |
| heartbeat | 817 | 周期心跳检查（HEARTBEAT.md） |
| update | 810 | 自更新 + 自动检查 |
| session | 1,200+ | 会话持久化 + 标签 + 收藏 + Markdown/HTML 导出 |
| permissions | 692 | 4 层安全模型（全局→项目→agent→运行时） |
| skills | 594 | Skill 加载器 + 14 个内置 skill（188 文件） |
| agents | 567 | Agent 配置加载（AGENT.md + config.yaml） |
| hooks | 514 | 生命周期钩子（pre/post tool use） |
| audit | 443 | JSON-lines 审计日志 + 密钥脱敏 |
| schedule | 442 | Cron 定时任务 CRUD + flock 保护 |
| instructions | 401 | 层级指令加载（全局→项目→local） |
| prompt | 277 | 系统 prompt 构建（含 communicating-with-user） |
| cwdctx | 新增 | 请求级 CWD 上下文 |
| uploads | 新增 | 文件上传管理 |
| runstatus | 新增 | 运行状态定义 |

## 数据流

```
用户输入
  → CLI/TUI 解析命令
  → config.Load() 加载配置
  → client.NewClient() 工厂选择 provider (Anthropic/OpenAI/Ollama)
  → AgentLoop.Run()
    → prompt.Builder 构建系统 prompt
    → Instructions.Loader 注入自定义指令
    → Memory.Load 加载 MEMORY.md
    → LLMClient.Chat() 调用模型
    → 解析响应：文本 → TUI 渲染 / 工具调用 → 执行
    → permissions.Check() 权限检查
    → loopdetect.Check() 循环检测 + normalize
    → hooks.Run() 钩子执行
    → audit.Log() 审计记录
    → toolbudget.Consume() 检查字符预算
    → compaction（时间/语义）压缩上下文
    → session.Save() 会话持久化（含标签/收藏）
  → 返回结果
```

## 安全模型

- 4 层权限：全局 → 项目 → Agent → 运行时
- Safeguard：危险命令拦截（rm -rf /、mkfs、fork bomb 等）
- 路径安全检查 + 审计日志 + 密钥脱敏
- ReadOnly 模式：只读工具中间件

## CLI 命令

```
starclaw chat <query>                   一次性查询
starclaw interactive                    交互式 TUI
starclaw daemon start/stop/status       后台服务管理
starclaw schedule list/create/update/   定时任务
  remove/enable/disable
starclaw sessions list/tag/untag/       会话管理
  favorite/unfavorite/export
starclaw mcp list/serve                 MCP 管理
starclaw completion bash/zsh/fish/      自动补全
  install
starclaw setup                          配置向导
starclaw update                         版本更新
```

## 模型提供商

| Provider | 模型 | 配置 |
|---|---|---|
| anthropic | Claude | `endpoint` + `api_key` + `model_tier` |
| openai | GPT-4o | `openai_api_key` + `openai_model` |
| ollama | llama3.1 等 | 本地 `http://localhost:11434` |

## 设计原则

- **接口驱动** — Tool/EventHandler/LLMClient 均为接口
- **并发安全** — flock + mutex + semaphore + TryLock
- **多模型** — 统一接口，3 provider 自由切换
- **渐进式复杂度** — 核心短（chat ~200 行），高级功能按需
- **跨平台** — macOS/Linux/Windows，平台工具降级优雅
