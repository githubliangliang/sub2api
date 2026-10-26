//go:build unit

package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/pkg/websearch"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type personalSearchProvider struct {
	search func(context.Context, websearch.SearchRequest) (*websearch.SearchResponse, string, error)
}

// Scheduling refreshes settings asynchronously, including after the request
// returns. Keep the fixture safe while cleanup disables search configuration.
type personalSearchSettingRepo struct {
	service.SettingRepository
	mu     sync.RWMutex
	values map[string]string
}

func (r *personalSearchSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}

func (r *personalSearchSettingRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func (p personalSearchProvider) SearchWithBestProvider(ctx context.Context, req websearch.SearchRequest) (*websearch.SearchResponse, string, error) {
	return p.search(ctx, req)
}

type personalSearchForbiddenUpstream struct {
	service.HTTPUpstream
	calls atomic.Int32
}

func (u *personalSearchForbiddenUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	u.calls.Add(1)
	return nil, errors.New("LLM upstream must not be called for emulated search")
}

func (u *personalSearchForbiddenUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

type personalSearchFailWriter struct {
	gin.ResponseWriter
	writes int
}

func (w *personalSearchFailWriter) Write(body []byte) (int, error) {
	w.writes++
	if w.writes > 1 {
		return 0, io.ErrClosedPipe
	}
	return w.ResponseWriter.Write(body)
}

type personalSearchHandlerRun struct {
	recorder  *httptest.ResponseRecorder
	usage     *openAIWSUsageHandlerUsageLogRepoStub
	slots     *concurrencyCacheMock
	upstream  *personalSearchForbiddenUpstream
	provider  int
	cancelled bool
}

func runPersonalSearchHandler(t *testing.T, stream, queued bool, failure string, failover bool) personalSearchHandlerRun {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Default.RateMultiplier = 1
	groupID := int64(9700)
	group := &service.Group{ID: groupID, Hydrated: true, Platform: service.PlatformOpenAI, Status: service.StatusActive, AllowMessagesDispatch: true, RateMultiplier: 1}
	account := service.Account{
		ID: 9701, Name: "personal-search", Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
		Status: service.StatusActive, Schedulable: true, Concurrency: 1, Priority: 1,
		Credentials: map[string]any{"api_key": "local-fixture", "base_url": "https://unused.invalid/v1"},
		Extra: map[string]any{
			"web_search_emulation":              "enabled",
			openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions),
		},
	}
	accounts := []service.Account{account}
	if failover {
		proxyID := int64(9710)
		accounts[0].ProxyID = &proxyID
		accounts[0].Proxy = &service.Proxy{ID: proxyID, Protocol: "http", Host: "unavailable.invalid", Port: 8080}
		next := account
		next.ID, next.Priority = 9704, 2
		accounts = append(accounts, next)
	}
	run := personalSearchHandlerRun{
		recorder: httptest.NewRecorder(),
		usage:    &openAIWSUsageHandlerUsageLogRepoStub{created: make(chan *service.UsageLog, 4)},
		slots: &concurrencyCacheMock{
			acquireUserSlotFn:    func(context.Context, int64, int, string) (bool, error) { return true, nil },
			acquireAccountSlotFn: func(context.Context, int64, int, string) (bool, error) { return true, nil },
		},
		upstream: &personalSearchForbiddenUpstream{},
	}
	settings := service.NewSettingService(&personalSearchSettingRepo{}, cfg)
	require.NoError(t, settings.SaveWebSearchEmulationConfig(context.Background(), &service.WebSearchEmulationConfig{
		Enabled: true, Providers: []service.WebSearchProviderConfig{{Type: "brave", APIKey: "local-fixture"}},
	}))
	t.Cleanup(func() {
		service.SetWebSearchManager(nil)
		require.NoError(t, settings.SaveWebSearchEmulationConfig(context.Background(), &service.WebSearchEmulationConfig{}))
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.SetWebSearchManager(personalSearchProvider{search: func(searchCtx context.Context, req websearch.SearchRequest) (*websearch.SearchResponse, string, error) {
		run.provider++
		require.Equal(t, "latest search query", req.Query)
		if failure == "provider" {
			return nil, "brave", errors.New("local provider failure")
		}
		if failure == "cancel" {
			cancel()
			return nil, "brave", searchCtx.Err()
		}
		if failover && req.ProxyURL != "" {
			return nil, "brave", websearch.ErrProxyUnavailable
		}
		return &websearch.SearchResponse{Query: req.Query, Results: []websearch.SearchResult{{URL: "https://example.com/result", Title: "Local result", Snippet: "fixture search content"}}}, "brave", nil
	}})

	concurrency := service.NewConcurrencyService(run.slots)
	billing := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billing.Stop)
	gw := service.NewOpenAIGatewayService(
		&openAIImagesFailoverAccountRepo{accounts: accounts}, run.usage, nil, nil, nil, nil, nil,
		cfg, nil, concurrency, service.NewBillingService(cfg, nil), nil, billing, run.upstream,
		&service.DeferredService{}, nil, nil, nil, nil, nil, settings, nil,
	)
	h := &OpenAIGatewayHandler{
		gatewayService: gw, billingCacheService: billing,
		apiKeyService:     service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg),
		concurrencyHelper: NewConcurrencyHelper(concurrency, SSEPingFormatClaude, 0),
		cfg:               cfg, maxAccountSwitches: 1,
	}
	var releaseWorker func()
	if queued {
		pool := newUsageRecordTestPool(t)
		h.usageRecordWorkerPool = pool
		started := make(chan struct{})
		unblock := make(chan struct{})
		var released atomic.Bool
		releaseWorker = func() {
			if released.CompareAndSwap(false, true) {
				close(unblock)
			}
		}
		t.Cleanup(releaseWorker)
		require.Equal(t, service.UsageRecordSubmitModeEnqueued, pool.Submit(func(context.Context) { close(started); <-unblock }))
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("usage worker did not start")
		}
	}
	apiKey := &service.APIKey{ID: 9702, UserID: 9703, GroupID: &groupID, Group: group, User: &service.User{ID: 9703, Balance: 100}}
	router := gin.New()
	router.POST("/v1/messages", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), apiKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: apiKey.UserID, Concurrency: 1})
		if failure == "write" {
			c.Writer = &personalSearchFailWriter{ResponseWriter: c.Writer}
		}
		h.Messages(c)
		if queued {
			c.Request = httptest.NewRequest(http.MethodPost, "/next-request", nil).WithContext(context.WithValue(context.Background(), ctxkey.RequestedPublicModel, "next-request-model"))
		}
	})
	body := fmt.Sprintf(`{"model":"gpt-5.2","stream":%t,"max_tokens":16,"tools":[{"type":"web_search_20250305","name":"web_search"}],"messages":[{"role":"user","content":"older query"},{"role":"assistant","content":"previous answer"},{"role":"user","content":"latest search query"}]}`, stream)
	requestCtx := context.WithValue(ctx, ctxkey.RequestedPublicModel, "public-search-alias")
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body)).WithContext(requestCtx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "personal-search-client")
	router.ServeHTTP(run.recorder, req)
	run.cancelled = ctx.Err() != nil
	if queued {
		require.Equal(t, uint64(1), h.usageRecordWorkerPool.Stats().WaitingTasks, "successful search must enqueue one usage record")
		releaseWorker()
		h.usageRecordWorkerPool.Stop()
	}
	return run
}

func TestOpenAIWebSearchMessagesHandler_UsageAndSlots(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, queued := range []bool{false, true} {
			t.Run(fmt.Sprintf("stream=%t/queued=%t", stream, queued), func(t *testing.T) {
				run := runPersonalSearchHandler(t, stream, queued, "", false)
				require.Equal(t, http.StatusOK, run.recorder.Code, run.recorder.Body.String())
				require.Equal(t, 1, run.provider)
				require.Zero(t, run.upstream.calls.Load())
				require.Equal(t, int32(1), atomic.LoadInt32(&run.slots.releaseUserCalled))
				require.Equal(t, int32(1), atomic.LoadInt32(&run.slots.releaseAccountCalled))
				if stream {
					require.Equal(t, 1, strings.Count(run.recorder.Body.String(), "event: message_stop"))
				} else {
					require.Equal(t, "end_turn", gjson.GetBytes(run.recorder.Body.Bytes(), "stop_reason").String())
				}
				require.Len(t, run.usage.created, 1)
				usage := <-run.usage.created
				require.EqualValues(t, 9703, usage.UserID)
				require.EqualValues(t, 9702, usage.APIKeyID)
				require.EqualValues(t, 9701, usage.AccountID)
				require.NotNil(t, usage.GroupID)
				require.EqualValues(t, 9700, *usage.GroupID)
				require.Equal(t, "gpt-5.2", usage.Model)
				require.Equal(t, "public-search-alias", usage.RequestedModel, "queued work must retain the original request, even after Gin reuse")
				require.NotNil(t, usage.UserAgent)
				require.Equal(t, "personal-search-client", *usage.UserAgent)
				require.Zero(t, usage.InputTokens)
				require.Zero(t, usage.OutputTokens)
				require.Zero(t, usage.TotalCost)
				require.NotEmpty(t, usage.RequestID)
			})
		}
	}
}

func TestOpenAIWebSearchMessagesHandler_FailuresReleaseSlots(t *testing.T) {
	for _, failure := range []string{"provider", "cancel", "write"} {
		t.Run(failure, func(t *testing.T) {
			run := runPersonalSearchHandler(t, true, false, failure, false)
			require.Equal(t, 1, run.provider)
			require.Zero(t, run.upstream.calls.Load())
			require.Equal(t, int32(1), atomic.LoadInt32(&run.slots.releaseUserCalled))
			require.Equal(t, int32(1), atomic.LoadInt32(&run.slots.releaseAccountCalled))
			require.Empty(t, run.usage.created, "unsuccessful emulation must not create successful usage")
			require.NotContains(t, run.recorder.Body.String(), "event: message_stop")
			if failure == "cancel" {
				require.True(t, run.cancelled)
			}
		})
	}
}

func TestOpenAIWebSearchMessagesHandler_ProviderFailoverReleasesAttempts(t *testing.T) {
	run := runPersonalSearchHandler(t, false, false, "", true)
	require.Equal(t, http.StatusOK, run.recorder.Code, run.recorder.Body.String())
	require.Equal(t, 2, run.provider)
	require.Zero(t, run.upstream.calls.Load())
	require.Equal(t, int32(1), atomic.LoadInt32(&run.slots.releaseUserCalled))
	require.Equal(t, int32(2), atomic.LoadInt32(&run.slots.releaseAccountCalled))
	require.Len(t, run.usage.created, 1)
	require.EqualValues(t, 9704, (<-run.usage.created).AccountID, "only the successful selected account owns usage")
}
