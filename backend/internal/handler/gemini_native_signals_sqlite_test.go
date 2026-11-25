//go:build unit

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestGeminiNativeSignalsHTTPAndSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	template := tier223SQLiteTemplate(t)
	usage := `"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}`
	content := `{"candidates":[{"content":{"parts":[{"text":"hello"}]},"finishReason":"STOP"}],` + usage + `}`
	policy := `{"candidates":[{"finishReason":"SAFETY"}],` + usage + `}`
	envelope := `{"error":{"code":503,"status":"UNAVAILABLE","message":"fixture unavailable"}}`
	sse := func(body string) string { return "data: " + body + "\n\n" }
	for _, tc := range []struct {
		name, body                         string
		stream, oauth, skip, prior, simple bool
		status, input, output              int
	}{
		{name: "JSON error with usage", body: `{"error":{"code":503,"status":"UNAVAILABLE","message":"fixture unavailable"},` + usage + `}`, status: 503, input: 10, output: 2},
		{name: "SSE error after output", body: sse(content) + sse(envelope), stream: true, status: 503, input: 10, output: 2},
		{name: "highest priority once", body: sse(policy) + sse(envelope) + sse(envelope), stream: true, status: 503, input: 10, output: 2},
		{name: "JSON content policy", body: policy, status: 400, input: 10, output: 2},
		{name: "SSE content policy", body: sse(policy), stream: true, status: 400, input: 10, output: 2},
		{name: "JSON prompt policy", body: `{"promptFeedback":{"blockReason":"SAFETY"},` + usage + `}`, status: 400, input: 10, output: 2},
		{name: "JSON empty candidates retains usage", body: `{"candidates":[],"usageMetadata":{"promptTokenCount":10}}`, status: 502, input: 10},
		{name: "SSE empty", body: ": keepalive\n\ndata: [DONE]\n\n", stream: true, status: 502},
		{name: "array picks error", body: `[` + policy + `,` + envelope + `]`, status: 503},
		{name: "OAuth streamed error", body: sse(`{"response":`+content+`}`) + sse(`{"response":`+envelope+`}`), stream: true, oauth: true, status: 503, input: 10, output: 2},
		{name: "OAuth aggregate observes discarded error", body: sse(`{"response":`+content+`}`) + sse(`{"response":`+envelope+`}`), oauth: true, status: 503, input: 10, output: 2},
		{name: "healthy STOP", body: content, input: 10, output: 2},
		{name: "skip matches original body without detail logging", body: `{"error":{"code":503,"status":"UNAVAILABLE","message":"fixture unavailable","padding":"` + strings.Repeat("x", 2500) + `","reason":"skip-this-envelope"},` + usage + `}`, skip: true, status: 503, input: 10, output: 2},
		{name: "policy after recovered attempt", body: policy, prior: true, status: 400, input: 10, output: 2},
		{name: "policy independent of earlier skip", body: policy, prior: true, skip: true, status: 400, input: 10, output: 2},
		{name: "simple mode retains record without billing", body: policy, simple: true, status: 400, input: 10, output: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetOpsErrorLoggerStateForTest(t)
			t.Cleanup(func() { StopOpsErrorLogWorkers() })
			db, client := tier223SQLiteLedger(t, template, service.PlatformGemini)
			_, err := db.Exec(`UPDATE accounts SET type='apikey',credentials='{"api_key":"fixture"}',extra='{"quota_limit":100,"quota_used":0}' WHERE id=3`)
			require.NoError(t, err)
			repo := repository.NewAccountRepository(client, db, nil)
			account, err := repo.GetByID(context.Background(), 3)
			require.NoError(t, err)
			accountRate := 0.5
			account.RateMultiplier = &accountRate
			if tc.oauth {
				account.Type = service.AccountTypeOAuth
				account.Credentials = map[string]any{"access_token": "fixture-token", "project_id": "fixture-project", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339)}
			}
			var calls atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				contentType := "application/json"
				if tc.stream || tc.oauth {
					contentType = "text/event-stream"
				}
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("X-Request-Id", "native-signals-fixture")
				_, _ = io.WriteString(w, tc.body)
			}))
			t.Cleanup(upstream.Close)
			target, err := url.Parse(upstream.URL)
			require.NoError(t, err)
			cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Ops: config.OpsConfig{Enabled: true}}
			if tc.simple {
				cfg.RunMode = config.RunModeSimple
			}
			cfg.Pricing.DataDir = t.TempDir()
			cfg.Pricing.UpdateIntervalHours = 24
			require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"), []byte(`{"gemini-2.5-flash":{"input_cost_per_token":0.000002,"output_cost_per_token":0.000003}}`), 0600))
			pricing := service.NewPricingService(cfg, nil)
			require.NoError(t, pricing.Initialize())
			t.Cleanup(pricing.Stop)
			redisServer := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			tokens := repository.NewGeminiTokenCache(rdb)
			if tc.oauth {
				require.NoError(t, tokens.SetAccessToken(context.Background(), service.GeminiTokenCacheKey(account), "fixture-token", time.Hour))
			}
			transport := tier223LocalHTTP{upstream.Client(), target}
			native := service.NewGeminiMessagesCompatService(repo, nil, nil, nil, service.NewGeminiTokenProvider(repo, tokens, nil), service.NewRateLimitService(repo, nil, cfg, nil, nil), transport, nil, cfg)
			gateway := tier223ClaudeGateway(repo, db, cfg, transport, nil, pricing)
			opsRepo := repository.NewOpsRepository(db)
			ops := service.NewOpsService(opsRepo, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
			var rules *service.ErrorPassthroughService
			if tc.skip && !tc.prior {
				ruleRepo := repository.NewErrorPassthroughRepository(client)
				_, err := ruleRepo.Create(context.Background(), &model.ErrorPassthroughRule{Name: "signal skip", Enabled: true, ErrorCodes: []int{503}, Keywords: []string{"skip-this-envelope"}, MatchMode: model.MatchModeAll, Platforms: []string{service.PlatformGemini}, PassthroughCode: true, PassthroughBody: true, SkipMonitoring: true})
				require.NoError(t, err)
				rules = service.NewErrorPassthroughService(ruleRepo, nil)
			}
			user := &service.User{ID: 1, Balance: 10, Status: service.StatusActive}
			groupID := int64(4)
			key := &service.APIKey{ID: 2, UserID: 1, User: user, Quota: 1, GroupID: &groupID, Group: &service.Group{ID: 4, Platform: service.PlatformGemini, Hydrated: true, RateMultiplier: 2}}
			forwardErrors := make(chan error, 1)
			router := gin.New()
			router.Use(OpsErrorLoggerMiddleware(ops))
			router.POST("/v1beta/models/*modelAction", func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.ClientRequestID, "native-signals-client"))
				c.Set(string(middleware.ContextKeyAPIKey), key)
				setOpsRequestContext(c, "gemini-2.5-flash", tc.stream)
				setOpsSelectedAccount(c, 3, service.PlatformGemini)
				service.BindErrorPassthroughService(c, rules)
				if tc.prior {
					c.Set(service.OpsUpstreamErrorsKey, []*service.OpsUpstreamErrorEvent{{Platform: service.PlatformGemini, AccountID: 3, UpstreamStatusCode: 503, Kind: "retry", Message: "earlier attempt"}})
					c.Set(service.OpsSkipPassthroughKey, tc.skip)
				}
				action := "generateContent"
				if tc.stream {
					action = "streamGenerateContent"
				}
				result, err := native.ForwardNative(c.Request.Context(), c, account, "gemini-2.5-flash", action, tc.stream, []byte(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`))
				if err == nil {
					if result.Usage.InputTokens != tc.input || result.Usage.OutputTokens != tc.output {
						t.Errorf("usage = %+v, want input=%d output=%d", result.Usage, tc.input, tc.output)
					}
					input := &service.RecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}}
					err = gateway.RecordUsage(c.Request.Context(), input)
					if err == nil {
						err = gateway.RecordUsage(c.Request.Context(), input)
					}
				}
				forwardErrors <- err
			})
			server := httptest.NewServer(router)
			t.Cleanup(server.Close)
			resp, err := server.Client().Post(server.URL+"/v1beta/models/gemini-2.5-flash:generateContent", "application/json", strings.NewReader(`{}`))
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.NoError(t, <-forwardErrors)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			if tc.oauth && !tc.stream {
				require.JSONEq(t, content, string(body), "aggregation keeps its pre-existing output selection")
			} else if tc.oauth {
				require.Equal(t, sse(content)+"\n"+sse(envelope)+"\n", string(body), "OAuth keeps its pre-existing unwrap and SSE framing")
			} else {
				require.Equal(t, tc.body, string(body))
			}
			require.EqualValues(t, 1, calls.Load(), "in-band outcomes must not retry")
			require.True(t, StopOpsErrorLogWorkers(), "drain the real Ops queue into SQLite")
			var logs int
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM ops_error_logs`).Scan(&logs))
			wantLogs := 1
			if tc.status == 0 || (tc.skip && !tc.prior) {
				wantLogs = 0
			} else if tc.prior && !tc.skip {
				wantLogs = 2
			}
			require.Equal(t, wantLogs, logs)
			if wantLogs > 0 {
				var status int
				var owner string
				var limited, stream bool
				var upstreamStatus sql.NullInt64
				require.NoError(t, db.QueryRow(`SELECT status_code,error_owner,is_business_limited,stream,upstream_status_code FROM ops_error_logs WHERE status_code>=400`).Scan(&status, &owner, &limited, &stream, &upstreamStatus))
				require.Equal(t, tc.status, status)
				require.Equal(t, tc.stream, stream)
				if tc.status == 400 {
					require.Equal(t, "client", owner)
					require.True(t, limited)
					require.False(t, upstreamStatus.Valid, "policy must not inherit earlier upstream attribution")
				} else {
					require.Equal(t, "provider", owner)
					require.False(t, limited)
					require.EqualValues(t, tc.status, upstreamStatus.Int64)
				}
			}
			overview, err := opsRepo.GetDashboardOverview(context.Background(), &service.OpsDashboardFilter{StartTime: time.Now().Add(-time.Hour), EndTime: time.Now().Add(time.Minute)})
			require.NoError(t, err)
			wantSLA := int64(0)
			if wantLogs > 0 && tc.status >= 500 {
				wantSLA = 1
			}
			require.Equal(t, wantSLA, overview.ErrorCountSLA)
			var balance, quota, accountCost, totalCost, actualCost float64
			var dedup int
			require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
			require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quota))
			require.NoError(t, db.QueryRow(`SELECT json_extract(extra,'$.quota_used') FROM accounts WHERE id=3`).Scan(&accountCost))
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
			require.NoError(t, db.QueryRow(`SELECT total_cost,actual_cost FROM usage_logs LIMIT 1`).Scan(&totalCost, &actualCost))
			standard := float64(tc.input)*0.000002 + float64(tc.output)*0.000003
			require.InDelta(t, standard, totalCost, 1e-10)
			require.InDelta(t, standard*2, actualCost, 1e-10)
			charged := standard * 2
			if tc.simple {
				charged = 0
				require.Zero(t, dedup)
			} else {
				require.Equal(t, 1, dedup)
			}
			require.InDelta(t, 10-charged, balance, 1e-10)
			require.InDelta(t, charged, quota, 1e-10)
			wantAccount := standard * 0.5
			if tc.simple || tc.oauth {
				wantAccount = 0
			}
			require.InDelta(t, wantAccount, accountCost, 1e-10)
			persisted, err := repo.GetByID(context.Background(), 3)
			require.NoError(t, err)
			require.Equal(t, service.StatusActive, persisted.Status)
			require.True(t, persisted.Schedulable)
			encoded, err := json.Marshal(persisted.Extra)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "cooldown")
		})
	}
}
