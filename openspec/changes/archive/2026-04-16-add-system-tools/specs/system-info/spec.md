# Specification: system_info Tool

## Overview

| Field | Value |
|-------|-------|
| Name | `system_info` |
| Type | Local Tool |
| Approval Required | No |
| Read-Only | Yes |

## Purpose

Provides portable system information across different operating systems. Useful for understanding the execution environment before taking action.

## Interface

### Input Parameters

None. Tool takes no arguments.

### Input Schema

```json
{
  "type": "object",
  "properties": {}
}
```

### Output

Plain text output with the following sections:

| Field | Description | Example |
|-------|-------------|---------|
| OS | Operating system name | `linux`, `darwin`, `windows` |
| Arch | CPU architecture | `amd64`, `arm64` |
| Hostname | System hostname | `my-laptop` |
| CPUs | Number of CPU cores | `8` |
| Memory | Memory info (platform-specific) | `Total: 16384 MB` |
| Disk | Disk usage (platform-specific) | `Used: 45%, Free: 100GB` |

### Output Format

```
OS: darwin
Arch: arm64
Hostname: my-macbook
CPUs: 8

Memory:
  Total: 16384 MB
  Available: 8192 MB

Disk:
Filesystem      Size   Used  Avail  Use%
/dev/disk1     500G   200G   300G   40%
```

## Platform Support

| Platform | OS | Arch | Hostname | CPUs | Memory | Disk |
|----------|----|------|----------|------|--------|------|
| Linux | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| macOS | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Windows | ✅ | ✅ | ✅ | ✅ | ⚠️ | ⚠️ |
| Other | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |

Legend: ✅ Full support, ⚠️ Partial, ❌ Not supported

## Implementation Notes

### Memory Info

- **Linux**: Parse `/proc/meminfo`
- **macOS**: Parse `vm_stat` output
- **Windows**: `syscall.GlobalMemoryStatusEx`
- **Other**: Not available

### Disk Info

- **Linux/macOS**: `df -h .` command
- **Windows**: `GetDiskFreeSpaceEx` or WMI
- **Other**: Not available

## Examples

### Example 1: macOS

**Input:** `{}`

**Output:**
```
OS: darwin
Arch: arm64
Hostname: macbook-pro
CPUs: 8

Memory:
  Total: 16384 MB
  Available: 4096 MB

Disk:
Filesystem      Size   Used  Avail  Use%
/dev/disk1s1   500Gi  200Gi  300Gi   40%
```

### Example 2: Linux

**Input:** `{}`

**Output:**
```
OS: linux
Arch: amd64
Hostname: ubuntu-server
CPUs: 4

Memory:
  Total: 8192 MB
  Free: 2048 MB

Disk:
Filesystem      Size   Used  Avail  Use%
/dev/sda1        50G   25G    25G   50%
```

## Error Handling

Tool should never fail - gracefully omit information that cannot be retrieved.

## Design Rationale

- **No parameters**: Simple, always-available tool
- **Plain text output**: Human-readable, easy for AI to parse
- **Graceful degradation**: Works on all platforms, best-effort for extras
