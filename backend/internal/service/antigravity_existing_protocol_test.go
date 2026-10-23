//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAntigravityExistingClaudeTierPrices(t *testing.T) {
	for _, dynamic := range []bool{false, true} {
		svc := newTestBillingService()
		if dynamic {
			svc.pricingService = &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				"claude-opus-5-5":   {InputCostPerToken: 4e-6, OutputCostPerToken: 20e-6, CacheReadInputTokenCost: 0.2e-6},
				"claude-sonnet-5-5": {InputCostPerToken: 2e-6, OutputCostPerToken: 10e-6, CacheReadInputTokenCost: 0.1e-6},
			}}
		}
		for _, base := range []string{"claude-opus-5-5", "claude-sonnet-5-5"} {
			want, err := svc.GetModelPricing(base)
			require.NoError(t, err)
			for _, tier := range []string{"high", "medium", "low"} {
				for _, prefix := range []string{"", "anthropic/"} {
					model := prefix + base + "-" + tier
					got, err := svc.GetModelPricing(model)
					require.NoError(t, err)
					require.Equal(t, want.InputPricePerToken, got.InputPricePerToken, "dynamic=%t model=%s", dynamic, model)
					require.Equal(t, want.OutputPricePerToken, got.OutputPricePerToken)
					require.Equal(t, want.CacheReadPricePerToken, got.CacheReadPricePerToken)
				}
			}
		}
	}
}

func TestAntigravityBufferedGeminiPreservesFunctionCallAndSignatures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stream := "data: {\"response\":{\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"functionCall\":{\"name\":\"weather\",\"args\":{\"city\":\"Tokyo\"}},\"thoughtSignature\":\"call-signature\"}]}}]}}\n\n" +
		"data: {\"response\":{\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":12,\"candidatesTokenCount\":3}}}\n\n"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.6-flash:generateContent", nil)
	svc := &AntigravityGatewayService{settingService: &SettingService{cfg: &config.Config{}}}
	result, err := svc.handleGeminiStreamToNonStreaming(c, &http.Response{Header: http.Header{}, Body: io.NopCloser(strings.NewReader(stream))}, time.Now())
	require.NoError(t, err)
	require.NotNil(t, result)
	parts := gjson.Get(rec.Body.String(), "candidates.0.content.parts").Array()
	require.Len(t, parts, 1)
	require.Equal(t, "weather", parts[0].Get("functionCall.name").String())
	require.Equal(t, "call-signature", parts[0].Get("thoughtSignature").String())
	require.Equal(t, "STOP", gjson.Get(rec.Body.String(), "candidates.0.finishReason").String())
}

func TestAntigravitySelfMappingSelectsExistingExactVariant(t *testing.T) {
	account := &Account{Platform: PlatformAntigravity, Credentials: map[string]any{"model_mapping": map[string]any{
		"gemini-3.6-flash": "gemini-3.6-flash", "gemini-3.6-flash-high": "gemini-3.6-flash-high",
	}}}
	model, matched := resolveGeminiThinkingVariantForLevel(account, "gemini-3.6-flash", "high")
	require.True(t, matched)
	require.Equal(t, "gemini-3.6-flash-high", model)
}

func TestAntigravityExistingSubscriptionTierSubIDs(t *testing.T) {
	for _, tc := range []struct{ id, plan string }{
		{"g1-ultra-lite-tier", "Ultra"}, {"g1-pro-plus-tier", "Pro"}, {"unknown-tier", "unknown-tier"},
	} {
		require.Equal(t, tc.plan, antigravity.TierIDToPlanType(tc.id))
		if tc.plan == "Ultra" || tc.plan == "Pro" {
			require.True(t, accountHasAntigravityPaidTier(&Account{Credentials: map[string]any{"plan_type": tc.id}}))
		}
	}
}

func TestAntigravityExplicitTierPriceKeepsPriority(t *testing.T) {
	svc := newTestBillingService()
	svc.pricingService = &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"claude-opus-5-5":      {InputCostPerToken: 4e-6},
		"claude-opus-5-5-high": {InputCostPerToken: 7e-6},
	}}
	for _, model := range []string{"claude-opus-5-5-high", "anthropic/claude-opus-5-5-high"} {
		price, err := svc.GetModelPricing(model)
		require.NoError(t, err)
		require.Equal(t, 7e-6, price.InputPricePerToken)
	}
}

func TestAntigravitySelfMappingDoesNotInventWildcardVariant(t *testing.T) {
	account := &Account{Platform: PlatformAntigravity, Credentials: map[string]any{"model_mapping": map[string]any{
		"gemini-2.5-flash": "gemini-2.5-flash", "gemini-2.5*": "gemini-2.5-flash",
	}}}
	_, matched := resolveGeminiThinkingVariantForLevel(account, "gemini-2.5-flash", "high")
	require.False(t, matched)
}

func TestAntigravityConnectionTestUsesMappedProtocol(t *testing.T) {
	for _, tc := range []struct {
		requested, target string
		wantTools         bool
	}{
		{"public-alias", "gemini-3.6-flash", false},
		{"gemini-3.6-flash-high", "claude-sonnet-5-5", true},
	} {
		t.Run(tc.requested, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{},
				Body: io.NopCloser(strings.NewReader("data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"OK\"}]}}]}}\n\n")),
			}}
			svc := &AntigravityGatewayService{
				settingService: NewSettingService(&antigravitySettingRepoStub{}, &config.Config{}),
				tokenProvider:  &AntigravityTokenProvider{}, httpUpstream: upstream,
			}
			account := &Account{ID: 109, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Credentials: map[string]any{
				"access_token": "test-token", "project_id": "test-project", "model_mapping": map[string]any{tc.requested: tc.target},
			}}
			result, err := svc.TestConnection(context.Background(), account, tc.requested)
			require.NoError(t, err)
			require.Equal(t, tc.target, result.MappedModel)
			require.Equal(t, "OK", result.Text)
			require.Equal(t, tc.wantTools, gjson.GetBytes(upstream.lastBody, "request.toolConfig").Exists(), "Claude and Gemini test builders must be selected by the mapped target")
		})
	}
}

func TestAntigravityForwardGeminiReplaysFunctionCallSignatures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{},
		Body: io.NopCloser(strings.NewReader("data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"OK\"}]},\"finishReason\":\"STOP\"}]}}\n\n")),
	}}
	svc := &AntigravityGatewayService{
		settingService: NewSettingService(&antigravitySettingRepoStub{}, &config.Config{}),
		tokenProvider:  &AntigravityTokenProvider{}, httpUpstream: upstream,
	}
	account := &Account{ID: 110, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Credentials: map[string]any{
		"access_token": "test-token", "project_id": "test-project", "model_mapping": map[string]any{"gemini-3.6-flash": "gemini-3.6-flash"},
	}}
	body := []byte(`{"contents":[{"role":"model","parts":[{"functionCall":{"name":"weather","args":{}},"thoughtSignature":"retained-signature"},{"functionCall":{"name":"time","args":{}}}]}]}`)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.6-flash:generateContent", strings.NewReader(string(body)))
	_, err := svc.ForwardGemini(context.Background(), c, account, "gemini-3.6-flash", "generateContent", false, body, false)
	require.NoError(t, err)
	parts := gjson.GetBytes(upstream.lastBody, "request.contents.0.parts").Array()
	require.Equal(t, "retained-signature", parts[0].Get("thoughtSignature").String())
	require.NotEmpty(t, parts[1].Get("thoughtSignature").String(), "missing signatures use the existing replay normalizer")
}
