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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSSEParseGateEscapedMetadata(t *testing.T) {
	for _, payload := range []string{
		`{"ty\u0070e":"message_stop"}`,
		`{"type":"content_block_delta","us\u0061ge":{"cache_creation_input_tokens":7}}`,
		`{"type":"content_block_delta","m\u006fdel":"claude-sonnet-4-6"}`,
	} {
		require.True(t, sseEventNeedsJSONParse("content_block_delta", []byte(payload), false, false))
	}
	require.False(t, sseEventNeedsJSONParse("content_block_delta", []byte(`{"type":"content_block_delta","delta":{"type":"text_delta","text":"plain"}}`), false, false))

	gin.SetMode(gin.TestMode)
	for _, keepalive := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "claude-code"}[keepalive], func(t *testing.T) {
			body := "event: content_block_delta\ndata: " + `{"ty\u0070e":"message_start","message":{"m\u006fdel":"model","us\u0061ge":{"input_tokens":3,"cache_creation":{"ephemeral_5m_input_tokens":2,"ephemeral_1h_input_tokens":5}}}}` + "\n\n" +
				"event: content_block_delta\ndata: " + `{"ty\u0070e":"message_delta","us\u0061ge":{"output_tokens":4}}` + "\n\n" +
				"event: content_block_delta\ndata: " + `{"ty\u0070e":"message_stop"}` + "\n\n"
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			if keepalive {
				c.Request.Header.Set("User-Agent", "claude-cli/2.1.321 (external, cli)")
			}
			resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
			result, err := newMinimalGatewayService().handleStreamingResponse(c.Request.Context(), resp, c, &Account{ID: 1}, time.Now(), "model", "model", false)
			require.NoError(t, err)
			require.Equal(t, body, w.Body.String())
			require.Equal(t, 3, result.usage.InputTokens)
			require.Equal(t, 4, result.usage.OutputTokens)
			require.Equal(t, 7, result.usage.CacheCreationInputTokens)
			require.Equal(t, 2, result.usage.CacheCreation5mTokens)
			require.Equal(t, 5, result.usage.CacheCreation1hTokens)
			require.Equal(t, "model", observedUpstreamResponseModel(c))
		})
	}
}

func BenchmarkS22NativeSSEMap(b *testing.B) {
	gin.SetMode(gin.TestMode)
	body := "event: message_start\ndata: " + `{"type":"message_start","message":{"model":"model","usage":{"input_tokens":3}}}` + "\n\n" +
		strings.Repeat("event: content_block_delta\ndata: "+`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"plain text"}}`+"\n\n", 100) +
		"event: message_delta\ndata: " + `{"type":"message_delta","usage":{"output_tokens":100}}` + "\n\nevent: message_stop\ndata: " + `{"type":"message_stop"}` + "\n\n"
	svc := newMinimalGatewayService()
	b.ReportAllocs()
	for b.Loop() {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
		if _, err := svc.handleStreamingResponse(context.Background(), resp, c, &Account{ID: 1}, time.Now(), "model", "model", false); err != nil {
			b.Fatal(err)
		}
	}
}
