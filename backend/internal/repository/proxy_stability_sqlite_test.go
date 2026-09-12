package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newProxyStabilitySQLiteRepo(t *testing.T) (*proxyRepository, *sql.DB) {
	t.Helper()
	db := newPGCompatSQLiteDB(t)
	require.NoError(t, ApplyMigrations(context.Background(), db))
	require.NoError(t, EnsureSQLiteAuxTables(context.Background(), db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	return newProxyRepositoryWithSQL(client, db), db
}

func createStabilityProxy(t *testing.T, repo *proxyRepository, name, mode string, expiry *time.Time, backup *int64) *service.Proxy {
	t.Helper()
	p := &service.Proxy{Name: name, Protocol: "http", Host: "localhost", Port: 8080, Status: service.StatusActive, FallbackMode: mode, ExpiresAt: expiry, BackupProxyID: backup, ExpiryWarnDays: 7}
	require.NoError(t, repo.Create(context.Background(), p))
	return p
}

func TestProxySQLiteBackupReferencesAreDirectedAndShared(t *testing.T) {
	repo, _ := newProxyStabilitySQLiteRepo(t)
	ctx := context.Background()
	backup := createStabilityProxy(t, repo, "backup", service.FallbackModeNone, nil, nil)
	first := createStabilityProxy(t, repo, "first", service.FallbackModeProxy, nil, &backup.ID)
	loaded, err := repo.GetByID(ctx, backup.ID)
	require.NoError(t, err)
	require.Nil(t, loaded.BackupProxyID, "a primary must not rewrite its backup's outgoing reference")
	second := createStabilityProxy(t, repo, "second", service.FallbackModeProxy, nil, &backup.ID)
	backup.Name = "renamed"
	require.NoError(t, repo.Update(ctx, backup))
	for _, id := range []int64{first.ID, second.ID} {
		got, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, &backup.ID, got.BackupProxyID)
	}
	last := createStabilityProxy(t, repo, "last", service.FallbackModeNone, nil, nil)
	backup.BackupProxyID, backup.FallbackMode = &last.ID, service.FallbackModeProxy
	require.NoError(t, repo.Update(ctx, backup))
	first.BackupProxyID = &last.ID
	require.NoError(t, repo.Update(ctx, first))
	first.BackupProxyID, first.FallbackMode = nil, service.FallbackModeNone
	require.NoError(t, repo.Update(ctx, first))
	got, err := repo.GetByID(ctx, second.ID)
	require.NoError(t, err)
	require.Equal(t, &backup.ID, got.BackupProxyID)
	got, err = repo.GetByID(ctx, backup.ID)
	require.NoError(t, err)
	require.Equal(t, &last.ID, got.BackupProxyID)
	got, err = repo.GetByID(ctx, last.ID)
	require.NoError(t, err)
	require.Nil(t, got.BackupProxyID)
}

func TestProxySQLiteLegacySymmetricBacklinkRemainsEditable(t *testing.T) {
	repo, _ := newProxyStabilitySQLiteRepo(t)
	ctx := context.Background()
	expiry := time.Now().UTC().Add(24 * time.Hour)
	// Use the unchanged legacy Ent builders to reproduce the old API writes.
	backup, err := repo.client.Proxy.Create().SetName("legacy-B").SetProtocol("http").SetHost("localhost").SetPort(8080).SetStatus(service.StatusActive).SetFallbackMode(service.FallbackModeNone).Save(ctx)
	require.NoError(t, err)
	primary, err := repo.client.Proxy.Create().SetName("legacy-A").SetProtocol("http").SetHost("localhost").SetPort(8081).SetStatus(service.StatusActive).SetFallbackMode(service.FallbackModeProxy).SetExpiresAt(expiry).SetBackupProxyID(backup.ID).Save(ctx)
	require.NoError(t, err)
	b, err := repo.GetByID(ctx, backup.ID)
	require.NoError(t, err)
	require.Equal(t, &primary.ID, b.BackupProxyID, "the old API silently created this inactive reverse reference")
	a, err := repo.GetByID(ctx, primary.ID)
	require.NoError(t, err)
	a.Name = "renamed-A"
	require.NoError(t, repo.Update(ctx, a), "a name-only edit of normal legacy data must succeed")
	a, err = repo.GetByID(ctx, primary.ID)
	require.NoError(t, err)
	require.Equal(t, &backup.ID, a.BackupProxyID)
	require.Equal(t, service.FallbackModeProxy, a.FallbackMode)
	require.True(t, a.ExpiresAt.Equal(expiry))
	b.Name = "renamed-B"
	require.NoError(t, repo.Update(ctx, b))
	c := createStabilityProxy(t, repo, "new-C", service.FallbackModeProxy, nil, &backup.ID)
	for _, id := range []int64{a.ID, c.ID} {
		got, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, &backup.ID, got.BackupProxyID)
	}
	b.FallbackMode = service.FallbackModeProxy
	require.Error(t, repo.Update(ctx, b), "activating the reverse reference would create an effective fallback cycle")
	unchanged, err := repo.GetByID(ctx, backup.ID)
	require.NoError(t, err)
	require.Equal(t, service.FallbackModeNone, unchanged.FallbackMode)
	missing := int64(99999)
	c.BackupProxyID = &missing
	require.ErrorIs(t, repo.Update(ctx, c), service.ErrProxyNotFound)
}

func TestProxySQLiteRejectsBackupCycles(t *testing.T) {
	repo, db := newProxyStabilitySQLiteRepo(t)
	ctx := context.Background()
	first := createStabilityProxy(t, repo, "first", service.FallbackModeProxy, nil, nil)
	second := createStabilityProxy(t, repo, "second", service.FallbackModeNone, nil, nil)
	_, err := db.ExecContext(ctx, `UPDATE proxies SET backup_proxy_id=$1 WHERE id=$2`, second.ID, first.ID)
	require.NoError(t, err)
	second.BackupProxyID, second.FallbackMode = &first.ID, service.FallbackModeProxy
	require.Error(t, repo.Update(ctx, second), "a directed cycle must not be persisted")
	first.BackupProxyID = &first.ID
	require.Error(t, repo.Update(ctx, first), "self backup must not be persisted")
}

func TestProxySQLiteRepeatedFallbackPreservesFirstOrigin(t *testing.T) {
	for _, direct := range []bool{false, true} {
		t.Run(map[bool]string{false: "next_backup", true: "direct"}[direct], func(t *testing.T) {
			repo, db := newProxyStabilitySQLiteRepo(t)
			ctx := context.Background()
			now := time.Now().UTC()
			past, soon, later := now.Add(-time.Hour), now.Add(time.Hour), now.Add(48*time.Hour)
			origin := createStabilityProxy(t, repo, "origin", service.FallbackModeProxy, &past, nil)
			middle := createStabilityProxy(t, repo, "middle", service.FallbackModeDirect, &soon, nil)
			last := createStabilityProxy(t, repo, "last", service.FallbackModeNone, &later, nil)
			_, err := db.ExecContext(ctx, `UPDATE proxies SET backup_proxy_id=$1 WHERE id=$2`, middle.ID, origin.ID)
			require.NoError(t, err)
			if !direct {
				_, err = db.ExecContext(ctx, `UPDATE proxies SET fallback_mode='proxy', backup_proxy_id=$1 WHERE id=$2`, last.ID, middle.ID)
				require.NoError(t, err)
			}
			account, err := repo.client.Account.Create().SetName("account").SetPlatform(service.PlatformOpenAI).SetType(service.AccountTypeAPIKey).SetCredentials(map[string]any{}).SetExtra(map[string]any{"upstream_billing_probe": map[string]any{"status": "ok"}}).SetProxyID(origin.ID).Save(ctx)
			require.NoError(t, err)
			changed, err := repo.SweepExpiredProxies(ctx, now)
			require.NoError(t, err)
			require.EqualValues(t, 1, changed)
			changed, err = repo.SweepExpiredProxies(ctx, now.Add(2*time.Hour))
			require.NoError(t, err)
			require.EqualValues(t, 1, changed, "a previously rerouted account must follow its current expired proxy")
			var current, original *int64
			var extra string
			require.NoError(t, db.QueryRowContext(ctx, `SELECT proxy_id, proxy_fallback_origin_id, extra FROM accounts WHERE id=$1`, account.ID).Scan(&current, &original, &extra))
			require.Equal(t, &origin.ID, original)
			if direct {
				require.Nil(t, current)
			} else {
				require.Equal(t, &last.ID, current)
			}
			require.JSONEq(t, `{}`, extra)
			var payload string
			require.NoError(t, db.QueryRowContext(ctx, `SELECT payload FROM scheduler_outbox WHERE event_type=$1 ORDER BY id DESC LIMIT 1`, service.SchedulerOutboxEventAccountBulkChanged).Scan(&payload))
			var event struct {
				AccountIDs []int64 `json:"account_ids"`
			}
			require.NoError(t, json.Unmarshal([]byte(payload), &event))
			require.Equal(t, []int64{account.ID}, event.AccountIDs)
			changed, err = repo.SweepExpiredProxies(ctx, now.Add(3*time.Hour))
			require.NoError(t, err)
			require.Zero(t, changed)
			require.NoError(t, newAccountRepositoryWithSQL(repo.client, db, nil).RevertProxyFallback(ctx, account.ID))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT proxy_id, proxy_fallback_origin_id FROM accounts WHERE id=$1`, account.ID).Scan(&current, &original))
			require.Equal(t, &origin.ID, current)
			require.Nil(t, original)
		})
	}
}
