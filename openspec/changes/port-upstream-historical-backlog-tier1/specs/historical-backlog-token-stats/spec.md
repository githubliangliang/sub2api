# Historical Backlog Token Statistics

## ADDED Requirements

### Requirement: Existing token statistics honor the selected data scope

The existing statistics endpoint SHALL include models selected by its time, platform and group filters without requiring a gpt name prefix, and SHALL preserve duration-weighted throughput, pagination and existing API compatibility.

#### Scenario: OpenAI model has no gpt prefix
- **WHEN** the selected OpenAI scope contains both a gpt model and a non-gpt OpenAI model
- **THEN** both models SHALL contribute to the total and result rows
- **AND** existing grouping and weighted TPS semantics SHALL be retained

#### Scenario: Operator selects Claude or Grok
- **WHEN** a supported platform or group is selected in the existing card
- **THEN** returned rows and the count SHALL reflect that same scope
- **AND** card wording SHALL describe the scope accurately

#### Scenario: Existing pagination and attribution
- **WHEN** the query uses TopN, pagination, an empty window or rows without a group
- **THEN** limits, ordering, totals, half-open time boundaries and existing group/account attribution SHALL remain consistent
- **AND** the change SHALL NOT alter recorded usage, account cost or user charges
