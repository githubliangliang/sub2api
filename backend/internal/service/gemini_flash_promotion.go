package service

import "time"

// The frozen v0.2.16 card is promotional through 2026-12-31 (UTC).
// Check at lookup, so long-running processes also stop using it at expiry.
var geminiFlashPromotionEnd = time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)

func expireGeminiFlashCatalogPromotion(pricing *LiteLLMModelPricing, at time.Time) *LiteLLMModelPricing {
	if pricing == nil || !pricing.geminiFlashPromotion || at.Before(geminiFlashPromotionEnd) {
		return pricing
	}
	next := *pricing
	next.InputCostPerToken *= 2
	next.OutputCostPerToken *= 2
	next.CacheReadInputTokenCost *= 2
	next.InputCostPerTokenPriority *= 2
	next.OutputCostPerTokenPriority *= 2
	next.CacheReadInputTokenCostPriority *= 2
	next.geminiFlashPromotion = false
	return &next
}
