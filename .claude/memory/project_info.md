---
name: project_info
description: StarClaw 项目信息
type: project
tags: project, starclaw
---

StarClaw 是一个简化版智能体系统，目标是复刻 Claude Code 的核心能力。

## 架构

```
starclaw/
├── core/      # Agent 循环、上下文
├── llm/       # OpenAI/Claude 客户端
├── tools/     # 内置工具
├── mcp/       # MCP 客户端
└── skills.py  # 技能系统
```

## 当前 Phase

Phase 4: 高级特性 - 记忆系统已实现

## 技术栈

- Python 3.9+
- OpenAI/Claude API
- MCP (Model Context Protocol)
- Textual (TUI)
