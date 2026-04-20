package tools

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemInfoTool_Info(t *testing.T) {
	tool := &SystemInfoTool{}
	info := tool.Info()

	assert.Equal(t, "system_info", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Description, "OS")
	assert.Contains(t, info.Parameters, "type")
	assert.Nil(t, info.Required)
}

func TestSystemInfoTool_RequiresApproval(t *testing.T) {
	tool := &SystemInfoTool{}
	assert.False(t, tool.RequiresApproval())
}

func TestSystemInfoTool_IsReadOnlyCall(t *testing.T) {
	tool := &SystemInfoTool{}
	assert.True(t, tool.IsReadOnlyCall("{}"))
	assert.True(t, tool.IsReadOnlyCall(""))
}

func TestSystemInfoTool_Run_BasicInfo(t *testing.T) {
	tool := &SystemInfoTool{}
	result, err := tool.Run(context.Background(), "{}")

	require.NoError(t, err)
	assert.False(t, result.IsError)

	content := result.Content

	// Check for basic fields that should always be present
	assert.Contains(t, content, "OS:")
	assert.Contains(t, content, "Arch:")
	assert.Contains(t, content, "Hostname:")
	assert.Contains(t, content, "CPUs:")

	// Verify OS matches runtime
	assert.Contains(t, content, runtime.GOOS)
	assert.Contains(t, content, runtime.GOARCH)
}

func TestSystemInfoTool_Run_NoArgs(t *testing.T) {
	tool := &SystemInfoTool{}
	result, err := tool.Run(context.Background(), "")

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "OS:")
}

func TestSystemInfoTool_Run_InvalidJSON(t *testing.T) {
	tool := &SystemInfoTool{}
	// Invalid JSON should be ignored (tool takes no arguments)
	result, err := tool.Run(context.Background(), "not valid json")

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "OS:")
}

func TestSystemInfoTool_CPUs(t *testing.T) {
	tool := &SystemInfoTool{}
	result, err := tool.Run(context.Background(), "{}")

	require.NoError(t, err)

	// Check CPUs field contains the expected count
	expectedCPUs := runtime.NumCPU()
	assert.Contains(t, result.Content, fmt.Sprintf("CPUs: %d", expectedCPUs))
}
