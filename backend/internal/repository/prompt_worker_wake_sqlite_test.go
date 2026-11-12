package repository

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	"github.com/stretchr/testify/require"
)

type sqliteWakeConfig struct {
	securityaudit.ConfigStore
	cfg securityaudit.ActiveConfig
}

func (c sqliteWakeConfig) Active() (securityaudit.ActiveConfig, bool) { return c.cfg, true }

type sqliteWakePayload struct {
	mu     sync.Mutex
	values map[int64]string
}

func (p *sqliteWakePayload) Set(_ context.Context, id int64, value string, _ time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.values[id] = value
	return nil
}
func (p *sqliteWakePayload) Get(_ context.Context, id int64) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.values[id], nil
}
func (p *sqliteWakePayload) Delete(_ context.Context, id int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.values, id)
	return nil
}
func (p *sqliteWakePayload) Ping(context.Context) error { return nil }

type observedSQLiteWakeRepo struct {
	securityaudit.JobRepository
	claims atomic.Int64
}

func (r *observedSQLiteWakeRepo) ClaimNextJob(ctx context.Context, now time.Time) (*securityaudit.Job, bool, error) {
	r.claims.Add(1)
	return r.JobRepository.ClaimNextJob(ctx, now)
}

type sqliteWakeScanner func(context.Context, securityaudit.ActiveEndpoint, string, []string) (*securityaudit.NormalizedResult, error)

func (f sqliteWakeScanner) Scan(ctx context.Context, e securityaudit.ActiveEndpoint, s string, ids []string) (*securityaudit.NormalizedResult, error) {
	return f(ctx, e, s, ids)
}

func TestPromptWorkerWakeAndRetryRunOnSQLite(t *testing.T) {
	registerSQLitePGCompatFunctions()
	db := openUsageStatsSQLite(t)
	repo := &observedSQLiteWakeRepo{JobRepository: securityaudit.NewPostgreSQLRepository(db)}
	payload := &sqliteWakePayload{values: map[int64]string{}}
	cfg := sqliteWakeConfig{cfg: securityaudit.ActiveConfig{RiskControlEnabled: true, Enabled: true, WorkerCount: 1, QueueCapacity: 8, Scanners: []string{"prompt_injection"}, Endpoints: []securityaudit.ActiveEndpoint{{ID: "fixture", Enabled: true, InputLimit: 1000}}}}
	scanned := make(chan struct{}, 2)
	runner := securityaudit.NewRunner(cfg, repo, payload, sqliteWakeScanner(func(context.Context, securityaudit.ActiveEndpoint, string, []string) (*securityaudit.NormalizedResult, error) {
		scanned <- struct{}{}
		return &securityaudit.NormalizedResult{Decision: securityaudit.EventPass}, nil
	}), securityaudit.NewAtomicMetrics())
	require.NoError(t, runner.Start(t.Context()))
	defer func() { require.NoError(t, runner.Shutdown(t.Context())) }()
	require.Eventually(t, func() bool { return repo.claims.Load() > 0 }, time.Second, time.Millisecond)
	create := func() *securityaudit.Job {
		job, err := repo.CreateStagingWithCapacity(t.Context(), securityaudit.PromptSnapshot{RequestID: "wake-fixture"}, 1, 3, 8)
		require.NoError(t, err)
		require.NoError(t, payload.Set(t.Context(), job.ID, "hello", time.Minute))
		return job
	}
	job := create()
	require.NoError(t, repo.PublishQueued(t.Context(), job.ID))
	var rawDue, normalizedDue int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM prompt_audit_jobs WHERE next_attempt_at<=?", time.Now().UTC()).Scan(&rawDue))
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM prompt_audit_jobs WHERE julianday(next_attempt_at)<=julianday(?)", time.Now().UTC()).Scan(&normalizedDue))
	t.Logf("SQLite due comparison raw=%d normalized=%d", rawDue, normalizedDue)
	started := time.Now()
	runner.Notify()
	select {
	case <-scanned:
		t.Logf("actual SQLite claim wake latency=%s", time.Since(started))
	case <-time.After(250 * time.Millisecond):
		t.Fatal("local wake waited for fallback polling")
	}
	require.Eventually(t, func() bool {
		var status string
		return db.QueryRow("SELECT status FROM prompt_audit_jobs WHERE id=?", job.ID).Scan(&status) == nil && status == "done"
	}, time.Second, time.Millisecond)
	// No Notify: a due retry must still be picked up by fallback polling.
	retry := create()
	_, err := db.Exec("UPDATE prompt_audit_jobs SET status='retry',next_attempt_at=? WHERE id=?", time.Now().Add(50*time.Millisecond).UTC(), retry.ID)
	require.NoError(t, err)
	select {
	case <-scanned:
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("fallback poll did not claim the due retry")
	}
}
