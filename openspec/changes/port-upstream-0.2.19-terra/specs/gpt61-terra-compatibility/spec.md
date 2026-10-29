## ADDED Requirements

### Requirement: Terra has a consistent model identity and effort contract
The system SHALL recognize GPT-6.1 Terra under the existing provider-prefix and spelling normalization rules, support none/low/medium/high/xhigh/max with medium as the default, and preserve explicit thinking-disable precedence.

#### Scenario: A supported spelling includes an effort or compact suffix
- **WHEN** a Terra request uses a source-supported effort/compact spelling, case variation, underscore normalization or provider prefix
- **THEN** its canonical upstream model remains gpt-6.1-terra and the requested supported effort is preserved

#### Scenario: Disabled thinking or unrelated model rules apply
- **WHEN** Terra disables thinking or a Sol request uses none/minimal
- **THEN** Terra retains none and the existing Sol rejection remains intact

#### Scenario: An unsupported Terra suffix is used
- **WHEN** pro, ultra or minimal is used as a Terra capability suffix
- **THEN** it is not treated as a supported Terra effort or silently routed to another known family

### Requirement: Terra tools use the existing Responses path
The system SHALL require Responses for Terra tool calls and SHALL reject tool-bearing only-Chat fallbacks before an upstream call using the client's existing error format.

#### Scenario: A Responses-capable account handles tools
- **WHEN** Chat, Messages or Responses compatibility receives a Terra tool request for a capable account
- **THEN** the existing Responses conversion preserves the model, effort and tool identity

#### Scenario: An account supports only Chat Completions
- **WHEN** Chat, Messages or Responses fallback has effective tools or legacy functions for Terra
- **THEN** it returns HTTP 400 before forwarding
- **AND** ordinary no-tool text retains its existing raw compatibility behavior

### Requirement: Account metadata takes precedence over Terra defaults
The system SHALL add a public default Terra description while preserving explicit account capabilities and raw upstream catalogs.

#### Scenario: The account has restrictive metadata
- **WHEN** an account declares text-only input, high-only reasoning or a smaller context limit
- **THEN** the manifest honors those declarations instead of overwriting them with public Terra defaults

#### Scenario: Upstream discovery returns an empty or populated catalog
- **WHEN** discovery processes the upstream catalog
- **THEN** the empty catalog remains empty, the populated catalog retains its raw data, and cache/ETag contents are not modified in place

#### Scenario: No account-specific declaration exists
- **WHEN** a configured Terra descriptor needs defaults
- **THEN** its default reasoning is medium with the six supported efforts
- **AND** the existing global default model remains unchanged

