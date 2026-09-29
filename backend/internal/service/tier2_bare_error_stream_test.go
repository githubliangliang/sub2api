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
)

func TestTier2BareErrorWaitsForAuthoritativeTerminal(t *testing.T) {
	for _, path := range []string{"native", "passthrough", "ws-http"} {
		for _, recover := range []bool{false, true} {
			t.Run(path+map[bool]string{false: "/eof", true: "/completed"}[recover], func(t *testing.T) {
				body := "data: " + `{"type":"response.created","response":{"id":"resp_recovered","status":"in_progress"}}` + "\n\n" +
					"event: error\ndata: " + `{"type":"error","error":{"code":"transient","message":"retrying"}}` + "\n\n"
				if recover {
					body += "data: " + `{"type":"response.completed","response":{"id":"resp_recovered","status":"completed","usage":{"input_tokens":8,"output_tokens":4}}}` + "\n\n"
				}
				resp := &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				account := &Account{ID: 113, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1}
				svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}, httpUpstream: &httpUpstreamRecorder{resp: resp}}
				var err error
				var tokens int
				switch path {
				case "native":
					var result *openaiStreamingResult
					result, err = svc.handleStreamingResponse(context.Background(), resp, c, account, time.Now(), "gpt-5.5", "gpt-5.5")
					if result != nil {
						tokens = result.usage.InputTokens
					}
				case "passthrough":
					var result *openaiStreamingResultPassthrough
					result, err = svc.handleStreamingResponsePassthrough(context.Background(), resp, c, account, time.Now(), "gpt-5.5", "gpt-5.5")
					if result != nil {
						tokens = result.usage.InputTokens
					}
				case "ws-http":
					payload := []byte(`{"type":"response.create","model":"gpt-5.5","input":"hi"}`)
					var result *OpenAIForwardResult
					result, err = svc.proxyOpenAIWSHTTPBridgeTurn(context.Background(), c, account, "fixture", payload, len(payload), "gpt-5.5", "", "", "", "", 2, func(data []byte) error { _, e := rec.Write(data); return e })
					if result != nil {
						tokens = result.Usage.InputTokens
					}
				}
				if recover {
					require.NoError(t, err)
					require.Equal(t, 8, tokens)
					require.Contains(t, rec.Body.String(), "response.completed")
					require.NotContains(t, rec.Body.String(), "retrying", "superseded error must not reach the client")
				} else {
					require.Error(t, err)
					require.Contains(t, rec.Body.String(), "response.failed")
					require.Equal(t, 1, strings.Count(rec.Body.String(), `"message":"retrying"`))
				}
			})
		}
	}
}
