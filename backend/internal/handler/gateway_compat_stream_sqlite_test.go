//go:build unit

package handler

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCompatibilityPartialUsageSettlesDeletedKeyOnceSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, endpoint := range []string{"chat/completions", "responses"} {
		for _, tc := range []struct {
			name   string
			remove bool
			cancel bool
		}{
			{"soft deleted", false, false}, {"missing", true, false}, {"canceled request", false, true},
		} {
			t.Run(endpoint+"/"+tc.name, func(t *testing.T) {
				db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared&_time_format=sqlite", t.Name()))
				require.NoError(t, err)
				defer db.Close()
				db.SetMaxOpenConns(1)
				for _, statement := range []string{
					`CREATE TABLE usage_billing_dedup (id INTEGER PRIMARY KEY AUTOINCREMENT, request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT, UNIQUE(request_id, api_key_id))`,
					`CREATE TABLE usage_billing_dedup_archive (request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT)`,
					`CREATE TABLE users (id INTEGER PRIMARY KEY, balance REAL, frozen_balance REAL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
					`CREATE TABLE api_keys (id INTEGER PRIMARY KEY, quota REAL, quota_used REAL, status TEXT, usage_5h REAL, usage_1d REAL, usage_7d REAL, window_5h_start DATETIME, window_1d_start DATETIME, window_7d_start DATETIME, deleted_at DATETIME, updated_at DATETIME)`,
					`CREATE TABLE accounts (id INTEGER PRIMARY KEY, extra TEXT, deleted_at DATETIME, updated_at DATETIME)`,
					`INSERT INTO users (id, balance) VALUES (9303, 100)`,
					`INSERT INTO api_keys (id, quota, quota_used, status, usage_5h, usage_1d, usage_7d, deleted_at) VALUES (9302, 20, 0, 'active', 0, 0, 0, CURRENT_TIMESTAMP)`,
					`INSERT INTO accounts (id, extra) VALUES (9301, '{"quota_limit":10}')`,
				} {
					_, err := db.Exec(statement)
					require.NoError(t, err)
				}
				if tc.remove {
					_, err := db.Exec(`DELETE FROM api_keys WHERE id=9302`)
					require.NoError(t, err)
				}
				ledger := repository.NewUsageBillingRepository(nil, db)
				stream := compatStreamUsagePartialStream
				if tc.cancel {
					stream += compatStreamUsageCompletion
				}
				run := runCompatStreamUsageHandlerWithLedger(t, endpoint, stream, true, ledger, tc.cancel)
				require.Len(t, run.billingRepo.commands, 1)
				cmd := run.billingRepo.commands[0]
				require.Equal(t, 12, cmd.InputTokens)
				require.Positive(t, cmd.BalanceCost)
				require.Positive(t, cmd.AccountQuotaCost)
				var balance, quota float64
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=9303`).Scan(&balance))
				require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=9301`).Scan(&quota))
				require.InDelta(t, 100-cmd.BalanceCost, balance, 1e-10)
				require.InDelta(t, cmd.AccountQuotaCost, quota, 1e-10)
				// Replay through the actual handler and recorder; the same request
				// must not settle the user/account a second time, even after deletion.
				replay := runCompatStreamUsageHandlerWithLedger(t, endpoint, stream, true, ledger, tc.cancel)
				require.Len(t, replay.billingRepo.commands, 1)
				var afterBalance, afterQuota float64
				var dedup int
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=9303`).Scan(&afterBalance))
				require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=9301`).Scan(&afterQuota))
				require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
				require.Equal(t, balance, afterBalance)
				require.Equal(t, quota, afterQuota)
				require.Equal(t, 1, dedup)
			})
		}
	}
}
