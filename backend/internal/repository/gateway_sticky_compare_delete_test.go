package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestStickySessionCompareDeleteMiniredis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewGatewayCache(rdb)
	conditional, ok := cache.(service.StickySessionAccountCompareDeleter)
	require.True(t, ok)
	ctx := context.Background()
	require.NoError(t, cache.SetSessionAccountID(ctx, 1, "sess", 5, time.Hour))
	require.NoError(t, cache.SetSessionAccountID(ctx, 2, "sess", 5, time.Hour))
	require.NoError(t, cache.SetSessionAccountID(ctx, 1, "other", 5, time.Hour))
	old, err := cache.GetSessionAccountID(ctx, 1, "sess")
	require.NoError(t, err)
	updated := make(chan struct{})
	go func() { _ = cache.SetSessionAccountID(ctx, 1, "sess", 6, time.Hour); close(updated) }()
	<-updated
	removed, err := conditional.DeleteSessionAccountIDIfMatch(ctx, 1, "sess", old)
	require.NoError(t, err)
	require.False(t, removed)
	bound, err := cache.GetSessionAccountID(ctx, 1, "sess")
	require.NoError(t, err)
	require.Equal(t, int64(6), bound)
	removed, err = conditional.DeleteSessionAccountIDIfMatch(ctx, 1, "sess", 6)
	require.NoError(t, err)
	require.True(t, removed)
	_, err = cache.GetSessionAccountID(ctx, 1, "sess")
	require.ErrorIs(t, err, service.ErrStickySessionNotFound)
	for _, scope := range []struct {
		group int64
		hash  string
	}{{2, "sess"}, {1, "other"}} {
		bound, err = cache.GetSessionAccountID(ctx, scope.group, scope.hash)
		require.NoError(t, err)
		require.Equal(t, int64(5), bound)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	removed, err = conditional.DeleteSessionAccountIDIfMatch(canceled, 1, "other", 5)
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, removed)
	bound, err = cache.GetSessionAccountID(ctx, 1, "other")
	require.NoError(t, err)
	require.Equal(t, int64(5), bound)
}
