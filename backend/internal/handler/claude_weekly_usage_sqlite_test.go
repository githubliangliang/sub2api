//go:build unit

package handler

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestClaudeWeeklyUsageSQLitePersistence(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "weekly.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, repository.ApplyMigrations(ctx, db))
	require.NoError(t, repository.EnsureSQLiteAuxTables(ctx, db))
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(1,'weekly','anthropic','setup-token','{}','{}')`)
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	accounts := repository.NewAccountRepository(client, db, nil)
	usageLogs := repository.NewUsageLogRepository(nil, db)
	rate := service.NewRateLimitService(accounts, usageLogs, &config.Config{}, nil, nil)
	usage := service.NewAccountUsageService(accounts, usageLogs, nil, nil, nil, nil, nil, nil, service.NewUsageCache(), nil, nil)
	reset := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	for _, exhausted := range []bool{false, true} {
		account, err := accounts.GetByID(ctx, 1)
		require.NoError(t, err)
		headers := http.Header{}
		utilization := "0.42"
		if exhausted {
			utilization = "1.0"
			headers.Set("anthropic-ratelimit-unified-7d-status", "rejected")
		}
		headers.Set("anthropic-ratelimit-unified-7d-utilization", utilization)
		headers.Set("anthropic-ratelimit-unified-7d-reset", strconv.FormatInt(reset.Unix(), 10))
		if exhausted {
			require.False(t, rate.HandleUpstreamError(ctx, account, 429, headers, nil))
		} else {
			rate.UpdateSessionWindow(ctx, account, headers)
		}
		// Read through the real repository and all public usage access patterns.
		single, err := usage.GetUsage(ctx, 1, true)
		require.NoError(t, err)
		passive, err := usage.GetPassiveUsage(ctx, 1)
		require.NoError(t, err)
		batch, failures, err := usage.GetUsageBatch(ctx, []int64{1}, true)
		require.NoError(t, err)
		require.Empty(t, failures)
		want := 42.0
		if exhausted {
			want = 100
		}
		for _, got := range []*service.UsageInfo{single, passive, batch[1]} {
			require.NotNil(t, got.SevenDay)
			require.Equal(t, want, got.SevenDay.Utilization)
			require.NotNil(t, got.SevenDay.ResetsAt)
			require.True(t, reset.Equal(*got.SevenDay.ResetsAt))
		}
		persisted, err := accounts.GetByID(ctx, 1)
		require.NoError(t, err)
		require.Nil(t, persisted.SessionWindowEnd, "weekly sampling must not invent a five-hour window")
		if exhausted {
			require.NotNil(t, persisted.RateLimitResetAt)
			require.True(t, reset.Equal(*persisted.RateLimitResetAt))
		}
	}
}
