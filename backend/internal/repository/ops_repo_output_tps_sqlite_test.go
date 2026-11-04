package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpsOutputTPSDistributionSQLite(t *testing.T) {
	db := openOpsUsageSQLite(t)
	ctx := context.Background()
	start := time.Date(2026, 1, 17, 0, 30, 0, 0, time.UTC)
	end := start.Add(6 * time.Hour)
	seedOpsUsageSQLite(t, db, start)
	_, err := db.ExecContext(ctx, `DELETE FROM usage_logs;
UPDATE accounts SET platform = 'anthropic' WHERE id = 9001;
INSERT INTO groups (id, name, platform) VALUES (9002, 'Other group', 'openai');`)
	require.NoError(t, err)
	requestNumber := 0
	insert := func(at time.Time, groupID any, output int, duration any, kind service.RequestType, images, imageTokens int, billing any) {
		t.Helper()
		requestNumber++
		_, err := db.ExecContext(ctx, `INSERT INTO usage_logs
(user_id, api_key_id, account_id, group_id, request_id, model, input_tokens, output_tokens,
 duration_ms, first_token_ms, request_type, image_count, image_output_tokens, billing_mode, created_at)
VALUES (9001,9001,9001,$1,$2,'gpt-test',100000,$3,$4,400,$5,$6,$7,$8,$9)`,
			groupID, fmt.Sprintf("output-tps-%d", requestNumber), output, duration, kind, images, imageTokens, billing, at)
		require.NoError(t, err)
	}
	// Unequal durations distinguish a mean of individual rates from a ratio of
	// sums; the head, full hours and tail must all contribute to every mode.
	insert(start, 9001, 1, 1000, service.RequestTypeSync, 0, 0, "token")
	insert(start.Add(time.Hour), 9001, 20, 2000, service.RequestTypeStream, 0, 0, nil)
	insert(start.Add(2*time.Hour), 9001, 100, 1000, service.RequestTypeWSV2, 0, 0, "")
	insert(end.Add(-time.Minute), 9001, 500, 500, service.RequestTypeCyberBlocked, 0, 0, "token")
	for _, duration := range []any{nil, 0, -1} {
		insert(start, 9001, 9999, duration, service.RequestTypeSync, 0, 0, "token")
	}
	for _, output := range []int{0, -1} {
		insert(start, 9001, output, 1000, service.RequestTypeSync, 0, 0, "token")
	}
	for _, kind := range []service.RequestType{service.RequestTypeLive, service.RequestTypeUnknown, service.RequestType(99)} {
		insert(start, 9001, 9999, 1000, kind, 0, 0, "token")
	}
	insert(start, 9001, 9999, 1000, service.RequestTypeSync, 1, 0, "token")
	insert(start, 9001, 9999, 1000, service.RequestTypeSync, 0, 1, "token")
	insert(start, 9001, 9999, 1000, service.RequestTypeSync, 0, 0, "image")
	insert(start.Add(-time.Millisecond), 9001, 9999, 1000, service.RequestTypeSync, 0, 0, "token")
	insert(end, 9001, 9999, 1000, service.RequestTypeSync, 0, 0, "token")
	insert(start, 9002, 9999, 1000, service.RequestTypeSync, 0, 0, "token")
	insert(start, nil, 50, 1000, service.RequestTypeSync, 0, 0, "token")

	repo := &opsRepository{db: db}
	groupID := int64(9001)
	filter := &service.OpsDashboardFilter{StartTime: start.In(time.FixedZone("UTC+8", 8*3600)), EndTime: end,
		Platform: " OpenAI ", GroupID: &groupID}
	assertStats := func(stats *service.OpsOutputTPS) {
		t.Helper()
		require.NotNil(t, stats)
		require.EqualValues(t, 4, stats.SampleCount)
		require.NotNil(t, stats.P5)
		require.NotNil(t, stats.P10)
		require.NotNil(t, stats.P50)
		require.NotNil(t, stats.Avg)
		require.InDelta(t, 2.35, *stats.P5, 1e-9)
		require.InDelta(t, 3.7, *stats.P10, 1e-9)
		require.InDelta(t, 55, *stats.P50, 1e-9)
		require.InDelta(t, 277.75, *stats.Avg, 1e-9)
	}
	for _, mode := range []service.OpsQueryMode{service.OpsQueryModeRaw, service.OpsQueryModeAuto} {
		filter.QueryMode = mode
		overview, err := repo.GetDashboardOverview(ctx, filter)
		require.NoError(t, err)
		assertStats(overview.OutputTPS)
		encoded, err := json.Marshal(overview)
		require.NoError(t, err)
		require.Contains(t, string(encoded), `"output_tps":{"p5":2.35,`)
	}
	require.NoError(t, repo.UpsertHourlyMetrics(ctx, start.Truncate(time.Hour), end))
	for _, mode := range []service.OpsQueryMode{service.OpsQueryModePreagg, service.OpsQueryModeAuto} {
		filter.QueryMode = mode
		overview, err := repo.GetDashboardOverview(ctx, filter)
		require.NoError(t, err)
		assertStats(overview.OutputTPS)
	}
	filter.Platform = service.PlatformAnthropic
	empty, err := repo.queryOutputTPS(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, &service.OpsOutputTPS{}, empty)
	encoded, err := json.Marshal(empty)
	require.NoError(t, err)
	require.JSONEq(t, `{"p5":null,"p10":null,"p50":null,"avg":null,"sample_count":0}`, string(encoded))
	// Group platform wins over the account's platform; ungrouped rows fall back.
	filter.GroupID = nil
	fallback, err := repo.queryOutputTPS(ctx, filter)
	require.NoError(t, err)
	require.EqualValues(t, 1, fallback.SampleCount)
	require.Equal(t, 50.0, *fallback.P50)

	// Keeping preaggregated history must not recreate cleaned-up rate samples.
	_, err = db.ExecContext(ctx, "DELETE FROM usage_logs WHERE group_id = 9001")
	require.NoError(t, err)
	filter.GroupID, filter.Platform, filter.QueryMode = &groupID, "openai", service.OpsQueryModePreagg
	overview, err := repo.GetDashboardOverview(ctx, filter)
	require.NoError(t, err)
	require.Positive(t, overview.SuccessCount)
	require.Equal(t, &service.OpsOutputTPS{}, overview.OutputTPS)
}

func TestOpsOutputTPSContinuousPercentilesSQLite(t *testing.T) {
	db := openOpsUsageSQLite(t)
	start := time.Date(2026, 1, 17, 0, 0, 0, 0, time.UTC)
	seedOpsUsageSQLite(t, db, start)
	repo := &opsRepository{db: db}
	for _, tc := range []struct {
		name  string
		rates []int
		p5    float64
		p10   float64
		p50   float64
		avg   float64
	}{
		{"single", []int{3}, 3, 3, 3, 3},
		{"two reversed", []int{100, 1}, 5.95, 10.9, 50.5, 50.5},
		{"odd with duplicates", []int{100, 10, 1, 10, 10}, 2.8, 4.6, 10, 26.2},
		{"integer ranks", []int{21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, 2, 3, 11, 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec("DELETE FROM usage_logs")
			require.NoError(t, err)
			for _, rate := range tc.rates {
				_, err := db.Exec(`INSERT INTO usage_logs (user_id, api_key_id, account_id, model,
output_tokens, duration_ms, request_type, created_at) VALUES (9001,9001,9001,'gpt-test',$1,1000,$2,$3)`,
					rate, service.RequestTypeSync, start)
				require.NoError(t, err)
			}
			stats, err := repo.queryOutputTPS(context.Background(), &service.OpsDashboardFilter{StartTime: start, EndTime: start.Add(time.Hour)})
			require.NoError(t, err)
			require.EqualValues(t, len(tc.rates), stats.SampleCount)
			require.InDelta(t, tc.p5, *stats.P5, 1e-9)
			require.InDelta(t, tc.p10, *stats.P10, 1e-9)
			require.InDelta(t, tc.p50, *stats.P50, 1e-9)
			require.InDelta(t, tc.avg, *stats.Avg, 1e-9)
		})
	}
}

func TestOpsOutputTPSQueryInterruptSQLite(t *testing.T) {
	// A large recursive view forces the real driver to execute long enough to
	// observe interruption without preparing millions of persistent rows.
	// A cancellation may discard the driver's connection, so use a file whose
	// schema survives reconnects rather than a connection-local :memory: DB.
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "interrupt.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE VIEW usage_logs AS
WITH RECURSIVE numbers(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM numbers WHERE n < 10000000)
SELECT n AS output_tokens, 1000 AS duration_ms, 1 AS request_type,
0 AS image_count, 0 AS image_output_tokens, 'token' AS billing_mode,
printf('2026-01-17 00:30:%02d+00:00', n % 60) AS created_at FROM numbers`)
	require.NoError(t, err)
	repo := &opsRepository{db: db}
	start := time.Date(2026, 1, 17, 0, 0, 0, 0, time.UTC)
	filter := &service.OpsDashboardFilter{StartTime: start, EndTime: start.Add(time.Hour)}
	for _, cancelCaller := range []bool{false, true} {
		t.Run(fmt.Sprint("cancel=", cancelCaller), func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc
			want := context.DeadlineExceeded
			if cancelCaller {
				ctx, cancel = context.WithCancel(context.Background())
				timer := time.AfterFunc(20*time.Millisecond, cancel)
				defer timer.Stop()
				want = context.Canceled
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
			}
			defer cancel()
			stats, err := repo.queryOutputTPS(ctx, filter)
			require.ErrorIs(t, err, want)
			require.Nil(t, stats)
		})
	}
	// The interrupted connection must remain usable for the next request.
	var one int
	require.NoError(t, db.QueryRow("SELECT 1").Scan(&one))
	require.Equal(t, 1, one)
}
