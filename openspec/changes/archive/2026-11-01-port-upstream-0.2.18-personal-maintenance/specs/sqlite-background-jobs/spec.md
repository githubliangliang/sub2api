## ADDED Requirements

### Requirement: Existing background jobs run on SQLite
The system SHALL start usage cleanup, account expiry and scheduled account test workers on SQLite according to their existing configuration and plan controls.

#### Scenario: Restart resumes a requested cleanup task
- **GIVEN** cleanup is enabled and a pending or stale running task exists
- **WHEN** the SQLite service starts and its cleanup worker runs
- **THEN** it processes the task's recorded filters in bounded batches and updates task state
- **AND** quota balances and billing dedup records are not recreated or reset by log removal

#### Scenario: Cleanup is disabled
- **GIVEN** usage cleanup is disabled
- **WHEN** services start or a new cleanup task is requested
- **THEN** no cleanup work executes and task creation reports the existing disabled response

### Requirement: Expired accounts are paused with consistent scheduling state
The system SHALL pause only undeleted, currently schedulable accounts with auto_pause_on_expired enabled and a reached expiry time, and propagate the change to scheduling.

#### Scenario: Expiry is reached
- **GIVEN** an eligible account expires at or before the worker's current time
- **WHEN** the expiry worker runs
- **THEN** the account becomes unschedulable and its scheduling snapshot is updated
- **AND** repeated runs do not produce another pause for the same account

#### Scenario: Account is outside the configured expiry rule
- **GIVEN** an account is undeleted and expires later, has no expiry, has auto-pause disabled, or is already unschedulable
- **WHEN** the worker runs
- **THEN** its scheduling state is not changed

### Requirement: Scheduled tests persist their lifecycle on SQLite
The system SHALL execute enabled due plans, persist test results, apply configured result retention and advance the plan's next run.

#### Scenario: Due test completes
- **GIVEN** an enabled plan is due
- **WHEN** the worker executes the existing account-test path
- **THEN** the result and next run are saved
- **AND** automatic recovery follows the existing success and recoverable-state rules

#### Scenario: Disabled or future plan
- **GIVEN** a plan is disabled or scheduled for the future
- **WHEN** due plans are listed
- **THEN** it is not selected for testing

### Requirement: Worker time and cancellation boundaries are supported
The system SHALL read supported legacy TEXT timestamps and native DATETIME values on the newly active paths, prevent overlapping execution, and stop background work when shutdown is requested.

#### Scenario: Legacy time value is read
- **GIVEN** a task or plan timestamp is returned as a string or byte slice
- **WHEN** it is scanned or compared to the due time
- **THEN** its supported UTC or offset time is honored without a Scan(*time.Time) type error

#### Scenario: Worker stops during waiting or execution
- **WHEN** shutdown cancels a worker
- **THEN** it stops waiting or cancels current work and does not execute another batch or plan
