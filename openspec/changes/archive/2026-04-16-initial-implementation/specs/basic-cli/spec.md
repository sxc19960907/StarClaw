## ADDED Requirements

### Requirement: One-shot CLI mode
The system SHALL support one-shot mode where a single query is processed and the program exits.

#### Scenario: Execute one-shot command
- **WHEN** the user runs `starclaw "list all Go files in this directory"`
- **THEN** the system SHALL process the query
- **AND** output the response
- **AND** exit with code 0

#### Scenario: One-shot with tool approval
- **GIVEN** a query requiring a tool call that needs approval
- **WHEN** the user runs in one-shot mode without --yes flag
- **THEN** the system SHALL prompt for approval
- **AND** wait for user input before proceeding

#### Scenario: One-shot with auto-approve
- **GIVEN** a query requiring tool calls
- **WHEN** the user runs with `-y` or `--yes` flag
- **THEN** all safe tool calls SHALL be executed without prompting

#### Scenario: One-shot error handling
- **GIVEN** a query that fails
- **WHEN** the command runs
- **THEN** the error SHALL be printed to stderr
- **AND** the program SHALL exit with non-zero code

### Requirement: Help and version flags
The system SHALL provide --help and --version flags.

#### Scenario: Show help
- **WHEN** the user runs `starclaw --help`
- **THEN** usage information SHALL be displayed
- **AND** available flags and subcommands SHALL be listed

#### Scenario: Show version
- **WHEN** the user runs `starclaw --version`
- **THEN** the current version SHALL be displayed (e.g., "starclaw version 0.1.0")

### Requirement: Setup command
The system SHALL provide a setup command for configuration.

#### Scenario: Run setup wizard
- **WHEN** the user runs `starclaw --setup`
- **THEN** the interactive setup wizard SHALL run
- **AND** prompt for endpoint and API key
- **AND** save the configuration

### Requirement: Configuration reload
The system SHALL support reloading configuration without restart (where applicable).

#### Scenario: Reload config in daemon mode
- **GIVEN** the daemon is running
- **WHEN** the user updates config and sends reload signal
- **THEN** the daemon SHALL reload configuration without restart

### Requirement: Subcommand structure
The system SHALL organize functionality into subcommands.

#### Scenario: Available subcommands
- **WHEN** the user runs `starclaw --help`
- **THEN** the following subcommands SHALL be listed:
  - `daemon` - Daemon mode operations
  - `version` - Show version
  - `update` - Check for updates

### Requirement: Interactive mode detection
The system SHALL detect whether stdin is a TTY.

#### Scenario: Non-TTY input
- **GIVEN** input is piped (e.g., `echo "query" | starclaw`)
- **WHEN** the program runs
- **THEN** it SHALL handle the input appropriately
- **AND** not prompt for interactive input

#### Scenario: TTY detection
- **GIVEN** input is from terminal
- **WHEN** the program runs without arguments
- **THEN** it SHALL start in interactive/TUI mode
