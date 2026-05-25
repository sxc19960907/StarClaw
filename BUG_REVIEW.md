# StarClaw 代码审查报告

审查了 249 个 Go 源文件，覆盖 20 个包。日期：2026-05-10。

---

## 严重级别（共 5 个）

### 1. `client/mock.go` — MockClient 数据竞争

**行号:** 9-17, 41-128

`MockClient` 所有字段（`response`, `toolCallName`, `toolCallArgs`, `responseFunc`, `callCount`, `lastMessages`, `lastTools`）均无互斥锁保护。并发测试中会触发数据竞争。`loop_blackbox_test.go` 已在 `blackboxMockClient` 中自行添加 `sync.Mutex`，说明这个问题是已知的。

```go
type MockClient struct {
    response       string
    toolCallName   string
    toolCallArgs   string
    responseFunc   func(input string) *MockMessage
    lastMessages   []Message
    lastTools      []ToolDef
    callCount      int
}
```

### 2. `agent/loop.go` — 流式处理导致重复 API 调用

**行号:** 575-588

`StreamChat` 成功后执行会穿透到 `a.llmClient.Chat()`，导致每次流式请求都发出第二次非流式 API 调用。单次请求消耗双倍 token。

```go
if attempt == 0 && a.enableStreaming {
    if streamer, ok := a.llmClient.(StreamingLLMClient); ok {
        err := streamer.StreamChat(...)
        if err == nil {
            // TODO: implement proper streaming response handling.
        }
    }
}
resp, err := a.llmClient.Chat(ctx, ...)  // always called
```

### 3. `context/sanitize.go` — `mergeConsecutiveRoles` 替换而非合并消息

**行号:** 65-78

函数名和注释说"合并"，但实际用新消息**覆盖**前一条消息，前一条消息内容被静默丢弃。`SanitizeHistory` 和 `ConsolidateRedundant` 两条路径都受影响。

```go
case "assistant", "user":
    out[len(out)-1] = msg   // REPLACES, not merges
    continue
```

### 4. `context/window.go` — `CompressOldToolResults` 重复计数轮次对

**行号:** 109-139

向后迭代时对 `[user][assistant]` 和 `[assistant][user]` 两个角色转换都计数，导致一个 assistant-user 轮次对被计为 2 次。造成系统性过度截断旧消息。

```go
if messages[i].Role != messages[i+1].Role {
    pairs++  // counts both transitions per turn
}
```

### 5. `daemon/checkpoint.go` — `sanitizeID` 路径遍历

**行号:** 62-69

`filepath.Clean("..")` 返回 `".."`，不被 `.` 检查捕获，且不包含需要替换的分隔符。ID 为 `../../tmp` 可逃逸出检查点目录。

```go
cleaned := filepath.Clean(id)
cleaned = strings.ReplaceAll(cleaned, string(filepath.Separator), "_")
if cleaned == "." || cleaned == "" {
    return "_default"
}
```

---

## 高优先级（共 14 个）

### client 包

**6. `client/sse.go:134-176`** — 流结束时最后一条 SSE 事件未被刷新（RFC 8895 违规）。`for scanner.Scan()` 循环仅在遇到空行时分发事件，循环结束后未刷新 `current` 中已累积的事件。

**7. `client/openai.go:158-166`** — 类型断言失败时返回空响应 `(resp, nil)` 而非错误，使 API 协议违规不可检测。

### agent 包

**8. `agent/loop.go:614-638`** — `isRetryableLLMError` 中 `"deadline exceeded"` 子串匹配优先于 `"context deadline exceeded"` 排除检查，导致上下文超时被误分类为可重试。

**9. `agent/registry.go:9-55`** — `ToolRegistry` 无互斥锁；`List()` 和 `Names()` 都调用 `sort.Strings(r.order)` 直接修改共享切片。

**10. `agent/readtracker.go:27-41`** — `MarkRead` 和 `HasRead` 对 `map[string]bool` 的读写无同步。

### tools 包

**11. `tools/safe_path.go:28-34`** — `IsSafePath` 被符号链接绕过。`filepath.Abs` 不解析符号链接，验证的是未解析路径，但后续 I/O 操作会跟随符号链接。

**12. `tools/safe_path.go:85-98`** — `IsPathUnderCWD` 前缀冲突。CWD `/home/user/project` 会错误匹配 `/home/user/project-other/secret.txt`。

**13. `tools/imaging.go:63-79`** — 图片工具遗漏 `ExpandHome` 和 `IsSafePath` 调用，可操作任意文件。`imagemagickResize` 原地覆盖文件。

**14. `tools/publish_to_web.go:118`** — `RequiresApproval()` 返回 false，agent 可自动执行该工具，静默将任意文件外泄到本地 web 服务器。

**15. `tools/screenshot.go:53-56`** — 截图输出路径无 `ExpandHome`/`IsSafePath` 验证。

**16. `tools/grep.go:170`** — 原生 grep 模式中绝对路径 `dir` 与 `os.DirFS(".")` 不兼容，静默返回无匹配。

### heartbeat 包

**17. `heartbeat/heartbeat.go:107-215`** — `Deps.RunAgent` 函数字段未检查 nil，零值 `Deps` 结构体会导致 panic。

**18. `heartbeat/heartbeat.go:150-240`** — `Start()` 前调用 `Close()` 导致死锁。`m.done` channel 只在 `Start()` 中创建 goroutine 来关闭。

**19. `heartbeat/heartbeat.go:158,236`** — 并发 `Start()`/`Close()` 在 `m.cancel` 上存在数据竞争。

---

## 中优先级（共 36 个）

### 安全相关

- **`session/store.go:45-55,98-101`** — `Store.Load` 和 `Store.Delete` 路径遍历。`id` 参数直接拼接到文件路径，`../../etc/crontab` 可逃逸会话目录。
- **`daemon/attachment.go:59`** — `sessionID` 未经消毒直接拼入附件目录路径。
- **`hooks/hooks.go:297-301`** — 绝对路径 `command` 使用原始字符串前缀检查 `starclawDir`，未调用 `filepath.Clean`，`/home/user/.starclaw/../../bin/sh` 可绕过限制。
- **`daemon/safeguard.go:86-99`** — `chmod` 检查被合并标志绕过（`-Rv` 既不等于 `-r` 也不等于 `-R`）。
- **`audit/redaction.go:14-15`** — JWT 正则表达式 `[A-Za-z0-9_-]*` 在长输入上存在灾难性回溯（ReDoS）风险。
- **`tools/http.go:97`** — HTTP 工具无 SSRF URL 验证，可访问云元数据端点等内部地址。
- **`tools/process.go:99`** — `pkill` 名称未经验证，缺少 `--` 分隔符。
- **`tools/notify.go:83`** — `notify-send` 缺少 `--`，以 `-` 开头的标题被解释为选项。
- **`tools/grep.go:94`** — ripgrep 调用缺少 `--` 分隔符，以 `--` 开头的模式被解释为标志。

### 数据竞争

- **`client.go:119-121, ollama.go:172-173, openai.go:240-241`** — `SetModel` 与 `Chat` 读取模型字段无同步。
- **`cwdctx/cwdctx.go:5-22`** — `CWDContext.Set`/`Get` 无互斥锁。
- **`daemon/server.go:89-91,210-215`** — `SetCancelFunc` 写入 `s.cancel` 与 `handleShutdown` 读取无同步。
- **`daemon/multi_handler.go:37`** — 接口相等比较脆弱，值接收器实现可能错误匹配。
- **`daemon/events.go:56-60`** — `Unsubscribe` 后 channel 保持打开，缓冲事件悬空。
- **`session/index.go:33-87`** — `Index` map 异步访问无互斥锁。
- **`schedule/schedule.go:85-166`** — `load()` 在索引文件上加 SH 锁，`lockedModify()` 在 `.lock` 文件上加 EX 锁，两个锁不协调。

### 逻辑/错误处理

- **`client.go:275`** — `json.Marshal(input)` 错误被忽略。
- **`client/mock.go:118-124`** — `GetLastMessages`/`GetLastTools` 暴露内部切片引用。
- **`client/openai.go:227-228`** — `Complete` 遗漏 `req.Thinking`。
- **`client/ollama.go:160-161`** — `Complete` 遗漏 `req.Thinking` 和 `req.ReasoningEffort`。
- **`client/sse.go:48-57`** — 缺少 `Last-Event-ID` 追踪用于重连。
- **`client/sse.go:73,108-124`** — 失败 SSE 连接的错误响应体被丢弃。
- **`agent/loop.go:474-487`** — spill 失败时跳过截断回退，超大内容留在消息缓冲区。
- **`agent/loop.go:482-486`** / **`agent/toolresult_budget.go:50`** — 字符串在字节位置截断，拆分多字节 UTF-8 字符。
- **`agent/partition_concurrency.go:67`** — 使用 `context.Background()` 而非可取消上下文。
- **`agent/watchdog.go:34-63`** — `AfterFunc` 回调在 `Stop()` 后仍可触发。
- **`agent/loopdetect.go:49-443`** — `LoopDetector` 无互斥锁，且正则表达式每次调用时编译。
- **`agent/spill.go:24`** — `sessionID`/`callID` 直接插入文件名，存在路径遍历。
- **`audit/audit.go:73-74`** — 写入错误被静默丢弃。
- **`audit/audit.go:89-95`** — `truncate` 按字节截断破坏多字节 UTF-8。
- **`context/persist.go:200-203`** — `crypto/rand.Read` 错误被忽略。
- **`context/persist.go:326-328`** — `os.Remove` 错误被静默丢弃。
- **`context/persist.go:63-69`** — `os.ReadFile` 在 `os.MkdirAll` 之前调用。
- **`context/window.go:57-67`** — `ShapeHistory` 假设 `messages[0]` 是系统消息且 `messages[1]` 是用户消息，未验证。
- **`tools/skill.go:158-164`** — 双重互斥锁获取引起 TOCTOU 竞争。
- **`tools/browser.go:190`** — 回退失败时错误变量使用外部 `err` 而非 `fbErr`。
- **`tools/grep.go:182-183`** — 仅按扩展名检测二进制文件，二进制内容文件可能被错误处理。
- **`tools/computer.go:292`** — 未识别修饰键静默默认为 `command`。
- **`schedule/schedule.go:157-160`** — JSON 反序列化错误被静默丢弃，部分解析结果可能覆盖原文件。
- **`permissions/permissions.go:348-354`** — 写入操作始终返回 `Ask`，即使路径在 `AllowedDirs` 中。
- **`permissions/permissions.go:293-306`** — `splitCompound` 不遵守 shell 引用。
- **`mcp/client.go:245-276`** — 传输错误重连绕过 `reconnectMu`，并发重连导致竞态。

---

## 低优先级（共 41 个）

### 被忽略的错误
- `client/sse.go:87-91` — `time.After` 在上下文取消时泄漏定时器
- `client/openai.go`/`ollama.go`/`client.go` — `Chat` 可能返回 `(nil, nil)`，`Complete` 访问 `resp.Content` 时空指针
- `client` 所有构造函数 — 缺少必需配置（空 apiKey/baseURL）验证
- `audit/audit.go` — 写入错误被忽略
- `hooks/hooks.go:322-324` — `json.Marshal` 错误被忽略
- `schedule/schedule.go:277-278` — `rand.Read` 错误被忽略
- `session/store.go:20` — `MkdirAll` 错误被忽略
- `daemon/approval.go:87` / `daemon/server.go:859` — `rand.Read` 错误被忽略
- `daemon/attachment.go:70` — `out.Close()` 错误被 `defer` 丢弃
- `daemon/server.go:823-833` — `writeJSON`/`writeError` 忽略 `Encode` 错误

### 逻辑/设计问题
- `agent/loop.go:210-264` — `lastRunStatus` 无同步访问
- `agent/spill.go:50` — `os.Remove` 对非空目录静默失败
- `agent/partition.go:139` — JSON 解析错误被静默丢弃
- `agent/loop.go:144-148` — `SwitchAgent` 无限追加 system prompt
- `agent/phase.go:7` — `import "testing"` 在生产代码中
- `agent/toolbudget.go:28-34` — 负数 chars 处理不一致
- `agent/timebasedcompact.go:33` — 首次压缩延迟 `maxAge`
- `agent/approval_cache.go:25-31` — map 异步访问无锁（潜在）
- `agent/loop.go:393-395` — `LoopForceStop` 未提前终止工具循环
- `client/ollama.go/openai.go/client.go` — 无界请求体序列化
- `cmd/root.go:465-520` — CLI 参数未验证直接传入 session manager（路径遍历风险）
- `cmd/schedule.go:63-105` — 调度 prompt 无长度验证
- `cmd/daemon.go:80-82` — 端口硬编码为 7533
- `cmd/root.go:282-287` — `isPathReference` 对以 `/` 开头的输入过于宽松
- `context/persist.go:326-328` — 合并清理时 `os.Remove` 错误被丢弃
- `skills/loader.go:75-79` — 目录读取错误静默跳过
- `tools/http.go:108` — 每次请求创建新 `http.Client`，无连接复用
- `tools/imaging.go:67` — 冗余验证检查
- `tools/wait.go:53` — `time.After` 在取消时泄漏定时器
- `daemon/server.go:504-511` — 重复 `json.Unmarshal`（复制粘贴错误）
- `daemon/server.go:320-327` — 基于 `strings.Contains` 的错误分类不稳健
- `daemon/rules.go:28,66` — `agentName` 路径遍历
- `daemon/session_cache.go:176-189` — `routeEntry` 永不过期（内存泄漏）
- `daemon/scheduler.go:129` — 调度运行使用与调度器主 context 相同的取消信号
- `session/manager.go:91-98` — `ResumeLatest` 按 `CreatedAt` 排序而非 `UpdatedAt`
- `mcp/supervisor.go:61-68` — `Stop()` 未调用 `wg.Wait()`，`Start()` 重新调用时可能产生双重循环
- `mcp/readiness.go:23` — 传入 nil manager 时空指针
- `permissions/permissions.go:406-418` — `expandHome` 不处理 `~otheruser`

---

## 按包分布

| 包 | 严重 | 高 | 中 | 低 | 总计 |
|---------|----------|------|--------|-----|-------|
| `agent/` | 1 | 3 | 7 | 9 | 20 |
| `client/` | 1 | 2 | 6 | 4 | 13 |
| `tools/` | — | 6 | 7 | 5 | 18 |
| `daemon/` | 1 | 3 | 6 | 6 | 16 |
| `cmd/audit/heartbeat/context/cwdctx/skills/prompt/instructions/runstatus/` | 2 | 3 | 8 | 7 | 20 |
| `session/mcp/hooks/permissions/schedule/` | — | 2 | 6 | 7 | 15 |

---

## 建议修复优先级

1. **立即** — 5 个严重级别（数据丢失、重复 API 调用、路径遍历写入）
2. **本周** — 8 个路径遍历/安全绕过（安全路径符号链接、会话/附件 ID、钩子命令解析）
3. **下周** — 跨 agent/client/tools/daemon 的 7+ 个数据竞争
4. **持续** — 改善错误处理纪律（数十处 I/O 和加密操作中 `_` 丢弃返回值）
