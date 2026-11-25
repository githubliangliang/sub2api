# Historical Backlog Gemini Native Signals

## ADDED Requirements

### Requirement: Native Gemini in-band outcomes receive accurate Ops attribution

Existing native Gemini JSON, SSE and aggregated responses SHALL report semantic upstream failures and request-scoped policy outcomes through the existing Ops contracts while preserving downstream status, response bytes and observed usage.

#### Scenario: HTTP 200 carries a Google error envelope
- **WHEN** an existing native response contains a valid Google upstream error envelope inside HTTP 200
- **THEN** Ops SHALL record the semantic failure and its intended status using existing logging and SLA rules
- **AND** downstream wire status and payload SHALL remain unchanged

#### Scenario: Model result is a request-scoped policy outcome
- **WHEN** prompt blocking or a recognized content filter explains the native result
- **THEN** the outcome SHALL be attributed to the request without account punishment or upstream SLA failure
- **AND** normal STOP, MAX_TOKENS and ordinary tool-related stops SHALL NOT be misclassified as upstream failures

#### Scenario: Wrapped, streamed, empty and mixed outcomes
- **WHEN** outcomes arrive in OAuth wrappers, arrays, SSE, aggregation or a genuinely empty response
- **THEN** bounded observation SHALL select and record the appropriate outcome once according to signal priority
- **AND** existing skip-monitoring rules, cancellation, read/write failure ownership and resource limits SHALL remain effective

#### Scenario: Accounting and client behavior remain stable
- **WHEN** a native response contains observed usage or a client disconnects
- **THEN** established usage, independent account cost, user charges, fixed quota and dedup behavior SHALL remain intact
- **AND** signal reporting SHALL NOT introduce retries, account penalties or new quota strategies
