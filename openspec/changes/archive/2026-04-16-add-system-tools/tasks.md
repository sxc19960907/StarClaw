# Tasks: Add System Tools

## Task 1: Create SystemInfoTool Core

**ID**: T1  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Create `SystemInfoTool` with basic cross-platform info (OS, Arch, Hostname, CPUs).

### Acceptance Criteria
- [x] `system_info.go` created with `SystemInfoTool` struct
- [x] Returns OS, Arch, Hostname, NumCPU
- [x] `RequiresApproval()` returns `false`
- [x] `IsReadOnlyCall()` returns `true`
- [x] Works on all platforms (Linux, macOS, Windows)

### Notes
Use `runtime` and `os` packages for portable info.

---

## Task 2: Add Platform-Specific Memory/Disk Info

**ID**: T2  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Add memory and disk information using build tags for platform-specific implementations.

### Acceptance Criteria
- [x] `system_info_linux.go` with /proc/meminfo parsing
- [x] `system_info_darwin.go` with vm_stat
- [x] `system_info_windows.go` with basic support (or stub)
- [x] `system_info_other.go` as fallback
- [x] Memory info shown on Linux/macOS
- [x] Disk info shown where supported
- [x] Graceful degradation on unsupported platforms

### Test Strategy
- Mock exec.Command for testing
- Test parsing logic independently

---

## Task 3: Write SystemInfoTool Unit Tests

**ID**: T3  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Create `system_info_test.go` with comprehensive tests.

### Acceptance Criteria
- [x] Test basic info fields are present
- [x] Test parsing functions for memory/disk
- [x] Test error handling for missing tools
- [x] All tests pass on CI (Linux/macOS/Windows)

---

## Task 4: Create HTTPTool

**ID**: T4  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Create `HTTPTool` with GET/POST/PUT/DELETE support.

### Acceptance Criteria
- [x] `http.go` created with `HTTPTool` struct
- [x] Supports GET, POST, PUT, DELETE
- [x] Supports custom headers
- [x] Supports request body
- [x] Configurable timeout (default 30s)
- [x] Response body truncated to 10KB
- [x] Returns status, headers, body
- [x] `RequiresApproval()` returns `true`

---

## Task 5: Implement SafeChecker for HTTP

**ID**: T5  
**Status**: completed  
**Owner**:  
**Blocked By**: T4

### Description
Add `IsSafeArgs()` method for auto-approving localhost GET requests.

### Acceptance Criteria
- [x] `IsSafeArgs()` implemented
- [x] Returns `true` for GET to localhost
- [x] Returns `true` for GET to 127.0.0.1
- [x] Returns `false` for non-GET methods
- [x] Returns `false` for external URLs
- [x] Handles invalid URLs safely

---

## Task 6: Write HTTPTool Unit Tests

**ID**: T6  
**Status**: completed  
**Owner**:  
**Blocked By**: T4, T5

### Description
Create `http_test.go` with mocked HTTP server tests.

### Acceptance Criteria
- [x] Test GET request
- [x] Test POST with body
- [x] Test custom headers
- [x] Test timeout
- [x] Test response body truncation
- [x] Test SafeChecker scenarios
- [x] Mock server for predictable responses
- [x] All tests pass

---

## Task 7: Register Both Tools

**ID**: T7  
**Status**: completed  
**Owner**:  
**Blocked By**: T1, T4

### Description
Add both tools to `RegisterLocalTools()`.

### Acceptance Criteria
- [x] `&SystemInfoTool{}` registered
- [x] `&HTTPTool{}` registered
- [x] Order: after file tools, before bash
- [x] Imports added

---

## Task 8: Write Integration Tests

**ID**: T8  
**Status**: completed  
**Owner**:  
**Blocked By**: T7

### Description
Add integration tests to verify tool registration.

### Acceptance Criteria
- [x] `system_info` in registry
- [x] `http` in registry
- [x] Tool schemas valid
- [x] Tests pass

---

## Task 9: Add Configuration for HTTP Timeout

**ID**: T9  
**Status**: pending  
**Owner**:  
**Blocked By**: T4

### Description
Optionally read default HTTP timeout from config.

### Acceptance Criteria
- [ ] Config supports `tools.http_timeout`
- [ ] HTTPTool reads config if provided
- [ ] Falls back to 30s if not configured

---

## Task 10: Cross-Platform CI Testing

**ID**: T10  
**Status**: pending  
**Owner**:  
**Blocked By**: T3, T6

### Description
Ensure tests pass on Linux, macOS, and Windows.

### Acceptance Criteria
- [ ] Linux tests pass
- [ ] macOS tests pass
- [ ] Windows tests pass (with stubs if needed)

---

## Dependencies

```
T1 (SystemInfo Core)
  └── T2 (Platform-specific)
  └── T3 (SystemInfo Tests)

T4 (HTTPTool)
  └── T5 (SafeChecker)
  └── T6 (HTTP Tests)
  └── T9 (Config - optional)

T1 + T4 ──▶ T7 (Register)
              └── T8 (Integration)

T3 + T6 + T8 ──▶ T10 (CI)
```

## Estimated Effort

- T1: 30 minutes
- T2: 1.5 hours (platform complexity)
- T3: 1 hour
- T4: 1 hour
- T5: 30 minutes
- T6: 1.5 hours (mock server)
- T7: 15 minutes
- T8: 30 minutes
- T9: 30 minutes (optional)
- T10: Ongoing (CI setup)

**Total**: ~7 hours
