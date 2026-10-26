//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestPersonalFeatures_ExplicitAccountScope(t *testing.T) {
	account := openAIModelNotFoundTempAccount()
	account.Credentials["temp_unschedulable_rules"] = []any{map[string]any{
		"error_code": 503, "keywords": []any{"unavailable"}, "duration_minutes": 5, "scope": "account",
	}}
	raw, err := json.Marshal(account.GetTempUnschedulableRules())
	require.NoError(t, err)
	require.Equal(t, "account", gjson.GetBytes(raw, "0.scope").String())
	repo := &modelNotFoundAccountRepoStub{}
	svc := &RateLimitService{accountRepo: repo}
	require.True(t, svc.HandleTempUnschedulable(context.Background(), account, http.StatusServiceUnavailable,
		[]byte(`{"error":{"message":"unavailable"}}`), "gpt-5.4"))
	require.Equal(t, 1, repo.tempCalls)
	require.Empty(t, repo.modelRateLimitCalls)
}

func TestPersonalFeatures_OpenAIAPIKeySearchMode(t *testing.T) {
	a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Extra: map[string]any{featureKeyWebSearchEmulation: "enabled"}}
	require.Equal(t, WebSearchModeEnabled, a.GetWebSearchEmulationMode())
}
