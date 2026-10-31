//go:build unit

package apicompat

import (
	"encoding/json"
	"testing"
)

// Regression coverage for the selected historical compatibility fixes.
func TestResponsesChatHistoryFiltersInvalidFunctionCall(t *testing.T) {
	input, err := json.Marshal([]map[string]any{
		{"type": "function_call", "call_id": "call_bad", "name": "read_file", "arguments": `{"path":`},
		{"type": "function_call_output", "call_id": "call_bad", "output": "failed"},
		{"role": "user", "content": "continue"},
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := ResponsesToChatCompletionsRequest(&ResponsesRequest{Model: "gpt-5.4", Input: input})
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range req.Messages {
		for _, call := range message.ToolCalls {
			if call.ID == "call_bad" {
				t.Errorf("invalid historical function call forwarded: arguments=%q", call.Function.Arguments)
			}
		}
		if message.ToolCallID == "call_bad" {
			t.Error("matching output of invalid historical call also forwarded")
		}
	}
}

func TestResponsesChatOutputRejectsInvalidFunctionArguments(t *testing.T) {
	for _, arguments := range []string{`{"path":`, `{"path":"ok"}`} {
		t.Run(arguments, func(t *testing.T) {
			resp := ChatCompletionsResponseToResponses(&ChatCompletionsResponse{
				ID: "chat_review", Choices: []ChatChoice{{Message: ChatMessage{
					Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call_review", Type: "function", Function: ChatFunctionCall{Name: "read_file", Arguments: arguments}}},
				}, FinishReason: "tool_calls"}},
			}, "gpt-5.4", nil, false, nil)
			found := false
			for _, item := range resp.Output {
				if item.Type == "function_call" {
					found = true
					if !json.Valid([]byte(item.Arguments)) {
						t.Errorf("invalid function call emitted with status=%q arguments=%q", item.Status, item.Arguments)
					}
				}
			}
			if json.Valid([]byte(arguments)) && !found {
				t.Error("valid function call was lost")
			}
		})
	}
}

func TestResponsesChatCustomToolNamespaceAlias(t *testing.T) {
	resp := ChatCompletionsResponseToResponses(&ChatCompletionsResponse{
		ID: "chat_alias", Choices: []ChatChoice{{Message: ChatMessage{
			Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call_alias", Type: "function", Function: ChatFunctionCall{Name: "functions__exec", Arguments: `{"input":"pwd"}`}}},
		}, FinishReason: "tool_calls"}},
	}, "gpt-5.4", map[string]bool{"exec": true}, false, map[string]NamespacedToolName{"functions__wait": {Namespace: "functions", Name: "wait"}})
	if len(resp.Output) != 1 {
		t.Fatalf("unexpected output count: %d", len(resp.Output))
	}
	item := resp.Output[0]
	if item.Type != "custom_tool_call" || item.Name != "exec" || item.Input != "pwd" {
		t.Fatalf("unambiguous custom alias not restored: type=%q name=%q input=%q", item.Type, item.Name, item.Input)
	}
}

func TestResponsesClientToolsClearsOrphanDeferredFlags(t *testing.T) {
	req := map[string]any{"tools": []any{
		map[string]any{"type": "function", "name": "read_file", "parameters": map[string]any{"type": "object"}, "defer_loading": true},
		map[string]any{"type": "tool_search"},
	}}
	if _, _, err := AdaptResponsesClientTools(req); err != nil {
		t.Fatal(err)
	}
	for _, raw := range req["tools"].([]any) {
		tool := raw.(map[string]any)
		if _, exists := tool["defer_loading"]; exists {
			t.Errorf("orphan defer_loading remains after built-in tool_search was lowered: %v", tool)
		}
	}
}

func TestResponsesStreamEventIncludesZeroSequence(t *testing.T) {
	for _, sequence := range []int{0, 1} {
		data, err := json.Marshal(ResponsesStreamEvent{Type: "response.created", SequenceNumber: sequence})
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(data, &object); err != nil {
			t.Fatal(err)
		}
		if _, exists := object["sequence_number"]; !exists {
			t.Errorf("sequence=%d omitted from serialized event: %s", sequence, data)
		}
	}
}

func TestResponsesChatToolIdentityAndArgumentBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, toolName, arguments, finish, wantType, wantName, wantNamespace string
		custom, functions                                                    map[string]bool
		namespaces                                                           map[string]NamespacedToolName
	}{
		{name: "valid length", toolName: "read", arguments: `{}`, finish: "length", wantType: "function_call", wantName: "read"},
		{name: "empty ordinary", toolName: "read", finish: "tool_calls", wantType: "function_call", wantName: "read"},
		{name: "freeform custom", toolName: "exec", arguments: "print('ok')", finish: "tool_calls", custom: map[string]bool{"exec": true}, wantType: "custom_tool_call", wantName: "exec"},
		{name: "explicit function wins", toolName: "functions__exec", arguments: `{}`, finish: "tool_calls", custom: map[string]bool{"exec": true}, functions: map[string]bool{"functions__exec": true}, namespaces: map[string]NamespacedToolName{"functions__wait": {Namespace: "functions", Name: "wait"}}, wantType: "function_call", wantName: "functions__exec"},
		{name: "namespace child wins", toolName: "functions__exec", arguments: `{}`, finish: "tool_calls", custom: map[string]bool{"exec": true}, namespaces: map[string]NamespacedToolName{"functions__exec": {Namespace: "functions", Name: "exec"}}, wantType: "function_call", wantName: "exec", wantNamespace: "functions"},
		{name: "unknown prefix stays ordinary", toolName: "other__exec", arguments: `{}`, finish: "tool_calls", custom: map[string]bool{"exec": true}, namespaces: map[string]NamespacedToolName{"functions__wait": {Namespace: "functions", Name: "wait"}}, wantType: "function_call", wantName: "other__exec"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := ChatCompletionsResponseToResponses(&ChatCompletionsResponse{Choices: []ChatChoice{{
				Message:      ChatMessage{Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call", Function: ChatFunctionCall{Name: tc.toolName, Arguments: tc.arguments}}}},
				FinishReason: tc.finish,
			}}}, "gpt-5.4", tc.custom, false, tc.namespaces, tc.functions)
			if len(response.Output) != 1 {
				t.Fatalf("output: %+v", response.Output)
			}
			item := response.Output[0]
			if item.Type != tc.wantType || item.Name != tc.wantName || item.Namespace != tc.wantNamespace {
				t.Fatalf("identity: %+v", item)
			}
			if tc.finish == "length" && response.Status != "incomplete" {
				t.Fatal("length response lost incomplete status")
			}
			if tc.arguments == "" && item.Arguments != "{}" {
				t.Fatal("empty argument default changed")
			}
		})
	}
}

func TestResponsesChatStreamValidLengthAndCustomAlias(t *testing.T) {
	for _, custom := range []bool{false, true} {
		state := NewChatCompletionsToResponsesStreamState("gpt-5.4")
		name, arguments := "read", `{}`
		if custom {
			state.CustomTools = map[string]bool{"exec": true}
			state.NamespaceTools = map[string]NamespacedToolName{"functions__wait": {Namespace: "functions", Name: "wait"}}
			name, arguments = "functions__exec", `{"input":"pwd"}`
		}
		index, finish := 0, "length"
		ChatCompletionsChunkToResponsesEvents(&ChatCompletionsChunk{Choices: []ChatChunkChoice{{Delta: ChatDelta{ToolCalls: []ChatToolCall{{Index: &index, ID: "call", Function: ChatFunctionCall{Name: name, Arguments: arguments}}}}, FinishReason: &finish}}}, state)
		if err := state.ValidateToolCallArguments(); err != nil {
			t.Fatal(err)
		}
		events := FinalizeChatCompletionsResponsesStream(state)
		found := false
		for _, event := range events {
			if event.Type == "response.output_item.done" && event.Item != nil && event.Item.CallID == "call" {
				found = true
				if custom && (event.Item.Type != "custom_tool_call" || event.Item.Name != "exec" || event.Item.Input != "pwd") {
					t.Fatalf("custom alias: %+v", event.Item)
				}
			}
		}
		if !found {
			t.Fatal("valid tool completion missing")
		}
	}
}

func TestResponsesDeferredFlagsRespectBuiltinAndDeclarationOwnership(t *testing.T) {
	declaration := map[string]any{"type": "function", "name": "read", "defer_loading": true}
	tools := []any{declaration, map[string]any{"type": "tool_search"}}
	if stripResponsesDeferredToolFlags(tools) {
		t.Fatal("valid built-in search lost deferred flags")
	}
	tools = []any{declaration, map[string]any{"type": "function", "name": "tool_search"}}
	if !stripResponsesDeferredToolFlags(tools) {
		t.Fatal("same-named function was treated as built-in")
	}
	if _, ok := tools[0].(map[string]any)["defer_loading"]; ok {
		t.Fatal("orphan flag remains")
	}
	if declaration["defer_loading"] != true {
		t.Fatal("shared declaration source mutated")
	}
}

func TestResponsesInvalidEmptyIDDoesNotDiscardCustomHistory(t *testing.T) {
	input, _ := json.Marshal([]map[string]any{
		{"type": "function_call", "name": "bad", "arguments": `{"x":`},
		{"type": "custom_tool_call", "name": "exec", "input": "pwd"},
		{"type": "custom_tool_call_output", "output": "custom output"},
		{"type": "function_call_output", "output": "bad output"},
		{"role": "user", "content": "continue"},
	})
	request, err := ResponsesToChatCompletionsRequest(&ResponsesRequest{Model: "gpt-5.4", Input: input})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, message := range request.Messages {
		if string(message.Content) == `"bad output"` {
			t.Fatal("invalid ordinary output replayed")
		}
		if string(message.Content) == `"custom output"` {
			found = true
		}
	}
	if !found {
		t.Fatal("valid custom output was discarded")
	}
}
