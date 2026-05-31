# Validate Real Agent Workflow

## Goal

Validate that StarClaw's agent loop can execute a realistic local multi-tool workflow as a general-purpose agent, without requiring live provider credentials.

## User Value

After the CLI smoke validation, this task verifies the actual agent runtime path: model response with tool calls, tool execution, iterative follow-up, file mutation, shell command execution, event reporting, and session persistence.

## Confirmed Facts

- Existing integration and black-box tests cover many individual tools and some agent-loop behavior.
- Black-box tests that use real Anthropic credentials are skipped unless API keys are configured.
- `client.MockClient` supports deterministic multi-turn tool call responses.
- `agent.AgentLoop` records session messages and auto-saves through `session.Manager`.
- Current coverage does not contain one deterministic test that chains read/search/write/edit/bash/session in a single realistic workflow.

## Requirements

- Add deterministic validation for a multi-turn agent workflow using `client.MockClient`.
- Cover at least file read, file write, file edit, shell command execution, event handler observation, and session persistence.
- Avoid real provider credentials, external network, desktop automation, and user home state.
- Keep the test isolated to temporary directories.
- Fix any narrow runtime bug discovered while adding the workflow test.

## Acceptance Criteria

- [x] A deterministic test exercises a realistic multi-tool agent loop.
- [x] The test verifies final response content.
- [x] The test verifies file system side effects from write/edit tools.
- [x] The test verifies shell command tool output reached the agent loop.
- [x] The test verifies session data is persisted and resumable.
- [x] `go test ./...` passes locally.
- [x] `go vet ./...` passes locally.
- [ ] CI passes after push.

## Validation Results

- Added `TestAgentRealisticLocalWorkflow`.
- Fixed CWD cleanup in `TestAgentMultipleToolCalls`, which leaked temporary CWD state into later tests.
- `go test ./tests -count=1` passed.
- `go test ./...` passed.
- `go vet ./...` passed.

## Out Of Scope

- Live Anthropic/OpenAI/Ollama validation.
- Full TUI workflow.
- MCP remote server integration.
- Desktop/browser/screenshot workflows.
- Large changes to tool schemas or prompt behavior unless required by a discovered blocker.
