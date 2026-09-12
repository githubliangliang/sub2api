package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardOpenAIWSV2_ClientCancellationPreservesCompletedImageUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	writer := &cancelOnFirstWriteResponseWriter{cancel: cancel}
	c, _ := gin.CreateTestContext(writer)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	c.Request.Header.Set("User-Agent", "unit-test-agent/1.0")
	conn := &openAIWSCancelSafeConn{openAIWSCaptureConn: &openAIWSCaptureConn{
		events: [][]byte{
			[]byte(`{"type":"response.created","response":{"id":"resp_image_cancel","model":"gpt-5.5"}}`),
			[]byte(`{"type":"response.output_text.delta","delta":"partial"}`),
			[]byte(`{"type":"response.output_item.done","item":{"id":"image_1","type":"image_generation_call","result":"finished-image","size":"1024x1024"}}`),
		},
	}}
	cfg := newOpenAIWSV2TestConfig()
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = true
	cfg.Gateway.OpenAIWS.IngressModeDefault = OpenAIWSIngressModeCtxPool
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 1
	pool := newOpenAIWSConnPool(cfg)
	defer pool.Close()
	pool.setClientDialerForTest(&openAIWSClientConnCancelDialer{conn: conn})
	svc := &OpenAIGatewayService{
		cfg: cfg, httpUpstream: &httpUpstreamRecorder{}, cache: &stubGatewayCache{},
		openaiWSResolver: NewOpenAIWSProtocolResolver(cfg), toolCorrector: NewCodexToolCorrector(), openaiWSPool: pool,
	}
	account := &Account{
		ID: 9102, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Status: StatusActive, Schedulable: true, Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra:       map[string]any{"openai_apikey_responses_websockets_v2_mode": OpenAIWSIngressModeCtxPool},
	}

	result, err := svc.Forward(ctx, c, account, []byte(`{"model":"gpt-5.5","stream":true,"input":[{"type":"input_text","text":"draw"}]}`))

	require.ErrorIs(t, err, context.Canceled)
	require.NotNil(t, result)
	require.True(t, result.ClientDisconnect)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, "resp_image_cancel", result.RequestID)
}
