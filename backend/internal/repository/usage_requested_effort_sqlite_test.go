package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRequestedEffortSQLiteUpgradeAndWriters(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "usage.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	require.NoError(t, applyMigrationsFS(ctx, db, migrationsWithout(t, "230_usage_log_requested_reasoning_effort.sql")))
	_, err = db.Exec("INSERT INTO usage_logs(user_id,api_key_id,account_id,request_id,model,reasoning_effort,actual_cost) VALUES(1,2,3,'legacy','m','high',0.5)")
	require.NoError(t, err)
	require.NoError(t, ApplyMigrations(ctx, db))
	repo := newUsageLogRepositoryWithSQL(nil, db)
	legacy, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	require.Nil(t, legacy.RequestedReasoningEffort)
	require.Equal(t, "high", *legacy.ReasoningEffort)
	require.Equal(t, 0.5, legacy.ActualCost)

	requested, sent := "xhigh", "high"
	for _, writer := range []string{"single", "batch", "best-effort"} {
		record := &service.UsageLog{UserID: 1, APIKeyID: 2, AccountID: 3, RequestID: writer,
			Model: "m", RequestedReasoningEffort: &requested, ReasoningEffort: &sent, ActualCost: 0.25, CreatedAt: time.Now().UTC()}
		switch writer {
		case "single":
			_, err = repo.createSingle(ctx, db, record)
		case "batch":
			_, err = repo.Create(ctx, record)
		default:
			err = repo.CreateBestEffort(ctx, record)
		}
		require.NoError(t, err)
		var id int64
		require.NoError(t, db.QueryRow("SELECT id FROM usage_logs WHERE request_id=?", writer).Scan(&id))
		stored, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, requested, *stored.RequestedReasoningEffort)
		require.Equal(t, sent, *stored.ReasoningEffort)
		require.Equal(t, 0.25, stored.ActualCost)
		inserted, err := repo.createSingle(ctx, db, record)
		require.NoError(t, err)
		require.False(t, inserted)
	}
}
