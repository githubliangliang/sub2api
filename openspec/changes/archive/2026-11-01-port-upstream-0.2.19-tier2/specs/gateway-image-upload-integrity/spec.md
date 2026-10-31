## ADDED Requirements

### Requirement: Image parts are forwarded intact or rejected
The system SHALL reject oversized OpenAI multipart parts rather than silently truncating them and SHALL retain the current 20 MiB per-part boundary.

#### Scenario: A part reaches or exceeds the boundary
- **WHEN** an image or mask part is exactly 20 MiB or one byte larger
- **THEN** the allowed part remains byte-identical and the oversized part receives a clear rejection before forwarding

#### Scenario: Multipart fields and retries are used
- **WHEN** a request has multiple parts, filenames, content types or retries
- **THEN** accepted payloads remain intact and no rejected partial payload is used for another attempt

### Requirement: Grok parsing obeys the existing total-body boundary
The system SHALL keep multipart bytes used for Grok parsing and review consistent with the bounded request body and SHALL not replace truncation with unbounded request-body acceptance.

#### Scenario: A bounded body contains a large part
- **WHEN** a request passes the established total-body limit
- **THEN** parsing retains the full accepted part or reports a clear failure, rather than silently reviewing a shortened image

#### Scenario: Total-body size, cancellation or download limits apply
- **WHEN** the full body is oversized, canceled or a separate download operation runs
- **THEN** the established body/cancellation/download boundaries remain in force
- **AND** the unselected 50 MiB upload expansion is not enabled

