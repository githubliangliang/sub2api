package antigravity

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGeminiUsageMapping_NonNegativeInputPreservesBuckets(t *testing.T) {
	for _, tc := range []struct {
		name, prompt string
		wantInput    int
	}{
		{"missing", "", 0},
		{"zero", `"promptTokenCount":0,`, 0},
		{"less than cached", `"promptTokenCount":3,`, 0},
		{"equal to cached", `"promptTokenCount":8,`, 0},
		{"greater than cached", `"promptTokenCount":13,`, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"candidates":[{"content":{"parts":[{"text":"hi"}],"role":"model"}}],"usageMetadata":{%s"cachedContentTokenCount":8,"candidatesTokenCount":3,"thoughtsTokenCount":2,"candidatesTokensDetails":[{"modality":"IMAGE","tokenCount":2}]}}`, tc.prompt)
			want := ClaudeUsage{InputTokens: tc.wantInput, CacheReadInputTokens: 8, OutputTokens: 5, ImageOutputTokens: 2}
			t.Run("non-stream", func(t *testing.T) {
				out, usage, err := TransformGeminiToClaude([]byte(body), "gemini-3.1-pro-preview")
				require.NoError(t, err)
				require.Equal(t, want, *usage)
				var response ClaudeResponse
				require.NoError(t, json.Unmarshal(out, &response))
				require.Equal(t, want, response.Usage)
			})
			t.Run("message_start", func(t *testing.T) {
				p := NewStreamingProcessor("gemini-3.1-pro-preview")
				out := p.ProcessLine(`data: {"response":` + body + `}`)
				found := false
				for _, line := range strings.Split(string(out), "\n") {
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					data := strings.TrimPrefix(line, "data: ")
					if gjson.Get(data, "type").String() == "message_start" {
						found = true
						var startUsage ClaudeUsage
						require.NoError(t, json.Unmarshal([]byte(gjson.Get(data, "message.usage").Raw), &startUsage))
						require.Equal(t, want, startUsage)
					}
				}
				require.True(t, found, "message_start must carry the corrected usage")
				_, usage := p.Finish()
				require.Equal(t, want, *usage)
			})
			t.Run("usage update", func(t *testing.T) {
				p := NewStreamingProcessor("gemini-3.1-pro-preview")
				p.ProcessLine(`data: {"response":{"usageMetadata":{"promptTokenCount":100}}}`)
				p.ProcessLine(`data: {"response":` + body + `}`)
				out, usage := p.Finish()
				require.Equal(t, want, *usage)
				for _, line := range strings.Split(string(out), "\n") {
					if strings.HasPrefix(line, "data: ") {
						data := strings.TrimPrefix(line, "data: ")
						if gjson.Get(data, "type").String() == "message_delta" {
							var deltaUsage ClaudeUsage
							require.NoError(t, json.Unmarshal([]byte(gjson.Get(data, "usage").Raw), &deltaUsage))
							require.Equal(t, want, deltaUsage)
							return
						}
					}
				}
				t.Fatal("message_delta with final usage missing")
			})
		})
	}
}
