## ADDED Requirements

### Requirement: Claude API integration
The system SHALL support Claude API via Anthropic's Go SDK.

#### Scenario: Send chat request
- **GIVEN** a valid API key and configured endpoint
- **WHEN** a chat request is sent with messages
- **THEN** the system SHALL return the LLM response

#### Scenario: Tool definition formatting
- **GIVEN** registered tools with their schemas
- **WHEN** sending a request to Claude
- **THEN** tools SHALL be formatted according to Claude's tool use API specification

#### Scenario: Parse tool use response
- **GIVEN** a Claude response containing tool_use blocks
- **WHEN** the response is parsed
- **THEN** the tool_use blocks SHALL be extracted with name and arguments

#### Scenario: Parse text response
- **GIVEN** a Claude response containing only text
- **WHEN** the response is parsed
- **THEN** the text content SHALL be returned as the final response

### Requirement: API key configuration
The system SHALL support configuring API key via configuration file.

#### Scenario: Use configured API key
- **GIVEN** `api_key` is set in config.yaml
- **WHEN** making an API request
- **THEN** the configured key SHALL be used in the Authorization header

#### Scenario: Missing API key
- **GIVEN** no API key is configured
- **WHEN** attempting to make an API request
- **THEN** the system SHALL return an error prompting for configuration

### Requirement: Model configuration
The system SHALL support configuring the model and parameters.

#### Scenario: Model selection
- **GIVEN** `model_tier` is set to "medium" in config
- **WHEN** making an API request
- **THEN** the appropriate model (e.g., "claude-4-sonnet") SHALL be used

#### Scenario: Temperature configuration
- **GIVEN** `temperature` is set to 0.5
- **WHEN** making an API request
- **THEN** the temperature parameter SHALL be passed to the API

#### Scenario: Max tokens configuration
- **GIVEN** `max_tokens` is set to 4096
- **WHEN** making an API request
- **THEN** the max_tokens parameter SHALL be passed to the API

### Requirement: Error handling
The system SHALL handle API errors gracefully.

#### Scenario: API authentication error
- **GIVEN** an invalid API key
- **WHEN** making a request
- **THEN** the system SHALL return a clear error: "Invalid API key. Please check your configuration."

#### Scenario: API rate limit
- **GIVEN** the API returns a rate limit error
- **WHEN** the error is handled
- **THEN** it SHALL be categorized as "transient" and potentially retryable

#### Scenario: Network error
- **GIVEN** a network timeout
- **WHEN** the error occurs
- **THEN** it SHALL be categorized as "transient" with retry suggestion

### Requirement: Token usage tracking
The system SHALL track and report token usage.

#### Scenario: Track usage per request
- **WHEN** an API request completes
- **THEN** the system SHALL record input_tokens and output_tokens
- **AND** make usage available to event handlers

#### Scenario: Cache tracking (Claude specific)
- **WHEN** using Claude's prompt caching
- **THEN** the system SHALL track cache_creation_tokens and cache_read_tokens
