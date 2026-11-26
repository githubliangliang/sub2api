# Historical Followup Request Ownership

## ADDED Requirements

### Requirement: User edits belong to the current loaded object and session

Existing user attribute and allowed-group editors SHALL apply asynchronous state and saves only to the current object and dialog session, and SHALL distinguish unavailable data from explicit empty edits.

#### Scenario: User selection changes before an attribute response arrives
- **WHEN** a response for a previous user resolves after another user becomes current
- **THEN** it SHALL NOT replace or emit the current user's attribute values
- **AND** stale errors and completion callbacks SHALL NOT affect the current request state

#### Scenario: Group configuration fails to load
- **WHEN** the current group configuration has not loaded successfully
- **THEN** both UI and method-level save paths SHALL prevent an empty replacement from being submitted
- **AND** a later successful load SHALL permit normal saving with established public and exclusive group semantics

#### Scenario: Dialog changes during save
- **WHEN** a user editor closes, reopens or changes object while a save is pending
- **THEN** the submitted object and values SHALL remain tied to the originating session
- **AND** its completion SHALL NOT modify a later session

### Requirement: Shared image reads belong to the active selection

The existing shared image-upload component SHALL cancel or ignore reads that no longer belong to the active selection while retaining sanitization and file limits.

#### Scenario: Earlier read finishes after a newer image
- **WHEN** an earlier image read completes after another file was selected
- **THEN** neither its value nor its error SHALL replace current state

#### Scenario: Image is removed or component unmounts
- **WHEN** the operator removes an image or leaves the component while a read is pending
- **THEN** a late callback SHALL NOT restore the image or update obsolete component state
- **AND** normal single-file image and SVG processing SHALL retain existing sanitization and size/type checks

### Requirement: Template association and apply state belong to one dialog session

The existing administrator template picker SHALL keep associated monitors, selected IDs, loading, errors and apply completion tied to the current template and dialog session.

#### Scenario: Templates are switched during loading
- **WHEN** an earlier template's association request resolves after another template is current
- **THEN** the old rows and default selection SHALL NOT replace the current template's state
- **AND** stale errors or loading completion SHALL be ignored

#### Scenario: An old apply finishes after reopening
- **WHEN** a submitted apply operation finishes after the picker has closed or reopened
- **THEN** it SHALL retain the originating template and IDs without closing or modifying the newer session
- **AND** current-session normal apply and duplicate-submit protection SHALL remain functional
