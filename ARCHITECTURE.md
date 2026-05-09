# StarClaw 架构文档

## 概述

StarClaw 是一个 AI 驱动的个人 CLI 工具，约 28,000 行 Go 代码，20 个 internal 包。

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLI / TUI                                │
│  chat · interactive · daemon · schedule · sessions · completion  │
├──────────────────────────────────────────────────────────────────┤
│                       Agent Loop                                 │
│  · Loop Detection  · Retry + Backoff  · Spill to Disk            │
│  · Read Tracker    · Context Compaction  · Thinking Mode         │
│  · RunStatus       · Context Bloat Detection                     │
├─────────────────────┬────────────────────┬───────────────────────┤
│    LLM Client       │    Tool System     │    Session Manager     │
│  (Anthropic API)    │  18 built-in tools │  JSON file-based       │
├─────────────────────┴────────────────────┴───────────────────────┤
│                         Daemon (Background Service)               │
│  HTTP Server (23 endpoints) · Cron Scheduler · Heartbeat         │
│  SessionCache · EventBus · ApprovalBroker                        │
├─────────────────────┬────────────────────┬───────────────────────┤
│   Infrastructure    │     Security       │      Skills            │
│  Config · Hooks     │  Permissions (4L)  │  14 bundled skills     │
│  Prompt Builder     │  Audit Logging     │  Instructions Loader   │
│  Watcher · Update   │  Path Validation   │  MCP Client            │
└─────────────────────┴────────────────────┴───────────────────────┘
```

## 核心模块

### Agent Loop (`internal/agent/`)

Agent 执行引擎，3,838 行。核心职责：

- **Run()** — 主循环：发送 prompt → 收到工具调用 → 执行 → 发回结果 → 重复
- **Loop Detection** — SHA256 哈希检测重复工具调用，三级响应（Continue/Nudge/ForceStop）
- **Retry + Backoff** — 瞬时错误自动重试，指数退避
- **Spill to Disk** — 50KB+ 工具结果写入临时文件
- **Context Compaction** — 对话过长时 LLM 摘要压缩
- **Thinking Mode** — adaptive/enabled extended thinking，streaming delta 支持
- **RunStatus** — context bloat 检测（工具结果占比 > 50% 告警）

### Tool System (`internal/tools/`)

18 个内置工具，3,801 行：

| 类别 | 工具 |
|---|---|
| 文件 | file_read, file_write, file_edit |
| 搜索 | glob, directory_list, grep (rg + VCS skip) |
| 执行 | bash (output cap), http |
| 推理 | think |
| 系统 | system_info, wait, version |
| 会话 | session_search, memory_append |
| 调度 | schedule_create, schedule_list, schedule_update, schedule_remove |
| 其他 | use_skill, publish_to_web |

工具实现 `agent.Tool` 接口：`Info()` + `Run(ctx, args)` + `RequiresApproval()`。

### Daemon (`internal/daemon/`)

后台服务，4,792 行，最大模块：

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

Bubbletea 终端 UI，1,735 行：

- Frog 像素动画启动画面（12 帧）
- Glamour Markdown 渲染
- 工具调用紧凑/详细双模式（✓/✗ 图标）
- 两栏启动头部：模型信息 + Tips + 最近活动

## 支撑模块

| 模块 | 行数 | 功能 |
|---|---|---|
| config | 935 | YAML 配置加载 + thinking/agent/tools 设置 |
| context | 1,431 | 上下文窗口管理 + 压缩 |
| mcp | 1,061 | MCP 协议客户端 |
| watcher | 924 | fsnotify 递归文件监听 + debounce |
| heartbeat | 817 | 周期心跳检查（HEARTBEAT.md） |
| update | 810 | 自更新 + 自动检查 |
| session | 807 | JSON 会话持久化 |
| permissions | 692 | 4 层安全模型（全局→项目→agent→运行时） |
| skills | 594 | Skill 加载器 + 14 个内置 skill |
| agents | 567 | Agent 配置加载（AGENT.md + config.yaml） |
| hooks | 514 | 生命周期钩子（pre/post tool use） |
| audit | 443 | JSON-lines 审计日志 + 密钥脱敏 |
| schedule | 442 | Cron 定时任务 CRUD + flock 保护 |
| instructions | 401 | 层级指令加载（全局→项目→local） |
| prompt | 277 | 系统 prompt 构建 |

## 数据流

```
用户输入
  → CLI/TUI 解析命令
  → config.Load() 加载配置
  → AgentLoop.Run()
    → prompt.Builder 构建系统 prompt
    → Instructions.Loader 注入自定义指令
    → Memory.Load 加载 MEMORY.md
    → client.Complete() 调用 LLM
    → 解析响应：文本 → TUI 渲染 / 工具调用 → 执行
    → permissions.Check() 权限检查
    → hooks.Run() 钩子执行
    → audit.Log() 审计记录
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
starclaw chat <query>         一次性查询
starclaw interactive          交互式 TUI
starclaw daemon start/stop/status  后台服务管理
starclaw schedule list/create/update/remove/enable/disable  定时任务
starclaw sessions              会话列表
starclaw setup                 配置向导
starclaw completion            自动补全
starclaw mcp list              MCP 服务器列表
starclaw update                版本更新
```

## 外部依赖

| 依赖 | 用途 |
|---|---|
| cobra | CLI 框架 |
| bubbletea + lipgloss + glamour | TUI |
| viper | 配置管理 |
| fsnotify | 文件监听 |
| mcp-go | MCP 协议 |
| testify | 测试断言 |

## 设计原则

- **零网络依赖** — 所有本地工具不访问外部服务
- **接口驱动** — Tool/EventHandler/LLMClient 均为接口，可测试可替换
- **并发安全** — flock + mutex + semaphore 保护共享状态
- **渐进式复杂度** — 核心路径短（chat 命令 ~200 行），高级功能按需加载
