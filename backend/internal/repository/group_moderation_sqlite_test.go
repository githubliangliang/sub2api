package repository

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupDeletionCleansModerationScopeSQLite(t *testing.T) {
	for _, cascade := range []bool{false, true} {
		t.Run(fmt.Sprint(cascade), func(t *testing.T) {
			repo, client := newGroupEntRepoSQLite(t)
			ctx := context.Background()
			g := &service.Group{Name: "scope", Platform: service.PlatformOpenAI, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard}
			require.NoError(t, repo.Create(ctx, g))
			settings := NewSettingRepository(client)
			original := fmt.Sprintf(`{"enabled":true,"group_ids":[9,%d,7,%d],"extra":{"number":9007199254740993}}`, g.ID, g.ID)
			require.NoError(t, settings.Set(ctx, service.SettingKeyContentModerationConfig, original))
			if cascade {
				_, err := repo.DeleteCascade(ctx, g.ID)
				require.NoError(t, err)
			} else {
				require.NoError(t, repo.Delete(ctx, g.ID))
			}
			value, err := settings.GetValue(ctx, service.SettingKeyContentModerationConfig)
			require.NoError(t, err)
			require.JSONEq(t, `{"enabled":true,"group_ids":[9,7],"extra":{"number":9007199254740993}}`, value)
			require.Contains(t, value, "9007199254740993")
		})
	}
}

func TestGroupDeletionPreservesMalformedModerationSQLite(t *testing.T) {
	for _, value := range []string{"", `{`, `null`, `[]`, `{"group_ids":null}`, `{"group_ids":"1"}`, `{"group_ids":[1,"bad"]}`, `{"enabled":true}`} {
		t.Run(value, func(t *testing.T) {
			repo, client := newGroupEntRepoSQLite(t)
			ctx := context.Background()
			g := &service.Group{Name: "scope", Platform: service.PlatformOpenAI, Status: service.StatusActive}
			require.NoError(t, repo.Create(ctx, g))
			settings := NewSettingRepository(client)
			require.NoError(t, settings.Set(ctx, service.SettingKeyContentModerationConfig, value))
			_, err := repo.DeleteCascade(ctx, g.ID)
			require.NoError(t, err)
			got, err := settings.GetValue(ctx, service.SettingKeyContentModerationConfig)
			require.NoError(t, err)
			require.Equal(t, value, got)
		})
	}
}

func TestGroupDeletionRollsBackOnModerationDatabaseErrorSQLite(t *testing.T) {
	repo, client := newGroupEntRepoSQLite(t)
	ctx := context.Background()
	g := &service.Group{Name: "scope", Platform: service.PlatformOpenAI, Status: service.StatusActive}
	require.NoError(t, repo.Create(ctx, g))
	settings := NewSettingRepository(client)
	value := fmt.Sprintf(`{"group_ids":[%d]}`, g.ID)
	require.NoError(t, settings.Set(ctx, service.SettingKeyContentModerationConfig, value))
	_, err := client.ExecContext(ctx, `CREATE TRIGGER fail_moderation_update BEFORE UPDATE ON settings BEGIN SELECT RAISE(ABORT, 'settings unavailable'); END`)
	require.NoError(t, err)
	_, err = repo.DeleteCascade(ctx, g.ID)
	require.ErrorContains(t, err, "settings unavailable")
	_, err = repo.GetByID(ctx, g.ID)
	require.NoError(t, err)
	got, err := settings.GetValue(ctx, service.SettingKeyContentModerationConfig)
	require.NoError(t, err)
	require.Equal(t, value, got)
}

func TestModerationSaveAndGroupDeleteSerializeSQLite(t *testing.T) {
	for _, deleteFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("delete_first=%t", deleteFirst), func(t *testing.T) {
			dsn := "file:" + filepath.Join(t.TempDir(), "groups.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_time_format=sqlite"
			repo, client := newGroupEntRepoSQLite(t, dsn)
			sqlDB, ok := repo.sql.(*sql.DB)
			require.True(t, ok)
			sqlDB.SetMaxOpenConns(4)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			g := &service.Group{Name: "scope", Platform: service.PlatformOpenAI, Status: service.StatusActive}
			require.NoError(t, repo.Create(ctx, g))
			settings := &settingRepository{client: client}
			oldValue := fmt.Sprintf(`{"group_ids":[%d],"note":"old"}`, g.ID)
			newValue := fmt.Sprintf(`{"group_ids":[%d],"note":"new"}`, g.ID)
			require.NoError(t, settings.Set(ctx, service.SettingKeyContentModerationConfig, oldValue))
			tx, err := client.Tx(ctx)
			require.NoError(t, err)
			defer func() { _ = tx.Rollback() }()
			txCtx := dbent.NewTxContext(ctx, tx)
			if deleteFirst {
				_, err = repo.DeleteCascade(txCtx, g.ID)
			} else {
				err = settings.SetContentModerationConfig(txCtx, newValue, []int64{g.ID})
			}
			require.NoError(t, err)
			started := make(chan struct{})
			done := make(chan error, 1)
			go func() {
				close(started)
				if deleteFirst {
					done <- settings.SetContentModerationConfig(ctx, newValue, []int64{g.ID})
				} else {
					_, err := repo.DeleteCascade(ctx, g.ID)
					done <- err
				}
			}()
			<-started
			select {
			case err := <-done:
				t.Fatalf("second writer bypassed the open transaction: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			require.NoError(t, tx.Commit())
			select {
			case err = <-done:
			case <-ctx.Done():
				t.Fatal("second writer did not finish")
			}
			if deleteFirst {
				require.ErrorIs(t, err, service.ErrGroupNotFound)
			} else {
				require.NoError(t, err)
			}
			value, err := settings.GetValue(ctx, service.SettingKeyContentModerationConfig)
			require.NoError(t, err)
			note := "new"
			if deleteFirst {
				note = "old"
			}
			require.JSONEq(t, fmt.Sprintf(`{"group_ids":[],"note":%q}`, note), value)
			_, err = repo.GetByID(ctx, g.ID)
			require.ErrorIs(t, err, service.ErrGroupNotFound)
		})
	}
}
