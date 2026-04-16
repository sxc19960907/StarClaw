# StarClaw 功能对齐参考文档

基于 ShanClaw 的深度代码分析，本文档提供 StarClaw 的功能对齐映射。

## ShanClaw 代码统计

```
语言: Go 1.25.7
文件数: 192 个 .go 文件
测试文件: 89 个 _test.go
代码行数: ~28,000 行 (不含测试)
核心模块: 22 个 internal 包
```

## 核心功能对齐矩阵

### Phase 1: 基础工具 (优先级最高)

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/tools/file_read.go` | `internal/tools/file_read.go` | 读取文件带行号 | ✅ 单元测试 |
| `internal/tools/file_write.go` | `internal/tools/file_write.go` | 写入文件 | ✅ 单元测试 |
| `internal/tools/file_edit.go` | `internal/tools/file_edit.go` | 编辑文件 | ✅ 单元测试 |
| `internal/tools/glob.go` | `internal/tools/glob.go` | 文件模式匹配 | ✅ 单元测试 |
| `internal/tools/grep.go` | `internal/tools/grep.go` | 内容搜索 | ✅ 单元测试 |
| `internal/tools/directory_list.go` | `internal/tools/directory_list.go` | 目录列表 | ✅ 单元测试 |
| `internal/tools/bash.go` | `internal/tools/bash.go` | Bash 执行 | ✅ 单元测试 + 安全测试 |
| `internal/tools/safe_path.go` | `internal/tools/safe_path.go` | 路径安全 | ✅ 安全测试 |

### Phase 2: Agent 核心

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/agent/tools.go` | `internal/agent/tools.go` | Tool 接口定义 | ✅ 接口测试 |
| `internal/agent/loop.go` | `internal/agent/loop.go` | Agent 主循环 ⭐ | ✅ 集成测试 |
| `internal/tools/register.go` | `internal/tools/register.go` | 工具注册表 | ✅ 注册测试 |
| `internal/client/client.go` | `internal/client/anthropic.go` | Claude API | ✅ Mock 测试 |
| `internal/client/client.go` | `internal/client/openai.go` | OpenAI API | ✅ Mock 测试 |

### Phase 3: 配置与 CLI

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/config/config.go` | `internal/config/config.go` | 配置管理 | ✅ 配置测试 |
| `internal/config/setup.go` | `internal/config/setup.go` | 初始化向导 | ✅ 交互测试 |
| `cmd/root.go` | `cmd/root.go` | 根命令 | ✅ E2E 测试 |
| `cmd/daemon.go` | `cmd/daemon.go` | Daemon 子命令 | ✅ 集成测试 |

### Phase 4: TUI

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/tui/app.go` | `internal/tui/app.go` | TUI 主应用 | ✅ 组件测试 |
| `internal/tui/header.go` | `internal/tui/header.go` | 头部组件 | ✅ 渲染测试 |
| `internal/tui/markdown.go` | `internal/tui/markdown.go` | Markdown 渲染 | ✅ 渲染测试 |

### Phase 5: Daemon 模式

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/daemon/server.go` | `internal/daemon/server.go` | HTTP 服务器 | ✅ API 测试 |
| `internal/daemon/client.go` | `internal/daemon/client.go` | WebSocket | ✅ 连接测试 |
| `internal/session/*.go` | `internal/session/*.go` | 会话管理 | ✅ 存储测试 |

### Phase 6: 高级功能

| ShanClaw 文件 | StarClaw 对应 | 功能 | 测试要求 |
|---------------|---------------|------|----------|
| `internal/mcp/*.go` | `internal/mcp/*.go` | MCP 客户端 | ✅ 集成测试 |
| `internal/skills/*.go` | `internal/skills/*.go` | 技能系统 | ✅ 加载测试 |
| `internal/schedule/*.go` | `internal/schedule/*.go` | 定时任务 | ✅ 调度测试 |
| `internal/audit/*.go` | `internal/audit/*.go` | 审计日志 | ✅ 日志测试 |

## 关键代码模式

### 1. Tool 接口模式

```go
// ShanClaw: internal/agent/tools.go

type Tool interface {
    Info() ToolInfo
    Run(ctx context.Context, args string) (ToolResult, error)
    RequiresApproval() bool
}

// 可选接口
type SafeChecker interface {
    IsSafeArgs(argsJSON string) bool
}

type ReadOnlyChecker interface {
    IsReadOnlyCall(argsJSON string) bool
}
```

**StarClaw 实现**: 保持完全一致

### 2. Tool 结果类型

```go
// ShanClaw: internal/agent/tools.go

type ToolResult struct {
    Content       string
    IsError       bool
    ErrorCategory ErrorCategory // transient/validation/business/permission
    IsRetryable   bool
    Images        []ImageBlock
}

// 错误分类
const (
    ErrCategoryTransient  ErrorCategory = "transient"
    ErrCategoryValidation ErrorCategory = "validation"
    ErrCategoryBusiness   ErrorCategory = "business"
    ErrCategoryPermission ErrorCategory = "permission"
)
```

**StarClaw 实现**: 保持完全一致

### 3. Agent Loop 核心

```go
// ShanClaw: internal/agent/loop.go

type AgentLoop struct {
    client            *client.GatewayClient
    tools             *ToolRegistry
    handler           EventHandler
    maxIter           int
    maxTokens         int
    // ... 其他字段
}

func (a *AgentLoop) Run(ctx context.Context, query string, history []Message) (*RunResult, error) {
    // 1. 构建系统提示
    // 2. 调用 LLM
    // 3. 处理工具调用
    // 4. 循环直到完成或达到最大迭代
}
```

**StarClaw 实现**: 简化版本，核心逻辑保持一致

### 4. 配置层级

```go
// ShanClaw: internal/config/config.go

// 优先级: local > project > global > default
// ~/.shannon/config.yaml          (global)
// ./.shannon/config.yaml          (project)
// ./.shannon/config.local.yaml    (local)
```

**StarClaw 实现**: 简化为两层 (global + local)

## 代码差异点

### StarClaw 不做的事

| ShanClaw 功能 | StarClaw 决策 | 理由 |
|---------------|---------------|------|
| Shannon Cloud 集成 | ❌ 移除 | 改用原生 Claude/OpenAI API |
| 复杂的 MCP 工具切换 | ❌ 简化 | 基础 MCP 支持即可 |
| launchd 集成 | ❌ 移除 | Linux 兼容，使用 cron |
| Chrome CDP 自动管理 | ❌ 简化 | 依赖外部 Chrome |
| Agent 复杂配置覆盖 | ❌ 简化 | 基础配置即可 |

### StarClaw 新增的事

| 功能 | 说明 |
|------|------|
| 多 Provider 支持 | Claude + OpenAI 切换 |
| 简化配置 | 减少层级，更易理解 |
| 纯 CLI 模式 | 无 TUI 的轻量模式 |

## 测试覆盖要求

### 每个 Phase 的测试标准

```
Phase 1 (工具):
- 每个工具函数单元测试
- 边界条件测试 (空文件、大文件、无权限)
- 安全测试 (路径遍历、命令注入)

Phase 2 (Agent):
- Mock LLM 客户端
- 工具调用链测试
- 错误处理测试
- 上下文管理测试

Phase 3 (配置):
- 配置读写测试
- 配置合并测试
- 加密存储测试

Phase 4 (TUI):
- 组件渲染测试
- 事件处理测试
- 状态管理测试

Phase 5 (Daemon):
- API 端点测试
- WebSocket 连接测试
- 会话持久化测试
```

## 开发顺序建议

```
1. 先实现基础工具 (file, bash, glob, grep)
   ↓ 验证: 命令行直接调用工具

2. 实现 Agent Loop 简化版
   ↓ 验证: 能进行简单对话

3. 添加配置系统
   ↓ 验证: 能配置 API Key

4. 实现 One-shot CLI
   ↓ 验证: 单次对话完整流程

5. 添加 TUI
   ↓ 验证: 交互式体验

6. 添加 Daemon 模式
   ↓ 验证: HTTP API 可用

7. 添加高级功能
   ↓ 验证: 完整生态
```

## 参考代码位置

```
/Users/timmy/PycharmProjects/ShanClaw/
├── cmd/                    # CLI 命令参考
├── internal/agent/         # Agent 核心参考 ⭐
├── internal/tools/         # 工具实现参考 ⭐
├── internal/config/        # 配置管理参考
├── internal/daemon/        # Daemon 参考
├── internal/tui/           # TUI 参考
├── internal/mcp/           # MCP 参考
├── internal/skills/        # 技能系统参考
├── internal/schedule/      # 定时任务参考
└── npm/                    # 分发包装参考
```

## 许可证说明

StarClaw 参考 ShanClaw 的设计和代码模式，但实现为独立项目。
ShanClaw: Apache-2.0 License
StarClaw: MIT License
