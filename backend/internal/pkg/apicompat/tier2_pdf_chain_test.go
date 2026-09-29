package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestTier2PDFResponsesAnthropicGeminiChain(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		wantPDF    bool
	}{
		{"inline", `"file_data":"data:application/pdf;base64,JVBERi0xLjQ="`, true},
		{"empty", `"file_data":"data:application/pdf;base64,"`, false},
		{"file-id", `"file_id":"file_remote"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := json.RawMessage(`[{"role":"user","content":[{"type":"input_text","text":"read"},{"type":"input_file",` + tc.file + `}]}]`)
			anthropic, err := ResponsesToAnthropicRequest(&ResponsesRequest{Model: "gemini-3.1-pro", Input: input})
			require.NoError(t, err)
			raw, err := json.Marshal(anthropic)
			require.NoError(t, err)
			var request antigravity.ClaudeRequest
			require.NoError(t, json.Unmarshal(raw, &request))
			gemini, err := antigravity.TransformClaudeToGemini(&request, "fixture-project", "gemini-3.1-pro")
			require.NoError(t, err)
			parts := gjson.GetBytes(gemini, "request.contents.0.parts").Array()
			if tc.wantPDF {
				require.Len(t, parts, 2)
				require.Equal(t, "application/pdf", parts[1].Get("inlineData.mimeType").String())
				require.Equal(t, "JVBERi0xLjQ=", parts[1].Get("inlineData.data").String())
			} else {
				require.Len(t, parts, 1)
				require.Equal(t, "read", parts[0].Get("text").String())
			}
		})
	}
}
