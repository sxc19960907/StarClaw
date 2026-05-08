# Heartbeat - Periodic Health Check

## Goal

实现周期性心跳检查，按 agent 配置定时运行 agent 自我诊断。

## Requirements

- 扫描 agents 目录中配置了 heartbeat 的 agent
- 读取 HEARTBEAT.md 作为检查清单
- Goal-driven 模式：发 prompt → agent 自查 → 返回 "HEARTBEAT_OK"
- 每个 agent 独立 ticker goroutine
- TryLock 防重叠
- Active hours 时段过滤
- 使用 SessionCache 注入会话上下文
- 最小 1 分钟间隔

## Acceptance Criteria

- [ ] 编译通过，go test 通过
- [ ] 覆盖：ticker 触发、OK 响应、非 OK 响应、重叠防护、时段过滤

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/heartbeat/heartbeat.go`
- 依赖: internal/daemon (SessionCache, RunAgent, ServerDeps), internal/agents, internal/watcher (InActiveHours)
- 文件: `internal/heartbeat/heartbeat.go` + `heartbeat_test.go`
