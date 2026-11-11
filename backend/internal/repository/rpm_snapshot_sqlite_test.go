package repository

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type countingRPMRepository struct {
	service.UserGroupRateRepository
	calls atomic.Int32
}

func (r *countingRPMRepository) GetRPMOverrideByUserAndGroup(ctx context.Context, user, group int64) (*int, error) {
	r.calls.Add(1)
	return r.UserGroupRateRepository.GetRPMOverrideByUserAndGroup(ctx, user, group)
}

func TestRPMSnapshotLoadedNilSQLiteAndMiniredis(t *testing.T) {
	ctx := context.Background()
	registerSQLitePGCompatFunctions()
	db, err := sql.Open("sqlite", "file:"+t.TempDir()+"/rpm.db?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
	_, err = db.Exec(`INSERT INTO groups(id,name,platform,rpm_limit) VALUES(901,'rpm','openai',100); INSERT INTO users(id,email,password_hash,balance) VALUES(901,'rpm@example.test','fixture',10); INSERT INTO api_keys(id,user_id,key,name,group_id,status) VALUES(901,901,'rpm-fixture','rpm',901,'active')`)
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	rates := &countingRPMRepository{UserGroupRateRepository: NewUserGroupRateRepository(db)}
	cfg := &config.Config{}
	cache := NewBillingCache(rdb)
	require.NoError(t, cache.SetUserBalance(ctx, 901, 10))
	auth := service.NewAPIKeyService(NewAPIKeyRepository(client, db), nil, nil, nil, rates, NewAPIKeyCache(rdb), cfg)
	billing := service.NewBillingCacheService(cache, nil, nil, nil, NewUserRPMCache(rdb), rates, cfg, nil)
	t.Cleanup(billing.Stop)
	key, err := auth.GetByKey(ctx, "rpm-fixture")
	require.NoError(t, err)
	require.Nil(t, key.User.UserGroupRPMOverride)
	require.EqualValues(t, 901, key.User.UserGroupRPMOverrideGroupID)
	loadedCalls := rates.calls.Load()
	for i := 0; i < 20; i++ {
		require.NoError(t, billing.CheckBillingEligibility(ctx, key.User, key, key.Group, nil, "openai"))
	}
	require.Equal(t, loadedCalls, rates.calls.Load(), "loaded nil override must not reread SQLite per request")
	other := *key.Group
	other.ID = 902
	require.NoError(t, billing.CheckBillingEligibility(ctx, key.User, key, &other, nil, "openai"))
	require.Equal(t, loadedCalls+1, rates.calls.Load(), "different group must query its own override")
	// Invalidate and reload after an explicit zero override is persisted.
	_, err = db.Exec(`INSERT INTO user_group_rate_multipliers(user_id,group_id,rate_multiplier,rpm_override) VALUES(901,901,1,0)`)
	require.NoError(t, err)
	mini.FastForward(24 * time.Hour)
	refreshed := service.NewAPIKeyService(NewAPIKeyRepository(client, db), nil, nil, nil, rates, NewAPIKeyCache(rdb), cfg)
	key, err = refreshed.GetByKey(ctx, "rpm-fixture")
	require.NoError(t, err)
	require.NotNil(t, key.User.UserGroupRPMOverride)
	require.Zero(t, *key.User.UserGroupRPMOverride)
}
