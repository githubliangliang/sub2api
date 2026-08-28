package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayCompatPoolMode429AllowsSameAccountRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		path string
		body []byte
		call func(*GatewayService, context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
	}{
		{
			name: "chat completions",
			path: "/v1/chat/completions",
			body: []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
			},
		},
		{
			name: "responses",
			path: "/v1/responses",
			body: []byte(`{"model":"claude-sonnet-4-5","input":"hello"}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsResponses(ctx, c, account, body, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"X-Request-Id": []string{"pool-429"}},
				Body:       io.NopCloser(http.NoBody),
			}}}
			svc := &GatewayService{
				cfg:                 &config.Config{},
				httpUpstream:        upstream,
				tlsFPProfileService: &TLSFingerprintProfileService{},
			}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
			account := &Account{
				ID:       1,
				Name:     "pool-account",
				Platform: PlatformAnthropic,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":   "test-key",
					"pool_mode": true,
				},
			}

			result, err := tt.call(svc, context.Background(), c, account, tt.body)
			require.Nil(t, result)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.Equal(t, http.StatusTooManyRequests, failoverErr.StatusCode)
			require.True(t, failoverErr.RetryableOnSameAccount)
			require.Equal(t, 1, upstream.callCount)
			require.Empty(t, recorder.Body.String())
		})
	}
}

func TestPoolModeRetryableOnSameAccountFlag(t *testing.T) {
	pool := &Account{
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":   "test-key",
			"pool_mode": true,
		},
	}
	nonPool := &Account{
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "test-key",
		},
	}

	require.True(t, poolModeRetryableOnSameAccount(pool, http.StatusTooManyRequests, false))
	require.False(t, poolModeRetryableOnSameAccount(pool, http.StatusTooManyRequests, true), "should-disable must suppress same-account retry")
	require.False(t, poolModeRetryableOnSameAccount(nonPool, http.StatusTooManyRequests, false), "non-pool must stay false")
	require.False(t, poolModeRetryableOnSameAccount(pool, http.StatusInternalServerError, false), "non-retryable status must stay false")
	require.False(t, poolModeRetryableOnSameAccount(nil, http.StatusTooManyRequests, false))
}

func TestGatewayCompatPoolModeRetryNegativeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type scenario struct {
		name      string
		status    int
		poolMode  bool
		wantRetry bool
	}
	scenarios := []scenario{
		{name: "non-pool 429", status: http.StatusTooManyRequests, poolMode: false, wantRetry: false},
		{name: "pool 500 not retryable", status: http.StatusInternalServerError, poolMode: true, wantRetry: false},
	}
	paths := []struct {
		name string
		path string
		body []byte
		call func(*GatewayService, context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
	}{
		{
			name: "chat completions",
			path: "/v1/chat/completions",
			body: []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
			},
		},
		{
			name: "responses",
			path: "/v1/responses",
			body: []byte(`{"model":"claude-sonnet-4-5","input":"hello"}`),
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsResponses(ctx, c, account, body, nil)
			},
		},
	}

	for _, sc := range scenarios {
		for _, tt := range paths {
			t.Run(sc.name+"/"+tt.name, func(t *testing.T) {
				upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
					StatusCode: sc.status,
					Header:     http.Header{"X-Request-Id": []string{"pool-neg"}},
					Body:       io.NopCloser(http.NoBody),
				}}}
				svc := &GatewayService{
					cfg:                 &config.Config{},
					httpUpstream:        upstream,
					tlsFPProfileService: &TLSFingerprintProfileService{},
				}
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
				creds := map[string]any{"api_key": "test-key"}
				if sc.poolMode {
					creds["pool_mode"] = true
				}
				account := &Account{
					ID:          1,
					Name:        "pool-account",
					Platform:    PlatformAnthropic,
					Type:        AccountTypeAPIKey,
					Credentials: creds,
				}

				result, err := tt.call(svc, context.Background(), c, account, tt.body)
				require.Nil(t, result)
				var failoverErr *UpstreamFailoverError
				require.ErrorAs(t, err, &failoverErr)
				require.Equal(t, sc.status, failoverErr.StatusCode)
				require.Equal(t, sc.wantRetry, failoverErr.RetryableOnSameAccount)
			})
		}
	}
}
