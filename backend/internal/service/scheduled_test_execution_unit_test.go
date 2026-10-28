//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type scheduledExecutionPlans struct {
	ScheduledTestPlanRepository
	plan       *ScheduledTestPlan
	updated    bool
	last, next time.Time
}

func (r *scheduledExecutionPlans) ListDue(context.Context, time.Time) ([]*ScheduledTestPlan, error) {
	return []*ScheduledTestPlan{r.plan}, nil
}

func (r *scheduledExecutionPlans) UpdateAfterRun(_ context.Context, _ int64, last, next time.Time) error {
	r.updated, r.last, r.next = true, last, next
	return nil
}

type scheduledExecutionResults struct {
	ScheduledTestResultRepository
	results   []*ScheduledTestResult
	keep      int
	createErr error
}

func (r *scheduledExecutionResults) Create(_ context.Context, value *ScheduledTestResult) (*ScheduledTestResult, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	copy := *value
	r.results = append(r.results, &copy)
	return &copy, nil
}

func (r *scheduledExecutionResults) PruneOldResults(_ context.Context, _ int64, keep int) error {
	r.keep = keep
	return nil
}

type scheduledExecutionHTTP struct {
	HTTPUpstream
	client *http.Client
}

func (u *scheduledExecutionHTTP) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.client.Do(req)
}

func TestScheduledTestWorkerExecutionAndRecovery(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		saveError bool
	}{
		{"success", http.StatusOK, false},
		{"upstream failure", http.StatusInternalServerError, false},
		{"result save failure", http.StatusOK, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				require.Equal(t, "/v1/messages", r.URL.Path)
				require.Equal(t, "fixture-key", r.Header.Get("x-api-key"))
				w.WriteHeader(tc.status)
				body := `{"error":{"message":"fixture upstream unavailable"}}`
				if tc.status == http.StatusOK {
					body = "data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"scheduled-ok\"}}\n\ndata: {\"type\":\"message_stop\"}\n\n"
				}
				_, err := io.WriteString(w, body)
				require.NoError(t, err)
			}))
			t.Cleanup(server.Close)
			accountRepo := &rateLimitClearRepoStub{getByIDAccount: &Account{
				ID: 42, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusError,
				Credentials: map[string]any{"api_key": "fixture-key", "base_url": server.URL},
			}}
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			accountTest := NewAccountTestService(accountRepo, nil, nil, nil, nil,
				&scheduledExecutionHTTP{client: server.Client()}, cfg, nil)
			plans := &scheduledExecutionPlans{plan: &ScheduledTestPlan{
				ID: 1, AccountID: 42, ModelID: "claude-sonnet-4-5", CronExpression: "* * * * *", Enabled: true, AutoRecover: true, MaxResults: 2,
			}}
			results := &scheduledExecutionResults{}
			if tc.saveError {
				results.createErr = errors.New("fixture storage unavailable")
			}
			worker := NewScheduledTestRunnerService(plans, NewScheduledTestService(plans, results), accountTest,
				NewRateLimitService(accountRepo, nil, cfg, nil, nil), cfg)
			t.Cleanup(worker.Stop)
			worker.runDuePlans(context.Background())
			require.EqualValues(t, 1, requests.Load())
			if tc.saveError {
				require.False(t, plans.updated, "failed result storage must not advance the plan")
				require.Zero(t, accountRepo.clearErrorCalls, "recovery requires a saved successful result")
				return
			}
			require.Len(t, results.results, 1)
			require.EqualValues(t, 1, results.results[0].PlanID)
			require.Equal(t, 2, results.keep)
			require.True(t, plans.updated)
			require.True(t, plans.next.After(plans.last))
			if tc.status == http.StatusOK {
				require.Equal(t, "success", results.results[0].Status)
				require.Equal(t, "scheduled-ok", results.results[0].ResponseText)
				require.Equal(t, 1, accountRepo.clearErrorCalls)
			} else {
				require.Equal(t, "failed", results.results[0].Status)
				require.Contains(t, results.results[0].ErrorMessage, "API returned 500")
				require.Zero(t, accountRepo.clearErrorCalls)
			}
		})
	}
}

func TestScheduledTestWorkerSkipsOverlappingTick(t *testing.T) {
	worker := NewScheduledTestRunnerService(&emptyScheduledPlans{}, nil, nil, nil, nil)
	t.Cleanup(worker.Stop)
	firstDone := make(chan struct{})
	go func() { defer close(firstDone); worker.runScheduled() }()
	require.Eventually(t, worker.running.Load, time.Second, time.Millisecond)
	secondDone := make(chan struct{})
	go func() { defer close(secondDone); worker.runScheduled() }()
	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Fatal("overlapping scheduled tick was not skipped")
	}
	worker.Stop()
	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("active scheduled tick did not stop")
	}
}
