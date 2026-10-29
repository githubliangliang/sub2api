## ADDED Requirements

### Requirement: Pagination arithmetic stays within native integer bounds
The system SHALL prevent numeric parsing, offsets and offset-plus-limit calculations from overflowing and SHALL retain each entry point's existing default, saturation or invalid-input response policy.

#### Scenario: Query integers are oversized or malformed
- **WHEN** pagination query values are overlong decimals, MaxInt, negative, empty or malformed
- **THEN** no wrapped offset or internal server error is produced and the entry point applies its defined policy

#### Scenario: Parameters are constructed without HTTP parsing
- **WHEN** shared pagination parameters are used directly for repository queries or in-memory API-key slicing
- **THEN** the offset and offset plus limit are safe, including the largest supported page

### Requirement: Administrator guard failures are client errors
The system SHALL return stable bad-request errors for forbidden last-administrator demotion and direct administrator disable/delete while preserving the existing protection rules.

#### Scenario: A forbidden administrator mutation is requested
- **WHEN** the request would violate an administrator guard
- **THEN** the handler returns HTTP 400 with a stable domain code and no forbidden database mutation

#### Scenario: Other administrators or infrastructure failures exist
- **WHEN** demotion is allowed because another administrator exists or a database operation fails
- **THEN** allowed demotion succeeds and infrastructure failure remains an HTTP 500 error

