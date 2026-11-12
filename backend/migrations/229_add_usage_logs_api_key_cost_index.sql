-- Cover per-key cost aggregation without fetching usage rows from the table.
CREATE INDEX IF NOT EXISTS idx_usage_logs_api_key_created_cost
    ON usage_logs (api_key_id, created_at, actual_cost);
