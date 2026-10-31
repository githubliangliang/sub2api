## ADDED Requirements

### Requirement: Authorized audit cleanup works on SQLite
The system SHALL support the existing operation-log clear-all action using SQLite-compatible SQL and preserve its synchronous cleanup trace.

#### Scenario: Verified administrator clears operation logs
- **GIVEN** an administrator session and a valid fresh TOTP code
- **WHEN** the administrator confirms clear-all
- **THEN** historical audit rows are deleted, the deleted count is returned and a new clear trace remains
- **AND** historical audit IDs are not reused
- **AND** the page refreshes and displays the existing success result

#### Scenario: Verification is missing or rejected
- **GIVEN** a missing/invalid TOTP code, disabled TOTP, or an admin API-key request
- **WHEN** clear-all is requested
- **THEN** the existing authorization or validation response is returned and no logs are cleared

#### Scenario: Database request fails
- **WHEN** the database operation is canceled or fails
- **THEN** the failure is reported and the UI allows a new verified attempt
