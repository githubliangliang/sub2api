package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestIsQueryTimeoutErr(t *testing.T) {
	if !isQueryTimeoutErr(context.DeadlineExceeded) {
		t.Fatalf("context.DeadlineExceeded should be treated as query timeout")
	}
	if !isQueryTimeoutErr(fmt.Errorf("wrapped: %w", context.DeadlineExceeded)) {
		t.Fatalf("wrapped context.DeadlineExceeded should be treated as query timeout")
	}
	if isQueryTimeoutErr(context.Canceled) {
		t.Fatalf("context.Canceled should not be treated as query timeout")
	}
	if isQueryTimeoutErr(fmt.Errorf("wrapped: %w", context.Canceled)) {
		t.Fatalf("wrapped context.Canceled should not be treated as query timeout")
	}
}

func TestOpsOutputTPSOverviewFailurePolicy(t *testing.T) {
	for _, tc := range []struct {
		name      string
		err       error
		delay     time.Duration
		caller    bool
		wantError bool
	}{
		{name: "own timeout keeps overview", err: context.DeadlineExceeded},
		{name: "cancellation propagates", err: context.Canceled, wantError: true},
		{name: "database errors propagate", err: errors.New("database unavailable"), wantError: true},
		{name: "real two second query budget", delay: 3 * time.Second},
		{name: "caller deadline propagates", delay: time.Second, caller: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newSQLMock(t)
			repo := &opsRepository{db: db}
			// Existing raw overview queries precede the independent rate query.
			for _, count := range []int{2, 13, 6, 2, 6, 2} {
				columns, values := make([]string, count), make([]driver.Value, count)
				for i := range columns {
					columns[i], values[i] = fmt.Sprint(i), int64(0)
				}
				mock.ExpectQuery("SELECT|WITH").WillReturnRows(sqlmock.NewRows(columns).AddRow(values...))
			}
			query := mock.ExpectQuery("WITH rates AS").WillDelayFor(tc.delay)
			if tc.err != nil {
				query.WillReturnError(tc.err)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"p5", "p10", "p50", "avg", "sample_count"}).AddRow(nil, nil, nil, nil, 0))
			}
			ctx := context.Background()
			if tc.caller {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}
			start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			overview, err := repo.GetDashboardOverview(ctx, &service.OpsDashboardFilter{StartTime: start, EndTime: start.Add(time.Hour), QueryMode: service.OpsQueryModeRaw})
			if tc.wantError {
				want := tc.err
				if tc.caller {
					want = context.DeadlineExceeded
				}
				require.ErrorIs(t, err, want)
				require.Nil(t, overview)
			} else {
				require.NoError(t, err)
				require.NotNil(t, overview)
				require.Nil(t, overview.OutputTPS)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
