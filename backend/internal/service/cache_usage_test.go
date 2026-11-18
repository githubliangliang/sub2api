package service

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCacheUsagePartialSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name       string
		snapshots  []string
		want       []int
		five, hour int
		cost       float64
	}{
		{"partial details", []string{`{"cache_creation":{"ephemeral_5m_input_tokens":20}}`, `{"cache_creation":{"ephemeral_1h_input_tokens":30}}`, `{"cache_creation":{"ephemeral_5m_input_tokens":0}}`}, []int{20, 50, 30}, 0, 30, 120},
		{"null preserves", []string{`{"cache_creation_input_tokens":100,"cache_creation":{"ephemeral_1h_input_tokens":200}}`, `{"cache_creation_input_tokens":null,"cache_creation":{"ephemeral_1h_input_tokens":null}}`}, []int{100, 100}, 0, 200, 400},
		{"contradiction ratio", []string{`{"cache_creation_input_tokens":100,"cache_creation":{"ephemeral_5m_input_tokens":50,"ephemeral_1h_input_tokens":200}}`}, []int{100}, 50, 200, 360},
		{"remainder five minutes", []string{`{"cache_creation_input_tokens":3000,"cache_creation":{"ephemeral_1h_input_tokens":3000}}`, `{"cache_creation_input_tokens":11000}`}, []int{3000, 11000}, 0, 3000, 28000},
		{"zero retains positive", []string{`{"cache_creation_input_tokens":100,"cache_creation":{"ephemeral_5m_input_tokens":20}}`, `{"cache_creation_input_tokens":0,"cache_creation":{"ephemeral_1h_input_tokens":30}}`}, []int{100, 100}, 20, 30, 260},
		{"zero derives details", []string{`{"cache_creation_input_tokens":0,"cache_creation":{"ephemeral_5m_input_tokens":20}}`, `{"cache_creation_input_tokens":0,"cache_creation":{"ephemeral_1h_input_tokens":30}}`}, []int{20, 50}, 20, 30, 160},
	} {
		for _, parser := range []string{"typed", "ordinary SSE", "passthrough SSE", "Responses SSE", "WS", "raw Chat"} {
			t.Run(tc.name+"/"+parser, func(t *testing.T) {
				var claude ClaudeUsage
				var openai OpenAIUsage
				for i, snapshot := range tc.snapshots {
					switch parser {
					case "typed":
						var src apicompat.AnthropicUsage
						require.NoError(t, json.Unmarshal([]byte(snapshot), &src))
						mergeAnthropicUsage(&claude, src)
					case "ordinary SSE", "passthrough SSE":
						body := `{"type":"message_delta","usage":` + snapshot + `}`
						if i == 0 {
							body = `{"type":"message_start","message":{"usage":` + snapshot + `}}`
						}
						if parser == "ordinary SSE" {
							(&GatewayService{}).parseSSEUsage(body, &claude)
						} else {
							(&GatewayService{}).parseSSEUsagePassthrough(body, &claude)
						}
					case "WS":
						parseOpenAIWSResponseUsageFromCompletedEvent([]byte(`{"type":"response.completed","response":{"usage":`+snapshot+`}}`), &openai)
					case "Responses SSE":
						(&OpenAIGatewayService{}).parseSSEUsageBytes([]byte(`{"type":"response.completed","response":{"usage":`+snapshot+`}}`), &openai)
					case "raw Chat":
						src := extractCCStreamUsage(`{"choices":[],"usage":` + snapshot + `}`)
						require.NotNil(t, src)
						mergeOpenAIUsageSnapshot(&openai, *src)
					}
					if parser == "typed" || parser == "ordinary SSE" || parser == "passthrough SSE" {
						require.Equal(t, tc.want[i], claude.CacheCreationInputTokens)
					} else {
						require.Equal(t, tc.want[i], openai.CacheCreationInputTokens)
					}
				}
				var tokens UsageTokens
				if parser == "typed" || parser == "ordinary SSE" || parser == "passthrough SSE" {
					require.Equal(t, tc.five, claude.CacheCreation5mTokens)
					require.Equal(t, tc.hour, claude.CacheCreation1hTokens)
					tokens = claude.billingUsageTokens()
				} else {
					require.Equal(t, tc.five, openai.CacheCreation5mTokens)
					require.Equal(t, tc.hour, openai.CacheCreation1hTokens)
					tokens = openai.billingUsageTokens()
				}
				pricing := &ModelPricing{SupportsCacheBreakdown: true, CacheCreation5mPrice: 2, CacheCreation1hPrice: 4}
				require.Equal(t, tc.cost, (&BillingService{}).computeCacheCreationCost(pricing, tokens, 0, 1))
			})
		}
	}
}

func TestCacheUsageProtocolRoundtrip(t *testing.T) {
	var input apicompat.AnthropicResponse
	require.NoError(t, json.Unmarshal([]byte(`{"id":"msg_cache","model":"claude-sonnet-4-6","usage":{"input_tokens":10,"output_tokens":2,"cache_creation_input_tokens":100,"cache_read_input_tokens":4,"cache_creation":{"ephemeral_5m_input_tokens":20,"ephemeral_1h_input_tokens":30}}}`), &input))
	responses := apicompat.AnthropicToResponsesResponse(&input)
	chat := apicompat.ResponsesToChatCompletions(responses, "claude-sonnet-4-6")
	replayed := apicompat.ChatCompletionsResponseToResponses(chat, "claude-sonnet-4-6", nil, false, nil)
	raw, err := json.Marshal(replayed)
	require.NoError(t, err)
	var decoded apicompat.ResponsesResponse
	require.NoError(t, json.Unmarshal(raw, &decoded))
	output := apicompat.ResponsesToAnthropic(&decoded, "claude-sonnet-4-6")
	require.Equal(t, input.Usage.CacheCreation, output.Usage.CacheCreation)
	require.Equal(t, 10, output.Usage.InputTokens)
	usage := copyOpenAIUsageFromResponsesUsage(decoded.Usage)
	tokens := usage.billingUsageTokens()
	require.Equal(t, 10, tokens.InputTokens)
	require.Equal(t, 20, tokens.CacheCreation5mTokens)
	require.Equal(t, 30, tokens.CacheCreation1hTokens)
	require.Equal(t, 100, tokens.CacheCreationTokens)
	require.Equal(t, int64(30), gjson.GetBytes(raw, "usage.cache_creation.ephemeral_1h_input_tokens").Int())
}
