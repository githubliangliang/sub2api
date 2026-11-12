package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type rotationUserRepo struct {
	service.UserRepository
	user *service.User
}

func (r *rotationUserRepo) GetByID(context.Context, int64) (*service.User, error) {
	copy := *r.user
	return &copy, nil
}

type rotationBarrierCache struct {
	service.RefreshTokenCache
	wg sync.WaitGroup
}

func (c *rotationBarrierCache) GetRefreshToken(ctx context.Context, hash string) (*service.RefreshTokenData, error) {
	data, err := c.RefreshTokenCache.GetRefreshToken(ctx, hash)
	c.wg.Done()
	c.wg.Wait()
	return data, err
}

type rotationFailingConsumeCache struct{ service.RefreshTokenCache }

func (c *rotationFailingConsumeCache) ConsumeRefreshToken(context.Context, string) (*service.RefreshTokenData, error) {
	return nil, errors.New("cache unavailable")
}

func rotationService(cache service.RefreshTokenCache, user *service.User) *service.AuthService {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "rotation-fixture-secret", AccessTokenExpireMinutes: 60, RefreshTokenExpireDays: 7}}
	return service.NewAuthService(nil, &rotationUserRepo{user: user}, nil, cache, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
}

func TestRefreshTokenRotationServiceUsesAtomicMiniredis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()
	cache := NewRefreshTokenCache(rdb)
	user := &service.User{ID: 71, Email: "rotation@example.invalid", Status: service.StatusActive, TokenVersion: 1, TokenVersionResolved: true}
	svc := rotationService(cache, user)
	pair, err := svc.GenerateTokenPair(t.Context(), user, "fixture-family")
	require.NoError(t, err)
	barrier := &rotationBarrierCache{RefreshTokenCache: cache}
	const workers = 8
	barrier.wg.Add(workers)
	concurrentService := rotationService(barrier, user)
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			_, err := concurrentService.RefreshTokenPair(t.Context(), pair.RefreshToken)
			results <- err
		}()
	}
	succeeded := 0
	for i := 0; i < workers; i++ {
		select {
		case err := <-results:
			if err == nil {
				succeeded++
			} else {
				require.ErrorIs(t, err, service.ErrRefreshTokenInvalid)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent refresh did not finish")
		}
	}
	require.Equal(t, 1, succeeded)
	_, err = svc.RefreshTokenPair(t.Context(), pair.RefreshToken)
	require.ErrorIs(t, err, service.ErrRefreshTokenInvalid)
}

func TestRefreshTokenRotationPrechecksAndCacheFailure(t *testing.T) {
	for _, scenario := range []string{"expired", "disabled", "version", "family", "consume failure", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer func() { _ = rdb.Close() }()
			cache := NewRefreshTokenCache(rdb)
			user := &service.User{ID: 72, Status: service.StatusActive, TokenVersion: 1, TokenVersionResolved: true}
			svc := rotationService(cache, user)
			pair, err := svc.GenerateTokenPair(t.Context(), user, "fixture-family")
			require.NoError(t, err)
			sum := sha256.Sum256([]byte(pair.RefreshToken))
			hash := hex.EncodeToString(sum[:])
			ctx := t.Context()
			want := service.ErrRefreshTokenInvalid
			switch scenario {
			case "expired":
				data, err := cache.GetRefreshToken(ctx, hash)
				require.NoError(t, err)
				data.ExpiresAt = time.Now().Add(-time.Second)
				require.NoError(t, cache.StoreRefreshToken(ctx, hash, data, time.Minute))
				want = service.ErrRefreshTokenExpired
			case "disabled":
				user.Status = "disabled"
				want = service.ErrUserNotActive
			case "version":
				user.TokenVersion++
				want = service.ErrTokenRevoked
			case "family":
				require.NoError(t, cache.DeleteTokenFamily(ctx, "fixture-family"))
			case "consume failure":
				svc = rotationService(&rotationFailingConsumeCache{RefreshTokenCache: cache}, user)
				want = service.ErrServiceUnavailable
			case "cancelled":
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelCtx
				want = service.ErrServiceUnavailable
			}
			result, err := svc.RefreshTokenPair(ctx, pair.RefreshToken)
			require.ErrorIs(t, err, want)
			require.Nil(t, result)
			if scenario == "consume failure" || scenario == "cancelled" {
				_, err := cache.GetRefreshToken(t.Context(), hash)
				require.NoError(t, err, "failed consumption must not issue or destroy a token")
			} else {
				_, err := cache.GetRefreshToken(t.Context(), hash)
				require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
			}
		})
	}
}
