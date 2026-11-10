-- Acceleration settings default to existing request and billing behavior.
ALTER TABLE groups ADD COLUMN force_openai_fast BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN free_openai_fast BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE channel_model_pricing ADD COLUMN fast_multiplier REAL CHECK (fast_multiplier > 0);
