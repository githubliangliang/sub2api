# Historical Followup Narrow Correctness

## ADDED Requirements

### Requirement: Generated Codex instructions reflect the routed model identity

Existing generated manifests SHALL avoid a GPT self-identification for non-GPT routed models while preserving official GPT templates and other instructions.

#### Scenario: Public alias routes to a non-GPT model
- **WHEN** the existing manifest generator selects a non-GPT model directly or through an alias
- **THEN** its generated instructions SHALL omit the false GPT self-identification according to the effective target
- **AND** other working instructions, placeholders and explicit upstream or user instructions SHALL remain intact

#### Scenario: Existing GPT descriptor
- **WHEN** the routed model is an existing GPT model
- **THEN** its established template SHALL remain byte-identical
- **AND** no new model-discovery system or model capability SHALL be introduced

### Requirement: Retained S3 secrets stay encrypted on configuration writes

Every existing S3 configuration write that retains a nonempty secret SHALL honor the established fixed-key and encryption boundary before persistence and return a redacted response.

#### Scenario: Operator leaves the existing secret blank
- **WHEN** an operator updates a non-secret configuration field and leaves the secret input empty
- **THEN** the retained secret SHALL remain encrypted in SQLite and usable through the established reader
- **AND** the response SHALL continue to omit the secret

#### Scenario: Fixed key or persistence fails
- **WHEN** the fixed-key requirement, encryption or persistence fails
- **THEN** the established failure contract SHALL prevent a plaintext secret from being persisted
- **AND** no backup, restore or object-store operation SHALL be started by saving configuration

### Requirement: Attribute form payloads preserve existing update semantics

Existing attribute forms SHALL submit numeric values as strings and explicit clears as empty strings while retaining the current integer validation and omitted-field semantics.

#### Scenario: Integer custom attribute is edited
- **WHEN** an operator edits a valid integer-valued attribute
- **THEN** the request SHALL contain a string acceptable to the existing attribute API
- **AND** zero, negative values and configured bounds SHALL follow the established server policy

#### Scenario: Description or placeholder is cleared
- **WHEN** an operator clears a saved description or placeholder
- **THEN** the request SHALL include an explicit empty string for that field
- **AND** genuinely omitted fields SHALL retain their existing meaning
