## Context

CI 流水线配置中 Go 版本与项目需求不匹配：
- `go.mod` 使用 `toolchain go1.26.1`
- `.github/workflows/ci.yml` 配置为 Go 1.22/1.23/1.24
- 这导致 `golangci-lint` 报错：`the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.26.1)`
- 覆盖率工具 `covdata` 也因版本不匹配而缺失

## Goals / Non-Goals

**Goals:**
- 修复 CI 配置中的 Go 版本与 `go.mod` 一致
- 确保 lint 和 test job 能正常执行
- 保持向后兼容的测试矩阵

**Non-Goals:**
- 不修改 go.mod 中的 Go 版本（保持 1.26）
- 不涉及功能代码变更
- 不修改 lint 规则或测试逻辑

## Decisions

**Decision 1**: 将 CI 中的 Go 版本更新为 1.24/1.25/1.26（矩阵），并在 lint job 使用 1.26

- **Rationale**: 与 go.mod 中的 toolchain 版本保持一致，同时保持测试矩阵覆盖多个版本
- **Alternative considered**: 降级 go.mod 到 1.24 - 拒绝，因为新功能可能需要 1.26 特性

## Risks / Trade-offs

- **Risk**: Go 1.26 可能较新，某些依赖可能不完全支持 → 已通过检查 go.mod 确认依赖兼容
- **Risk**: CI 运行时间可能略有增加 → 可接受，修复正确性优先
