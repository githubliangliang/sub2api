package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func TestOpenAIForward_InstructionsFollowAccountProtocol(t *testing.T) {
	explicit, empty := "Only answer in JSON.", ""
	for _, tc := range []struct {
		name         string
		oauth        bool
		passthrough  bool
		instructions *string
		wantDefault  bool
	}{
		{name: "api_key_omitted"},
		{name: "api_key_explicit", instructions: &explicit},
		{name: "api_key_empty", instructions: &empty},
		{name: "api_key_passthrough", passthrough: true},
		{name: "oauth_default", oauth: true, wantDefault: true},
		{name: "oauth_explicit", oauth: true, instructions: &explicit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := newOpenAIRejectedFieldTestAccount()
			if tc.oauth {
				account = newOpenAIOAuthNamespaceTestAccount()
			}
			if account.Extra == nil {
				account.Extra = map[string]any{}
			}
			account.Extra["openai_passthrough"] = tc.passthrough
			body := []byte(`{"model":"gpt-5.5","stream":false,"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello"}],"internal_chat_message_metadata_passthrough":{"content_item_kinds":["input_text"],"keep":"user-metadata"}}]}`)
			if tc.instructions != nil {
				var err error
				body, err = sjson.SetBytes(body, "instructions", *tc.instructions)
				require.NoError(t, err)
			}
			upstream := &httpUpstreamRecorder{resp: newOpenAIRejectedFieldTestResponse(http.StatusOK, `{"id":"resp_scope","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)}
			result, err := newOpenAIRejectedFieldTestService(upstream).Forward(context.Background(), newOpenAIRejectedFieldTestContext(body), account, body)
			require.NoError(t, err)
			require.NotNil(t, result)
			instructions := gjson.GetBytes(upstream.lastBody, "instructions")
			switch {
			case tc.wantDefault:
				require.NotEmpty(t, instructions.String())
			case tc.instructions != nil:
				require.True(t, instructions.Exists())
				require.Equal(t, *tc.instructions, instructions.String())
			default:
				require.False(t, instructions.Exists())
			}
			require.False(t, gjson.GetBytes(upstream.lastBody, "input.0.internal_chat_message_metadata_passthrough.content_item_kinds").Exists())
			require.Equal(t, "user-metadata", gjson.GetBytes(upstream.lastBody, "input.0.internal_chat_message_metadata_passthrough.keep").String())
			require.True(t, gjson.GetBytes(body, "input.0.internal_chat_message_metadata_passthrough.content_item_kinds").Exists())
		})
	}
}
