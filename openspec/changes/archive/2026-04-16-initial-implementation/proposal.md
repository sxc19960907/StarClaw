## Why

StarClaw 是一个 AI Agent CLI 工具，受 ShanClaw 启发，旨在提供一个高性能、可扩展的本地 AI 助手。
当前需要构建一个最小可用的基础版本，包含核心工具系统、配置管理、Agent 对话循环和基础的 TUI 界面。
这将为后续的高级功能（Daemon 模式、MCP、技能系统等）奠定坚实基础。

## What Changes

- **项目骨架搭建**: 初始化 Go 项目结构，包含 cmd/、internal/ 等目录
- **配置系统**: 实现基于 YAML 的配置管理，支持 API Key 配置和初始化向导
- **核心工具**: 实现基础工具集（file_read, file_write, file_edit, glob, grep, bash, directory_list）
- **Agent 核心**: 实现 Agent 对话循环，支持 LLM 调用和工具调用链
- **LLM 客户端**: 集成 Claude API（Anthropic SDK）
- **基础 CLI**: 实现 One-shot 模式（单次对话）和简单的交互式 CLI
- **TUI 界面**: 实现基于 Bubbletea 的基础交互界面
- **测试覆盖**: 每个功能点都有对应的单元测试和集成测试

## Capabilities

### New Capabilities

- `config-management`: 配置加载、保存、层级管理和初始化向导
- `core-tools`: 基础文件和系统工具集（file, bash, glob, grep 等）
- `agent-loop`: Agent 对话循环和工具调用管理
- `llm-client`: LLM API 客户端（Claude/OpenAI）
- `basic-cli`: 基础命令行界面（One-shot 和简单交互）
- `tui-interface`: 基于 Bubbletea 的终端用户界面

### Modified Capabilities

- （无现有 capability 需要修改）

## Impact

- **代码结构**: 新建完整的 Go 项目结构，包含 15+ 个包
- **依赖**: 引入 cobra、viper、bubbletea、anthropic-go 等依赖
- **配置**: 创建 `~/.starclaw/config.yaml` 配置文件
- **构建**: 添加 Makefile 和 goreleaser 配置
- **分发**: 准备 npm 包装脚本用于二进制分发
