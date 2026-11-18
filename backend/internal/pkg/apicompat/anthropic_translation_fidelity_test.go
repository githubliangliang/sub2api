package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAnthropicToResponses_MidConversationSystemAuthority(t *testing.T) {
	out, err := AnthropicToResponses(&AnthropicRequest{
		Model: "gpt-6.1-sol", System: json.RawMessage(`"initial instruction"`),
		Messages: []AnthropicMessage{
			{Role: "user", Content: json.RawMessage(`"user text"`)},
			{Role: "assistant", Content: json.RawMessage(`[{"type":"tool_use","id":"call_1","name":"run","input":{}}]`)},
			{Role: "system", Content: json.RawMessage(`[{"type":"text","text":"hook output"},{"type":"tool_result","tool_use_id":"call_1","content":"untrusted tool result"}]`)},
			{Role: "system", Content: json.RawMessage(`"extra instruction"`)},
			{Role: "user", Content: json.RawMessage(`"last user text"`)},
		},
	})
	require.NoError(t, err)
	input := gjson.ParseBytes(out.Input).Array()
	require.Len(t, input, 7)
	require.Equal(t, "developer", input[0].Get("role").String())
	require.Equal(t, "initial instruction", input[0].Get("content.0.text").String())
	require.Equal(t, "user", input[1].Get("role").String())
	require.Equal(t, "function_call", input[2].Get("type").String())
	require.Equal(t, "function_call_output", input[3].Get("type").String())
	require.Equal(t, "call_1", input[3].Get("call_id").String())
	require.False(t, input[3].Get("role").Exists())
	require.Contains(t, input[3].Get("output").String(), "untrusted tool result")
	require.Equal(t, "developer", input[4].Get("role").String())
	require.Equal(t, "hook output", input[4].Get("content.0.text").String())
	require.Equal(t, "developer", input[5].Get("role").String())
	require.Equal(t, "user", input[6].Get("role").String())
}

func TestAnthropicToResponses_SystemToolResultMediaStaysUser(t *testing.T) {
	items, err := anthropicMsgToResponsesItems(AnthropicMessage{
		Role: "system", Content: json.RawMessage(`[{"type":"text","text":"instruction"},{"type":"tool_result","tool_use_id":"call_1","content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aGVsbG8="}}]}]`),
	})
	require.NoError(t, err)
	require.Len(t, items, 3)
	require.Equal(t, "function_call_output", items[0].Type)
	require.Equal(t, "developer", items[1].Role)
	require.Equal(t, "instruction", gjson.GetBytes(items[1].Content, "0.text").String())
	require.Equal(t, "user", items[2].Role)
	require.Equal(t, "input_image", gjson.GetBytes(items[2].Content, "0.type").String())
}

func TestChatResponseFormatStrictPresence(t *testing.T) {
	for _, strict := range []string{"", "true", "false", "null"} {
		t.Run("strict="+strict, func(t *testing.T) {
			schema := `{"name":"answer","schema":{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"]}`
			if strict != "" {
				schema += `,"strict":` + strict
			}
			schema += `}`
			raw := json.RawMessage(`{"type":"json_schema","json_schema":` + schema + `}`)
			original := append(json.RawMessage(nil), raw...)
			out := chatResponseFormatToResponsesTextFormat(raw)
			want := strict
			if want == "" {
				want = "false"
			}
			require.Equal(t, want, gjson.GetBytes(out, "strict").Raw)
			require.JSONEq(t, gjson.Get(schema, "schema").Raw, gjson.GetBytes(out, "schema").Raw)
			require.Equal(t, original, raw)
		})
	}
	for _, raw := range []string{`{"type":"text"}`, `{"type":"json_object"}`} {
		require.Equal(t, json.RawMessage(raw), chatResponseFormatToResponsesTextFormat(json.RawMessage(raw)))
	}
}
