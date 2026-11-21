package repository

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestOpsTokenRateDurationWeightedSQLite(t *testing.T) {
	db := openOpsUsageSQLite(t)
	start := time.Date(2026, 11, 21, 1, 0, 0, 0, time.UTC)
	seedOpsUsageSQLite(t, db, start)
	_, err := db.Exec(`DELETE FROM usage_logs; INSERT INTO usage_logs(user_id,api_key_id,account_id,request_id,model,output_tokens,duration_ms,created_at) VALUES (9001,9001,9001,'a','gpt-test',100,1000,$1),(9001,9001,9001,'b','gpt-test',100,10000,$1)`, start)
	require.NoError(t, err)
	r, err := (&opsRepository{db: db}).GetOpenAITokenStats(context.Background(), &service.OpsOpenAITokenStatsFilter{StartTime: start.Add(-time.Second), EndTime: start.Add(time.Hour), TopN: 10})
	require.NoError(t, err)
	require.Len(t, r.Items, 1)
	require.InDelta(t, 18.18, *r.Items[0].AvgTokensPerSec, 0.01)
}
