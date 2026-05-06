# Daemon Sub4: Client - HTTP Client for Daemon

## Goal

实现 HTTP 客户端，供 CLI 与 daemon 通信。

## Requirements

- `Client` 结构体：baseURL, http.Client
- 方法：Health(), Status(), Message(req), Cancel(req), Shutdown()
- Schedule CRUD：ListSchedules(), GetSchedule(), CreateSchedule(), PatchSchedule(), DeleteSchedule()
- JSON 序列化/反序列化请求和响应
- 错误处理：非 2xx 响应视为错误

## Acceptance Criteria

- [ ] 所有方法可以正确发出 HTTP 请求
- [ ] 单元测试用 httptest 模拟服务端验证

## Technical Notes

- 参考: `/Users/timmy/PycharmProjects/ShanClaw/internal/daemon/client.go` (421行)
- 使用 net/http/httptest 做测试
