package apicompat

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfiguredMinOutputTokens_AllConverters(t *testing.T) {
	for _, floor := range []int{16, 128, 256} {
		for _, input := range []int{0, -1, 1, 16, 128, 512} {
			t.Run(fmt.Sprintf("floor_%d_input_%d", floor, input), func(t *testing.T) {
				anthropic := &AnthropicRequest{Model: "gpt-4o", MaxTokens: input,
					Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}}
				responses, err := AnthropicToResponsesWithMinOutputTokens(anthropic, floor)
				require.NoError(t, err)
				chat, err := AnthropicToChatCompletionsRequestWithMinOutputTokens(anthropic, floor)
				require.NoError(t, err)
				cc := &ChatCompletionsRequest{Model: "gpt-4o", MaxTokens: &input,
					Messages: []ChatMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}}
				fromChat, err := ChatCompletionsToResponsesWithMinOutputTokens(cc, floor)
				require.NoError(t, err)
				for _, value := range []*int{responses.MaxOutputTokens, chat.MaxCompletionTokens, fromChat.MaxOutputTokens} {
					if input <= 0 {
						require.Nil(t, value)
					} else {
						require.NotNil(t, value)
						require.Equal(t, max(input, floor), *value)
					}
				}
			})
		}
	}
}

func TestConfiguredMinOutputTokens_FieldPrecedenceAndLegacyDefault(t *testing.T) {
	high, low := 500, 1
	req := &ChatCompletionsRequest{Model: "gpt-4o", MaxTokens: &high, MaxCompletionTokens: &low,
		Messages: []ChatMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}}
	configured, err := ChatCompletionsToResponsesWithMinOutputTokens(req, 16)
	require.NoError(t, err)
	require.Equal(t, 16, *configured.MaxOutputTokens)
	legacy, err := ChatCompletionsToResponses(req)
	require.NoError(t, err)
	require.Equal(t, 128, *legacy.MaxOutputTokens)
	req.MaxTokens, req.MaxCompletionTokens = nil, nil
	missing, err := ChatCompletionsToResponsesWithMinOutputTokens(req, 256)
	require.NoError(t, err)
	require.Nil(t, missing.MaxOutputTokens)
}
