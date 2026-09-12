package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestInsertSystemMetricsNullableIntegerMetrics(t *testing.T) {
	createdAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	zero := 0

	tests := []struct {
		name         string
		dbConnActive *int
		wantDBActive driver.Value
	}{
		{name: "explicit zero is preserved", dbConnActive: &zero, wantDBActive: int64(0)},
		{name: "unavailable metric remains null", dbConnActive: nil, wantDBActive: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			t.Cleanup(func() { _ = db.Close() })

			args := make([]driver.Value, 40)
			args[0] = createdAt
			args[1] = int64(1)
			for i := 4; i <= 12; i++ {
				args[i] = int64(0)
			}
			args[35] = tt.wantDBActive

			mock.ExpectExec("INSERT INTO ops_system_metrics").
				WithArgs(args...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			repo := &opsRepository{db: db}
			err = repo.InsertSystemMetrics(context.Background(), &service.OpsInsertSystemMetricsInput{
				CreatedAt:    createdAt,
				DBConnActive: tt.dbConnActive,
			})
			if err != nil {
				t.Fatalf("InsertSystemMetrics: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet SQL expectations: %v", err)
			}
		})
	}
}

func TestInsertSystemMetrics_SQLiteDistinguishesZeroFromMissing(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	for _, migration := range []string{"033_ops_monitoring_vnext.sql", "042b_add_ops_system_metrics_switch_count.sql"} {
		ddl, err := os.ReadFile("../../migrations/" + migration)
		require.NoError(t, err)
		for _, stmt := range splitSQLStatements(string(ddl)) {
			if strings.Contains(stmt, "CREATE TABLE IF NOT EXISTS ops_system_metrics (") || strings.Contains(stmt, "ALTER TABLE ops_system_metrics ADD COLUMN") {
				_, err := db.Exec(stmt)
				require.NoError(t, err)
			}
		}
	}

	repo := &opsRepository{db: db}
	for _, tc := range []struct {
		name    string
		present bool
		value   int
	}{
		{"zero", true, 0}, {"nonzero", true, 7}, {"missing", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := &service.OpsInsertSystemMetricsInput{}
			if tc.present {
				v, v64 := tc.value, int64(tc.value)
				input.DurationP50Ms, input.DurationP90Ms, input.DurationP95Ms, input.DurationP99Ms, input.DurationMaxMs = &v, &v, &v, &v, &v
				input.TTFTP50Ms, input.TTFTP90Ms, input.TTFTP95Ms, input.TTFTP99Ms, input.TTFTMaxMs = &v, &v, &v, &v, &v
				input.MemoryUsedMB, input.MemoryTotalMB = &v64, &v64
				input.RedisConnTotal, input.RedisConnIdle = &v, &v
				input.DBConnActive, input.DBConnIdle, input.DBConnWaiting = &v, &v, &v
				input.GoroutineCount, input.ConcurrencyQueueDepth = &v, &v
			}
			zeroGroup := int64(0)
			input.GroupID = &zeroGroup
			require.NoError(t, repo.InsertSystemMetrics(context.Background(), input))
			columns := []string{
				"duration_p50_ms", "duration_p90_ms", "duration_p95_ms", "duration_p99_ms", "duration_max_ms",
				"ttft_p50_ms", "ttft_p90_ms", "ttft_p95_ms", "ttft_p99_ms", "ttft_max_ms",
				"memory_used_mb", "memory_total_mb", "redis_conn_total", "redis_conn_idle",
				"db_conn_active", "db_conn_idle", "db_conn_waiting", "goroutine_count", "concurrency_queue_depth",
			}
			for _, column := range columns {
				var got sql.NullInt64
				require.NoError(t, db.QueryRow("SELECT "+column+" FROM ops_system_metrics ORDER BY id DESC LIMIT 1").Scan(&got))
				require.Equal(t, sql.NullInt64{Int64: int64(tc.value), Valid: tc.present}, got, column)
			}
			var group sql.NullInt64
			require.NoError(t, db.QueryRow("SELECT group_id FROM ops_system_metrics ORDER BY id DESC LIMIT 1").Scan(&group))
			require.False(t, group.Valid, "business ID zero semantics remain unchanged")
		})
	}
}
