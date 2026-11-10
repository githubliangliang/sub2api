package repository

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestFastTiersSQLiteUpgradePersistenceAndAuthProjection(t *testing.T) {
	ctx := context.Background()
	registerSQLitePGCompatFunctions()
	db, err := sql.Open("sqlite", "file:"+t.TempDir()+"/fast.db?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()
	db.SetMaxOpenConns(1)
	baseline := fstest.MapFS{}
	entries, err := migrations.FS.ReadDir(".")
	require.NoError(t, err)
	for _, e := range entries {
		if e.Name() >= "228_" {
			continue
		}
		data, err := migrations.FS.ReadFile(e.Name())
		require.NoError(t, err)
		baseline[e.Name()] = &fstest.MapFile{Data: data, Mode: 0444}
	}
	require.NoError(t, applyMigrationsFS(ctx, db, baseline))
	_, err = db.Exec(`INSERT INTO groups(id,name,platform) VALUES(901,'existing-fast','openai'); INSERT INTO users(id,email,password_hash) VALUES(901,'fast@example.invalid','fixture'); INSERT INTO api_keys(id,user_id,key,name,group_id,status) VALUES(901,901,'fast-sqlite-key','fast',901,'active')`)
	require.NoError(t, err)
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	groupRepo := NewGroupRepository(client, db)
	group, err := groupRepo.GetByID(ctx, 901)
	require.NoError(t, err)
	require.False(t, group.ForceOpenAIFast)
	require.False(t, group.FreeOpenAIFast)
	group.ForceOpenAIFast = true
	group.FreeOpenAIFast = true
	require.NoError(t, groupRepo.Update(ctx, group))
	keyRepo := NewAPIKeyRepository(client, db)
	key, err := keyRepo.GetByKeyForAuth(ctx, "fast-sqlite-key")
	require.NoError(t, err)
	require.True(t, key.Group.ForceOpenAIFast)
	require.True(t, key.Group.FreeOpenAIFast)
	clone := *group
	clone.ID = 0
	clone.Name = "new-fast"
	require.NoError(t, groupRepo.Create(ctx, &clone))
	loaded, err := groupRepo.GetByID(ctx, clone.ID)
	require.NoError(t, err)
	require.True(t, loaded.ForceOpenAIFast)
	require.True(t, loaded.FreeOpenAIFast)
	_, err = db.Exec(`INSERT INTO channels(id,name,status) VALUES(901,'fast-channel','active')`)
	require.NoError(t, err)
	channelRepo := NewChannelRepository(db)
	multiplier := 3.0
	price := 0.001
	pricing := service.ChannelModelPricing{ChannelID: 901, Platform: service.PlatformOpenAI, Models: []string{"gpt-6-astra"}, BillingMode: service.BillingModeToken, InputPrice: &price, FastMultiplier: &multiplier}
	require.NoError(t, channelRepo.CreateModelPricing(ctx, &pricing))
	prices, err := channelRepo.ListModelPricing(ctx, 901)
	require.NoError(t, err)
	require.Len(t, prices, 1)
	require.Equal(t, 3.0, *prices[0].FastMultiplier)
	pricing.FastMultiplier = nil
	require.NoError(t, channelRepo.UpdateModelPricing(ctx, &pricing))
	prices, err = channelRepo.ListModelPricing(ctx, 901)
	require.NoError(t, err)
	require.Nil(t, prices[0].FastMultiplier)
	group.ForceOpenAIFast = false
	group.FreeOpenAIFast = false
	require.NoError(t, groupRepo.Update(ctx, group))
	key, err = keyRepo.GetByKeyForAuth(ctx, "fast-sqlite-key")
	require.NoError(t, err)
	require.False(t, key.Group.ForceOpenAIFast)
	require.False(t, key.Group.FreeOpenAIFast)
}
