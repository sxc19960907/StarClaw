# Daemon Sub5: Server - Full HTTP API Server

## Goal

实现完整的 HTTP REST API 服务器，包含 39 个端点（排除 Chrome/CDP 3 个），覆盖健康检查、定时任务、agent 运行、管理等功能。

## Requirements

### 核心端点（必须）

| 模块 | 端点 | 用途 |
|---|---|---|
| Health | GET /health, GET /status | 健康检查 + 运行状态 |
| Message | POST /message | 同步/SSE 运行 agent |
| Control | POST /cancel, POST /shutdown | 取消任务/停止服务 |
| Events | GET /events | SSE 实时事件推送 |
| Schedule | GET/POST /schedules, GET/PATCH/DELETE /schedules/{id} | 定时任务 CRUD |

### 管理端点（按已有模块实现）

| 模块 | 端点 | 
|---|---|
| Agents | GET/POST /agents, GET/PUT/DELETE /agents/{name}, PUT/DELETE config/commands/skills |
| Skills | GET/PUT/DELETE skills 管理 + install |
| Config | GET/PATCH /config, GET /config/status, POST /config/reload |
| Instructions | GET/PUT /instructions |
| Sessions | GET /sessions, DELETE /sessions/{id}, GET /sessions/search |
| Permissions | GET /permissions, POST /permissions/request |
| Approval | POST /approval |

### 核心类型

- `Server` 结构体：port, deps (*ServerDeps), http.Server, eventBus, approvalBroker
- `NewServer(port, deps, version)` 构造函数
- `Start(ctx)` — 注册路由 + ListenAndServe
- 所有 handler 用 Go 1.22 路由模式（`GET /health`）

## Acceptance Criteria

- [ ] 所有核心端点正确响应
- [ ] SSE /events 推送正确
- [ ] /message 支持同步 JSON 和 SSE 流式
- [ ] 单元测试用 httptest 覆盖所有端点

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/server.go` (2215行)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/router.go` (510行)
- ServerDeps 已在 types.go 定义，包含所有模块引用
- 使用 Go 1.22+ net/http 原生的路由模式
- 排除 Chrome/CDP 端点
- 对于 StarClaw 尚无 CLI 的模块（如 agents/skills config），handler 可返回 "not implemented yet"
