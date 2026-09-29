//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestTier2CodexWebSocketBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, nextWindow, wantPrevious string
		failedWindow, lite             bool
	}{
		{name: "new window", nextWindow: "b"},
		{name: "same window", nextWindow: "a", wantPrevious: "resp_first"},
		{name: "missing window", wantPrevious: "resp_first"},
		{name: "failed window", nextWindow: "b", failedWindow: true},
		{name: "mapped Lite", nextWindow: "b", lite: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newSchedulerTestOpenAIWSV2Config()
			cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
			cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
			cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
			cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 3
			cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3
			model := "gpt-5.1"
			account := &Account{ID: 143, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, Credentials: map[string]any{"api_key": "fixture"}, Extra: map[string]any{"responses_websockets_v2_enabled": true}}
			if tc.lite {
				model = "gpt-5.5"
				account.Type = AccountTypeOAuth
			}
			response := func(id, kind string) []byte {
				body, _ := json.Marshal(map[string]any{"type": kind, "response": map[string]any{"id": id, "model": model, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}}})
				return body
			}
			secondKind := "response.completed"
			if tc.failedWindow {
				secondKind = "response.failed"
			}
			capture := &openAIWSCaptureConn{events: [][]byte{response("resp_first", "response.completed"), response("resp_second", secondKind)}}
			if tc.failedWindow {
				capture.events = append(capture.events, response("resp_third", "response.completed"))
			}
			pool := newOpenAIWSConnPool(cfg)
			pool.setClientDialerForTest(&openAIWSCaptureDialer{conn: capture})
			svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: &httpUpstreamRecorder{}, cache: &stubGatewayCache{}, openaiWSResolver: NewOpenAIWSProtocolResolver(cfg), toolCorrector: NewCodexToolCorrector(), openaiWSPool: pool}
			done := make(chan error, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := coderws.Accept(w, r, nil)
				if err != nil {
					done <- err
					return
				}
				defer conn.CloseNow()
				ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
				defer cancel()
				_, first, err := conn.Read(ctx)
				if err != nil {
					done <- err
					return
				}
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = r
				done <- svc.ProxyResponsesWebSocketFromClient(ctx, c, conn, account, "fixture", first, nil)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
			require.NoError(t, err)
			defer client.CloseNow()
			send := func(window, previous, id string) {
				metadata := map[string]any{}
				if window != "" {
					metadata["x-codex-window-id"] = window
				}
				if tc.lite {
					metadata[responsesLiteWSMetadataKey] = "true"
				}
				request := map[string]any{"type": "response.create", "model": model, "store": true, "input": []any{map[string]any{"type": "input_text", "text": "hello"}}, "client_metadata": metadata}
				if previous != "" {
					request["previous_response_id"] = previous
				}
				raw, err := json.Marshal(request)
				require.NoError(t, err)
				require.NoError(t, client.Write(ctx, coderws.MessageText, raw))
				_, result, err := client.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, id, gjson.GetBytes(result, "response.id").String())
			}
			send("a", "", "resp_first")
			send(tc.nextWindow, "resp_first", "resp_second")
			if tc.failedWindow {
				send("a", "resp_first", "resp_third")
			}
			require.NoError(t, client.Close(coderws.StatusNormalClosure, "done"))
			require.NoError(t, <-done)
			require.GreaterOrEqual(t, len(capture.writes), 2)
			require.Equal(t, tc.wantPrevious, gjson.Get(requestToJSONString(capture.writes[1]), "previous_response_id").String())
			if tc.failedWindow {
				require.Equal(t, "resp_first", gjson.Get(requestToJSONString(capture.writes[2]), "previous_response_id").String())
			}
			if tc.lite {
				for _, payload := range capture.writes {
					require.False(t, isOpenAIResponsesLiteWebSocketPayload([]byte(requestToJSONString(payload))))
				}
			}
		})
	}
}

func TestTier2MappedLiteCreatePayloadDoesNotMutateInput(t *testing.T) {
	for _, kind := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		input := map[string]any{"model": "gpt-5.5", "client_metadata": map[string]any{responsesLiteWSMetadataKey: "true", "keep": "value"}}
		out := (&OpenAIGatewayService{}).buildOpenAIWSCreatePayload(input, &Account{Platform: PlatformOpenAI, Type: kind})
		require.Equal(t, kind == AccountTypeAPIKey, isOpenAIResponsesLiteWebSocketPayload([]byte(requestToJSONString(out))))
		require.True(t, isOpenAIResponsesLiteWebSocketPayload([]byte(requestToJSONString(input))))
	}
}
