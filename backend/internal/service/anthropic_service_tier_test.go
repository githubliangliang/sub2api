package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAnthropicForwardRecordsFastServiceTier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		for _, streaming := range []bool{false, true} {
			for _, tc := range []struct{ reported, expected string }{{"fast", "fast"}, {"", "fast"}, {"standard", "standard"}} {
				t.Run(fmt.Sprintf("passthrough=%t/stream=%t/reported=%s", passthrough, streaming, tc.reported), func(t *testing.T) {
					body := []byte(fmt.Sprintf(`{"model":"claude-opus-5","max_tokens":64,"stream":%t,"speed":"fast","messages":[{"role":"user","content":"hello"}]}`, streaming))
					parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), PlatformAnthropic)
					require.NoError(t, err)
					payload := fmt.Sprintf(`{"id":"msg_fast","type":"message","model":"claude-opus-5","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":11,"output_tokens":3,"speed":%q}}`, tc.reported)
					contentType := "application/json"
					if streaming {
						contentType = "text/event-stream"
						payload = "data: {\"type\":\"message_start\",\"message\":" + payload + "}\n\n" +
							"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":3}}\n\n" +
							"data: {\"type\":\"message_stop\"}\n\n"
					}
					upstream := &anthropicHTTPUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK,
						Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(payload))}}
					account := newAnthropicOAuthAccountForPartialUsageTest()
					if passthrough {
						account.Type = AccountTypeAPIKey
						account.Credentials = map[string]any{"api_key": "fixture", "base_url": "https://api.example.test"}
						account.Extra = map[string]any{"anthropic_passthrough": true}
					}
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
					result, err := newForwardPartialUsageServiceForTest(upstream).Forward(context.Background(), c, account, parsed)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, "fast", gjson.GetBytes(upstream.lastBody, "speed").String())
					// Exercise the actual log builder without changing the token-cost policy.
					log := (&GatewayService{}).buildRecordUsageLog(context.Background(), &recordUsageCoreInput{}, result,
						&APIKey{ID: 1}, &User{ID: 1}, account, nil, result.Model, 1, 1, 1, 0, false,
						&CostBreakdown{TotalCost: 0.25, ActualCost: 0.25}, &recordUsageOpts{})
					require.NotNil(t, log.ServiceTier, "forwarding speed=fast must reach the persisted usage log")
					require.Equal(t, tc.expected, *log.ServiceTier)
					require.Equal(t, 0.25, log.ActualCost)
				})
			}
		}
	}
}

func TestAnthropicServiceSpeedObservation(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint("passthrough=", passthrough), func(t *testing.T) {
			usage := &ClaudeUsage{}
			parse := (&GatewayService{}).parseSSEUsage
			if passthrough {
				parse = (&GatewayService{}).parseSSEUsagePassthrough
			}
			parse(`{"type":"message_start","message":{"usage":{"input_tokens":10,"speed":" FAST "}}}`, usage)
			require.Equal(t, "fast", usage.Speed)
			parse(`{"type":"message_delta","usage":{"output_tokens":3}}`, usage)
			require.Equal(t, "fast", usage.Speed, "a frame without speed must retain the observed value")
			parse(`{"type":"message_delta","usage":{"speed":"standard","output_tokens":4}}`, usage)
			require.Equal(t, "standard", usage.Speed, "the explicit provider speed must replace the request/earlier observation")
			parse(`{"type":"ping","usage":{"speed":"fast"}}`, usage)
			parse(`{"type":"message_delta","usage":{"speed":"unknown"}}`, usage)
			require.Equal(t, "standard", usage.Speed)
		})
	}
	for _, value := range []string{"priority", "flex", "turbo", "", "null"} {
		require.Empty(t, normalizeAnthropicSpeed(value))
	}
}
