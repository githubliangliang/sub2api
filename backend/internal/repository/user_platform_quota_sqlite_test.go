package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/ent/userplatformquota"
	"github.com/stretchr/testify/require"
)

func newUserPlatformQuotaSQLite(t *testing.T, legacyWindows bool) (*userPlatformQuotaRepository, *dbent.Client, int64) {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "quota.db") + "?_time_format=sqlite&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	groupRepo, client := newGroupEntRepoSQLite(t, dsn)
	db, ok := groupRepo.sql.(*sql.DB)
	require.True(t, ok)
	db.SetMaxOpenConns(4)
	if legacyWindows {
		_, err := client.ExecContext(context.Background(), "DROP TABLE user_platform_quotas")
		require.NoError(t, err)
		ddl, err := os.ReadFile("../../migrations/142_user_platform_quotas.sql")
		require.NoError(t, err)
		_, err = client.ExecContext(context.Background(), string(ddl))
		require.NoError(t, err)
	}
	user, err := client.User.Create().SetEmail("quota@example.invalid").SetPasswordHash("fixture-hash").Save(context.Background())
	require.NoError(t, err)
	return &userPlatformQuotaRepository{client: client}, client, user.ID
}

func TestUserPlatformQuotaSQLiteInitialSkipsUnlimitedAndPreservesZero(t *testing.T) {
	repo, client, id := newUserPlatformQuotaSQLite(t, false)
	ctx := context.Background()
	zero, limit := 0.0, 10.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{
		{UserID: id, Platform: "anthropic"},
		{UserID: id, Platform: "openai", DailyLimitUSD: &zero},
		{UserID: id, Platform: "gemini", MonthlyLimitUSD: &limit},
	}))
	count, err := client.UserPlatformQuota.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, count)
	got, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, &zero, got.DailyLimitUSD)
	weekly := 20.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{
		{UserID: id, Platform: "openai", DailyLimitUSD: &limit, WeeklyLimitUSD: &weekly},
	}))
	got, err = repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Equal(t, &zero, got.DailyLimitUSD)
	require.Equal(t, &weekly, got.WeeklyLimitUSD)
}

func TestUserPlatformQuotaSQLiteClearDoesNotReviveAndReconfigureStartsFresh(t *testing.T) {
	repo, client, id := newUserPlatformQuotaSQLite(t, false)
	ctx := context.Background()
	limit, nextLimit := 10.0, 20.0
	now := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	require.NoError(t, repo.UpsertForUser(ctx, id, []UserPlatformQuotaRecord{{Platform: "openai", DailyLimitUSD: &limit}}))
	require.NoError(t, repo.IncrementUsageWithReset(ctx, id, "openai", 3, now))
	before, err := client.UserPlatformQuota.Query().Only(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.UpsertForUser(ctx, id, []UserPlatformQuotaRecord{{Platform: "openai"}}))
	cleared, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Nil(t, cleared)
	require.ErrorIs(t, repo.ResetExpiredWindow(ctx, id, "openai", "daily", now), ErrUserPlatformQuotaNotFound)
	require.NoError(t, repo.IncrementUsageWithReset(ctx, id, "openai", 4, now))
	require.NoError(t, repo.BatchSnapshotUsage(ctx, []UserPlatformQuotaSnapshot{{
		UserID: id, Platform: "openai", DailyUsageUSD: 8, DailyWindowStart: now, WeeklyWindowStart: now, MonthlyWindowStart: now,
	}}, now))
	count, err := client.UserPlatformQuota.Query().Count(mixins.SkipSoftDelete(ctx))
	require.NoError(t, err)
	require.Equal(t, 1, count, "usage writes must not create a replacement for a cleared record")
	deleted, err := client.UserPlatformQuota.Get(mixins.SkipSoftDelete(ctx), before.ID)
	require.NoError(t, err)
	require.NotNil(t, deleted.DeletedAt)
	require.Equal(t, before.DailyUsageUsd, deleted.DailyUsageUsd)
	require.NoError(t, repo.UpsertForUser(ctx, id, []UserPlatformQuotaRecord{{Platform: "openai", DailyLimitUSD: &nextLimit}}))
	active, err := client.UserPlatformQuota.Query().Where(userplatformquota.DeletedAtIsNil()).Only(ctx)
	require.NoError(t, err)
	require.NotEqual(t, before.ID, active.ID)
	require.Zero(t, active.DailyUsageUsd)
	require.Nil(t, active.DailyWindowStart)
	require.Nil(t, active.WeeklyWindowStart)
	require.Nil(t, active.MonthlyWindowStart)
}

func TestUserPlatformQuotaSQLiteUsageMissingIsNoop(t *testing.T) {
	repo, client, id := newUserPlatformQuotaSQLite(t, false)
	ctx := context.Background()
	now := time.Now().UTC()
	require.NoError(t, repo.IncrementUsageWithReset(ctx, id, "openai", 1, now))
	require.NoError(t, repo.BatchSnapshotUsage(ctx, []UserPlatformQuotaSnapshot{{UserID: id, Platform: "openai"}, {UserID: id + 10000, Platform: "gemini"}}, now))
	count, err := client.UserPlatformQuota.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestUserPlatformQuotaSQLiteSnapshotsUpdateExistingAndChunkLargeInput(t *testing.T) {
	repo, client, id := newUserPlatformQuotaSQLite(t, false)
	ctx := context.Background()
	zero := 0.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{{UserID: id, Platform: "openai", DailyLimitUSD: &zero}}))
	now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
	snaps := make([]UserPlatformQuotaSnapshot, 4100)
	for i := range snaps {
		snaps[i] = UserPlatformQuotaSnapshot{UserID: int64(i + 10000), Platform: "openai", DailyUsageUSD: 7, DailyWindowStart: now, WeeklyWindowStart: now, MonthlyWindowStart: now}
	}
	snaps[0].UserID = id
	require.NoError(t, repo.BatchSnapshotUsage(ctx, snaps, now))
	got, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Equal(t, 7.0, got.DailyUsageUSD)
	require.Equal(t, &zero, got.DailyLimitUSD)
	count, err := client.UserPlatformQuota.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestUserPlatformQuotaSQLiteSnapshotChunksRollBackTogether(t *testing.T) {
	repo, client, id := newUserPlatformQuotaSQLite(t, false)
	ctx := context.Background()
	second, err := client.User.Create().SetEmail("second@example.invalid").SetPasswordHash("fixture-hash").Save(ctx)
	require.NoError(t, err)
	limit := 10.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{{UserID: id, Platform: "openai", DailyLimitUSD: &limit}, {UserID: second.ID, Platform: "openai", DailyLimitUSD: &limit}}))
	_, err = client.ExecContext(ctx, fmt.Sprintf(`CREATE TRIGGER reject_second_snapshot BEFORE UPDATE ON user_platform_quotas WHEN NEW.user_id = %d BEGIN SELECT RAISE(ABORT, 'snapshot fixture failure'); END`, second.ID))
	require.NoError(t, err)
	now := time.Now().UTC()
	snaps := make([]UserPlatformQuotaSnapshot, 1001)
	for i := range snaps {
		snaps[i] = UserPlatformQuotaSnapshot{UserID: int64(i + 10000), Platform: "openai", DailyUsageUSD: 9}
	}
	snaps[0].UserID = id
	snaps[len(snaps)-1].UserID = second.ID
	require.ErrorContains(t, repo.BatchSnapshotUsage(ctx, snaps, now), "snapshot fixture failure")
	got, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Zero(t, got.DailyUsageUSD)
}

func TestUserPlatformQuotaSQLiteLegacyWindowColumnsRoundTrip(t *testing.T) {
	repo, _, id := newUserPlatformQuotaSQLite(t, true)
	ctx := context.Background()
	limit := 10.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{{UserID: id, Platform: "openai", DailyLimitUSD: &limit}}))
	now := time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)
	require.NoError(t, repo.IncrementUsageWithReset(ctx, id, "openai", 2, now))
	got, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.NotNil(t, got.DailyWindowStart)
	require.NotNil(t, got.WeeklyWindowStart)
	require.NoError(t, repo.IncrementUsageWithReset(ctx, id, "openai", 3, now.Add(time.Minute)))
	got, err = repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Equal(t, 5.0, got.DailyUsageUSD)
}

func TestUserPlatformQuotaSQLiteConcurrentIncrementsDoNotLoseUsage(t *testing.T) {
	repo, _, id := newUserPlatformQuotaSQLite(t, true)
	ctx := context.Background()
	limit := 100.0
	require.NoError(t, repo.BulkInsertInitial(ctx, []UserPlatformQuotaRecord{{UserID: id, Platform: "openai", DailyLimitUSD: &limit}}))
	now := time.Date(2026, 10, 28, 2, 0, 0, 0, time.UTC)
	const requests = 12
	errors := make(chan error, requests)
	var workers sync.WaitGroup
	for i := 0; i < requests; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			errors <- repo.IncrementUsageWithReset(ctx, id, "openai", 1, now)
		}()
	}
	workers.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	got, err := repo.GetByUserPlatform(ctx, id, "openai")
	require.NoError(t, err)
	require.Equal(t, float64(requests), got.DailyUsageUSD)
	require.Equal(t, float64(requests), got.WeeklyUsageUSD)
	require.Equal(t, float64(requests), got.MonthlyUsageUSD)
}
