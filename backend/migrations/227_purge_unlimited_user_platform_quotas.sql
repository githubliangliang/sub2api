-- Unlimited platforms have no quota record. Keep every configured limit, including zero.
-- This deletes only historical placeholders; gateway usage logs and billing records are untouched.
-- Repeating the cleanup has no further effect.
DELETE FROM user_platform_quotas
WHERE daily_limit_usd IS NULL
  AND weekly_limit_usd IS NULL
  AND monthly_limit_usd IS NULL;
