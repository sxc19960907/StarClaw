# Memory System

## Goal

完整记忆系统：agent 通过 `memory_append` 工具写入 MEMORY.md（flock 保护 + 溢出），跨会话自动加载，写入后不触发重读以保护 prompt cache。

## Requirements

### 1. BoundedAppend（补全到 persist.go）
- flock 文件锁保护并发写入
- 行数上限 150 行，超出则写入 `auto-YYYY-MM-DD-xxxxxx.md` 并加一行指针
- 输入校验：非空 content

### 2. memory_append 工具
- 实现 `agent.Tool` 接口
- 从 context 中获取 memoryDir（由 AgentLoop.Run 注入）
- 调用 `BoundedAppend(memoryDir, content)`
- 写入后**不重新加载** MEMORY.md（保护 prompt cache）
- 注册到 `RegisterLocalTools()`

### 3. ConsolidateMemory（碎片整理）
- 当 auto-*.md 文件 ≥ 12 个且上次 GC ≥ 7 天时触发
- LLM 去重合并 auto 条目 + MEMORY.md 中的 auto section
- 保留用户手写内容不变
- 通过 `Completer` 接口调用 small tier
- 原子写入（tmp + rename）

### 4. Memory 加载优化
- `loadMemory()` 只在 Run() 开头调用一次
- memory_append 写入后不触发 loadMemory()
- 新 session 启动时自动加载最新 MEMORY.md

## Acceptance Criteria

- [ ] memory_append 工具注册并可通过 agent.Tool 调用
- [ ] 并发写入 MEMORY.md 有 flock 保护
- [ ] 超过 150 行自动溢出到 detail 文件
- [ ] ConsolidateMemory: ≥12 auto 文件 → LLM 合并去重
- [ ] 写入后不重读，下个 session 能看到新记忆
- [ ] 所有新功能有单元测试

## Decision (ADR-lite)

**Context**: 记忆系统需要在会话内写入持久化，但不能触发 prompt cache 失效

**Decision**: memory_append 写入文件后，不调用 loadMemory() 重读。新记忆只在下次 session 启动时加载到系统 prompt。ConsolidateMemory 使用 small tier 的 LLM 调用做碎片整理。

**Consequences**: agent 在当前会话看不到刚写的记忆（可接受，因为写记忆时的上下文本身就在当前对话中）。prompt cache 不会因文件写入而失效。

## Out of Scope

- 跨 agent 实例的实时记忆同步
- 记忆向量化/语义搜索
- UI 编辑记忆

## Technical Notes

- `internal/context/persist.go` — 补全 BoundedAppend (120行)
- `internal/tools/memory_append.go` — 新建工具 (60行)
- `internal/context/persist.go` — 补全 ConsolidateMemory (100行)
- `internal/tools/register.go` — 注册新工具
- 参考: ShanClaw `internal/tools/memory_append.go` + `internal/context/persist.go`
