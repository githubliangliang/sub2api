# Upstream 0.2.25 Tier 1 Compatibility

## ADDED Requirements

### Requirement: Account-aware client error redaction

Existing OpenAI passthrough error messages exposed to a client SHALL redact configured account credentials and upstream host names while retaining non-sensitive recovery information and the existing query-parameter redaction.

#### Scenario: Context-window message contains account identity
- **WHEN** an upstream context-window error embeds the configured host or account credential
- **THEN** the client SHALL receive the actionable context error without those sensitive values
- **AND** error status, account side effects and strict response-header filtering SHALL retain their established behavior

#### Scenario: Unicode case folding changes byte length
- **WHEN** the message contains case-folding characters, invalid UTF-8, or literal regular-expression characters beside a host
- **THEN** replacement SHALL avoid panic and remove the configured host without slicing by transformed byte offsets
- **AND** a failed host-pattern interpretation SHALL fail closed to a generic message

#### Scenario: Generic forbidden error remains sanitized
- **WHEN** an unrelated forbidden upstream error has no matched disclosure rule
- **THEN** the existing generic sanitized response SHALL remain in effect and SHALL be written once
