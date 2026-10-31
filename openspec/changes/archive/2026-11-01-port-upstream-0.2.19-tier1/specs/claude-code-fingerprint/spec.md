## ADDED Requirements

### Requirement: Claude Code fingerprints use JavaScript text semantics
The system SHALL sample the first user text at zero-based UTF-16 code-unit indexes 4, 7 and 20, join the sampled units, encode that sequence using JavaScript UTF-8 semantics, and retain the existing salt/version/hash format.

#### Scenario: User text contains multibyte characters
- **WHEN** a Claude Code request contains Chinese characters or emoji before a sampled index
- **THEN** the fingerprint matches an independent JavaScript oracle on every existing body/header consumer

#### Scenario: Sampled units include surrogate halves
- **WHEN** sampled UTF-16 units are paired or unpaired after joining
- **THEN** their hash input follows JavaScript encoding of that joined sequence

#### Scenario: Text is short or very long
- **WHEN** the text is shorter than a sampled index or extends far beyond index 20
- **THEN** absent units use the existing zero-character padding and sampling reads only the required prefix
- **AND** ASCII output and repeated version synchronization remain stable

