package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpsTokenStatsScopeSQLite(t *testing.T) {
	db := openOpsUsageSQLite(t)
	start := time.Date(2026, 11, 25, 1, 0, 0, 0, time.UTC)
	seedOpsUsageSQLite(t, db, start)
	_, err := db.Exec(`DELETE FROM usage_logs;
INSERT INTO groups(id,name,platform) VALUES(9002,'Claude','anthropic'),(9003,'Grok','grok'),(9004,'Legacy','');
INSERT INTO accounts(id,name,platform,type,credentials) VALUES(9002,'Claude','anthropic','apikey','{}'),(9003,'Grok','grok','apikey','{}');`)
	require.NoError(t, err)
	for _, row := range []struct {
		model    string
		group    any
		account  int
		at       time.Time
		output   int
		duration any
	}{
		{"gpt-shared", 9001, 9002, start, 100, 1000},
		{"gpt-shared", 9001, 9002, start.Add(time.Second), 100, 10000},
		{"o3", 9001, 9001, start, 10, 1000},
		{"claude-sonnet", 9002, 9001, start, 20, 1000},
		{"grok-4", 9003, 9001, start, 30, 1000},
		{"o3-ungrouped", nil, 9001, start, 40, 1000},
		{"claude-legacy", 9004, 9002, start, 50, 1000},
		{"gpt-shared", 9002, 9002, start, 9, 3000},
		{"zero-duration", nil, 9001, start, 7, 0},
		{"null-duration", nil, 9001, start, 8, nil},
		{"before-window", 9001, 9001, start.Add(-time.Nanosecond), 10, 1000},
		{"at-window-end", 9001, 9001, start.Add(time.Hour), 10, 1000},
	} {
		_, err := db.Exec(`INSERT INTO usage_logs(user_id,api_key_id,account_id,group_id,model,output_tokens,duration_ms,created_at,total_cost,actual_cost) VALUES(9001,9001,$1,$2,$3,$4,$5,$6,0.1,0.04)`, row.account, row.group, row.model, row.output, row.duration, row.at)
		require.NoError(t, err)
	}
	group := func(id int64) *int64 { return &id }
	repo := &opsRepository{db: db}
	for _, tc := range []struct {
		name     string
		platform string
		group    *int64
		models   []string
	}{
		{"all", "", nil, []string{"gpt-shared", "claude-legacy", "claude-sonnet", "grok-4", "null-duration", "o3", "o3-ungrouped", "zero-duration"}},
		{"openai", " OPENAI ", nil, []string{"gpt-shared", "null-duration", "o3", "o3-ungrouped", "zero-duration"}},
		{"anthropic", "anthropic", nil, []string{"claude-legacy", "claude-sonnet", "gpt-shared"}},
		{"grok", "grok", nil, []string{"grok-4"}},
		{"group wins over account", "openai", group(9001), []string{"gpt-shared", "o3"}},
		{"crossed scope is empty", "grok", group(9001), []string{}},
		{"group without platform falls back", "anthropic", group(9004), []string{"claude-legacy"}},
		{"group only", "", group(9002), []string{"claude-sonnet", "gpt-shared"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filter := &service.OpsOpenAITokenStatsFilter{TimeRange: "1h", StartTime: start, EndTime: start.Add(time.Hour), Platform: tc.platform, GroupID: tc.group, TopN: 20}
			result, err := repo.GetOpenAITokenStats(context.Background(), filter)
			require.NoError(t, err)
			require.EqualValues(t, len(tc.models), result.Total)
			models := make([]string, 0, len(result.Items))
			for _, item := range result.Items {
				models = append(models, item.Model)
				if item.Model == "zero-duration" || item.Model == "null-duration" {
					require.Nil(t, item.AvgTokensPerSec)
				}
			}
			require.Equal(t, tc.models, models)
			if tc.name == "group wins over account" {
				require.InDelta(t, 18.18, *result.Items[0].AvgTokensPerSec, 0.001)
				require.EqualValues(t, 2, result.Items[0].RequestCount)
			}
			filter.TopN = 1
			top, err := repo.GetOpenAITokenStats(context.Background(), filter)
			require.NoError(t, err)
			require.Equal(t, result.Total, top.Total)
			require.LessOrEqual(t, len(top.Items), 1)
			filter.TopN, filter.PageSize = 0, 2
			paged := make([]string, 0)
			for page := 1; page <= len(tc.models)/2+1; page++ {
				filter.Page = page
				part, err := repo.GetOpenAITokenStats(context.Background(), filter)
				require.NoError(t, err)
				require.Equal(t, result.Total, part.Total)
				for _, item := range part.Items {
					paged = append(paged, item.Model)
				}
			}
			require.Equal(t, tc.models, paged)
			filter.StartTime = filter.EndTime
			empty, err := repo.GetOpenAITokenStats(context.Background(), filter)
			require.NoError(t, err)
			require.Zero(t, empty.Total)
			require.Empty(t, empty.Items)
		})
	}
	var rows int
	var userCost, accountCost float64
	require.NoError(t, db.QueryRow(`SELECT COUNT(*),SUM(total_cost),SUM(actual_cost) FROM usage_logs`).Scan(&rows, &userCost, &accountCost))
	require.Equal(t, 12, rows)
	require.InDelta(t, 1.2, userCost, 1e-9)
	require.InDelta(t, 0.48, accountCost, 1e-9)
}
