//go:build unit

package handler

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// Deliver a complete usage frame before the next read observes client cancellation.
// This keeps the cancellation boundary independent of scanner read sizes.
type geminiAbortSQLiteBody struct {
	reader *strings.Reader
	cancel context.CancelFunc
}

func (b *geminiAbortSQLiteBody) Read(p []byte) (int, error) {
	if b.reader.Len() > 0 {
		return b.reader.Read(p)
	}
	b.cancel()
	return 0, context.Canceled
}

func (*geminiAbortSQLiteBody) Close() error { return nil }

type geminiAbortSQLiteUpstream struct {
	service.HTTPUpstream
	body  io.ReadCloser
	calls int
}

func (u *geminiAbortSQLiteUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	u.calls++
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":      []string{"text/event-stream"},
			"X-Request-Id":      []string{"gemini-abort-sqlite"},
			"X-Goog-Request-Id": []string{"gemini-abort-sqlite"},
		},
		Body: u.body,
	}, nil
}

type geminiAbortSQLiteBillingRepo struct {
	compatStreamUsageBillingRepo
	canceledTasks atomic.Int32
	usageLog      *service.UsageLog
}

type geminiAbortSQLiteOptions struct {
	dropOverflow    bool
	reuseGinRequest bool
}

func (r *geminiAbortSQLiteBillingRepo) Apply(ctx context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error) {
	if ctx.Err() != nil {
		r.canceledTasks.Add(1)
	}
	return r.compatStreamUsageBillingRepo.Apply(ctx, cmd)
}

func newGeminiAbortSQLiteLedger(t *testing.T) (*sql.DB, service.UsageBillingRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared&_time_format=sqlite", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		`CREATE TABLE usage_billing_dedup (id INTEGER PRIMARY KEY AUTOINCREMENT, request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT, UNIQUE(request_id, api_key_id))`,
		`CREATE TABLE usage_billing_dedup_archive (request_id TEXT, api_key_id INTEGER, request_fingerprint TEXT)`,
		`CREATE TABLE users (id INTEGER PRIMARY KEY, balance REAL, frozen_balance REAL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE api_keys (id INTEGER PRIMARY KEY, quota REAL, quota_used REAL, status TEXT, usage_5h REAL, usage_1d REAL, usage_7d REAL, window_5h_start DATETIME, window_1d_start DATETIME, window_7d_start DATETIME, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE accounts (id INTEGER PRIMARY KEY, extra TEXT, deleted_at DATETIME, updated_at DATETIME)`,
		`INSERT INTO users (id, balance) VALUES (9403, 100)`,
		`INSERT INTO api_keys (id, quota, quota_used, status, usage_5h, usage_1d, usage_7d) VALUES (9402, 20, 0, 'active', 0, 0, 0)`,
		`INSERT INTO accounts (id, extra) VALUES (9401, '{"quota_limit":10,"quota_used":0}')`,
	} {
		_, err := db.Exec(statement)
		require.NoError(t, err)
	}
	return db, repository.NewUsageBillingRepository(nil, db)
}

func runGeminiAbortSQLiteHandler(t *testing.T, protocol string, ledger service.UsageBillingRepository, withUsage bool) *geminiAbortSQLiteBillingRepo {
	return runGeminiAbortSQLiteHandlerWithOptions(t, protocol, ledger, withUsage, geminiAbortSQLiteOptions{})
}

func runGeminiAbortSQLiteHandlerWithOptions(t *testing.T, protocol string, ledger service.UsageBillingRepository, withUsage bool, options geminiAbortSQLiteOptions) *geminiAbortSQLiteBillingRepo {
	t.Helper()
	cfg := &config.Config{Pricing: config.PricingConfig{DataDir: t.TempDir(), UpdateIntervalHours: 24}}
	require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"), []byte(
		`{"gemini-2.5-flash":{"input_cost_per_token":0.000002,"output_cost_per_token":0.000003}}`,
	), 0600))
	pricing := service.NewPricingService(cfg, nil)
	require.NoError(t, pricing.Initialize())
	t.Cleanup(pricing.Stop)
	groupID := int64(9400)
	group := &service.Group{ID: groupID, Hydrated: true, Platform: service.PlatformGemini, Status: service.StatusActive, RateMultiplier: 1}
	account := &service.Account{
		ID: 9401, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey,
		Status: service.StatusActive, Schedulable: true, Concurrency: 1,
		Credentials:   map[string]any{"api_key": "test-key"},
		Extra:         map[string]any{"quota_limit": 10.0},
		AccountGroups: []service.AccountGroup{{AccountID: 9401, GroupID: groupID}},
	}
	snapshot := service.NewSchedulerSnapshotService(&fakeSchedulerCache{accounts: []*service.Account{account}}, nil, nil, nil, nil)
	billingRepo := &geminiAbortSQLiteBillingRepo{compatStreamUsageBillingRepo: compatStreamUsageBillingRepo{UsageBillingRepository: ledger}}
	usageRepo := &openAIWSUsageHandlerUsageLogRepoStub{created: make(chan *service.UsageLog, 2)}
	billingCache := service.NewBillingCacheService(&compatStreamEligibilityCache{}, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billingCache.Stop)
	poolOptions := service.UsageRecordWorkerPoolOptions{
		WorkerCount: 1, QueueSize: 4, TaskTimeout: 5 * time.Second,
		OverflowPolicy: config.UsageRecordOverflowPolicySync, AutoScaleEnabled: false,
	}
	if options.dropOverflow {
		poolOptions.QueueSize = 1
		poolOptions.OverflowPolicy = config.UsageRecordOverflowPolicyDrop
	}
	pool := service.NewUsageRecordWorkerPoolWithOptions(poolOptions)
	t.Cleanup(pool.Stop)
	blocked, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseWorker := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseWorker)
	require.Equal(t, service.UsageRecordSubmitModeEnqueued, pool.Submit(func(ctx context.Context) {
		close(blocked)
		select {
		case <-release:
		case <-ctx.Done():
		}
	}))
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("usage worker did not start")
	}
	if options.dropOverflow {
		require.Equal(t, service.UsageRecordSubmitModeEnqueued, pool.Submit(func(context.Context) {}))
		require.Equal(t, uint64(1), pool.Stats().WaitingTasks)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ctx = context.WithValue(ctx, ctxkey.Group, group)
	ctx = context.WithValue(ctx, ctxkey.ClientRequestID, "gemini-abort-sqlite-"+protocol)
	stream := `data: {"candidates":[{"content":{"role":"model","parts":[{"text":"partial"}]}}]`
	if withUsage {
		stream += `,"usageMetadata":{"promptTokenCount":12,"candidatesTokenCount":7}`
	}
	stream += "}\n\n"
	upstream := &geminiAbortSQLiteUpstream{body: &geminiAbortSQLiteBody{reader: strings.NewReader(stream), cancel: cancel}}
	gw := service.NewGatewayService(
		nil, &fakeGroupRepo{group: group}, usageRepo, billingRepo, nil, nil, nil, nil, cfg,
		snapshot, nil, service.NewBillingService(cfg, pricing), nil, nil, nil, upstream,
		&service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	h := &GatewayHandler{
		gatewayService: gw, billingCacheService: billingCache, cfg: cfg,
		geminiCompatService:   service.NewGeminiMessagesCompatService(nil, nil, nil, nil, nil, nil, upstream, nil, cfg),
		usageRecordWorkerPool: pool, maxAccountSwitches: 1, maxAccountSwitchesGemini: 1,
		concurrencyHelper: NewConcurrencyHelper(service.NewConcurrencyService(&fakeConcurrencyCache{}), SSEPingFormatClaude, 0),
	}
	apiKey := &service.APIKey{
		ID: 9402, UserID: 9403, GroupID: &groupID, Group: group,
		User: &service.User{ID: 9403, Balance: 100},
	}
	route, path := "/v1/messages", "/v1/messages"
	body := `{"model":"gemini-2.5-flash","stream":true,"max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`
	call := h.Messages
	switch protocol {
	case "chat_completions":
		route, path, call = "/v1/chat/completions", "/v1/chat/completions", h.ChatCompletions
	case "native":
		route, path, call = "/v1beta/models/*modelAction", "/v1beta/models/gemini-2.5-flash:streamGenerateContent?alt=sse", h.GeminiV1BetaModels
		body = `{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`
	}
	router := gin.New()
	router.POST(route, func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), apiKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: apiKey.UserID})
		call(c)
		if options.reuseGinRequest {
			// Gin can reuse this request field once the gateway handler returns.
			// The worker remains blocked until after this deterministic replacement.
			nextCtx := context.WithValue(context.Background(), ctxkey.RequestedPublicModel, "next-request-model")
			c.Request = httptest.NewRequest(http.MethodPost, "/next-request", nil).WithContext(nextCtx)
		}
	})
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.ErrorIs(t, ctx.Err(), context.Canceled)
	require.Equal(t, 1, upstream.calls, "partial output must not trigger another upstream attempt")
	billingRepo.mu.Lock()
	pendingCommands := len(billingRepo.commands)
	billingRepo.mu.Unlock()
	if options.dropOverflow {
		require.Equal(t, 1, pendingCommands, "observed usage must fall back to synchronous billing when the drop queue is full")
		require.Equal(t, uint64(1), pool.Stats().DroppedQueueFull)
	} else {
		require.Zero(t, pendingCommands, "billing must wait in the real worker queue")
	}
	if withUsage && !options.dropOverflow {
		require.Equal(t, uint64(1), pool.Stats().WaitingTasks, "observed usage must be submitted after cancellation")
	} else if !options.dropOverflow {
		require.Zero(t, pool.Stats().WaitingTasks, "no observed usage must not create a billing task")
	}
	releaseWorker()
	pool.Stop()
	require.Zero(t, billingRepo.canceledTasks.Load(), "queued billing must use a context independent of the canceled request")
	if withUsage {
		require.Len(t, billingRepo.commands, 1)
		require.Len(t, usageRepo.created, 1)
		billingRepo.usageLog = <-usageRepo.created
	} else {
		require.Empty(t, billingRepo.commands)
		require.Empty(t, usageRepo.created)
	}
	return billingRepo
}

func TestGeminiStreamAbortDropQueueOverflowStillSettlesSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"messages", "chat_completions", "native"} {
		t.Run(protocol, func(t *testing.T) {
			db, ledger := newGeminiAbortSQLiteLedger(t)
			runGeminiAbortSQLiteHandlerWithOptions(t, protocol, ledger, true, geminiAbortSQLiteOptions{dropOverflow: true})
			var balance, quota float64
			var dedup int
			require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=9403`).Scan(&balance))
			require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=9401`).Scan(&quota))
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
			require.InDelta(t, 100-(12*0.000002+7*0.000003), balance, 1e-10)
			require.InDelta(t, 12*0.000002+7*0.000003, quota, 1e-10)
			require.Equal(t, 1, dedup)
		})
	}
}

func TestGeminiStreamAbortQueuedUsageCapturesRequestMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"messages", "chat_completions", "native"} {
		t.Run(protocol, func(t *testing.T) {
			_, ledger := newGeminiAbortSQLiteLedger(t)
			run := runGeminiAbortSQLiteHandlerWithOptions(t, protocol, ledger, true, geminiAbortSQLiteOptions{reuseGinRequest: true})
			require.Equal(t, "gemini-2.5-flash", run.usageLog.RequestedModel, "queued work must use the original channel metadata snapshot")
			require.NotContains(t, run.usageLog.RequestedModel, "next-request")
		})
	}
}

func TestGeminiStreamAbortUsageQueueSettlesOnceSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"messages", "chat_completions", "native"} {
		for _, withUsage := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/usage=%t", protocol, withUsage), func(t *testing.T) {
				db, ledger := newGeminiAbortSQLiteLedger(t)
				first := runGeminiAbortSQLiteHandler(t, protocol, ledger, withUsage)
				var balance, quota float64
				var dedup int
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=9403`).Scan(&balance))
				require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=9401`).Scan(&quota))
				require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
				if withUsage {
					cmd := first.commands[0]
					require.Equal(t, 12, cmd.InputTokens)
					require.Equal(t, 7, cmd.OutputTokens)
					require.Positive(t, cmd.BalanceCost)
					require.Positive(t, cmd.AccountQuotaCost)
					require.InDelta(t, 12*0.000002+7*0.000003, cmd.BalanceCost, 1e-10)
					require.InDelta(t, 12*0.000002+7*0.000003, cmd.AccountQuotaCost, 1e-10)
					require.InDelta(t, 100-cmd.BalanceCost, balance, 1e-10)
					require.InDelta(t, cmd.AccountQuotaCost, quota, 1e-10)
					require.Equal(t, 1, dedup)
				} else {
					require.Equal(t, float64(100), balance)
					require.Zero(t, quota)
					require.Zero(t, dedup)
				}
				runGeminiAbortSQLiteHandler(t, protocol, ledger, withUsage)
				var replayBalance, replayQuota float64
				var replayDedup int
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=9403`).Scan(&replayBalance))
				require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=9401`).Scan(&replayQuota))
				require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&replayDedup))
				require.Equal(t, balance, replayBalance)
				require.Equal(t, quota, replayQuota)
				require.Equal(t, dedup, replayDedup)
			})
		}
	}
}
