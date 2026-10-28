//go:build unit

package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cacheHitHTTPUpstream struct{ client *http.Client }

type cacheHitQuotaUpdater struct{}

func (cacheHitQuotaUpdater) UpdateQuotaUsed(context.Context, int64, float64) error {
	return errors.New("SQLite atomic billing must update quota in its transaction")
}

func (cacheHitQuotaUpdater) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return errors.New("SQLite atomic billing must update rate usage in its transaction")
}

func (s cacheHitHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return s.client.Do(req)
}

func (s cacheHitHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxyURL, accountID, concurrency)
}

func TestUsagePromptCacheHitSQLiteSettlement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
			ctx := context.Background()
			db := openPersonalMaintenanceSQLite(t)
			require.NoError(t, ApplyMigrations(ctx, db))
			require.NoError(t, EnsureSQLiteAuxTables(ctx, db))
			for _, statement := range []string{
				`INSERT INTO users (id,email,password_hash,role,balance) VALUES (1,'user@example.test','hash','user',10)`,
				`INSERT INTO api_keys (id,user_id,key,name,quota,quota_used,status) VALUES (2,1,'cache-hit-key','test',1,0,'active')`,
				`INSERT INTO accounts (id,name,platform,type,credentials,extra) VALUES (3,'cache-hit-account','openai','apikey','{}','{}')`,
			} {
				_, err := db.ExecContext(ctx, statement)
				require.NoError(t, err)
			}
			const response = `{"id":"resp_cache_hit","object":"response","model":"gpt-5.4","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]}],"usage":{"prompt_tokens":100,"completion_tokens":5,"prompt_cache_hit_tokens":80,"prompt_cache_miss_tokens":20}}`
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/responses" {
					http.Error(w, "unexpected endpoint", http.StatusNotFound)
					return
				}
				w.Header().Set("X-Request-Id", "cache-hit-settlement")
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":"+response+"}\n\ndata: [DONE]\n\n")
				} else {
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprint(w, response)
				}
			}))
			t.Cleanup(upstream.Close)
			cfg := &config.Config{
				Default: config.DefaultConfig{RateMultiplier: 1},
				Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
					Enabled: false, AllowInsecureHTTP: true,
				}},
			}
			billing := service.NewBillingService(cfg, nil)
			svc := service.NewOpenAIGatewayService(nil, NewUsageLogRepository(nil, db), NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, billing, nil, nil, cacheHitHTTPUpstream{client: upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, nil, nil, nil, nil, nil, nil)
			account := &service.Account{ID: 3, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test", "base_url": upstream.URL}}
			body := fmt.Sprintf(`{"model":"gpt-5.4","input":"hello","stream":%t}`, stream)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
			result, err := svc.Forward(ctx, c, account, []byte(body))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotEmpty(t, result.RequestID)
			user := &service.User{ID: 1, Role: service.RoleUser, Status: service.StatusActive, Balance: 10}
			key := &service.APIKey{ID: 2, UserID: 1, User: user, Key: "cache-hit-key", Quota: 1, Status: service.StatusActive}
			input := &service.OpenAIRecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: cacheHitQuotaUpdater{}}
			require.NoError(t, svc.RecordUsage(ctx, input))
			require.NoError(t, svc.RecordUsage(ctx, input))
			var count, dedup, ordinary, cached, output int
			var balance, quotaUsed, cost float64
			require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM usage_logs`).Scan(&count))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT input_tokens,cache_read_tokens,output_tokens,actual_cost FROM usage_logs`).Scan(&ordinary, &cached, &output, &cost))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT balance FROM users WHERE id=1`).Scan(&balance))
			require.NoError(t, db.QueryRowContext(ctx, `SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quotaUsed))
			require.Equal(t, 1, count)
			require.Equal(t, 1, dedup)
			require.Equal(t, 20, ordinary)
			require.Equal(t, 80, cached)
			require.Equal(t, 5, output)
			wantCost, err := billing.CalculateCost("gpt-5.4", service.UsageTokens{InputTokens: 20, CacheReadTokens: 80, OutputTokens: 5}, 1)
			require.NoError(t, err)
			require.InDelta(t, wantCost.ActualCost, cost, 1e-12)
			require.InDelta(t, 10-wantCost.ActualCost, balance, 1e-12)
			require.InDelta(t, wantCost.ActualCost, quotaUsed, 1e-12)
		})
	}
}
