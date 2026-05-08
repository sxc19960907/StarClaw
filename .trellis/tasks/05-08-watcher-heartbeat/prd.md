# 实现 Watcher 和 Heartbeat 模块

## Goal

对照 ShanClaw 的 `internal/watcher/` 和 `internal/heartbeat/`，在 StarClaw 中实现文件系统监听和心跳健康检查模块。

## What I already know

### Watcher（ShanClaw: 370行 + 280行测试）

- 基于 `github.com/fsnotify/fsnotify` 实现递归目录监听
- 跳过规则：node_modules, .git, vendor 等 20 个默认忽略目录
- 最大 4096 个监听目录上限，防止 fd 耗尽
- Glob 匹配 + agent 绑定：每个 watch 绑定到特定 agent
- Debounce 防抖：默认 2 秒，可配
- Rate limiting：agent 执行间隔限制
- Active hours：时段过滤
- 运行时自动添加新建目录
- Stats 指标跟踪
- RunFunc 回调触发 agent 执行

### Heartbeat（ShanClaw: 330行 + 170行测试）

- 按 agent 配置周期性心跳检查
- 读取 agent 目录的 HEARTBEAT.md 作为检查清单
- Goal-driven 模式：发 prompt → agent 用工具自查 → 返回 "HEARTBEAT_OK" 或描述问题
- 每个 agent 独立 goroutine ticker
- TryLock 防重叠（一次只允许一个 tick 运行）
- Active hours 时段过滤
- 最小 1 分钟间隔
- Session 持久化：心跳结果保存到 agent 会话
- EventBus 告警推送

### StarClaw 适配调整

| 功能 | ShanClaw | StarClaw |
|---|---|---|
| SessionCache | 有 | 无 → 简化，每次新建会话 |
| WSClient (cloud) | 有 | 无 → 移除 |
| client.Message | 有 | 无 → 直接使用 agent.ToolResult |
| daemon.ServerDeps | 完整 | 简化版 deps |

## Requirements

### Watcher

- 创建 `Watcher` 结构体，支持 `New(agentWatches, runFn, opts...)`
- 递归监听目录，跳过默认忽略列表
- Glob 匹配 + agent 绑定
- Debounce 事件批处理（默认 2s）
- Rate limiting（可选）
- Active hours（可选）
- 动态添加新目录
- Stats 导出
- 充分测试：目录监控、事件批处理、glob 匹配、debounce、rate limit、active hours

### Heartbeat

- 创建 `Manager` 结构体，扫描 agents 目录中配置了 heartbeat 的 agent
- 每个 agent 独立 ticker goroutine
- 读取 HEARTBEAT.md 作为检查清单
- 调用 RunAgent 执行心跳检查
- 检查响应是否为 "HEARTBEAT_OK"
- TryLock 防重叠
- Active hours 时段过滤
- 充分测试：ticker 触发、OK 响应、非 OK 响应、重叠防护、时段过滤

## Decision (ADR-lite)

**Context**: 两个独立模块，无互相依赖，可并行开发

**Decision**: 拆分为 2 个子任务并行实现

**Consequences**: 各自独立可测试、可提交

## Implementation Plan

```
Sub1: Watcher — internal/watcher/watcher.go + watcher_test.go
Sub2: Heartbeat — internal/heartbeat/heartbeat.go + heartbeat_test.go
```

两个子任务无依赖，可并行派发。

## Out of Scope

- SessionCache 集成（StarClaw 暂无）
- WSClient 云推送
- TranscriptCollector（使用 StarClaw 原生类型）

## Technical Notes

- 依赖: `github.com/fsnotify/fsnotify`（需 go get）
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/watcher/watcher.go`
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/heartbeat/heartbeat.go`
- StarClaw 模块路径: `github.com/starclaw/starclaw`
- Watcher 可完全搬迁，Heartbeat 需简化适配（去掉 SessionCache/WSClient 依赖）
