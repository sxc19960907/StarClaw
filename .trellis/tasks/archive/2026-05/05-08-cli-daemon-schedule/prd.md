# CLI Daemon 和 Schedule 命令

## Goal

对照 ShanClaw 实现 CLI 子命令：`starclaw daemon start/stop/status` 和 `starclaw schedule list/create/update/remove/enable/disable`。

## Requirements

### schedule 命令（`cmd/schedule.go`）
- `schedule list` — 表格输出所有定时任务
- `schedule create --agent --cron --prompt` — 创建定时任务
- `schedule update <id> --cron --prompt` — 更新
- `schedule remove <id>` — 删除
- `schedule enable <id>` / `schedule disable <id>` — 启用/禁用
- 直接使用 `internal/schedule.Manager`

### daemon 命令（`cmd/daemon.go`）
- `daemon start` — 启动后台 HTTP 服务 + scheduler + heartbeat + watcher
- `daemon stop` — 通过 HTTP /shutdown 优雅停止
- `daemon status` — 查询 HTTP /status 显示运行状态
- 使用 `internal/daemon/*` 现有模块

### 注册（`cmd/root.go`）
- `rootCmd.AddCommand(daemonCmd)`
- `rootCmd.AddCommand(scheduleCmd)`

## Acceptance Criteria

- [ ] `starclaw schedule list` 输出正确
- [ ] `starclaw daemon start` 启动服务并响应 HTTP
- [ ] `go build` 通过，CLI 可正常使用

## Implementation Plan

```
Sub1: cmd/schedule.go — schedule CRUD 命令
Sub2: cmd/daemon.go — daemon start/stop/status
```

两个无依赖，可并行。

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/cmd/schedule.go` (166行)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/cmd/daemon.go` (482行，大幅简化)
- StarClaw 简化版 daemon start：去掉 WS client、launchd、pidfile、MCP supervisor
