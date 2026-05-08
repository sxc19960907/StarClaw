# Shell 深度集成

## Goal

让 StarClaw 成为顺手的第一公民 CLI 工具：自动补全、管道增强、CWD 感知。

## Requirements

### 1. Shell Completion（`cmd/completion.go`）
- `starclaw completion bash` — 生成 bash 补全脚本
- `starclaw completion zsh` — 生成 zsh 补全脚本  
- `starclaw completion fish` — 生成 fish 补全脚本
- Cobra 原生支持，只需注册命令

### 2. 一键安装脚本（`cmd/completion.go`）
- `starclaw completion install` — 自动检测 shell，安装补全到配置
- 打印安装路径和生效方式

### 3. 管道模式增强（`cmd/root.go`）
- 当通过管道传入时，自动检测内容类型
- 将 CWD 信息注入 prompt 上下文
- 管道模式下默认提高 max_iterations（批处理通常需要更多步骤）

### 4. CWD 感知
- `--cd` flag 让 agent 切换到指定目录执行
- 默认传递当前工作目录给 agent

## Acceptance Criteria

- [ ] `starclaw completion bash` 输出有效脚本
- [ ] `starclaw completion install` 一键安装
- [ ] 管道模式传递 CWD 信息
- [ ] `go build` 通过

## Technical Notes

- Cobra 1.10.2 自带 `GenZshCompletion`/`GenBashCompletion`/`GenFishCompletion`
- `os.Getwd()` 获取当前目录
