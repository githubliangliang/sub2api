//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVertexSharedCapacity429ClassificationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, details string
		shared        bool
	}{
		{"absent", "", true},
		{"null", `,"details":null`, true},
		{"empty array", `,"details":[]`, true},
		{"quota failure", `,"details":[{"@type":"type.googleapis.com/google.rpc.QuotaFailure"}]`, false},
		{"retry info", `,"details":[{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"5s"}]`, false},
		{"unknown details", `,"details":[{"kind":"future_error"}]`, false},
		{"object details", `,"details":{"kind":"quota"}`, false},
		{"empty object details", `,"details":{}`, false},
		{"string details", `,"details":"quota"`, false},
		{"boolean details", `,"details":false`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later."` + tc.details + `}}`)
			require.Equal(t, tc.shared, isVertexSharedCapacity429(body))
		})
	}
	for _, body := range []string{
		`{"error":{"status":"RESOURCE_EXHAUSTED","message":"Quota exhausted. Please try again later."}}`,
		`{"error":{"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted: Provisioned Throughput exceeded. Please try again later."}}`,
		`{"error":{"status":"UNAVAILABLE","message":"Resource exhausted. Please try again later."}}`,
		`{"error":{"status":"RESOURCE_EXHAUSTED","message":"backend busy"}}`,
		`{"error":`,
	} {
		require.False(t, isVertexSharedCapacity429([]byte(body)), body)
	}
}

type vertexS07QuotaRepo struct {
	errorPolicyRepoStub
	rateLimitCalls int
}

func (r *vertexS07QuotaRepo) SetRateLimited(context.Context, int64, time.Time) error {
	r.rateLimitCalls++
	return nil
}

func TestGeminiSharedCapacityRetriesWithinRequestWithoutAccountPenalty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	useVertexTokenEndpoint(t, http.StatusOK, `{"access_token":"test-token","expires_in":3600}`)
	for _, fw := range geminiSingleCheckForwarders {
		t.Run(fw.name, func(t *testing.T) {
			repo := &vertexS07QuotaRepo{}
			upstream := &geminiSequenceUpstreamStub{responses: []geminiSequenceUpstreamResponse{
				{status: http.StatusTooManyRequests, body: `{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later."}}`},
				{status: http.StatusOK, body: `{"candidates":[{"content":{"role":"model","parts":[{"text":"OK"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1}}`},
			}}
			svc := &GeminiMessagesCompatService{
				accountRepo: repo, tokenProvider: NewGeminiTokenProvider(repo, nil, nil), httpUpstream: upstream,
				rateLimitService: NewRateLimitService(repo, nil, &config.Config{}, nil, nil), cfg: &config.Config{},
			}
			err, _ := fw.forward(t, svc, newVertexTokenTestAccount(t))
			require.NoError(t, err)
			require.Equal(t, 2, upstream.calls)
			require.Zero(t, repo.rateLimitCalls)
			require.Zero(t, repo.tempCalls)
			require.Zero(t, repo.setErrCalls)
		})
	}
}

func TestGeminiVertexTransientTokenErrorsDoNotDisableAnyEntryPoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tokenStatus := range []int{http.StatusBadRequest, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(tokenStatus), func(t *testing.T) {
			useVertexTokenEndpoint(t, tokenStatus, `{"error":"invalid_grant","error_description":"Invalid JWT Signature."}`)
			for _, fw := range geminiSingleCheckForwarders {
				t.Run(fw.name, func(t *testing.T) {
					svc, upstream, repo := newVertexTokenErrorService()
					err, c := fw.forward(t, svc, newVertexTokenTestAccount(t))
					require.Error(t, err)
					var failover *UpstreamFailoverError
					require.False(t, errors.As(err, &failover))
					require.Zero(t, repo.setErrCalls)
					require.Zero(t, repo.tempCalls)
					require.Zero(t, upstream.calls)
					require.Equal(t, http.StatusBadGateway, c.Writer.Status())
				})
			}
		})
	}
}

func TestGeminiVertexCanceledTokenFetchNeverDisablesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	useVertexTokenEndpoint(t, http.StatusBadRequest, vertexAccountNotFoundTokenBody)
	for _, protocol := range []string{"native", "messages", "chat"} {
		t.Run(protocol, func(t *testing.T) {
			svc, upstream, repo := newVertexTokenErrorService()
			account := newVertexTokenTestAccount(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			body := []byte(`{"model":"gemini-2.5-flash","max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(ctx)
			var err error
			switch protocol {
			case "native":
				_, err = svc.ForwardNative(ctx, c, account, "gemini-2.5-flash", "generateContent", false, []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
			case "messages":
				_, err = svc.Forward(ctx, c, account, body)
			case "chat":
				_, err = svc.ForwardAsChatCompletions(ctx, c, account, body)
			}
			require.ErrorIs(t, err, context.Canceled)
			require.Zero(t, repo.setErrCalls)
			require.Zero(t, repo.tempCalls)
			require.Zero(t, upstream.calls)
			require.False(t, c.Writer.Written())
		})
	}
}

func TestGeminiVertexCancellationDuringTokenFetchNeverDisablesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	entered := make(chan struct{}, 3)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		entered <- struct{}{}
		select {
		case <-req.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })
	original := vertexServiceAccountTokenURL
	vertexServiceAccountTokenURL = server.URL
	t.Cleanup(func() { vertexServiceAccountTokenURL = original })
	for _, protocol := range []string{"native", "messages", "chat"} {
		t.Run(protocol, func(t *testing.T) {
			svc, upstream, repo := newVertexTokenErrorService()
			account := newVertexTokenTestAccount(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			go func() {
				select {
				case <-entered:
					cancel()
				case <-ctx.Done():
				}
			}()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			body := []byte(`{"model":"gemini-2.5-flash","max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(ctx)
			var err error
			switch protocol {
			case "native":
				_, err = svc.ForwardNative(ctx, c, account, "gemini-2.5-flash", "generateContent", false, []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
			case "messages":
				_, err = svc.Forward(ctx, c, account, body)
			case "chat":
				_, err = svc.ForwardAsChatCompletions(ctx, c, account, body)
			}
			require.ErrorIs(t, err, context.Canceled)
			require.Zero(t, repo.setErrCalls)
			require.Zero(t, repo.tempCalls)
			require.Zero(t, upstream.calls)
			require.False(t, c.Writer.Written())
		})
	}
}

func TestVertexTokenURIRejectsCredentialOverride(t *testing.T) {
	require.Equal(t, vertexDefaultTokenURL, vertexServiceAccountTokenURL)
	key, err := parseVertexServiceAccountJSON([]byte(`{"project_id":"project","client_email":"test@example.com","private_key":"test-key","token_uri":"http://169.254.169.254/token"}`))
	require.NoError(t, err)
	require.Equal(t, vertexDefaultTokenURL, key.TokenURI)
}

func TestVertexAccountQuotaAndRetryInfoPreserveCooldown(t *testing.T) {
	for _, tc := range []struct {
		name    string
		details string
		delay   time.Duration
	}{
		{"retry info", `[{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"70s"}]`, 70 * time.Second},
		{"project quota reset", `[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","metadata":{"quotaResetDelay":"2m"}}]`, 2 * time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &rateLimit429AccountRepoStub{}
			svc := &GeminiMessagesCompatService{accountRepo: repo}
			before := time.Now()
			body := []byte(`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later.","details":` + tc.details + `}}`)
			svc.handleGeminiUpstreamError(context.Background(), &Account{ID: 606, Platform: PlatformGemini, Type: AccountTypeServiceAccount}, http.StatusTooManyRequests, http.Header{}, body)
			require.Equal(t, 1, repo.rateLimitCalls)
			require.WithinDuration(t, before.Add(tc.delay), repo.lastRateLimitReset, 2*time.Second)
		})
	}
}

func TestGeminiNonVertexSharedCapacityStillAppliesAccountCooldown(t *testing.T) {
	repo := &rateLimit429AccountRepoStub{}
	svc := &GeminiMessagesCompatService{accountRepo: repo}
	body := []byte(`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later."}}`)
	svc.handleGeminiUpstreamError(context.Background(), &Account{ID: 607, Platform: PlatformGemini, Type: AccountTypeAPIKey}, http.StatusTooManyRequests, http.Header{}, body)
	require.Equal(t, 1, repo.rateLimitCalls)
}
