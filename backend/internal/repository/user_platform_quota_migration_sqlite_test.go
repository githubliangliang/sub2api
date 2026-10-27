package repository

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestUserPlatformQuotaMigration227UpgradesRealSQLite226Database(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	baseline := fstest.MapFS{}
	entries, err := migrations.FS.ReadDir(".")
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.Name() == "227_purge_unlimited_user_platform_quotas.sql" {
			continue
		}
		data, err := migrations.FS.ReadFile(entry.Name())
		require.NoError(t, err)
		baseline[entry.Name()] = &fstest.MapFile{Data: data, Mode: 0444}
	}
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.Exec(`INSERT INTO users (id,email,password_hash,role) VALUES (90001,'upgrade@example.invalid','fixture-hash','user');
		INSERT INTO user_platform_quotas (user_id,platform,daily_limit_usd,weekly_limit_usd,monthly_limit_usd,daily_usage_usd,deleted_at) VALUES
		(90001,'anthropic',NULL,NULL,NULL,9,NULL),
		(90001,'openai',0,NULL,NULL,3,NULL),
		(90001,'gemini',NULL,10,NULL,5,NULL),
		(90001,'antigravity',NULL,NULL,0,17,NULL),
		(90001,'anthropic',NULL,NULL,NULL,7,'2026-10-28');`)
	require.NoError(t, err)
	require.NoError(t, ApplyMigrations(ctx, db))
	var count, applied int
	var usage float64
	require.NoError(t, db.QueryRow(`SELECT COUNT(*),SUM(daily_usage_usd) FROM user_platform_quotas`).Scan(&count, &usage))
	require.Equal(t, 3, count)
	require.Equal(t, 25.0, usage)
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE filename='227_purge_unlimited_user_platform_quotas.sql'`).Scan(&applied))
	require.Equal(t, 1, applied)
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM user_platform_quotas`).Scan(&count))
	require.Equal(t, 3, count)
}
