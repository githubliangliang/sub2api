//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestExistingSolUltrafastPricing(t *testing.T) {
	svc := newTestBillingService()
	for _, model := range []string{"gpt-6.1-sol", "openai/gpt-6.1-sol-high", "gpt-6.1-sol-max", "gpt-6-astra"} {
		for _, n := range []int{271999, 272000, 272001} {
			t.Run(model+"/"+strconv.Itoa(n), func(t *testing.T) {
				tokens := UsageTokens{InputTokens: n - 3000, CacheReadTokens: 2000, CacheCreationTokens: 1000, OutputTokens: 500}
				standard, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "")
				require.NoError(t, err)
				ultra, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "ultrafast")
				require.NoError(t, err)
				fast, err := svc.CalculateCostWithServiceTier(model, tokens, 1, "fast")
				require.NoError(t, err)
				require.InDelta(t, standard.InputCost*6, ultra.InputCost, 1e-10)
				require.InDelta(t, standard.OutputCost*6, ultra.OutputCost, 1e-10)
				require.InDelta(t, standard.CacheReadCost*6, ultra.CacheReadCost, 1e-10)
				require.InDelta(t, standard.CacheCreationCost*6, ultra.CacheCreationCost, 1e-10)
				require.InDelta(t, standard.TotalCost*2, fast.TotalCost, 1e-10)
				require.Equal(t, standard.LongContextBillingApplied, ultra.LongContextBillingApplied)
			})
		}
		for _, custom := range []float64{0, 1e-6} {
			fast := 0.0
			card, err := svc.GetModelPricingWithChannel(model, &ChannelModelPricing{
				InputPrice: &custom, OutputPrice: &custom, CacheReadPrice: &custom, CacheWritePrice: &custom, FastMultiplier: &fast,
			})
			require.NoError(t, err)
			require.Equal(t, 6.0, configuredServiceTierMultiplier("ultrafast", card))
			require.Zero(t, configuredServiceTierMultiplier("fast", card))
			cost := svc.computeTokenBreakdown(card, UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000}, 1, "ultrafast", false)
			require.InDelta(t, custom*24000, cost.TotalCost, 1e-10)
		}
	}
	for _, model := range []string{"gpt-6-sol", "gpt-6.1-terra", "gpt-5.6-sol"} {
		card, err := svc.GetModelPricing(model)
		require.NoError(t, err)
		require.Equal(t, 2.0, configuredServiceTierMultiplier("ultrafast", card))
	}
}

func TestUpdateSessionWindow_WeeklyRolloverBoundaries(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	future, expired := now.Add(24*time.Hour), now.Add(-time.Hour)
	for _, scenario := range []string{"future_seconds", "future_millis", "expired_seconds", "expired_millis", "missing_reset", "weekly_only"} {
		t.Run(scenario, func(t *testing.T) {
			oldEnd := now.Add(-time.Minute)
			extra := map[string]any{"session_window_utilization": 0.8, "passive_usage_7d_utilization": 0.9, "passive_usage_sampled_at": now.Add(-time.Hour).Format(time.RFC3339)}
			reset := future.Unix()
			switch scenario {
			case "future_millis":
				reset = future.UnixMilli()
			case "expired_seconds":
				reset = expired.Unix()
			case "expired_millis":
				reset = expired.UnixMilli()
			}
			if scenario != "missing_reset" {
				extra["passive_usage_7d_reset"] = strconv.FormatInt(reset, 10)
			}
			headers := http.Header{}
			if scenario == "weekly_only" {
				headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.95")
				headers.Set("anthropic-ratelimit-unified-7d-reset", strconv.FormatInt(future.UnixMilli(), 10))
			} else {
				headers.Set("anthropic-ratelimit-unified-5h-status", "allowed")
				headers.Set("anthropic-ratelimit-unified-5h-reset", strconv.FormatInt(now.Add(5*time.Hour).UnixMilli(), 10))
			}
			repo := &sessionWindowMockRepo{}
			account := &Account{ID: 7, Platform: PlatformAnthropic, Type: AccountTypeSetupToken, Extra: extra, SessionWindowEnd: &oldEnd}
			newRateLimitServiceForTest(repo).UpdateSessionWindow(context.Background(), account, headers)
			for _, call := range repo.updateExtraCalls {
				for key, value := range call.Updates {
					extra[key] = value
				}
			}
			kept := strings.HasPrefix(scenario, "future") || scenario == "weekly_only"
			if kept {
				require.NotNil(t, extra["passive_usage_7d_utilization"])
				require.NotNil(t, extra["passive_usage_sampled_at"])
			} else {
				require.Nil(t, extra["passive_usage_7d_utilization"])
				require.Nil(t, extra["passive_usage_7d_reset"])
				require.Nil(t, extra["passive_usage_sampled_at"])
			}
			if scenario == "weekly_only" {
				require.Empty(t, repo.sessionWindowCalls)
				require.Equal(t, 0.95, extra["passive_usage_7d_utilization"])
				require.Equal(t, future.Unix(), extra["passive_usage_7d_reset"])
			} else {
				require.Nil(t, extra["session_window_utilization"])
				require.Len(t, repo.sessionWindowCalls, 1)
			}
		})
	}
}

func TestUpdateSessionWindow_IndependentWeeklySampleSurvives(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	oldEnd := now.Add(-time.Minute)
	extra := map[string]any{
		"passive_usage_7d_utilization": 0.9, "passive_usage_7d_reset": now.Add(-time.Hour).Unix(),
		"passive_usage_7d_oi_utilization": 0.7, "passive_usage_7d_oi_reset": now.Add(24 * time.Hour).Unix(),
		"passive_usage_sampled_at": now.Add(-time.Hour).Format(time.RFC3339),
	}
	sampledAt := extra["passive_usage_sampled_at"]
	repo := &sessionWindowMockRepo{}
	account := &Account{ID: 7, Extra: extra, SessionWindowEnd: &oldEnd}
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-status", "allowed")
	headers.Set("anthropic-ratelimit-unified-5h-reset", strconv.FormatInt(now.Add(5*time.Hour).Unix(), 10))
	newRateLimitServiceForTest(repo).UpdateSessionWindow(context.Background(), account, headers)
	for _, call := range repo.updateExtraCalls {
		for key, value := range call.Updates {
			extra[key] = value
		}
	}
	require.Nil(t, extra["passive_usage_7d_utilization"])
	require.Nil(t, extra["passive_usage_7d_reset"])
	require.Equal(t, 0.7, extra["passive_usage_7d_oi_utilization"])
	require.Equal(t, sampledAt, extra["passive_usage_sampled_at"])
}

func TestClaudeAccountDefaultAndExplicitTestModel(t *testing.T) {
	for _, explicit := range []string{"", "claude-opus-4-6"} {
		t.Run("model="+explicit, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("data: {\"type\":\"message_stop\"}\n\n"))}}
			account := &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key"}}
			svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
			require.NoError(t, svc.testClaudeAccountConnection(c, account, explicit))
			want := explicit
			if want == "" {
				want = "claude-sonnet-4-6"
			}
			require.Equal(t, want, gjson.GetBytes(upstream.lastBody, "model").String())
			require.Contains(t, rec.Body.String(), `"model":"`+want+`"`)
		})
	}
}
