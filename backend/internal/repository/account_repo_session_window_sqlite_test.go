package repository

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestActiveFiveHourWindowBoundsAndCostsSQLite(t *testing.T) {
	ctx := context.Background()
	db := openUsageStatsSQLite(t)
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	repo := newAccountRepositoryWithSQL(client, db, nil)
	start := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	oldStart, oldEnd := start.Add(-5*time.Hour), start
	account, err := client.Account.Create().SetName("active window").SetPlatform(service.PlatformAnthropic).
		SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).
		SetSessionWindowStart(oldStart).SetSessionWindowEnd(oldEnd).SetSessionWindowStatus("allowed").
		SetExtra(map[string]any{"window_cost_limit": 10.0, "window_cost_sticky_reserve": 1.0}).Save(ctx)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO usage_logs (user_id, api_key_id, account_id, request_id, model, input_tokens, output_tokens, total_cost, actual_cost, created_at) VALUES
		(1, 1, ?, 'previous-window', 'claude-sonnet-4-6', 10, 20, 20, 20, ?),
		(1, 1, ?, 'current-window', 'claude-sonnet-4-6', 10, 20, 2, 2, ?)`, account.ID, start.Add(-time.Minute), account.ID, start.Add(time.Minute))
	require.NoError(t, err)
	end := start.Add(5 * time.Hour)
	require.NoError(t, repo.UpdateSessionWindowEnd(ctx, account.ID, end))
	current, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, current.SessionWindowStart.Equal(start))
	require.True(t, current.SessionWindowEnd.Equal(end))
	require.Equal(t, "allowed", current.SessionWindowStatus)
	stats, err := newUsageLogRepositoryWithSQL(client, db).GetAccountWindowStats(ctx, account.ID, current.GetCurrentWindowStartTime())
	require.NoError(t, err)
	require.EqualValues(t, 1, stats.Requests)
	require.Equal(t, 2.0, stats.StandardCost)
	require.Equal(t, service.WindowCostSchedulable, current.CheckWindowCostSchedulability(stats.StandardCost))
	var events int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM scheduler_outbox WHERE event_type=? AND account_id=?`, service.SchedulerOutboxEventAccountChanged, account.ID).Scan(&events))
	require.Equal(t, 1, events)
	// Modernc reads DATETIME values written with millisecond precision as time.Time.
	_, err = db.Exec(`UPDATE accounts SET session_window_start=?, session_window_end=? WHERE id=?`, start.Format("2006-01-02 15:04:05.000"), end.Format("2006-01-02 15:04:05.000"), account.ID)
	require.NoError(t, err)
	current, err = repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, current.SessionWindowStart.Equal(start))
	require.True(t, current.SessionWindowEnd.Equal(end))
}
