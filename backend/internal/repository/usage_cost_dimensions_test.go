package repository

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUsageLogRepositoryGetModelStatsAccountFilterKeepsActualCost(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	accountID := int64(42)

	mock.ExpectQuery("COALESCE\\(SUM\\(actual_cost\\), 0\\) as actual_cost,\\s+COALESCE\\(SUM\\(COALESCE\\(account_stats_cost, total_cost\\) \\* COALESCE\\(account_rate_multiplier, 1\\)\\), 0\\) as account_cost").
		WithArgs(start, end, accountID).
		WillReturnRows(sqlmock.NewRows([]string{
			"model", "requests", "input_tokens", "output_tokens",
			"cache_creation_tokens", "cache_read_tokens", "total_tokens",
			"cost", "actual_cost", "account_cost",
		}).AddRow("gemini-3.8-flash", int64(3), int64(10), int64(20), int64(0), int64(0), int64(30), 6.0, 2.4, 0.9))

	results, err := repo.GetModelStatsWithUsageFiltersBySource(context.Background(), start, end, usagestats.UsageLogFilters{AccountID: accountID}, usagestats.ModelSourceRequested)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, 2.4, results[0].ActualCost)
	require.Equal(t, 0.9, results[0].AccountCost)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryGetEndpointStatsAccountFilterKeepsActualCost(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	accountID := int64(42)

	mock.ExpectQuery("COALESCE\\(SUM\\(actual_cost\\), 0\\) as actual_cost,\\s+COALESCE\\(SUM\\(COALESCE\\(account_stats_cost, total_cost\\) \\* COALESCE\\(account_rate_multiplier, 1\\)\\), 0\\) as account_cost").
		WithArgs(start, end, accountID).
		WillReturnRows(sqlmock.NewRows([]string{"endpoint", "requests", "total_tokens", "cost", "actual_cost", "account_cost"}).
			AddRow("/v1beta/models", int64(3), int64(30), 6.0, 2.4, 0.9))

	results, err := repo.GetEndpointStatsWithFilters(context.Background(), start, end, 0, 0, accountID, 0, "", nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, 2.4, results[0].ActualCost)
	require.Equal(t, 0.9, results[0].AccountCost)
	require.NoError(t, mock.ExpectationsWereMet())
}
