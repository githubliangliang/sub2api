-- [sqlite-converted] from PostgreSQL migration: 232_group_reasoning_effort_over_limit.sql
-- Per-group access control when an explicit OpenAI/Codex reasoning effort
-- exceeds max_reasoning_effort. Existing groups keep the previous behaviour
-- (automatically downgrade to the ceiling). DEFAULT must stay downgrade so
-- unmodified groups keep silent clamp after upgrade.

ALTER TABLE groups ADD COLUMN max_reasoning_effort_over_limit VARCHAR(20) NOT NULL DEFAULT 'downgrade';
