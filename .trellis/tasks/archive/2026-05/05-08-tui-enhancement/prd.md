# TUI Enhancement

## Goal

对照 ShanClaw 增强 StarClaw TUI：markdown 渲染、启动动画、工具格式化。

## Requirements

### 1. markdown.go — Glamour 渲染
- Markdown → ANSI 终端输出
- 自动剥离 Sources/References 段落
- 合并多余空行
- 依赖 `github.com/charmbracelet/glamour`

### 2. header.go — 启动头部动画
- 两栏布局：左栏 frog 像素动画 + 模型/CWD/endpoint，右栏 Tips + Recent activity
- 12 帧启动动画 (crouch→jump→land→blink)
- "StarClaw CLI" 标题（非 "Shannon CLI"）
- 适配 StarClaw 的导入路径

### 3. frog.go — Frog 像素动画
- 8×10 像素网格，half-block Unicode 渲染
- 5 个姿态：base, blink_half, blink_closed, crouch, jump
- lipgloss ANSI 颜色映射

### 4. toolformat.go — 工具调用格式化
- 工具参数摘要提取（按工具类型）
- 紧凑单行 + 展开详细两种格式
- ✓/✗ 状态图标
- 工具结果摘要行

### 5. app.go 集成
- 使用 header 渲染启动画面
- 使用 markdown 渲染 LLM 响应
- 使用 toolformat 渲染工具调用/结果
- 整体 TUI 体验提升

## Acceptance Criteria

- [ ] go build 通过
- [ ] 现有测试不变
- [ ] 无需额外外部依赖（glamour 可能已有）

## Technical Notes

- 源: `/Users/timmy/PycharmProjects/ShanClaw/internal/tui/`
- 需要 `go get github.com/charmbracelet/glamour`
- 适配 import path: `Kocoro-lab/ShanClaw` → `starclaw/starclaw`
- "Shannon" → "StarClaw" 文案替换
