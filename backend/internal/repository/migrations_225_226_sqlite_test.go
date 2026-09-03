package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigrations225And226ApplyOnFreshSQLite(t *testing.T) {
	applyAndAssert := func(t *testing.T, dsnName string) {
		t.Helper()
		dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=foreign_keys(1)&_time_format=sqlite", dsnName)
		db, err := sql.Open("sqlite", dsn)
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		db.SetMaxOpenConns(1)

		require.NoError(t, ApplyMigrations(context.Background(), db))

		for _, filename := range []string{
			"225_group_reasoning_effort_over_limit.sql",
			"226_channel_cache_write_1h_pricing.sql",
		} {
			var n int
			err := db.QueryRowContext(context.Background(),
				`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, filename,
			).Scan(&n)
			require.NoError(t, err)
			require.Equalf(t, 1, n, "expected schema_migrations row for %s", filename)
		}

		var overLimit int
		err = db.QueryRowContext(context.Background(),
			`SELECT COUNT(*) FROM pragma_table_info('groups') WHERE name = $1`,
			"max_reasoning_effort_over_limit",
		).Scan(&overLimit)
		require.NoError(t, err)
		require.Equal(t, 1, overLimit, "groups.max_reasoning_effort_over_limit")

		for _, table := range []string{
			"channel_model_pricing",
			"channel_pricing_intervals",
			"channel_account_stats_model_pricing",
			"channel_account_stats_pricing_intervals",
		} {
			var n int
			err := db.QueryRowContext(context.Background(),
				fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = $1`, table),
				"cache_write_1h_price",
			).Scan(&n)
			require.NoError(t, err)
			require.Equalf(t, 1, n, "expected %s.cache_write_1h_price", table)
		}
	}

	t.Run("db1", func(t *testing.T) { applyAndAssert(t, t.Name()) })
	t.Run("db2", func(t *testing.T) { applyAndAssert(t, t.Name()) })
}
