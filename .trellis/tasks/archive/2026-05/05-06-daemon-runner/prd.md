# Daemon Sub3: Runner - Agent Execution Pipeline

## Goal

实现 RunAgent 函数：接收 RunAgentRequest，创建/恢复会话，运行 agent loop，保存结果。

## Requirements

- `RunAgent(ctx, deps, req, handler) (RunAgentResponse, error)` — 核心入口
- 会话管理：命名的 agent 复用长期会话，默认 agent 每次新建
- Cancel 支持：通过 channel 取消运行中的 agent
- 结果保存：agent loop 完成后保存 session
- EventHandler 转发：将 agent loop 事件转给 handler

## Acceptance Criteria

- [ ] RunAgent 能成功运行 agent loop 并返回结果
- [ ] 命名 agent 会话复用逻辑正确
- [ ] 单元测试覆盖基本流程

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/runner.go` (741行)
- 可大幅简化：去掉 cloud_delegate 相关逻辑
- 依赖: internal/agent, internal/session, internal/config, internal/tools
