package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponsesConvertersSerializeCreatedAt(t *testing.T) {
	assertTime := func(t *testing.T, response *ResponsesResponse, want int64) {
		t.Helper()
		encoded, err := json.Marshal(response)
		require.NoError(t, err)
		var wire map[string]any
		require.NoError(t, json.Unmarshal(encoded, &wire))
		value, ok := wire["created_at"].(float64)
		require.True(t, ok, "created_at must be a JSON number: %s", encoded)
		require.Positive(t, value)
		if want > 0 {
			require.EqualValues(t, want, value)
		}
	}
	assertTime(t, AnthropicToResponsesResponse(&AnthropicResponse{ID: "m", Model: "claude-sonnet-4-6"}), 0)
	assertTime(t, ChatCompletionsResponseToResponses(&ChatCompletionsResponse{ID: "c", Created: 1700000000}, "gpt-6.1-sol", nil, false, nil), 1700000000)
	anthropic := NewAnthropicEventToResponsesState()
	anthropic.Created = 1700000001
	assertTime(t, makeResponsesCreatedEvent(anthropic).Response, anthropic.Created)
	assertTime(t, makeResponsesCompletedEvent(anthropic, "completed", nil).Response, anthropic.Created)
	chat := NewChatCompletionsToResponsesStreamState("gpt-6.1-sol")
	chat.Created = 1700000002
	assertTime(t, ensureChatToResponsesCreated(chat)[0].Response, chat.Created)
}
