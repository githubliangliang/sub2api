//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestTier2ToolSchemaBeforeDispatch(t *testing.T) {
	for _, path := range []string{"messages-responses", "messages-chat", "responses"} {
		t.Run(path, func(t *testing.T) {
			body := `{"model":"gpt-5.4","max_tokens":1000,"tools":[{"name":"lookup","input_schema":{"type":"object","required":null,"properties":{"query":{"type":"string","default":{"required":null}}}}}],"messages":[{"role":"user","content":"hello"}]}`
			if path == "responses" {
				body = `{"model":"gpt-5.4","tools":[{"type":"function","name":"lookup","parameters":{"type":"object","required":null,"properties":{"query":{"type":"string","default":{"required":null}}}}}],"input":"hello"}`
			}
			payload := "data: " + `{"type":"response.completed","response":{"id":"resp_test","model":"gpt-5.4","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":10,"output_tokens":2}}}` + "\n\n"
			extra := map[string]any{"openai_responses_supported": true}
			contentType, schemaPath := "text/event-stream", "tools.0.parameters"
			if path == "messages-chat" {
				extra["openai_responses_supported"] = false
				contentType, schemaPath = "application/json", "tools.0.function.parameters"
				payload = `{"id":"chat_test","model":"gpt-5.4","choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2}}`
			}
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(payload))}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: extra, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://api.openai.com"}}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
			var err error
			if path == "responses" {
				_, err = svc.Forward(context.Background(), c, account, []byte(body))
			} else {
				_, err = svc.ForwardAsAnthropic(context.Background(), c, account, []byte(body), "", "")
			}
			require.NoError(t, err)
			require.True(t, gjson.ValidBytes(upstream.lastBody))
			require.False(t, gjson.GetBytes(upstream.lastBody, schemaPath+".required").Exists())
			require.Equal(t, "null", gjson.GetBytes(upstream.lastBody, schemaPath+".properties.query.default.required").Raw)
		})
	}
}

func TestTier2ToolSchemaNativeAnthropicDispatch(t *testing.T) {
	for _, path := range []string{"messages", "responses", "chat"} {
		t.Run(path, func(t *testing.T) {
			body := `{"model":"claude-sonnet-4-6","max_tokens":100,"messages":[{"role":"user","content":"hello"}],"tools":[{"name":"lookup","input_schema":{"type":"object","required":null,"properties":{"query":{"type":"string"}}}}]}`
			if path == "responses" {
				body = `{"model":"claude-sonnet-4-6","input":"hello","tools":[{"type":"function","name":"lookup","parameters":{"type":"object","required":null,"properties":{"query":{"type":"string"}}}}]}`
			}
			if path == "chat" {
				body = `{"model":"claude-sonnet-4-6","messages":[{"role":"user","content":"hello"}],"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object","required":null,"properties":{"query":{"type":"string"}}}}}]}`
			}
			upstream := &anthropicHTTPUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(namespaceToolAnthropicStream()))}}
			svc := &GatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			if path == "messages" {
				upstream.resp.Header.Set("Content-Type", "application/json")
				upstream.resp.Body = io.NopCloser(strings.NewReader(`{"id":"msg_fixture","model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`))
			}
			account := newAnthropicAPIKeyAccountForTest()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+path, strings.NewReader(body))
			var err error
			switch path {
			case "messages":
				_, err = svc.Forward(context.Background(), c, account, &ParsedRequest{Model: "claude-sonnet-4-6", Body: NewRequestBodyRef([]byte(body))})
			case "responses":
				_, err = svc.ForwardAsResponses(context.Background(), c, account, []byte(body), nil)
			case "chat":
				_, err = svc.ForwardAsChatCompletions(context.Background(), c, account, []byte(body), nil)
			}
			require.NoError(t, err)
			require.False(t, gjson.GetBytes(upstream.lastBody, "tools.0.input_schema.required").Exists())
			require.Equal(t, "string", gjson.GetBytes(upstream.lastBody, "tools.0.input_schema.properties.query.type").String())
		})
	}
}
