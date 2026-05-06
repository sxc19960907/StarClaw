# General Utility Tools

## Goal

实现 3 个通用工具：session_search（会话搜索）、version（版本查询）、wait（轮询等待）。

## Requirements

### session_search
- 搜索历史会话内容（通过 session.Manager）
- 参数：query（关键词）、limit（结果上限）
- 返回匹配的会话摘要和匹配行

### version
- 返回 StarClaw 版本信息
- 显示 Go 版本、OS/Arch

### wait
- 轮询等待工具，替代 `bash sleep`
- 参数：seconds（等待秒数，默认 5，最大 30）
- 用于等待异步操作（构建、服务启动等）

## Acceptance Criteria

- [ ] 3 个工具注册到 RegisterLocalTools
- [ ] 各工具 info/run/approval 正确实现
- [ ] 单元测试覆盖

## Technical Notes

- session_search 参考: ShanClaw `internal/tools/session_search.go` (71行)
- 其他 2 个按 StarClaw 风格自行实现
