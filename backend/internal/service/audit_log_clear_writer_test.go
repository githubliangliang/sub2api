package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type auditClearWriterRepo struct {
	AuditLogRepository
	mu      sync.Mutex
	entries []*AuditLog
	started chan struct{}
	release chan struct{}
}

func (r *auditClearWriterRepo) BatchInsert(_ context.Context, entries []*AuditLog) (int64, error) {
	if r.started != nil {
		close(r.started)
		<-r.release
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entries...)
	return int64(len(entries)), nil
}

func (r *auditClearWriterRepo) Count(context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.entries)), nil
}

func (r *auditClearWriterRepo) TruncateAll(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = nil
	return nil
}

func (r *auditClearWriterRepo) Insert(_ context.Context, entry *AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
	return nil
}

func TestAuditClearWaitsForActiveFlushAndRecordStaysNonBlocking(t *testing.T) {
	repo := &auditClearWriterRepo{started: make(chan struct{}), release: make(chan struct{})}
	svc := NewAuditLogService(repo, nil)
	svc.Start()
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(repo.release) }); svc.Stop() })
	for range auditLogBatchSize {
		svc.Record(&AuditLog{Action: "old"})
	}
	select {
	case <-repo.started:
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not start")
	}
	// A blocked database writer must not stall management requests enqueueing logs.
	recorded := make(chan struct{})
	go func() { svc.Record(&AuditLog{Action: "buffered"}); close(recorded) }()
	select {
	case <-recorded:
	case <-time.After(time.Second):
		t.Fatal("Record blocked behind the database write")
	}
	cleared := make(chan error, 1)
	go func() {
		_, err := svc.ClearAll(context.Background(), &AuditLog{})
		cleared <- err
	}()
	release.Do(func() { close(repo.release) })
	select {
	case err := <-cleared:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("clear did not complete")
	}
	// The second queued entry is in the old generation, so Stop cannot restore it.
	svc.Stop()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	require.Len(t, repo.entries, 1)
	require.Equal(t, AuditActionAuditLogClear, repo.entries[0].Action)
}
