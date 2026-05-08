# Thinking Sub2: Agent Loop Integration

## Goal

将 Thinking Mode、Streaming、Model Override 集成到 AgentLoop。

## Requirements

### 1. AgentLoop 新增字段
- thinking *client.ThinkingConfig
- reasoningEffort string
- specificModel string
- enableStreaming bool

### 2. Setter 方法
- SetThinking(cfg *client.ThinkingConfig)
- SetReasoningEffort(effort string)
- SetSpecificModel(model string)
- SetEnableStreaming(enable bool)

### 3. EventHandler 扩展
- 在 `internal/agent/tools.go` 的 EventHandler 接口中新增 OnStreamDelta(delta string)

### 4. LLM 调用改造
- 首次尝试 streaming（如果 enableStreaming && handler != nil）
- 重试时跳过 streaming
- 在 completion request 中传入 thinking, reasoningEffort, specificModel

### 5. thinkingConfig 构建逻辑
```go
func buildThinkingConfig(cfg config.AgentConfig) *client.ThinkingConfig {
    if !cfg.Thinking {
        return nil
    }
    tc := &client.ThinkingConfig{}
    switch cfg.ThinkingMode {
    case "adaptive":
        tc.Type = "adaptive"
    case "enabled":
        tc.Type = "enabled"
        tc.BudgetTokens = cfg.ThinkingBudget
    }
    return tc
}
```

## Acceptance Criteria

- [ ] 编译通过
- [ ] 现有测试全部通过（不破坏已有功能）
- [ ] 新方法有单元测试

## Technical Notes

- 改动: `internal/agent/loop.go` (+~60行), `internal/agent/tools.go` (+1行)
- 参考: ShanClaw `internal/agent/loop.go` SetThinking/streaming logic
