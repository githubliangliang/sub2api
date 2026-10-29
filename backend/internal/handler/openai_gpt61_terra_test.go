//go:build unit

package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT61TerraWebSocketUsagePreservesEffortAndMapping(t *testing.T) {
	for _, effort := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
		t.Run(effort, func(t *testing.T) {
			got := runOpenAIResponsesWebSocketUsageLogCase(t, openAIResponsesWSUsageLogCase{
				firstPayload:        fmt.Sprintf(`{"type":"response.create","model":"public-terra","stream":false,"reasoning":{"effort":%q}}`, effort),
				userAgent:           testStringPtr("codex_cli_rs/0.125.0 terra"),
				channelMapping:      map[string]string{"public-terra": "terra-channel"},
				accountModelMapping: map[string]any{"terra-channel": "gpt-6.1-terra", "public-terra": "gpt-6.1-terra"},
			})
			require.Equal(t, "gpt-6.1-terra", gjson.GetBytes(got.upstreamFirstPayload, "model").String())
			require.Equal(t, effort, gjson.GetBytes(got.upstreamFirstPayload, "reasoning.effort").String())
			require.NotNil(t, got.log.ReasoningEffort)
			require.Equal(t, effort, *got.log.ReasoningEffort)
			require.Equal(t, "public-terra", got.log.RequestedModel)
			require.NotNil(t, got.log.UpstreamModel)
			require.Equal(t, "gpt-6.1-terra", *got.log.UpstreamModel)
			require.NotNil(t, got.log.UpstreamResponseModel)
			require.Equal(t, "gpt-6.1-terra", *got.log.UpstreamResponseModel)
			require.NotNil(t, got.log.ModelMappingChain)
			require.Equal(t, "public-terra→terra-channel→gpt-6.1-terra", *got.log.ModelMappingChain)
			require.True(t, got.log.OpenAIWSMode)
			require.Equal(t, 2, got.log.InputTokens)
			require.Equal(t, 1, got.log.OutputTokens)
			require.InDelta(t, 7e-6, got.log.TotalCost, 1e-12)
			require.Equal(t, "public-terra", gjson.GetBytes(got.clientEvents[0], "response.model").String())
		})
	}
}
