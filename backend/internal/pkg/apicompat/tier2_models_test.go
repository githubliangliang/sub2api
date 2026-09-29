package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTier2ReasoningGenerations(t *testing.T) {
	for _, tc := range []struct {
		model     string
		reasoning bool
	}{
		{"gpt-6-astra", true}, {"gpt-6-sol", true}, {"gpt-6-luna", true},
		{"gpt-7-future", true}, {"gpt-5.5", true}, {"GPT-6-Astra", true},
		{"gpt-4o", false}, {"gpt-image-1", false}, {"gpt-audio", false},
	} {
		t.Run(tc.model, func(t *testing.T) {
			temp := 0.7
			out, err := AnthropicToResponses(&AnthropicRequest{Model: tc.model, Temperature: &temp, TopP: &temp, Messages: []AnthropicMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}})
			require.NoError(t, err)
			if tc.reasoning {
				require.Nil(t, out.Temperature)
				require.Nil(t, out.TopP)
			} else {
				require.Equal(t, &temp, out.Temperature)
				require.Equal(t, &temp, out.TopP)
			}
		})
	}
}

func TestTier2OpusAdaptiveAndSignedHistory(t *testing.T) {
	for _, model := range []string{"claude-opus-5-5", "anthropic/claude-opus-5.5"} {
		t.Run(model, func(t *testing.T) {
			out, err := ResponsesToAnthropicRequest(&ResponsesRequest{Model: model, Input: json.RawMessage(`"hello"`), Reasoning: &ResponsesReasoning{Effort: "xhigh"}})
			require.NoError(t, err)
			require.NotNil(t, out.Thinking)
			require.Equal(t, "adaptive", out.Thinking.Type)
			require.Equal(t, "xhigh", out.OutputConfig.Effort)
			_, err = ResponsesToAnthropicRequest(&ResponsesRequest{Model: model, Input: json.RawMessage(`"hello"`), ToolChoice: json.RawMessage(`"required"`)})
			require.ErrorContains(t, err, "forced tool_choice")
			resp := AnthropicToResponsesResponse(&AnthropicResponse{Model: model, Content: []AnthropicContentBlock{{Type: "thinking", Signature: "signed-history"}, {Type: "text", Text: "answer"}}})
			require.Len(t, resp.Output, 2)
			require.NotEmpty(t, resp.Output[0].EncryptedContent)
		})
	}
}
