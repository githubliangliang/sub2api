//go:build unit

package repository

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOpenAIRefreshCASAndOutboxSQLite(t *testing.T) {
	repo, client := newGrokCatalogSQLiteTestRepository(t)
	ctx := context.Background()
	expected := map[string]any{"access_token": "fixture-old", "refresh_token": "fixture-refresh", "nested": map[string]any{"a": 1, "b": true}}
	replacement := map[string]any{"access_token": "fixture-new", "refresh_token": "fixture-next"}
	a, err := client.Account.Create().SetName("local rotation fixture").SetPlatform(service.PlatformOpenAI).SetType(service.AccountTypeOAuth).SetCredentials(expected).Save(ctx)
	require.NoError(t, err)
	db := repo.sql.(*sql.DB)
	var events int
	countEvents := func() int {
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM scheduler_outbox WHERE account_id = ?", a.ID).Scan(&events))
		return events
	}
	before := countEvents()
	// A semantically equal snapshot has a different object order and numeric type.
	expected["nested"] = map[string]any{"b": true, "a": float64(1)}
	ok, err := repo.UpdateOpenAIOAuthCredentialsIfUnchanged(ctx, a.ID, expected, nil, replacement)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, before+1, countEvents())
	ok, err = repo.UpdateOpenAIOAuthCredentialsIfUnchanged(ctx, a.ID, expected, nil, map[string]any{"access_token": "stale"})
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, before+1, countEvents())
	got, err := client.Account.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, replacement, got.Credentials)

	_, err = db.Exec(`CREATE TRIGGER fail_rotation_outbox BEFORE INSERT ON scheduler_outbox BEGIN SELECT RAISE(ABORT, 'fixture outbox failure'); END`)
	require.NoError(t, err)
	ok, err = repo.UpdateOpenAIOAuthCredentialsIfUnchanged(ctx, a.ID, replacement, nil, expected)
	require.Error(t, err)
	require.False(t, ok)
	got, err = client.Account.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, replacement, got.Credentials, "outbox failure must roll back credentials")
}

type localOpenAIRotationExecutor struct{ cancel context.CancelFunc }

func (e *localOpenAIRotationExecutor) CanRefresh(*service.Account) bool                  { return true }
func (e *localOpenAIRotationExecutor) NeedsRefresh(*service.Account, time.Duration) bool { return true }
func (e *localOpenAIRotationExecutor) CacheKey(*service.Account) string {
	return "openai:local-rotation"
}
func (e *localOpenAIRotationExecutor) Refresh(context.Context, *service.Account) (map[string]any, error) {
	e.cancel()
	return map[string]any{"access_token": "local-next", "refresh_token": "local-next-refresh"}, nil
}

func TestOpenAIRotationCanceledParentSQLiteMiniredis(t *testing.T) {
	repo, client := newGrokCatalogSQLiteTestRepository(t)
	account, err := client.Account.Create().SetName("local rotation").SetPlatform(service.PlatformOpenAI).SetType(service.AccountTypeOAuth).SetStatus(service.StatusActive).SetCredentials(map[string]any{"access_token": "local-old", "refresh_token": "local-old-refresh"}).Save(context.Background())
	require.NoError(t, err)
	a, err := repo.GetByID(context.Background(), account.ID)
	require.NoError(t, err)
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer func() { require.NoError(t, rdb.Close()) }()
	cache := NewGeminiTokenCache(rdb)
	require.NoError(t, cache.SetAccessToken(context.Background(), "openai:local-rotation", "local-old", time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := service.NewOAuthRefreshAPI(repo, cache).RefreshIfNeeded(ctx, a, &localOpenAIRotationExecutor{cancel: cancel}, time.Hour)
	require.NoError(t, err)
	require.True(t, result.Refreshed)
	require.Equal(t, "local-next-refresh", result.Account.GetOpenAIRefreshToken())
	require.False(t, mini.Exists(oauthTokenKeyPrefix+"openai:local-rotation"))
	require.False(t, mini.Exists(oauthRefreshLockKeyPrefix+"openai:local-rotation"))
	stored, err := client.Account.Get(context.Background(), a.ID)
	require.NoError(t, err)
	require.Equal(t, "local-next", stored.Credentials["access_token"])
}

func TestOpenAIRefreshConcurrentCASHasOneWinnerSQLite(t *testing.T) {
	repo, client := newGrokCatalogSQLiteWALRepository(t)
	ctx := context.Background()
	expected := map[string]any{"access_token": "local-old", "refresh_token": "local-refresh"}
	a, err := client.Account.Create().SetName("concurrent local rotation").SetPlatform(service.PlatformOpenAI).SetType(service.AccountTypeOAuth).SetCredentials(expected).Save(ctx)
	require.NoError(t, err)
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make(chan bool, 2)
	errs := make(chan error, 2)
	for _, token := range []string{"local-first", "local-second"} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			<-start
			ok, err := repo.UpdateOpenAIOAuthCredentialsIfUnchanged(ctx, a.ID, expected, nil, map[string]any{"access_token": token, "refresh_token": token})
			results <- ok
			errs <- err
		}(token)
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	winners := 0
	for ok := range results {
		if ok {
			winners++
		}
	}
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, 1, winners)
	var events int
	require.NoError(t, repo.sql.(*sql.DB).QueryRow("SELECT COUNT(*) FROM scheduler_outbox WHERE account_id=?", a.ID).Scan(&events))
	require.Equal(t, 1, events)
}
