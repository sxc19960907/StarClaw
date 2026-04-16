package agent

import (
	"context"
)

// ToolInfo describes a tool for the LLM
type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]any
	Required    []string
}

// ErrorCategory classifies tool errors
type ErrorCategory string

const (
	ErrCategoryTransient  ErrorCategory = "transient"
	ErrCategoryValidation ErrorCategory = "validation"
	ErrCategoryBusiness   ErrorCategory = "business"
	ErrCategoryPermission ErrorCategory = "permission"
)

// ToolSource classifies tool origin
type ToolSource string

const (
	SourceLocal   ToolSource = "local"
	SourceMCP     ToolSource = "mcp"
	SourceGateway ToolSource = "gateway"
)

// ToolSourcer is optional interface for tools to declare origin
type ToolSourcer interface {
	ToolSource() ToolSource
}

// ToolResult contains tool execution result
type ToolResult struct {
	Content       string
	IsError       bool
	ErrorCategory ErrorCategory
	IsRetryable   bool
}

// TransientError returns a transient error result
func TransientError(msg string) ToolResult {
	return ToolResult{
		Content:       "[transient error] " + msg,
		IsError:       true,
		ErrorCategory: ErrCategoryTransient,
		IsRetryable:   true,
	}
}

// ValidationError returns a validation error result
func ValidationError(msg string) ToolResult {
	return ToolResult{
		Content:       "[validation error] " + msg,
		IsError:       true,
		ErrorCategory: ErrCategoryValidation,
	}
}

// BusinessError returns a business error result
func BusinessError(msg string) ToolResult {
	return ToolResult{
		Content:       "[business error] " + msg,
		IsError:       true,
		ErrorCategory: ErrCategoryBusiness,
	}
}

// PermissionError returns a permission error result
func PermissionError(msg string) ToolResult {
	return ToolResult{
		Content:       "[permission error] " + msg,
		IsError:       true,
		ErrorCategory: ErrCategoryPermission,
	}
}

// Tool is the interface for all tools
type Tool interface {
	Info() ToolInfo
	Run(ctx context.Context, args string) (ToolResult, error)
	RequiresApproval() bool
}

// SafeChecker is optional interface for tools to indicate safe arguments
type SafeChecker interface {
	IsSafeArgs(argsJSON string) bool
}

// ReadOnlyChecker is optional interface for tools to indicate read-only calls
type ReadOnlyChecker interface {
	IsReadOnlyCall(argsJSON string) bool
}
