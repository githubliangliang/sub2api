package migrations

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigration227PurgesOnlyUnlimitedPlatformQuotasSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	_, err = db.Exec(`CREATE TABLE user_platform_quotas (
		id INTEGER PRIMARY KEY, daily_limit_usd REAL, weekly_limit_usd REAL,
		monthly_limit_usd REAL, daily_usage_usd REAL, deleted_at DATETIME);
		INSERT INTO user_platform_quotas VALUES
		(1,NULL,NULL,NULL,9,NULL),(2,0,NULL,NULL,3,NULL),(3,NULL,0,NULL,4,NULL),
		(4,NULL,NULL,0,5,NULL),(5,10,NULL,NULL,6,NULL),(6,NULL,NULL,NULL,7,'2026-10-28'),
		(7,NULL,20,NULL,8,'2026-10-28');`)
	require.NoError(t, err)
	query, err := FS.ReadFile("227_purge_unlimited_user_platform_quotas.sql")
	require.NoError(t, err)
	result, err := db.Exec(string(query))
	require.NoError(t, err)
	removed, err := result.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 2, removed)
	var count int
	var usage float64
	require.NoError(t, db.QueryRow("SELECT COUNT(*), SUM(daily_usage_usd) FROM user_platform_quotas").Scan(&count, &usage))
	require.Equal(t, 5, count)
	require.Equal(t, 26.0, usage)
	result, err = db.Exec(string(query))
	require.NoError(t, err)
	removed, err = result.RowsAffected()
	require.NoError(t, err)
	require.Zero(t, removed)
	var deleted int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM user_platform_quotas WHERE id=7 AND deleted_at IS NOT NULL").Scan(&deleted))
	require.Equal(t, 1, deleted)
}
