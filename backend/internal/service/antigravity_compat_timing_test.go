package service

import (
	"context"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestAntigravityCompatGeminiThinkingConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("streaming responses tier suffix enables includeThoughts without thinkingLevel", func(t *testing.T) {
		upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{antigravityCompatSuccessResponse()}}
		svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
		account := newAntigravityCompatAccount(AccountTypeOAuth)
		mapping, ok := account.Credentials["model_mapping"].(map[string]any)
		require.True(t, ok)
		mapping["gemini-3.8-flash-high"] = "gemini-3.8-flash-high"

		body := []byte(`{"model":"gemini-3.8-flash-high","stream":true,"input":"inspect repository"}`)
		c, recorder := newAntigravityCompatContext(http.MethodPost, "/v1/responses", body)

		_, err := svc.ForwardAsResponses(context.Background(), c, account, body, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Len(t, upstream.requestBodies, 1)

		cfg := gjson.GetBytes(upstream.requestBodies[0], "request.generationConfig.thinkingConfig")
		require.True(t, cfg.Get("includeThoughts").Bool(), "thinkingConfig=%s", cfg.Raw)
		require.False(t, cfg.Get("thinkingLevel").Exists(), "tier-suffixed model must not set thinkingLevel: %s", cfg.Raw)
		require.False(t, cfg.Get("thinkingBudget").Exists(), "tier-suffixed model must not set thinkingBudget: %s", cfg.Raw)
	})

	t.Run("streaming chat completions tier suffix enables includeThoughts without thinkingLevel", func(t *testing.T) {
		upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{antigravityCompatSuccessResponse()}}
		svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
		account := newAntigravityCompatAccount(AccountTypeOAuth)
		mapping, ok := account.Credentials["model_mapping"].(map[string]any)
		require.True(t, ok)
		mapping["gemini-3.8-flash-high"] = "gemini-3.8-flash-high"

		body := []byte(`{"model":"gemini-3.8-flash-high","stream":true,"messages":[{"role":"user","content":"inspect repository"}]}`)
		c, recorder := newAntigravityCompatContext(http.MethodPost, "/v1/chat/completions", body)

		_, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Len(t, upstream.requestBodies, 1)

		cfg := gjson.GetBytes(upstream.requestBodies[0], "request.generationConfig.thinkingConfig")
		require.True(t, cfg.Get("includeThoughts").Bool(), "thinkingConfig=%s", cfg.Raw)
		require.False(t, cfg.Get("thinkingLevel").Exists(), "thinkingConfig=%s", cfg.Raw)
	})

	t.Run("streaming responses bare model with low effort selects low variant and enables includeThoughts", func(t *testing.T) {
		upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{antigravityCompatSuccessResponse()}}
		svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
		account := newAntigravityCompatAccount(AccountTypeOAuth)
		mapping, ok := account.Credentials["model_mapping"].(map[string]any)
		require.True(t, ok)
		mapping["gemini-3.8-flash-low"] = "gemini-3.8-flash-low"
		mapping["gemini-3.8-flash-high"] = "gemini-3.8-flash-high"

		body := []byte(`{"model":"gemini-3.8-flash","stream":true,"reasoning":{"effort":"low"},"input":"quick check"}`)
		c, recorder := newAntigravityCompatContext(http.MethodPost, "/v1/responses", body)

		result, err := svc.ForwardAsResponses(context.Background(), c, account, body, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Len(t, upstream.requestBodies, 1)
		require.Equal(t, "gemini-3.8-flash-low", gjson.GetBytes(upstream.requestBodies[0], "model").String())
		require.Equal(t, "gemini-3.8-flash-low", result.UpstreamModel)

		cfg := gjson.GetBytes(upstream.requestBodies[0], "request.generationConfig.thinkingConfig")
		require.True(t, cfg.Get("includeThoughts").Bool(), "thinkingConfig=%s", cfg.Raw)
		require.False(t, cfg.Get("thinkingLevel").Exists(), "thinkingConfig=%s", cfg.Raw)
	})

	t.Run("streaming responses explicit bare model sets thinkingLevel and includeThoughts", func(t *testing.T) {
		upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{antigravityCompatSuccessResponse()}}
		svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
		account := newAntigravityCompatAccount(AccountTypeOAuth)
		mapping, ok := account.Credentials["model_mapping"].(map[string]any)
		require.True(t, ok)
		mapping["gemini-3.5-flash"] = "gemini-3.5-flash"

		body := []byte(`{"model":"gemini-3.5-flash","stream":true,"reasoning":{"effort":"medium"},"input":"check"}`)
		c, recorder := newAntigravityCompatContext(http.MethodPost, "/v1/responses", body)

		_, err := svc.ForwardAsResponses(context.Background(), c, account, body, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Len(t, upstream.requestBodies, 1)

		cfg := gjson.GetBytes(upstream.requestBodies[0], "request.generationConfig.thinkingConfig")
		require.Equal(t, "medium", cfg.Get("thinkingLevel").String(), "thinkingConfig=%s", cfg.Raw)
		require.True(t, cfg.Get("includeThoughts").Bool(), "thinkingConfig=%s", cfg.Raw)
	})

	t.Run("non-streaming responses do not force includeThoughts", func(t *testing.T) {
		upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{antigravityCompatSuccessResponse()}}
		svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
		account := newAntigravityCompatAccount(AccountTypeOAuth)
		mapping, ok := account.Credentials["model_mapping"].(map[string]any)
		require.True(t, ok)
		mapping["gemini-3.8-flash-high"] = "gemini-3.8-flash-high"

		nonStreamBody := []byte(`{"model":"gemini-3.8-flash-high","stream":false,"input":"inspect"}`)
		c, rec := newAntigravityCompatContext(http.MethodPost, "/v1/responses", nonStreamBody)
		_, err := svc.ForwardAsResponses(context.Background(), c, account, nonStreamBody, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Len(t, upstream.requestBodies, 1)
		require.False(t, gjson.GetBytes(upstream.requestBodies[0], "request.generationConfig.thinkingConfig.includeThoughts").Bool())
	})
}

func TestAntigravityCompatHeaderWaitExceedingPreContentLimitDoesNotAbortStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []antigravityCompatProtocol{antigravityCompatResponses, antigravityCompatChatCompletions} {
		name := "responses"
		path := "/v1/responses"
		wantDone := "event: response.completed"
		if protocol == antigravityCompatChatCompletions {
			name = "chat_completions"
			path = "/v1/chat/completions"
			wantDone = "data: [DONE]"
		}
		for _, committedKeepalive := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/keepalive=%t", name, committedKeepalive), func(t *testing.T) {
				svc := newAntigravityCompatService(config.GatewayConfig{
					MaxLineSize:               defaultMaxLineSize,
					StreamDataIntervalTimeout: 180,
					StreamKeepaliveInterval:   10,
				}, nil)
				c, recorder := newAntigravityCompatContext(http.MethodPost, path, nil)
				if committedKeepalive {
					c.Header("Content-Type", "text/event-stream")
					c.Status(http.StatusOK)
					MarkResponseCommitted(c)
					_, err := c.Writer.Write([]byte(": ping\n\n"))
					require.NoError(t, err)
					c.Writer.Flush()
				}

				requestStart := time.Now().Add(-3 * time.Minute)
				call := &antigravityCompatUpstreamCall{
					request: antigravityCompatRequest{
						protocol:      protocol,
						originalModel: "gemini-3.8-flash-high",
						clientStream:  true,
						startTime:     requestStart,
					},
					billingModel: "gemini-3.8-flash-high",
					prefix:       "test-long-header-wait",
				}

				result, err := svc.consumeAntigravityCompatResponse(
					c.Request.Context(),
					c,
					newAntigravityCompatAccount(AccountTypeOAuth),
					call,
					antigravityCompatSuccessResponse(),
				)
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, result.FirstTokenMs)
				require.GreaterOrEqual(t, *result.FirstTokenMs, int((3*time.Minute - time.Second).Milliseconds()))
				require.Contains(t, recorder.Body.String(), wantDone)
				require.NotContains(t, recorder.Body.String(), "stream_timeout")
			})
		}
	}
}

func TestAntigravityCompatInterleavedThoughtChunksResetIdleTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newAntigravityCompatService(config.GatewayConfig{
		MaxLineSize:               defaultMaxLineSize,
		StreamDataIntervalTimeout: 1,
	}, nil)
	c, recorder := newAntigravityCompatContext(http.MethodPost, "/v1/responses", nil)
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       reader,
	}

	done := make(chan error, 1)
	go func() {
		_, err := svc.handleResponsesStreamingFromAntigravity(c, resp, time.Now(), "gemini-3.8-flash-high")
		done <- err
	}()

	_, err := io.WriteString(writer, `data: {"response":{"responseId":"resp_44276","candidates":[{"content":{"role":"model","parts":[{"text":"Checking files..."}]}}]}}`+"\n\n")
	require.NoError(t, err)
	for i := 0; i < 4; i++ {
		time.Sleep(350 * time.Millisecond)
		_, err = io.WriteString(writer, fmt.Sprintf(`data: {"response":{"responseId":"resp_44276","candidates":[{"content":{"role":"model","parts":[{"thought":true,"text":"step %d "}]}}]}}`+"\n\n", i+1))
		require.NoError(t, err)
	}
	_, err = io.WriteString(writer, `data: {"response":{"responseId":"resp_44276","candidates":[{"content":{"role":"model","parts":[{"thoughtSignature":"sig_done","functionCall":{"name":"exec_command","args":{"cmd":"ls"},"id":"call_1"}}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":100,"candidatesTokenCount":20,"thoughtsTokenCount":50}}}`+"\n\n")
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	require.NoError(t, <-done)
	out := recorder.Body.String()
	require.Contains(t, out, "event: response.reasoning_summary_text.delta")
	require.Contains(t, out, "event: response.completed")
	require.NotContains(t, out, "stream_timeout")
}
