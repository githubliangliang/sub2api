package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOAuthWebSearchHistoryPassthroughTwoTurns(t *testing.T) {
	for _, lite := range []bool{false, true} {
		t.Run(fmt.Sprint(lite), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			upstream := newStagedPassthroughConn()
			cfg := passthroughLifecycleConfig()
			cfg.Gateway.OpenAIWS.OAuthEnabled = true
			account := passthroughLifecycleAccount()
			account.Type = AccountTypeOAuth
			account.Credentials = map[string]any{"access_token": "test-token"}
			account.Extra = map[string]any{"openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModePassthrough}
			server, done := startPassthroughLifecycleServer(t, ctx, newPassthroughLifecycleService(cfg, upstream), account)
			defer server.Close()
			payload := fmt.Sprintf(`{"type":"response.create","model":"gpt-5.5","instructions":"x","input":[{"type":"web_search_call","sequence":9007199254740993},{"type":"compaction_trigger"}],"tools":[],"client_metadata":{"ws_request_header_x_openai_internal_codex_responses_lite":"%t"}}`, lite)
			client := dialPassthroughLifecycleClientWithPayload(t, server, payload)
			defer func() { _ = client.CloseNow() }()
			for turn := 1; turn <= 2; turn++ {
				if turn > 1 {
					require.NoError(t, client.Write(ctx, coderws.MessageText, []byte(payload)))
				}
				body := requirePassthroughUpstreamWrite(t, upstream, 3*time.Second)
				require.Equal(t, "none", gjson.GetBytes(body, "tool_choice").String())
				require.Equal(t, "9007199254740993", gjson.GetBytes(body, "input.0.sequence").Raw)
				path := "tools.0"
				if lite {
					path = "input.1.tools.0"
					require.Equal(t, "compaction_trigger", gjson.GetBytes(body, "input.2.type").String())
				}
				require.Equal(t, "web_search", gjson.GetBytes(body, path+".type").String())
				require.Equal(t, gjson.False, gjson.GetBytes(body, path+".external_web_access").Type)
				upstream.Send(fmt.Sprintf(`{"type":"response.completed","response":{"id":"resp_%d","model":"gpt-5.5","usage":{"input_tokens":1,"output_tokens":1}}}`, turn))
				_, err := readPassthroughLifecycleFrame(t, client, 3*time.Second)
				require.NoError(t, err)
			}
			require.NoError(t, client.Close(coderws.StatusNormalClosure, "done"))
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("websocket relay did not finish")
			}
		})
	}
}
