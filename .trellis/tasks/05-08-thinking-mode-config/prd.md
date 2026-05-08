# Thinking Mode 和配置扩展

## Goal

对照 ShanClaw 实现 extended thinking 支持和配置扩展，提升 agent loop 的推理能力和流式体验。

## Scope

### 1. Thinking Mode（~3 文件改动）

| 改动 | 文件 | 说明 |
|---|---|---|
| ThinkingConfig 类型 | `internal/client/client.go` | Type, BudgetTokens 字段 |
| StreamDelta 类型 + CompleteStream | `internal/client/client.go` | SSE 流式 completion |
| Config 新增 5 字段 | `internal/config/config.go` | thinking/thinking_mode/thinking_budget/reasoning_effort/model |
| AgentConfig 扩展 | `internal/config/config.go` | 同上，加到 AgentConfig |
| AgentLoop 字段 | `internal/agent/loop.go` | thinking, reasoningEffort, specificModel, enableStreaming |
| Setter 方法 | `internal/agent/loop.go` | SetThinking, SetReasoningEffort, SetSpecificModel, SetEnableStreaming |
| OnStreamDelta | `internal/agent/tools.go` | EventHandler 新方法 |
| LLM 调用改造 | `internal/agent/loop.go` | 首次尝试 streaming，重试 fallback 到 non-streaming |

### 2. 配置扩展

| 字段 | 默认值 | 说明 |
|---|---|---|
| `agent.thinking` | true | 启用 extended thinking |
| `agent.thinking_mode` | "adaptive" | adaptive 或 enabled |
| `agent.thinking_budget` | 10000 | thinking token budget |
| `agent.reasoning_effort` | "" | reasoning effort level |
| `agent.model` | "" | 特定 model 覆盖 |
| `tools.grep_max_results` | 100 | grep 最大结果数 |

### 3. ThinkingConfig JSON

```go
type ThinkingConfig struct {
    Type         string `json:"type"`                    // "adaptive", "enabled", "disabled"
    BudgetTokens int    `json:"budget_tokens,omitempty"` // only for "enabled" mode
}
```

### 4. 逻辑

- thinking 启用 → 根据 thinking_mode 决定 Type
  - "adaptive" → ThinkingConfig{Type: "adaptive"}
  - "enabled" → ThinkingConfig{Type: "enabled", BudgetTokens: thinking_budget}
- thinking 禁用 → 发空 ThinkingConfig 或不发
- streaming 启用 → 首次尝试 CompleteStream，fallback 到普通 Complete
- 重试时跳过 streaming（避免重复 delta）

## Acceptance Criteria

- [ ] Config 新字段解析正确，默认值生效
- [ ] ThinkingConfig 正确序列化为 JSON
- [ ] AgentLoop 接受 thinking/reasoningEffort/model 配置
- [ ] OnStreamDelta 回调正确触发
- [ ] 所有现有测试通过，新测试覆盖新增逻辑

## Implementation Plan

拆分为 2 个子任务：
```
Sub1: Config + Client types (ThinkingConfig, StreamDelta, config fields)
Sub2: Agent loop integration (streaming, thinking, model override)
```

Sub1 必须先完成（Sub2 依赖新类型）。

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/config/config.go` (agent.thinking 相关)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/client/gateway.go` (ThinkingConfig, CompleteStream)
- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/agent/loop.go` (SetThinking, streaming logic)
- StarClaw 当前 loop.go ~434 行，改动量约 +100 行
- StarClaw 当前 config.go ~286 行，改动量约 +40 行
