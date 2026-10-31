## ADDED Requirements

### Requirement: Existing Grok catalog consumers share strict parsing
The system SHALL use the same valid model-ID extraction rules for manual synchronization and background observation, preserve custom ID case, and distinguish an empty valid shape from an invalid response.

#### Scenario: A valid catalog contains duplicate or custom IDs
- **WHEN** a supported catalog shape contains protocol IDs, duplicate entries or custom case-sensitive IDs
- **THEN** both consumers return the same deduplicated deterministic IDs

#### Scenario: A failed response resembles JSON
- **WHEN** an error object, an entry lacking a protocol ID, invalid JSON or an empty valid catalog is received
- **THEN** it is not saved as successful model IDs
- **AND** empty_catalog is distinguishable from invalid_response while a previous success is retained

### Requirement: Catalog observations preserve ordered success and failure state
The system SHALL retain the last successful catalog while recording failures and SHALL use validated timestamps and source scope to prevent stale or mismatched observations from becoming fresh state.

#### Scenario: An older or invalid observation completes later
- **WHEN** an older-started request completes after a newer observation or an observation has invalid time values
- **THEN** successful state does not move backward and invalid time data does not create freshness

#### Scenario: Scope changes or a legacy snapshot is read
- **WHEN** credentials are used through a different catalog scope or an old snapshot lacks new scope metadata
- **THEN** the old success remains readable but is not falsely treated as a fresh catalog for the new scope

### Requirement: SQLite catalog writes and ordinary edits preserve managed state
The system SHALL atomically update only the catalog snapshot using semantic expected-value comparison and bounded conflict retries, and ordinary account editing SHALL preserve the latest database-owned catalog and upstream metadata.

#### Scenario: Independent writers compete
- **WHEN** catalog observation overlaps another catalog write or an unrelated extra update
- **THEN** only a valid compare-and-swap applies, other extra keys survive, and retries remain bounded without unlocked fallback

#### Scenario: A caller-owned transaction is rolled back
- **WHEN** a catalog write executes inside an external transaction that has not committed or is rolled back
- **THEN** SQL and ORM use that transaction and no scheduler cache publication exposes uncommitted data

#### Scenario: An old form, missing keys or injected state is submitted
- **WHEN** ordinary editing submits old or absent catalog/metadata values
- **THEN** database-owned current values win, including when editing overlaps observation

#### Scenario: Cancellation, soft deletion or a wrong platform occurs
- **WHEN** the context cancels, the account is soft-deleted, the platform is not Grok, or the repository lacks the required CAS capability
- **THEN** the write fails or does not apply under the existing error contract without reviving data or weakening concurrency

