## ADDED Requirements

### Requirement: Existing client configuration surfaces expose Terra consistently
The system SHALL add Terra to the existing model whitelist, mapping presets and supported client configuration generators while preserving global defaults and other platform/model behavior.

#### Scenario: Terra is selected in an existing model surface
- **WHEN** a user chooses a whitelist entry or mapping preset
- **THEN** the canonical gpt-6.1-terra identity is represented consistently

#### Scenario: An existing client configuration is generated
- **WHEN** the user generates the existing OpenCode client configuration for Terra
- **THEN** the output parses successfully and includes source-defined context/output defaults and none/low/medium/high/xhigh/max variants

#### Scenario: Another model or platform is configured
- **WHEN** an existing non-Terra configuration is generated
- **THEN** its established output and global default remain unchanged
- **AND** no new account platform, menu, schema or payment capability is introduced

