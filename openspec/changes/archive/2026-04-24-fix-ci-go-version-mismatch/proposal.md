## Why

CI 流水线持续失败，原因是 Go 版本不匹配。项目 `go.mod` 指定了 `toolchain go1.26.1`，但 CI 配置使用 Go 1.24，导致 `golangci-lint` 无法运行且覆盖率工具 `covdata` 缺失。这阻碍了代码合并和发布流程。

## What Changes

- 更新 `.github/workflows/ci.yml` 中的 Go 版本从 1.24 到 1.26
- 更新所有测试矩阵中的 Go 版本（1.22, 1.23, 1.24 → 1.24, 1.25, 1.26）
- 同步 `lint` job 的 Go 版本到 1.26
- 验证修复后 CI 能通过 lint 和 test 阶段

## Capabilities

### New Capabilities
- 无（这是配置修复，不涉及新功能）

### Modified Capabilities
- 无（这是配置修复，不涉及需求变更）

## Impact

- `.github/workflows/ci.yml` - CI 配置文件
- 所有后续 PR 的 CI 检查将恢复正常
- 无代码逻辑变更，无 API 变更
