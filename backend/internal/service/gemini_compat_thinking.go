package service

import (
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type claudeThinkingIntent struct {
	thinkingType string
	tier         string
	budgetTokens int
}

func parseClaudeThinkingIntent(body []byte) claudeThinkingIntent {
	intent := claudeThinkingIntent{thinkingType: strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "thinking.type").String()))}
	if intent.thinkingType == "disabled" {
		intent.tier = "minimal"
		return intent
	}
	switch strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "output_config.effort").String())) {
	case "none", "minimal":
		intent.tier = "minimal"
	case "low", "medium":
		intent.tier = strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "output_config.effort").String()))
	case "high", "xhigh", "max":
		intent.tier = "high"
	}
	if intent.thinkingType == "enabled" {
		intent.budgetTokens = int(gjson.GetBytes(body, "thinking.budget_tokens").Int())
	}
	return intent
}

// Adapt only the existing Gemini 2.5 and 3 families. Explicit tier mappings
// choose the variant themselves; omitted effort retains the provider default.
func applyGeminiCompatThinkingConfig(body, claudeBody []byte, model string, includeThoughts bool) []byte {
	name := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(model)), "models/")
	levelModel := strings.HasPrefix(name, "gemini-3-") || strings.HasPrefix(name, "gemini-3.")
	budgetModel := strings.HasPrefix(name, "gemini-2.5-")
	if !levelModel && !budgetModel {
		return body
	}
	for _, kind := range []string{"image", "tts", "embedding", "live", "audio"} {
		if strings.Contains(name, kind) {
			return body
		}
	}
	intent := parseClaudeThinkingIntent(claudeBody)
	set := func(field string, value any) {
		if next, err := sjson.SetBytes(body, "generationConfig.thinkingConfig."+field, value); err == nil {
			body = next
		}
	}
	thinkingOff := false
	if !hasGeminiThinkingVariantSuffix(name) {
		if levelModel {
			tier := intent.tier
			if tier == "" && intent.budgetTokens > 0 {
				switch {
				case intent.budgetTokens <= 1024:
					tier = "low"
				case intent.budgetTokens <= 8192:
					tier = "medium"
				default:
					tier = "high"
				}
			}
			flash := strings.Contains(name, "flash")
			if tier == "minimal" && !flash {
				tier = "low"
			}
			if tier == "medium" && !flash && !strings.HasPrefix(name, "gemini-3.1-pro") {
				tier = "high"
			}
			if tier != "" {
				set("thinkingLevel", tier)
			}
		} else {
			maximum, minimum := 24576, 1
			pro := strings.Contains(name, "pro")
			if pro {
				maximum, minimum = 32768, 128
			} else if strings.Contains(name, "flash-lite") {
				minimum = 512
			}
			budget, explicit := intent.budgetTokens, intent.budgetTokens > 0
			switch intent.tier {
			case "minimal":
				budget, explicit = 0, true
				if pro {
					budget = 128
				}
			case "low":
				budget, explicit = 1024, true
			case "medium":
				budget, explicit = 8192, true
			case "high":
				budget, explicit = maximum, true
			}
			if explicit {
				if budget > 0 {
					budget = min(max(budget, minimum), maximum)
				}
				set("thinkingBudget", budget)
				thinkingOff = budget == 0
			}
		}
	}
	if includeThoughts && !thinkingOff {
		set("includeThoughts", true)
	}
	return body
}
