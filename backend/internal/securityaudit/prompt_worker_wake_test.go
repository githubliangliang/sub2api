package securityaudit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnqueuerNotifiesOnlyAfterPublish(t *testing.T) {
	cases := []struct {
		name       string
		publishErr error
		want       int
	}{
		{name: "published", want: 1},
		{name: "publish failed", publishErr: errors.New("db down"), want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeJobRepository{createJob: &Job{ID: 41}, publishErr: tc.publishErr}
			enqueuer := NewEnqueuer(&fakeConfigStore{cfg: asyncConfig(), active: true}, repo, &fakePayloadStore{values: map[int64]string{}})
			notified := 0
			enqueuer.notify = func() { notified++ }
			_ = enqueuer.Enqueue(context.Background(), asyncRequest())
			require.Equal(t, tc.want, notified)
		})
	}
}

type startingPayloadStore struct {
	PayloadStore
	entered chan struct{}
	release chan struct{}
}

func (s *startingPayloadStore) Ping(context.Context) error {
	close(s.entered)
	<-s.release
	return nil
}

func TestRunnerShutdownWaitsForStartingWorkers(t *testing.T) {
	payload := &startingPayloadStore{PayloadStore: &fakePayloadStore{values: map[int64]string{}}, entered: make(chan struct{}), release: make(chan struct{})}
	runner := NewRunner(&fakeConfigStore{cfg: asyncConfig(), active: true}, &fakeJobRepository{}, payload, PromptScannerFunc(func(context.Context, ActiveEndpoint, string, []string) (*NormalizedResult, error) {
		return integrationResult(EventPass), nil
	}), NewAtomicMetrics())
	started := make(chan error, 1)
	go func() { started <- runner.Start(t.Context()) }()
	<-payload.entered
	stopped := make(chan error, 1)
	go func() { stopped <- runner.Shutdown(t.Context()) }()
	select {
	case err := <-stopped:
		close(payload.release)
		<-started
		t.Fatalf("Shutdown returned before starting workers were registered: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(payload.release)
	require.NoError(t, <-started)
	require.NoError(t, <-stopped)
	runner.mu.Lock()
	require.Nil(t, runner.cancel)
	runner.mu.Unlock()
}

func TestRunnerConcurrentLifecycleAndCoalescedWake(t *testing.T) {
	runner := NewRunner(&fakeConfigStore{cfg: asyncConfig(), active: true}, &fakeJobRepository{}, &fakePayloadStore{values: map[int64]string{}}, PromptScannerFunc(func(context.Context, ActiveEndpoint, string, []string) (*NormalizedResult, error) {
		return integrationResult(EventPass), nil
	}), NewAtomicMetrics())
	for round := 0; round < 4; round++ {
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				require.NoError(t, runner.Start(t.Context()))
				for n := 0; n < 100; n++ {
					runner.Notify()
				}
				require.NoError(t, runner.Shutdown(t.Context()))
			}()
		}
		wg.Wait()
	}
	require.NoError(t, runner.Shutdown(t.Context()))
}

func TestEnqueuerAdmissionAndPayloadFailureNeverWake(t *testing.T) {
	for _, failure := range []error{ErrQueueFull, ErrQueueAdmissionBusy, errors.New("database unavailable")} {
		enqueuer := NewEnqueuer(&fakeConfigStore{cfg: asyncConfig(), active: true}, &fakeJobRepository{createErr: failure}, &fakePayloadStore{})
		notified := false
		enqueuer.notify = func() { notified = true }
		require.ErrorIs(t, enqueuer.Enqueue(t.Context(), asyncRequest()), failure)
		require.False(t, notified)
	}
	enqueuer := NewEnqueuer(&fakeConfigStore{cfg: asyncConfig(), active: true}, &fakeJobRepository{createJob: &Job{ID: 41}}, &fakePayloadStore{setErr: errors.New("payload down")})
	notified := false
	enqueuer.notify = func() { notified = true }
	require.Error(t, enqueuer.Enqueue(t.Context(), asyncRequest()))
	require.False(t, notified)
}

// Both cases must finish before the first fallback tick (workerIdlePoll), so they
// exercise the start-up drain and the wake path rather than polling.
func TestWorkerClaimsWithoutWaitingForIdlePoll(t *testing.T) {
	newRunner := func(repo *fakeJobRepository) *Runner {
		payload := &fakePayloadStore{values: map[int64]string{51: "abc"}}
		return NewRunner(&fakeConfigStore{cfg: asyncConfig(), active: true}, repo, payload, PromptScannerFunc(func(context.Context, ActiveEndpoint, string, []string) (*NormalizedResult, error) {
			return integrationResult(EventPass), nil
		}), NewAtomicMetrics())
	}
	completed := func(repo *fakeJobRepository) func() bool {
		return func() bool { repo.mu.Lock(); defer repo.mu.Unlock(); return repo.completeCount == 1 }
	}
	stop := func(t *testing.T, runner *Runner) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, runner.Shutdown(ctx))
	}

	t.Run("jobs left from a previous run are drained at start", func(t *testing.T) {
		repo := &fakeJobRepository{claimQueue: []*Job{workerJob(1, 3)}}
		runner := newRunner(repo)
		require.NoError(t, runner.Start(context.Background()))
		defer stop(t, runner)
		require.Eventually(t, completed(repo), workerIdlePoll/2, 5*time.Millisecond)
	})

	t.Run("notify wakes an idle worker", func(t *testing.T) {
		repo := &fakeJobRepository{}
		runner := newRunner(repo)
		require.NoError(t, runner.Start(context.Background()))
		defer stop(t, runner)
		// wait for the start-up drain to find an empty queue
		require.Eventually(t, func() bool { repo.mu.Lock(); defer repo.mu.Unlock(); return repo.claims > 0 }, workerIdlePoll/2, time.Millisecond)
		repo.mu.Lock()
		repo.claimQueue = append(repo.claimQueue, workerJob(1, 3))
		repo.mu.Unlock()
		runner.Notify()
		require.Eventually(t, completed(repo), workerIdlePoll/2, 5*time.Millisecond)
	})

	t.Run("notify never blocks", func(t *testing.T) {
		runner := newRunner(&fakeJobRepository{})
		for i := 0; i < 3; i++ {
			runner.Notify()
		}
		var nilRunner *Runner
		nilRunner.Notify()
	})
}
