package service

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

func anthropicUsageServiceTier(usage ClaudeUsage) *string {
	speed := normalizeAnthropicSpeed(usage.Speed)
	if speed == "" {
		return nil
	}
	return &speed
}

func (usage OpenAIUsage) billingUsageTokens() UsageTokens {
	creation := usage.CacheCreationInputTokens
	if creation <= 0 {
		creation = max(usage.CacheCreation5mTokens, 0) + max(usage.CacheCreation1hTokens, 0)
	}
	return UsageTokens{
		InputTokens:      max(usage.InputTokens-usage.CacheReadInputTokens-creation, 0),
		ImageInputTokens: usage.ImageInputTokens, OutputTokens: usage.OutputTokens,
		CacheCreationTokens: creation, CacheReadTokens: usage.CacheReadInputTokens,
		CacheCreation5mTokens: usage.CacheCreation5mTokens, CacheCreation1hTokens: usage.CacheCreation1hTokens,
		ImageOutputTokens: usage.ImageOutputTokens,
	}
}

// Preserve independently updated TTL snapshots without interpreting missing or
// null fields as zero. A positive aggregate remains authoritative over details.
func mergeOpenAICacheCreationUsage(node gjson.Result, dst *OpenAIUsage) {
	cache := ClaudeUsage{
		CacheCreationInputTokens: dst.CacheCreationInputTokens,
		CacheCreation5mTokens:    dst.CacheCreation5mTokens, CacheCreation1hTokens: dst.CacheCreation1hTokens,
		cacheCreationInputTokensPresent:     dst.cacheCreationInputTokensPresent,
		cacheCreationInputTokensFromDetails: dst.cacheCreationInputTokensFromDetails,
		cacheCreation5mTokensPresent:        dst.cacheCreation5mTokensPresent,
		cacheCreation1hTokensPresent:        dst.cacheCreation1hTokensPresent,
	}
	source := apicompat.AnthropicUsage{
		CacheCreationInputTokens:        dst.CacheCreationInputTokens,
		CacheCreationInputTokensPresent: dst.cacheCreationInputTokensPresent,
		CacheCreation:                   &apicompat.AnthropicCacheCreation{},
	}
	for _, path := range []string{"cache_creation_5m_tokens", "cache_creation.ephemeral_5m_input_tokens", "input_tokens_details.cache_creation.ephemeral_5m_input_tokens", "prompt_tokens_details.cache_creation.ephemeral_5m_input_tokens"} {
		if value := node.Get(path); value.Exists() && value.Type != gjson.Null {
			tokens := int(value.Int())
			source.CacheCreation.Ephemeral5mInputTokens = &tokens
			break
		}
	}
	for _, path := range []string{"cache_creation_1h_tokens", "cache_creation.ephemeral_1h_input_tokens", "input_tokens_details.cache_creation.ephemeral_1h_input_tokens", "prompt_tokens_details.cache_creation.ephemeral_1h_input_tokens"} {
		if value := node.Get(path); value.Exists() && value.Type != gjson.Null {
			tokens := int(value.Int())
			source.CacheCreation.Ephemeral1hInputTokens = &tokens
			break
		}
	}
	mergeAnthropicCacheCreationUsage(&cache, source)
	dst.CacheCreationInputTokens = cache.CacheCreationInputTokens
	dst.CacheCreation5mTokens, dst.CacheCreation1hTokens = cache.CacheCreation5mTokens, cache.CacheCreation1hTokens
	dst.cacheCreationInputTokensPresent = cache.cacheCreationInputTokensPresent
	dst.cacheCreationInputTokensFromDetails = cache.cacheCreationInputTokensFromDetails
	dst.cacheCreation5mTokensPresent, dst.cacheCreation1hTokensPresent = cache.cacheCreation5mTokensPresent, cache.cacheCreation1hTokensPresent
}

func mergeOpenAIUsageSnapshot(dst *OpenAIUsage, src OpenAIUsage) {
	if src.InputTokens > 0 {
		dst.InputTokens = src.InputTokens
	}
	if src.OutputTokens > 0 {
		dst.OutputTokens = src.OutputTokens
	}
	if src.ImageInputTokens > 0 {
		dst.ImageInputTokens = src.ImageInputTokens
	}
	if src.ImageOutputTokens > 0 {
		dst.ImageOutputTokens = src.ImageOutputTokens
	}
	if src.CacheReadInputTokens > 0 {
		dst.CacheReadInputTokens = src.CacheReadInputTokens
	}
	if src.cacheCreationInputTokensPresent && !src.cacheCreationInputTokensFromDetails && (src.CacheCreationInputTokens > 0 || dst.CacheCreationInputTokens <= 0) {
		dst.CacheCreationInputTokens = src.CacheCreationInputTokens
		dst.cacheCreationInputTokensPresent = true
		dst.cacheCreationInputTokensFromDetails = src.cacheCreationInputTokensFromDetails
	}
	if src.cacheCreation5mTokensPresent {
		dst.CacheCreation5mTokens = src.CacheCreation5mTokens
		dst.cacheCreation5mTokensPresent = true
	}
	if src.cacheCreation1hTokensPresent {
		dst.CacheCreation1hTokens = src.CacheCreation1hTokens
		dst.cacheCreation1hTokensPresent = true
	}
	if (!dst.cacheCreationInputTokensPresent || dst.cacheCreationInputTokensFromDetails || dst.CacheCreationInputTokens <= 0) &&
		(src.cacheCreation5mTokensPresent || src.cacheCreation1hTokensPresent) {
		dst.CacheCreationInputTokens = dst.CacheCreation5mTokens + dst.CacheCreation1hTokens
		dst.cacheCreationInputTokensFromDetails = true
	}
	if src.ProviderCostUSD != nil {
		dst.ProviderCostUSD = src.ProviderCostUSD
	}
}
