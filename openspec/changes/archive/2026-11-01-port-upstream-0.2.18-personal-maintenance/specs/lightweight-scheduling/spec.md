## ADDED Requirements

### Requirement: Single-account retry detection is lazy per request
The system SHALL defer the Antigravity single-account check until its retry or model-limit logic requires it and memoize the result only within the request.

#### Scenario: A normal request does not need the check
- **WHEN** a successful request or another provider path does not evaluate Antigravity single-account retry
- **THEN** setting up its request metadata performs no account-group probe

#### Scenario: Multiple consumers require the check
- **GIVEN** a request whose Antigravity retry logic requires the check
- **WHEN** multiple consumers, including concurrent consumers, read it
- **THEN** the query runs at most once using the original group ID snapshot

#### Scenario: Retry policy is explicitly set
- **GIVEN** an explicit single-account retry true or false value
- **WHEN** retry metadata is read
- **THEN** the explicit value is honored without evaluating the lazy probe

### Requirement: Mapping signatures do not allocate and preserve cache invalidation
The system SHALL compute mapping signatures without temporary per-call heap allocations while retaining content-sensitive invalidation of existing account mapping caches.

#### Scenario: Map traversal order changes
- **GIVEN** two raw maps with equal entries and different insertion order
- **WHEN** signatures are computed
- **THEN** signatures are equal

#### Scenario: Mapping changes in place
- **GIVEN** an account mapping cache is populated
- **WHEN** a target changes in place while map identity and size stay equal
- **THEN** the next mapping read returns the new target

#### Scenario: Signature is repeatedly computed
- **WHEN** supported raw mappings are signed repeatedly
- **THEN** the measured per-call allocation count is zero
