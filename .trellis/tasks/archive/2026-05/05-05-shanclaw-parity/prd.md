# ShanClaw Parity: Remaining Features

## Goal

对比 ShanClaw 后确定 StarClaw 还需要实现哪些功能，排定优先级。

## What I already know

上次 session 已经实现了：
- MCP Client（真实 mcp-go 集成）
- MCP Tool Adapter（MCPTool 适配器）
- Config Merge（MergeAgentConfig）
- Agent 集成（SwitchAgent, memory, --agent flag）
- 全面测试覆盖

## Gap Analysis

### 缺失的 internal 包（8 个）

| 包 | ShanClaw 功能 | 优先级评估 |
|---|---|---|
| `context/` | 上下文窗口管理、压缩、摘要 | ★★★★★ |
| `permissions/` | 工具权限 allow/deny 系统 | ★★★★ |
| `hooks/` | 生命周期钩子（before/after） | ★★★ |
| `prompt/` | 系统 prompt 构建器 | ★★★ |
| `daemon/` | 后台 HTTP 服务 + 调度 | ★★★ |
| `schedule/` | cron 定时任务 | ★★ |
| `heartbeat/` | 连接健康监控 | ★★ |
| `watcher/` | 文件系统监听 | ★★ |
| `instructions/` | 指令文件加载 | ★★ |

### Agent Loop 差距

| 功能 | 说明 |
|---|---|
| **Context Compaction** | 对话过长时自动压缩，防止超 token 限制 |
| **Loop Detection** | 检测 agent 循环调用同一工具 |
| **Spill to Disk** | 长对话溢出到磁盘 |
| **Thinking Mode** | extended thinking 支持 |
| **Read Tracker** | 追踪已读文件 |
| **Micro-compaction** | 微压缩 |
| **Retry Logic** | API 失败重试 |

### 缺失的工具

| 类别 | 工具 |
|---|---|
| macOS 桌面 | accessibility, applescript, browser, clipboard, ghostty, imaging, notify, screenshot, process |
| 基础设施 | cloud_delegate, memory_append, schedule, server, session_search, version, wait |
| Computer Use | computer（鼠标/键盘控制） |

### 缺失的 Bundled Skills（14 个）

algorithmic-art, brand-guidelines, canvas-design, claude-api, doc-coauthoring, frontend-design, internal-comms, mcp-builder, skill-creator, slack-gif-creator, theme-factory, web-artifacts-builder, webapp-testing

## Requirements (evolving)

待确定优先级后填充。

## Open Questions

* 优先实现哪个？Agent Loop 核心能力 vs 权限系统 vs macOS 工具？

## Acceptance Criteria (evolving)

* [ ] 明确功能优先级排序
* [ ] 确定 MVP scope

## Out of Scope (explicit)

待定

## Technical Notes

- 参考项目: `/Users/timmy/PycharmProjects/ShanClaw`
- 上次 session 已实现: MCP client, MCP tool adapter, config merge, agent 集成
- 代码可直接从 ShanClaw 搬迁适配
