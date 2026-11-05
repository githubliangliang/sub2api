package service

import "strings"

// Anthropic uses speed independently of its service_tier capacity classification.
// Only explicit speed values identify Fast; a beta header alone does not.
func normalizeAnthropicSpeed(raw string) string {
	switch value := strings.ToLower(strings.TrimSpace(raw)); value {
	case "fast", "standard":
		return value
	default:
		return ""
	}
}

func observeAnthropicUsageSpeed(usage *ClaudeUsage, raw string) {
	if speed := normalizeAnthropicSpeed(raw); speed != "" {
		usage.Speed = speed
	}
}
