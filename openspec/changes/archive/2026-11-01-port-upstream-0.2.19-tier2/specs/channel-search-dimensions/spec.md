## ADDED Requirements

### Requirement: Channel updates preserve known and unknown search dimensions
The system SHALL merge only the Web Search emulation platform map on partial channel updates while preserving existing whole-block semantics for other features and avoiding mutation of caller-owned maps.

#### Scenario: A client knows only one platform
- **GIVEN** a channel has Anthropic and OpenAI search settings
- **WHEN** a client changes one dimension or omits the search key
- **THEN** other existing dimensions survive and supplied values take precedence

#### Scenario: A client explicitly clears or disables search
- **WHEN** a platform is explicitly false or the search key is an empty object or null
- **THEN** the specified disable or clear remains expressible and is not restored by inheritance

#### Scenario: Other feature keys are supplied
- **WHEN** incoming features include unrelated nested objects or unknown search dimensions
- **THEN** unrelated keys retain their historical replacement behavior, unknown search dimensions are retained as defined, and inputs remain unchanged

### Requirement: The existing channel editor saves actual supported dimensions
The system SHALL expose and persist the search setting for platforms with an existing local consumer, including OpenAI, without importing absent upstream platforms.

#### Scenario: The administrator enables or disables a platform section
- **WHEN** the channel editor submits its current sections
- **THEN** supported platforms are written explicitly with their current true or false values

#### Scenario: Global settings hide the control
- **WHEN** the search control is hidden by its global setting and the channel is saved
- **THEN** unrelated editing does not silently erase previously configured dimensions

