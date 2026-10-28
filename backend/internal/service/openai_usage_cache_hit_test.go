//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIUsage_PromptCacheHitFallback(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		want         int
	}{
		{"top level hit and miss", `"prompt_cache_hit_tokens":80,"prompt_cache_miss_tokens":20`, 80},
		{"prompt nested positive", `"prompt_cache_hit_tokens":80,"prompt_tokens_details":{"cached_tokens":64}`, 64},
		{"prompt nested zero", `"prompt_cache_hit_tokens":80,"prompt_tokens_details":{"cached_tokens":0}`, 0},
		{"input nested positive", `"prompt_cache_hit_tokens":80,"input_tokens_details":{"cached_tokens":32}`, 32},
		{"input nested zero wins", `"prompt_cache_hit_tokens":80,"input_tokens_details":{"cached_tokens":0},"prompt_tokens_details":{"cached_tokens":64}`, 0},
		{"legacy cached tokens", `"prompt_cache_hit_tokens":80,"cached_tokens":30`, 30},
		{"legacy cache read tokens", `"prompt_cache_hit_tokens":80,"cache_read_tokens":40`, 40},
		{"legacy cache read input", `"prompt_cache_hit_tokens":80,"cache_read_input_tokens":50`, 50},
		{"negative hit", `"prompt_cache_hit_tokens":-80`, 0},
		{"miss only", `"prompt_cache_miss_tokens":100`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			usage, ok := openAIUsageFromGJSON(gjson.Parse(`{"prompt_tokens":100,"completion_tokens":5,` + tc.fields + `}`))
			require.True(t, ok)
			require.Equal(t, 100, usage.InputTokens, "prompt tokens remain the total input bucket")
			require.Equal(t, 5, usage.OutputTokens)
			require.Equal(t, tc.want, usage.CacheReadInputTokens)
			require.Zero(t, usage.CacheCreationInputTokens)
		})
	}
}

func TestOpenAIUsage_PromptCacheHitHTTPAndSSE(t *testing.T) {
	const response = `{"id":"resp_cache_hit","object":"response","model":"gpt-5.4","status":"completed","output":[],"usage":{"prompt_tokens":100,"completion_tokens":5,"prompt_cache_hit_tokens":80,"prompt_cache_miss_tokens":20}}`
	const sse = "data: {\"type\":\"response.completed\",\"response\":" + response + "}\n\ndata: [DONE]\n\n"
	gin.SetMode(gin.TestMode)
	for _, format := range []string{"json", "sse_to_json", "streaming"} {
		t.Run(format, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			payload, contentType := response, "application/json"
			if format != "json" {
				payload, contentType = sse, "text/event-stream"
			}
			resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(payload))}
			t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })
			svc := &OpenAIGatewayService{cfg: &config.Config{}}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			var usage *OpenAIUsage
			if format == "streaming" {
				result, err := svc.handleStreamingResponse(context.Background(), resp, c, account, time.Now(), "gpt-5.4", "gpt-5.4")
				require.NoError(t, err)
				usage = result.usage
			} else {
				result, err := svc.handleNonStreamingResponse(context.Background(), resp, c, account, "gpt-5.4", "gpt-5.4")
				require.NoError(t, err)
				usage = result.usage
			}
			require.NotNil(t, usage)
			require.Equal(t, 100, usage.InputTokens)
			require.Equal(t, 80, usage.CacheReadInputTokens)
			require.Equal(t, 20, usage.InputTokens-usage.CacheReadInputTokens)
			require.Equal(t, 5, usage.OutputTokens)
			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
