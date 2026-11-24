# Upstream 0.2.25 Tier 2 Compatibility

## ADDED Requirements

### Requirement: Passthrough deterministic error fidelity

Existing OpenAI passthrough final errors SHALL preserve sanitized, bounded standard scalar fields for allowed deterministic 400 errors and administrator passthrough-body rules without forwarding arbitrary upstream fields or sensitive headers.

#### Scenario: Invalid encrypted content or context window error
- **WHEN** an allowed upstream 400 contains type, code, param and message
- **THEN** the client SHALL receive the applicable standard fields with account-aware redaction
- **AND** object, array, boolean and oversized identifier fields SHALL be rejected rather than serialized as text

#### Scenario: Blank code wraps an embedded error
- **WHEN** a missing, null or whitespace code accompanies a supported embedded error message
- **THEN** the outward code SHALL agree with the existing compatible error classification
- **AND** explicit malformed structured code values SHALL not enable arbitrary-field disclosure

### Requirement: Administrator error rule consistency

The passthrough final-error path SHALL apply existing administrator rules using the original upstream status, body and account platform, retaining rule precedence, explicit custom messages and monitoring semantics.

#### Scenario: Custom response or passthrough body is configured
- **WHEN** an applicable rule selects a custom status/message or a passthrough body
- **THEN** the configured result SHALL determine the final client error within the established sanitized envelope
- **AND** unmatched errors and non-400 disclosure boundaries SHALL retain their existing behavior

#### Scenario: Monitoring rule matches a code outside the extracted message
- **WHEN** detail logging is disabled and skip_monitoring matches only the original structured error code
- **THEN** matching SHALL still use that original body and SHALL set the established monitoring marker
- **AND** already-working status-rule suppression SHALL remain effective

### Requirement: Compact failure lifecycle fidelity

After compact keepalive commits HTTP 200, final error handling SHALL emit exactly one ordered response.failed terminal event, preserve deterministic client codes, and rewrite capacity-shed codes to the established retryable client code.

#### Scenario: Encrypted-content error follows heartbeat
- **WHEN** the heartbeat has committed and the upstream returns invalid_encrypted_content
- **THEN** the client SHALL receive a single failed terminal event with that code and sanitized message
- **AND** sequence numbers SHALL advance through the existing request sequence allocator

#### Scenario: Capacity code reaches structured compact output
- **WHEN** structured error forwarding would expose server_is_overloaded or slow_down after heartbeat
- **THEN** the terminal client code SHALL be retryable while internal monitoring retains the original classification
- **AND** stopping, cancellation and failure accounting SHALL preserve existing ownership and deduplication

### Requirement: Bounded Gorilla security maintenance

The selected dependency maintenance SHALL use an official Gorilla fix for GO-2026-6278 and SHALL preserve existing administrator WebSocket security and interoperability while recording the exact module graph and test limitations.

#### Scenario: Exact security candidate is evaluated
- **WHEN** the Gorilla version is changed
- **THEN** its source, checksums, transitive module selection and full behavioral diff SHALL be verified
- **AND** the change SHALL not import upstream group synchronization or unrelated dependencies

#### Scenario: Existing WebSocket consumers remain compatible
- **WHEN** administrator or existing gateway WebSocket traffic executes with the selected module
- **THEN** authentication, Origin, single subprotocol selection, limits, control frames and connection release SHALL retain their established behavior
- **AND** genuine CONNECT and Basic-auth tunnel controls SHALL pass while historical dependency-fixture failures remain explicitly recorded and investigated
