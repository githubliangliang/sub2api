package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func convertLegacyFunctionHistory(t *testing.T, msgs []ChatMessage) []ResponsesInputItem {
	t.Helper()
	resp, err := ChatCompletionsToResponses(&ChatCompletionsRequest{Model: "gpt-4o", Messages: msgs})
	require.NoError(t, err)
	var items []ResponsesInputItem
	require.NoError(t, json.Unmarshal(resp.Input, &items))
	return items
}

// Legacy function calling has no call IDs: the assistant turn carries
// function_call and the result is a role=function message keyed by name. The
// Responses API needs a function_call item and a function_call_output with the
// same call_id, otherwise the output is rejected as having no matching call.
func TestChatCompletionsToResponses_LegacyFunctionCallHistoryPaired(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: json.RawMessage(`"Weather in Paris and Rome?"`)},
		{Role: "assistant", Content: json.RawMessage(`"Checking Paris."`), FunctionCall: &ChatFunctionCall{Name: "get_weather", Arguments: `{"city":"Paris"}`}},
		{Role: "function", Name: "get_weather", Content: json.RawMessage(`"sunny"`)},
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "get_weather", Arguments: `{"city":"Rome"}`}},
		{Role: "function", Name: "get_weather", Content: json.RawMessage(`"rainy"`)},
	}

	items := convertLegacyFunctionHistory(t, msgs)
	require.Len(t, items, 6)

	require.Equal(t, "user", items[0].Role)
	require.Equal(t, "assistant", items[1].Role)

	firstCall, firstOutput := items[2], items[3]
	require.Equal(t, "function_call", firstCall.Type)
	require.Equal(t, "get_weather", firstCall.Name)
	require.Equal(t, `{"city":"Paris"}`, firstCall.Arguments)
	require.Equal(t, "function_call_output", firstOutput.Type)
	require.Equal(t, "sunny", firstOutput.Output)
	require.NotEmpty(t, firstCall.CallID)
	require.Equal(t, firstCall.CallID, firstOutput.CallID)

	secondCall, secondOutput := items[4], items[5]
	require.Equal(t, "function_call", secondCall.Type)
	require.Equal(t, `{"city":"Rome"}`, secondCall.Arguments)
	require.Equal(t, "function_call_output", secondOutput.Type)
	require.Equal(t, "rainy", secondOutput.Output)
	require.Equal(t, secondCall.CallID, secondOutput.CallID)
	require.NotEqual(t, firstCall.CallID, secondCall.CallID)

	// Synthesized IDs must be stable across turns so the replayed history keeps
	// the same prompt prefix.
	require.Equal(t, items, convertLegacyFunctionHistory(t, msgs))
}

func TestChatCompletionsToResponses_LegacyFunctionCallIDAvoidsToolCallIDs(t *testing.T) {
	items := convertLegacyFunctionHistory(t, []ChatMessage{
		{Role: "user", Content: json.RawMessage(`"hi"`)},
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "lookup", Arguments: `{}`}},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"legacy result"`)},
		{Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call_legacy_1", Type: "function", Function: ChatFunctionCall{Name: "lookup", Arguments: `{}`}}}},
		{Role: "tool", ToolCallID: "call_legacy_1", Content: json.RawMessage(`"tool result"`)},
	})
	require.Len(t, items, 5)

	legacyCall, legacyOutput := items[1], items[2]
	require.Equal(t, "function_call", legacyCall.Type)
	require.Equal(t, legacyCall.CallID, legacyOutput.CallID)
	require.NotEqual(t, "call_legacy_1", legacyCall.CallID)

	require.Equal(t, "call_legacy_1", items[3].CallID)
	require.Equal(t, "call_legacy_1", items[4].CallID)
}

func TestChatCompletionsToResponses_LegacyFunctionResultWithoutCallKeepsName(t *testing.T) {
	items := convertLegacyFunctionHistory(t, []ChatMessage{
		{Role: "user", Content: json.RawMessage(`"hi"`)},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"orphan"`)},
	})
	require.Len(t, items, 2)
	require.Equal(t, "function_call_output", items[1].Type)
	require.Equal(t, "lookup", items[1].CallID)
}

func TestChatCompletionsToResponses_LegacyFunctionQueueAndModernPrecedence(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "lookup", Arguments: `{"n":1}`}},
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "lookup", Arguments: `{"n":2}`}},
		{Role: "function", Name: "other", Content: json.RawMessage(`"orphan"`)},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"first"`)},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"second"`)},
		{Role: "assistant", Content: json.RawMessage(`"text"`), FunctionCall: &ChatFunctionCall{Name: "ignored", Arguments: `{}`}, ToolCalls: []ChatToolCall{{ID: "call_legacy_1", Type: "function", Function: ChatFunctionCall{Name: "modern", Arguments: `{}`}}}},
		{Role: "tool", ToolCallID: "call_legacy_2", Content: json.RawMessage(`"reserved"`)},
	}
	items := convertLegacyFunctionHistory(t, msgs)
	require.Len(t, items, 8)
	require.Equal(t, "call_legacy_3", items[0].CallID)
	require.Equal(t, "call_legacy_4", items[1].CallID)
	require.Equal(t, "other", items[2].CallID)
	require.Equal(t, items[0].CallID, items[3].CallID)
	require.Equal(t, items[1].CallID, items[4].CallID)
	require.Equal(t, "assistant", items[5].Role)
	require.Equal(t, "modern", items[6].Name)
	require.Nil(t, msgs[0].ToolCalls, "conversion must not mutate the replayed history")
}
