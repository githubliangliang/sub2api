## ADDED Requirements

### Requirement: Compatible usage recognizes prompt cache hits
The system SHALL recognize prompt_cache_hit_tokens as a top-level cache-read fallback without summing multiple alternative representations or replacing canonical nested values.

#### Scenario: Only top-level cache hit counters are supplied
- **WHEN** compatible usage contains prompt_tokens 100, prompt_cache_hit_tokens 80 and prompt_cache_miss_tokens 20
- **THEN** total input and cache-read usage retain their respective values under the existing token-bucket contract

#### Scenario: Canonical nested cache usage is supplied
- **WHEN** a nested cached_tokens field is explicitly present, including zero
- **THEN** it takes precedence over the new top-level fallback

#### Scenario: Usage reaches storage
- **WHEN** the same successful request passes through HTTP or SSE parsing and settlement
- **THEN** the token buckets agree and the existing SQLite billing dedup prevents duplicate settlement

### Requirement: Existing Gemini image accounting recognizes file references
The system SHALL count image parts represented by inlineData, inline_data, fileData or file_data when image MIME and a nonempty string data or URI are present.

#### Scenario: A custom image alias returns file references
- **WHEN** an existing Gemini-compatible path observes two valid file-reference image parts under an unknown model alias
- **THEN** two actual outputs are retained instead of zero

#### Scenario: A part has multiple spellings or empty content
- **WHEN** one part contains multiple recognized representations, an empty URI, or a non-image MIME
- **THEN** the part is counted at most once and invalid image shells are not counted

#### Scenario: SSE repeats accumulated content
- **WHEN** repeated chunks contain accumulated image parts or the client disconnects after usage was observed
- **THEN** the counter retains the largest observed count and the established interrupted-usage contract
- **AND** no URI download or new endpoint/billing mode is enabled

