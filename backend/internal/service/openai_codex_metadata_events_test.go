package service

import (
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

func TestRelayStallsMetadataBeforeCapacityFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, metadata := range []string{
		`{"type":"response.metadata","response_id":"r","metadata":{"route":"test"}}`,
		`{"type":"codex.rate_limits","rate_limits":{}}`,
		`{"type":"codex.response.metadata","response_id":"r","metadata":{}}`,
	} {
		for _, passthrough := range []bool{false, true} {
			name := metadata
			if passthrough {
				name += "/passthrough"
			}
			t.Run(name, func(t *testing.T) {
				svc := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}}}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
				stream := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"r\"}}\n\n" +
					"data: " + metadata + "\n\n" +
					"data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"server_is_overloaded\",\"message\":\"Our servers are currently overloaded. Please try again later.\"}}}\n\n"
				resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(stream))}
				account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
				var err error
				if passthrough {
					var result *openaiStreamingResultPassthrough
					result, err = svc.handleStreamingResponsePassthrough(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
					require.NotNil(t, result)
					require.Nil(t, result.firstTokenMs)
				} else {
					var result *openaiStreamingResult
					result, err = svc.handleStreamingResponse(c.Request.Context(), resp, c, account, time.Now(), "model", "model")
					require.NotNil(t, result)
					require.Nil(t, result.firstTokenMs)
				}
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.False(t, c.Writer.Written())
				require.Empty(t, rec.Body.String(), "metadata must stay attempt-local until real output")
				require.True(t, failover.RequestScopedTransient)
			})
		}
	}
}

func TestCodexMetadataClassificationRetainsRealOutput(t *testing.T) {
	for _, kind := range []string{"response.output_text.delta", "response.output_item.added", "response.function_call_arguments.delta", "response.completed", "response.done", "codex.custom_output"} {
		require.False(t, openAIStreamEventIsMetadata(kind), kind)
	}
	for _, kind := range []string{"codex.rate_limits", "codex.response.metadata", "response.metadata"} {
		require.True(t, openAIStreamEventIsMetadata(kind), kind)
	}
}
