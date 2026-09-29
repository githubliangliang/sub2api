//go:build unit

package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/stretchr/testify/require"
)

func TestTier2ModelBillingAliases(t *testing.T) {
	billing := newTestBillingService()
	for _, tc := range []struct {
		model                        string
		input, output, cached, write float64
	}{
		{"gpt-6-sol", 2e-6, 10e-6, 0.2e-6, 2.5e-6},
		{"openai/gpt-6-sol-max", 2e-6, 10e-6, 0.2e-6, 2.5e-6},
		{"gpt-6-luna", 0.1e-6, 0.5e-6, 0.01e-6, 0.125e-6},
		{"claude-opus-5-5", 4e-6, 20e-6, 0.2e-6, 5e-6},
		{"anthropic/claude-opus-5.5", 4e-6, 20e-6, 0.2e-6, 5e-6},
		{"grok-4.7-latest", 2e-6, 6e-6, 0.5e-6, 0},
	} {
		t.Run(tc.model, func(t *testing.T) {
			p, err := billing.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.InDelta(t, tc.input, p.InputPricePerToken, 1e-12)
			require.InDelta(t, tc.output, p.OutputPricePerToken, 1e-12)
			require.InDelta(t, tc.cached, p.CacheReadPricePerToken, 1e-12)
			require.InDelta(t, tc.write, p.CacheCreationPricePerToken, 1e-12)
		})
	}
}

func TestTier2Grok47BridgeAndMapping(t *testing.T) {
	require.True(t, xai.IsGrokTextResponsesModelID("grok-4.7"))
	require.Equal(t, "grok-4.7", xai.ResolveGrokTextResponsesModelID("grok-4.7-latest"))
	effort, ok := normalizeGrokReasoningEffortValue("xhigh", "grok-4.7")
	require.True(t, ok)
	require.Equal(t, "xhigh", effort)
	require.True(t, grokChatResponsesRuntimeEligible("grok-4.7", "session"))
	account := &Account{Platform: PlatformGrok, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"grok-4.7": "custom-model"}}}
	require.Equal(t, "custom-model", account.GetMappedModel("grok-4.7"))
}
