# Agent Loop Upgrade

## Goal

从 ShanClaw 搬迁 Loop Detection、Spill to Disk、Retry Logic、Read Tracker 四个模块到 StarClaw，提升 agent 循环的健壮性和可靠性。

## Requirements

### 1. Loop Detection
- 检测 agent 是否反复调用同一个工具（SHA256 哈希比对）
- 三种响应级别：LoopContinue（正常）、LoopNudge（提示换方法）、LoopForceStop（强制终止）
- 记录工具调用历史：名称、参数哈希、结果签名、是否 error

### 2. Spill to Disk
- 工具结果超过 50,000 字符时写入磁盘临时文件
- 上下文中保留前 2,000 字符预览
- 会话结束时清理临时文件
- 需要传入 StarClaw 配置目录（~/.starclaw/tmp/）

### 3. Retry Logic
- API 调用失败时自动重试（transient errors）
- 区分可重试错误（网络超时、429）和不可重试错误（401）
- 退避策略

### 4. Read Tracker
- 追踪 agent 已读取的文件（通过 context 传递）
- 避免重复读取同一文件
- 记录读取文件的路径
- 提供 WithReadTracker / ReadTrackerFromContext

## Acceptance Criteria

- [ ] Loop Detection: 连续 3 次相同工具调用触发 LoopForceStop
- [ ] Spill: 50KB+ 工具结果写入磁盘，上下文只保留预览
- [ ] Retry: transient error 自动重试最多 3 次
- [ ] Read Tracker: 已读文件记录在 context 中，可查询
- [ ] 所有新模块有单元测试

## Technical Approach

直接从 ShanClaw 搬迁代码，适配 StarClaw 的 import path 和配置路径：
- `internal/agent/loopdetect.go` — 从 ShanClaw 搬迁
- `internal/agent/spill.go` — 从 ShanClaw 搬迁，`shannonDir` → `starclawDir`
- `internal/agent/readtracker.go` — 从 ShanClaw 搬迁
- Retry logic — 从 ShanClaw loop.go 提取，集成到现有 Run() 中

## Out of Scope

- Micro-compact（LLM 摘要压缩）
- Context Compaction（对话窗口压缩）
- Deferred Execution
- Partition Concurrency
- Normalize

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/agent/`
- StarClaw 当前 loop.go: 288 行（需扩展）
- ShahClaw loop.go: 2444 行（仅提取 retry 部分）
- StarClaw 配置目录: `~/.starclaw/`（对应 ShanClaw 的 `~/.shannon/`）
- 所有模块是独立添加，不改变现有 loop.go 核心逻辑
