//go:build unit

package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func terraPricingTestServices(t *testing.T) map[string]*BillingService {
	t.Helper()
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	resource := &PricingService{}
	resource.pricingData, err = resource.parsePricingData(body)
	require.NoError(t, err)
	return map[string]*BillingService{
		"resource":         NewBillingService(&config.Config{}, resource),
		"pricing fallback": NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{}}),
		"billing fallback": NewBillingService(&config.Config{}, nil),
	}
}

func TestGPT61TerraFrozenPricingAndFallbacks(t *testing.T) {
	for name, svc := range terraPricingTestServices(t) {
		for _, model := range []string{"gpt-6.1-terra", "openai/GPT_6.1_TERRA_MAX", "gpt-6.1-terra-openai-compact"} {
			t.Run(name+"/"+model, func(t *testing.T) {
				pricing, err := svc.GetModelPricing(model)
				require.NoError(t, err)
				require.InDelta(t, 1e-6, pricing.InputPricePerToken, 1e-12)
				require.InDelta(t, 2e-6, pricing.InputPricePerTokenPriority, 1e-12)
				require.InDelta(t, 5e-6, pricing.OutputPricePerToken, 1e-12)
				require.InDelta(t, 10e-6, pricing.OutputPricePerTokenPriority, 1e-12)
				require.InDelta(t, 1.25e-6, pricing.CacheCreationPricePerToken, 1e-12)
				require.InDelta(t, 2.5e-6, pricing.CacheCreationPricePerTokenPriority, 1e-12)
				require.InDelta(t, 0.05e-6, pricing.CacheReadPricePerToken, 1e-12)
				require.InDelta(t, 0.1e-6, pricing.CacheReadPricePerTokenPriority, 1e-12)
				require.Equal(t, 272_000, pricing.LongContextInputThreshold)
				require.InDelta(t, 2, pricing.LongContextInputMultiplier, 1e-12)
				require.InDelta(t, 1.5, pricing.LongContextOutputMultiplier, 1e-12)
			})
		}
	}
}

func TestGPT61TerraLongContextBoundaryAndTokenBuckets(t *testing.T) {
	for name, svc := range terraPricingTestServices(t) {
		for _, count := range []int{271_999, 272_000, 272_001} {
			for _, tier := range []string{"", "fast", "priority", "flex"} {
				t.Run(fmt.Sprintf("%s/%d/%s", name, count, tier), func(t *testing.T) {
					tokens := UsageTokens{InputTokens: 100_000, CacheCreationTokens: 100_000, CacheReadTokens: count - 200_000, OutputTokens: 10, CacheCreation5mTokens: 40_000, CacheCreation1hTokens: 60_000}
					cost, err := svc.CalculateCostWithServiceTier("gpt-6.1-terra", tokens, 1, tier)
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
					require.InDelta(t, float64(tokens.InputTokens)*1e-6*inputFactor*tierFactor, cost.InputCost, 1e-12)
					require.InDelta(t, 100_000*1.25e-6*inputFactor*tierFactor, cost.CacheCreationCost, 1e-12)
					require.InDelta(t, float64(tokens.CacheReadTokens)*0.05e-6*inputFactor*tierFactor, cost.CacheReadCost, 1e-12)
					require.InDelta(t, 10*5e-6*outputFactor*tierFactor, cost.OutputCost, 1e-12)
				})
			}
		}
	}
}

func TestGPT61TerraExplicitZeroAndPricingPrecedence(t *testing.T) {
	pricing := &PricingService{}
	var err error
	pricing.pricingData, err = pricing.parsePricingData([]byte(`{"gpt-6.1-terra":{"litellm_provider":"openai","input_cost_per_token":0.000003,"output_cost_per_token":0.000007,"cache_read_input_token_cost":0.0000002,"cache_creation_input_token_cost":0,"cache_creation_input_token_cost_priority":0.000009}}`))
	require.NoError(t, err)
	original := *pricing.pricingData["gpt-6.1-terra"]
	svc := NewBillingService(&config.Config{}, pricing)
	for _, tier := range []string{"", "fast", "priority", "flex"} {
		t.Run("zero/"+tier, func(t *testing.T) {
			cost, err := svc.CalculateCostWithServiceTier("openai/gpt-6.1-terra-max", UsageTokens{CacheCreationTokens: 300_000}, 1, tier)
			require.NoError(t, err)
			require.Zero(t, cost.CacheCreationCost)
		})
	}
	base, err := svc.GetModelPricing("gpt-6.1-terra")
	require.NoError(t, err)
	require.InDelta(t, 3e-6, base.InputPricePerToken, 1e-12)
	require.InDelta(t, 7e-6, base.OutputPricePerToken, 1e-12)
	require.True(t, base.CacheCreationPriceExplicit)
	zero, input, output, cacheRead := 0.0, 9e-6, 21e-6, 0.3e-6
	channel, err := svc.GetModelPricingWithChannel("gpt-6.1-terra", &ChannelModelPricing{InputPrice: &input, OutputPrice: &output, CacheWritePrice: &zero, CacheReadPrice: &cacheRead})
	require.NoError(t, err)
	require.Equal(t, input, channel.InputPricePerToken)
	require.Equal(t, input, channel.InputPricePerTokenPriority)
	require.Equal(t, output, channel.OutputPricePerToken)
	require.Equal(t, cacheRead, channel.CacheReadPricePerTokenPriority)
	require.Zero(t, channel.CacheCreationPricePerTokenPriority)
	require.Equal(t, original, *pricing.pricingData["gpt-6.1-terra"])
	for _, fallback := range terraPricingTestServices(t) {
		card, err := fallback.GetModelPricing("gpt-6.1-terra")
		require.NoError(t, err)
		before := *card
		_, err = fallback.GetModelPricingWithChannel("gpt-6.1-terra", &ChannelModelPricing{InputPrice: &input, CacheWritePrice: &zero})
		require.NoError(t, err)
		after, err := fallback.GetModelPricing("gpt-6.1-terra")
		require.NoError(t, err)
		require.Equal(t, before, *after)
	}
}

func TestGPT61TerraPricingOverridePreservesFrozenDefaults(t *testing.T) {
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	pricing := newPricingServiceWithOverride(t, `{"gpt-6.1-terra":{"input_cost_per_token":0.000004,"cache_creation_input_token_cost":0}}`)
	pricing.pricingData, err = pricing.parsePricingData(body)
	require.NoError(t, err)
	card := pricing.GetModelPricing("gpt-6.1-terra")
	require.NotNil(t, card)
	require.Equal(t, 4e-6, card.InputCostPerToken)
	require.Equal(t, 5e-6, card.OutputCostPerToken)
	require.Equal(t, 0.05e-6, card.CacheReadInputTokenCost)
	require.Zero(t, card.CacheCreationInputTokenCost)
	require.True(t, card.CacheCreationInputTokenCostExplicit)
	cost, err := NewBillingService(&config.Config{}, pricing).CalculateCostWithServiceTier("gpt-6.1-terra", UsageTokens{CacheCreationTokens: 300_000}, 1, "priority")
	require.NoError(t, err)
	require.Zero(t, cost.CacheCreationCost)
}
