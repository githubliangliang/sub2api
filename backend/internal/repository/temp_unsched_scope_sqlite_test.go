package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTempUnschedScopeSQLiteWithEmbeddedRedis(t *testing.T) {
	for _, scope := range []string{service.TempUnschedScopeModel, service.TempUnschedScopeAccount} {
		t.Run(scope, func(t *testing.T) {
			ctx := context.Background()
			repo, client := newOllamaCloudUsageSQLiteRepository(t)
			_, err := repo.sql.ExecContext(ctx, `CREATE TABLE scheduler_outbox (
				id INTEGER PRIMARY KEY AUTOINCREMENT, event_type TEXT NOT NULL, account_id INTEGER,
				group_id INTEGER, payload TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, dedup_key TEXT UNIQUE)`)
			require.NoError(t, err)
			disabled := false
			cfg := &config.Config{RunMode: config.RunModeSimple, Redis: config.RedisConfig{Enabled: &disabled}}
			rdb := InitRedis(cfg)
			t.Cleanup(func() { _ = rdb.Close() })
			require.NoError(t, rdb.Ping(ctx).Err())
			cache := NewSchedulerCache(rdb)
			tempCache := NewTempUnschedCache(rdb)
			repo.schedulerCache = cache
			credentials := map[string]any{
				"api_key": "test", "model_mapping": map[string]any{"alias-a": "model-a"},
				"temp_unschedulable_enabled": true,
				"temp_unschedulable_rules": []any{map[string]any{
					"error_code": 503, "keywords": []any{"capacity"}, "duration_minutes": 1, "scope": scope,
				}},
			}
			created, err := client.Account.Create().SetName("scoped").SetPlatform(service.PlatformOpenAI).
				SetType(service.AccountTypeAPIKey).SetCredentials(credentials).SetExtra(map[string]any{}).Save(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { _ = cache.DeleteAccount(ctx, created.ID); _ = tempCache.DeleteTempUnsched(ctx, created.ID) })
			account, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			svc := service.NewRateLimitService(repo, nil, cfg, nil, tempCache)
			require.True(t, svc.HandleTempUnschedulable(ctx, account, 503, []byte(`{"error":{"message":"capacity"}}`), "model-a"))
			persisted, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			snapshot, err := cache.GetAccount(ctx, created.ID)
			require.NoError(t, err)
			require.NotNil(t, snapshot)
			state, err := tempCache.GetTempUnsched(ctx, created.ID)
			require.NoError(t, err)
			for _, item := range []*service.Account{persisted, snapshot} {
				require.False(t, item.IsSchedulableForModelWithContext(ctx, "alias-a"))
				if scope == service.TempUnschedScopeModel {
					require.Nil(t, item.TempUnschedulableUntil)
					require.True(t, item.IsSchedulableForModelWithContext(ctx, "model-b"))
				} else {
					require.NotNil(t, item.TempUnschedulableUntil)
					require.False(t, item.IsSchedulableForModelWithContext(ctx, "model-b"))
				}
			}
			if scope == service.TempUnschedScopeModel {
				require.Nil(t, state, "model isolation must not write an account-wide Redis block")
				require.NoError(t, repo.SetModelRateLimit(ctx, created.ID, "model-a", time.Now().Add(-time.Minute), "expired"))
				restored, err := repo.GetByID(ctx, created.ID)
				require.NoError(t, err)
				require.True(t, restored.IsSchedulableForModelWithContext(ctx, "alias-a"))
			} else {
				require.NotNil(t, state)
				require.Equal(t, service.TempUnschedScopeAccount, state.Scope)
				require.Empty(t, state.Model)
			}
		})
	}
}
