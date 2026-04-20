# Design: System Tools

## Overview

Two tools to enhance system interaction:
1. `system_info` - Get system information (cross-platform)
2. `http` - Make HTTP requests with safety controls

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     System Tools                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────┐         ┌───────────────┐               │
│  │ SystemInfoTool│         │   HTTPTool    │               │
│  │               │         │               │               │
│  │ - OS/Arch     │         │ - GET/POST    │               │
│  │ - Hostname    │         │ - Headers     │               │
│  │ - CPUs        │         │ - Body        │               │
│  │ - Memory*     │         │ - Timeout     │               │
│  │ - Disk*       │         │               │               │
│  └───────┬───────┘         └───────┬───────┘               │
│          │                         │                       │
│          ▼                         ▼                       │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ Platform-Specific│   │  SafeChecker    │                │
│  │  Implementations │   │  (localhost GET)│                │
│  │                  │   └─────────────────┘                │
│  │  - Linux         │                                      │
│  │  - Darwin        │                                      │
│  │  - Windows       │                                      │
│  │  - Other         │                                      │
│  └─────────────────┘                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘

* Platform-specific
```

## SystemInfoTool Design

### Core Implementation

```go
// internal/tools/system_info.go
package tools

type SystemInfoTool struct{}

func (t *SystemInfoTool) Info() agent.ToolInfo {
    return agent.ToolInfo{
        Name:        "system_info",
        Description: "Get system information: OS, architecture, hostname, CPU count, memory, and disk usage.",
        Parameters: map[string]any{
            "type":       "object",
            "properties": map[string]any{},
        },
        Required: nil,
    }
}

func (t *SystemInfoTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
    hostname, _ := os.Hostname()
    
    var sb strings.Builder
    fmt.Fprintf(&sb, "OS: %s\n", runtime.GOOS)
    fmt.Fprintf(&sb, "Arch: %s\n", runtime.GOARCH)
    fmt.Fprintf(&sb, "Hostname: %s\n", hostname)
    fmt.Fprintf(&sb, "CPUs: %d\n", runtime.NumCPU())
    
    // Platform-specific info
    if memInfo := getMemoryInfo(); memInfo != "" {
        fmt.Fprintf(&sb, "\nMemory:\n%s", memInfo)
    }
    if diskInfo := getDiskInfo(); diskInfo != "" {
        fmt.Fprintf(&sb, "\nDisk:\n%s", diskInfo)
    }
    
    return agent.ToolResult{Content: sb.String()}, nil
}

func (t *SystemInfoTool) RequiresApproval() bool { return false }
func (t *SystemInfoTool) IsReadOnlyCall(string) bool { return true }

// Platform-specific functions (implemented in platform files)
var getMemoryInfo func() string
var getDiskInfo func() string
```

### Platform-Specific Implementations

```go
// internal/tools/system_info_linux.go
//go:build linux
package tools

import (
    "os"
    "strconv"
    "strings"
)

func init() {
    getMemoryInfo = getLinuxMemoryInfo
    getDiskInfo = getLinuxDiskInfo
}

func getLinuxMemoryInfo() string {
    data, err := os.ReadFile("/proc/meminfo")
    if err != nil {
        return ""
    }
    
    var total, available uint64
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "MemTotal:") {
            total = parseMeminfoKB(line)
        } else if strings.HasPrefix(line, "MemAvailable:") {
            available = parseMeminfoKB(line)
        }
    }
    
    if total == 0 {
        return ""
    }
    
    return fmt.Sprintf("  Total: %d MB\n  Available: %d MB\n", 
        total/1024, available/1024)
}

func getLinuxDiskInfo() string {
    // Use df -h .
    cmd := exec.Command("df", "-h", ".")
    out, err := cmd.Output()
    if err != nil {
        return ""
    }
    return string(out)
}
```

```go
// internal/tools/system_info_darwin.go
//go:build darwin
package tools

func init() {
    getMemoryInfo = getDarwinMemoryInfo
    getDiskInfo = getDarwinDiskInfo
}

func getDarwinMemoryInfo() string {
    // Use vm_stat
    cmd := exec.Command("vm_stat")
    out, err := cmd.Output()
    if err != nil {
        return ""
    }
    // Parse vm_stat output
    // ... implementation ...
}
```

```go
// internal/tools/system_info_other.go
//go:build !linux && !darwin
package tools

func init() {
    getMemoryInfo = func() string { return "" }
    getDiskInfo = func() string { return "" }
}
```

## HTTPTool Design

### Implementation

```go
// internal/tools/http.go
package tools

type HTTPTool struct {
    DefaultTimeout int // From config, or 30
}

type httpArgs struct {
    URL     string            `json:"url"`
    Method  string            `json:"method,omitempty"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    string            `json:"body,omitempty"`
    Timeout int               `json:"timeout,omitempty"`
}

func (t *HTTPTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
    var args httpArgs
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
    }
    
    // Set defaults
    method := strings.ToUpper(args.Method)
    if method == "" {
        method = "GET"
    }
    
    timeout := 30 * time.Second
    if t.DefaultTimeout > 0 {
        timeout = time.Duration(t.DefaultTimeout) * time.Second
    }
    if args.Timeout > 0 {
        timeout = time.Duration(args.Timeout) * time.Second
    }
    
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // Build request
    var bodyReader io.Reader
    if args.Body != "" {
        bodyReader = strings.NewReader(args.Body)
    }
    
    req, err := http.NewRequestWithContext(ctx, method, args.URL, bodyReader)
    if err != nil {
        return agent.ValidationError(fmt.Sprintf("invalid request: %v", err)), nil
    }
    
    // Add headers
    for k, v := range args.Headers {
        req.Header.Set(k, v)
    }
    
    // Execute
    client := &http.Client{Timeout: timeout}
    resp, err := client.Do(req)
    if err != nil {
        return agent.TransientError(fmt.Sprintf("request failed: %v", err)), nil
    }
    defer resp.Body.Close()
    
    // Read body (limited to 10KB)
    body, err := io.ReadAll(io.LimitReader(resp.Body, 10240))
    if err != nil {
        return agent.TransientError(fmt.Sprintf("error reading body: %v", err)), nil
    }
    
    // Format response
    var sb strings.Builder
    fmt.Fprintf(&sb, "Status: %d %s\n\nHeaders:\n", resp.StatusCode, resp.Status)
    for k, vals := range resp.Header {
        for _, v := range vals {
            fmt.Fprintf(&sb, "  %s: %s\n", k, v)
        }
    }
    fmt.Fprintf(&sb, "\nBody:\n%s", string(body))
    
    return agent.ToolResult{Content: sb.String()}, nil
}

func (t *HTTPTool) RequiresApproval() bool { return true }

func (t *HTTPTool) IsSafeArgs(argsJSON string) bool {
    var args httpArgs
    if err := json.Unmarshal([]byte(argsJSON), &err); err != nil {
        return false
    }
    
    // Must be GET
    method := strings.ToUpper(args.Method)
    if method == "" {
        method = "GET"
    }
    if method != "GET" {
        return false
    }
    
    // Must be localhost
    parsed, err := url.Parse(args.URL)
    if err != nil {
        return false
    }
    
    host := parsed.Hostname()
    return host == "localhost" || host == "127.0.0.1"
}
```

## Testing Strategy

### SystemInfoTool Tests

```go
// internal/tools/system_info_test.go

func TestSystemInfoTool_BasicInfo(t *testing.T) {
    tool := &SystemInfoTool{}
    result, err := tool.Run(context.Background(), "{}")
    
    require.NoError(t, err)
    require.False(t, result.IsError)
    require.Contains(t, result.Content, "OS:")
    require.Contains(t, result.Content, "Arch:")
    require.Contains(t, result.Content, "Hostname:")
    require.Contains(t, result.Content, "CPUs:")
}

func TestSystemInfoTool_NoApproval(t *testing.T) {
    tool := &SystemInfoTool{}
    require.False(t, tool.RequiresApproval())
}
```

### HTTPTool Tests

```go
// internal/tools/http_test.go

func TestHTTPTool_GET(t *testing.T) {
    // Start mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "GET", r.Method)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()
    
    tool := &HTTPTool{}
    result, err := tool.Run(context.Background(), fmt.Sprintf(`{"url": "%s"}`, server.URL))
    
    require.NoError(t, err)
    require.False(t, result.IsError)
    require.Contains(t, result.Content, "200")
    require.Contains(t, result.Content, `"status": "ok"`)
}

func TestHTTPTool_SafeChecker_Localhost(t *testing.T) {
    tool := &HTTPTool{}
    
    // Safe: GET to localhost
    assert.True(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "GET"}`))
    assert.True(t, tool.IsSafeArgs(`{"url": "http://127.0.0.1:3000"}`))
    
    // Not safe: POST to localhost
    assert.False(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "POST"}`))
    
    // Not safe: GET to external
    assert.False(t, tool.IsSafeArgs(`{"url": "http://example.com", "method": "GET"}`))
}
```

## Integration Points

### Registration

Both tools added to `RegisterLocalTools()`:

```go
func RegisterLocalTools(cfg *config.Config) *agent.ToolRegistry {
    reg := agent.NewToolRegistry()
    
    // ... existing tools ...
    
    reg.Register(&SystemInfoTool{})
    reg.Register(&HTTPTool{DefaultTimeout: cfg.Tools.HTTPTimeout})
    
    // ... bash tool ...
    
    return reg
}
```

### Config (Optional)

Add to config:

```go
type ToolsConfig struct {
    // ... existing fields ...
    HTTPTimeout int `mapstructure:"http_timeout" yaml:"http_timeout"`
}
```

## Platform Considerations

| Platform | Memory | Disk | Notes |
|----------|--------|------|-------|
| Linux | /proc/meminfo | df -h | Full support |
| macOS | vm_stat | df -h | Full support |
| Windows | WMI | WMI | Basic support |
| Other | N/A | N/A | Graceful degradation |

## Error Handling

| Tool | Error | Category | Behavior |
|------|-------|----------|----------|
| system_info | Any | None | Omit info, continue |
| http | Invalid URL | Validation | Return validation error |
| http | Network error | Transient | Return transient error (retryable) |
| http | Timeout | Transient | Return transient error |
