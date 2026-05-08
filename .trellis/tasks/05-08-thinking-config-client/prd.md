# Thinking Sub1: Config + Client Types

## Goal

新增 ThinkingConfig/StreamDelta 类型 + 扩展 Config 字段。

## Requirements

### Config 扩展 (`internal/config/config.go`)

AgentConfig 新增字段：
```go
Thinking        bool   // default true
ThinkingMode    string // "adaptive" (default) or "enabled"
ThinkingBudget  int    // default 10000
ReasoningEffort string // default ""
Model           string // specific model override (empty = use model_tier)
```

ToolsConfig 新增：
```go
GrepMaxResults int // default 100
```

默认值 + 验证逻辑（thinking_mode 只能是 "adaptive" 或 "enabled"）。

### Client 新增 (`internal/client/client.go`)

```go
type ThinkingConfig struct {
    Type         string `json:"type"`
    BudgetTokens int    `json:"budget_tokens,omitempty"`
}

type StreamDelta struct {
    Text string
}
```

请求结构中新增 Thinking + ReasoningEffort + SpecificModel 字段。

## Acceptance Criteria

- [ ] Config 新字段正确解析
- [ ] ThinkingConfig JSON 序列化正确
- [ ] 编译通过，现有测试不破坏

## Technical Notes

- 改动文件: `internal/config/config.go`, `internal/client/client.go`
- 参考: ShanClaw `internal/config/config.go` 和 `internal/client/gateway.go`
