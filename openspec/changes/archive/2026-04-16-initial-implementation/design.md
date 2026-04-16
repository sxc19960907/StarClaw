## Context

StarClaw 是一个 AI Agent CLI 工具，参考 ShanClaw 的设计。项目目前处于空白状态，需要从 0 开始构建。

### 参考信息

- **ShanClaw**: 28,000+ 行 Go 代码，89 个测试文件，使用 cobra/viper/bubbletea
- **目标**: 实现最小可用版本（MVP），包含核心工具、配置、Agent 循环和基础 TUI
- **语言**: Go 1.22+
- **分发**: npm 包装 + GitHub Releases

## Goals / Non-Goals

**Goals:**
1. 构建可编译运行的项目骨架
2. 实现配置管理和初始化向导
3. 实现 7 个核心工具（file_read, file_write, file_edit, glob, grep, bash, directory_list）
4. 实现 Agent 对话循环（支持工具调用链）
5. 集成 Claude API 客户端
6. 实现 One-shot CLI 模式
7. 实现基础 TUI 交互界面
8. 每个功能都有完整测试覆盖

**Non-Goals:**
- Daemon 模式（Phase 2）
- MCP 客户端（Phase 2）
- 技能系统（Phase 2）
- 定时任务（Phase 2）
- 审计日志（Phase 2）
- OpenAI 支持（Phase 1.5，Claude 优先）
- Web 界面（Phase 3）

## Decisions

### 1. 工具接口设计

**决策**: 完全参考 ShanClaw 的 Tool 接口

```go
type Tool interface {
    Info() ToolInfo
    Run(ctx context.Context, args string) (ToolResult, error)
    RequiresApproval() bool
}
```

**理由**: ShanClaw 的设计经过实战检验，保持兼容便于迁移和对比

### 2. 配置系统

**决策**: 使用 viper + YAML，简化为两层配置

```
~/.starclaw/config.yaml       (global)
./.starclaw/config.local.yaml (local - optional)
```

**对比**: ShanClaw 有四层（default, global, project, local）
**理由**: 减少复杂度，global 足够管理个人设置，local 用于敏感信息

### 3. Agent 循环架构

**决策**: 参考 ShanClaw loop.go 的核心逻辑，但简化实现

核心流程:
1. 构建系统提示 + 工具定义
2. 调用 LLM
3. 解析响应（文本或 tool_use）
4. 执行工具调用
5. 将结果返回 LLM
6. 循环直到无 tool_use 或达到最大迭代

**理由**: 保持架构清晰，避免过早优化

### 4. TUI 框架

**决策**: 使用 Charmbracelet 生态
- bubbletea: TUI 框架
- lipgloss: 样式
- glamour: Markdown 渲染

**理由**: ShanClaw 已验证此组合，文档完善

### 5. LLM 客户端

**决策**: 优先实现 Claude API（Anthropic SDK）

**理由**: 
- Claude 在编码任务上表现优秀
- 官方 Go SDK 可用
- OpenAI 可在后续迭代添加

### 6. 测试策略

**决策**: 每个 Phase 都必须有可用性验证

- 单元测试: 覆盖每个函数
- 集成测试: 模块间协作
- E2E 测试: 完整用户场景

**理由**: 防止"全量但不可用"的情况

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 启动延迟（无 native 二进制） | 用户体验 | 后续添加 goreleaser 构建 |
| 工具安全性（bash 执行） | 安全 | 强制审批 + 白名单机制 |
| LLM API 成本 | 成本 | 添加 token 使用统计 |
| 配置迁移（后续扩展） | 兼容性 | 使用版本号管理配置 |
| 单文件体积（TUI 依赖） | 分发 | 使用 -ldflags="-s -w" 压缩 |

## Migration Plan

无需迁移（新项目）

部署步骤:
1. 开发完成并通过所有测试
2. goreleaser 构建多平台二进制
3. GitHub Release 发布
4. npm 包更新 install.js
5. 用户通过 `npm install -g @starclaw/cli` 安装

## Open Questions

1. **API Key 存储**: 是否加密？如何保护？
   - 建议: 使用系统 keychain（macOS）或文件权限（Linux）

2. **会话存储**: 是否保留对话历史？
   - 建议: Phase 1 简化，只保留当前会话，不持久化

3. **工具审批**: 是否所有工具都需要审批？
   - 建议: file_read 和 glob 可配置为免审批，其他需要
