//go:build unit

package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMessagesConvertedEffortPolicyAndFinalValidation(t *testing.T) {
	for _, chat := range []bool{false, true} {
		for _, tc := range []struct {
			name, model, fields, from, to, ceiling, limit, want string
			status                                              int
		}{
			{name: "disabled mapped", model: "gpt-6.1-sol", fields: `,"thinking":{"type":"disabled"}`, from: "none", to: "low", want: "low"},
			{name: "disabled capped", model: "gpt-6.1-sol", fields: `,"thinking":{"type":"disabled"}`, from: "none", to: "xhigh", ceiling: "medium", want: "medium"},
			{name: "disabled denied", model: "gpt-6.1-sol", fields: `,"thinking":{"type":"disabled"}`, from: "none", to: "xhigh", ceiling: "low", limit: ReasoningEffortOverLimitDeny, status: 403},
			{name: "disabled still invalid", model: "gpt-6.1-sol", fields: `,"thinking":{"type":"disabled"}`, status: 400},
			{name: "explicit invalid cannot be rescued", model: "gpt-6.1-sol", fields: `,"output_config":{"effort":"none"}`, from: "none", to: "low", status: 400},
			{name: "omitted default", model: "gpt-6.1-sol", from: "medium", to: "low", ceiling: "low", want: "medium"},
			{name: "max", model: "gpt-6.1-sol", fields: `,"output_config":{"effort":"max"}`, want: "max"},
			{name: "Terra none", model: "gpt-6.1-terra", fields: `,"thinking":{"type":"disabled"}`, want: "none"},
			{name: "Terra max", model: "gpt-6.1-terra", fields: `,"output_config":{"effort":"max"}`, want: "max"},
		} {
			t.Run(fmt.Sprintf("%s/chat=%t", tc.name, chat), func(t *testing.T) {
				body := []byte(`{"model":"public","max_tokens":256,"messages":[{"role":"user","content":"hello"}]` + tc.fields + `}`)
				account := rawChatCompletionsTestAccount()
				account.Credentials["model_mapping"] = map[string]any{"public": tc.model}
				mode := openai_compat.ResponsesSupportModeForceResponses
				if chat {
					mode = openai_compat.ResponsesSupportModeForceChatCompletions
				}
				account.Extra = map[string]any{openai_compat.ExtraKeyResponsesMode: string(mode)}
				response := `{"id":"chat_effort","choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`
				contentType := "application/json"
				if !chat {
					contentType = "text/event-stream"
					response = `data: {"type":"response.completed","response":{"id":"resp_effort","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":3,"output_tokens":1}}}` + "\n\n"
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(response))}}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
				var mappings []ReasoningEffortMapping
				if tc.from != "" {
					mappings = []ReasoningEffortMapping{{From: tc.from, To: tc.to}}
				}
				ctx := WithOpenAIReasoningEffortPolicyForMessages(context.Background(), body, tc.ceiling, mappings, tc.limit)
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body)).WithContext(ctx)
				result, err := svc.ForwardAsAnthropic(ctx, c, account, body, "", "")
				if tc.status != 0 {
					require.Error(t, err)
					require.Equal(t, tc.status, rec.Code)
					require.Nil(t, upstream.lastReq)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				path := "reasoning.effort"
				if chat {
					path = "reasoning_effort"
				}
				require.Equal(t, tc.want, gjson.GetBytes(upstream.lastBody, path).String())
				require.NotNil(t, result.ReasoningEffort)
				require.Equal(t, tc.want, *result.ReasoningEffort)
			})
		}
	}
}

func TestCountTokensDisabledThinkingAcceptsMappedSolAndTerra(t *testing.T) {
	for _, model := range []string{"gpt-6.1-sol", "gpt-6.1-terra"} {
		account := rawChatCompletionsTestAccount()
		account.Credentials["model_mapping"] = map[string]any{"public": model}
		prepared, err := prepareOpenAIInputTokensCountRequest([]byte(`{"model":"public","thinking":{"type":"disabled"},"messages":[{"role":"user","content":"hi"}]}`), account, "")
		require.NoError(t, err)
		require.Equal(t, model, prepared.UpstreamModel)
		require.NotEmpty(t, prepared.Request.Input)
	}
}
