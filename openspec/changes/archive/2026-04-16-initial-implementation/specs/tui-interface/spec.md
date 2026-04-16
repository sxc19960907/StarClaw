## ADDED Requirements

### Requirement: TUI startup
The system SHALL launch a terminal UI when run without arguments.

#### Scenario: Launch TUI
- **WHEN** the user runs `starclaw` without arguments
- **THEN** the TUI SHALL start with an input prompt

#### Scenario: TUI header
- **WHEN** the TUI starts
- **THEN** it SHALL display:
  - Application name and version
  - Current model information
  - Connection status

### Requirement: Message input
The system SHALL provide a text input area for user messages.

#### Scenario: Type and submit message
- **GIVEN** the TUI is running
- **WHEN** the user types a message and presses Enter
- **THEN** the message SHALL be sent to the Agent
- **AND** the input SHALL be cleared

#### Scenario: Multi-line input
- **GIVEN** the TUI is running
- **WHEN** the user types a multi-line message (Shift+Enter for newline)
- **THEN** the complete message SHALL be submitted

#### Scenario: Input history
- **GIVEN** the user has sent previous messages
- **WHEN** the user presses Up arrow
- **THEN** previous messages SHALL be recalled for editing

### Requirement: Message display
The system SHALL display conversation history in the TUI.

#### Scenario: Display user message
- **WHEN** a user sends a message
- **THEN** it SHALL appear in the conversation area with user styling

#### Scenario: Display assistant response
- **WHEN** the assistant responds
- **THEN** the response SHALL appear with assistant styling
- **AND** Markdown SHALL be rendered appropriately

#### Scenario: Display tool calls
- **WHEN** a tool is called
- **THEN** a tool call indicator SHALL be displayed showing:
  - Tool name
  - Arguments (collapsed by default)
  - Execution status (pending/running/completed)

#### Scenario: Display tool results
- **WHEN** a tool completes
- **THEN** the result SHALL be displayed (optionally collapsible)
- **AND** execution time SHALL be shown

### Requirement: Streaming output
The system SHALL support streaming assistant responses.

#### Scenario: Stream text response
- **WHEN** the assistant is generating a response
- **THEN** text SHALL appear incrementally (word by word)
- **AND** the user SHALL see progress in real-time

#### Scenario: Stream with tool calls
- **GIVEN** a response containing both text and tool calls
- **WHEN** streaming
- **THEN** text SHALL stream first
- **AND** tool call SHALL appear when invoked
- **AND** final text SHALL stream after tool result

### Requirement: Tool approval UI
The system SHALL prompt for approval before executing sensitive tools.

#### Scenario: Approval dialog
- **GIVEN** a tool call requiring approval
- **WHEN** the tool is invoked
- **THEN** an approval dialog SHALL appear showing:
  - Tool name
  - Arguments
  - Approve / Deny buttons

#### Scenario: Approve tool
- **GIVEN** the approval dialog is displayed
- **WHEN** the user approves
- **THEN** the tool SHALL execute
- **AND** the result SHALL be displayed

#### Scenario: Deny tool
- **GIVEN** the approval dialog is displayed
- **WHEN** the user denies
- **THEN** the tool SHALL NOT execute
- **AND** an error SHALL be sent to the LLM

### Requirement: Keyboard shortcuts
The system SHALL support keyboard shortcuts for common actions.

#### Scenario: Quit shortcut
- **GIVEN** the TUI is running
- **WHEN** the user presses Ctrl+C or Ctrl+Q
- **THEN** the application SHALL exit gracefully

#### Scenario: Clear screen
- **GIVEN** the TUI is running
- **WHEN** the user presses Ctrl+L
- **THEN** the conversation SHALL be cleared from display

### Requirement: Responsive layout
The system SHALL adapt to terminal size changes.

#### Scenario: Resize terminal
- **GIVEN** the TUI is running
- **WHEN** the terminal is resized
- **THEN** the layout SHALL adjust to fit the new size
- **AND** text SHALL reflow appropriately

### Requirement: Error display
The system SHALL display errors gracefully in the TUI.

#### Scenario: Display error
- **GIVEN** an error occurs
- **WHEN** the error needs to be displayed
- **THEN** it SHALL appear with error styling
- **AND** not crash the TUI
