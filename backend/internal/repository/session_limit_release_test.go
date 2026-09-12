package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestSessionLimitUnregisterFreesOnlyRequestedSlot(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewSessionLimitCache(rdb, 5)
	ctx := context.Background()
	for _, id := range []string{"failed", "served"} {
		allowed, err := cache.RegisterSession(ctx, 42, id, 2, 5*time.Minute)
		require.NoError(t, err)
		require.True(t, allowed)
	}
	allowed, err := cache.RegisterSession(ctx, 42, "next", 2, 5*time.Minute)
	require.NoError(t, err)
	require.False(t, allowed)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cache.UnregisterSession(ctx, 42, "failed"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	require.NoError(t, cache.UnregisterSession(ctx, 42, "missing"))
	require.NoError(t, cache.UnregisterSession(ctx, 42, ""))
	allowed, err = cache.RegisterSession(ctx, 42, "next", 2, 5*time.Minute)
	require.NoError(t, err)
	require.True(t, allowed)
	active, err := cache.IsSessionActive(ctx, 42, "served")
	require.NoError(t, err)
	require.True(t, active)
	count, err := cache.GetActiveSessionCount(ctx, 42)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}
