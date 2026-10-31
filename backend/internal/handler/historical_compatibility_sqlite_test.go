//go:build unit

package handler

import (
	"bufio"
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
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Exercise the actual HTTP transport, Responses writer, terminal handler and
// queued SQLite settlement together. The consumer requires integer sequences.
func TestHistoricalToolsStrictHTTPAndSQLiteUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	templatePath := filepath.Join(t.TempDir(), "template.db")
	db, err := sql.Open("sqlite", "file:"+templatePath+"?_time_format=sqlite")
	require.NoError(t, err)
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))
	require.NoError(t, repository.EnsureSQLiteAuxTables(context.Background(), db))
	require.NoError(t, db.Close())
	template, err := os.ReadFile(templatePath)
	require.NoError(t, err)
	for _, scenario := range []string{"valid", "truncated", "custom-alias"} {
		t.Run(scenario, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "usage.db")
			require.NoError(t, os.WriteFile(path, template, 0600))
			db, err := sql.Open("sqlite", "file:"+path+"?_time_format=sqlite&_pragma=foreign_keys(1)")
			require.NoError(t, err)
			db.SetMaxOpenConns(1)
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
			t.Cleanup(func() { _ = client.Close() })
			for _, statement := range []string{
				`INSERT INTO users(id,email,password_hash,role,balance) VALUES(1,'compat@example.test','hash','user',10)`,
				`INSERT INTO api_keys(id,user_id,key,name,quota,status) VALUES(2,1,'fixture-key','compat',1,'active')`,
				`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(3,'compat','openai','apikey','{}','{}')`,
			} {
				_, err := db.Exec(statement)
				require.NoError(t, err)
			}
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/chat/completions" {
					t.Errorf("unexpected upstream path: %s", r.URL.Path)
				}
				name, arguments := "read", `{}`
				if scenario == "truncated" {
					arguments = `{"path":`
				}
				if scenario == "custom-alias" {
					name, arguments = "functions__exec", `{"input":"pwd"}`
				}
				chunk, err := json.Marshal(map[string]any{"id": "chat_fixture", "choices": []any{map[string]any{
					"index": 0, "delta": map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "call", "type": "function", "function": map[string]any{"name": name, "arguments": arguments}}}}, "finish_reason": "tool_calls",
				}}})
				if err != nil {
					t.Error(err)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("X-Request-ID", "compat-"+scenario)
				_, _ = fmt.Fprintf(w, "data: %s\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":1000,\"completion_tokens\":10,\"total_tokens\":1010}}\n\ndata: [DONE]\n\n", chunk)
			}))
			t.Cleanup(upstream.Close)
			cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
			svc := service.NewOpenAIGatewayService(repository.NewAccountRepository(client, db, nil), repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, nil, grokAuditHTTP{upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, nil, nil, nil, nil, nil, nil)
			account := &service.Account{ID: 3, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "fixture-key", "base_url": upstream.URL}, Extra: map[string]any{openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions)}}
			user := &service.User{ID: 1, Status: service.StatusActive, Balance: 10}
			key := &service.APIKey{ID: 2, UserID: 1, User: user, Key: "fixture-key", Quota: 1, Status: service.StatusActive}
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
				t.Fatal("worker did not start")
			}
			h := &OpenAIGatewayHandler{gatewayService: svc, usageRecordWorkerPool: pool}
			type forwarded struct {
				input *service.OpenAIRecordUsageInput
				ctx   context.Context
				err   error
			}
			ready, settled := make(chan forwarded, 1), make(chan error, 1)
			router := gin.New()
			router.POST("/v1/responses", func(c *gin.Context) {
				body, err := io.ReadAll(c.Request.Body)
				if err != nil {
					t.Error(err)
					return
				}
				ctx := context.WithValue(c.Request.Context(), ctxkey.ClientRequestID, "compat-client-"+scenario)
				c.Request = c.Request.WithContext(ctx)
				result, forwardErr := svc.Forward(ctx, c, account, body)
				if result == nil {
					t.Error("missing usage result")
					return
				}
				input := &service.OpenAIRecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}}
				h.submitOpenAIUsageRecordTask(ctx, result, func(taskCtx context.Context) { settled <- svc.RecordUsage(taskCtx, input) })
				ready <- forwarded{input: input, ctx: context.WithoutCancel(ctx), err: forwardErr}
				if forwardErr != nil {
					h.handleStreamingAwareError(c, http.StatusBadGateway, "upstream_error", "invalid tool arguments", true)
				}
			})
			gateway := httptest.NewServer(router)
			t.Cleanup(gateway.Close)
			tools := `[ {"type":"function","name":"read","parameters":{"type":"object"}} ]`
			if scenario == "custom-alias" {
				tools = `[{"type":"custom","name":"exec"},{"type":"namespace","name":"functions","tools":[{"type":"function","name":"wait","parameters":{"type":"object"}}]}]`
			}
			body := `{"model":"gpt-5.4","input":"hi","stream":true,"tools":` + tools + `}`
			response, err := gateway.Client().Post(gateway.URL+"/v1/responses", "application/json", strings.NewReader(body))
			require.NoError(t, err)
			defer response.Body.Close()
			last := int64(-1)
			terminal, tool := "", ""
			scanner := bufio.NewScanner(response.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if !strings.HasPrefix(line, "data: ") || line == "data: [DONE]" {
					continue
				}
				payload := strings.TrimPrefix(line, "data: ")
				var event struct {
					Type     string `json:"type"`
					Sequence *int64 `json:"sequence_number"`
				}
				require.NoError(t, json.Unmarshal([]byte(payload), &event))
				require.NotNil(t, event.Sequence, "strict consumer requires sequence_number")
				require.Greater(t, *event.Sequence, last)
				last = *event.Sequence
				if event.Type == "response.completed" || event.Type == "response.failed" {
					terminal = event.Type
				}
				if event.Type == "response.output_item.done" {
					kind := gjson.Get(payload, "item.type").String()
					if kind == "function_call" {
						require.True(t, json.Valid([]byte(gjson.Get(payload, "item.arguments").String())))
						tool = kind
					}
					if kind == "custom_tool_call" {
						require.Equal(t, "exec", gjson.Get(payload, "item.name").String())
						require.Equal(t, "pwd", gjson.Get(payload, "item.input").String())
						tool = kind
					}
				}
			}
			require.NoError(t, scanner.Err())
			var forward forwarded
			select {
			case forward = <-ready:
			case <-time.After(5 * time.Second):
				t.Fatal("forwarding result was not delivered")
			}
			if scenario == "truncated" {
				require.Error(t, forward.err)
				require.Equal(t, "response.failed", terminal)
				require.Empty(t, tool)
			} else {
				require.NoError(t, forward.err)
				require.Equal(t, "response.completed", terminal)
				require.NotEmpty(t, tool)
			}
			unblock()
			select {
			case err := <-settled:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("queued usage did not settle")
			}
			var inputTokens, outputTokens, count int
			var cost, balance, quota float64
			require.NoError(t, db.QueryRow(`SELECT input_tokens,output_tokens,actual_cost FROM usage_logs`).Scan(&inputTokens, &outputTokens, &cost))
			require.Equal(t, 1000, inputTokens)
			require.Equal(t, 10, outputTokens)
			require.Positive(t, cost)
			require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
			require.NoError(t, db.QueryRow(`SELECT quota_used FROM api_keys WHERE id=2`).Scan(&quota))
			require.InDelta(t, 10-cost, balance, 1e-12)
			require.InDelta(t, cost, quota, 1e-12)
			require.NoError(t, svc.RecordUsage(forward.ctx, forward.input))
			require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM usage_billing_dedup`).Scan(&count))
			require.Equal(t, 1, count)
			var after float64
			require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&after))
			require.Equal(t, balance, after)
		})
	}
}
