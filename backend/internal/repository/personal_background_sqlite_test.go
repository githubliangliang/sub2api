package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func preparePersonalBackgroundSQLite(t *testing.T, legacyText bool) *sql.DB {
	t.Helper()
	db := openPersonalMaintenanceSQLite(t)
	_, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, quota_used REAL NOT NULL DEFAULT 7);
		CREATE TABLE accounts (id INTEGER PRIMARY KEY);
		INSERT INTO users (id) VALUES (1), (2);
		INSERT INTO accounts (id) VALUES (1);
		CREATE TABLE usage_logs (id INTEGER PRIMARY KEY, created_at DATETIME, user_id INTEGER,
			api_key_id INTEGER, account_id INTEGER, group_id INTEGER, model TEXT,
			request_type INTEGER, stream BOOLEAN, billing_type INTEGER);
		CREATE TABLE usage_billing_dedup (request_id TEXT PRIMARY KEY);
		INSERT INTO usage_billing_dedup VALUES ('settled-request')`)
	require.NoError(t, err)
	for _, name := range []string{"042_add_usage_cleanup_tasks.sql", "043_add_usage_cleanup_cancel_audit.sql", "066_add_scheduled_test_tables.sql", "070_add_scheduled_test_auto_recover.sql"} {
		ddl, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		require.NoError(t, err)
		text := string(ddl)
		if legacyText {
			text = strings.ReplaceAll(text, "DATETIME", "TEXT")
		}
		_, err = db.Exec(text)
		require.NoError(t, err)
	}
	return db
}

func TestUsageCleanupRepositorySQLiteLifecycle(t *testing.T) {
	for _, legacyText := range []bool{false, true} {
		name := "datetime"
		if legacyText {
			name = "legacy_text"
		}
		t.Run(name, func(t *testing.T) {
			db := preparePersonalBackgroundSQLite(t, legacyText)
			ctx := context.Background()
			repo := NewUsageCleanupRepository(nil, db)
			start := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			userID := int64(1)
			filters := service.UsageCleanupFilters{StartTime: start, EndTime: start.Add(time.Hour), UserID: &userID}
			for i, row := range []struct {
				at   time.Time
				user int64
			}{
				{start.Add(time.Minute), 1}, {start.Add(2 * time.Minute), 1},
				{start.Add(time.Minute), 2}, {start.Add(-time.Minute), 1},
			} {
				_, err := db.Exec(`INSERT INTO usage_logs (id, created_at, user_id) VALUES (?, ?, ?)`, i+1, row.at, row.user)
				require.NoError(t, err)
			}
			task := &service.UsageCleanupTask{Status: service.UsageCleanupStatusPending, Filters: filters, CreatedBy: 1}
			require.NoError(t, repo.CreateTask(ctx, task))
			claimed, err := repo.ClaimNextPendingTask(ctx, 1800)
			require.NoError(t, err)
			require.Equal(t, task.ID, claimed.ID)
			require.NotNil(t, claimed.StartedAt)
			require.False(t, claimed.CreatedAt.IsZero())
			active, err := repo.ClaimNextPendingTask(ctx, 1800)
			require.NoError(t, err)
			require.Nil(t, active)
			var total int64
			for range 3 {
				deleted, err := repo.DeleteUsageLogsBatch(ctx, filters, 1)
				require.NoError(t, err)
				total += deleted
				require.NoError(t, repo.UpdateTaskProgress(ctx, task.ID, total))
			}
			require.EqualValues(t, 2, total)
			require.NoError(t, repo.MarkTaskSucceeded(ctx, task.ID, total))
			tasks, _, err := repo.ListTasks(ctx, pagination.PaginationParams{Page: 1, PageSize: 20})
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			require.Equal(t, service.UsageCleanupStatusSucceeded, tasks[0].Status)
			require.NotNil(t, tasks[0].FinishedAt)
			var remaining, dedup int
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_logs`).Scan(&remaining))
			require.Equal(t, 2, remaining)
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
			require.Equal(t, 1, dedup)
			var quota float64
			require.NoError(t, db.QueryRow(`SELECT quota_used FROM users WHERE id=1`).Scan(&quota))
			require.Equal(t, 7.0, quota)
		})
	}
}

func TestUsageCleanupRepositorySQLiteStaleAndCanceled(t *testing.T) {
	db := preparePersonalBackgroundSQLite(t, false)
	ctx := context.Background()
	repo := NewUsageCleanupRepository(nil, db)
	filters := service.UsageCleanupFilters{StartTime: time.Now().Add(-time.Hour), EndTime: time.Now()}
	task := &service.UsageCleanupTask{Status: service.UsageCleanupStatusRunning, Filters: filters, CreatedBy: 1, DeletedRows: 5}
	require.NoError(t, repo.CreateTask(ctx, task))
	_, err := db.Exec(`UPDATE usage_cleanup_tasks SET started_at=? WHERE id=?`, time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano), task.ID)
	require.NoError(t, err)
	claimed, err := repo.ClaimNextPendingTask(ctx, 1800)
	require.NoError(t, err)
	require.NotNil(t, claimed, "stale RFC3339 timestamps must be compared chronologically")
	require.EqualValues(t, 5, claimed.DeletedRows)
	canceled, err := repo.CancelTask(ctx, task.ID, 1)
	require.NoError(t, err)
	require.True(t, canceled)
	require.NoError(t, repo.MarkTaskSucceeded(ctx, task.ID, 6))
	require.NoError(t, repo.MarkTaskFailed(ctx, task.ID, 6, "late failure"))
	status, err := repo.GetTaskStatus(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, service.UsageCleanupStatusCanceled, status, "late worker completion must preserve cancellation")
	next, err := repo.ClaimNextPendingTask(ctx, 1800)
	require.NoError(t, err)
	require.Nil(t, next)
}

func TestUsageCleanupWorkerSQLiteResumesPending(t *testing.T) {
	db := preparePersonalBackgroundSQLite(t, false)
	repo := NewUsageCleanupRepository(nil, db)
	ctx := context.Background()
	now := time.Now().UTC()
	_, err := db.Exec(`INSERT INTO usage_logs (id, created_at, user_id) VALUES (1, ?, 1)`, now.Add(-time.Minute))
	require.NoError(t, err)
	task := &service.UsageCleanupTask{Status: service.UsageCleanupStatusPending,
		Filters: service.UsageCleanupFilters{StartTime: now.Add(-time.Hour), EndTime: now}, CreatedBy: 1}
	require.NoError(t, repo.CreateTask(ctx, task))
	wheel, err := service.NewTimingWheelService()
	require.NoError(t, err)
	t.Cleanup(wheel.Stop)
	cfg := &config.Config{}
	cfg.Database.Driver = "sqlite"
	cfg.UsageCleanup.Enabled = true
	cfg.UsageCleanup.WorkerIntervalSeconds = 1
	cfg.UsageCleanup.BatchSize = 1
	cfg.UsageCleanup.TaskTimeoutSeconds = 10
	worker := service.ProvideUsageCleanupService(repo, wheel, nil, cfg)
	t.Cleanup(worker.Stop)
	require.Eventually(t, func() bool {
		status, err := repo.GetTaskStatus(ctx, task.ID)
		return err == nil && status == service.UsageCleanupStatusSucceeded
	}, 4*time.Second, 20*time.Millisecond)
	worker.Stop()
	second := &service.UsageCleanupTask{Status: service.UsageCleanupStatusPending, Filters: task.Filters, CreatedBy: 1}
	require.NoError(t, repo.CreateTask(ctx, second))
	time.Sleep(1100 * time.Millisecond)
	status, err := repo.GetTaskStatus(ctx, second.ID)
	require.NoError(t, err)
	require.Equal(t, service.UsageCleanupStatusPending, status)
}

func TestUsageCleanupRepositorySQLiteCanceledTerminalState(t *testing.T) {
	db := preparePersonalBackgroundSQLite(t, false)
	repo := NewUsageCleanupRepository(nil, db)
	ctx := context.Background()
	task := &service.UsageCleanupTask{Status: service.UsageCleanupStatusRunning, CreatedBy: 1,
		Filters: service.UsageCleanupFilters{StartTime: time.Now().Add(-time.Hour), EndTime: time.Now()}}
	require.NoError(t, repo.CreateTask(ctx, task))
	canceled, err := repo.CancelTask(ctx, task.ID, 1)
	require.NoError(t, err)
	require.True(t, canceled)
	require.NoError(t, repo.MarkTaskSucceeded(ctx, task.ID, 1))
	require.NoError(t, repo.MarkTaskFailed(ctx, task.ID, 1, "late worker error"))
	status, err := repo.GetTaskStatus(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, service.UsageCleanupStatusCanceled, status)
}

func TestScheduledTestRepositorySQLiteLifecycle(t *testing.T) {
	for _, legacyText := range []bool{false, true} {
		name := "datetime"
		if legacyText {
			name = "legacy_text"
		}
		t.Run(name, func(t *testing.T) {
			db := preparePersonalBackgroundSQLite(t, legacyText)
			ctx := context.Background()
			plans := NewScheduledTestPlanRepository(db)
			results := NewScheduledTestResultRepository(db)
			now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
			past, future := now.Add(-time.Minute), now.Add(time.Minute)
			plan, err := plans.Create(ctx, &service.ScheduledTestPlan{AccountID: 1, ModelID: "test-model", CronExpression: "*/30 * * * *", Enabled: true, MaxResults: 2, AutoRecover: true, NextRunAt: &past})
			require.NoError(t, err)
			require.NotNil(t, plan.NextRunAt)
			require.True(t, past.Equal(*plan.NextRunAt))
			_, err = db.Exec(`UPDATE scheduled_test_plans SET next_run_at=? WHERE id=?`, past.In(time.FixedZone("offset", 8*3600)).Format(time.RFC3339), plan.ID)
			require.NoError(t, err)
			_, err = plans.Create(ctx, &service.ScheduledTestPlan{AccountID: 1, Enabled: true, NextRunAt: &future})
			require.NoError(t, err)
			_, err = plans.Create(ctx, &service.ScheduledTestPlan{AccountID: 1, Enabled: false, NextRunAt: &past})
			require.NoError(t, err)
			_, err = plans.Create(ctx, &service.ScheduledTestPlan{AccountID: 1, Enabled: true})
			require.NoError(t, err)
			due, err := plans.ListDue(ctx, now)
			require.NoError(t, err)
			require.Len(t, due, 1)
			require.Equal(t, plan.ID, due[0].ID)
			plan.ModelID = "updated-model"
			plan.NextRunAt = &future
			plan, err = plans.Update(ctx, plan)
			require.NoError(t, err)
			require.Equal(t, "updated-model", plan.ModelID)
			require.NoError(t, plans.UpdateAfterRun(ctx, plan.ID, now, future))
			loaded, err := plans.GetByID(ctx, plan.ID)
			require.NoError(t, err)
			require.True(t, now.Equal(*loaded.LastRunAt))
			listed, err := plans.ListByAccountID(ctx, 1)
			require.NoError(t, err)
			require.Len(t, listed, 4)
			var ids []int64
			for range 4 {
				result, err := results.Create(ctx, &service.ScheduledTestResult{PlanID: plan.ID, Status: "success", ResponseText: "ok", StartedAt: past, FinishedAt: now})
				require.NoError(t, err)
				require.False(t, result.CreatedAt.IsZero())
				ids = append(ids, result.ID)
			}
			_, err = db.Exec(`UPDATE scheduled_test_results SET created_at=?`, now)
			require.NoError(t, err)
			require.NoError(t, results.PruneOldResults(ctx, plan.ID, 2))
			kept, err := results.ListByPlanID(ctx, plan.ID, 10)
			require.NoError(t, err)
			require.Len(t, kept, 2)
			require.Equal(t, []int64{ids[3], ids[2]}, []int64{kept[0].ID, kept[1].ID})
			require.NoError(t, plans.Delete(ctx, plan.ID))
			kept, err = results.ListByPlanID(ctx, plan.ID, 10)
			require.NoError(t, err)
			require.Empty(t, kept)
		})
	}
}

func TestAccountExpirySQLiteFormatsAndEligibility(t *testing.T) {
	registerSQLitePGCompatFunctions()
	repo, client := newOllamaCloudUsageSQLiteRepository(t)
	ctx := context.Background()
	_, err := repo.sql.ExecContext(ctx, `CREATE TABLE scheduler_outbox (id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL, account_id INTEGER, group_id INTEGER, payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, dedup_key TEXT UNIQUE)`)
	require.NoError(t, err)
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	var eligibleIDs []int64
	for _, tc := range []struct {
		name, expires                               string
		automatic, schedulable, deleted, wantPaused bool
	}{
		{"sqlite", "2026-10-28 11:59:59", true, true, false, true},
		{"rfc3339", "2026-10-28T11:59:59Z", true, true, false, true},
		{"offset", "2026-10-28T19:59:59+08:00", true, true, false, true},
		{"boundary", "2026-10-28T12:00:00Z", true, true, false, true},
		{"future", "2026-10-28T12:00:01Z", true, true, false, false},
		{"manual", "2026-10-28 11:59:59", false, true, false, false},
		{"paused", "2026-10-28 11:59:59", true, false, false, false},
		{"deleted", "2026-10-28 11:59:59", true, true, true, false},
	} {
		account, err := client.Account.Create().SetName(tc.name).SetPlatform(service.PlatformOpenAI).
			SetType(service.AccountTypeAPIKey).SetCredentials(map[string]any{"api_key": "test"}).
			SetSchedulable(tc.schedulable).SetAutoPauseOnExpired(tc.automatic).Save(ctx)
		require.NoError(t, err)
		var deletedAt any
		if tc.deleted {
			deletedAt = now
		}
		_, err = repo.sql.ExecContext(ctx, `UPDATE accounts SET expires_at=?, deleted_at=? WHERE id=?`, tc.expires, deletedAt, account.ID)
		require.NoError(t, err)
		if tc.wantPaused {
			eligibleIDs = append(eligibleIDs, account.ID)
		}
	}
	paused, err := repo.AutoPauseExpiredAccounts(ctx, now)
	require.NoError(t, err)
	require.EqualValues(t, len(eligibleIDs), paused)
	for _, id := range eligibleIDs {
		account, err := client.Account.Get(ctx, id)
		require.NoError(t, err)
		require.False(t, account.Schedulable)
	}
	paused, err = repo.AutoPauseExpiredAccounts(ctx, now)
	require.NoError(t, err)
	require.Zero(t, paused)
	var events int
	rows, err := repo.sql.QueryContext(ctx, `SELECT COUNT(*) FROM scheduler_outbox`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&events))
	require.Equal(t, 1, events)
}

func TestAccountExpirySQLiteSnapshotRefresh(t *testing.T) {
	registerSQLitePGCompatFunctions()
	repo, client := newOllamaCloudUsageSQLiteRepository(t)
	ctx := context.Background()
	_, err := repo.sql.ExecContext(ctx, `CREATE TABLE scheduler_outbox (id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL, account_id INTEGER, group_id INTEGER, payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP, dedup_key TEXT UNIQUE)`)
	require.NoError(t, err)
	account, err := client.Account.Create().SetName("expired").SetPlatform(service.PlatformAnthropic).
		SetType(service.AccountTypeAPIKey).SetCredentials(map[string]any{"api_key": "test"}).
		SetSchedulable(true).SetAutoPauseOnExpired(true).Save(ctx)
	require.NoError(t, err)
	now := time.Now().UTC()
	_, err = repo.sql.ExecContext(ctx, `UPDATE accounts SET expires_at=? WHERE id=?`, now.Add(-time.Hour).Format("2006-01-02 15:04:05"), account.ID)
	require.NoError(t, err)
	cache := &quotaSnapshotCacheStub{}
	repo.schedulerCache = cache
	paused, err := repo.AutoPauseExpiredAccounts(ctx, now)
	require.NoError(t, err)
	require.EqualValues(t, 1, paused)
	require.Len(t, cache.setAccounts, 1, "expiry must refresh the current account snapshot immediately")
	require.False(t, cache.setAccounts[0].Schedulable)
}
