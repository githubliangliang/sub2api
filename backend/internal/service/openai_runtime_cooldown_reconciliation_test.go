//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type cooldownReconciliationRepo struct {
	AccountRepository
	account  *Account
	writeErr error
}

func (r *cooldownReconciliationRepo) GetByID(context.Context, int64) (*Account, error) {
	cloned := *r.account
	return &cloned, nil
}

func (r *cooldownReconciliationRepo) SetTempUnschedulable(_ context.Context, _ int64, until time.Time, _ string) error {
	if r.writeErr != nil {
		return r.writeErr
	}
	r.account.TempUnschedulableUntil = &until
	return nil
}

func TestRuntimeCooldownSnapshotLagFailsOpenBeforeDBRecheck(t *testing.T) {
	ctx := context.Background()
	account := &Account{ID: 93, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "test"}}
	stale := *account
	repo := &cooldownReconciliationRepo{account: account}
	cache := &openAISnapshotCacheStub{snapshotAccounts: []*Account{&stale}, accountsByID: map[int64]*Account{93: &stale}}
	snapshot := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)
	svc := &OpenAIGatewayService{cfg: &config.Config{RunMode: config.RunModeSimple}, schedulerSnapshot: snapshot, accountRepo: repo}
	svc.tempUnscheduleOpenAITransportError(ctx, account, "proxy credential expired")
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	candidates, err := svc.listSchedulableAccounts(ctx, nil, PlatformOpenAI)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(&candidates[0], "gpt-5"), "a lagging snapshot intentionally clears account-level runtime block")
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
	hydrated, err := svc.hydrateSelectedAccount(ctx, &candidates[0])
	require.NoError(t, err)
	require.Nil(t, hydrated.TempUnschedulableUntil, "hydration uses the account cache and can also lag")
	require.Nil(t, svc.recheckSelectedOpenAIAccountFromDBBeforeProfit(ctx, hydrated, nil, PlatformOpenAI, "gpt-5", false, ""), "persisted active cooldown still rejects the DB recheck")
	repo.account.TempUnschedulableUntil = nil
	require.NotNil(t, svc.recheckSelectedOpenAIAccountFromDBBeforeProfit(ctx, hydrated, nil, PlatformOpenAI, "gpt-5", false, ""))
}

func TestRuntimeCooldownPersistenceFailureFailsOpen(t *testing.T) {
	account := &Account{ID: 94, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	svc := &OpenAIGatewayService{accountRepo: &cooldownReconciliationRepo{account: account, writeErr: errors.New("database write failed")}}
	svc.tempUnscheduleOpenAITransportError(context.Background(), account, "proxy refused")
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5"))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestRuntimeCooldownActivePersistedFieldsRetainBlock(t *testing.T) {
	for _, field := range []string{"temporary", "rate_limit", "overload"} {
		t.Run(field, func(t *testing.T) {
			until := time.Now().Add(time.Minute)
			account := &Account{ID: 95, Platform: PlatformOpenAI}
			switch field {
			case "temporary":
				account.TempUnschedulableUntil = &until
			case "rate_limit":
				account.RateLimitResetAt = &until
			case "overload":
				account.OverloadUntil = &until
			}
			svc := &OpenAIGatewayService{}
			svc.BlockAccountScheduling(account, until, "persisted")
			require.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5"))
		})
	}
}
