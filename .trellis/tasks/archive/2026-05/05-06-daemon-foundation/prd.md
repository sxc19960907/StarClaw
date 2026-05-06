# Daemon Sub1: Foundation (types + events + approval)

## Goal

定义 daemon 模块的共享类型、事件总线、审批代理——所有其他子任务的基础。

## Requirements

### types.go — 共享类型

- `Channel` 常量: ChannelCLI, ChannelHTTP, ChannelSchedule
- `RunAgentRequest`: Text, Agent, Source, Channel, Sender, NewSession, Model, RequestID, Attachments
- `RunAgentResponse`: SessionID, Messages, Usage, Error
- `ServerDeps`: 汇集所有内部模块引用（StarclawDir, Config, AgentLoop, SessionMgr, ScheduleMgr, ToolReg, SkillsDir, etc.）
- `ApprovalRequest/ApprovalResolvedPayload`: 远程审批类型

### events.go — 事件总线

- `EventBus` 结构体：订阅者管理，发布 SSE 事件
- `Subscribe/Unsubscribe/Publish` 方法
- 事件类型常量：tool_call, tool_result, text, approval_needed 等

### approval.go — 审批代理

- `ApprovalBroker`：等待审批决议的通道
- `WaitForApproval/Resolve` 方法
- 超时处理

## Acceptance Criteria

- [ ] types.go 定义所有共享类型，编译通过
- [ ] events.go EventBus 有 Subscribe/Unsubscribe/Publish 测试
- [ ] approval.go ApprovalBroker 有 Wait/Resolve/Timeout 测试
- [ ] 所有类型与 ShanClaw 兼容（字段名、JSON tag 一致）

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/types.go` (133行)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/events.go` (71行)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/approval.go` (158行)
- StarClaw 模块路径: `github.com/starclaw/starclaw/internal/daemon`
- ServerDeps 不需要包含所有字段，可以按后续子任务逐步补全
