package service

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

// Every other platform's connection test sends the admin's prompt (default "hi") and lets the
// model answer. The Antigravity text test sent "." with maxOutputTokens 1, ignoring the prompt,
// so thinking models (gemini-4-argon, gemini-3.8-flash) spent that token thinking and the test
// showed "sent: hi" with an empty response.
func TestAntigravityBuildGeminiTestRequest_UsesPromptAndRoomToAnswer(t *testing.T) {
	t.Parallel()

	svc := &AntigravityGatewayService{}
	for _, tc := range []struct{ prompt, want string }{{"hello there", "hello there"}, {"", "hi"}, {"   ", "hi"}} {
		raw, err := svc.buildGeminiTestRequestWithPrompt("proj-123", "gemini-4-argon-high", tc.prompt)
		require.NoError(t, err)
		require.Equal(t, tc.want, gjson.GetBytes(raw, "request.contents.0.parts.0.text").String())
		require.EqualValues(t, antigravityTestMaxOutputTokens, gjson.GetBytes(raw, "request.generationConfig.maxOutputTokens").Int())
		require.GreaterOrEqual(t, antigravityTestMaxOutputTokens, 128, "a thinking model needs room to answer")
	}
}

func TestAntigravityBuildClaudeTestRequest_UsesPromptAndRoomToAnswer(t *testing.T) {
	t.Parallel()

	svc := &AntigravityGatewayService{}
	raw, err := svc.buildClaudeTestRequest("proj-123", "claude-sonnet-4-6", "hello there")
	require.NoError(t, err)
	require.Contains(t, string(raw), "hello there")
	require.NotContains(t, string(raw), `"text":"."`)

	raw, err = svc.buildClaudeTestRequest("proj-123", "claude-sonnet-4-6", "")
	require.NoError(t, err)
	require.Contains(t, string(raw), `"hi"`)
}
