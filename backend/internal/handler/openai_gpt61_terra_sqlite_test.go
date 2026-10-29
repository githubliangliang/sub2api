//go:build unit

package handler

import (
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
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT61TerraAsyncUsageSQLiteSettlementAndDedup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Migrate the full schema once, then give each protocol an independent copy.
	// Repeating every SQLite ALTER TABLE under -race dominates the protocol check.
	templatePath := filepath.Join(t.TempDir(), "migrated.db")
	templateDB, err := sql.Open("sqlite", "file:"+templatePath+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = templateDB.Close() })
	require.NoError(t, repository.ApplyMigrations(context.Background(), templateDB))
	require.NoError(t, repository.EnsureSQLiteAuxTables(context.Background(), templateDB))
	require.NoError(t, templateDB.Close())
	template, err := os.ReadFile(templatePath)
	require.NoError(t, err)
	for _, protocol := range []string{"native-json", "native-sse", "chat-buffered", "chat-sse", "messages-buffered", "messages-sse"} {
		for _, effort := range []string{"none", "max"} {
			t.Run(protocol+"/"+effort, func(t *testing.T) {
				dbPath := filepath.Join(t.TempDir(), "usage.db")
				require.NoError(t, os.WriteFile(dbPath, template, 0600))
				db, err := sql.Open("sqlite", "file:"+dbPath+"?_time_format=sqlite&_pragma=foreign_keys(1)")
				require.NoError(t, err)
				t.Cleanup(func() { _ = db.Close() })
				db.SetMaxOpenConns(1)
				for _, statement := range []string{
					`INSERT INTO users(id,email,password_hash,role,balance) VALUES(1,'terra@example.test','hash','user',10)`,
					`INSERT INTO api_keys(id,user_id,key,name,quota,status) VALUES(2,1,'synthetic-terra-key','terra',1,'active')`,
					`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(3,'terra','openai','apikey','{}','{}')`,
				} {
					_, err := db.Exec(statement)
					require.NoError(t, err)
				}
				cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false, AllowInsecureHTTP: true}}}
				const response = `{"id":"resp_terra_sqlite","model":"gpt-6.1-terra","status":"completed","service_tier":"priority","output":[{"id":"fc_terra","type":"function_call","call_id":"call_lookup","name":"lookup","arguments":"{}"}],"usage":{"input_tokens":1000,"output_tokens":10,"input_tokens_details":{"cached_tokens":200,"cache_write_tokens":300}}}`
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					payload, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					require.Equal(t, "/v1/responses", r.URL.Path)
					require.Equal(t, "gpt-6.1-terra", gjson.GetBytes(payload, "model").String())
					require.Equal(t, effort, gjson.GetBytes(payload, "reasoning.effort").String())
					require.Equal(t, "lookup", gjson.GetBytes(payload, "tools.0.name").String())
					w.Header().Set("X-Request-Id", "terra-"+protocol+"-"+effort)
					if protocol == "native-json" {
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprint(w, response)
						return
					}
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":"+response+"}\n\ndata: [DONE]\n\n")
				}))
				t.Cleanup(upstream.Close)
				client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
				t.Cleanup(func() { _ = client.Close() })
				accountRepo := repository.NewAccountRepository(client, db, nil)
				svc := service.NewOpenAIGatewayService(accountRepo, repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, nil, grokAuditHTTP{upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, nil, nil, nil, nil, nil, nil)
				account := &service.Account{ID: 3, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "synthetic-terra-key", "base_url": upstream.URL, "model_mapping": map[string]any{"public-terra": "gpt-6.1-terra"}}}
				stream := strings.HasSuffix(protocol, "sse")
				body := fmt.Sprintf(`{"model":"public-terra","input":"hi","reasoning":{"effort":%q},"service_tier":"priority","stream":%t,"tools":[{"type":"function","name":"lookup","parameters":{"type":"object"}}]}`, effort, stream)
				if strings.HasPrefix(protocol, "chat") {
					body = fmt.Sprintf(`{"model":"public-terra","messages":[{"role":"user","content":"hi"}],"reasoning_effort":%q,"service_tier":"priority","stream":%t,"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"}}}]}`, effort, stream)
				} else if strings.HasPrefix(protocol, "messages") {
					body = fmt.Sprintf(`{"model":"public-terra","max_tokens":128,"messages":[{"role":"user","content":"hi"}],"output_config":{"effort":%q},"service_tier":"priority","stream":%t,"tools":[{"name":"lookup","input_schema":{"type":"object"}}]}`, effort, stream)
				}
				parent, cancelParent := context.WithCancel(context.WithValue(context.Background(), ctxkey.ClientRequestID, "terra-client-"+protocol+"-"+effort))
				t.Cleanup(cancelParent)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body)).WithContext(parent)
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 2})
				var result *service.OpenAIForwardResult
				if strings.HasPrefix(protocol, "chat") {
					result, err = svc.ForwardAsChatCompletions(parent, c, account, []byte(body), "", "")
				} else if strings.HasPrefix(protocol, "messages") {
					result, err = svc.ForwardAsAnthropic(parent, c, account, []byte(body), "", "")
				} else {
					result, err = svc.Forward(parent, c, account, []byte(body))
				}
				require.NoError(t, err)
				require.NotNil(t, result.ReasoningEffort)
				require.Equal(t, effort, *result.ReasoningEffort)
				require.Equal(t, "gpt-6.1-terra", result.UpstreamResponseModel)
				user := &service.User{ID: 1, Status: service.StatusActive, Balance: 10}
				key := &service.APIKey{ID: 2, UserID: 1, User: user, Key: "synthetic-terra-key", Quota: 1, Status: service.StatusActive}
				pool := service.NewUsageRecordWorkerPoolWithOptions(service.UsageRecordWorkerPoolOptions{WorkerCount: 1, QueueSize: 2, TaskTimeout: 5 * time.Second, AutoScaleEnabled: false})
				t.Cleanup(pool.Stop)
				entered, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				unblock := func() { once.Do(func() { close(release) }) }
				t.Cleanup(unblock)
				pool.Submit(func(ctx context.Context) {
					close(entered)
					select {
					case <-release:
					case <-ctx.Done():
					}
				})
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("usage worker did not start")
				}
				h := &OpenAIGatewayHandler{gatewayService: svc, usageRecordWorkerPool: pool}
				done := make(chan error, 1)
				input := &service.OpenAIRecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}, BillingModelSource: "response_model"}
				h.submitOpenAIUsageRecordTask(parent, result, func(ctx context.Context) { done <- svc.RecordUsage(ctx, input) })
				require.EqualValues(t, 1, pool.Stats().WaitingTasks)
				cancelParent()
				c.Request = httptest.NewRequest(http.MethodPost, "/other-request", nil)
				c.Set("upstream_response_model_observer", "other-request-model")
				unblock()
				select {
				case err := <-done:
					require.NoError(t, err)
				case <-time.After(5 * time.Second):
					t.Fatal("queued usage did not finish")
				}
				var requested, sent, observed, recordedEffort string
				var inputTokens, outputTokens, cacheRead, cacheWrite int
				var cost, balance, quotaUsed float64
				require.NoError(t, db.QueryRow(`SELECT model,upstream_model,upstream_response_model,reasoning_effort,input_tokens,output_tokens,cache_read_tokens,cache_creation_tokens,actual_cost FROM usage_logs`).Scan(&requested, &sent, &observed, &recordedEffort, &inputTokens, &outputTokens, &cacheRead, &cacheWrite, &cost))
				require.Equal(t, "public-terra", requested)
				require.Equal(t, "gpt-6.1-terra", sent)
				require.Equal(t, "gpt-6.1-terra", observed)
				require.Equal(t, effort, recordedEffort)
				require.Equal(t, 500, inputTokens)
				require.Equal(t, 10, outputTokens)
				require.Equal(t, 200, cacheRead)
				require.Equal(t, 300, cacheWrite)
				require.InDelta(t, 0.00187, cost, 1e-12)
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
				require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quotaUsed))
				require.InDelta(t, 10-cost, balance, 1e-12)
				require.InDelta(t, cost, quotaUsed, 1e-12)
				require.NoError(t, svc.RecordUsage(context.WithoutCancel(parent), input))
				var afterBalance, afterQuota float64
				var dedupCount int
				require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&afterBalance))
				require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&afterQuota))
				require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedupCount))
				require.Equal(t, balance, afterBalance)
				require.Equal(t, quotaUsed, afterQuota)
				require.Equal(t, 1, dedupCount)
			})
		}
	}
}
