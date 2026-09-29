package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTier2RevertProxyInvalidatesOnlyChangedIdentitySQLite(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(strconv.FormatBool(changed), func(t *testing.T) {
			repo, db := newProxyStabilitySQLiteRepo(t)
			origin := createStabilityProxy(t, repo, "origin", service.FallbackModeNone, nil, nil)
			fallback := createStabilityProxy(t, repo, "fallback", service.FallbackModeNone, nil, nil)
			current := origin.ID
			if changed {
				current = fallback.ID
			}
			_, err := db.Exec(`INSERT INTO accounts (id,name,platform,type,credentials,extra,proxy_id,proxy_fallback_origin_id) VALUES (900,'probe','openai','apikey','{}','{"upstream_billing_probe":{"rate":1},"keep":"value"}',?,?)`, current, origin.ID)
			require.NoError(t, err)
			accounts := newAccountRepositoryWithSQL(repo.client, db, nil)
			require.NoError(t, accounts.RevertProxyFallback(context.Background(), 900))
			var extra string
			var proxyID int64
			require.NoError(t, db.QueryRow(`SELECT extra,proxy_id FROM accounts WHERE id=900`).Scan(&extra, &proxyID))
			require.Equal(t, origin.ID, proxyID)
			if changed {
				require.JSONEq(t, `{"keep":"value"}`, extra)
			} else {
				require.JSONEq(t, `{"upstream_billing_probe":{"rate":1},"keep":"value"}`, extra)
			}
		})
	}
}

func TestTier2ProxySweepRejectsStaleSnapshotSQLite(t *testing.T) {
	for _, mutation := range []string{"renew", "disable", "fallback", "backup", "target_disabled", "unchanged"} {
		t.Run(mutation, func(t *testing.T) {
			repo, db := newProxyStabilitySQLiteRepo(t)
			now := time.Now().UTC()
			expired := now.Add(-time.Hour)
			backup := createStabilityProxy(t, repo, "backup", service.FallbackModeNone, nil, nil)
			other := createStabilityProxy(t, repo, "other", service.FallbackModeNone, nil, nil)
			origin := createStabilityProxy(t, repo, "origin", service.FallbackModeProxy, &expired, &backup.ID)
			snapshot := *origin
			_, err := db.Exec(`INSERT INTO accounts (id,name,platform,type,credentials,proxy_id) VALUES (901,'sweep','openai','apikey','{}',?)`, origin.ID)
			require.NoError(t, err)
			switch mutation {
			case "renew":
				_, err = db.Exec(`UPDATE proxies SET expires_at=? WHERE id=?`, now.Add(time.Hour), origin.ID)
			case "disable":
				_, err = db.Exec(`UPDATE proxies SET status='inactive' WHERE id=?`, origin.ID)
			case "fallback":
				_, err = db.Exec(`UPDATE proxies SET fallback_mode='none' WHERE id=?`, origin.ID)
			case "backup":
				_, err = db.Exec(`UPDATE proxies SET backup_proxy_id=? WHERE id=?`, other.ID, origin.ID)
			case "target_disabled":
				_, err = db.Exec(`UPDATE proxies SET status='inactive' WHERE id=?`, backup.ID)
			}
			require.NoError(t, err)
			ids, err := repo.sweepOneExpiredProxy(context.Background(), snapshot, now, &backup.ID, true)
			require.NoError(t, err)
			var current int64
			require.NoError(t, db.QueryRow(`SELECT proxy_id FROM accounts WHERE id=901`).Scan(&current))
			if mutation == "unchanged" {
				require.Equal(t, []int64{901}, ids)
				require.Equal(t, backup.ID, current)
			} else {
				require.Empty(t, ids)
				require.Equal(t, origin.ID, current)
			}
		})
	}
}
