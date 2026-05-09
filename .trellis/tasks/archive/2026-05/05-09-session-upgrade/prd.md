# 会话管理升级

## Goal

增强会话管理：标签系统、收藏、导出 Markdown/HTML。

## Requirements

### 1. 数据模型扩展 (`internal/session/session.go`)

Session 新增字段：
```go
Tags     []string `json:"tags,omitempty"`
Favorite bool     `json:"favorite"`
```

### 2. 标签管理 (`internal/session/manager.go`)

- `AddTag(sessionID, tag)` — 添加标签
- `RemoveTag(sessionID, tag)` — 移除标签
- `SetFavorite(sessionID, favorite)` — 收藏/取消
- `SearchByTag(tag) []SessionSummary` — 按标签搜索
- `ListFavorites() []SessionSummary` — 列出收藏

### 3. 导出 (`internal/session/export.go`)

- `ExportMarkdown(session) string` — 导出为 Markdown
- `ExportHTML(session) string` — 导出为 HTML
- `ExportToFile(session, format, path)` — 导出到文件

### 4. CLI 增强 (`cmd/sessions.go`)

- `starclaw sessions tag <id> <tag>` — 添加标签
- `starclaw sessions untag <id> <tag>` — 移除标签
- `starclaw sessions favorite <id>` — 收藏
- `starclaw sessions unfavorite <id>` — 取消收藏
- `starclaw sessions export <id> --format md|html` — 导出
- `sessions list` 显示标签和收藏状态

## Acceptance Criteria

- [ ] 标签增删查功能正常
- [ ] Markdown/HTML 导出格式正确
- [ ] 收藏功能正常
- [ ] CLI 新命令可用
- [ ] 现有测试不破坏

## Technical Notes

- 改动文件: `internal/session/session.go`, `internal/session/manager.go`, `internal/session/export.go`, `cmd/sessions.go`
- Session 用 JSON 文件存储，新字段向后兼容（omitempty）
