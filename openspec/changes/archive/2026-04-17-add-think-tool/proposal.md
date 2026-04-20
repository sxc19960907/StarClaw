# Proposal: Add Think Tool

## Summary

Add a `think` tool that allows the AI to explicitly reason and plan before taking action. This tool provides a dedicated scratchpad for the model's internal monologue, making its reasoning process explicit and traceable.

## Motivation

Currently, the AI may output reasoning as plain text before calling tools, but this is implicit and can be mixed with other content. By providing a dedicated `think` tool:

1. **Explicit reasoning** - The model can signal "I'm thinking about this" explicitly
2. **Better UX** - The UI can render think blocks differently (collapsible, styled)
3. **Debugging** - Easier to trace the model's decision-making process
4. **Chain of thought** - Encourages step-by-step reasoning for complex tasks

## Scope

### In Scope
- Create `ThinkTool` struct implementing the `Tool` interface
- Register tool in the tool registry
- Update system prompt to encourage tool usage
- Unit tests for tool functionality

### Out of Scope
- UI styling for think blocks (can be added later)
- Automatic summarization of thoughts
- Persistence of thoughts separately from messages

## Success Criteria

- [ ] AI can call `think` tool with a `thought` parameter
- [ ] Tool returns the thought content as the result
- [ ] Tool requires no approval (read-only operation)
- [ ] System prompt encourages using think tool for planning
- [ ] Unit tests cover happy path and error cases
- [ ] Integration tests verify tool registration

## Reference

Based on ShanClaw's implementation: `internal/tools/think.go`
