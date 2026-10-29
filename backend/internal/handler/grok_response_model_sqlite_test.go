//go:build unit

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type grokAuditHTTP struct{ client *http.Client }

func (u grokAuditHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.client.Do(req)
}

func (u grokAuditHTTP) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

type grokAuditQuotaUpdater struct{}

func (grokAuditQuotaUpdater) UpdateQuotaUsed(context.Context, int64, float64) error {
	return errors.New("usage must settle through the SQLite billing transaction")
}

func (grokAuditQuotaUpdater) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return errors.New("usage must settle through the SQLite billing transaction")
}

func TestGrokResponseModelAuditAsyncSQLiteBillingBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"native-json", "native-sse", "chat-buffered", "messages-buffered"} {
		t.Run(protocol, func(t *testing.T) {
			db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared&_time_format=sqlite&_pragma=foreign_keys(1)")
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			db.SetMaxOpenConns(1)
			require.NoError(t, repository.ApplyMigrations(context.Background(), db))
			require.NoError(t, repository.EnsureSQLiteAuxTables(context.Background(), db))
			for _, statement := range []string{
				`INSERT INTO users(id,email,password_hash,role,balance) VALUES(1,'grok@example.test','hash','user',10)`,
				`INSERT INTO api_keys(id,user_id,key,name,quota,status) VALUES(2,1,'synthetic-key','audit',1,'active')`,
				`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(3,'audit','grok','apikey','{}','{}')`,
			} {
				_, err := db.Exec(statement)
				require.NoError(t, err)
			}
			cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Pricing: config.PricingConfig{DataDir: t.TempDir(), UpdateIntervalHours: 24}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false, AllowInsecureHTTP: true}}}
			require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"), []byte(`{"grok-4.6":{"input_cost_per_token":0.003,"output_cost_per_token":0.003},"grok-4.7":{"input_cost_per_token":0.001,"output_cost_per_token":0.001}}`), 0600))
			pricing := service.NewPricingService(cfg, nil)
			require.NoError(t, pricing.Initialize())
			t.Cleanup(pricing.Stop)
			const response = `{"id":"resp_audit","model":"grok-4.7","status":"completed","output":[{"id":"msg_audit","type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":2,"output_tokens":1}}`
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/v1/responses", r.URL.Path)
				w.Header().Set("X-Request-Id", "grok-audit-"+protocol)
				if protocol == "native-json" {
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprint(w, response)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = fmt.Fprint(w, "data: {\"type\":\"response.created\",\"response\":{\"model\":\"grok-4.5\"}}\n\ndata: {\"type\":\"response.completed\",\"response\":"+response+"}\n\ndata: [DONE]\n\n")
			}))
			t.Cleanup(upstream.Close)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
			t.Cleanup(func() { _ = client.Close() })
			accountRepo := repository.NewAccountRepository(client, db, nil)
			svc := service.NewOpenAIGatewayService(accountRepo, repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, pricing), nil, nil, grokAuditHTTP{upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, service.NewGrokTokenProvider(accountRepo, nil), nil, nil, nil, nil, nil)
			account := &service.Account{ID: 3, Platform: service.PlatformGrok, Type: service.AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "synthetic-key", "base_url": upstream.URL}}
			if protocol == "chat-buffered" {
				account.Type = service.AccountTypeOAuth
				account.Status = service.StatusActive
				account.Schedulable = true
				account.Credentials["access_token"] = "synthetic-access"
				account.Credentials["refresh_token"] = "synthetic-refresh"
				account.Credentials["expires_at"] = time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
				credentials, err := json.Marshal(account.Credentials)
				require.NoError(t, err)
				_, err = db.Exec("UPDATE accounts SET type='oauth', credentials=? WHERE id=3", string(credentials))
				require.NoError(t, err)
			}
			if protocol == "chat-buffered" {
				current, err := accountRepo.GetByID(context.Background(), account.ID)
				require.NoError(t, err)
				require.Equal(t, account.Credentials, current.Credentials)
				_, _, tokenErr := svc.GetAccessToken(context.Background(), account)
				require.NoError(t, tokenErr)
			}
			body := `{"model":"grok-4.6","input":"hi","stream":false}`
			if protocol == "native-sse" {
				body = `{"model":"grok-4.6","input":"hi","stream":true}`
			}
			if strings.Contains(protocol, "buffered") {
				body = `{"model":"grok-4.6","messages":[{"role":"user","content":"hi"}],"max_tokens":128,"stream":false}`
			}
			parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "grok-audit-"+protocol)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body)).WithContext(parent)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 2})
			var result *service.OpenAIForwardResult
			switch protocol {
			case "chat-buffered":
				result, err = svc.ForwardAsChatCompletions(parent, c, account, []byte(body), "audit-cache-identity", "")
			case "messages-buffered":
				result, err = svc.ForwardAsAnthropic(parent, c, account, []byte(body), "", "")
			default:
				result, err = svc.Forward(parent, c, account, []byte(body))
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, "grok-4.7", result.UpstreamResponseModel)
			require.Equal(t, protocol != "native-json", result.UpstreamResponseModelConflict)
			require.Equal(t, "grok-4.6", result.Model)
			require.Equal(t, "grok-4.6", result.UpstreamModel)
			user := &service.User{ID: 1, Status: service.StatusActive, Balance: 10}
			key := &service.APIKey{ID: 2, UserID: 1, User: user, Key: "synthetic-key", Quota: 1, Status: service.StatusActive}
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
			// Reusing the request context after forwarding cannot replace the queued audit.
			c.Request = httptest.NewRequest(http.MethodPost, "/other-request", nil)
			c.Set("upstream_response_model_observer", "other-request-model")
			unblock()
			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("queued usage did not finish")
			}
			var raw, sent string
			var mismatch bool
			var cost float64
			require.NoError(t, db.QueryRow(`SELECT upstream_response_model,upstream_model,upstream_model_mismatch,actual_cost FROM usage_logs`).Scan(&raw, &sent, &mismatch, &cost))
			require.Equal(t, "grok-4.7", raw)
			require.Equal(t, "grok-4.6", sent)
			require.True(t, mismatch)
			if result.UpstreamResponseModelConflict {
				require.InDelta(t, 0.009, cost, 1e-12)
			} else {
				require.InDelta(t, 0.003, cost, 1e-12)
			}
			if result.UpstreamResponseModelConflict {
				control := *result
				control.RequestID += "-no-conflict"
				control.UpstreamResponseModelConflict = false
				controlInput := *input
				controlInput.Result = &control
				require.NoError(t, svc.RecordUsage(context.Background(), &controlInput))
				var controlCost float64
				require.NoError(t, db.QueryRow(`SELECT actual_cost FROM usage_logs WHERE request_id=?`, control.RequestID).Scan(&controlCost))
				require.InDelta(t, 0.003, controlCost, 1e-12)
			}
		})
	}
}
