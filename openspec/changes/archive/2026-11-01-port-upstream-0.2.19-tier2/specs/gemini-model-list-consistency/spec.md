## ADDED Requirements

### Requirement: Declared Gemini model catalogs use existing mapping semantics
The system SHALL derive native model listings and model-detail membership from declared mappings when all schedulable Gemini accounts have effective mappings, and SHALL preserve upstream forwarding when any account is unmapped.

#### Scenario: All accounts declare mappings
- **WHEN** a group contains only mapped schedulable Gemini accounts
- **THEN** native listing uses the deduplicated declared model source and detail membership recognizes existing normalized and prefix-wildcard mappings

#### Scenario: An account is unmapped or no declaration covers the model
- **WHEN** any schedulable Gemini account is unmapped or a fully declared group has no mapping covering a queried model
- **THEN** the unmapped group retains its existing upstream model endpoint behavior and the fully declared group returns the established not-found response

#### Scenario: Custom display lists and mixed routing apply
- **WHEN** existing custom display rules, forced Antigravity, composite/mixed routing or a path guard apply
- **THEN** those local boundaries remain intact without treating display configuration as new request authorization
- **AND** no new ModelAllowlist or Vertex schema/capability is introduced

### Requirement: Existing Gemini defaults and selection lists stay consistent
The system SHALL remove the retired models identified by the frozen source from public default lists, use the already supported 3.5-flash for ordinary unspecified connection tests, and preserve explicit and legacy account choices.

#### Scenario: An ordinary test omits its model
- **WHEN** a non-legacy Gemini account is tested without an explicit model
- **THEN** the existing supported 3.5-flash default is used consistently by the appropriate frontend/backend selectors

#### Scenario: Legacy or explicit choices are used
- **WHEN** Google One conservative mappings, an explicit old model or a custom alias is configured
- **THEN** those choices remain usable under their existing account contract

#### Scenario: The combined selector includes existing Antigravity models
- **WHEN** the user selects from the combined model options
- **THEN** existing supported Antigravity entries are included without duplicates
- **AND** unselected Argon, 3.6/3.7/3.8 or new flash-lite-image capability entries are not added

