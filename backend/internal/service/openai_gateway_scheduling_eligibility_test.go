//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAICompatibleAccountEligibilityFailureReason_NamesFirstVeto(t *testing.T) {
	ctx := withOpenAIQuotaAutoPauseSettings(context.Background(), OpsOpenAIAccountQuotaAutoPauseSettings{DefaultThreshold7d: 0.9})
	healthy := &Account{
		ID:          1,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
	}
	require.Equal(t, "", openAICompatibleAccountEligibilityFailureReason(ctx, healthy, PlatformOpenAI, "gpt-5.4", false, ""))
	require.True(t, isOpenAICompatibleAccountEligibleForRequest(ctx, healthy, PlatformOpenAI, "gpt-5.4", false, ""))

	require.Equal(t, "account_nil", openAICompatibleAccountEligibilityFailureReason(ctx, nil, PlatformOpenAI, "gpt-5.4", false, ""))

	grok := *healthy
	grok.Platform = PlatformGrok
	require.Equal(t, "platform_mismatch", openAICompatibleAccountEligibilityFailureReason(ctx, &grok, PlatformOpenAI, "gpt-5.4", false, ""))

	paused := &Account{
		ID:          2,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"codex_7d_used_percent":  95.0,
			"codex_7d_reset_at":      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"codex_usage_updated_at": time.Now().Add(-time.Minute).Format(time.RFC3339),
		},
	}
	require.Equal(t, "quota_auto_pause_7d", openAICompatibleAccountEligibilityFailureReason(ctx, paused, PlatformOpenAI, "gpt-5.4", false, ""))
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(ctx, paused, PlatformOpenAI, "gpt-5.4", false, ""))

	unschedulable := *healthy
	unschedulable.Schedulable = false
	require.Equal(t, "not_schedulable", openAICompatibleAccountEligibilityFailureReason(ctx, &unschedulable, PlatformOpenAI, "gpt-5.4", false, ""))

	mapped := &Account{
		ID:          3,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-4o": "gpt-4o"}},
	}
	require.Equal(t, "model_not_supported", openAICompatibleAccountEligibilityFailureReason(ctx, mapped, PlatformOpenAI, "gpt-5.4-mini", false, ""))
}

func TestOpenAICompatibleAccountEligibilityFailureReason_ModelRateLimitedWhileSchedulable(t *testing.T) {
	ctx := context.Background()
	account := &Account{
		ID:          11,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			modelRateLimitsKey: map[string]any{
				"gpt-5.4": map[string]any{
					"rate_limit_reset_at": time.Now().Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
	}
	require.True(t, account.IsSchedulable())
	require.False(t, account.IsSchedulableForModelWithContext(ctx, "gpt-5.4"))
	require.Equal(t, "model_rate_limited", openAICompatibleAccountEligibilityFailureReason(ctx, account, PlatformOpenAI, "gpt-5.4", false, ""))
	require.False(t, isOpenAICompatibleAccountEligibleForRequest(ctx, account, PlatformOpenAI, "gpt-5.4", false, ""))
}

func TestClearActualOpenAIUpstreamEndpoint_BetweenFailoverAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	SetActualOpenAIUpstreamEndpoint(c, "/v1/responses")
	require.Equal(t, "/v1/responses", GetActualOpenAIUpstreamEndpoint(c))
	ClearActualOpenAIUpstreamEndpoint(c)
	require.Empty(t, GetActualOpenAIUpstreamEndpoint(c))
	SetActualOpenAIUpstreamEndpoint(c, "/v1/chat/completions")
	require.Equal(t, "/v1/chat/completions", GetActualOpenAIUpstreamEndpoint(c))
}
