package antigravity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMalformedFunctionCallTransform(t *testing.T) {
	for _, body := range []string{
		`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}`,
		`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"thoughtSignature":"test-signature"}]}}]}`,
		`{"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}}`,
		`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"},{"finishReason":"STOP","content":{"parts":[{"text":"other candidate"}]}}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			response, usage, err := TransformGeminiToClaude([]byte(body), "test-model")
			require.ErrorContains(t, err, "MALFORMED_FUNCTION_CALL")
			require.ErrorIs(t, err, ErrMalformedFunctionCall)
			require.Nil(t, response)
			require.Nil(t, usage, "malformed results must not become successful usage")
		})
	}
}

func TestMalformedFunctionCallTransform_UsesSelectedCandidate(t *testing.T) {
	response, usage, err := TransformGeminiToClaude([]byte(`{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"hello"}]}},{"finishReason":"MALFORMED_FUNCTION_CALL"}]}`), "test-model")
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Contains(t, string(response), `"text":"hello"`)
}
