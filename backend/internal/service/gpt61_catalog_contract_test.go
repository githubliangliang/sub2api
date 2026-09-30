//go:build unit

package service

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGPT61SolPublicCatalogPreservesOfficialExtensions(t *testing.T) {
	const groupID int64 = 733
	svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		groupID: {{ID: 23, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"public-alias": "gpt-6.1-sol"}}}},
	}}}
	body, err := svc.BuildCodexModelsManifestForGroup(context.Background(), &Group{ID: groupID, Platform: PlatformComposite}, "", []string{"public-alias"})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, "public-alias", models[0]["slug"])
	var official map[string]any
	require.NoError(t, json.Unmarshal(openai.CodexGPT61SolMetadata, &official))
	for _, field := range []string{"prefer_websockets", "guardian", "minimal_client_version", "available_in_plans", "supports_reasoning_effort_updates"} {
		require.Contains(t, models[0], field)
		require.Equal(t, official[field], models[0][field], field)
	}
	messages := models[0]["model_messages"].(map[string]any)
	for field, value := range official["model_messages"].(map[string]any) {
		require.Contains(t, messages, field)
		require.Equal(t, value, messages[field], field)
	}
}

func TestGPT61SolCatalogPreservesExplicitAccountMetadata(t *testing.T) {
	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-6.1-sol","context_window":300000,"default_reasoning_level":"high","service_tiers":[],"model_messages":{"instructions_template":"operator instructions","auto_review":{"enabled":false}},"unknown":{"kept":true}}]}`)}
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.EqualValues(t, 300000, models[0]["context_window"])
	require.Equal(t, "high", models[0]["default_reasoning_level"])
	require.Empty(t, models[0]["service_tiers"])
	require.Equal(t, map[string]any{"kept": true}, models[0]["unknown"])
	messages := models[0]["model_messages"].(map[string]any)
	require.Equal(t, "operator instructions", messages["instructions_template"])
	require.Equal(t, false, messages["auto_review"].(map[string]any)["enabled"])
}
