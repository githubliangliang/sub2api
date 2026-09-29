package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTier2AntigravityMalformedStreamOnlyRetriesWithoutContent(t *testing.T) {
	for _, tc := range []struct {
		name, part string
		empty      bool
	}{
		{"malformed-empty", ``, true},
		{"signature-only", `{"thought":true,"thoughtSignature":"signature"}`, true},
		{"text", `{"text":"answer"}`, false},
		{"tool", `{"functionCall":{"name":"lookup","args":{"q":"test"}}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			payload := `data: {"response":{"candidates":[{"content":{"role":"model","parts":[` + tc.part + `]},"finishReason":"MALFORMED_FUNCTION_CALL"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":2}}}` + "\n\n"
			response := &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(payload))}
			result, err := (&AntigravityGatewayService{}).handleAntigravityCompatStream(c, response, time.Now(), "gemini-3.8-flash", newAntigravityResponsesStreamAdapter("gemini-3.8-flash"), "test")
			if tc.empty {
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Nil(t, result)
				require.Empty(t, recorder.Body.String())
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotEmpty(t, recorder.Body.String())
			}
		})
	}
}

func TestTier2AntigravityBareModelFallback(t *testing.T) {
	svc := &AntigravityGatewayService{}
	account := &Account{Platform: PlatformAntigravity, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-3.8-flash-high": "gemini-3.8-flash-high"}}}
	require.Equal(t, "gemini-3.8-flash-high", svc.getMappedModel(account, "gemini-3.8-flash"))
}

func TestTier2AntigravityControlEventsAreNotContent(t *testing.T) {
	for _, event := range []*apicompat.AnthropicStreamEvent{
		{Type: "message_stop"},
		{Type: "content_block_delta", Delta: &apicompat.AnthropicDelta{Type: "signature_delta", Signature: "signature"}},
		{Type: "message_delta", Delta: &apicompat.AnthropicDelta{StopReason: "end_turn"}},
	} {
		require.False(t, isMeaningfulAntigravityCompatEvent(event))
	}
	for _, event := range []*apicompat.AnthropicStreamEvent{
		{Type: "content_block_delta", Delta: &apicompat.AnthropicDelta{Type: "text_delta", Text: "answer"}},
		{Type: "content_block_start", ContentBlock: &apicompat.AnthropicContentBlock{Type: "tool_use", Name: "lookup"}},
	} {
		require.True(t, isMeaningfulAntigravityCompatEvent(event))
	}
}
