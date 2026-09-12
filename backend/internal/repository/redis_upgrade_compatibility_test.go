package repository

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisUpgradeActiveIndexCleanupBounds(t *testing.T) {
	mr := miniredis.RunT(t)
	now := time.Unix(1_700_000_000, 0)
	mr.SetTime(now)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewConcurrencyCache(client, 15, 60)
	ctx := context.Background()
	for _, key := range []string{accountActiveIndexKey, userActiveIndexKey} {
		members := make([]redis.Z, 0, 1004)
		for id := 1; id <= 1003; id++ {
			members = append(members, redis.Z{Score: float64(now.Unix()), Member: strconv.Itoa(id)})
		}
		members = append(members, redis.Z{Score: float64(now.Unix() + 1), Member: "future"})
		require.NoError(t, client.ZAdd(ctx, key, members...).Err())
	}
	require.NoError(t, cache.CleanupExpiredAccountSlotKeys(ctx))
	for _, key := range []string{accountActiveIndexKey, userActiveIndexKey} {
		require.EqualValues(t, 4, client.ZCard(ctx, key).Val(), "only 1000 due entries may be removed per index")
		require.Equal(t, float64(now.Unix()+1), client.ZScore(ctx, key, "future").Val())
	}
	require.NoError(t, cache.CleanupExpiredAccountSlotKeys(ctx))
	for _, key := range []string{accountActiveIndexKey, userActiveIndexKey} {
		require.Equal(t, []string{"future"}, client.ZRange(ctx, key, 0, -1).Val(), "score equal to now is due; future score is untouched")
	}
}

func TestRedisUpgradeQueueIndexPreservesLiveLocksAndBatchLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	now := time.Unix(1_700_000_000, 0)
	mr.SetTime(now)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewUserMsgQueueCache(client)
	ctx := context.Background()
	// Due records are ordered before the future record, including exact-now.
	require.NoError(t, client.Set(ctx, umqLockKey(1), "live", time.Minute).Err())
	require.NoError(t, client.Set(ctx, umqLockKey(2), "no-ttl", 0).Err())
	require.NoError(t, client.Set(ctx, umqLockKey(4), "future", 0).Err())
	require.NoError(t, client.ZAdd(ctx, umqLockIndexKey,
		redis.Z{Score: float64(now.UnixMilli() - 2), Member: "1"},
		redis.Z{Score: float64(now.UnixMilli() - 1), Member: "2"},
		redis.Z{Score: float64(now.UnixMilli()), Member: "3"},
		redis.Z{Score: float64(now.UnixMilli() + 1), Member: "4"},
	).Err())
	cleaned, err := cache.ReconcileExpiredLockCandidates(ctx, 1)
	require.NoError(t, err)
	require.Zero(t, cleaned)
	require.Equal(t, "live", client.Get(ctx, umqLockKey(1)).Val())
	require.Greater(t, client.ZScore(ctx, umqLockIndexKey, "1").Val(), float64(now.UnixMilli()))
	require.Equal(t, "no-ttl", client.Get(ctx, umqLockKey(2)).Val(), "batch bound must leave second due lock for next call")
	cleaned, err = cache.ReconcileExpiredLockCandidates(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, 1, cleaned)
	require.ErrorIs(t, client.Get(ctx, umqLockKey(2)).Err(), redis.Nil)
	require.ErrorIs(t, client.ZScore(ctx, umqLockIndexKey, "3").Err(), redis.Nil)
	require.Equal(t, "future", client.Get(ctx, umqLockKey(4)).Val())
	require.EqualValues(t, 2, client.ZCard(ctx, umqLockIndexKey).Val())
}

func TestRedisUpgradeEmbeddedPoolRecoversAfterCanceledWaiter(t *testing.T) {
	disabled := false
	client := InitRedis(&config.Config{Redis: config.RedisConfig{Enabled: &disabled, PoolSize: 1, MinIdleConns: 1, DialTimeoutSeconds: 1, ReadTimeoutSeconds: 1, WriteTimeoutSeconds: 1}})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	held := client.Conn()
	require.NoError(t, held.Ping(ctx).Err())
	cache := NewUserMsgQueueCache(client)
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	_, err := cache.GetLastCompletedMs(waitCtx, 99901)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NoError(t, held.Close())

	var workers sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		workers.Go(func() {
			for turn := 0; turn < 8; turn++ {
				id := int64(100000 + worker*10 + turn)
				request := fmt.Sprintf("worker-%d-turn-%d", worker, turn)
				opCtx, done := context.WithTimeout(ctx, time.Second)
				acquired, err := cache.AcquireLock(opCtx, id, request, 1000)
				if err != nil || !acquired {
					t.Errorf("acquire after canceled pool waiter: acquired=%v err=%v", acquired, err)
					done()
					return
				}
				released, err := cache.ReleaseLock(opCtx, id, request)
				if err != nil || !released {
					t.Errorf("release after canceled pool waiter: released=%v err=%v", released, err)
				}
				done()
			}
		})
	}
	workers.Wait()
	require.NoError(t, client.Ping(ctx).Err())
}
