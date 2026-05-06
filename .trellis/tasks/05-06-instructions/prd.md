# Instructions Loader

## Goal

实现指令文件加载器：从全局和项目级 `.md` 文件加载自定义指令，注入到系统 prompt。

## Requirements

- LoadInstructions(starclawDir, projectDir, maxTokens) — 层级加载
  - 全局: ~/.starclaw/instructions.md → ~/.starclaw/rules/*.md
  - 项目: .starclaw/instructions.md → .starclaw/rules/*.md → .starclaw/instructions.local.md
- LoadMemory(starclawDir, maxLines) — 加载 MEMORY.md
- 去重：高优先级文件的相同行覆盖低优先级
- 源头注释：`<!-- from: path -->`
- Token 预算截断

## Out of Scope

- LoadCustomCommands（需要 BuiltinCommands 设施）

## Technical Notes

- 对照: ShanClaw `internal/instructions/loader.go` (378行)
