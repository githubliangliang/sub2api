package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayAnthropicRequestRejectsForeignOAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, platform := range []string{PlatformOpenAI, PlatformGrok, PlatformGemini, PlatformAntigravity} {
		for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
			t.Run(platform+"/"+typ, func(t *testing.T) {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				account := &Account{ID: 25, Platform: platform, Type: typ}
				svc := &GatewayService{cfg: &config.Config{}}
				req, _, err := svc.buildUpstreamRequest(context.Background(), c, account,
					[]byte(`{"model":"gpt-6-astra","messages":[]}`), "synthetic-token", "oauth", "gpt-6-astra", true, false)
				require.Nil(t, req, "foreign credentials must never produce an Anthropic request")
				var failover *UpstreamFailoverError
				require.ErrorAs(t, err, &failover)
				require.Equal(t, GatewayFailureReason("anthropic_messages_foreign_oauth"), failover.Reason)
				require.True(t, failover.ShouldRetryNextAccount())
				require.False(t, failover.RetryableOnSameAccount)
				require.False(t, failover.IsCredentialFailure())
			})
		}
	}
}

func TestGatewayCountTokensRejectsForeignOAuthBeforeTokenFetch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, platform := range []string{PlatformOpenAI, PlatformGrok, PlatformGemini, PlatformAntigravity} {
		t.Run(platform, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)
			account := &Account{ID: 25, Platform: platform, Type: AccountTypeOAuth}
			parsed := &ParsedRequest{Model: "gpt-6-astra", Body: NewRequestBodyRef([]byte(`{"model":"gpt-6-astra","messages":[]}`))}
			// No token provider or HTTP client is installed: neither may be used.
			err := (&GatewayService{}).ForwardCountTokens(context.Background(), c, account, parsed)
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, rec.Code)
			require.Contains(t, rec.Body.String(), "not_found_error")
		})
	}
}

func TestCompositeOwnershipRejectsOAuthAmbiguityDespiteCooldown(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformGrok} {
		for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
			t.Run(platform+"/"+typ, func(t *testing.T) {
				until := time.Now().Add(time.Hour)
				repo := &compositeOwnershipAccountRepo{accounts: []Account{
					{ID: 1, Platform: platform, Type: typ, TempUnschedulableUntil: &until,
						Credentials: map[string]any{"model_mapping": map[string]any{"gpt-6-astra": "gpt-6-astra"}}},
					{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeAPIKey,
						Credentials: map[string]any{"model_mapping": map[string]any{"gpt-6-astra": "claude-sonnet-4-6"}}},
				}}
				repo.schedulableAccounts = repo.accounts[1:]
				svc := &GatewayService{accountRepo: repo}
				resolver := NewCompositeRouteResolver(nil)
				resolver.SetModelOwnershipResolver(svc.resolveCompositeModelOwnership)
				decision, err := resolver.Resolve(context.Background(), 7, "gpt-6-astra", CompositeRouteEndpointResponses)
				require.NoError(t, err)
				require.False(t, decision.Matched, "ownership must depend on configured accounts, including cooldown accounts")
				wantReason := "model is exposed by multiple provider platforms"
				if typ != AccountTypeAPIKey {
					wantReason = "model is exposed by multiple provider platforms including OAuth accounts"
				}
				require.Equal(t, wantReason, decision.Reason)
				require.Equal(t, 1, repo.availabilityCalls)
			})
		}
	}
}
