## ADDED Requirements

### Requirement: Configuration file management
The system SHALL support loading and saving configuration from YAML files.

#### Scenario: Load existing configuration
- **WHEN** the application starts and `~/.starclaw/config.yaml` exists
- **THEN** the system SHALL load configuration from that file

#### Scenario: Create default configuration
- **WHEN** the application starts and `~/.starclaw/config.yaml` does not exist
- **THEN** the system SHALL create a default configuration file

### Requirement: Configuration hierarchy
The system SHALL support local configuration overriding global configuration.

#### Scenario: Local config overrides global
- **GIVEN** `~/.starclaw/config.yaml` has `endpoint: https://api.global.com`
- **AND** `./.starclaw/config.local.yaml` has `endpoint: https://api.local.com`
- **WHEN** the application loads configuration from the project directory
- **THEN** the endpoint SHALL be `https://api.local.com`

### Requirement: Interactive setup wizard
The system SHALL provide an interactive setup wizard for first-time configuration.

#### Scenario: First-time setup
- **WHEN** the user runs `starclaw --setup`
- **THEN** the system SHALL prompt for API endpoint
- **AND** prompt for API key
- **AND** save the configuration to `~/.starclaw/config.yaml`

#### Scenario: Setup with missing required fields
- **WHEN** the user runs `starclaw` without API key configured
- **THEN** the system SHALL prompt "未检测到 API Key，启动配置向导..."
- **AND** guide the user through setup

### Requirement: Secure API key storage
The system SHALL securely store API keys with restricted file permissions.

#### Scenario: API key file permissions
- **WHEN** the configuration is saved with an API key
- **THEN** the config file SHALL have permissions 0600 (owner read/write only)

### Requirement: Configuration validation
The system SHALL validate configuration on load and report errors clearly.

#### Scenario: Invalid configuration
- **GIVEN** `~/.starclaw/config.yaml` contains invalid YAML
- **WHEN** the application loads configuration
- **THEN** the system SHALL return a clear error message indicating the file and line

## MODIFIED Requirements

（无修改）

## REMOVED Requirements

（无删除）
