package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTier2AccountStatsLongContextGate(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
		svc := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
		svc.channelService = newTestChannelServiceForStats(t, &Channel{ID: 1, Status: StatusActive}, 1, PlatformOpenAI)
		svc.resolver = NewModelPricingResolver(svc.channelService, svc.billingService)
		apiKey := &APIKey{ID: 100, GroupID: i64p(1), Group: &Group{ID: 1, Platform: PlatformOpenAI, LongContextPricingEnabled: true}}
		err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
			Result: &OpenAIForwardResult{RequestID: "tier2-cost", Usage: OpenAIUsage{InputTokens: 300000, OutputTokens: 2000}, Model: "gpt-5.4", Duration: time.Second},
			APIKey: apiKey, User: &User{ID: 200}, Account: &Account{ID: 300, Platform: PlatformOpenAI, Extra: map[string]any{openAILongContextBillingEnabledKey: enabled}},
		})
		require.NoError(t, err)
		require.NotNil(t, usageRepo.lastLog)
		require.NotNil(t, usageRepo.lastLog.AccountStatsCost)
		want := 0.78
		if enabled {
			want = 1.545
		}
		require.InDelta(t, want, *usageRepo.lastLog.AccountStatsCost, 1e-10)
		require.InDelta(t, 1.545, usageRepo.lastLog.TotalCost, 1e-10, "group sale gate remains enabled")
	}
}

func TestTier2ChannelImagePriceInheritsCatalogUnlessExplicit(t *testing.T) {
	bs := newTestBillingServiceWithPrices(map[string]*ModelPricing{"gpt-5.4": {InputPricePerToken: 1e-6, OutputPricePerToken: 2e-6, ImageInputPricePerToken: 3e-6, ImageOutputPricePerToken: 4e-6}})
	zero := 0.0
	for _, price := range []*float64{nil, &zero} {
		channel := &ChannelModelPricing{ImageOutputPrice: price}
		pricing, err := bs.GetModelPricingWithChannel("gpt-5.4", channel)
		require.NoError(t, err)
		want := 4e-6
		if price != nil {
			want = 0
		}
		require.Equal(t, want, pricing.ImageOutputPricePerToken)
		require.Equal(t, price != nil, pricing.ImageOutputPriceExplicit)
		require.Equal(t, 3e-6, pricing.ImageInputPricePerToken)
		resolved := &ResolvedPricing{BasePricing: bs.fallbackPrices["gpt-5.4"]}
		(&ModelPricingResolver{}).applyTokenOverrides(channel, resolved)
		require.Equal(t, want, resolved.BasePricing.ImageOutputPricePerToken)
		require.Equal(t, price != nil, resolved.BasePricing.ImageOutputPriceExplicit)
	}
}
