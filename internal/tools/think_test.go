package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThinkTool_Info(t *testing.T) {
	tool := &ThinkTool{}
	info := tool.Info()

	assert.Equal(t, "think", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Description, "plan")
	assert.Contains(t, info.Parameters, "type")
	assert.Contains(t, info.Parameters, "properties")
	assert.Equal(t, []string{"thought"}, info.Required)
}

func TestThinkTool_RequiresApproval(t *testing.T) {
	tool := &ThinkTool{}
	assert.False(t, tool.RequiresApproval())
}

func TestThinkTool_IsReadOnlyCall(t *testing.T) {
	tool := &ThinkTool{}
	assert.True(t, tool.IsReadOnlyCall("{}"))
	assert.True(t, tool.IsReadOnlyCall(`{"thought": "test"}`))
}

func TestThinkTool_Run_ValidThought(t *testing.T) {
	tool := &ThinkTool{}
	result, err := tool.Run(context.Background(), `{"thought": "I need to plan this task"}`)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "I need to plan this task", result.Content)
}

func TestThinkTool_Run_EmptyThought(t *testing.T) {
	tool := &ThinkTool{}
	result, err := tool.Run(context.Background(), `{"thought": ""}`)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "thought is required")
}

func TestThinkTool_Run_MissingThought(t *testing.T) {
	tool := &ThinkTool{}
	result, err := tool.Run(context.Background(), `{}`)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "thought is required")
}

func TestThinkTool_Run_InvalidJSON(t *testing.T) {
	tool := &ThinkTool{}
	result, err := tool.Run(context.Background(), "not valid json")

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestThinkTool_Run_LargeThought(t *testing.T) {
	tool := &ThinkTool{}
	// Create a 2000 character thought
	largeThought := strings.Repeat("This is a test sentence. ", 80)
	require.True(t, len(largeThought) > 1000)

	args := `{"thought": "` + largeThought + `"}`
	result, err := tool.Run(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, largeThought, result.Content)
}

func TestThinkTool_Run_MultilineThought(t *testing.T) {
	tool := &ThinkTool{}
	// Newlines in JSON must be escaped as \n
	thought := "Step 1: Analyze the code\nStep 2: Find patterns\nStep 3: Refactor"
	args := `{"thought": "Step 1: Analyze the code\nStep 2: Find patterns\nStep 3: Refactor"}`

	result, err := tool.Run(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, thought, result.Content)
}

func TestThinkTool_Run_SpecialCharacters(t *testing.T) {
	tool := &ThinkTool{}
	args := `{"thought": "Use \u0060code\u0060 and \"quotes\" and \\backslashes"}`

	result, err := tool.Run(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "code")
}

func TestThinkTool_Run_ReturnsToolResult(t *testing.T) {
	tool := &ThinkTool{}
	result, err := tool.Run(context.Background(), `{"thought": "test"}`)

	require.NoError(t, err)
	assert.IsType(t, agent.ToolResult{}, result)
	assert.Empty(t, result.ErrorCategory)
	assert.False(t, result.IsRetryable)
}
