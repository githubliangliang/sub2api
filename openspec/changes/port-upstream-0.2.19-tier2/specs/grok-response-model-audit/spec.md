## ADDED Requirements

### Requirement: Grok response-model audits survive existing conversions
The system SHALL carry the upstream response model and its conflict flag through native Responses, existing Chat/Messages conversions and the usage-write boundary without substituting client or mapped model identifiers.

#### Scenario: Native Responses uses JSON or SSE
- **WHEN** an upstream Grok response contains a raw model declaration
- **THEN** the forward result and eventual usage record retain that declaration as the response-model audit

#### Scenario: Buffered events declare different models
- **WHEN** a nonterminal event declares one Grok model and a terminal event declares another during Chat or Messages buffering
- **THEN** the terminal model wins and the conflict flag survives the forward result, asynchronous snapshot and usage-write boundary
- **AND** the SQLite record retains the response model and the established billing fallback is verified without adding a conflict column or changing migration 227

#### Scenario: Many events or another provider are processed
- **WHEN** a stream has many long model declarations or the account is not Grok
- **THEN** prior-model state stays within existing field limits and fixed memory
- **AND** the additional buffered prior audit does not alter other providers' behavior

