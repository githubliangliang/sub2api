//go:build unit

package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilterOpenAIResponsesNoneReasoningEffortForAccount_APIKeyPassthroughPreservesRequest(t *testing.T) {
	body := []byte(`{"model":"qwen3.8-27b","input":"hi","max_output_tokens":20,"reasoning":{"effort":"none"},"presence_penalty":1.5}`)
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://compat.example/v1"}, Extra: map[string]any{"openai_passthrough": true}}
	got, err := filterOpenAIResponsesNoneReasoningEffortForAccount(account, body)
	require.NoError(t, err)
	require.JSONEq(t, string(body), string(got))
}
