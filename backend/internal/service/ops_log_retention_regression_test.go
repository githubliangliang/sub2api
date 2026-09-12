//go:build unit

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestOpsAccessPersistenceRuntimeLifecycle(t *testing.T) {
	require.NoError(t, logger.Init(logger.InitOptions{Level: "info", Format: "json", Output: logger.OutputOptions{ToStdout: true}}))
	repo := newRuntimeSettingRepoStub()
	sink := &OpsSystemLogSink{}
	svc := &OpsService{settingRepo: repo, systemLogSink: sink}
	access := &logger.LogEvent{Level: "info", Fields: map[string]any{"component": "http.access"}}

	svc.initRuntimeSettings(context.Background())
	require.False(t, sink.shouldIndex(access), "old configurations must default access persistence off")
	for _, event := range []*logger.LogEvent{{Level: "warn", Component: "http.access"}, {Level: "error", Component: "http.access"}, {Level: "info", Component: "audit.settings"}} {
		require.True(t, sink.shouldIndex(event))
	}
	var update OpsRuntimeLogConfig
	require.NoError(t, json.Unmarshal([]byte(`{"level":"info","persist_access_logs":true,"caller":true,"stacktrace_level":"error","retention_days":7}`), &update))
	_, err := svc.UpdateRuntimeLogConfig(context.Background(), &update, 1)
	require.NoError(t, err)
	require.True(t, sink.shouldIndex(access))
	// A failed write must roll back both the logger and access persistence.
	repo.setFn = func(string, string) error { return errors.New("write failed") }
	require.NoError(t, json.Unmarshal([]byte(`{"persist_access_logs":false}`), &update))
	_, err = svc.UpdateRuntimeLogConfig(context.Background(), &update, 1)
	require.ErrorContains(t, err, "write failed")
	require.True(t, sink.shouldIndex(access))
	repo.setFn = nil
	_, err = svc.ResetRuntimeLogConfig(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, sink.shouldIndex(access))
}

func TestOpsRuntimeRetentionDoesNotInheritErrorLogWindow(t *testing.T) {
	svc := &OpsService{cfg: &config.Config{Ops: config.OpsConfig{Cleanup: config.OpsCleanupConfig{ErrorLogRetentionDays: 90}}}}
	cfg, err := svc.GetRuntimeLogConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, 30, cfg.RetentionDays)
}

func TestOpsAdvancedDefaultDoesNotDisableConfiguredCleanup(t *testing.T) {
	svc := &OpsService{cfg: &config.Config{Ops: config.OpsConfig{Cleanup: config.OpsCleanupConfig{Enabled: true}}}}
	svc.initRuntimeSettings(context.Background())
	cfg, err := svc.GetOpsAdvancedSettings(context.Background())
	require.NoError(t, err)
	require.True(t, cfg.DataRetention.CleanupEnabled)
}

type scheduledRetentionRepo struct {
	OpsRepository
	done chan *OpsUpsertJobHeartbeatInput
}

func (r *scheduledRetentionRepo) UpsertJobHeartbeat(_ context.Context, input *OpsUpsertJobHeartbeatInput) error {
	r.done <- input
	return nil
}

func TestOpsScheduledCleanupUsesIndependentSystemRetentionOnSQLite(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		db.SetMaxOpenConns(1)
		for table, column := range map[string]string{
			"ops_error_logs": "created_at", "ops_ingress_reject_aggregates": "bucket_start", "ops_alert_events": "created_at",
			"ops_system_logs": "created_at", "ops_system_log_cleanup_audits": "created_at", "ops_system_metrics": "created_at",
			"ops_metrics_hourly": "bucket_start", "ops_metrics_daily": "bucket_date",
		} {
			_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, %s DATETIME NOT NULL)", table, column))
			require.NoError(t, err)
			_, err = db.Exec(fmt.Sprintf("INSERT INTO %s VALUES (1,?), (2,?)", table), time.Now().Add(-10*24*time.Hour), time.Now().Add(-3*24*time.Hour))
			require.NoError(t, err)
		}
		settings := newRuntimeSettingRepoStub()
		settings.values[SettingKeyOpsRuntimeLogConfig] = `{"retention_days":7}`
		repo := &scheduledRetentionRepo{done: make(chan *OpsUpsertJobHeartbeatInput, 1)}
		cfg := &config.Config{RunMode: config.RunModeSimple, Ops: config.OpsConfig{Enabled: true, Cleanup: config.OpsCleanupConfig{
			Enabled: true, Schedule: "* * * * *", ErrorLogRetentionDays: 90, MinuteMetricsRetentionDays: 30, HourlyMetricsRetentionDays: 30,
		}}}
		cleanup := NewOpsCleanupService(repo, db, nil, cfg, nil, settings)
		cleanup.Start()
		t.Cleanup(cleanup.Stop)
		synctest.Wait()
		time.Sleep(time.Minute)
		synctest.Wait()
		select {
		case heartbeat := <-repo.done:
			require.Nil(t, heartbeat.LastError)
			require.NotNil(t, heartbeat.LastSuccessAt)
		default:
			t.Fatal("scheduled cleanup did not execute")
		}
		for table, want := range map[string]int{"ops_system_logs": 1, "ops_system_log_cleanup_audits": 1, "ops_error_logs": 2} {
			var count int
			require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&count))
			require.Equal(t, want, count, table)
		}
	})
}
