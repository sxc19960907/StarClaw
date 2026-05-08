# SessionCache - Session Manager Pool

## Goal

实现 session.Manager 的缓存池，为 daemon 的 router 和 heartbeat 提供会话上下文管理。

## Requirements

- `SessionCache` 结构体：routes map + managers map + starclawDir
- `NewSessionCache(starclawDir) *SessionCache`
- `GetOrCreate(agent string) *session.Manager` — 按 agent 获取或创建 session.Manager
- `GetOrCreateManager(sessionsDir string) *session.Manager` — 按目录获取
- `SessionsDir(agent string) string` — 返回 agent 的 sessions 目录
- `ResolveLatestSession(routeKey, sessionsDir) (sessionID, messages, error)` — 获取最新会话
- `AppendToSession(routeKey, sessionsDir, sessionID, messages) error` — 追加消息到会话
- `LockRoute(key)`, `UnlockRoute(key)` — 路由锁
- TryLock 防重入
- routeEntry: mu (sync.Mutex), manager, cancel, done, lastAccess

## Acceptance Criteria

- [ ] 编译通过，集成到 internal/daemon/
- [ ] 单元测试覆盖：create, getOrCreate cache hit, resolve latest, append, lock/unlock, error on missing
- [ ] 不依赖 networking，纯本地测试

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/router.go` (SessionCache 部分)
- StarClaw 现有: `internal/session/Manager` — NewManager, ResumeLatest, NewSession, Save
- 文件: `internal/daemon/session_cache.go` + `session_cache_test.go`
