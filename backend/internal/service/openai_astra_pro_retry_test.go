package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const openAICodexResponsesTestURL = "https://chatgpt.com/backend-api/codex/responses"

// newOpenAIAstraProRetryContext returns a gin context plus its recorder so the
// caller can assert the service-layer client error shape (type/code/param/message).
func newOpenAIAstraProRetryContext(body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, recorder
}

// TestOpenAIGatewayService_OAuthAstraProModeRejectionPassesThrough verifies that
// when the Codex OAuth upstream rejects an Astra pro+max request with a
// deterministic HTTP 400 on error.param=reasoning.mode, the gateway passes the
// client error through unchanged (upstream type/code/param/message, code not
// treated as official), makes exactly one upstream call, keeps reasoning.mode=pro
// and reasoning.effort=max in the sent body, and does not replay without mode.
func TestOpenAIGatewayService_OAuthAstraProModeRejectionPassesThrough(t *testing.T) {
	body := []byte(`{"model":"gpt-6-astra","reasoning":{"mode":"pro","effort":"max"},"input":"hello"}`)
	upstream := &httpUpstreamRecorder{resp: newOpenAIRejectedFieldTestResponse(
		http.StatusBadRequest,
		`{"error":{"type":"invalid_request_error","code":"model_specific_rejection","message":"reasoning.mode is not supported for this model","param":"reasoning.mode"}}`,
	)}
	svc := newOpenAIRejectedFieldTestService(upstream)
	c, recorder := newOpenAIAstraProRetryContext(body)

	_, err := svc.Forward(context.Background(), c, newOpenAIOAuthNamespaceTestAccount(), body)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, upstream.bodies, 1, "deterministic 400 must not trigger a retry")
	require.Len(t, upstream.requests, 1)
	require.Equal(t, openAICodexResponsesTestURL, upstream.requests[0].URL.String())

	sent := upstream.bodies[0]
	require.Equal(t, "gpt-6-astra", gjson.GetBytes(sent, "model").String())
	require.Equal(t, "pro", gjson.GetBytes(sent, "reasoning.mode").String())
	require.Equal(t, "max", gjson.GetBytes(sent, "reasoning.effort").String())

	respJSON := recorder.Body.String()
	require.Equal(t, "invalid_request_error", gjson.Get(respJSON, "error.type").String())
	require.Equal(t, "model_specific_rejection", gjson.Get(respJSON, "error.code").String())
	require.Equal(t, "reasoning.mode", gjson.Get(respJSON, "error.param").String())
	require.Equal(t, "reasoning.mode is not supported for this model", gjson.Get(respJSON, "error.message").String())
}

// Exercise the fork's existing encrypted-content recovery; retrying rejected
// input.status is a separate upstream behavior absent from this fork.
func TestOpenAIGatewayService_OAuthAstraProModeKeptAcrossEncryptedContentRetry(t *testing.T) {
	body := []byte(`{"model":"gpt-6-astra","stream":true,"instructions":"test","reasoning":{"mode":"pro","effort":"max"},"input":[{"type":"reasoning","id":"rs-1","encrypted_content":"stale","summary":[]},{"type":"message","role":"user","content":"hello"}]}`)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusBadRequest, `{"error":{"code":"invalid_encrypted_content","message":"The encrypted content could not be verified"}}`),
		newOpenAIRejectedFieldTestResponse(http.StatusOK, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_ok\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"total_tokens\":2}}}\n\ndata: [DONE]\n\n"),
	}}
	upstream.responses[1].Header.Set("Content-Type", "text/event-stream")
	svc := newOpenAIRejectedFieldTestService(upstream)
	c, _ := newOpenAIAstraProRetryContext(body)

	result, err := svc.Forward(context.Background(), c, newOpenAIOAuthNamespaceTestAccount(), body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.bodies, 2)
	require.Len(t, upstream.requests, 2)
	for i := range upstream.bodies {
		require.Equal(t, openAICodexResponsesTestURL, upstream.requests[i].URL.String(), "attempt %d URL", i)
		sent := upstream.bodies[i]
		require.Equal(t, "gpt-6-astra", gjson.GetBytes(sent, "model").String(), "attempt %d model", i)
		require.Equal(t, "pro", gjson.GetBytes(sent, "reasoning.mode").String(), "attempt %d reasoning.mode", i)
		require.Equal(t, "max", gjson.GetBytes(sent, "reasoning.effort").String(), "attempt %d reasoning.effort", i)
	}
	require.Equal(t, "stale", gjson.GetBytes(upstream.bodies[0], "input.0.encrypted_content").String())
	require.False(t, gjson.GetBytes(upstream.bodies[1], "input.0.encrypted_content").Exists())
}
