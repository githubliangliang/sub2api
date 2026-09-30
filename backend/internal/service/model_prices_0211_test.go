//go:build unit

package service

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSonnet55AndGPT61PricingSourcesAndOverrides(t *testing.T) {
	data, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	catalog := &PricingService{}
	catalog.pricingData, err = catalog.parsePricingData(data)
	require.NoError(t, err)
	for source, svc := range map[string]*BillingService{"fallback": newTestBillingService(), "catalog": NewBillingService(&config.Config{}, catalog)} {
		for _, model := range []string{"claude-sonnet-5-5", "anthropic/claude-sonnet-5.5", "us.anthropic.claude-sonnet-5-5"} {
			t.Run(source+"/"+model, func(t *testing.T) {
				cost, err := svc.CalculateCost(model, UsageTokens{InputTokens: 100000, OutputTokens: 500, CacheReadTokens: 1000, CacheCreationTokens: 1000, CacheCreation5mTokens: 400, CacheCreation1hTokens: 600}, 1)
				require.NoError(t, err)
				require.InDelta(t, 0.2, cost.InputCost, 1e-10)
				require.InDelta(t, 0.0034, cost.CacheCreationCost, 1e-10)
				require.InDelta(t, 0.0002, cost.CacheReadCost, 1e-10)
				require.InDelta(t, 0.005, cost.OutputCost, 1e-10)
			})
		}
		for _, model := range []string{"gpt-6.1-sol", "openai/gpt-6.1-sol-max"} {
			for _, tier := range []string{"", "priority"} {
				for _, gate := range []bool{false, true} {
					t.Run(source+"/"+model+"/"+tier+"/gate="+map[bool]string{false: "off", true: "on"}[gate], func(t *testing.T) {
						cost, err := svc.CalculateCostUnified(CostInput{Model: model, Tokens: UsageTokens{InputTokens: 300000, OutputTokens: 500, CacheReadTokens: 1000, CacheCreationTokens: 1000}, RateMultiplier: 1, ServiceTier: tier, LongContextBillingEnabled: &gate})
						require.NoError(t, err)
						factor, im, om := 1.0, 1.0, 1.0
						if tier == "priority" {
							factor = 2
						}
						if gate {
							im, om = 2, 1.5
						}
						require.InDelta(t, 0.6*factor*im, cost.InputCost, 1e-10)
						require.InDelta(t, 0.005*factor*om, cost.OutputCost, 1e-10)
						require.InDelta(t, 0.0001*factor*im, cost.CacheReadCost, 1e-10)
						require.InDelta(t, 0.0025*factor*im, cost.CacheCreationCost, 1e-10)
						require.Equal(t, gate, cost.LongContextBillingApplied)
					})
				}
			}
		}
		for _, model := range []string{"claude-sonnet-5-5", "gpt-6.1-sol"} {
			zero := 0.0
			p, err := svc.GetModelPricingWithChannel(model, &ChannelModelPricing{InputPrice: &zero, OutputPrice: &zero, CacheWritePrice: &zero, CacheReadPrice: &zero})
			require.NoError(t, err)
			cost := svc.computeTokenBreakdown(p, UsageTokens{InputTokens: 300000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000}, 1, "priority", true)
			require.Zero(t, cost.TotalCost, source+"/"+model)
		}
	}
}
