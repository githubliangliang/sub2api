-- [sqlite-converted] from upstream 232_channel_cache_write_1h_pricing.sql
-- Channel custom pricing: split 1h cache-write from the existing 5m cache_write_price.
-- NULL on the new column keeps pre-split behavior: a lone cache_write_price
-- still covers both TTL tiers. Empty admin field must store NULL, not 0.

ALTER TABLE channel_model_pricing ADD COLUMN cache_write_1h_price NUMERIC(20,12);
ALTER TABLE channel_pricing_intervals ADD COLUMN cache_write_1h_price NUMERIC(20,12);
ALTER TABLE channel_account_stats_model_pricing ADD COLUMN cache_write_1h_price NUMERIC(20,12);
ALTER TABLE channel_account_stats_pricing_intervals ADD COLUMN cache_write_1h_price NUMERIC(20,12);
