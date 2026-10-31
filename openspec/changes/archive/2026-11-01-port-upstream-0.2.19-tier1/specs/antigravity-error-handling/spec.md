## ADDED Requirements

### Requirement: Malformed nonstream tool responses are failures
The system SHALL reject a selected upstream candidate ending with MALFORMED_FUNCTION_CALL before producing a successful nonstream response, including when collecting upstream SSE for a nonstream client.

#### Scenario: Direct or wrapped JSON reports malformed tools
- **WHEN** the selected response candidate has the malformed finish reason
- **THEN** the existing compatibility error format reports an upstream failure and no empty successful message is sent

#### Scenario: An error tail has no content parts
- **WHEN** collected SSE ends with a malformed-tool frame without parts
- **THEN** the failure reason survives aggregation and no fabricated success terminal is emitted

#### Scenario: Normal results or cancellation occur
- **WHEN** the upstream returns ordinary text/tools or the request is canceled
- **THEN** established transformation, passthrough, cancellation and failure-usage behavior remain intact

