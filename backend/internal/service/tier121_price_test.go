package service

import (
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestTier121SonnetPrice(t *testing.T) {
	data, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	catalog := &PricingService{}
	catalog.pricingData, err = catalog.parsePricingData(data)
	require.NoError(t, err)
	fallback := &BillingService{fallbackPrices: make(map[string]*ModelPricing)}
	fallback.initFallbackPricing()
	for source, s := range map[string]*BillingService{"billing": fallback, "catalog": NewBillingService(&config.Config{}, catalog), "remote-missing": NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{}})} {
		for _, model := range []string{"claude-sonnet-5-5", "anthropic/claude-sonnet-5.5", "us.anthropic.claude-sonnet-5-5"} {
			t.Run(source+"/"+model, func(t *testing.T) {
				tokens := UsageTokens{InputTokens: 1000, OutputTokens: 100, CacheReadTokens: 1000000}
				ordinary, err := s.calculateCostInternal(model, tokens, 1, "", nil)
				require.NoError(t, err)
				fast, err := s.calculateCostInternal(model, tokens, 1, "priority", nil)
				require.NoError(t, err)
				require.InDelta(t, 0.1, ordinary.CacheReadCost, 1e-12)
				require.InDelta(t, 0.2, fast.CacheReadCost, 1e-12)
				require.InDelta(t, 2*ordinary.InputCost, fast.InputCost, 1e-12)
				require.InDelta(t, 2*ordinary.OutputCost, fast.OutputCost, 1e-12)
				for _, override := range []float64{0, 0.7e-6} {
					cost, err := s.calculateCostInternal(model, tokens, 1, "priority", &ChannelModelPricing{CacheReadPrice: &override})
					require.NoError(t, err)
					require.InDelta(t, 1e6*override, cost.CacheReadCost, 1e-12)
					multiplier := 3.0
					cost, err = s.calculateCostInternal(model, tokens, 1, "priority", &ChannelModelPricing{CacheReadPrice: &override, FastMultiplier: &multiplier})
					require.NoError(t, err)
					require.InDelta(t, 3e6*override, cost.CacheReadCost, 1e-12)
				}
			})
		}
	}
}
