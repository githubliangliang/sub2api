package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func openPersonalMaintenanceSQLite(t *testing.T) *sql.DB {
	t.Helper()
	registerSQLitePGCompatFunctions()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "maintenance.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

func TestAuditLogClearAllSQLite(t *testing.T) {
	db := openPersonalMaintenanceSQLite(t)
	ctx := context.Background()
	// Exercise the production auxiliary-table DDL, including its legacy TEXT timestamp.
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	repo := NewAuditLogRepository(db)
	svc := service.NewAuditLogService(repo, nil)
	t.Cleanup(svc.Stop)

	inserted, err := repo.BatchInsert(ctx, []*service.AuditLog{
		{Action: "admin.account.create", StatusCode: 200},
		{Action: "admin.account.update", StatusCode: 400},
	})
	require.NoError(t, err)
	require.EqualValues(t, 2, inserted)
	actorID := int64(7)
	trace := &service.AuditLog{
		ActorUserID: &actorID, ActorEmail: "admin@example.test", ActorRole: "admin",
		AuthMethod: "jwt", Method: "POST", Path: "/api/v1/admin/audit-logs/clear",
		ClientIP: "127.0.0.1", UserAgent: "test-browser", StatusCode: 200,
	}

	deleted, err := svc.ClearAll(ctx, trace)
	require.NoError(t, err)
	require.EqualValues(t, 2, deleted)
	result, err := repo.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Len(t, result.Logs, 1)
	entry := result.Logs[0]
	require.Equal(t, service.AuditActionAuditLogClear, entry.Action)
	require.EqualValues(t, deleted, entry.Extra["deleted_rows"])
	require.Equal(t, actorID, *entry.ActorUserID)
	require.Equal(t, trace.ClientIP, entry.ClientIP)
	require.Equal(t, trace.UserAgent, entry.UserAgent)
	require.Greater(t, entry.ID, inserted, "clearing must not reuse historical audit IDs")
	require.False(t, entry.CreatedAt.IsZero())

	deleted, err = svc.ClearAll(ctx, &service.AuditLog{StatusCode: 200})
	require.NoError(t, err)
	require.EqualValues(t, 1, deleted)
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
}

func TestAuditLogClearAllSQLiteCanceled(t *testing.T) {
	db := openPersonalMaintenanceSQLite(t)
	ctx := context.Background()
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	repo := NewAuditLogRepository(db)
	require.NoError(t, repo.Insert(ctx, &service.AuditLog{Action: "keep"}))
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	require.ErrorIs(t, repo.TruncateAll(canceled), context.Canceled)
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
}
