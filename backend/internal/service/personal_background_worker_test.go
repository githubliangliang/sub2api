package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type expiryWorkerRepository struct {
	AccountRepository
	calls chan struct{}
}

func (r *expiryWorkerRepository) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	r.calls <- struct{}{}
	return 0, nil
}

type emptyScheduledPlans struct {
	ScheduledTestPlanRepository
	calls atomic.Int64
}

func (r *emptyScheduledPlans) ListDue(context.Context, time.Time) ([]*ScheduledTestPlan, error) {
	r.calls.Add(1)
	return nil, nil
}

func TestAccountExpiryWorkerSQLiteStarts(t *testing.T) {
	repo := &expiryWorkerRepository{calls: make(chan struct{}, 4)}
	cfg := &config.Config{}
	cfg.Database.Driver = "sqlite"
	worker := ProvideAccountExpiryService(repo, cfg)
	t.Cleanup(worker.Stop)
	select {
	case <-repo.calls:
	case <-time.After(time.Second):
		t.Fatal("SQLite account-expiry worker was not started")
	}
}

func TestAccountExpiryWorkerStartsOnce(t *testing.T) {
	repo := &expiryWorkerRepository{calls: make(chan struct{}, 4)}
	worker := NewAccountExpiryService(repo, time.Hour)
	t.Cleanup(worker.Stop)
	worker.Start()
	select {
	case <-repo.calls:
	case <-time.After(time.Second):
		t.Fatal("expiry worker did not start")
	}
	worker.Start()
	select {
	case <-repo.calls:
		t.Fatal("repeated Start created another expiry worker")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestScheduledTestWorkerSQLiteStarts(t *testing.T) {
	plans := &emptyScheduledPlans{}
	cfg := &config.Config{}
	cfg.Database.Driver = "sqlite"
	cfg.Timezone = "Asia/Shanghai"
	worker := ProvideScheduledTestRunnerService(plans, NewScheduledTestService(plans, nil), &AccountTestService{}, nil, cfg)
	t.Cleanup(worker.Stop)
	require.NotNil(t, worker.cron, "SQLite scheduled-test runner was not started")
	require.Len(t, worker.cron.Entries(), 1)
	worker.Start()
	require.Len(t, worker.cron.Entries(), 1)
}

func TestScheduledTestWorkerStopCancelsDelayedTick(t *testing.T) {
	plans := &emptyScheduledPlans{}
	worker := NewScheduledTestRunnerService(plans, nil, nil, nil, nil)
	done := make(chan struct{})
	go func() { defer close(done); worker.runScheduled() }()
	t.Cleanup(func() {
		worker.Stop()
		select {
		case <-done:
		case <-time.After(11 * time.Second):
			t.Error("scheduled tick did not terminate")
		}
	})
	time.Sleep(10 * time.Millisecond)
	worker.Stop()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop did not cancel the scheduled tick's delay")
	}
	require.Zero(t, plans.calls.Load())
}
