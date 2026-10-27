package antigravity

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamingThinkingSignatureWire(t *testing.T) {
	for _, text := range []string{"plan", ""} {
		t.Run("thinking_"+text, func(t *testing.T) {
			p := NewStreamingProcessor("claude-sonnet-4-6")
			parts, err := json.Marshal(V1InternalResponse{Response: GeminiResponse{
				Candidates: []GeminiCandidate{{Content: &GeminiContent{Parts: []GeminiPart{{Text: text, Thought: true, ThoughtSignature: "upstream-signature"}}}, FinishReason: "STOP"}},
			}})
			require.NoError(t, err)
			wire := string(p.ProcessLine("data: " + string(parts)))
			var start, signature bool
			for _, line := range strings.Split(wire, "\n") {
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				var event map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event))
				if event["type"] == "content_block_start" {
					block, ok := event["content_block"].(map[string]any)
					require.True(t, ok)
					if block["type"] == "thinking" {
						require.Equal(t, "", block["thinking"])
						require.Equal(t, "", block["signature"])
						start = true
					}
				}
				if delta, ok := event["delta"].(map[string]any); ok && delta["type"] == "signature_delta" {
					require.Equal(t, "upstream-signature", delta["signature"])
					signature = true
				}
			}
			require.True(t, start, wire)
			require.True(t, signature, wire)
		})
	}
}
