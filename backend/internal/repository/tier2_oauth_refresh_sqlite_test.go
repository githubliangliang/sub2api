package repository

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTier2PausedOAuthRefreshCandidatesSQLite(t *testing.T) {
	ctx := context.Background()
	db := newPGCompatSQLiteDB(t)
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	repo := newAccountRepositoryWithSQL(client, db, nil)
	_, err := db.Exec(`INSERT INTO accounts (id,name,platform,type,status,schedulable,credentials) VALUES
		(1,'healthy','openai','oauth','active',1,'{"refresh_token":"fixture"}'),
		(2,'paused','openai','oauth','active',0,'{"refresh_token":"fixture"}'),
		(3,'rejected','openai','oauth','error',0,'{"refresh_token":"fixture"}'),
		(4,'retry','openai','oauth','active',1,'{"refresh_token":"fixture"}')`)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE accounts SET temp_unschedulable_until=datetime('now','+1 hour'),temp_unschedulable_reason='token refresh retry exhausted: fixture' WHERE id=4`)
	require.NoError(t, err)
	page, err := repo.ListOAuthRefreshCandidatePage(ctx, service.OAuthRefreshPageOptions{Platforms: []string{service.PlatformOpenAI}, Limit: 20, ActiveOnly: true, RequireRefreshToken: true, ExcludeRetryCooldown: true})
	require.NoError(t, err)
	ids := make([]int64, 0, len(page.Accounts))
	for _, account := range page.Accounts {
		ids = append(ids, account.ID)
	}
	require.Equal(t, []int64{1, 2}, ids)
}
