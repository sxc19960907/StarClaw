# Watcher - File System Monitoring

## Goal

实现文件系统监听器，当项目文件变化时自动触发 agent 执行。基于 fsnotify 实现递归目录监控。

## Requirements

- 基于 `github.com/fsnotify/fsnotify`
- 递归监听目录，默认跳过 20 个生成/缓存目录
- Glob 匹配 + agent 绑定
- Debounce 防抖（默认 2s）
- Rate limiting（按 agent 限流）
- 运行时自动添加新建目录
- 最大 4096 目录上限
- Stats 统计导出
- RunFunc 回调触发 agent

## Acceptance Criteria

- [ ] 编译通过，go test 通过
- [ ] 覆盖：glob 匹配、debounce 批处理、rate limit、目录跳过、动态目录添加

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/watcher/watcher.go`
- 依赖: `github.com/fsnotify/fsnotify`
- 文件: `internal/watcher/watcher.go` + `watcher_test.go`
