# Specification: http Tool

## Overview

| Field | Value |
|-------|-------|
| Name | `http` |
| Type | Local Tool |
| Approval Required | Yes (configurable) |
| Safe Operations | GET to localhost/127.0.0.1 |

## Purpose

Make HTTP requests to APIs, local services, or web resources. Supports common HTTP methods with configurable headers, body, and timeout.

## Interface

### Input Parameters

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | Yes | - | Request URL |
| `method` | string | No | `GET` | HTTP method (GET, POST, PUT, DELETE) |
| `headers` | object | No | `{}` | Request headers as key-value pairs |
| `body` | string | No | `""` | Request body (for POST/PUT) |
| `timeout` | integer | No | `30` | Timeout in seconds |

### Input Schema

```json
{
  "type": "object",
  "properties": {
    "url": {
      "type": "string",
      "description": "Request URL"
    },
    "method": {
      "type": "string",
      "description": "HTTP method (default: GET)"
    },
    "headers": {
      "type": "object",
      "description": "Request headers as key-value pairs"
    },
    "body": {
      "type": "string",
      "description": "Request body"
    },
    "timeout": {
      "type": "integer",
      "description": "Timeout in seconds (default: 30)"
    }
  },
  "required": ["url"]
}
```

### Output

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | Formatted response |
| `is_error` | boolean | `true` if network/request error |

### Output Format

```
Status: 200 OK

Headers:
  Content-Type: application/json
  Content-Length: 1234

Body:
{...response body...}
```

### Response Body Truncation

Response body is limited to 10KB. If truncated:
```
Body:
{...first 10KB...}
[...truncated...]
```

## Error Handling

| Error Type | Condition | Result Category |
|------------|-----------|-----------------|
| Validation | Invalid URL | `validation` |
| Transient | Network timeout | `transient` |
| Transient | Connection refused | `transient` |
| Business | HTTP 4xx/5xx | `business` (not error) |

## Safety Controls

### Auto-Approve Conditions

A request is auto-approved (no user prompt) when:
1. Method is GET (case-insensitive)
2. Host is `localhost` or `127.0.0.1`

All other requests require explicit approval.

### Examples

| URL | Method | Auto-Approve |
|-----|--------|--------------|
| `http://localhost:8080/api` | GET | ✅ Yes |
| `http://127.0.0.1:3000` | GET | ✅ Yes |
| `http://localhost/api` | POST | ❌ No (not GET) |
| `http://example.com` | GET | ❌ No (not localhost) |

## Examples

### Example 1: GET Request

**Input:**
```json
{
  "url": "http://localhost:8080/health",
  "method": "GET"
}
```

**Output:**
```
Status: 200 OK

Headers:
  Content-Type: application/json

Body:
{"status": "healthy", "uptime": 3600}
```

### Example 2: POST with Body

**Input:**
```json
{
  "url": "http://api.example.com/users",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token123"
  },
  "body": "{\"name\": \"Alice\", \"email\": \"alice@example.com\"}"
}
```

**Output:**
```
Status: 201 Created

Headers:
  Content-Type: application/json
  Location: /users/123

Body:
{"id": 123, "name": "Alice", "email": "alice@example.com"}
```

### Example 3: Timeout

**Input:**
```json
{
  "url": "http://slow.example.com/data",
  "timeout": 5
}
```

**Output:** (if timeout occurs)
```
[transient error] request failed: context deadline exceeded
```

## Security Considerations

1. **Exfiltration risk**: HTTP tool could leak data to external servers
   - Mitigation: Approval required (except localhost GET)

2. **Local service scanning**: Could probe local ports
   - Mitigation: GET to localhost is safe; POST/PUT still requires approval

3. **Credential exposure**: Headers may contain auth tokens
   - Mitigation: Audit logging should redact common auth header patterns

## Configuration

Optional config support:

```yaml
tools:
  http_timeout: 30  # Default timeout in seconds
```

## Design Rationale

- **Common methods only**: GET/POST/PUT/DELETE cover most use cases
- **Header support**: Essential for API authentication
- **Body as string**: Flexible for JSON, XML, plain text
- **Timeout configurable**: Different APIs have different SLAs
- **10KB truncation**: Prevent overwhelming context with large responses
