# OpenAI 多模型支持

## Goal

添加 OpenAI API 支持，让用户可以在 Anthropic 和 OpenAI 之间切换，也可以配置自定义 endpoint 连接兼容 API（如 ollama）。

## Requirements

### 1. LLMClient 接口化 (`internal/client/`)
定义 `LLMClient` 接口：
```go
type LLMClient interface {
    Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions) (*Response, error)
}
```

### 2. Anthropic 实现 (`internal/client/anthropic.go`)
将现有的 Anthropic API 调用逻辑封装为 `AnthropicClient`，实现 `LLMClient` 接口。

### 3. OpenAI 实现 (`internal/client/openai.go`)
`OpenAIClient` 结构体，实现 `LLMClient` 接口：
- 请求格式转换为 OpenAI Chat Completions API
- 工具调用格式转换为 OpenAI function calling
- 响应格式统一为 `Response` 结构

### 4. Config 扩展 (`internal/config/`)
新增字段：
```go
Provider string // "anthropic" (default) or "openai"
OpenAIAPIKey string
OpenAIEndpoint string // default "https://api.openai.com/v1"
OpenAIModel string // default "gpt-4o"
```

### 5. 客户端工厂
```go
func NewClient(cfg *Config) (LLMClient, error)
```
根据 `cfg.Provider` 创建对应的客户端。

### 6. SSH 集成
更新 `cmd/root.go` 使用 `NewClient` 工厂方法，所有调用方使用接口。

## Acceptance Criteria

- [ ] OpenAI Chat Completions API 调用成功
- [ ] 工具调用（function calling）正常工作
- [ ] 配置切换 `provider: openai` 即可使用
- [ ] 向后兼容，Anthropic 默认不变
- [ ] 现有测试全部通过

## Implementation Plan

拆分为 2 个子任务：
```
Sub1: client 抽象化 + OpenAI 实现 (internal/client/)
Sub2: config + CLI 集成
```

Sub1 必须先完成（Sub2 依赖新接口）。

## Technical Notes

- OpenAI Chat Completions API: POST https://api.openai.com/v1/chat/completions
- Function calling 格式: tools 数组 + tool_choice
- 支持 OPENAI_API_KEY 和 OPENAI_BASE_URL 环境变量
