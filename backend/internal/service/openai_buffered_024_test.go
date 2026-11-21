//go:build unit

package service

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func read024Buffered(t *testing.T, account *Account, events ...string) *apicompat.ResponsesResponse {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, event := range events {
			_, _ = w.Write([]byte("data: " + event + "\n\n"))
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	final, _, acc, _, err := (&OpenAIGatewayService{}).readOpenAICompatBufferedTerminal(resp, account, "test", "req")
	require.NoError(t, err)
	require.NotNil(t, final)
	if final.Status != "failed" {
		acc.SupplementResponseOutput(final)
	}
	return final
}
func TestBuffered024HostedAndCustomReconstruction(t *testing.T) {
	out := read024Buffered(t, nil,
		`{"type":"response.output_item.done","output_index":0,"item":{"id":"ws_1","type":"web_search_call","status":"completed","action":{"query":"q"}}}`,
		`{"type":"response.output_item.added","output_index":1,"item":{"id":"ct_1","type":"custom_tool_call","call_id":"call_1","name":"exec","input":""}}`,
		`{"type":"response.custom_tool_call_input.delta","item_id":"ct_1","output_index":1,"delta":"dir /b"}`,
		`{"type":"response.output_item.done","output_index":1,"item":{"id":"ct_1","type":"custom_tool_call","call_id":"call_1","name":"exec","input":""}}`,
		`{"type":"response.completed","response":{"id":"r","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2}}}`)
	require.Len(t, out.Output, 2)
	require.Equal(t, "web_search_call", out.Output[0].Type)
	require.Equal(t, "custom_tool_call", out.Output[1].Type)
	require.Equal(t, "dir /b", out.Output[1].Input)
	anthropic := apicompat.ResponsesToAnthropic(out, "m")
	require.Len(t, anthropic.Content, 3)
	require.Equal(t, "server_tool_use", anthropic.Content[0].Type)
	require.JSONEq(t, `{"input":"dir /b"}`, string(anthropic.Content[2].Input))
	chat := apicompat.ResponsesToChatCompletions(out, "m")
	require.Equal(t, "dir /b", chat.Choices[0].Message.ToolCalls[0].Function.Arguments)
}
func TestBuffered024IndependentContentIndexesAndDoneAuthority(t *testing.T) {
	out := read024Buffered(t, nil,
		`{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":0,"delta":"old"}`,
		`{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":1,"delta":"second"}`,
		`{"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"m","role":"assistant","content":[{"type":"output_text","text":"FINAL"},{"type":"output_text","text":""}]}}`,
		`{"type":"response.completed","response":{"id":"r","status":"completed","output":[]}}`)
	require.Len(t, out.Output, 1)
	require.Len(t, out.Output[0].Content, 2)
	require.Equal(t, "FINAL", out.Output[0].Content[0].Text)
	require.Equal(t, "second", out.Output[0].Content[1].Text)
}
func TestBuffered024TerminalAndSearchControls(t *testing.T) {
	final := read024Buffered(t, nil,
		`{"type":"response.output_item.done","output_index":0,"item":{"type":"message","content":[{"type":"output_text","text":"EARLY"}]}}`,
		`{"type":"response.completed","response":{"id":"r","status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"TERMINAL"}]}]}}`)
	require.Equal(t, "TERMINAL", final.Output[0].Content[0].Text)
	failed := read024Buffered(t, nil,
		`{"type":"response.output_item.done","output_index":0,"item":{"type":"message","content":[{"type":"output_text","text":"early"}]}}`,
		`{"type":"response.failed","response":{"id":"r","status":"failed","output":[],"error":{"message":"failed"}}}`)
	require.Empty(t, failed.Output)
	search := read024Buffered(t, &Account{Platform: PlatformGrok},
		`{"type":"response.output_item.done","output_index":0,"item":{"type":"custom_tool_call","call_id":"xs_123","name":"x_keyword_search","input":"{}"}}`,
		`{"type":"response.completed","response":{"id":"r","status":"completed","output":[]}}`)
	require.Empty(t, apicompat.ResponsesToChatCompletions(search, "grok").Choices[0].Message.ToolCalls)
}
func TestBridge024DynamicTools(t *testing.T) {
	body := []byte(`{"model":"alias","tools":[{"type":"function","function":{"name":"first"}}],"messages":[{"role":"system","content":"rules"},{"role":"system","tools":[{"type":"function","function":{"name":"second","parameters":{"type":"object","default":9007199254740993}}}]},{"role":"user","content":"hello","tools":[{"type":"function","function":{"name":"ignored"}}]}]}`)
	var req apicompat.ChatCompletionsRequest
	require.NoError(t, parseChatRequestForBridge(body, &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, &req))
	require.Len(t, req.Messages, 2)
	require.Len(t, req.Tools, 2)
	require.Equal(t, "first", req.Tools[0].Function.Name)
	require.Equal(t, "second", req.Tools[1].Function.Name)
	require.Contains(t, string(req.Tools[1].Function.Parameters), "9007199254740993")
	encoded, err := json.Marshal(req)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(encoded), "9007199254740993"))
}

func TestStop024HTTPFallbackAndNativeWhitelist(t *testing.T) {
	for _, fallback := range []bool{true, false} {
		body := []byte(`{"model":"compat-model","input":"hello","stop":["END"],"stream":false}`)
		account := forceChatResponsesFallbackAccount()
		if !fallback {
			account.Extra = map[string]any{"openai_responses_mode": "force_responses"}
		}
		upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 400, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"done"}}`))}}
		svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(string(body)))
		if fallback {
			_, _ = svc.Forward(context.Background(), c, account, body)
			require.JSONEq(t, `["END"]`, gjson.GetBytes(upstream.lastBody, "stop").Raw)
		} else {
			// The Cursor-shaped Chat entry is the existing native Responses shortcut.
			_, _ = svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
			require.NotNil(t, upstream.lastReq)
			require.False(t, gjson.GetBytes(upstream.lastBody, "stop").Exists())
		}
	}
}

func TestBuffered024HostedRawFields(t *testing.T) {
	for _, kind := range []string{"file_search_call", "computer_call", "code_interpreter_call", "image_generation_call", "web_search_call", "x_search_call"} {
		out := read024Buffered(t, nil,
			`{"type":"response.output_item.done","output_index":0,"item":{"type":"`+kind+`","id":"hosted","status":"completed","provider_payload":{"nested":9007199254740993},"action":{"query":"q","sources":[{"url":"https://example.invalid"}]}}}`,
			`{"type":"response.completed","response":{"id":"r","status":"completed","output":[]}}`)
		require.Len(t, out.Output, 1)
		raw, err := json.Marshal(out.Output[0])
		require.NoError(t, err)
		require.Equal(t, "9007199254740993", gjson.GetBytes(raw, "provider_payload.nested").Raw)
		require.Equal(t, "https://example.invalid", gjson.GetBytes(raw, "action.sources.0.url").String())
	}
}
