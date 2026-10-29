//go:build unit

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGrokOrdinaryEditPreservesDatabaseCatalogSQLite(t *testing.T) {
	_, client := newGrokCatalogSQLiteTestRepository(t)
	ctx := context.Background()
	credentials := map[string]any{"access_token": "local-probe-token"}
	current, err := client.Account.Create().SetName("local catalog probe").SetPlatform(service.PlatformGrok).
		SetType(service.AccountTypeOAuth).SetCredentials(credentials).
		SetExtra(map[string]any{"grok_observed_models": map[string]any{"models": []any{"new-model"}}, service.UpstreamModelMetadataExtraKey: map[string]any{"fresh": true}}).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	stale := &service.Account{ID: current.ID, Name: current.Name, Platform: current.Platform, Type: current.Type,
		Credentials: credentials, Extra: map[string]any{"grok_observed_models": map[string]any{"models": []any{"old-model"}}, service.UpstreamModelMetadataExtraKey: map[string]any{"fresh": false}}}
	merged, err := lockAndMergeAccountProbeExtra(ctx, client, stale, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	models := merged["grok_observed_models"].(map[string]any)["models"].([]any)
	if len(models) != 1 || models[0] != "new-model" {
		t.Fatalf("ordinary edit overwrote catalog %v", models)
	}
	if merged[service.UpstreamModelMetadataExtraKey].(map[string]any)["fresh"] != true {
		t.Fatal("ordinary edit overwrote fresh metadata")
	}

}

func TestGrokCatalogCASSQLiteSemanticExpectedAndSingleKey(t *testing.T) {
	repo, client := newGrokCatalogSQLiteTestRepository(t)
	ctx := context.Background()
	account, err := client.Account.Create().SetName("catalog").SetPlatform(service.PlatformGrok).
		SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).
		SetExtra(map[string]any{"other": true, "grok_observed_models": map[string]any{"models": []any{"old"}, "nested": map[string]any{"a": 1, "b": false}}}).Save(ctx)
	require.NoError(t, err)
	expected := json.RawMessage(` { "nested": { "b": false, "a": 1.0 }, "models": ["old"] } `)
	next := json.RawMessage(`{"models":["Custom-ID"],"fetched_at":"2026-10-29T00:00:00Z"}`)
	applied, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, expected, next)
	require.NoError(t, err)
	require.True(t, applied)
	got, err := client.Account.Get(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, true, got.Extra["other"])
	require.Equal(t, []any{"Custom-ID"}, got.Extra["grok_observed_models"].(map[string]any)["models"])
	applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID, expected, json.RawMessage(`null`))
	require.NoError(t, err)
	require.False(t, applied)
	for _, raw := range []json.RawMessage{json.RawMessage(`{`), json.RawMessage(`{"models":}`)} {
		applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID, raw, next)
		require.ErrorIs(t, err, errGrokCatalogInvalidJSON)
		require.False(t, applied)
		applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID, next, raw)
		require.ErrorIs(t, err, errGrokCatalogInvalidJSON)
		require.False(t, applied)
	}
}

func TestGrokCatalogCASSQLiteLegacyNullExtra(t *testing.T) {
	repo, client := newGrokCatalogSQLiteTestRepository(t)
	ctx := context.Background()
	account, err := client.Account.Create().SetName("legacy null").SetPlatform(service.PlatformGrok).
		SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).Save(ctx)
	require.NoError(t, err)
	_, err = repo.sql.ExecContext(ctx, `UPDATE accounts SET extra='null' WHERE id=?`, account.ID)
	require.NoError(t, err)
	applied, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, json.RawMessage(`{"models":["fresh"]}`))
	require.NoError(t, err)
	require.True(t, applied)
	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Contains(t, got.Extra, "grok_observed_models", "a legacy JSON null root must persist the accepted CAS")
}

type grokCatalogSchedulerCache struct {
	service.SchedulerCache
	writes atomic.Int32
}

func (c *grokCatalogSchedulerCache) SetAccount(context.Context, *service.Account) error {
	c.writes.Add(1)
	return nil
}

func TestGrokCatalogCASSQLiteCallerTransactionRollbackAndCache(t *testing.T) {
	for _, implicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "context", true: "transaction-client"}[implicit], func(t *testing.T) {
			repo, client := newGrokCatalogSQLiteTestRepository(t)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			account, err := client.Account.Create().SetName("transaction").SetPlatform(service.PlatformGrok).
				SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).SetExtra(map[string]any{"other": true}).Save(ctx)
			require.NoError(t, err)
			cache := &grokCatalogSchedulerCache{}
			repo.schedulerCache = cache
			tx, err := client.Tx(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { _ = tx.Rollback() })
			txCtx := dbent.NewTxContext(ctx, tx)
			writer := repo
			if implicit {
				writer = newAccountRepositoryWithSQL(tx.Client(), tx.Client(), cache)
				txCtx = ctx
			}
			scope, err := service.NewGrokCatalogScope(account.ID, service.AccountTypeOAuth, "https://api.x.ai/v1/models")
			require.NoError(t, err)
			now := time.Now().UTC().Add(-time.Second).Format(time.RFC3339Nano)
			_, err = service.PersistGrokCatalogObservation(txCtx, writer, account.ID, service.GrokCatalogObservation{
				Models: []string{"transaction-model"}, Attempt: service.GrokCatalogAttempt{StartedAt: now, FinishedAt: now, Status: "success", Scope: scope},
			})
			require.NoError(t, err)
			require.NoError(t, writer.UpdateExtra(txCtx, account.ID, map[string]any{"uncommitted": true}))
			read, err := writer.GetByID(txCtx, account.ID)
			require.NoError(t, err)
			require.Contains(t, read.Extra, "grok_observed_models")
			require.Equal(t, true, read.Extra["uncommitted"])
			require.Zero(t, cache.writes.Load())
			require.NoError(t, tx.Rollback())
			read, err = repo.GetByID(ctx, account.ID)
			require.NoError(t, err)
			require.NotContains(t, read.Extra, "grok_observed_models")
			require.NotContains(t, read.Extra, "uncommitted")
			require.Zero(t, cache.writes.Load())
			applied, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, json.RawMessage(`{"models":["committed"]}`))
			require.NoError(t, err)
			require.True(t, applied)
			require.EqualValues(t, 1, cache.writes.Load())
		})
	}
}

func TestGrokCatalogCASSQLiteRejectsDeletedPlatformAndCancellation(t *testing.T) {
	repo, client := newGrokCatalogSQLiteTestRepository(t)
	ctx := context.Background()
	account, err := client.Account.Create().SetName("guard").SetPlatform(service.PlatformOpenAI).
		SetType(service.AccountTypeAPIKey).SetCredentials(map[string]any{}).Save(ctx)
	require.NoError(t, err)
	applied, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.False(t, applied)
	_, err = client.Account.UpdateOneID(account.ID).SetPlatform(service.PlatformGrok).SetDeletedAt(time.Now()).Save(ctx)
	require.NoError(t, err)
	applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.False(t, applied)
	applied, err = repo.CompareAndSwapGrokObservedModels(ctx, account.ID+1, nil, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.False(t, applied)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	applied, err = repo.CompareAndSwapGrokObservedModels(canceled, account.ID, nil, json.RawMessage(`{}`))
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, applied)
}

func TestGrokCatalogCASSQLiteConcurrentWritersAndExtraUpdates(t *testing.T) {
	repo, client := newGrokCatalogSQLiteWALRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	account, err := client.Account.Create().SetName("race").SetPlatform(service.PlatformGrok).
		SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).SetExtra(map[string]any{"other": 0}).Save(ctx)
	require.NoError(t, err)
	var applied atomic.Int32
	var wg sync.WaitGroup
	errors := make(chan error, 9)
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ok, err := repo.CompareAndSwapGrokObservedModels(ctx, account.ID, nil, json.RawMessage(`{"models":["winner"]}`))
			if ok {
				applied.Add(1)
			}
			errors <- err
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 1; i <= 20; i++ {
			if err := repo.UpdateExtra(ctx, account.ID, map[string]any{"other": i}); err != nil {
				errors <- err
				return
			}
		}
		errors <- nil
	}()
	close(start)
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, applied.Load())
	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.EqualValues(t, 20, got.Extra["other"])
	require.Equal(t, []any{"winner"}, got.Extra["grok_observed_models"].(map[string]any)["models"])
}

func TestGrokOrdinaryEditMissingAndInjectedManagedKeysSQLite(t *testing.T) {
	for _, incoming := range []map[string]any{nil, {}, {"grok_observed_models": "injected", service.UpstreamModelMetadataExtraKey: false}} {
		repo, client := newGrokCatalogSQLiteTestRepository(t)
		ctx := context.Background()
		account, err := client.Account.Create().SetName("edit").SetPlatform(service.PlatformGrok).
			SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).
			SetExtra(map[string]any{"grok_observed_models": map[string]any{"models": []any{"fresh"}}, service.UpstreamModelMetadataExtraKey: map[string]any{"fresh": true}}).Save(ctx)
		require.NoError(t, err)
		edit, err := repo.GetByID(ctx, account.ID)
		require.NoError(t, err)
		edit.Name = "edited"
		edit.Extra = incoming
		require.NoError(t, repo.Update(ctx, edit))
		got, err := repo.GetByID(ctx, account.ID)
		require.NoError(t, err)
		require.Equal(t, []any{"fresh"}, got.Extra["grok_observed_models"].(map[string]any)["models"])
		require.Equal(t, true, got.Extra[service.UpstreamModelMetadataExtraKey].(map[string]any)["fresh"])
	}
}

func newGrokCatalogSQLiteTestRepository(t *testing.T) (*accountRepository, *dbent.Client) {
	t.Helper()
	repo, client := newOllamaCloudUsageSQLiteRepository(t)
	require.NoError(t, EnsureSQLiteAuxTables(context.Background(), repo.sql.(*sql.DB)))
	return repo, client
}

func newGrokCatalogSQLiteWALRepository(t *testing.T) (*accountRepository, *dbent.Client) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "catalog.db")+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)&_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(8)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(context.Background()))
	require.NoError(t, EnsureSQLiteAuxTables(context.Background(), db))
	return newAccountRepositoryWithSQL(client, db, nil), client
}

func TestGrokOrdinaryEditOverlapsCatalogObservationSQLite(t *testing.T) {
	repo, client := newGrokCatalogSQLiteWALRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	account, err := client.Account.Create().SetName("overlap").SetPlatform(service.PlatformGrok).
		SetType(service.AccountTypeOAuth).SetCredentials(map[string]any{}).
		SetExtra(map[string]any{"grok_observed_models": map[string]any{"models": []any{"old"}}, service.UpstreamModelMetadataExtraKey: map[string]any{"fresh": false}}).Save(ctx)
	require.NoError(t, err)
	stale, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	stale.Name = "edited"
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })
	txCtx := dbent.NewTxContext(ctx, tx)
	applied, err := repo.CompareAndSwapGrokObservedModels(txCtx, account.ID, json.RawMessage(`{"models":["old"]}`), json.RawMessage(`{"models":["fresh"]}`))
	require.NoError(t, err)
	require.True(t, applied)
	require.NoError(t, repo.UpdateExtra(txCtx, account.ID, map[string]any{service.UpstreamModelMetadataExtraKey: map[string]any{"fresh": true}}))
	started, finished := make(chan struct{}), make(chan error, 1)
	go func() { close(started); finished <- repo.Update(ctx, stale) }()
	<-started
	require.NoError(t, tx.Commit())
	select {
	case err := <-finished:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "edited", got.Name)
	require.Equal(t, []any{"fresh"}, got.Extra["grok_observed_models"].(map[string]any)["models"])
	require.Equal(t, true, got.Extra[service.UpstreamModelMetadataExtraKey].(map[string]any)["fresh"])
}
