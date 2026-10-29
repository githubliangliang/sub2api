## ADDED Requirements

### Requirement: Model refresh rechecks cache freshness inside a flight
The system SHALL reuse a fresh model cache entry found after entering a refresh flight and SHALL retain existing cache identity, stale refresh and conditional-request behavior.

#### Scenario: Another flight fills the cache between miss and entry
- **GIVEN** a caller observed a miss and another completed refresh subsequently stored a fresh entry
- **WHEN** the original caller enters the flight
- **THEN** it receives that entry without another upstream fetch

#### Scenario: The entry is missing or stale
- **WHEN** no fresh entry exists at flight entry
- **THEN** the existing upstream fetch and ETag/304 rules apply

#### Scenario: Concurrent accounts and cancellation
- **WHEN** callers for independent credential identities refresh concurrently or a waiting caller cancels
- **THEN** entries stay isolated and the established shared-flight cancellation contract is preserved

