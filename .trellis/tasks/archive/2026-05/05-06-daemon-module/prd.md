# 实现 Daemon 后台服务模块 (完整版 - 方案C)

## Goal

对照 ShanClaw 的 `internal/daemon/`，在 StarClaw 中完整实现后台守护进程，包含 42 个 HTTP 端点、cron 调度器、agent 执行器、HTTP 客户端。

## What I already know

### ShanClaw 端点清单（42 个）

| 模块 | 端点数 | 端点 |
|---|---|---|
| Health/Status | 2 | GET /health, GET /status |
| Schedule | 5 | GET/POST /schedules, GET/PATCH/DELETE /schedules/{id} |
| Message/Cancel/Shutdown | 3 | POST /message, POST /cancel, POST /shutdown |
| Events | 1 | GET /events (SSE) |
| Agents | 10 | GET/POST /agents, GET/PUT/DELETE /agents/{name}, PUT/DELETE config/commands/skills |
| Skills | 16 | GET/PUT/DELETE skills 及子资源 management |
| Config | 4 | GET/PATCH /config, GET /config/status, POST /config/reload |
| Instructions | 2 | GET/PUT /instructions |
| Sessions | 2 | GET /sessions, DELETE /sessions/{id}, GET /sessions/search |
| Permissions | 2 | GET /permissions, POST /permissions/request |
| Approval | 1 | POST /approval |
| Chrome/CDP | 3 | GET /chrome/status, POST /chrome/show, POST /chrome/hide (暂不实现) |

### StarClaw 已有依赖

- `internal/schedule/` — Manager CRUD ✅
- `internal/agent/` — agent loop + Tool interface ✅
- `internal/session/` — session.Manager ✅
- `internal/config/` — Config struct + StarclawDir ✅
- `internal/instructions/` — loader ✅
- `internal/agents/` — LoadAgent, ListAgents ✅
- `internal/permissions/` — CheckToolCall ✅
- `internal/skills/` — use_skill ✅

## Decision (ADR-lite)

**Context**: 方案 B（~10 端点）vs 方案 C（42 端点完整对等）

**Decision**: 选择方案 C，拆分为 5 个子任务按依赖顺序实现。

**Consequences**: 工作量约 2500+ 行代码（排除 Chrome/CDP），需要引入 gronx 依赖（用于 IsDue），server.go 是最大的单文件。

## Out of Scope (explicit)

- launchd 服务安装
- macOS TCC 权限管理
- PID 文件管理
- Chrome/CDP 端点（3 个）

## Implementation Plan

按依赖关系拆分为 5 个子任务：

```
Sub1: Foundation (types + events + approval)
  ↓
Sub2: Scheduler (cron execution engine)
  ↓
Sub3: Runner (RunAgent + cancel logic)
  ↓
Sub4: Client (HTTP client for CLI↔daemon)
  ↓
Sub5: Server (full HTTP server, 42 endpoints)
```

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/`
- StarClaw 模块路径: `github.com/starclaw/starclaw`
- 需要 `gronx.IsDue` 做 cron 到期判断（仅此一处引入外部依赖）
- RunAgent 可简化：去掉 cloud_delegate 相关逻辑
- ServerDeps 需要汇集所有内部模块引用
