package antigravity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolResultPreservesImagesAndDocumentsInOrder(t *testing.T) {
	parts, _, err := buildParts(json.RawMessage(`[{"type":"tool_result","tool_use_id":"tool_1","content":[
		{"type":"text","text":"done"},
		{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aGVsbG8="}},
		{"type":"document","source":{"type":"base64","media_type":"application/pdf","data":"JVBERg=="}},
		{"type":"image","source":{"type":"url","url":"https://images.example.test/remote.png"}}
	]}]`), map[string]string{"tool_1": "read_file"}, false)
	require.NoError(t, err)
	require.Len(t, parts, 3, "functionResponse must be followed by inline image/document data; remote media is never fetched")
	require.NotNil(t, parts[0].FunctionResponse)
	require.Equal(t, "done", parts[0].FunctionResponse.Response["result"])
	require.Equal(t, "image/png", parts[1].InlineData.MimeType)
	require.Equal(t, "aGVsbG8=", parts[1].InlineData.Data)
	require.Equal(t, "application/pdf", parts[2].InlineData.MimeType)
}
