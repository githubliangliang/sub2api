package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestUsageBillingRepositoryApplySQLiteQuotaWindows(t *testing.T) {
	for _, tt := range []struct {
		name     string
		mode     string
		resetAt  string
		timezone string
		active   bool
		disabled bool
	}{
		{name: "fixed_expired_utc", mode: "fixed", resetAt: "2020-01-01T00:00:00Z", timezone: "UTC"},
		{name: "fixed_expired_shanghai", mode: "fixed", resetAt: "2020-01-01T00:00:00Z", timezone: "Asia/Shanghai"},
		{name: "fixed_missing_reset", mode: "fixed", timezone: "UTC"},
		{name: "fixed_invalid_reset", mode: "fixed", resetAt: "invalid", timezone: "UTC"},
		{name: "fixed_active", mode: "fixed", timezone: "Asia/Shanghai", active: true},
		{name: "rolling_expired", mode: "rolling", timezone: "UTC"},
		{name: "rolling_active", mode: "rolling", timezone: "UTC", active: true},
		{name: "disabled", mode: "fixed", timezone: "UTC", disabled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			db.SetMaxOpenConns(1)
			for _, statement := range []string{
				`CREATE TABLE usage_billing_dedup (id INTEGER PRIMARY KEY AUTOINCREMENT, request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT, UNIQUE(request_id, api_key_id))`,
				`CREATE TABLE usage_billing_dedup_archive (request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT)`,
				`CREATE TABLE accounts (id INTEGER PRIMARY KEY, extra TEXT, deleted_at DATETIME, updated_at DATETIME)`,
				`CREATE TABLE scheduler_outbox (id INTEGER PRIMARY KEY AUTOINCREMENT, event_type TEXT, account_id INTEGER, group_id INTEGER, payload TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, dedup_key TEXT UNIQUE)`,
			} {
				_, err = db.Exec(statement)
				require.NoError(t, err)
			}

			now := time.Now().UTC()
			location, err := time.LoadLocation(tt.timezone)
			require.NoError(t, err)
			// Keep reset boundaries away from the test's execution time.
			resetHour := (now.In(location).Hour() + 2) % 24
			resetDay := (int(now.In(location).Weekday()) + 1) % 7
			start := now.AddDate(0, 0, -14).Format(time.RFC3339Nano)
			if tt.active {
				start = now.Format(time.RFC3339Nano)
			}
			extra := map[string]any{
				"quota_used": 100.0, "quota_daily_used": 9.0, "quota_weekly_used": 19.0,
				"quota_daily_limit": 3.0, "quota_weekly_limit": 3.0,
				"quota_daily_reset_mode": tt.mode, "quota_weekly_reset_mode": tt.mode,
				"quota_daily_reset_hour": resetHour, "quota_weekly_reset_hour": resetHour,
				"quota_weekly_reset_day": resetDay, "quota_reset_timezone": tt.timezone,
				"quota_daily_start": start, "quota_weekly_start": start,
				"unrelated_setting": "preserved",
			}
			if tt.disabled {
				extra["quota_daily_limit"], extra["quota_weekly_limit"] = 0.0, 0.0
			}
			if tt.resetAt != "" {
				extra["quota_daily_reset_at"], extra["quota_weekly_reset_at"] = tt.resetAt, tt.resetAt
			}
			if tt.mode == "fixed" && tt.active {
				service.ComputeQuotaResetAt(extra)
			}
			encoded, err := json.Marshal(extra)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO accounts (id, extra) VALUES (3, $1)`, string(encoded))
			require.NoError(t, err)

			readExtra := func() map[string]any {
				t.Helper()
				var raw string
				require.NoError(t, db.QueryRow(`SELECT extra FROM accounts WHERE id = 3`).Scan(&raw))
				var got map[string]any
				require.NoError(t, json.Unmarshal([]byte(raw), &got))
				return got
			}
			repo := NewUsageBillingRepository(nil, db)
			wantDaily, wantWeekly := 0.0, 0.0
			if tt.active || tt.disabled {
				wantDaily, wantWeekly = 9, 19
			}
			var previous map[string]any
			for i, amount := range []float64{1.25, 2.5} {
				cmd := &service.UsageBillingCommand{
					RequestID: fmt.Sprintf("quota-%d", i), APIKeyID: 2, UserID: 1,
					AccountID: 3, AccountType: service.AccountTypeAPIKey, AccountQuotaCost: amount,
				}
				result, err := repo.Apply(context.Background(), cmd)
				require.NoError(t, err)
				require.True(t, result.Applied)
				require.NotNil(t, result.QuotaState)
				if !tt.disabled {
					wantDaily += amount
					wantWeekly += amount
				}
				require.InDelta(t, wantDaily, result.QuotaState.DailyUsed, 1e-9)
				require.InDelta(t, wantWeekly, result.QuotaState.WeeklyUsed, 1e-9)
				got := readExtra()
				require.InDelta(t, wantDaily, got["quota_daily_used"], 1e-9)
				require.InDelta(t, wantWeekly, got["quota_weekly_used"], 1e-9)
				require.Equal(t, "preserved", got["unrelated_setting"])
				for _, dimension := range []string{"daily", "weekly"} {
					resetKey := "quota_" + dimension + "_reset_at"
					startKey := "quota_" + dimension + "_start"
					if tt.mode == "fixed" && !tt.disabled && i > 0 {
						resetString, ok := got[resetKey].(string)
						require.True(t, ok, "fixed windows must persist their next reset")
						resetAt, err := time.Parse(time.RFC3339, resetString)
						require.NoError(t, err)
						require.True(t, resetAt.After(now))
						require.Equal(t, resetHour, resetAt.In(location).Hour())
						if dimension == "weekly" {
							require.Equal(t, time.Weekday(resetDay), resetAt.In(location).Weekday())
						}
						if tt.active {
							require.Equal(t, extra[resetKey], got[resetKey])
						}
					} else if tt.mode == "rolling" {
						require.NotContains(t, got, resetKey)
					}
					if previous != nil {
						require.Equal(t, previous[resetKey], got[resetKey])
						require.Equal(t, previous[startKey], got[startKey])
					}
				}
				duplicate, err := repo.Apply(context.Background(), cmd)
				require.NoError(t, err)
				require.False(t, duplicate.Applied)
				require.Equal(t, got, readExtra(), "duplicate billing must leave quota state unchanged")
				previous = got
			}
			require.InDelta(t, 103.75, previous["quota_used"], 1e-9)
			var outboxCount int
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM scheduler_outbox`).Scan(&outboxCount))
			if tt.active || tt.disabled {
				require.Zero(t, outboxCount)
			} else {
				require.Equal(t, 1, outboxCount, "crossing the new window's quota must refresh scheduling")
			}
		})
	}
}

func TestUpdateFixedQuotaResetAtWeeklyDefault(t *testing.T) {
	for _, tt := range []struct {
		name string
		day  any
		want string
	}{
		{name: "default_monday", want: "2026-10-18T16:00:00Z"},
		{name: "explicit_sunday", day: float64(0), want: "2026-10-17T16:00:00Z"},
		{name: "explicit_monday", day: float64(1), want: "2026-10-18T16:00:00Z"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			extra := map[string]any{
				"quota_weekly_reset_mode": "fixed", "quota_reset_timezone": "Asia/Shanghai",
			}
			if tt.day != nil {
				extra["quota_weekly_reset_day"] = tt.day
			}
			updateFixedQuotaResetAt(extra, "weekly", time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC))
			require.Equal(t, tt.want, extra["quota_weekly_reset_at"])
		})
	}
}

func TestUsageBillingRepositoryApplySQLiteAppliesAllEffects(t *testing.T) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	for _, statement := range []string{
		`CREATE TABLE usage_billing_dedup (id INTEGER PRIMARY KEY AUTOINCREMENT, request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT, UNIQUE(request_id, api_key_id))`,
		`CREATE TABLE usage_billing_dedup_archive (request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT)`,
		`CREATE TABLE users (id INTEGER PRIMARY KEY, balance REAL, frozen_balance REAL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE api_keys (id INTEGER PRIMARY KEY, quota REAL, quota_used REAL, status TEXT, usage_5h REAL, usage_1d REAL, usage_7d REAL, window_5h_start DATETIME, window_1d_start DATETIME, window_7d_start DATETIME, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE accounts (id INTEGER PRIMARY KEY, extra TEXT, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE scheduler_outbox (id INTEGER PRIMARY KEY AUTOINCREMENT, event_type TEXT, account_id INTEGER, group_id INTEGER, payload TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, dedup_key TEXT UNIQUE)`,
		`INSERT INTO users (id, balance) VALUES (1, 10)`,
		`INSERT INTO api_keys (id, quota, quota_used, status, usage_5h, usage_1d, usage_7d, window_5h_start, window_1d_start, window_7d_start) VALUES (2, 2, 0, 'active', 0, 0, 0, '2020-01-01', '2020-01-01', '2020-01-01')`,
		`INSERT INTO accounts (id, extra) VALUES (3, '{"quota_limit":10,"quota_daily_limit":5,"quota_weekly_limit":8}')`,
	} {
		_, err = db.Exec(statement)
		require.NoError(t, err)
	}

	repo := NewUsageBillingRepository(nil, db)
	result, err := repo.Apply(context.Background(), &service.UsageBillingCommand{
		RequestID:           "sqlite-all-effects",
		APIKeyID:            2,
		UserID:              1,
		AccountID:           3,
		AccountType:         service.AccountTypeAPIKey,
		BalanceCost:         1.25,
		APIKeyQuotaCost:     2,
		APIKeyRateLimitCost: 3,
		AccountQuotaCost:    4,
	})
	require.NoError(t, err)
	require.True(t, result.Applied)
	require.True(t, result.APIKeyQuotaExhausted)
	require.NotNil(t, result.QuotaState)
	require.InDelta(t, 4, result.QuotaState.TotalUsed, 0.000001)
	require.InDelta(t, 4, result.QuotaState.DailyUsed, 0.000001)
	require.InDelta(t, 4, result.QuotaState.WeeklyUsed, 0.000001)

	var balance, quotaUsed, usage5h float64
	var status string
	require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id = 1`).Scan(&balance))
	require.NoError(t, db.QueryRow(`SELECT quota_used, status, usage_5h FROM api_keys WHERE id = 2`).Scan(&quotaUsed, &status, &usage5h))
	require.InDelta(t, 8.75, balance, 0.000001)
	require.InDelta(t, 2, quotaUsed, 0.000001)
	require.Equal(t, service.StatusAPIKeyQuotaExhausted, status)
	require.InDelta(t, 3, usage5h, 0.000001)

	var dailyStart, weeklyStart string
	require.NoError(t, db.QueryRow(`SELECT json_extract(extra, '$.quota_daily_start'), json_extract(extra, '$.quota_weekly_start') FROM accounts WHERE id = 3`).Scan(&dailyStart, &weeklyStart))
	_, err = time.Parse(time.RFC3339Nano, dailyStart)
	require.NoError(t, err)
	_, err = time.Parse(time.RFC3339Nano, weeklyStart)
	require.NoError(t, err)

	duplicate, err := repo.Apply(context.Background(), &service.UsageBillingCommand{
		RequestID: "sqlite-all-effects", APIKeyID: 2, UserID: 1,
		AccountID: 3, AccountType: service.AccountTypeAPIKey,
		BalanceCost: 1.25, APIKeyQuotaCost: 2, APIKeyRateLimitCost: 3, AccountQuotaCost: 4,
	})
	require.NoError(t, err)
	require.False(t, duplicate.Applied)
}
