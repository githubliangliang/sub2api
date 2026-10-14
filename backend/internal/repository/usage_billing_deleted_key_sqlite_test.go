package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageBillingDeletedKeySQLite(t *testing.T) {
	for _, tc := range []struct {
		name, setup string
		balanceCost float64
		wantError   bool
	}{
		{"soft deleted", `UPDATE api_keys SET deleted_at = CURRENT_TIMESTAMP`, 1.25, false},
		{"missing", `DELETE FROM api_keys`, 1.25, false},
		{"simple mode command", `UPDATE api_keys SET deleted_at = CURRENT_TIMESTAMP`, 0, false},
		{"quota database error", `CREATE TRIGGER fail_quota BEFORE UPDATE OF quota_used ON api_keys BEGIN SELECT RAISE(ABORT, 'quota failure'); END`, 1.25, true},
		{"rate database error", `CREATE TRIGGER fail_rate BEFORE UPDATE OF usage_5h ON api_keys BEGIN SELECT RAISE(ABORT, 'rate failure'); END`, 1.25, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
				`INSERT INTO users (id, balance) VALUES (1, 10)`,
				`INSERT INTO api_keys (id, quota, quota_used, status, usage_5h, usage_1d, usage_7d) VALUES (2, 20, 0, 'active', 0, 0, 0)`,
				`INSERT INTO accounts (id, extra) VALUES (3, '{"quota_limit":10}')`,
				tc.setup,
			} {
				_, err = db.Exec(statement)
				require.NoError(t, err)
			}
			repo := NewUsageBillingRepository(nil, db)
			cmd := &service.UsageBillingCommand{
				RequestID: "deleted-key", APIKeyID: 2, UserID: 1, AccountID: 3,
				AccountType: service.AccountTypeAPIKey, BalanceCost: tc.balanceCost,
				APIKeyQuotaCost: 2, APIKeyRateLimitCost: 3, AccountQuotaCost: 4,
			}
			result, err := repo.Apply(context.Background(), cmd)
			wantBalance, wantQuota, wantDedup := 10.0, 0.0, 0
			if tc.wantError {
				require.Error(t, err)
				require.NotErrorIs(t, err, service.ErrAPIKeyNotFound)
			} else {
				require.NoError(t, err)
				require.True(t, result.Applied)
				require.False(t, result.APIKeyQuotaExhausted)
				wantBalance, wantQuota, wantDedup = 10-tc.balanceCost, 4, 1
				duplicate, err := repo.Apply(context.Background(), cmd)
				require.NoError(t, err)
				require.False(t, duplicate.Applied)
			}
			var balance, quota, keyQuota, keyRate float64
			var dedup int
			require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
			require.NoError(t, db.QueryRow(`SELECT COALESCE(json_extract(extra, '$.quota_used'), 0) FROM accounts WHERE id=3`).Scan(&quota))
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
			require.NoError(t, db.QueryRow(`SELECT COALESCE(SUM(quota_used), 0), COALESCE(SUM(usage_5h+usage_1d+usage_7d), 0) FROM api_keys`).Scan(&keyQuota, &keyRate))
			require.InDelta(t, wantBalance, balance, 1e-9)
			require.InDelta(t, wantQuota, quota, 1e-9)
			require.Equal(t, wantDedup, dedup)
			require.Zero(t, keyQuota)
			require.Zero(t, keyRate)
		})
	}
}
