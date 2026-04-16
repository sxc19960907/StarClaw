## ADDED Requirements

### Requirement: File read tool
The system SHALL provide a tool to read file contents with line numbers.

#### Scenario: Read entire file
- **GIVEN** a file at `/path/to/file.txt` containing "Hello\nWorld"
- **WHEN** the tool is invoked with `{"path": "/path/to/file.txt"}`
- **THEN** the result SHALL contain "   1 | Hello\n   2 | World"

#### Scenario: Read file with offset and limit
- **GIVEN** a file with 100 lines
- **WHEN** the tool is invoked with `{"path": "/path/to/file.txt", "offset": 10, "limit": 5}`
- **THEN** the result SHALL contain lines 11-15 with their line numbers

#### Scenario: Read non-existent file
- **GIVEN** a file that does not exist
- **WHEN** the tool is invoked
- **THEN** the result SHALL be an error with category "validation"

#### Scenario: Read file without permission
- **GIVEN** a file the user cannot read
- **WHEN** the tool is invoked
- **THEN** the result SHALL be an error with category "permission"

### Requirement: File write tool
The system SHALL provide a tool to write content to files.

#### Scenario: Write new file
- **WHEN** the tool is invoked with `{"path": "/path/to/new.txt", "content": "Hello"}`
- **THEN** the file SHALL be created with content "Hello"
- **AND** the result SHALL indicate success

#### Scenario: Overwrite existing file
- **GIVEN** an existing file
- **WHEN** the tool is invoked to write to that path
- **THEN** the file SHALL be overwritten
- **AND** the result SHALL indicate success with a note about overwriting

### Requirement: File edit tool
The system SHALL provide a tool to edit specific lines in files.

#### Scenario: Replace content in file
- **GIVEN** a file containing "Hello World"
- **WHEN** the tool is invoked with `{"path": "/path/to/file.txt", "old_string": "World", "new_string": "Universe"}`
- **THEN** the file SHALL contain "Hello Universe"

#### Scenario: Edit with multiple matches
- **GIVEN** a file with multiple occurrences of "old"
- **WHEN** the tool is invoked to replace "old" with "new"
- **THEN** only the first occurrence SHALL be replaced

### Requirement: Glob tool
The system SHALL provide a tool to find files matching patterns.

#### Scenario: Find Go files
- **GIVEN** a directory with `.go` files
- **WHEN** the tool is invoked with `{"pattern": "**/*.go"}`
- **THEN** the result SHALL list all Go files recursively

#### Scenario: Glob with no matches
- **GIVEN** a pattern that matches no files
- **WHEN** the tool is invoked
- **THEN** the result SHALL be an empty list (not an error)

### Requirement: Grep tool
The system SHALL provide a tool to search file contents with regex.

#### Scenario: Search for pattern
- **GIVEN** files containing the search term
- **WHEN** the tool is invoked with `{"pattern": "func main", "path": "/path/to/dir"}`
- **THEN** the result SHALL list matching files with line numbers and content

#### Scenario: Search with no results
- **GIVEN** a pattern that doesn't match
- **WHEN** the tool is invoked
- **THEN** the result SHALL indicate no matches found (not an error)

### Requirement: Bash tool
The system SHALL provide a tool to execute shell commands.

#### Scenario: Execute safe command
- **WHEN** the tool is invoked with `{"command": "ls -la"}`
- **THEN** the command SHALL execute and return output

#### Scenario: Command timeout
- **GIVEN** a command that runs longer than the timeout
- **WHEN** the tool is invoked with `{"command": "sleep 100", "timeout": 5}`
- **THEN** the result SHALL be a transient error indicating timeout

#### Scenario: Command with exit error
- **GIVEN** a command that exits with non-zero status
- **WHEN** the tool is invoked
- **THEN** the result SHALL contain the exit code and stderr output

#### Scenario: Safe command auto-approval
- **GIVEN** a command in the safe list (e.g., "ls", "pwd", "git status")
- **WHEN** the tool is invoked
- **THEN** the command MAY execute without user approval (configurable)

### Requirement: Directory list tool
The system SHALL provide a tool to list directory contents.

#### Scenario: List directory
- **WHEN** the tool is invoked with `{"path": "/path/to/dir"}`
- **THEN** the result SHALL list all files and directories with their types and sizes

#### Scenario: List non-existent directory
- **GIVEN** a directory that does not exist
- **WHEN** the tool is invoked
- **THEN** the result SHALL be an error with category "validation"

### Requirement: Tool registration
The system SHALL support registering and discovering tools dynamically.

#### Scenario: Register new tool
- **GIVEN** a tool implementation satisfying the Tool interface
- **WHEN** the tool is registered with the registry
- **THEN** the tool SHALL be available for the Agent to use

#### Scenario: List available tools
- **GIVEN** multiple tools registered
- **WHEN** the registry is queried for tools
- **THEN** all registered tools SHALL be returned with their Info
