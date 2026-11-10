## ADDED Requirements

### Requirement: Consistent accelerated request tiers
The gateway SHALL preserve supported explicit tiers across native and compatible HTTP and WebSocket routes and apply configured policies consistently.

#### Scenario: Explicit Ultrafast
- **WHEN** an allowed request selects ultrafast
- **THEN** the upstream receives ultrafast and usage records retain its effective tier.

### Requirement: Separate downstream Free Fast cost
The gateway SHALL bill eligible Free Fast requests at Standard downstream while retaining accelerated upstream cost.

#### Scenario: Free Fast enabled
- **WHEN** an OpenAI priority request belongs to an eligible group with Free Fast enabled
- **THEN** downstream token cost uses Standard prices and upstream token cost retains priority prices.

### Requirement: Durable acceleration configuration
The system SHALL persist acceleration configuration in SQLite and propagate it through cached authentication and admin forms.

#### Scenario: Save and reload
- **WHEN** an administrator updates acceleration configuration
- **THEN** subsequent requests use the saved settings after cache invalidation.

### Requirement: Explicit tier validation
The gateway SHALL reject invalid explicit service tiers before forwarding on HTTP and response.create WebSocket requests.

#### Scenario: Invalid Messages compatibility tier
- **WHEN** an OpenAI-routed Messages request specifies an unknown, empty, or non-string service tier
- **THEN** it returns a client error without an upstream request.

### Requirement: Account-aware tier settlement
OpenAI billing SHALL retain the requested tier when the response would raise cost and SHALL treat Codex OAuth default echoes as non-authoritative.

#### Scenario: Public API downgrade
- **WHEN** a public API Ultrafast request is served at priority
- **THEN** the usage row is settled at priority.

#### Scenario: Codex default echo
- **WHEN** a Codex OAuth Fast request returns default
- **THEN** the priority request tier remains the billable tier.

### Requirement: Account-owned catalog capabilities
The model catalog SHALL preserve explicit account tier declarations, including empty lists, without deriving OAuth Ultrafast access from subscription names.

#### Scenario: Explicit empty upstream tier list
- **WHEN** account metadata supplies an empty tier list
- **THEN** fallback catalog construction does not add Ultrafast.

### Requirement: Optional pricing overrides
Fast pricing overrides SHALL accept finite positive multipliers, preserve explicit zero token prices, and leave model-owned Ultrafast multipliers independent.

#### Scenario: Invalid multiplier
- **WHEN** an administrator supplies a zero or negative Fast multiplier
- **THEN** the API returns HTTP 400 without persisting the invalid configuration.
