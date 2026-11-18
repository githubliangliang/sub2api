package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func costDimensionsSQLite(t *testing.T) (*usageLogRepository, time.Time, time.Time) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "costs.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE usage_logs (
	 user_id INTEGER,api_key_id INTEGER,account_id INTEGER,group_id INTEGER,
	 model TEXT,requested_model TEXT,upstream_model TEXT,upstream_response_model TEXT,upstream_model_mismatch BOOLEAN,
	 request_type INTEGER,stream BOOLEAN,billing_type INTEGER,billing_mode TEXT,
	 input_tokens INTEGER,output_tokens INTEGER,cache_creation_tokens INTEGER,cache_read_tokens INTEGER,
	 total_cost REAL,actual_cost REAL,account_stats_cost REAL,account_rate_multiplier REAL,
	 duration_ms INTEGER,inbound_endpoint TEXT,upstream_endpoint TEXT,created_at DATETIME)`)
	require.NoError(t, err)
	start := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	for i, row := range []struct {
		model, requested, upstream string
		cost, actual               float64
		stats                      any
		rate                       float64
	}{
		{"gemini-3.8-flash", "gemini-flash-public", "gemini-3.8-flash", 6, 2.4, 2.25, 0.4},
		{"gemini-3.1-pro", "gemini-pro-public", "gemini-3.1-pro", 10, 1, nil, 0.2},
		{"gemini-3.8-flash", "gemini-flash-free", "gemini-3.8-flash", 1, 0.7, 0, 0.4},
	} {
		_, err = db.Exec(`INSERT INTO usage_logs VALUES (1,2,7,3,$1,$2,$3,'reported-model',$4,$5,0,$6,'token',10,20,0,0,$7,$8,$9,$10,100,'/v1/messages','/v1beta/models',$11)`,
			row.model, row.requested, row.upstream, i == 0, int16(service.RequestTypeSync), int16(service.BillingTypeBalance), row.cost, row.actual, row.stats, row.rate, start.Add(time.Hour))
		require.NoError(t, err)
	}
	_, err = db.Exec(`INSERT INTO usage_logs VALUES (1,2,8,3,'other','other','other','other',0,1,0,0,'token',10,20,0,0,99,99,99,1,100,'/v1/messages','/v1beta/models',$1)`, start.Add(time.Hour))
	require.NoError(t, err)
	_, err = db.Exec(`ALTER TABLE usage_logs ADD COLUMN openai_ws_mode BOOLEAN DEFAULT 0`)
	require.NoError(t, err)
	_, err = db.Exec(`ALTER TABLE usage_logs ADD COLUMN image_count INTEGER DEFAULT 0`)
	require.NoError(t, err)
	_, err = db.Exec(`ALTER TABLE usage_logs ADD COLUMN video_count INTEGER DEFAULT 0`)
	require.NoError(t, err)
	return &usageLogRepository{sql: db}, start, start.Add(24 * time.Hour)
}

func TestUsageCostDimensionsAllFiltersSQLite(t *testing.T) {
	repo, start, end := costDimensionsSQLite(t)
	ctx := context.Background()
	requestType := int16(service.RequestTypeSync)
	stream := false
	billingType := int8(service.BillingTypeBalance)
	for _, filters := range []usagestats.UsageLogFilters{
		{AccountID: 7, StartTime: &start, EndTime: &end},
		{AccountID: 7, UserID: 1, APIKeyID: 2, GroupID: 3, StartTime: &start, EndTime: &end, RequestType: &requestType, Stream: &stream, BillingType: &billingType, BillingMode: "token"},
	} {
		stats, err := repo.GetStatsWithFilters(ctx, filters)
		require.NoError(t, err)
		require.EqualValues(t, 3, stats.TotalRequests)
		require.InDelta(t, 17, stats.TotalCost, 1e-10)
		require.InDelta(t, 4.1, stats.TotalActualCost, 1e-10)
		require.NotNil(t, stats.TotalAccountCost)
		require.InDelta(t, 2.9, *stats.TotalAccountCost, 1e-10)
		for _, rows := range [][]usagestats.EndpointStat{stats.Endpoints, stats.UpstreamEndpoints, stats.EndpointPaths} {
			require.Len(t, rows, 1)
			require.InDelta(t, 17, rows[0].Cost, 1e-10)
			require.InDelta(t, 4.1, rows[0].ActualCost, 1e-10)
			require.InDelta(t, 2.9, rows[0].AccountCost, 1e-10)
		}
		for _, source := range []string{usagestats.ModelSourceRequested, usagestats.ModelSourceUpstream, usagestats.ModelSourceMapping} {
			models, err := repo.GetModelStatsWithUsageFiltersBySource(ctx, start, end, filters, source)
			require.NoError(t, err)
			var standard, actual, account float64
			for _, model := range models {
				standard += model.Cost
				actual += model.ActualCost
				account += model.AccountCost
			}
			require.InDelta(t, 17, standard, 1e-10)
			require.InDelta(t, 4.1, actual, 1e-10)
			require.InDelta(t, 2.9, account, 1e-10)
		}
	}
	mismatch := true
	filters := usagestats.UsageLogFilters{AccountID: 7, StartTime: &start, EndTime: &end, UpstreamModelMismatch: &mismatch}
	stats, err := repo.GetStatsWithFilters(ctx, filters)
	require.NoError(t, err)
	require.EqualValues(t, 1, stats.TotalRequests)
	require.InDelta(t, 2.4, stats.TotalActualCost, 1e-10)
	require.InDelta(t, 0.9, *stats.TotalAccountCost, 1e-10)
	for _, source := range []struct{ source, model string }{
		{usagestats.ModelSourceRequested, "gemini-flash-public"},
		{usagestats.ModelSourceUpstream, "gemini-3.1-pro"},
		{usagestats.ModelSourceMapping, "gemini-flash-public -> gemini-3.8-flash"},
	} {
		filters := usagestats.UsageLogFilters{AccountID: 7, Model: source.model, ModelFilterSource: source.source}
		models, err := repo.GetModelStatsWithUsageFiltersBySource(ctx, start, end, filters, source.source)
		require.NoError(t, err)
		require.Len(t, models, 1)
		require.Equal(t, source.model, models[0].Model)
	}
	batched, err := repo.GetGeminiUsageTotalsBatch(ctx, []int64{7}, start, end)
	require.NoError(t, err)
	require.InDelta(t, 0.9, batched[7].FlashCost, 1e-10)
	require.InDelta(t, 2, batched[7].ProCost, 1e-10)
}
