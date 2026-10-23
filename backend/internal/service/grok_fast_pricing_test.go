package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGrok47FastPricingResourceAndFallback(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)
	dynamic := &PricingService{}
	dynamic.pricingData, err = dynamic.parsePricingData(data)
	require.NoError(t, err)

	for _, mode := range []struct {
		name    string
		pricing *PricingService
	}{{"offline", nil}, {"resource", dynamic}} {
		t.Run(mode.name, func(t *testing.T) {
			svc := NewBillingService(&config.Config{}, mode.pricing)
			for _, model := range []string{"grok-4.7-fast", "grok-4.7-build-fast", "GROK-4.7-FAST", "xai/grok-4.7-fast", "x-ai/grok-4.7-build-fast", "grok/grok-4.7-fast"} {
				t.Run(model, func(t *testing.T) {
					if mode.pricing != nil {
						catalog := mode.pricing.GetModelPricing(model)
						require.NotNil(t, catalog, "must resolve from the resource, not just the offline fallback")
						require.Equal(t, 200000, catalog.LongContextInputTokenThreshold)
						require.InDelta(t, 1.5, catalog.LongContextInputCostMultiplier, 1e-15)
						require.InDelta(t, 1.5, catalog.LongContextOutputCostMultiplier, 1e-15)
					}
					pricing, err := svc.GetModelPricing(model)
					require.NoError(t, err)
					require.InDelta(t, 4e-6, pricing.InputPricePerToken, 1e-15)
					require.InDelta(t, 1e-6, pricing.CacheReadPricePerToken, 1e-15)
					require.InDelta(t, 12e-6, pricing.OutputPricePerToken, 1e-15)
					require.True(t, svc.HasIdentifiedTokenPricing(model))
					for _, tc := range []struct {
						name                          string
						tokens                        UsageTokens
						long                          bool
						inPrice, cachePrice, outPrice float64
					}{
						{"199999", UsageTokens{InputTokens: 149999, CacheReadTokens: 50000, OutputTokens: 10000}, false, 4e-6, 1e-6, 12e-6},
						{"200000", UsageTokens{InputTokens: 150000, CacheReadTokens: 50000, OutputTokens: 10000}, true, 6e-6, 1.5e-6, 18e-6},
						{"cached only", UsageTokens{CacheReadTokens: 200000, OutputTokens: 10000}, true, 6e-6, 1.5e-6, 18e-6},
					} {
						t.Run(tc.name, func(t *testing.T) {
							cost, err := svc.CalculateCost(model, tc.tokens, 1)
							require.NoError(t, err)
							require.Equal(t, tc.long, cost.LongContextBillingApplied)
							require.InDelta(t, float64(tc.tokens.InputTokens)*tc.inPrice, cost.InputCost, 1e-12)
							require.InDelta(t, float64(tc.tokens.CacheReadTokens)*tc.cachePrice, cost.CacheReadCost, 1e-12)
							require.InDelta(t, float64(tc.tokens.OutputTokens)*tc.outPrice, cost.OutputCost, 1e-12)
						})
					}
				})
			}
			// Explicit channel prices still take precedence, including zero; the
			// shared default card must not be mutated by those overrides.
			inputPrice, cachePrice, outputPrice := 7e-6, 0.0, 21e-6
			override, err := svc.GetModelPricingWithChannel("grok-4.7-fast", &ChannelModelPricing{
				InputPrice: &inputPrice, CacheReadPrice: &cachePrice, OutputPrice: &outputPrice,
			})
			require.NoError(t, err)
			require.InDelta(t, 7e-6, override.InputPricePerToken, 1e-15)
			require.Zero(t, override.CacheReadPricePerToken)
			require.InDelta(t, 21e-6, override.OutputPricePerToken, 1e-15)
			fast, err := svc.GetModelPricing("grok-4.7-fast")
			require.NoError(t, err)
			require.InDelta(t, 4e-6, fast.InputPricePerToken, 1e-15)
			standard, err := svc.GetModelPricing("grok-4.7")
			require.NoError(t, err)
			require.InDelta(t, 2e-6, standard.InputPricePerToken, 1e-15)
			require.InDelta(t, 0.5e-6, standard.CacheReadPricePerToken, 1e-15)
			require.InDelta(t, 6e-6, standard.OutputPricePerToken, 1e-15)
		})
	}
}

func TestGrok47FastPricingSourceLongContextFields(t *testing.T) {
	// The upstream snapshot and downloaded catalogs express the long-context
	// rates with above_200k_tokens fields rather than local multiplier fields.
	dynamic := &PricingService{}
	var err error
	dynamic.pricingData, err = dynamic.parsePricingData([]byte(`{"grok-4.7-fast":{
		"input_cost_per_token":0.000004,"input_cost_per_token_above_200k_tokens":0.000006,
		"cache_read_input_token_cost":0.000001,"cache_read_input_token_cost_above_200k_tokens":0.0000015,
		"output_cost_per_token":0.000012,"output_cost_per_token_above_200k_tokens":0.000018,
		"litellm_provider":"xai"
	}}`))
	require.NoError(t, err)
	svc := NewBillingService(&config.Config{}, dynamic)
	cost, err := svc.CalculateCost("x-ai/grok-4.7-build-fast", UsageTokens{InputTokens: 150000, CacheReadTokens: 50000, OutputTokens: 10000}, 1)
	require.NoError(t, err)
	require.True(t, cost.LongContextBillingApplied)
	require.InDelta(t, 0.9, cost.InputCost, 1e-12)
	require.InDelta(t, 0.075, cost.CacheReadCost, 1e-12)
	require.InDelta(t, 0.18, cost.OutputCost, 1e-12)
}
