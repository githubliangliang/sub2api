//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpstream024AlphaSearch403Scope(t *testing.T) {
	for _, pat := range []bool{false, true} {
		for _, tc := range []struct {
			body       string
			wantWrites int
		}{
			{`{"detail":"Literature search is unavailable."}`, 0},
			{"", 0},
			{`{"error":{"code":"deactivated_workspace","message":"Workspace is deactivated"}}`, 1},
		} {
			t.Run(tc.body+"/PAT="+map[bool]string{false: "false", true: "true"}[pat], func(t *testing.T) {
				body := []byte(`{"id":"search-session","model":"gpt-6.2-sol","commands":{"literature_search_query":[{"q":"protein folding"}]}}`)
				c, rec := gin.CreateTestContext(httptest.NewRecorder())
				_ = rec
				c.Request = httptest.NewRequest(http.MethodPost, "/alpha/search", bytes.NewReader(body))
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}}
				repo := &alphaSearchAccountStateRepo{}
				cfg := &config.Config{}
				svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream, accountRepo: repo, rateLimitService: NewRateLimitService(repo, nil, cfg, nil, nil)}
				credentials := map[string]any{"access_token": "synthetic-token", "chatgpt_account_id": "synthetic-account"}
				if pat {
					credentials["auth_mode"] = OpenAIAuthModePersonalAccessToken
				}
				account := &Account{ID: 45, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Concurrency: 1, Credentials: credentials}
				result, err := svc.ForwardAlphaSearch(context.Background(), c, account, body)
				require.Nil(t, result)
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, http.StatusForbidden, failover.StatusCode)
				require.Equal(t, tc.wantWrites, repo.setErrorCalls)
				require.False(t, c.Writer.Written())
			})
		}
	}
}

func TestUpstream024RequestTimeoutFailover(t *testing.T) {
	svc := &OpenAIGatewayService{}
	for _, tc := range []struct {
		status int
		body   string
		want   bool
	}{
		{408, `{"error":{"message":"Model is overloaded, please retry later"}}`, true},
		{408, `Request Timeout`, true},
		{408, `{"error":{"code":"context_length_exceeded","message":"Maximum context length exceeded"}}`, false},
		{400, `{"error":{"message":"Invalid image size"}}`, false},
		{429, `{"error":{"message":"Rate limited"}}`, true},
	} {
		require.Equal(t, tc.want, svc.shouldFailoverOpenAIUpstreamResponse(tc.status, extractUpstreamErrorMessage([]byte(tc.body)), []byte(tc.body)), tc.body)
	}
}

func TestUpstream024Images408BeforeOutput(t *testing.T) {
	body := []byte(`{"model":"gpt-image-1","prompt":"draw a cat"}`)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusRequestTimeout, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"Upstream timed out"}}`))}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	result, err := svc.ForwardImages(context.Background(), c, &Account{ID: 51, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "synthetic-key"}}, body, parsed, "")
	require.Nil(t, result)
	var failover *UpstreamFailoverError
	require.ErrorAs(t, err, &failover)
	require.Equal(t, http.StatusRequestTimeout, failover.StatusCode)
	require.False(t, c.Writer.Written())
}
