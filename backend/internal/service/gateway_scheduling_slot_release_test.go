//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// hydrate 失败时，已经占用的并发槽位必须在返回前释放，否则调用方拿不到
// ReleaseFunc，槽位会一直泄漏到 TTL 到期。
func TestGatewayServiceNewSelectionResultReleasesSlotOnHydrationError(t *testing.T) {
	svc := &GatewayService{
		schedulerSnapshot: &SchedulerSnapshotService{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	released := 0
	selection, err := svc.newSelectionResult(ctx, "intake", &Account{ID: 42}, true, func() { released++ }, nil)

	require.Error(t, err)
	require.Nil(t, selection)
	require.Equal(t, 1, released)
}

// hydrate 成功时不能提前释放槽位，ReleaseFunc 应原样交给调用方。
func TestGatewayServiceNewSelectionResultKeepsReleaseOnSuccess(t *testing.T) {
	account := &Account{ID: 42, Platform: PlatformAnthropic}
	cache := &snapshotHydrationCache{
		accounts: map[int64]*Account{account.ID: account},
	}
	svc := &GatewayService{
		schedulerSnapshot: NewSchedulerSnapshotService(cache, nil, nil, nil, nil),
	}

	released := 0
	selection, err := svc.newSelectionResult(context.Background(), "intake", account, true, func() { released++ }, nil)

	require.NoError(t, err)
	require.NotNil(t, selection)
	require.Zero(t, released)
	require.NotNil(t, selection.ReleaseFunc)

	selection.ReleaseFunc()
	require.Equal(t, 1, released)
}

// Use the real snapshot fallback and session release paths on all failure shapes.
type tier121HydrationRepo struct {
	AccountRepository
	err error
}

func (r tier121HydrationRepo) GetByID(context.Context, int64) (*Account, error) { return nil, r.err }
func TestGatewayServiceNewSelectionResultReleasesBothResources(t *testing.T) {
	for _, mode := range []string{"cancelled", "repository-error", "disappeared"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			repo := tier121HydrationRepo{}
			if mode == "cancelled" {
				cancel()
			}
			if mode == "repository-error" {
				repo.err = errors.New("synthetic repository failure")
			}
			sessions := newSessionLimitReleaseCacheStub()
			svc := &GatewayService{schedulerSnapshot: NewSchedulerSnapshotService(nil, nil, repo, nil, nil), sessionLimitCache: sessions}
			released := 0
			result, err := svc.newSelectionResult(ctx, "turn-session", newSessionLimitTestAccount(), true, func() { released++ }, nil)
			require.Error(t, err)
			require.Nil(t, result)
			require.Equal(t, 1, released)
			require.Equal(t, []string{"turn-session"}, sessions.unregistered[42])
			// Waiting selections own no concurrency release function.
			_, err = svc.newSelectionResult(ctx, "wait-session", newSessionLimitTestAccount(), false, nil, &AccountWaitPlan{})
			require.Error(t, err)
			require.Equal(t, 1, released)
			require.Equal(t, []string{"turn-session", "wait-session"}, sessions.unregistered[42])
		})
	}
}
