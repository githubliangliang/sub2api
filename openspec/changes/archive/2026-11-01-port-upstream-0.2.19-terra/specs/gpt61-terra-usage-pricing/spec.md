## ADDED Requirements

### Requirement: Terra usage follows the existing pricing and settlement contract
The system SHALL use the frozen Terra resource and dedicated fallback prices within existing pricing precedence, cache-token, priority, long-context and SQLite dedup rules without enabling additional commercial mechanisms.

#### Scenario: Frozen default prices are used
- **WHEN** no explicit account/channel pricing overrides Terra defaults
- **THEN** resource and both fallback layers agree on input/cache-read/cache-write/output values of 1/0.05/1.25/5 per million tokens and their source-defined priority values

#### Scenario: Explicit zero or custom prices exist
- **WHEN** an explicit zero cache-write value, channel override or account-specific pricing is supplied
- **THEN** the existing precedence and explicit-value semantics remain intact without mutating shared default cards

#### Scenario: The existing long-context count crosses its threshold
- **WHEN** the established sum of noncached input and cache buckets is 271999, 272000 or 272001
- **THEN** default Terra long-context multipliers apply only above 272000, using the source input factor 2 and output factor 1.5

#### Scenario: Usage crosses protocol and storage boundaries
- **WHEN** successful Terra HTTP/SSE or an existing WS bridge hands observed usage to the handler and SQLite settlement
- **THEN** all token buckets and response/request model attribution survive the asynchronous snapshot
- **AND** the existing dedup, fixed quota window and configured/soft-deleted quota contract remains intact

