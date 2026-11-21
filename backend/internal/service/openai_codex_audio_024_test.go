package service

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCodexCatalogAudioInput(t *testing.T) {
	const sol = "gpt-6.2-sol"
	const astra = "gpt-6.2-astra"
	manifest, err := json.Marshal(map[string][]json.RawMessage{"models": {openai.CodexGPT62SolMetadata, openai.CodexGPT62AstraMetadata}})
	require.NoError(t, err)
	_, metadata, err := extractUpstreamModelCatalog(manifest, false)
	require.NoError(t, err)
	t.Run("sync keeps the declared input modalities", func(t *testing.T) {
		require.Equal(t, []string{"text", "image", "audio"}, metadata[sol].InputModalities)
		require.Equal(t, []string{"text", "image"}, metadata[astra].InputModalities)
	})
	subscription := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	synced := Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	synced.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: metadata})
	key := Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	native := Account{ID: 4, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"base_url": "https://native.example/v1"}}
	native.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: metadata})
	bridged := Account{ID: 5, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: native.Credentials, Extra: map[string]any{"openai_responses_mode": "force_chat_completions"}}
	bridged.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: metadata})
	for _, tc := range []struct {
		name     string
		model    string
		accounts []Account
		want     []any
	}{
		{"Codex subscription", sol, []Account{subscription}, []any{"text", "image", "audio"}},
		{"synced Codex subscription", sol, []Account{synced}, []any{"text", "image", "audio"}},
		{"Astra excludes audio", astra, []Account{synced}, []any{"text", "image"}},
		{"public API key excludes subscription-only audio", sol, []Account{key}, []any{"text", "image"}},
		{"mixed subscription and API key", sol, []Account{synced, key}, []any{"text", "image"}},
		{"native compatible route declares audio", sol, []Account{native}, []any{"text", "image", "audio"}},
		{"mixed native and chat bridge", sol, []Account{native, bridged}, []any{"text", "image"}},
		{"empty group has no audio route", sol, nil, []any{"text", "image"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{1: tc.accounts}}}
			body, err := svc.BuildCodexModelsManifestForGroup(context.Background(), &Group{ID: 1, Platform: PlatformOpenAI}, "", []string{tc.model})
			require.NoError(t, err)
			require.Equal(t, tc.want, decodeCodexManifestModels(t, body)[0]["input_modalities"])
		})
	}
	for _, tc := range []struct {
		account *Account
		want    []any
	}{
		{&native, []any{"text", "image", "audio"}},
		{&bridged, []any{"text", "image"}},
	} {
		response := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-6.2-sol"}]}`)}
		require.NoError(t, (&OpenAIGatewayService{}).CompleteAPIKeyCodexModelsManifestForClient(response, tc.account))
		require.Equal(t, tc.want, decodeCodexManifestModels(t, response.Body)[0]["input_modalities"])
	}
}

func TestCodexCatalogAudioFromLiveAPIKeyManifest(t *testing.T) {
	native := Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"base_url": "https://native.example/v1"}}
	for _, fixture := range []struct {
		name, body string
		converted  bool
	}{
		{"Codex manifest", `{"models":[{"slug":"gpt-6.2-sol","input_modalities":["text","image","audio"]}]}`, false},
		{"OpenAI list", `{"data":[{"id":"gpt-6.2-sol","input_modalities":["text","image","audio"]}]}`, true},
		{"nested modalities", `{"data":[{"id":"gpt-6.2-sol","modalities":{"input":["text","image","audio"]}}]}`, true},
	} {
		for _, tc := range []struct {
			name  string
			extra map[string]any
			want  []any
		}{
			{"native Responses", nil, []any{"text", "image", "audio"}},
			{"forced Chat bridge", map[string]any{"openai_responses_mode": "force_chat_completions"}, []any{"text", "image"}},
			{"probed Chat fallback", map[string]any{"openai_responses_supported": false}, []any{"text", "image"}},
		} {
			t.Run(fixture.name+"/"+tc.name, func(t *testing.T) {
				account := native
				account.Extra = tc.extra
				body := convertOpenAIModelListToCodexManifestForAccount([]byte(fixture.body), &account)
				response := &CodexModelsManifest{
					Body:                         body,
					upstreamSourceBody:           []byte(fixture.body),
					convertedFromOpenAIModelList: fixture.converted,
				}
				require.NoError(t, (&OpenAIGatewayService{}).CompleteAPIKeyCodexModelsManifestForClient(response, &account))
				require.Equal(t, tc.want, decodeCodexManifestModels(t, response.Body)[0]["input_modalities"])
			})
		}
	}
}
