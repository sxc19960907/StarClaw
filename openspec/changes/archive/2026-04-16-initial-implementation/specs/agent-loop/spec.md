## ADDED Requirements

### Requirement: Agent conversation loop
The system SHALL implement a conversation loop that can process user input and tool calls.

#### Scenario: Simple conversation without tools
- **GIVEN** a user query that doesn't require tools
- **WHEN** the Agent loop processes the query
- **THEN** the LLM SHALL respond with text only
- **AND** the loop SHALL return the response

#### Scenario: Conversation with single tool call
- **GIVEN** a user query requiring a tool (e.g., "read file.txt")
- **WHEN** the Agent loop processes the query
- **THEN** the LLM SHALL request a tool call
- **AND** the tool SHALL be executed
- **AND** the result SHALL be sent back to LLM
- **AND** the LLM SHALL provide a final response

#### Scenario: Conversation with tool chain
- **GIVEN** a user query requiring multiple sequential tools
- **WHEN** the Agent loop processes the query
- **THEN** the first tool SHALL be called and executed
- **AND** the result SHALL be sent to LLM
- **AND** the LLM SHALL request the second tool
- **AND** this SHALL continue until the task is complete

#### Scenario: Tool execution error
- **GIVEN** a tool call that fails
- **WHEN** the tool is executed
- **THEN** the error SHALL be categorized (transient/validation/business/permission)
- **AND** the error information SHALL be sent to LLM
- **AND** the LLM MAY retry or report the error

### Requirement: Iteration limit
The system SHALL enforce a maximum iteration limit to prevent infinite loops.

#### Scenario: Maximum iterations reached
- **GIVEN** `max_iterations` is set to 5
- **AND** a conversation requiring more than 5 tool calls
- **WHEN** the Agent loop processes the query
- **THEN** after 5 iterations, the loop SHALL stop
- **AND** return a partial result with a warning about the limit

### Requirement: Context management
The system SHALL maintain conversation context across the loop.

#### Scenario: Context preservation
- **GIVEN** a multi-turn conversation
- **WHEN** the Agent processes subsequent queries
- **THEN** the context SHALL include previous user messages and assistant responses

### Requirement: Tool result truncation
The system SHALL truncate large tool results to prevent token overflow.

#### Scenario: Large file read
- **GIVEN** a tool result exceeding `result_truncation` limit (default 30000 chars)
- **WHEN** the result is sent to LLM
- **THEN** the result SHALL be truncated with a note about truncation

### Requirement: Error categorization
The system SHALL categorize tool errors for appropriate handling.

#### Scenario: Transient error retry
- **GIVEN** a network timeout error
- **WHEN** the error is categorized
- **THEN** it SHALL be marked as "transient" and retryable

#### Scenario: Validation error no retry
- **GIVEN** invalid arguments to a tool
- **WHEN** the error is categorized
- **THEN** it SHALL be marked as "validation" and not retryable

### Requirement: Event handling
The system SHALL provide hooks for loop events (tool call, result, text, etc.).

#### Scenario: Tool call event
- **WHEN** a tool is about to be called
- **THEN** an event SHALL be fired with tool name and arguments
- **AND** event handlers SHALL be able to observe or intercept

#### Scenario: Tool result event
- **WHEN** a tool completes
- **THEN** an event SHALL be fired with the result
- **AND** event handlers SHALL receive the result and execution time
