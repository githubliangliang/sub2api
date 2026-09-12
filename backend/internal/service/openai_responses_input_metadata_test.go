package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestStripOpenAIResponsesInputContentItemKindsPreservesOtherFields(t *testing.T) {
	item := json.RawMessage(`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"content_item_kinds","internal_chat_message_metadata_passthrough":{"content_item_kinds":["nested-keep"]}}],"internal_chat_message_metadata_passthrough":{"content_item_kinds":["output_text"],"keep":9007199254740993},"sequence":9007199254740993}`)
	items := make([]json.RawMessage, 8)
	for index := range items {
		items[index] = item
	}
	body, err := json.Marshal(map[string]any{"input": items, "sequence": json.Number("9007199254740993")})
	require.NoError(t, err)
	original := string(body)

	stripped, err := stripOpenAIResponsesInputContentItemKinds(body)

	require.NoError(t, err)
	require.Len(t, gjson.GetBytes(stripped, "input").Array(), 8)
	for _, output := range gjson.GetBytes(stripped, "input").Array() {
		require.False(t, output.Get("internal_chat_message_metadata_passthrough.content_item_kinds").Exists())
		require.Equal(t, "9007199254740993", output.Get("internal_chat_message_metadata_passthrough.keep").Raw)
		require.Equal(t, "9007199254740993", output.Get("sequence").Raw)
		require.Equal(t, "content_item_kinds", output.Get("content.0.text").String())
		require.Equal(t, "nested-keep", output.Get("content.0.internal_chat_message_metadata_passthrough.content_item_kinds.0").String())
	}
	require.Equal(t, "9007199254740993", gjson.GetBytes(stripped, "sequence").Raw)
	require.Equal(t, original, string(body))
	secondPass, err := stripOpenAIResponsesInputContentItemKinds(stripped)
	require.NoError(t, err)
	require.Equal(t, stripped, secondPass)
}

func TestStripOpenAIResponsesInputContentItemKindsLeavesOtherShapesByteExact(t *testing.T) {
	for _, body := range []string{
		`{ "input": "content_item_kinds" }`,
		`{"input":null}`,
		`{"input":[]}`,
		`{"input":[null,7,"content_item_kinds",{"content":"hello"}]}`,
		`{"input":[{"internal_chat_message_metadata_passthrough":{"keep":true}}]}`,
		`{"input":[{"content":[{"internal_chat_message_metadata_passthrough":{"content_item_kinds":["nested-keep"]}}]}]}`,
		`{"internal_chat_message_metadata_passthrough":{"content_item_kinds":["root-keep"]}}`,
	} {
		stripped, err := stripOpenAIResponsesInputContentItemKinds([]byte(body))
		require.NoError(t, err)
		require.Equal(t, body, string(stripped))
	}
}

func TestStripOpenAIResponsesInputContentItemKindsHandlesEmptyAndEscapedMetadata(t *testing.T) {
	for _, metadata := range []string{`{"content_item_kinds":null}`, `{"content_item_kin\u0064s":[]}`} {
		body := []byte(`{"input":[{"internal_chat_message_metadata_passthrough":` + metadata + `}]}`)

		stripped, err := stripOpenAIResponsesInputContentItemKinds(body)

		require.NoError(t, err)
		require.JSONEq(t, `{"input":[{"internal_chat_message_metadata_passthrough":{}}]}`, string(stripped))
	}
}

func TestPrepareOpenAIWSHTTPBridgeBodyStripsInputContentItemKinds(t *testing.T) {
	for _, account := range []*Account{
		newOpenAIOAuthNamespaceTestAccount(),
		newOpenAIRejectedFieldTestAccount(),
		{Platform: PlatformGrok, Type: AccountTypeAPIKey},
	} {
		t.Run(account.Platform+"/"+account.Type, func(t *testing.T) {
			payload := []byte(`{"type":"response.create","model":"gpt-5.5","input":[{"role":"assistant","content":"hello","internal_chat_message_metadata_passthrough":{"content_item_kinds":["output_text"],"keep":9007199254740993}}]}`)

			body, err := prepareOpenAIWSHTTPBridgeBody(account, payload)

			require.NoError(t, err)
			require.Equal(t, account.Platform == PlatformGrok, gjson.GetBytes(body, "input.0.internal_chat_message_metadata_passthrough.content_item_kinds").Exists())
			require.Equal(t, "9007199254740993", gjson.GetBytes(body, "input.0.internal_chat_message_metadata_passthrough.keep").Raw)
			require.Equal(t, "hello", gjson.GetBytes(body, "input.0.content").String())
		})
	}
}

func TestOpenAIGatewayService_HTTPStripsInputContentItemKindsBeforeFirstForward(t *testing.T) {
	for _, accountType := range []string{"oauth", "apikey"} {
		for _, passthrough := range []bool{false, true} {
			for _, path := range []string{"/v1/responses", "/v1/responses/compact"} {
				t.Run(fmt.Sprintf("%s/passthrough=%t%s", accountType, passthrough, path), func(t *testing.T) {
					account := newOpenAIRejectedFieldTestAccount()
					if accountType == "oauth" {
						account = newOpenAIOAuthNamespaceTestAccount()
					}
					if account.Extra == nil {
						account.Extra = make(map[string]any)
					}
					account.Extra["openai_passthrough"] = passthrough
					body := []byte(`{
						"model":"gpt-5.5","stream":false,"instructions":"test",
						"input":[
							{"type":"message","role":"user","content":[{"type":"input_text","text":"hello"}],"internal_chat_message_metadata_passthrough":{"content_item_kinds":["input_text"],"keep":"user-metadata"}},
							{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}],"internal_chat_message_metadata_passthrough":{"content_item_kinds":["output_text"],"keep":"assistant-metadata"}}
						]
					}`)
					response := newOpenAIRejectedFieldTestResponse(http.StatusOK, `{"id":"resp_metadata_ok","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)
					if accountType == "oauth" && passthrough && path == "/v1/responses" {
						response = newOpenAIRejectedFieldTestResponse(http.StatusOK,
							"data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n"+
								"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_metadata_ok\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\n")
						response.Header.Set("Content-Type", "text/event-stream")
					}
					upstream := &httpUpstreamRecorder{responses: []*http.Response{response}}
					c := newOpenAIRejectedFieldTestContext(body)
					c.Request.URL.Path = path

					result, err := newOpenAIRejectedFieldTestService(upstream).Forward(context.Background(), c, account, body)

					require.NoError(t, err)
					require.NotNil(t, result)
					require.Len(t, upstream.bodies, 1)
					forwarded := upstream.bodies[0]
					for index, want := range []string{"user-metadata", "assistant-metadata"} {
						metadataPath := fmt.Sprintf("input.%d.internal_chat_message_metadata_passthrough", index)
						require.False(t, gjson.GetBytes(forwarded, metadataPath+".content_item_kinds").Exists(), "rejected metadata must be removed before the first upstream request")
						require.Equal(t, want, gjson.GetBytes(forwarded, metadataPath+".keep").String())
					}
					require.Equal(t, "hello", gjson.GetBytes(forwarded, "input.0.content.0.text").String())
					require.Equal(t, "hi", gjson.GetBytes(forwarded, "input.1.content.0.text").String())
					require.True(t, gjson.GetBytes(body, "input.1.internal_chat_message_metadata_passthrough.content_item_kinds").Exists(), "the original request must remain available for account failover")
				})
			}
		}
	}
}
