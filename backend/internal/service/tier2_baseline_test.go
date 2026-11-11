//go:build unit

package service

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestTier2GrokAgentMessage(t *testing.T) {
	raw, _, err := patchGrokResponsesBodyWithClientTools([]byte(`{"model":"grok-4.5","input":[{"type":"agent_message","content":[{"type":"input_text","text":"task:"},{"type":"encrypted_content","encrypted_content":"inspect repository"}]}]}`), "grok-4.5")
	require.NoError(t, err)
	require.Equal(t, "message", gjson.GetBytes(raw, "input.0.type").String())
	require.Contains(t, gjson.GetBytes(raw, "input.0.content").String(), "inspect repository")
}

func TestTier2TerraMinimal(t *testing.T) {
	require.Error(t, validateGPT61CompatRequest([]byte(`{"reasoning_effort":"minimal"}`), "gpt-6.1-terra"))
}

func TestTier2TerraSubscription(t *testing.T) {
	require.True(t, openAISubscriptionLongContextExempt(&Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}, &OpenAIForwardResult{Model: "gpt-6.1-terra"}))
}

func TestGrokAgentMessagePreservesUnrelatedRawInput(t *testing.T) {
	const untouched = `{"type":"message", "role":"user", "content":"ordinary", "exact":9007199254740993}`
	body := []byte(`{"model":"grok-4.5","input":[` + untouched + `,{"type":"agent_message","content":[{"type":"input_text","text":"envelope"},{"type":"encrypted_content","encrypted_content":"body"}]},{"type":"agent_message","content":[]}]}`)
	got, err := normalizeGrokAgentMessages(body)
	require.NoError(t, err)
	require.Contains(t, string(got), untouched)
	require.EqualValues(t, 2, gjson.GetBytes(got, "input.#").Int())
	require.Equal(t, "user", gjson.GetBytes(got, "input.1.role").String())
	require.Contains(t, gjson.GetBytes(got, "input.1.content").String(), "envelope")
	require.Contains(t, gjson.GetBytes(got, "input.1.content").String(), "body")
	plain := []byte(`{"input":[` + untouched + `]}`)
	out, err := normalizeGrokAgentMessages(plain)
	require.NoError(t, err)
	require.Equal(t, plain, out)
}
