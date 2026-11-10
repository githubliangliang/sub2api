//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestUltrafastCompatibilityPreservesExplicitTier(t *testing.T) {
	body, tier, err := normalizeResponsesBodyServiceTier([]byte(`{"model":"gpt-6-astra","service_tier":"ultrafast"}`))
	require.NoError(t, err)
	require.Equal(t, "ultrafast", tier)
	require.Equal(t, "ultrafast", gjson.GetBytes(body, "service_tier").String())
}

func TestFastGroupPolicyHTTPAndWS(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	for _, tier := range []string{"", "flex", "ultrafast"} {
		ctx := context.WithValue(context.Background(), ctxkey.Group, &Group{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ForceOpenAIFast: true})
		body := []byte(`{"type":"response.create","model":"gpt-6-astra","service_tier":"` + tier + `"}`)
		if tier == "" {
			body = []byte(`{"type":"response.create","model":"gpt-6-astra"}`)
		}
		svc := &OpenAIGatewayService{}
		got, err := svc.applyOpenAIFastPolicyToBody(ctx, account, "gpt-6-astra", body)
		require.NoError(t, err)
		require.Equal(t, "priority", gjson.GetBytes(got, "service_tier").String())
		got, blocked, err := svc.applyOpenAIFastPolicyToWSResponseCreate(ctx, account, "gpt-6-astra", body)
		require.NoError(t, err)
		require.Nil(t, blocked)
		require.Equal(t, "priority", gjson.GetBytes(got, "service_tier").String())
	}
}
func TestFastBillingAccountAuthority(t *testing.T) {
	priority, standard, ultra := "priority", "default", "ultrafast"
	require.Equal(t, "priority", *resolvedOpenAIServiceTier(&Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, &standard, &priority))
	require.Equal(t, "default", *resolvedOpenAIServiceTier(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, &standard, &priority))
	require.Equal(t, "priority", *resolvedOpenAIServiceTier(nil, &ultra, &priority))
	require.False(t, groupBillsOpenAIFastAtStandard(&APIKey{Group: &Group{Platform: PlatformOpenAI, FreeOpenAIFast: true}}, &Account{Platform: PlatformOpenAI}, ultra))
}

func TestFastTiersReachUpstreamAcrossHTTPProtocols(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, route := range []string{"responses", "passthrough", "chat", "messages", "responses-chat", "messages-chat"} {
		for _, stream := range []bool{false, true} {
			for _, tier := range []string{"fast", "ultrafast"} {
				t.Run(fmt.Sprintf("%s/%t/%s", route, stream, tier), func(t *testing.T) {
					model := "gpt-6-astra"
					request := map[string]any{"model": model, "stream": stream, "service_tier": tier, "input": "hi", "max_tokens": 64, "messages": []any{map[string]any{"role": "user", "content": "hi"}}}
					body, err := json.Marshal(request)
					require.NoError(t, err)
					canonical := tier
					if tier == "fast" {
						canonical = "priority"
					}
					rawChat := strings.HasSuffix(route, "-chat")
					response := `{"id":"resp_fast","model":"gpt-6-astra","status":"completed","service_tier":"` + canonical + `","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":5,"output_tokens":2}}`
					contentType := "application/json"
					// Converted Responses routes request SSE upstream even for buffered clients.
					if !rawChat && (stream || route == "chat" || route == "messages" || route == "passthrough") {
						response = "data: {\"type\":\"response.completed\",\"response\":" + response + "}\n\ndata: [DONE]\n\n"
						contentType = "text/event-stream"
					}
					if rawChat {
						response = `{"id":"chat_fast","object":"chat.completion","model":"gpt-6-astra","service_tier":"` + canonical + `","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`
						if stream {
							response = "data: " + strings.Replace(response, `"message":`, `"delta":`, 1) + "\n\ndata: [DONE]\n\n"
							contentType = "text/event-stream"
						}
					}
					upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(response))}}
					account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "fixture", "base_url": "https://api.example.test"}, Extra: map[string]any{}}
					if route == "passthrough" {
						account.Type = AccountTypeOAuth
						account.Credentials = map[string]any{"access_token": "fixture", "chatgpt_account_id": "fixture"}
						account.Extra["openai_passthrough"] = true
					}
					if rawChat {
						account = forceChatMessagesFallbackAccount()
						account.Credentials["base_url"] = "https://api.example.test"
					}
					svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(string(body)))
					var result *OpenAIForwardResult
					switch route {
					case "chat":
						result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
					case "messages", "messages-chat":
						result, err = svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
					default:
						result, err = svc.Forward(context.Background(), c, account, body)
					}
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, canonical, gjson.GetBytes(upstream.lastBody, "service_tier").String())
					require.Equal(t, canonical, optionalStringValue(result.ServiceTier))
				})
			}
		}
	}
}

func TestNativeClaudeFastTokenCost(t *testing.T) {
	svc := &GatewayService{billingService: newTestBillingService()}
	result := &ForwardResult{Usage: ClaudeUsage{InputTokens: 100, OutputTokens: 20}}
	normal := svc.calculateTokenCost(context.Background(), result, &APIKey{}, "claude-opus-4-6", 1, &recordUsageOpts{})
	fast := "fast"
	result.ServiceTier = &fast
	accelerated := svc.calculateTokenCost(context.Background(), result, &APIKey{}, "claude-opus-4-6", 1, &recordUsageOpts{})
	require.Positive(t, normal.TotalCost)
	require.InDelta(t, normal.TotalCost*2, accelerated.TotalCost, 1e-12)
}

func TestMessagesCompatibilityRejectsInvalidServiceTierBeforeForwarding(t *testing.T) {
	for _, tier := range []string{`"turbo"`, `""`, `42`} {
		for _, rawChat := range []bool{false, true} {
			account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			if rawChat {
				account = forceChatMessagesFallbackAccount()
			}
			upstream := &httpUpstreamRecorder{}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			body := []byte(`{"model":"gpt-6-astra","max_tokens":32,"messages":[{"role":"user","content":"hi"}],"service_tier":` + tier + `}`)
			_, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid service_tier")
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Nil(t, upstream.lastReq)
		}
	}
}

func TestFastMultiplierValidationReturnsClientError(t *testing.T) {
	for _, multiplier := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		err := checkPricesNotNegative(ChannelModelPricing{FastMultiplier: &multiplier})
		require.Equal(t, http.StatusBadRequest, infraerrors.Code(err))
	}
	multiplier := 2.5
	require.NoError(t, checkPricesNotNegative(ChannelModelPricing{FastMultiplier: &multiplier}))
	require.NoError(t, checkPricesNotNegative(ChannelModelPricing{}))
}
