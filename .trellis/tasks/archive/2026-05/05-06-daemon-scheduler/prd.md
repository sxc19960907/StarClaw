# Daemon Sub2: Scheduler - Cron Execution Engine

## Goal

实现 cron 调度器：每分钟 tick，评估已启用的定时任务是否到期，派发 agent 执行。

## Requirements

- `Scheduler` 结构体：manager (*schedule.Manager), deps (*ServerDeps), gronx, lastFired map, sem (bounded concurrency)
- `NewScheduler(mgr, deps)` 构造函数
- `Start(ctx)` — 立即 tick 一次，对齐下一分钟边界，然后每分钟循环
- `tick(ctx)` — EvaluateDue + 非阻塞 goroutine 派发（sem 限流 5 并发）
- `EvaluateDue(now) []schedule.Schedule` — 过滤 enabled + gronx.IsDue + 去重同一分钟
- `runSchedule(ctx, sched)` — 构造 RunAgentRequest，调用 RunAgent
- `scheduleHandler` — 静默 EventHandler（auto-approve 工具调用）
- 依赖 `gronx` 做 IsDue 判断（go get github.com/adhocore/gronx）

## Acceptance Criteria

- [ ] Scheduler 每分钟 tick 一次，对齐 wall-clock
- [ ] 正确评估 cron 表达式（*/5, ranges, lists）
- [ ] 去重：同一 schedule 同一分钟只触发一次
- [ ] 并发限制：最多 5 个 schedule 同时执行
- [ ] 禁用 schedule 不被触发
- [ ] 单元测试覆盖 EvaluateDue 逻辑

## Technical Notes

- 参考: `internal/schedule.Manager` (已实现)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/scheduler.go` (162行)
- 依赖: github.com/adhocore/gronx (网络可能不可用，考虑自实现 IsDue)
- gronx 之前 go get 超时，备选方案是自实现一个简单的 cron due checker
