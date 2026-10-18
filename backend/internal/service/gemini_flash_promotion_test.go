package service

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGeminiFlashPromotionExpiresWithoutRestart(t *testing.T) {
	svc := NewBillingService(nil, nil)
	for _, model := range []string{"gemini-3.6-flash", "gemini-3.6-flash-high", "models/gemini-3.6-flash-low", "gemini-3-6-flash", "google/gemini-3.6-flash-high"} {
		before, err := svc.getModelPricingAt(model, geminiFlashPromotionEnd.Add(-time.Nanosecond))
		require.NoError(t, err)
		require.InDelta(t, 0.75e-6, before.InputPricePerToken, 1e-15)
		after, err := svc.getModelPricingAt(model, geminiFlashPromotionEnd)
		require.NoError(t, err)
		require.InDelta(t, 1.5e-6, after.InputPricePerToken, 1e-15)
		require.InDelta(t, 7.5e-6, after.OutputPricePerToken, 1e-15)
		require.InDelta(t, 0.15e-6, after.CacheReadPricePerToken, 1e-15)
		require.InDelta(t, 0.75e-6, before.InputPricePerToken, 1e-15, "lookup must not mutate the shared snapshot")
	}
}

func TestGeminiFlashCatalogPromotionExpiresAndPreservesOverrides(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)
	for _, override := range []bool{false, true} {
		cfg := &config.Config{}
		if override {
			cfg.Pricing.OverrideFile = filepath.Join(t.TempDir(), "override.json")
			require.NoError(t, os.WriteFile(cfg.Pricing.OverrideFile, []byte(`{"gemini-3.6-flash":{"input_cost_per_token":0.00000075}}`), 0600))
		}
		catalog := &PricingService{cfg: cfg}
		catalog.pricingData, err = catalog.parsePricingData(data)
		require.NoError(t, err)
		svc := NewBillingService(cfg, catalog)
		before, err := svc.getModelPricingAt("gemini-3.6-flash", geminiFlashPromotionEnd.Add(-time.Second))
		require.NoError(t, err)
		require.InDelta(t, 0.75e-6, before.InputPricePerToken, 1e-15)
		after, err := svc.getModelPricingAt("gemini-3.6-flash", geminiFlashPromotionEnd)
		require.NoError(t, err)
		factor := 2.0
		if override {
			factor = 1
		}
		require.InDelta(t, 0.75e-6*factor, after.InputPricePerToken, 1e-15)
		require.InDelta(t, 1.35e-6*factor, after.InputPricePerTokenPriority, 1e-15)
		require.InDelta(t, 0.75e-6, catalog.pricingData["gemini-3.6-flash"].InputCostPerToken, 1e-15)
	}
}
