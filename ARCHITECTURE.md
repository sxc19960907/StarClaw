# StarClaw 架构文档

## 概述

StarClaw 是一个 AI 驱动的个人 CLI 工具，约 32,000 行 Go 代码，20 个 internal 包，33 个内置工具。

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLI / TUI                                │
│  chat · interactive · daemon · schedule · sessions · completion  │
│  mcp serve · mcp list · setup · version · update                 │
├──────────────────────────────────────────────────────────────────┤
│                       Agent Loop                                 │
│  · Loop Detection (9 patterns) · Retry + Backoff · Spill to Disk│
│  · Read Tracker · Context Compaction · Thinking Mode             │
│  · RunStatus · Bloat Detection · Time-based Compaction           │
│  · Semantic Consolidation · Streaming (OnPreamble+OnStreamDelta) │
├─────────────────────┬────────────────────┬───────────────────────┤
│    LLM Client       │    Tool System     │    Session Manager     │
│  Anthropic+OpenAI   │  33 built-in tools │  Tags+Favorites+Export │
│  (unified interface)│  + MCP Server      │  JSON file-based       │
├─────────────────────┴────────────────────┴───────────────────────┤
│                   Daemon (Background Service)                     │
│  HTTP Server (23 endpoints) · Cron Scheduler · Heartbeat         │
│  SessionCache · EventBus · ApprovalBroker · File Watcher         │
├─────────────────────┬────────────────────┬───────────────────────┤
│   Infrastructure    │     Security       │      Skills            │
│  Config · Hooks     │  Permissions (4L)  │  14 bundled skills     │
│  Prompt Builder     │  Audit Logging     │  Instructions Loader   │
│  Watcher · Update   │  Path Validation   │  MCP Client+Server     │
└─────────────────────┴────────────────────┴───────────────────────┘
```

## 核心模块

### Agent Loop (`internal/agent/`)

Agent 执行引擎，~4,500 行。核心职责：

- **Run()** — 主循环：发送 prompt → 收到工具调用 → 执行 → 发回结果 → 重复
- **Loop Detection** — 9 种模式检测（ExactDup, IdentityCycle, UnproductiveStreak, FileReadRepeat, ToolModeFlipFlop, SleepPattern, SearchEscalation, NoProgress, SuccessAfterError）
- **Retry + Backoff** — 瞬时错误自动重试，指数退避
- **Spill to Disk** — 50KB+ 工具结果写入临时文件
- **Context Compaction** — 对话过长时 LLM 摘要压缩
- **Time-based Compaction** — 超过时限的旧工具结果自动压缩
- **Semantic Consolidation** — 相同文件重复读取/相同 grep 合并
- **Thinking Mode** — adaptive/enabled extended thinking，streaming delta 支持
- **RunStatus** — context bloat 检测（工具结果占比 > 50% 告警）

### Tool System (`internal/tools/`)

33 个内置工具，~4,500 行：

| 类别 | 工具 |
|---|---|
| 文件 | file_read, file_write, file_edit |
| 搜索 | glob, directory_list, grep (rg + VCS skip + glob + mtime) |
| 执行 | bash (output cap), http |
| 推理 | think |
| 系统 | system_info, wait, version |
| 会话 | session_search, memory_append |
| 调度 | schedule_create, schedule_list, schedule_update, schedule_remove |
| 桌面 | clipboard, notify, screenshot, applescript, process |
| 发布 | publish_to_web |
| 技能 | use_skill |
| 协议 | MCP server (stdio) |

工具实现 `agent.Tool` 接口：`Info()` + `Run(ctx, args)` + `RequiresApproval()`。
可选接口：`ReadOnlyChecker`, `SafeChecker`, `ToolSourcer`。

### LLM Client (`internal/client/`)

统一的多模型接口 `LLMClient`：

- **AnthropicClient** — Claude API（默认）
- **OpenAIClient** — GPT-4o / 兼容 API（ollama 等）
- 配置 `provider: anthropic|openai` 切换
- 支持 Thinking Mode、Streaming Delta、Model Override

### Daemon (`internal/daemon/`)

后台服务，~5,000 行，最大模块：

```
┌─────────────────────────────────────────────┐
│              HTTP Server (:7533)             │
│  23 端点: health/status/message/schedules/   │
│  agents/sessions/permissions/approval/events │
├──────────────┬──────────────┬────────────────┤
│  Scheduler   │   Runner     │  SessionCache  │
│  每分钟 tick │  RunAgent()  │  Manager 缓存池 │
│  cron 评估   │  会话管理    │  Route Locking  │
└──────────────┴──────────────┴────────────────┘
```

- **Server** — Go 1.22+ 路由模式，JSON + SSE 双协议
- **Scheduler** — 内置 cron 解析器，并发上限 5，同分钟去重
- **Runner** — RunAgent：创建/恢复会话 → 运行 agent loop → 保存
- **SessionCache** — 懒加载 session.Manager 池，route-level 锁
- **ApprovalBroker** — 远程审批：等待 → 超时/批准/拒绝
- **EventBus** — SSE 事件广播

### TUI (`internal/tui/`)

Bubbletea 终端 UI，~2,000 行：

- Frog 像素动画启动画面（12 帧，Unicode half-block 渲染）
- Glamour Markdown 渲染
- 工具调用紧凑/详细双模式（✓/✗ 图标）
- 两栏启动头部：模型信息 + Tips + 最近活动

## 支撑模块

| 模块 | 行数 | 功能 |
|---|---|---|
| config | 1,000+ | YAML 配置 + thinking/agent/tools + OpenAI 支持 |
| context | 1,700+ | 上下文窗口管理 + 压缩 + 语义合并 |
| mcp | 1,100 | MCP 客户端 + stdio server |
| watcher | 924 | fsnotify 递归文件监听 + debounce + glob |
| heartbeat | 817 | 周期心跳检查（HEARTBEAT.md） |
| update | 810 | 自更新 + 自动检查 |
| session | 1,200+ | 会话持久化 + 标签 + 收藏 + 导出 |
| permissions | 692 | 4 层安全模型（全局→项目→agent→运行时） |
| skills | 594 | Skill 加载器 + 14 个内置 skill（188 文件） |
| agents | 567 | Agent 配置加载（AGENT.md + config.yaml） |
| hooks | 514 | 生命周期钩子（pre/post tool use） |
| audit | 443 | JSON-lines 审计日志 + 密钥脱敏 |
| schedule | 442 | Cron 定时任务 CRUD + flock 保护 |
| instructions | 401 | 层级指令加载（全局→项目→local） |
| prompt | 277 | 系统 prompt 构建（含 communicating-with-user） |

## 数据流

```
用户输入
  → CLI/TUI 解析命令
  → config.Load() 加载配置
  → client.NewClient() 工厂选择 provider
  → AgentLoop.Run()
    → prompt.Builder 构建系统 prompt
    → Instructions.Loader 注入自定义指令
    → Memory.Load 加载 MEMORY.md
    → LLMClient.Chat() 调用 Anthropic/OpenAI
    → 解析响应：文本 → TUI 渲染 / 工具调用 → 执行
    → permissions.Check() 权限检查
    → loopdetect.Check() 循环检测
    → hooks.Run() 钩子执行
    → audit.Log() 审计记录
    → compaction（时间/语义）压缩上下文
    → session.Save() 会话持久化
  → 返回结果
```

## 安全模型

4 层权限检查（`internal/permissions/`）：

1. **全局配置** — allowed/denied 工具列表
2. **项目配置** — .starclaw/config.local.yaml 覆盖
3. **Agent 配置** — 每个 agent 独立 allow/deny
4. **运行时** — 用户交互审批 + auto-approve

## CLI 命令

```
starclaw chat <query>              一次性查询
starclaw interactive               交互式 TUI
starclaw daemon start/stop/status  后台服务管理
starclaw schedule list/create/update/remove/enable/disable  定时任务
starclaw sessions list/tag/untag/favorite/unfavorite/export 会话管理
starclaw mcp list/serve            MCP 管理
starclaw completion bash/zsh/fish/install  自动补全
starclaw setup                     配置向导
starclaw update                    版本更新
```

## 配置示例

```yaml
provider: anthropic              # anthropic (default) or openai
endpoint: https://api.anthropic.com
api_key: sk-ant-xxx
model_tier: medium

# OpenAI
openai_api_key: sk-xxx
openai_model: gpt-4o

agent:
  thinking: true
  thinking_mode: adaptive
  max_iterations: 25

tools:
  bash_timeout: 120
  grep_max_results: 100

daemon:
  port: 7533
```

## 设计原则

- **接口驱动** — Tool/EventHandler/LLMClient 均为接口，可测试可替换
- **并发安全** — flock + mutex + semaphore 保护共享状态
- **多模型支持** — 统一 LLMClient 接口，Anthropic/OpenAI 自由切换
- **渐进式复杂度** — 核心路径短（chat ~200 行），高级功能按需加载
- **平台兼容** — macOS/Linux/Windows 跨平台工具支持

## 外部依赖

| 依赖 | 用途 |
|---|---|
| cobra | CLI 框架 |
| bubbletea + lipgloss + glamour | TUI |
| viper | 配置管理 |
| fsnotify | 文件监听 |
| mcp-go | MCP 协议 |
| testify | 测试断言 |
