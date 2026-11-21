package apicompat

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResponses024StopPreserved(t *testing.T) {
	for _, stop := range []string{`"END"`, `["END","STOP"]`, ``} {
		raw := `{"model":"gpt-6.2-sol","input":"hello"`
		if stop != "" {
			raw += `,"stop":` + stop
		}
		raw += `}`
		var req ResponsesRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		chat, err := ResponsesToChatCompletionsRequest(&req)
		require.NoError(t, err)
		if stop != "" {
			require.JSONEq(t, stop, string(chat.Stop))
		} else {
			require.Empty(t, chat.Stop)
		}
	}
}

func TestChat024StopRoundTrip(t *testing.T) {
	req := &ChatCompletionsRequest{Model: "compat-model", Messages: []ChatMessage{{Role: "user", Content: json.RawMessage(`"hello"`)}}, Stop: json.RawMessage(`["END"]`)}
	responses, err := ChatCompletionsToResponses(req)
	require.NoError(t, err)
	chat, err := ResponsesToChatCompletionsRequest(responses)
	require.NoError(t, err)
	require.JSONEq(t, `["END"]`, string(chat.Stop))
}
