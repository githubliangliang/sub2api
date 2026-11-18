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
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGrokServerSearchHTTPProtocolsAndPrefixScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, platform := range []string{service.PlatformGrok, service.PlatformOpenAI} {
		for _, path := range []string{"chat", "messages", "responses"} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/stream=%t", platform, path, stream), func(t *testing.T) {
					model := "grok-4.5"
					if platform == service.PlatformOpenAI {
						model = "gpt-5.4"
					}
					items := `[{"type":"custom_tool_call","id":"ctc_1","call_id":"xs_search","name":"x_search","input":"{\"q\":\"query\"}"},{"type":"custom_tool_call","id":"ctc_2","call_id":"as_search","name":"web_search","input":"{\"q\":\"query\"}"},{"type":"custom_tool_call","id":"ctc_3","call_id":"ordinary-call","name":"client_tool","input":"{\"cmd\":\"pwd\"}"},{"type":"message","role":"assistant","content":[{"type":"output_text","text":"found it"}]}]`
					sse := `data: {"type":"response.created","response":{"id":"resp_search","model":"` + model + `"}}` + "\n\n" + `data: {"type":"response.output_text.delta","output_index":3,"delta":"found it"}` + "\n\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_search\",\"model\":\"" + model + "\",\"status\":\"completed\",\"output\":" + items + ",\"usage\":{\"input_tokens\":5,\"output_tokens\":2}}}\n\n"
					var toolEvents strings.Builder
					for index, call := range []struct{ id, name, input string }{{"xs_search", "x_search", `{"q":"query"}`}, {"as_search", "web_search", `{"q":"query"}`}, {"ordinary-call", "client_tool", `{"cmd":"pwd"}`}} {
						fmt.Fprintf(&toolEvents, "data: {\"type\":\"response.output_item.added\",\"output_index\":%d,\"item\":{\"type\":\"custom_tool_call\",\"id\":\"ctc_%d\",\"call_id\":%q,\"name\":%q,\"input\":\"\"}}\n\n", index, index+1, call.id, call.name)
						fmt.Fprintf(&toolEvents, "data: {\"type\":\"response.custom_tool_call_input.delta\",\"output_index\":%d,\"item_id\":\"ctc_%d\",\"delta\":%q}\n\n", index, index+1, call.input)
						fmt.Fprintf(&toolEvents, "data: {\"type\":\"response.output_item.done\",\"output_index\":%d,\"item\":{\"type\":\"custom_tool_call\",\"id\":\"ctc_%d\",\"call_id\":%q,\"name\":%q,\"input\":%q}}\n\n", index, index+1, call.id, call.name, call.input)
					}
					sse = strings.Replace(sse, `data: {"type":"response.output_text.delta"`, toolEvents.String()+`data: {"type":"response.output_text.delta"`, 1)
					sse = strings.Replace(sse, `data: {"type":"response.completed"`, `data: {"type":"response.output_item.done","output_index":3,"item":{"type":"message","id":"msg_1","status":"completed","role":"assistant","content":[{"type":"output_text","text":"found it"}]}}`+"\n\n"+`data: {"type":"response.completed"`, 1)
					if path != "responses" && !stream {
						sse = strings.Replace(sse, `"output":`+items, `"output":[]`, 1)
					}
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = io.Copy(io.Discard, r.Body)
						w.Header().Set("Content-Type", "text/event-stream")
						_, _ = io.WriteString(w, sse)
					}))
					t.Cleanup(server.Close)
					target, err := url.Parse(server.URL)
					require.NoError(t, err)
					account := &service.Account{ID: 1, Platform: platform, Type: service.AccountTypeOAuth, Status: service.StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture-key", "access_token": "fixture-token", "refresh_token": "fixture-refresh", "expires_at": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339), "chatgpt_account_id": "fixture-account", "base_url": server.URL}}
					if platform == service.PlatformGrok {
						account.Type = service.AccountTypeOAuth
						account.Credentials["base_url"] = xai.DefaultCLIBaseURL
					}
					repo := &grokCredentialHandlerRepo{accounts: []service.Account{*account}, missingOnGet: map[int64]bool{}}
					cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
					svc := service.NewOpenAIGatewayService(repo, nil, nil, nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, nil, tier223LocalHTTP{server.Client(), target}, &service.DeferredService{}, nil, service.NewGrokTokenProvider(repo, nil), nil, nil, nil, nil, nil)
					body := []byte(fmt.Sprintf(`{"model":%q,"stream":%t,"messages":[{"role":"user","content":"hello"}]}`, model, stream))
					if path == "messages" {
						body = []byte(fmt.Sprintf(`{"model":%q,"stream":%t,"max_tokens":32,"messages":[{"role":"user","content":"hello"}]}`, model, stream))
					}
					if path == "chat" {
						body = []byte(fmt.Sprintf(`{"model":%q,"stream":%t,"messages":[{"role":"user","content":"hello"}]}`, model, stream))
					}
					if path == "responses" {
						body = []byte(fmt.Sprintf(`{"model":%q,"stream":%t,"input":"hello"}`, model, stream))
					}
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+path, strings.NewReader(string(body)))
					c.Set("api_key", &service.APIKey{ID: 88000})
					var result *service.OpenAIForwardResult
					switch path {
					case "chat":
						result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
					case "messages":
						result, err = svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
					default:
						result, err = svc.Forward(context.Background(), c, account, body)
					}
					require.NoError(t, err)
					require.Contains(t, w.Body.String(), "found it")
					require.Contains(t, w.Body.String(), "client_tool")
					if platform == service.PlatformGrok {
						require.Equal(t, 2, result.SearchCount)
					}
					if platform == service.PlatformGrok && path != "responses" {
						require.NotContains(t, w.Body.String(), "xs_search")
						require.NotContains(t, w.Body.String(), "as_search")
					} else {
						require.Contains(t, w.Body.String(), "xs_search")
						require.Contains(t, w.Body.String(), "as_search")
					}
				})
			}
		}
	}
}
