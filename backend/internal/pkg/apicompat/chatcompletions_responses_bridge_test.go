package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponsesInputToChatMessages_DeveloperRoleMapsToSystem(t *testing.T) {
	messages, err := responsesInputToChatMessages("", json.RawMessage(`[{"role":"developer","content":"follow project instructions"}]`))
	require.NoError(t, err)
	require.Len(t, messages, 1)

	assert.Equal(t, "system", messages[0].Role)
	assert.JSONEq(t, `"follow project instructions"`, string(messages[0].Content))
}

func TestResponsesInputToChatMessages_KeepsChatCompletionRoles(t *testing.T) {
	input := json.RawMessage(`[
		{"role":"system","content":"system message"},
		{"role":"user","content":"user message"},
		{"role":"assistant","content":"assistant message"},
		{"role":"tool","content":"tool message"}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 4)

	assert.Equal(t, []string{"system", "user", "assistant", "tool"}, chatMessageRoles(messages))
}

func TestResponsesInputToChatMessages_EmptyRoleFallsBackToUser(t *testing.T) {
	messages, err := responsesInputToChatMessages("", json.RawMessage(`[{"role":"","content":"hello"}]`))
	require.NoError(t, err)
	require.Len(t, messages, 1)

	assert.Equal(t, "user", messages[0].Role)
}

func TestResponsesInputToChatMessages_LeadingDeveloperRolesMergeIntoOneSystem(t *testing.T) {
	input := json.RawMessage(`[
		{"role":" Developer ","content":"one"},
		{"role":"\tDEVELOPER\n","content":"two"}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 1)

	assert.Equal(t, []string{"system"}, chatMessageRoles(messages))
	assert.JSONEq(t, `"one\n\ntwo"`, string(messages[0].Content))
}

func TestResponsesToChatCompletionsRequest_InstructionsAndInputDeveloperRole(t *testing.T) {
	req := &ResponsesRequest{
		Model:        "gpt-4o",
		Instructions: "Use concise answers.",
		Input: json.RawMessage(`[
			{"role":"developer","content":[{"type":"input_text","text":"Prefer JSON."}]},
			{"role":"user","content":"Hello"}
		]`),
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	require.Len(t, out.Messages, 2)

	assert.Equal(t, []string{"system", "user"}, chatMessageRoles(out.Messages))
	assert.JSONEq(t, `"Use concise answers.\n\nPrefer JSON."`, string(out.Messages[0].Content))
	assert.JSONEq(t, `"Hello"`, string(out.Messages[1].Content))
}

func TestResponsesInputToChatMessages_MidConversationDeveloperBecomesUser(t *testing.T) {
	input := json.RawMessage(`[
		{"type":"message","role":"user","content":[{"type":"input_text","text":"hello"}]},
		{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}]},
		{"type":"message","role":"developer","content":[{"type":"input_text","text":"<model_switch> switched model"}]},
		{"type":"message","role":"system","content":[{"type":"input_text","text":"be terse"}]},
		{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]}
	]`)

	messages, err := responsesInputToChatMessages("", input)
	require.NoError(t, err)
	require.Len(t, messages, 5)

	assert.Equal(t, []string{"user", "assistant", "user", "user", "user"}, chatMessageRoles(messages))
	assert.JSONEq(t, `"<model_switch> switched model"`, string(messages[2].Content))
	assert.JSONEq(t, `"be terse"`, string(messages[3].Content))
}

func TestResponsesToChatCompletionsRequest_TextFormatJsonObject(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-4o",
		Input: json.RawMessage(`[
			{"role":"user","content":"Return JSON"}
		]`),
		Text: &ResponsesText{
			Format: json.RawMessage(`{"type":"json_object"}`),
		},
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"json_object"}`, string(out.ResponseFormat))
}

func TestResponsesToChatCompletionsRequest_TextFormatJsonSchema(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-4o",
		Input: json.RawMessage(`[
			{"role":"user","content":"Return structured JSON"}
		]`),
		Text: &ResponsesText{
			Format: json.RawMessage(`{
				"type":"json_schema",
				"name":"answer",
				"schema":{
					"type":"object",
					"properties":{"ok":{"type":"boolean"}},
					"required":["ok"],
					"additionalProperties":false
				},
				"strict":true
			}`),
		},
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"type":"json_schema",
		"json_schema":{
			"name":"answer",
			"schema":{
				"type":"object",
				"properties":{"ok":{"type":"boolean"}},
				"required":["ok"],
				"additionalProperties":false
			},
			"strict":true
		}
	}`, string(out.ResponseFormat))
}

func TestResponsesToChatCompletionsRequest_ParallelToolCalls(t *testing.T) {
	parallel := false
	req := &ResponsesRequest{
		Model: "gpt-4o",
		Input: json.RawMessage(`[
			{"role":"user","content":"Use tools"}
		]`),
		ParallelToolCalls: &parallel,
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	require.NotNil(t, out.ParallelToolCalls)
	assert.False(t, *out.ParallelToolCalls)

	payload, err := json.Marshal(out)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"parallel_tool_calls":false`)
}

func chatMessageRoles(messages []ChatMessage) []string {
	roles := make([]string, 0, len(messages))
	for _, message := range messages {
		roles = append(roles, message.Role)
	}
	return roles
}
