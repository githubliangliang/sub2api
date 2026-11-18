package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
)

var rateCancelFunctionID atomic.Uint64

func atomicRatesDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "rates.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE user_group_rate_multipliers (
		user_id INTEGER, group_id INTEGER, rate_multiplier REAL, rpm_override INTEGER,
		created_at DATETIME, updated_at DATETIME, PRIMARY KEY(user_id,group_id))`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO user_group_rate_multipliers VALUES
		(1,7,2,10,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
		(3,7,3,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
		(4,8,4,40,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	return db
}

func TestGroupRateSyncAtomicSQLite(t *testing.T) {
	for _, rpm := range []bool{false, true} {
		for _, failure := range []string{"insert", "delete", "cancel"} {
			t.Run(fmt.Sprintf("rpm=%t/%s", rpm, failure), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var function string
				if failure == "cancel" {
					function = fmt.Sprintf("rate_cancel_%d", rateCancelFunctionID.Add(1))
					require.NoError(t, sqlite.RegisterScalarFunction(function, 0, func(_ *sqlite.FunctionContext, _ []driver.Value) (driver.Value, error) {
						cancel()
						return int64(0), nil
					}))
				}
				db := atomicRatesDB(t)
				expectedRows := 3
				if rpm {
					_, err := db.Exec(`INSERT INTO user_group_rate_multipliers VALUES(5,7,NULL,42,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
					require.NoError(t, err)
					expectedRows++
				}
				trigger := `CREATE TRIGGER fail_sync BEFORE INSERT ON user_group_rate_multipliers WHEN NEW.user_id=2 BEGIN SELECT RAISE(ABORT,'insert failure'); END`
				switch failure {
				case "delete":
					trigger = `CREATE TRIGGER fail_sync BEFORE DELETE ON user_group_rate_multipliers BEGIN SELECT RAISE(ABORT,'delete failure'); END`
				case "cancel":
					trigger = `CREATE TRIGGER fail_sync AFTER UPDATE ON user_group_rate_multipliers BEGIN SELECT ` + function + `(); END`
				}
				_, err := db.Exec(trigger)
				require.NoError(t, err)
				repo := NewUserGroupRateRepository(db)
				if rpm {
					limit := 20
					err = repo.SyncGroupRPMOverrides(ctx, 7, []service.GroupRPMOverrideInput{{UserID: 2, RPMOverride: &limit}})
				} else {
					err = repo.SyncGroupRateMultipliers(ctx, 7, []service.GroupRateMultiplierInput{{UserID: 2, RateMultiplier: 4}})
				}
				require.Error(t, err)
				if failure == "cancel" {
					require.ErrorIs(t, ctx.Err(), context.Canceled)
				}
				var count int
				require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM user_group_rate_multipliers WHERE
					(user_id=1 AND group_id=7 AND rate_multiplier=2 AND rpm_override=10) OR
					(user_id=3 AND group_id=7 AND rate_multiplier=3 AND rpm_override IS NULL) OR
					(user_id=4 AND group_id=8 AND rate_multiplier=4 AND rpm_override=40) OR
					(user_id=5 AND group_id=7 AND rate_multiplier IS NULL AND rpm_override=42)`).Scan(&count))
				require.Equal(t, expectedRows, count, "rollback preserves both columns and other groups")
			})
		}
	}
}

func TestGroupRateSyncLastWinsSQLite(t *testing.T) {
	db := atomicRatesDB(t)
	repo := NewUserGroupRateRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.SyncGroupRateMultipliers(ctx, 7, []service.GroupRateMultiplierInput{
		{UserID: 1, RateMultiplier: 5}, {UserID: 2, RateMultiplier: 6}, {UserID: 1, RateMultiplier: 0},
	}))
	var rate float64
	var rpm int
	require.NoError(t, db.QueryRow(`SELECT rate_multiplier,rpm_override FROM user_group_rate_multipliers WHERE user_id=1 AND group_id=7`).Scan(&rate, &rpm))
	require.Zero(t, rate)
	require.Equal(t, 10, rpm)
	limit := 20
	require.NoError(t, repo.SyncGroupRPMOverrides(ctx, 7, []service.GroupRPMOverrideInput{
		{UserID: 1, RPMOverride: &limit}, {UserID: 2}, {UserID: 1}, {UserID: 2, RPMOverride: &limit},
	}))
	var nullableRPM sql.NullInt64
	require.NoError(t, db.QueryRow(`SELECT rate_multiplier,rpm_override FROM user_group_rate_multipliers WHERE user_id=1 AND group_id=7`).Scan(&rate, &nullableRPM))
	require.Zero(t, rate)
	require.False(t, nullableRPM.Valid, "the final nil wins over an earlier RPM value")
	require.NoError(t, db.QueryRow(`SELECT rpm_override FROM user_group_rate_multipliers WHERE user_id=2 AND group_id=7`).Scan(&rpm))
	require.Equal(t, limit, rpm)
	require.NoError(t, repo.SyncGroupRateMultipliers(ctx, 7, nil))
	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM user_group_rate_multipliers WHERE group_id=7`).Scan(&count))
	require.Equal(t, 1, count, "clearing rates preserves the RPM-only row")
	require.NoError(t, repo.SyncGroupRPMOverrides(ctx, 7, nil))
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM user_group_rate_multipliers`).Scan(&count))
	require.Equal(t, 1, count, "empty synchronization removes only rows with both columns NULL")
}

func TestGroupRateSyncBindSafeSQLite(t *testing.T) {
	db := atomicRatesDB(t)
	repo := NewUserGroupRateRepository(db)
	entries := make([]service.GroupRateMultiplierInput, 10000)
	rpms := make([]service.GroupRPMOverrideInput, len(entries))
	limit := 20
	for i := range entries {
		entries[i] = service.GroupRateMultiplierInput{UserID: int64(i + 1), RateMultiplier: float64(i)}
		rpms[i] = service.GroupRPMOverrideInput{UserID: int64(i + 1), RPMOverride: &limit}
	}
	require.NoError(t, repo.SyncGroupRateMultipliers(context.Background(), 7, entries))
	require.NoError(t, repo.SyncGroupRPMOverrides(context.Background(), 7, rpms))
	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM user_group_rate_multipliers WHERE group_id=7 AND rate_multiplier=user_id-1 AND rpm_override=20`).Scan(&count))
	require.Equal(t, 10000, count)
}
