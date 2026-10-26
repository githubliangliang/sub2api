//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestConfiguredMinOutputTokens_DefaultAndInstanceIsolation(t *testing.T) {
	require.Equal(t, 128, configuredMinOutputTokens(nil))
	require.Equal(t, 128, configuredMinOutputTokens(&config.Config{}))
	for _, floor := range []int{16, 256} {
		t.Run(fmt.Sprint(floor), func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{Gateway: config.GatewayConfig{MinOutputTokens: floor}}
			for range 20 {
				req := &apicompat.AnthropicRequest{Model: "gpt-4o", MaxTokens: 1,
					Messages: []apicompat.AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}}
				out, err := anthropicToResponses(cfg, req)
				require.NoError(t, err)
				require.Equal(t, floor, *out.MaxOutputTokens)
			}
		})
	}
}

func TestMinOutputTokens_ForwardAsAnthropicRawChatUsesInstanceConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, floor := range []int{16, 256} {
		t.Run(fmt.Sprint(floor), func(t *testing.T) {
			t.Parallel()
			body := []byte(`{"model":"gpt-4o","max_tokens":1,"messages":[{"role":"user","content":"hello"}],"stream":false}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200,
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"id":"chatcmpl","object":"chat.completion","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))}}
			cfg := rawChatCompletionsTestConfig()
			cfg.Gateway.MinOutputTokens = floor
			svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream}
			result, err := svc.ForwardAsAnthropic(context.Background(), c, forceChatMessagesFallbackAccount(), body, "", "")
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, int64(floor), gjson.GetBytes(upstream.lastBody, "max_completion_tokens").Int())
			require.Equal(t, 1, result.Usage.OutputTokens, "configured limit must not overwrite observed usage")
		})
	}
}

func TestMinOutputTokens_NativeResponsesAndWebSocketRemainNative(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	for _, payload := range []string{
		`{"model":"gpt-4o","input":"hello","max_output_tokens":1}`,
		`{"type":"response.create","model":"gpt-4o","input":"hello","max_output_tokens":1}`,
	} {
		httpBody, _, err := normalizeOpenAIResponsesCompatibilityBodyWithOptions([]byte(payload), account, openAIResponsesCompatibilityOptions{})
		require.NoError(t, err)
		require.Equal(t, payload, string(httpBody))
		for range 2 {
			wsBody, _, err := normalizeOpenAIResponsesWebSocketCompatibilityBody([]byte(payload), account, false)
			require.NoError(t, err)
			require.Equal(t, payload, string(wsBody))
		}
		compact, _, err := normalizeOpenAIResponsesCompatibilityBodyWithOptions([]byte(payload), account, openAIResponsesCompatibilityOptions{Compact: true})
		require.NoError(t, err)
		require.Equal(t, payload, string(compact))
	}
}
