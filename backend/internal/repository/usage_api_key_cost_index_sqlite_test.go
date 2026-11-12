package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyCostCoveringIndexSQLiteUpgradeAndCost(t *testing.T) {
	ctx := context.Background()
	baseline := fstest.MapFS{}
	entries, err := migrations.FS.ReadDir(".")
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.Name() >= "229_" {
			continue
		}
		data, err := migrations.FS.ReadFile(entry.Name())
		require.NoError(t, err)
		baseline[entry.Name()] = &fstest.MapFile{Data: data, Mode: 0444}
	}
	open := func(name string) *sql.DB {
		db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), name)+"?_time_format=sqlite")
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		db.SetMaxOpenConns(1)
		return db
	}
	db := open("baseline.db")
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	start := time.Now().UTC().Truncate(24 * time.Hour).Add(-48 * time.Hour)
	end := start.Add(30 * 24 * time.Hour)
	insert := func(db *sql.DB, offset, count int) time.Duration {
		tx, err := db.Begin()
		require.NoError(t, err)
		stmt, err := tx.Prepare(`INSERT INTO usage_logs(user_id,api_key_id,account_id,request_id,model,actual_cost,created_at) VALUES(1,?,1,?,'index-fixture',?,?)`)
		require.NoError(t, err)
		started := time.Now()
		for n := offset; n < offset+count; n++ {
			_, err = stmt.Exec(n%10+1, fmt.Sprint(n), float64(n%7)/1000, start.Add(time.Duration(n)*13*time.Second))
			require.NoError(t, err)
		}
		require.NoError(t, stmt.Close())
		require.NoError(t, tx.Commit())
		return time.Since(started)
	}
	insert(db, 0, 20000)
	repo := newUsageLogRepositoryWithSQL(nil, db)
	before, err := repo.GetBatchAPIKeyUsageStats(ctx, []int64{1, 2, 3, 99}, start, end)
	require.NoError(t, err)
	plan := func() string {
		rows, err := db.Query(`EXPLAIN QUERY PLAN SELECT api_key_id, SUM(actual_cost) FILTER(WHERE created_at>=? AND created_at<?), SUM(actual_cost) FILTER(WHERE created_at>=?) FROM usage_logs WHERE api_key_id IN(1,2,3) AND created_at>=? GROUP BY api_key_id`, start, end, start.Add(48*time.Hour), start)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var details []string
		for rows.Next() {
			var id, parent, unused int
			var detail string
			require.NoError(t, rows.Scan(&id, &parent, &unused, &detail))
			details = append(details, detail)
		}
		require.NoError(t, rows.Err())
		return strings.Join(details, "; ")
	}
	measureQuery := func() time.Duration {
		times := make([]time.Duration, 7)
		for i := range times {
			started := time.Now()
			_, err := repo.GetBatchAPIKeyUsageStats(ctx, []int64{1, 2, 3, 99}, start, end)
			require.NoError(t, err)
			times[i] = time.Since(started)
		}
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		return times[len(times)/2]
	}
	bytes := func() int64 {
		var pages, size int64
		require.NoError(t, db.QueryRow("PRAGMA page_count").Scan(&pages))
		require.NoError(t, db.QueryRow("PRAGMA page_size").Scan(&size))
		return pages * size
	}
	baselinePlan, baselineQuery, baselineBytes := plan(), measureQuery(), bytes()
	// Clone identical data so the write workload starts from the same rows.
	clonePath := filepath.Join(t.TempDir(), "write-control.db")
	_, err = db.Exec("VACUUM INTO ?", clonePath)
	require.NoError(t, err)
	control, err := sql.Open("sqlite", "file:"+clonePath+"?_time_format=sqlite")
	require.NoError(t, err)
	defer func() { _ = control.Close() }()
	started := time.Now()
	require.NoError(t, ApplyMigrations(ctx, db))
	buildTime := time.Since(started)
	require.NoError(t, ApplyMigrations(ctx, db), "repeated startup preserves migration checksums")
	indexPlan, indexedQuery, indexedBytes := plan(), measureQuery(), bytes()
	require.Contains(t, indexPlan, "COVERING INDEX idx_usage_logs_api_key_created_cost")
	after, err := repo.GetBatchAPIKeyUsageStats(ctx, []int64{1, 2, 3, 99}, start, end)
	require.NoError(t, err)
	require.Equal(t, before, after)
	baselineWrites := insert(control, 20000, 5000)
	indexedWrites := insert(db, 20000, 5000)
	t.Logf("modernc SQLite; synthetic 20000 rows, same 5000-row write workload; before_plan=%s after_plan=%s", baselinePlan, indexPlan)
	t.Logf("index_build=%s query_median_7_before=%s after=%s database_bytes_before=%d after=%d increase=%d writes_before=%s after=%s", buildTime, baselineQuery, indexedQuery, baselineBytes, indexedBytes, indexedBytes-baselineBytes, baselineWrites, indexedWrites)
	if peak, err := os.ReadFile("/sys/fs/cgroup/memory.peak"); err == nil {
		t.Logf("cgroup memory.peak=%s", strings.TrimSpace(string(peak)))
	}
}
