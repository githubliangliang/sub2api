//go:build unit

package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAICacheTTLAndToolSchemaHTTPWSSettlementSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	template := tier223SQLiteTemplate(t)
	for _, mode := range []struct {
		name, path              string
		stream, passthrough, ws bool
	}{
		{"json", "responses", false, false, false}, {"sse", "responses", true, false, false}, {"buffered-sse", "responses", false, false, false},
		{"oauth-pass", "responses", true, true, false}, {"chat-stream", "chat", true, false, false}, {"chat-buffer", "chat", false, false, false},
		{"messages-stream", "messages", true, false, false}, {"messages-buffer", "messages", false, false, false}, {"ws", "responses", false, false, true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			db, client := tier223SQLiteLedger(t, template, service.PlatformOpenAI)
			cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
			if mode.ws {
				cfg.Gateway.OpenAIWS = config.GatewayOpenAIWSConfig{Enabled: true, APIKeyEnabled: true, ResponsesWebsocketsV2: true, MaxConnsPerAccount: 1, QueueLimitPerConn: 2, DialTimeoutSeconds: 3, ReadTimeoutSeconds: 5, WriteTimeoutSeconds: 3}
			}
			response := `{"id":"resp_cache","object":"response","model":"gpt-5.4","status":"completed","output":[{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"answer"}]}],"usage":{"input_tokens":1000,"output_tokens":10,"input_tokens_details":{"cached_tokens":200,"cache_write_tokens":300},"cache_creation":{"ephemeral_5m_input_tokens":100,"ephemeral_1h_input_tokens":100}}}`
			captured := make(chan []byte, 1)
			upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if mode.ws {
					conn, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						t.Error(err)
						return
					}
					defer conn.Close()
					_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
					_, payload, err := conn.ReadMessage()
					if err != nil {
						t.Error(err)
						return
					}
					captured <- payload
					_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"response.created","response":{"id":"resp_cache","model":"gpt-5.4"}}`))
					_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"response.completed","response":`+response+`}`))
					return
				}
				payload, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					return
				}
				captured <- payload
				w.Header().Set("X-Request-Id", "openai-cache-fixture")
				if mode.name == "json" {
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, response)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, `data: {"type":"response.created","response":{"id":"resp_cache","model":"gpt-5.4","output":[]}}`+"\n\n"+`data: {"type":"response.output_text.delta","output_index":0,"content_index":0,"delta":"answer"}`+"\n\ndata: {\"type\":\"response.completed\",\"response\":"+response+"}\n\n")
			}))
			t.Cleanup(server.Close)
			target, err := url.Parse(server.URL)
			require.NoError(t, err)
			pricing := tier223Pricing(t, cfg, false)
			svc := service.NewOpenAIGatewayService(repository.NewAccountRepository(client, db, nil), repository.NewUsageLogRepository(nil, db), repository.NewUsageBillingRepository(nil, db), nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, pricing), nil, nil, tier223LocalHTTP{server.Client(), target}, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil)
			t.Cleanup(svc.CloseOpenAIWSPool)
			accountRate := 0.5
			account := &service.Account{ID: 3, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Concurrency: 1, RateMultiplier: &accountRate, Credentials: map[string]any{"api_key": "fixture-key", "base_url": server.URL, "access_token": "fixture-token", "chatgpt_account_id": "fixture-account"}, Extra: map[string]any{"responses_websockets_v2_enabled": mode.ws, "openai_passthrough": mode.passthrough}}
			if mode.passthrough {
				account.Type = service.AccountTypeOAuth
			}
			tools := `[{"type":"function","name":"choose","parameters":{"type":["object","null"],"properties":{"value":{"type":"string","pattern":"\\p{L}","default":{"pattern":"instance literal","big":9007199254740993}}},"required":null,"additionalProperties":false}}]`
			body := []byte(fmt.Sprintf(`{"model":"gpt-5.4","stream":%t,"store":false,"input":[{"type":"item_reference","id":"at_123e4567-e89b-12d3-a456-426614174000"},{"type":"additional_tools","id":"at_123e4567-e89b-12d3-a456-426614174000","tools":[]},{"type":"message","role":"user","content":"hello"}],"tools":%s}`, mode.stream, tools))
			if mode.path == "chat" {
				body = []byte(fmt.Sprintf(`{"model":"gpt-5.4","stream":%t,"messages":[{"role":"user","content":"hello"}],"tools":[{"type":"function","function":{"name":"choose","parameters":{"type":"object","properties":{"value":{"type":"string","pattern":"\\p{L}","default":{"big":9007199254740993}}}}}}]}`, mode.stream))
			}
			if mode.path == "messages" {
				body = []byte(fmt.Sprintf(`{"model":"gpt-5.4","stream":%t,"max_tokens":32,"messages":[{"role":"user","content":"hello"}],"tools":[{"name":"choose","input_schema":{"type":"object","properties":{"value":{"type":"string","pattern":"\\p{L}","default":{"big":9007199254740993}}}}}]}`, mode.stream))
			}
			original := append([]byte(nil), body...)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+mode.path, strings.NewReader(string(body)))
			var result *service.OpenAIForwardResult
			switch mode.path {
			case "chat":
				result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
			case "messages":
				result, err = svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
			default:
				result, err = svc.Forward(context.Background(), c, account, body)
			}
			require.NoError(t, err)
			require.Equal(t, original, body)
			require.Equal(t, mode.ws, result.OpenAIWSMode)
			require.Equal(t, []int{1000, 10, 300, 200, 100, 100}, []int{result.Usage.InputTokens, result.Usage.OutputTokens, result.Usage.CacheCreationInputTokens, result.Usage.CacheReadInputTokens, result.Usage.CacheCreation5mTokens, result.Usage.CacheCreation1hTokens})
			select {
			case wire := <-captured:
				parameters := gjson.GetBytes(wire, "tools.0.parameters")
				require.True(t, parameters.Exists(), "%s", wire)
				require.False(t, parameters.Get("properties.value.pattern").Exists(), "OpenAI wire must drop an unsupported regexp")
				require.Equal(t, "9007199254740993", parameters.Get("properties.value.default.big").Raw)
				if mode.passthrough {
					require.Equal(t, "at_123e4567e89b12d3a456426614174000", gjson.GetBytes(wire, "input.0.id").String())
					require.Equal(t, "at_123e4567e89b12d3a456426614174000", gjson.GetBytes(wire, "input.1.id").String())
				}
			case <-time.After(time.Second):
				t.Fatal("upstream request missing")
			}
			groupID := int64(4)
			user := &service.User{ID: 1, Balance: 10, Status: service.StatusActive}
			key := &service.APIKey{ID: 2, UserID: 1, User: user, Quota: 1, GroupID: &groupID, Group: &service.Group{ID: 4, Platform: service.PlatformOpenAI, Hydrated: true, RateMultiplier: 2}}
			input := &service.OpenAIRecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}}
			require.NoError(t, svc.RecordUsage(context.Background(), input))
			require.NoError(t, svc.RecordUsage(context.Background(), input))
			tier223AssertLedger(t, db, 0.00283, 500, 300, 100, 100)
		})
	}
}
