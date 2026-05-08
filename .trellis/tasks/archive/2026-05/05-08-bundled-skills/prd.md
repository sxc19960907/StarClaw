# Bundled Skills

## Goal

从 ShanClaw 搬迁 14 个内置 skill 到 StarClaw，放到 `internal/skills/bundled/skills/` 目录。

## Requirements

- 复制全部 14 个 skill 目录到 `internal/skills/bundled/skills/`
- 保持 SKILL.md frontmatter 和内容不变（这些 skill 是通用的）
- 检查并适配任何引用路径（如 import path 从 ShanClaw → StarClaw）
- 复制 THIRD_PARTY_NOTICES.md
- 不需要修改代码，纯文件复制

## Skill 清单

| Skill | 文件数 | 用途 |
|---|---|---|
| algorithmic-art | 4 | 算法生成艺术 |
| brand-guidelines | 2 | 品牌规范 |
| canvas-design | 83 | Canvas 设计 |
| claude-api | 26 | Claude API/SDK 开发 |
| doc-coauthoring | 1 | 文档协作 |
| frontend-design | 2 | 前端设计 |
| internal-comms | 6 | 内部通讯 |
| mcp-builder | 10 | MCP 构建器 |
| skill-creator | 18 | Skill 创建器 |
| slack-gif-creator | 7 | Slack GIF 创建 |
| theme-factory | 13 | 主题工厂 |
| web-artifacts-builder | 5 | Web artifacts |
| webapp-testing | 6 | Web 应用测试 |

## Acceptance Criteria

- [ ] 14 个 skill 目录完整复制
- [ ] go build 通过（不引入编译错误）
- [ ] 无 ShanClaw import path 引用

## Technical Notes

- 源: `/Users/timmy/PycharmProjects/ShanClaw/internal/skills/bundled/skills/`
- 目标: `/Users/timmy/PycharmProjects/StarClaw/internal/skills/bundled/skills/`
- `cp -r` 即可，检查是否有 Go 源文件需要调整 import path
