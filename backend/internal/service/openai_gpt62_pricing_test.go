//go:build unit

package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

func gpt62PricingTestServices(t *testing.T) map[string]*BillingService {
	t.Helper()
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	resource := &PricingService{}
	resource.pricingData, err = resource.parsePricingData(body)
	require.NoError(t, err)
	return map[string]*BillingService{
		"resource":         NewBillingService(nil, resource),
		"pricing fallback": NewBillingService(nil, &PricingService{pricingData: map[string]*LiteLLMModelPricing{}}),
		"billing fallback": NewBillingService(nil, nil),
	}
}

func TestGPT62UnifiedPricingUsesFrozenInstant(t *testing.T) {
	cutoff := time.Date(2027, time.March, 1, 8, 0, 0, 0, time.UTC)
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for source, svc := range gpt62PricingTestServices(t) {
			for _, useResolver := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/resolver=%t", source, model, useResolver), func(t *testing.T) {
					t.Parallel()
					resolver := NewModelPricingResolver(nil, svc)
					resolved := resolver.Resolve(context.Background(), PricingInput{Model: model})
					original := *resolved.BasePricing
					input := CostInput{Model: model, RateMultiplier: 1.3, ServiceTier: "ultrafast",
						Tokens: UsageTokens{InputTokens: 100_000, CacheCreationTokens: 100_000, CacheReadTokens: 72_001, OutputTokens: 10}}
					if useResolver {
						input.Resolver, input.Resolved = resolver, resolved
					}
					for _, test := range []struct {
						at    time.Time
						ratio float64
					}{
						{cutoff.Add(-time.Nanosecond), 3}, {cutoff, 6}, {cutoff.Add(-time.Nanosecond), 3},
					} {
						input.PricingAt = test.at
						cost, err := svc.CalculateCostUnified(input)
						require.NoError(t, err)
						require.True(t, cost.LongContextBillingApplied)
						require.InDelta(t, 100_000*original.InputPricePerToken*2*test.ratio, cost.InputCost, 1e-12)
						require.InDelta(t, 100_000*original.CacheCreationPricePerToken*2*test.ratio, cost.CacheCreationCost, 1e-12)
						require.InDelta(t, 72_001*original.CacheReadPricePerToken*2*test.ratio, cost.CacheReadCost, 1e-12)
						require.InDelta(t, 10*original.OutputPricePerToken*1.5*test.ratio, cost.OutputCost, 1e-12)
						require.InDelta(t, cost.TotalCost*1.3, cost.ActualCost, 1e-12)
					}
					require.Equal(t, original, *resolved.BasePricing)
				})
			}
		}
	}
}

func TestGPT62DelayedRecordUsageAndIndependentAccountCost(t *testing.T) {
	cutoff := time.Date(2027, time.March, 1, 8, 0, 0, 0, time.UTC)
	for _, model := range []struct {
		id       string
		standard float64
	}{{"gpt-6.2-sol", 0.00187}, {"gpt-6.2-astra", 0.00748}} {
		for _, useResolver := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/resolver=%t", model.id, useResolver), func(t *testing.T) {
				repo := &openAIRecordUsageLogRepoStub{inserted: true}
				svc := newOpenAIRecordUsageServiceForTest(repo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
				svc.channelService = newTestChannelServiceForStats(t, &Channel{ID: 1, Status: StatusActive}, 17, PlatformOpenAI)
				group := &Group{ID: 17, Platform: PlatformOpenAI, RateMultiplier: 1, LongContextPricingEnabled: true}
				userStandard := model.standard
				if useResolver {
					input, output, cacheWrite, cacheRead := 9e-6, 21e-6, 0.0, 0.3e-6
					group.ModelPricing = []ChannelModelPricing{{Models: []string{model.id}, InputPrice: &input, OutputPrice: &output, CacheWritePrice: &cacheWrite, CacheReadPrice: &cacheRead}}
					svc.resolver = NewModelPricingResolver(svc.channelService, svc.billingService)
					userStandard = 0.00477
				} else {
					svc.resolver = nil
				}
				for _, test := range []struct {
					at    time.Time
					ratio float64
				}{
					{cutoff.Add(-time.Nanosecond), 3}, {cutoff, 6}, {cutoff.Add(-time.Nanosecond), 3},
				} {
					tier := "ultrafast"
					err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
						Result: &OpenAIForwardResult{RequestID: fmt.Sprint(test.at.UnixNano()), Model: model.id, UpstreamModel: model.id, ServiceTier: &tier,
							Usage: OpenAIUsage{InputTokens: 1000, OutputTokens: 10, CacheCreationInputTokens: 300, CacheReadInputTokens: 200}},
						APIKey: &APIKey{ID: 1007, GroupID: i64p(17), Group: group}, User: &User{ID: 2007},
						Account: &Account{ID: 3007, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, PricingAt: test.at,
					})
					require.NoError(t, err)
					require.NotNil(t, repo.lastLog)
					require.InDelta(t, userStandard*test.ratio, repo.lastLog.TotalCost, 1e-12)
					require.NotNil(t, repo.lastLog.AccountStatsCost)
					require.InDelta(t, model.standard*test.ratio, *repo.lastLog.AccountStatsCost, 1e-12)
					require.WithinDuration(t, time.Now(), repo.lastLog.CreatedAt, time.Minute)
					if useResolver {
						require.Zero(t, repo.lastLog.CacheCreationCost)
					}
				}
			})
		}
	}
}

func TestGPT62PricingCardsAndLongContext(t *testing.T) {
	for _, model := range []struct {
		id                                   string
		input, output, cacheWrite, cacheRead float64
	}{
		{"gpt-6.2-sol", 2e-6, 10e-6, 2.5e-6, 0.1e-6},
		{"gpt-6.2-astra", 8e-6, 40e-6, 10e-6, 0.4e-6},
	} {
		for name, svc := range gpt62PricingTestServices(t) {
			t.Run(name+"/"+model.id, func(t *testing.T) {
				for _, alias := range []string{model.id, "openai/" + model.id + "-max", model.id + "-openai-compact"} {
					pricing, err := svc.GetModelPricing(alias)
					require.NoError(t, err)
					require.Equal(t, model.input, pricing.InputPricePerToken)
					require.Equal(t, model.output, pricing.OutputPricePerToken)
					require.Equal(t, model.cacheWrite, pricing.CacheCreationPricePerToken)
					require.Equal(t, model.cacheRead, pricing.CacheReadPricePerToken)
					require.Equal(t, 272_000, pricing.LongContextInputThreshold)
				}
				for _, count := range []int{271_999, 272_000, 272_001} {
					for _, tier := range []string{"", "fast", "priority", "flex"} {
						t.Run(fmt.Sprintf("%d/%s", count, tier), func(t *testing.T) {
							tokens := UsageTokens{InputTokens: 100_000, CacheCreationTokens: 100_000, CacheReadTokens: count - 200_000, OutputTokens: 10}
							cost, err := svc.CalculateCostWithServiceTier(model.id, tokens, 1, tier)
							require.NoError(t, err)
							inputFactor, outputFactor, tierFactor := 1.0, 1.0, 1.0
							if count > 272_000 {
								inputFactor, outputFactor = 2, 1.5
							}
							if tier == "fast" || tier == "priority" {
								tierFactor = 2
							} else if tier == "flex" {
								tierFactor = 0.5
							}
							require.Equal(t, count > 272_000, cost.LongContextBillingApplied)
							require.InDelta(t, float64(tokens.InputTokens)*model.input*inputFactor*tierFactor, cost.InputCost, 1e-12)
							require.InDelta(t, float64(tokens.CacheCreationTokens)*model.cacheWrite*inputFactor*tierFactor, cost.CacheCreationCost, 1e-12)
							require.InDelta(t, float64(tokens.CacheReadTokens)*model.cacheRead*inputFactor*tierFactor, cost.CacheReadCost, 1e-12)
							require.InDelta(t, float64(tokens.OutputTokens)*model.output*outputFactor*tierFactor, cost.OutputCost, 1e-12)
						})
					}
				}
			})
		}
	}
}

func TestGPT62IdentityAndProtocolCapabilities(t *testing.T) {
	for _, base := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		t.Run(base, func(t *testing.T) {
			require.Contains(t, openai.DefaultModelIDs(), base)
			for _, suffix := range []string{"", "-low", "-medium", "-high", "-xhigh", "-max", "-openai-compact"} {
				model := "openai/" + base + suffix
				require.Equal(t, base, normalizeKnownOpenAICodexModel(model))
				require.Equal(t, base, normalizeCodexModel(model))
				require.Equal(t, base, NormalizeOpenAICompatRequestedModel(model))
				require.True(t, shouldAutoInjectPromptCacheKeyForCompat(model))
				require.True(t, openai.RequiresResponsesForToolCalls(model))
				require.True(t, configuredCodexSupportsPriorityServiceTier(model))
			}
		})
	}
}

func TestGPT62UnknownNamesDoNotGuessPrices(t *testing.T) {
	for _, model := range []string{"gpt-6.2", "gpt-6.2-foo", "gpt-6.2-sol-preview", "gpt-6.2-astra-law"} {
		require.Empty(t, normalizeKnownOpenAICodexModel(model))
		for _, svc := range gpt62PricingTestServices(t) {
			_, err := svc.GetModelPricing(model)
			require.ErrorIs(t, err, ErrModelPricingUnavailable)
		}
	}
	pricing := newStubPricingServiceFromJSON(t, `{"gpt-6.2-sol-preview":{"input_cost_per_token":0.000003}}`)
	card, err := NewBillingService(nil, pricing).GetModelPricing("gpt-6.2-sol-preview")
	require.NoError(t, err)
	require.Equal(t, 3e-6, card.InputPricePerToken)
}

func TestGPT62UltrafastPromotionBoundary(t *testing.T) {
	pacific, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	cutoff := time.Date(2027, time.March, 1, 0, 0, 0, 0, pacific)
	require.Equal(t, time.Date(2027, time.March, 1, 8, 0, 0, 0, time.UTC), cutoff.UTC())
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for name, svc := range gpt62PricingTestServices(t) {
			t.Run(name+"/"+model, func(t *testing.T) {
				original := *svc.fallbackPrices[model]
				for _, test := range []struct {
					at    time.Time
					ratio float64
				}{
					{cutoff.Add(-time.Nanosecond), 3}, {cutoff, 6}, {cutoff.Add(time.Nanosecond), 6},
					{cutoff.Add(-time.Nanosecond).UTC(), 3},
					{cutoff.In(time.FixedZone("Shanghai", 8*60*60)), 6},
					{cutoff.Add(-time.Nanosecond), 3},
				} {
					pricing, err := svc.getModelPricingAt(model, test.at)
					require.NoError(t, err)
					standard := svc.computeTokenBreakdown(pricing, UsageTokens{InputTokens: 100, OutputTokens: 50, CacheCreationTokens: 40, CacheReadTokens: 20}, 1, "", true)
					ultra := svc.computeTokenBreakdown(pricing, UsageTokens{InputTokens: 100, OutputTokens: 50, CacheCreationTokens: 40, CacheReadTokens: 20}, 1.7, "ultrafast", true)
					require.InDelta(t, standard.InputCost*test.ratio, ultra.InputCost, 1e-12)
					require.InDelta(t, standard.OutputCost*test.ratio, ultra.OutputCost, 1e-12)
					require.InDelta(t, standard.CacheCreationCost*test.ratio, ultra.CacheCreationCost, 1e-12)
					require.InDelta(t, standard.CacheReadCost*test.ratio, ultra.CacheReadCost, 1e-12)
					require.InDelta(t, standard.TotalCost*test.ratio*1.7, ultra.ActualCost, 1e-12)
				}
				require.Equal(t, original, *svc.fallbackPrices[model])
			})
		}
	}
}
