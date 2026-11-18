//go:build unit

package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func tier223SQLiteTemplate(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "template.db")
	db, err := sql.Open("sqlite", "file:"+path+"?_time_format=sqlite")
	require.NoError(t, err)
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))
	require.NoError(t, repository.EnsureSQLiteAuxTables(context.Background(), db))
	require.NoError(t, db.Close())
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	return body
}

func tier223SQLiteLedger(t *testing.T, template []byte, platform string) (*sql.DB, *dbent.Client) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.db")
	require.NoError(t, os.WriteFile(path, template, 0600))
	db, err := sql.Open("sqlite", "file:"+path+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	for _, statement := range []string{
		`INSERT INTO users(id,email,password_hash,role,balance) VALUES(1,'tier223@example.test','hash','user',10)`,
		`INSERT INTO api_keys(id,user_id,key,name,quota,status) VALUES(2,1,'tier223-fixture-key','tier223',1,'active')`,
	} {
		_, err := db.Exec(statement)
		require.NoError(t, err)
	}
	_, err = db.Exec(`INSERT INTO accounts(id,name,platform,type,credentials,extra,status,schedulable) VALUES(3,'tier223',?,'oauth','{"access_token":"old-token","refresh_token":"retained-refresh"}','{}','active',1)`, platform)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO groups(id,name,platform,rate_multiplier) VALUES(4,'tier223',?,2)`, platform)
	require.NoError(t, err)
	return db, client
}

// The production request builder retains its endpoint and headers. This test
// adapter sends that request over TCP to the local protocol fixture.
type tier223LocalHTTP struct {
	client *http.Client
	target *url.URL
}

func (u tier223LocalHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	copy := req.Clone(req.Context())
	copy.URL.Scheme, copy.URL.Host = u.target.Scheme, u.target.Host
	return u.client.Do(copy)
}

func (u tier223LocalHTTP) DoWithTLS(req *http.Request, proxy string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, accountID, concurrency)
}

func tier223ClaudeGateway(repo service.AccountRepository, db *sql.DB, cfg *config.Config, upstream service.HTTPUpstream, cache service.ClaudeTokenCache, pricing *service.PricingService) *service.GatewayService {
	return service.NewGatewayService(repo, nil, repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg,
		nil, nil, service.NewBillingService(cfg, pricing), service.NewRateLimitService(repo, nil, cfg, nil, nil), nil, nil, upstream,
		&service.DeferredService{}, service.NewClaudeTokenProvider(repo, cache, nil), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func tier223ForwardClaude(ctx context.Context, c *gin.Context, svc *service.GatewayService, account *service.Account, path string, stream bool) (*service.ForwardResult, error) {
	return tier223ForwardClaudeModel(ctx, c, svc, account, path, stream, "claude-sonnet-4-6")
}

func tier223ForwardClaudeModel(ctx context.Context, c *gin.Context, svc *service.GatewayService, account *service.Account, path string, stream bool, model string) (*service.ForwardResult, error) {
	body := []byte(fmt.Sprintf(`{"model":%q,"max_tokens":32,"stream":%t,"messages":[{"role":"user","content":"hello"}]}`, model, stream))
	switch path {
	case "messages":
		parsed, err := service.ParseGatewayRequest(service.NewRequestBodyRef(body), service.PlatformAnthropic)
		if err != nil {
			return nil, err
		}
		return svc.Forward(ctx, c, account, parsed)
	case "chat/completions":
		return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
	default:
		body = []byte(fmt.Sprintf(`{"model":%q,"input":"hello","stream":%t}`, model, stream))
		return svc.ForwardAsResponses(ctx, c, account, body, nil)
	}
}

func tier223ClaudeSSE(usage string) string {
	return "event: message_start\ndata: " + `{"type":"message_start","message":{"id":"msg_tier223","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[],"usage":` + usage + "}}\n\n" +
		"event: content_block_start\ndata: " + `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
		"event: content_block_delta\ndata: " + `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello"}}` + "\n\n" +
		"event: content_block_stop\ndata: " + `{"type":"content_block_stop","index":0}` + "\n\n" +
		"event: message_delta\ndata: " + `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":7}}` + "\n\nevent: message_stop\ndata: " + `{"type":"message_stop"}` + "\n\n"
}

func TestClaudeTokenRotationBarrierHTTPAndSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	template := tier223SQLiteTemplate(t)
	for _, path := range []string{"messages", "chat/completions", "responses"} {
		for _, stream := range []bool{false, true} {
			for _, scenario := range []string{"rotated", "unchanged", "second-401", "wrong-status", "canceled"} {
				t.Run(fmt.Sprintf("%s/stream=%t/%s", path, stream, scenario), func(t *testing.T) {
					db, client := tier223SQLiteLedger(t, template, service.PlatformAnthropic)
					repo := repository.NewAccountRepository(client, db, nil)
					account, err := repo.GetByID(context.Background(), 3)
					require.NoError(t, err)
					redisServer := miniredis.RunT(t)
					rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
					t.Cleanup(func() { _ = rdb.Close() })
					cache := repository.NewGeminiTokenCache(rdb)
					require.NoError(t, cache.SetAccessToken(context.Background(), service.ClaudeTokenCacheKey(account), "old-token", time.Hour))
					// A hydrated snapshot may already contain the new token while the cache is old.
					account.Credentials["access_token"] = "new-token"
					arrived, release := make(chan []byte, 1), make(chan struct{})
					var releaseOnce sync.Once
					releaseRequest := func() { releaseOnce.Do(func() { close(release) }) }
					t.Cleanup(releaseRequest)
					var calls atomic.Int32
					var first []byte
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						payload, readErr := io.ReadAll(r.Body)
						if readErr != nil {
							t.Error(readErr)
							return
						}
						call := calls.Add(1)
						if call == 1 {
							if r.Header.Get("Authorization") != "Bearer old-token" {
								t.Errorf("first wire token: %q", r.Header.Get("Authorization"))
							}
							first = payload
							arrived <- payload
							select {
							case <-release:
							case <-r.Context().Done():
								return
							}
						} else {
							if call != 2 || r.Header.Get("Authorization") != "Bearer new-token" {
								t.Errorf("unexpected retry %d: %q", call, r.Header.Get("Authorization"))
							}
							if string(first) != string(payload) {
								t.Error("replayed request bytes changed")
							}
							_, updateErr := db.Exec(`UPDATE accounts SET credentials='{"access_token":"third-token","refresh_token":"retained-refresh"}' WHERE id=3`)
							if updateErr != nil {
								t.Error(updateErr)
							}
						}
						if call == 1 || scenario == "second-401" {
							status := http.StatusUnauthorized
							if scenario == "wrong-status" {
								status = http.StatusBadRequest
							}
							w.WriteHeader(status)
							_, _ = io.WriteString(w, `{"type":"error","error":{"type":"authentication_error","message":"access token revoked"}}`)
							return
						}
						if path == "messages" && !stream {
							w.Header().Set("Content-Type", "application/json")
							_, _ = io.WriteString(w, `{"id":"msg_tier223","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":11,"output_tokens":7}}`)
						} else {
							w.Header().Set("Content-Type", "text/event-stream")
							_, _ = io.WriteString(w, tier223ClaudeSSE(`{"input_tokens":11}`))
						}
					}))
					t.Cleanup(server.Close)
					target, err := url.Parse(server.URL)
					require.NoError(t, err)
					cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
					svc := tier223ClaudeGateway(repo, db, cfg, tier223LocalHTTP{server.Client(), target}, cache, nil)
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+path, nil).WithContext(ctx)
					done := make(chan error, 1)
					go func() { _, forwardErr := tier223ForwardClaude(ctx, c, svc, account, path, stream); done <- forwardErr }()
					select {
					case <-arrived:
					case <-time.After(3 * time.Second):
						t.Fatal("old token did not reach the HTTP barrier")
					}
					stored := "new-token"
					if scenario == "unchanged" {
						stored = "old-token"
					}
					credentials, err := json.Marshal(map[string]string{"access_token": stored, "refresh_token": "retained-refresh"})
					require.NoError(t, err)
					_, err = db.Exec(`UPDATE accounts SET credentials=? WHERE id=3`, string(credentials))
					require.NoError(t, err, "commit the new credential while the old HTTP request is in flight")
					if scenario == "canceled" {
						cancel()
					}
					releaseRequest()
					select {
					case err = <-done:
					case <-time.After(3 * time.Second):
						t.Fatal("forward did not complete")
					}
					wantCalls := int32(1)
					if scenario == "rotated" || scenario == "second-401" {
						wantCalls = 2
					}
					require.Equal(t, wantCalls, calls.Load())
					if scenario == "rotated" {
						require.NoError(t, err)
						require.Contains(t, w.Body.String(), "hello")
						require.NotContains(t, w.Body.String(), "revoked")
					} else {
						require.Error(t, err)
					}
					if scenario == "canceled" {
						require.ErrorIs(t, ctx.Err(), context.Canceled)
						var cooldown sql.NullString
						require.NoError(t, db.QueryRow(`SELECT temp_unschedulable_until FROM accounts WHERE id=3`).Scan(&cooldown))
						require.False(t, cooldown.Valid, "a canceled request must not persist a cooldown")
					}
					var after string
					require.NoError(t, db.QueryRow(`SELECT json_extract(credentials,'$.refresh_token') FROM accounts WHERE id=3`).Scan(&after))
					require.Equal(t, "retained-refresh", after, "401 recovery must not refresh or persist credentials")
					if scenario == "rotated" {
						cached, _ := cache.GetAccessToken(context.Background(), service.ClaudeTokenCacheKey(account))
						require.NotEqual(t, "old-token", cached)
					}
				})
			}
		}
	}
}

func tier223Pricing(t *testing.T, cfg *config.Config, zero bool) *service.PricingService {
	t.Helper()
	cfg.Pricing.DataDir = t.TempDir()
	cfg.Pricing.UpdateIntervalHours = 24
	prices := map[string]any{}
	for _, model := range []struct {
		name  string
		scale float64
	}{{"claude-sonnet-4-6", 1}, {"claude-opus-4-6", 10}, {"gpt-5.4", 1}} {
		scale := model.scale
		if zero {
			scale = 0
		}
		prices[model.name] = map[string]float64{
			"input_cost_per_token": 2e-6 * scale, "output_cost_per_token": 3e-6 * scale,
			"cache_read_input_token_cost": 1e-6 * scale, "cache_creation_input_token_cost": 4e-6 * scale,
			"cache_creation_input_token_cost_above_1hr": 8e-6 * scale,
		}
	}
	body, err := json.Marshal(prices)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"), body, 0600))
	require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.sha256"), []byte(fmt.Sprintf("%x", sha256.Sum256(body))), 0600))
	pricing := service.NewPricingService(cfg, nil)
	require.NoError(t, pricing.Initialize())
	t.Cleanup(pricing.Stop)
	return pricing
}

func tier223AssertLedger(t *testing.T, db *sql.DB, expectedStandard float64, input, aggregate, five, hour int) {
	t.Helper()
	var total, actual, accountCost, balance, quota float64
	var gotInput, gotAggregate, gotFive, gotHour, count int
	require.NoError(t, db.QueryRow(`SELECT total_cost,actual_cost,COALESCE(account_stats_cost,total_cost*COALESCE(account_rate_multiplier,1)),input_tokens,cache_creation_tokens,cache_creation_5m_tokens,cache_creation_1h_tokens FROM usage_logs`).Scan(&total, &actual, &accountCost, &gotInput, &gotAggregate, &gotFive, &gotHour))
	require.InDelta(t, expectedStandard, total, 1e-12)
	require.InDelta(t, expectedStandard*2, actual, 1e-12)
	require.InDelta(t, expectedStandard*0.5, accountCost, 1e-12)
	require.Equal(t, []int{input, aggregate, five, hour}, []int{gotInput, gotAggregate, gotFive, gotHour})
	require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
	require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quota))
	require.InDelta(t, 10-actual, balance, 1e-12)
	require.InDelta(t, actual, quota, 1e-12)
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&count))
	require.Equal(t, 1, count)
}

func TestClaudeCacheTTLTransportObservationAndSQLiteSettlement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	template := tier223SQLiteTemplate(t)
	for _, mode := range []struct {
		path                string
		stream, passthrough bool
	}{
		{"messages", false, false}, {"messages", true, false}, {"messages", false, true}, {"messages", true, true},
		{"chat/completions", false, false}, {"chat/completions", true, false}, {"responses", false, false}, {"responses", true, false},
	} {
		for _, scenario := range []string{"split-updates", "contradictory-details", "zero-price", "model-downgrade", "model-conflict"} {
			t.Run(fmt.Sprintf("%s/stream=%t/pass=%t/%s", mode.path, mode.stream, mode.passthrough, scenario), func(t *testing.T) {
				db, client := tier223SQLiteLedger(t, template, service.PlatformAnthropic)
				repo := repository.NewAccountRepository(client, db, nil)
				account, err := repo.GetByID(context.Background(), 3)
				require.NoError(t, err)
				account.Credentials = map[string]any{"access_token": "fixture-token", "api_key": "fixture-key"}
				if mode.passthrough {
					account.Type = service.AccountTypeAPIKey
				}
				upstreamModel := "claude-sonnet-4-6"
				if strings.HasPrefix(scenario, "model-") {
					upstreamModel = "claude-opus-4-6"
					account.Credentials["model_mapping"] = map[string]any{"claude-sonnet-4-6": upstreamModel}
				}
				accountRate := 0.5
				account.RateMultiplier = &accountRate
				five, hour := 100, 100
				standard := 0.00283
				if scenario == "contradictory-details" {
					five, hour = 200, 200
					standard = 0.00303
				}
				if scenario == "zero-price" {
					standard = 0
				}
				if scenario == "model-conflict" {
					standard *= 10
				}
				usage := fmt.Sprintf(`{"input_tokens":500,"output_tokens":10,"cache_creation_input_tokens":300,"cache_read_input_tokens":200,"cache_creation":{"ephemeral_5m_input_tokens":%d,"ephemeral_1h_input_tokens":%d}}`, five, hour)
				streamBody := tier223ClaudeSSE(`{"input_tokens":500,"cache_creation_input_tokens":300,"cache_read_input_tokens":200}`)
				// Independent cache-only frames must retain the aggregate and the other TTL.
				updates := fmt.Sprintf("event: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"cache_creation\":{\"ephemeral_5m_input_tokens\":%d}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"cache_creation\":{\"ephemeral_1h_input_tokens\":%d}}}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"cache_creation_input_tokens\":0,\"output_tokens\":10}}\n\n", five, hour)
				if scenario == "model-conflict" {
					updates += "event: message_delta\ndata: " + `{"type":"message_delta","model":"claude-opus-4-6"}` + "\n\n"
				}
				streamBody = strings.Replace(streamBody, "event: message_stop", updates+"event: message_stop", 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					payload, readErr := io.ReadAll(r.Body)
					if readErr != nil {
						t.Error(readErr)
						return
					}
					if got := gjson.GetBytes(payload, "model").String(); got != upstreamModel {
						t.Errorf("sent model: %s, want %s", got, upstreamModel)
					}
					w.Header().Set("X-Request-Id", "tier223-cache")
					if mode.path == "messages" && !mode.stream {
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprintf(w, `{"id":"msg_tier223","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":%s}`, usage)
					} else {
						w.Header().Set("Content-Type", "text/event-stream")
						_, _ = io.WriteString(w, streamBody)
					}
				}))
				t.Cleanup(server.Close)
				target, err := url.Parse(server.URL)
				require.NoError(t, err)
				cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
				pricing := tier223Pricing(t, cfg, scenario == "zero-price")
				svc := tier223ClaudeGateway(repo, db, cfg, tier223LocalHTTP{server.Client(), target}, nil, pricing)
				ctx := service.WithCompositeRouteDecision(context.Background(), service.CompositeRouteDecision{Matched: true, PublicModel: "opaque public alias", UpstreamModel: upstreamModel, TargetPlatform: service.PlatformAnthropic})
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+mode.path, nil).WithContext(ctx)
				result, err := tier223ForwardClaudeModel(ctx, c, svc, account, mode.path, mode.stream, upstreamModel)
				require.NoError(t, err)
				require.Equal(t, []int{500, 10, 300, five, hour}, []int{result.Usage.InputTokens, result.Usage.OutputTokens, result.Usage.CacheCreationInputTokens, result.Usage.CacheCreation5mTokens, result.Usage.CacheCreation1hTokens})
				require.Contains(t, w.Body.String(), "opaque public alias")
				require.Equal(t, "claude-sonnet-4-6", result.UpstreamResponseModel)
				conflict := scenario == "model-conflict" && (mode.path != "messages" || mode.stream)
				require.Equal(t, conflict, result.UpstreamResponseModelConflict)
				if scenario == "model-conflict" && !conflict {
					standard = 0.00283
				}
				groupID := int64(4)
				user := &service.User{ID: 1, Balance: 10, Status: service.StatusActive}
				key := &service.APIKey{ID: 2, UserID: 1, User: user, Quota: 1, GroupID: &groupID, Group: &service.Group{ID: 4, Platform: service.PlatformAnthropic, Hydrated: true, RateMultiplier: 2}}
				input := &service.RecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}, ChannelUsageFields: service.ChannelUsageFields{BillingModelSource: "response_model"}}
				require.NoError(t, svc.RecordUsage(ctx, input))
				require.NoError(t, svc.RecordUsage(ctx, input))
				tier223AssertLedger(t, db, standard, 500, 300, five, hour)
			})
		}
	}
}
