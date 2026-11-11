package service

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIRequestPatches_ForwardReleasesPatchedBodyWhenAttemptEnds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5","stream":false,"reasoning":{"effort":"minimal"},"max_tokens":1024,"input":[{"type":"message","content":"hi"}]}`)
	const limit = 1 << 20
	budget := httputil.NewBodyMemoryBudget(limit)
	ctx, release := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	t.Cleanup(release)
	for attempt := range 100 {
		upstream := &httpUpstreamRecorder{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":1,"output_tokens":2}}`)),
			},
		}
		cfg := &config.Config{}
		cfg.Security.URLAllowlist.Enabled = false
		svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream}
		account := &Account{
			ID:          int64(attempt + 1),
			Name:        "openai-apikey",
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Concurrency: 1,
			Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://example.com"},
			Extra:       map[string]any{"use_responses_api": true},
		}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil).WithContext(ctx)
		SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

		_, err := svc.Forward(ctx, c, account, body)
		require.NoError(t, err)
		require.NotNil(t, upstream.lastReq)
		// Failover attempts must not accumulate rewritten bodies until the handler returns.
		probe, releaseProbe := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
		require.NoError(t, httputil.ReserveRequestBodyMemory(probe, limit), "attempt %d retained its patched body", attempt)
		releaseProbe()
	}
}

func TestOpenAIRequestPatches_ForwardStopsBeforeUpstreamOnOverload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name string
		body string
	}{
		{
			name: "final raw patch flush",
			body: `{"model":"gpt-5","stream":false,"reasoning":{"effort":"minimal"},"max_tokens":1024,"input":[{"type":"message","content":"hi"}]}`,
		},
		{
			name: "flush before schema decode",
			body: `{"model":"client-alias","stream":false,"reasoning":{"effort":"minimal"},"text":{"format":{"type":"json_schema","name":"test","schema":{"type":"object","properties":{}}}},"input":[{"type":"message","content":"hi"}]}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"usage":{"input_tokens":1,"output_tokens":2}}`)),
				},
			}
			cfg := &config.Config{}
			cfg.Security.URLAllowlist.Enabled = false
			svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream}
			account := &Account{
				ID:          1,
				Name:        "openai-apikey",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{
					"api_key":       "sk-test",
					"base_url":      "https://example.com",
					"model_mapping": map[string]any{"client-alias": "gpt-5"},
				},
				Extra: map[string]any{"use_responses_api": true},
			}
			ctx, release := httputil.WithRequestBodyMemoryBudget(context.Background(), httputil.NewBodyMemoryBudget(1))
			t.Cleanup(release)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil).WithContext(ctx)
			SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)

			result, err := svc.Forward(ctx, c, account, []byte(tt.body))
			require.ErrorIs(t, err, httputil.ErrRequestBodyMemoryBudget)
			require.Nil(t, result)
			require.Nil(t, upstream.lastReq, "overload must not fall back to unbudgeted serialization or forwarding")
		})
	}
}
