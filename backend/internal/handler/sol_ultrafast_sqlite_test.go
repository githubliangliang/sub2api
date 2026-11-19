//go:build unit

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSolUltrafastHTTPAndWSSettlementSQLite(t *testing.T) {
	ctx := context.Background()
	cutoff := time.Date(2027, time.March, 1, 8, 0, 0, 0, time.UTC)
	templatePath := filepath.Join(t.TempDir(), "template.db")
	templateDB, err := sql.Open("sqlite", "file:"+templatePath+"?_time_format=sqlite")
	require.NoError(t, err)
	require.NoError(t, repository.ApplyMigrations(ctx, templateDB))
	require.NoError(t, repository.EnsureSQLiteAuxTables(ctx, templateDB))
	require.NoError(t, templateDB.Close())
	template, err := os.ReadFile(templatePath)
	require.NoError(t, err)
	for _, transport := range []string{"http", "ws"} {
		for _, model := range []struct {
			name, alias string
			standard    float64
			pricingAt   time.Time
			ultrafast   float64
		}{
			{"gpt-6.1-sol", "public-sol", 0.00187, time.Time{}, 6},
			{"gpt-6-astra", "public-astra", 0.00945, time.Time{}, 6},
			{"gpt-6.2-sol", "public-62-sol", 0.00187, cutoff.Add(-time.Nanosecond), 3},
			{"gpt-6.2-astra", "public-62-astra", 0.00748, cutoff.Add(-time.Nanosecond), 3},
			{"gpt-6.2-sol", "public-62-sol", 0.00187, cutoff, 6},
			{"gpt-6.2-astra", "public-62-astra", 0.00748, cutoff, 6},
		} {
			for _, scenario := range []struct {
				name, tier                         string
				free                               bool
				upstreamMultiplier, userMultiplier float64
			}{
				{"ultrafast", "ultrafast", false, model.ultrafast, model.ultrafast}, {"fast", "priority", false, 2, 2}, {"free-fast", "priority", true, 2, 1}, {"free-ultrafast-control", "ultrafast", true, model.ultrafast, model.ultrafast},
			} {
				name := transport + "/" + model.name + "/" + scenario.name
				if !model.pricingAt.IsZero() {
					name += "/" + model.pricingAt.Format(time.RFC3339Nano)
				}
				t.Run(name, func(t *testing.T) {
					dbPath := filepath.Join(t.TempDir(), "settlement.db")
					require.NoError(t, os.WriteFile(dbPath, template, 0600))
					db, err := sql.Open("sqlite", "file:"+dbPath+"?_time_format=sqlite&_pragma=foreign_keys(1)")
					require.NoError(t, err)
					t.Cleanup(func() { _ = db.Close() })
					db.SetMaxOpenConns(1)
					for _, statement := range []string{
						`INSERT INTO users(id,email,password_hash,role,balance) VALUES(1,'sol@example.test','hash','user',10)`,
						`INSERT INTO api_keys(id,user_id,key,name,quota,status) VALUES(2,1,'sol-fixture-key','sol',1,'active')`,
						`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(3,'sol','openai','apikey','{}','{}')`,
					} {
						_, err := db.Exec(statement)
						require.NoError(t, err)
					}
					_, err = db.Exec(`INSERT INTO groups(id,name,platform,free_openai_fast) VALUES(4,'fast','openai',?)`, scenario.free)
					require.NoError(t, err)
					response := map[string]any{
						"id": "resp_ultra", "model": model.name, "status": "completed", "service_tier": scenario.tier,
						"output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "answer"}}}},
						"usage":  map[string]any{"input_tokens": 1000, "output_tokens": 10, "input_tokens_details": map[string]any{"cached_tokens": 200, "cache_write_tokens": 300}},
					}
					encoded, err := json.Marshal(response)
					require.NoError(t, err)
					captured := make(chan []byte, 1)
					upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
					upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if transport == "ws" {
							conn, err := upgrader.Upgrade(w, r, nil)
							if err != nil {
								t.Error(err)
								return
							}
							defer conn.Close()
							_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
							_, payload, err := conn.ReadMessage()
							if err != nil {
								t.Error(err)
								return
							}
							captured <- payload
							_ = conn.WriteJSON(map[string]any{"type": "response.created", "response": map[string]any{"id": "resp_ultra", "model": model.name}})
							_ = conn.WriteJSON(map[string]any{"type": "response.completed", "response": response})
							return
						}
						payload, err := io.ReadAll(r.Body)
						if err != nil {
							t.Error(err)
							return
						}
						captured <- payload
						w.Header().Set("Content-Type", "text/event-stream")
						w.Header().Set("X-Request-Id", "ultra-http")
						_, _ = fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":%s}\n\n", encoded)
					}))
					t.Cleanup(upstream.Close)
					cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
					if transport == "ws" {
						cfg.Gateway.OpenAIWS = config.GatewayOpenAIWSConfig{Enabled: true, APIKeyEnabled: true, ResponsesWebsocketsV2: true, MaxConnsPerAccount: 1, QueueLimitPerConn: 2, DialTimeoutSeconds: 3, ReadTimeoutSeconds: 5, WriteTimeoutSeconds: 3}
					}
					client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
					t.Cleanup(func() { _ = client.Close() })
					deferred := service.NewDeferredService(nil, nil, time.Hour)
					svc := service.NewOpenAIGatewayService(repository.NewAccountRepository(client, db, nil), repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, nil, grokAuditHTTP{upstream.Client()}, deferred, nil, nil, nil, nil, nil, nil, nil)
					t.Cleanup(svc.CloseOpenAIWSPool)
					account := &service.Account{ID: 3, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Concurrency: 1,
						Credentials: map[string]any{"api_key": "sol-fixture-key", "base_url": upstream.URL, "model_mapping": map[string]any{model.alias: model.name}}, Extra: map[string]any{"responses_websockets_v2_enabled": transport == "ws"}}
					body := fmt.Sprintf(`{"model":%q,"input":"hi","stream":false,"service_tier":%q,"reasoning":{"effort":"low"}}`, model.alias, scenario.tier)
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
					c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 2})
					result, err := svc.Forward(ctx, c, account, []byte(body))
					require.NoError(t, err)
					require.Equal(t, transport == "ws", result.OpenAIWSMode)
					require.NotNil(t, result.ServiceTier)
					require.Equal(t, scenario.tier, *result.ServiceTier)
					select {
					case payload := <-captured:
						require.Equal(t, model.name, gjson.GetBytes(payload, "model").String())
						require.Equal(t, scenario.tier, gjson.GetBytes(payload, "service_tier").String())
					case <-time.After(time.Second):
						t.Fatal("upstream did not receive the request")
					}
					user := &service.User{ID: 1, Status: service.StatusActive, Balance: 10}
					groupID := int64(4)
					key := &service.APIKey{ID: 2, UserID: 1, User: user, Quota: 1, Status: service.StatusActive, GroupID: &groupID, Group: &service.Group{ID: 4, Platform: service.PlatformOpenAI, Hydrated: true, RateMultiplier: 1, FreeOpenAIFast: scenario.free}}
					input := &service.OpenAIRecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}, BillingModelSource: "response_model", PricingAt: model.pricingAt}
					require.NoError(t, svc.RecordUsage(ctx, input))
					var cost, total, balance, quota float64
					var tier string
					require.NoError(t, db.QueryRow(`SELECT actual_cost,total_cost,service_tier FROM usage_logs`).Scan(&cost, &total, &tier))
					require.InDelta(t, model.standard*scenario.upstreamMultiplier, total, 1e-12)
					require.InDelta(t, model.standard*scenario.userMultiplier, cost, 1e-12)
					require.Equal(t, scenario.tier, tier)
					require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
					require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quota))
					require.InDelta(t, 10-cost, balance, 1e-12)
					require.InDelta(t, cost, quota, 1e-12)
					require.NoError(t, svc.RecordUsage(ctx, input))
					var afterBalance, afterQuota float64
					var dedup int
					require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&afterBalance))
					require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&afterQuota))
					require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&dedup))
					require.Equal(t, balance, afterBalance)
					require.Equal(t, quota, afterQuota)
					require.Equal(t, 1, dedup)
				})
			}
		}
	}
}
